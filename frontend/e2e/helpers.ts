import { createHmac, randomUUID } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { expect, type Page } from "@playwright/test";

export const ADMIN = {
  email: process.env.ADMIN_EMAIL || "admin@meridian.test",
  password: process.env.ADMIN_PASSWORD || "admin12345",
};

const API = (process.env.E2E_BASE_URL || "http://localhost:5173") + "/api/v1";

// The checkout signature is HMAC-SHA256(order_id|payment_id, key_secret),
// so tests need the API's Razorpay key secret: from the environment, or
// the repository's .env.
function razorpayKeySecret(): string {
  if (process.env.RAZORPAY_KEY_SECRET) return process.env.RAZORPAY_KEY_SECRET;
  const envFile = path.resolve(process.cwd(), "..", ".env");
  if (existsSync(envFile)) {
    const line = readFileSync(envFile, "utf8")
      .split(/\r?\n/)
      .find((entry) => entry.startsWith("RAZORPAY_KEY_SECRET="));
    if (line) return line.slice("RAZORPAY_KEY_SECRET=".length).trim().replace(/^["']|["']$/g, "");
  }
  throw new Error("Set RAZORPAY_KEY_SECRET (the Go API's Razorpay key secret) to run payment e2e tests");
}

export function signCheckout(gatewayOrderId: string, paymentId: string): string {
  return createHmac("sha256", razorpayKeySecret()).update(`${gatewayOrderId}|${paymentId}`).digest("hex");
}

export type CheckoutMode = "success" | "fail" | "dismiss";

// Replaces Razorpay's checkout.js with a stand-in that answers like the
// real widget: success calls the handler with a correctly signed response,
// fail reports payment.failed and closes, dismiss just closes. Everything
// else (POST /payments creating a real Razorpay test order, POST
// /payments/verify checking the signature) runs for real.
export async function stubRazorpay(page: Page, initialMode: CheckoutMode = "success") {
  let mode: CheckoutMode = initialMode;
  await page.exposeFunction("__e2eCheckoutMode", () => mode);
  await page.exposeFunction("__e2eSign", (orderId: string, paymentId: string) => signCheckout(orderId, paymentId));
  await page.route("https://checkout.razorpay.com/v1/checkout.js", (route) =>
    route.fulfill({
      contentType: "application/javascript",
      body: `
        window.Razorpay = class {
          constructor(options) { this.options = options; this.failed = []; }
          on(event, callback) { if (event === "payment.failed") this.failed.push(callback); }
          async open() {
            const mode = await window.__e2eCheckoutMode();
            const options = this.options;
            setTimeout(async () => {
              if (mode === "success") {
                const paymentId = "pay_e2e" + Math.random().toString(36).slice(2, 12);
                const signature = await window.__e2eSign(options.order_id, paymentId);
                options.handler({ razorpay_order_id: options.order_id, razorpay_payment_id: paymentId, razorpay_signature: signature });
              } else {
                if (mode === "fail") this.failed.forEach((cb) => cb({ error: { description: "Card declined by issuer" } }));
                options.modal.ondismiss();
              }
            }, 50);
          }
        };`,
    }),
  );
  return { setMode: (next: CheckoutMode) => (mode = next) };
}

export function uniqueEmail(prefix: string) {
  return `${prefix}-${randomUUID().slice(0, 8)}@e2e.test`;
}

async function api<T>(method: string, route: string, body?: unknown, token?: string): Promise<{ status: number; data: T; error?: string }> {
  const response = await fetch(API + route, {
    method,
    headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const json = (await response.json()) as { data: T; error?: string };
  return { status: response.status, data: json.data, error: json.error };
}

const FIXTURE_CATEGORY = "E2E Fixtures";
const createdProducts: string[] = [];

export const backend = {
  async login(email: string, password: string): Promise<string> {
    const result = await api<{ token: string }>("POST", "/auth/login", { email, password });
    expect(result.status, result.error).toBe(200);
    return result.data.token;
  },
  async registerCustomer(email = uniqueEmail("customer"), password = "password123") {
    const result = await api<{ id: string }>("POST", "/users", { first_name: "Test", last_name: "Shopper", email, password });
    expect(result.status, result.error).toBe(201);
    return { id: result.data.id, email, password, token: await backend.login(email, password) };
  },
  async createProduct(adminToken: string, stock: number, price = 499) {
    const suffix = randomUUID().slice(0, 8).toUpperCase();
    const categories = await api<{ id: string; name: string }[]>("GET", "/categories");
    let category = categories.data.find((c) => c.name === FIXTURE_CATEGORY);
    if (!category) {
      const created = await api<{ id: string; name: string }>("POST", "/categories", { name: FIXTURE_CATEGORY, description: "Created by e2e tests" }, adminToken);
      expect(created.status, created.error).toBe(201);
      category = created.data;
    }
    const product = await api<{ id: string; name: string; sku: string; category: { id: string } }>(
      "POST",
      "/products",
      {
        name: `E2E Fixture ${suffix}`,
        sku: `E2E-${suffix}`,
        price,
        mrp: Math.round(price * 1.25),
        brand: "E2E Brand",
        description: "Fixture product. Made for tests.",
        category_id: category.id,
      },
      adminToken,
    );
    expect(product.status, product.error).toBe(201);
    createdProducts.push(product.data.id);
    const inventory = await api("POST", "/inventory", { product_id: product.data.id, available_quantity: stock }, adminToken);
    expect(inventory.status, inventory.error).toBe(201);
    return product.data;
  },
  // Removes the products, the fixture category and the test accounts (all
  // soft deletes: orders keep their history), so test runs leave the
  // storefront and customer list as they found them.
  async cleanup() {
    const token = await backend.login(ADMIN.email, ADMIN.password);
    for (const id of createdProducts.splice(0)) await api("DELETE", `/products/${id}`, undefined, token);
    const categories = await api<{ id: string; name: string; product_count: number }[]>("GET", "/categories");
    const fixture = categories.data.find((c) => c.name === FIXTURE_CATEGORY);
    if (fixture && fixture.product_count === 0) await api("DELETE", `/categories/${fixture.id}`, undefined, token);
    // Accounts registered by the tests (uniqueEmail) all use @e2e.test.
    const users = await api<{ id: string; email: string }[]>("GET", "/users", undefined, token);
    for (const user of users.data.filter((u) => u.email.endsWith("@e2e.test"))) {
      await api("DELETE", `/users/${user.id}`, undefined, token);
    }
  },
  async createOrder(customerToken: string, productId: string, quantity: number) {
    const order = await api<{ id: string }>("POST", "/orders", { items: [{ product_id: productId, quantity }] }, customerToken);
    expect(order.status, order.error).toBe(201);
    return order.data;
  },
  async review(token: string, productId: string, rating: number, title: string) {
    const result = await api<{ id: string }>("PUT", `/products/${productId}/reviews`, { rating, title, body: "Written by an e2e test." }, token);
    expect(result.status, result.error).toBe(200);
    return result.data;
  },
  async product(productId: string) {
    const result = await api<{ rating_average: number; rating_count: number }>("GET", `/products/${productId}`);
    expect(result.status, result.error).toBe(200);
    return result.data;
  },
  async inventory(token: string, productId: string) {
    const result = await api<{ available_quantity: number; reserved_quantity: number }>("GET", `/inventory/${productId}`, undefined, token);
    expect(result.status, result.error).toBe(200);
    return result.data;
  },
  async placePaidOrder(customerToken: string, productId: string, quantity: number) {
    const order = { data: await backend.createOrder(customerToken, productId, quantity) };
    const checkout = await api<{ gateway_order_id: string; payment_id: string }>("POST", "/payments", { order_id: order.data.id }, customerToken);
    expect(checkout.status, checkout.error).toBe(201);
    const paymentId = "pay_e2e" + randomUUID().replace(/-/g, "").slice(0, 10);
    const verified = await api(
      "POST",
      "/payments/verify",
      {
        razorpay_order_id: checkout.data.gateway_order_id,
        razorpay_payment_id: paymentId,
        razorpay_signature: signCheckout(checkout.data.gateway_order_id, paymentId),
      },
      customerToken,
    );
    expect(verified.status, verified.error).toBe(200);
    return { orderId: order.data.id, paymentId: checkout.data.payment_id };
  },
  async order(token: string, orderId: string) {
    const result = await api<{ status: string }>("GET", `/orders/${orderId}`, undefined, token);
    expect(result.status, result.error).toBe(200);
    return result.data;
  },
};

export async function signIn(page: Page, email: string, password: string) {
  await page.goto("/login");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: "Sign in" }).click();
}

export function shortId(id: string) {
  return "#" + id.replace(/-/g, "").slice(0, 8).toUpperCase();
}
