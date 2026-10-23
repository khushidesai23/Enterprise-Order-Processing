import { expect, test } from "@playwright/test";
import { ADMIN, backend, signIn, stubRazorpay, uniqueEmail } from "./helpers";

test.afterAll(() => backend.cleanup());

test.describe("storefront", () => {
  test("anonymous visitor browses, searches and filters the catalogue", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 25, 1234);

    await page.goto("/");
    await expect(page).toHaveURL(/\/shop$/);
    await expect(page.getByTestId("product-card").first()).toBeVisible();

    // Search by SKU narrows to exactly our product.
    await page.getByLabel("Filter products").fill(product.sku);
    await expect(page.getByTestId("product-card")).toHaveCount(1);
    await expect(page.getByTestId("product-card")).toContainText(product.name);

    // Category chip uses GET /products/category/:id.
    await page.getByLabel("Filter products").fill("");
    await page.getByRole("button", { name: /^E2E Fixtures/ }).click();
    await expect(page).toHaveURL(/category=/);
    await expect(page.getByRole("heading", { name: "E2E Fixtures" })).toBeVisible();
    for (const card of await page.getByTestId("product-card").all()) {
      await expect(card).toContainText("E2E Fixtures");
    }

    // Price filter excludes it, then includes it.
    await page.getByLabel("Maximum price").fill("1000");
    await expect(page.getByText(product.name)).toHaveCount(0);
    await page.getByLabel("Maximum price").fill("2000");
    await expect(page.getByText(product.name)).toBeVisible();

    // Product detail page.
    await page.getByRole("link", { name: product.name }).click();
    await expect(page.getByRole("heading", { name: product.name })).toBeVisible();
    await expect(page.getByText("₹1,234")).toBeVisible();
    await expect(page.getByText("In stock")).toBeVisible();
  });

  test("protected pages send anonymous visitors to sign in", async ({ page }) => {
    await page.goto("/shop/orders");
    await expect(page).toHaveURL(/\/login$/);
    await page.goto("/admin");
    await expect(page).toHaveURL(/\/login$/);
  });

  test("signing out forgets the page; an expired session returns to it", async ({ page }) => {
    const customer = await backend.registerCustomer();
    await signIn(page, customer.email, customer.password);
    await expect(page).toHaveURL(/\/shop$/);
    await page.goto("/shop/orders");
    await page.getByRole("button", { name: "Account menu" }).click();
    await page.getByRole("menuitem", { name: "Sign out" }).click();
    await expect(page).toHaveURL(/\/shop$/);

    // The next person to sign in lands on their own start page.
    await signIn(page, ADMIN.email, ADMIN.password);
    await expect(page).toHaveURL(/\/admin$/);

    // A rejected token sends the user to sign in and back afterwards.
    await page.evaluate(() => localStorage.setItem("meridian.token", "expired"));
    await page.goto("/shop/orders");
    await expect(page).toHaveURL(/\/login$/);
    await page.getByLabel("Email").fill(customer.email);
    await page.getByLabel("Password").fill(customer.password);
    await page.getByRole("button", { name: "Sign in" }).click();
    await expect(page).toHaveURL(/\/shop\/orders$/);
  });

  test("new customer registers, buys, pays and sees the order", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 10, 750);
    await stubRazorpay(page, "success");

    const email = uniqueEmail("shopper");
    await page.goto("/register");
    await page.getByLabel("First name").fill("Asha");
    await page.getByLabel("Last name").fill("Rao");
    await page.getByLabel("Email").fill(email);
    await page.getByLabel("Password", { exact: true }).fill("password123");
    await page.getByLabel("Confirm password").fill("password123");
    await page.getByRole("button", { name: "Create account" }).click();
    await expect(page).toHaveURL(/\/shop$/);
    await expect(page.getByRole("button", { name: "Account menu" })).toContainText("Asha");

    // Add 2 from the product page, then 1 more in the cart.
    await page.goto(`/shop/products/${product.id}`);
    await page.getByRole("button", { name: "Increase quantity" }).click();
    await page.getByRole("button", { name: "Add to cart" }).click();
    await expect(page.getByTestId("cart-count")).toHaveText("2");
    await page.goto("/shop/cart");
    await expect(page.getByTestId("cart-line")).toHaveCount(1);
    await page.getByRole("button", { name: "Increase quantity" }).click();
    await expect(page.getByTestId("cart-count")).toHaveText("3");
    await expect(page.getByText("₹2,250").first()).toBeVisible();

    await page.getByRole("button", { name: "Proceed to checkout" }).click();
    await expect(page).toHaveURL(/\/shop\/checkout$/);
    await page.getByRole("button", { name: /Place order & pay/ }).click();

    // POST /orders -> POST /payments (real Razorpay test order) -> verify.
    await expect(page).toHaveURL(/\/shop\/orders\/[0-9a-f-]{36}$/, { timeout: 30_000 });
    await expect(page.getByText("Thank you! Your order is confirmed.")).toBeVisible();
    await expect(page.getByText("Captured")).toBeVisible();
    await expect(page.getByTestId("cart-count")).toHaveCount(0);

    const orderId = page.url().split("/").pop()!;
    expect((await backend.order(adminToken, orderId)).status).toBe("PAID");
    const stock = await backend.inventory(adminToken, product.id);
    expect(stock).toEqual({ product_id: product.id, available_quantity: 7, reserved_quantity: 3 });

    // Order history lists it.
    await page.goto("/shop/orders");
    await expect(page.getByTestId("order-row").first()).toContainText("Paid");
    await expect(page.getByTestId("order-row").first()).toContainText("₹2,250");

    // Profile update.
    await page.goto("/shop/account");
    await page.getByLabel("First name").fill("Ashwini");
    await page.getByRole("button", { name: "Save changes" }).click();
    await expect(page.getByText("Profile updated")).toBeVisible();
    await expect(page.getByRole("button", { name: "Account menu" })).toContainText("Ashwini");
  });

  test("closed or failed payment can be retried, or the order cancelled", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 5, 300);
    const customer = await backend.registerCustomer();
    const razorpay = await stubRazorpay(page, "dismiss");

    await signIn(page, customer.email, customer.password);
    await expect(page).toHaveURL(/\/shop$/);
    await page.goto(`/shop/products/${product.id}`);
    await page.getByRole("button", { name: "Buy now" }).click();
    await page.getByRole("button", { name: "Proceed to checkout" }).click();
    await page.getByRole("button", { name: /Place order & pay/ }).click();

    // Closing Razorpay leaves the order awaiting payment with stock held.
    await expect(page).toHaveURL(/\/shop\/orders\//, { timeout: 30_000 });
    await expect(page.getByText("Payment pending", { exact: true }).first()).toBeVisible();
    const orderId = page.url().split("/").pop()!;
    expect((await backend.order(customer.token, orderId)).status).toBe("PAYMENT_PENDING");

    // A failed attempt is reported and can be retried.
    razorpay.setMode("fail");
    await page.getByRole("button", { name: "Pay now" }).click();
    await expect(page.getByText("Payment failed")).toBeVisible();

    // Cancelling releases the reserved stock.
    await page.getByRole("button", { name: "Cancel order" }).first().click();
    await page.getByRole("dialog").getByRole("button", { name: "Cancel order" }).click();
    await expect(page.getByText("Order cancelled", { exact: true })).toBeVisible();
    await expect(page.getByText("This order was cancelled.")).toBeVisible();
    expect(await backend.inventory(adminToken, product.id)).toMatchObject({ available_quantity: 5, reserved_quantity: 0 });

    // Retry path: a new order, pay after a dismissal.
    razorpay.setMode("dismiss");
    await page.goto(`/shop/products/${product.id}`);
    await page.getByRole("button", { name: "Buy now" }).click();
    await page.getByRole("button", { name: "Proceed to checkout" }).click();
    await page.getByRole("button", { name: /Place order & pay/ }).click();
    await expect(page.getByRole("button", { name: "Pay now" })).toBeVisible({ timeout: 30_000 });
    razorpay.setMode("success");
    await page.getByRole("button", { name: "Pay now" }).click();
    await expect(page.getByText("Payment received")).toBeVisible();
    await expect(page.getByText("Captured")).toBeVisible();
  });

  test("cart blocks checkout when it asks for more than is in stock", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 1, 100);

    await page.goto(`/shop/products/${product.id}`);
    await page.getByRole("button", { name: "Add to cart" }).click();
    await expect(page.getByRole("button", { name: "Add to cart" })).toBeDisabled();

    // Someone else buys the last unit.
    const other = await backend.registerCustomer();
    await backend.placePaidOrder(other.token, product.id, 1);

    await page.goto("/shop/cart");
    await expect(page.getByText("Out of stock")).toBeVisible();
    await expect(page.getByRole("button", { name: /check out/i })).toBeDisabled();
  });
});
