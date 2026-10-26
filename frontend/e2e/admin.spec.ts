import { expect, test, type Page } from "@playwright/test";
import { randomUUID } from "node:crypto";
import { ADMIN, backend, shortId, signIn } from "./helpers";

async function signInAsAdmin(page: Page) {
  await signIn(page, ADMIN.email, ADMIN.password);
  await expect(page).toHaveURL(/\/admin$/);
}

test.afterAll(() => backend.cleanup());

test.describe("admin console", () => {
  test("customers cannot open the admin console", async ({ page }) => {
    const customer = await backend.registerCustomer();
    await signIn(page, customer.email, customer.password);
    await expect(page).toHaveURL(/\/shop$/);
    await page.goto("/admin");
    await expect(page).toHaveURL(/\/shop$/);
  });

  test("overview shows KPIs, charts, recent orders and stock warnings", async ({ page }) => {
    await signInAsAdmin(page);
    await expect(page.getByRole("heading", { name: "Overview" })).toBeVisible();
    for (const label of ["Captured revenue", "Total orders", "Low-stock products", "Customers"]) {
      await expect(page.getByText(label, { exact: true }).first()).toBeVisible();
    }
    await expect(page.getByRole("heading", { name: "Orders by status" })).toBeVisible();
    await expect(page.locator("path.recharts-rectangle").first()).toBeVisible();
    await expect(page.getByRole("heading", { name: "Recent orders" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Low-stock warnings" })).toBeVisible();
  });

  test("category and product management", async ({ page }) => {
    const suffix = randomUUID().slice(0, 6).toUpperCase();
    const category = `Garden ${suffix}`;
    const product = `Watering Can ${suffix}`;
    await signInAsAdmin(page);

    // Create, rename and later delete a category.
    await page.getByRole("link", { name: "Categories", exact: true }).click();
    await page.getByRole("button", { name: "New category" }).click();
    await page.getByLabel("Name").fill(category);
    await page.getByLabel("Description").fill("Outdoor tools");
    await page.getByRole("button", { name: "Create category" }).click();
    await expect(page.getByText("Category created")).toBeVisible();
    const row = page.getByTestId("admin-category-row").filter({ hasText: category });
    await expect(row).toContainText("Outdoor tools");
    await row.getByRole("button", { name: `Edit ${category}` }).click();
    await page.getByLabel("Description").fill("Outdoor tools and planters");
    await page.getByRole("button", { name: "Save changes" }).click();
    await expect(row).toContainText("Outdoor tools and planters");

    // Create a product in it with starting stock.
    await page.getByRole("link", { name: "Products", exact: true }).click();
    await page.getByRole("button", { name: "New product" }).click();
    await page.getByLabel("Name").fill(product);
    await page.getByLabel("Description").fill("Galvanised steel, 5 litres.");
    await page.getByLabel("SKU").fill(`GDN-${suffix}`);
    await page.getByLabel("Price (₹)").fill("899.50");
    await page.getByLabel("Category", { exact: true }).selectOption({ label: category });
    await page.getByLabel("Starting stock").fill("3");
    await page.getByRole("button", { name: "Create product" }).click();
    await expect(page.getByText("Product created")).toBeVisible();
    await page.getByLabel("Search products").fill(suffix);
    const productRow = page.getByTestId("admin-product-row").filter({ hasText: product });
    await expect(productRow).toContainText("₹899.50");
    await expect(productRow).toContainText("Low stock · 3");

    // Duplicate SKU is rejected with the API's message.
    await page.getByRole("button", { name: "New product" }).click();
    await page.getByLabel("Name").fill(`${product} copy`);
    await page.getByLabel("SKU").fill(`GDN-${suffix}`);
    await page.getByLabel("Price (₹)").fill("10");
    await page.getByLabel("Category", { exact: true }).selectOption({ label: category });
    await page.getByRole("button", { name: "Create product" }).click();
    await expect(page.getByText("Product SKU already exists")).toBeVisible();
    await page.getByRole("button", { name: "Cancel" }).click();

    // Edit price; the storefront shows the change.
    await productRow.getByRole("button", { name: `Edit ${product}` }).click();
    await expect(page.getByLabel("Description")).toHaveValue("Galvanised steel, 5 litres.");
    await page.getByLabel("Price (₹)").fill("949");
    await page.getByRole("button", { name: "Save changes" }).click();
    await expect(productRow).toContainText("₹949");

    // A category with products cannot be deleted.
    await page.getByRole("link", { name: "Categories", exact: true }).click();
    await expect(row.getByRole("button", { name: `Delete ${category}` })).toBeDisabled();

    // Delete the product, then the now-empty category.
    await page.getByRole("link", { name: "Products", exact: true }).click();
    await page.getByLabel("Search products").fill(suffix);
    await productRow.getByRole("button", { name: `Delete ${product}` }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Delete product" }).click();
    await expect(page.getByText("Product deleted")).toBeVisible();
    await expect(productRow).toHaveCount(0);

    await page.getByRole("link", { name: "Categories", exact: true }).click();
    await row.getByRole("button", { name: `Delete ${category}` }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Delete category" }).click();
    await expect(page.getByText("Category deleted")).toBeVisible();
    await expect(row).toHaveCount(0);
  });

  test("inventory stock operations and low-stock warnings", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 4);
    await signInAsAdmin(page);
    await page.getByRole("link", { name: "Inventory", exact: true }).click();

    await page.getByLabel("Search inventory").fill(product.sku);
    const row = page.getByTestId("inventory-row");
    await expect(row).toHaveCount(1);
    await expect(row).toContainText(product.name);
    await expect(row).toContainText("Low stock · 4");

    await row.getByRole("button", { name: "Add" }).click();
    await page.getByLabel("Quantity").fill("20");
    await page.getByRole("dialog").getByRole("button", { name: "Add" }).click();
    await expect(row).toContainText("In stock");
    await expect(row.locator("td").nth(1)).toHaveText("24");

    await row.getByRole("button", { name: "Remove" }).click();
    await page.getByLabel("Quantity").fill("4");
    await page.getByRole("dialog").getByRole("button", { name: "Remove" }).click();
    await expect(row.locator("td").nth(1)).toHaveText("20");

    // Removing more than is available surfaces the API error.
    await row.getByRole("button", { name: "Remove" }).click();
    await page.getByLabel("Quantity").fill("500");
    await page.getByRole("dialog").getByRole("button", { name: "Remove" }).click();
    await expect(page.getByText("Could not update inventory")).toBeVisible();
    await page.getByRole("button", { name: "Cancel" }).click();

    // Reservation lifecycle from the overflow menu.
    await row.getByRole("button", { name: `More stock actions for ${product.name}` }).click();
    await page.getByRole("menuitem", { name: "Reserve stock" }).click();
    await page.getByLabel("Quantity").fill("5");
    await page.getByRole("dialog").getByRole("button", { name: "Reserve" }).click();
    await expect(row.locator("td").nth(2)).toHaveText("5");
    await row.getByRole("button", { name: `More stock actions for ${product.name}` }).click();
    await page.getByRole("menuitem", { name: "Release reservation" }).click();
    await page.getByLabel("Quantity").fill("2");
    await page.getByRole("dialog").getByRole("button", { name: "Release" }).click();
    await expect(row.locator("td").nth(2)).toHaveText("3");
    await row.getByRole("button", { name: `More stock actions for ${product.name}` }).click();
    await page.getByRole("menuitem", { name: "Confirm reservation" }).click();
    await page.getByLabel("Quantity").fill("3");
    await page.getByRole("dialog").getByRole("button", { name: "Confirm" }).click();
    await expect(row.locator("td").nth(2)).toHaveText("0");

    // Set exact levels.
    await row.getByRole("button", { name: `More stock actions for ${product.name}` }).click();
    await page.getByRole("menuitem", { name: "Set exact levels" }).click();
    await page.getByLabel("Available").fill("2");
    await page.getByLabel("Reserved").fill("0");
    await page.getByRole("dialog").getByRole("button", { name: "Save" }).click();
    await expect(row).toContainText("Low stock · 2");

    expect(await backend.inventory(adminToken, product.id)).toMatchObject({ available_quantity: 2, reserved_quantity: 0 });
  });

  test("order fulfilment from paid to delivered", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 6);
    const customer = await backend.registerCustomer();
    const { orderId } = await backend.placePaidOrder(customer.token, product.id, 2);

    await signInAsAdmin(page);
    await page.getByRole("link", { name: "Orders", exact: true }).click();
    await page.getByRole("tab", { name: /^Paid/ }).click();
    await expect(page).toHaveURL(/status=PAID/);
    await page.getByRole("link", { name: shortId(orderId), exact: true }).click();
    await expect(page.getByRole("heading", { name: new RegExp(shortId(orderId)) })).toBeVisible();
    await expect(page.getByText(customer.email)).toBeVisible();
    await expect(page.getByText("Captured")).toBeVisible();

    for (const [action, label] of [
      ["Mark as packed", "Packed"],
      ["Mark as shipped", "Shipped"],
      ["Mark as delivered", "Delivered"],
    ] as const) {
      await page.getByRole("button", { name: action }).click();
      await page.getByRole("dialog").getByRole("button", { name: action }).click();
      await expect(page.getByRole("heading").getByText(label, { exact: true })).toBeVisible();
    }
    expect((await backend.order(adminToken, orderId)).status).toBe("DELIVERED");
    // Shipping consumed the reserved stock.
    expect(await backend.inventory(adminToken, product.id)).toMatchObject({ available_quantity: 4, reserved_quantity: 0 });

    // The customer sees it delivered.
    await page.getByRole("button", { name: "Account menu" }).click();
    await page.getByRole("menuitem", { name: "Sign out" }).click();
    await signIn(page, customer.email, customer.password);
    await expect(page).toHaveURL(/\/shop$/);
    await page.goto(`/shop/orders/${orderId}`);
    await expect(page.getByText("Delivered").first()).toBeVisible();
  });

  test("admin cancels an unpaid order", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 3);
    const customer = await backend.registerCustomer();
    const order = await backend.createOrder(customer.token, product.id, 2);

    await signInAsAdmin(page);
    await page.goto(`/admin/orders/${order.id}`);
    await page.getByRole("button", { name: "Cancel order" }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Cancel order" }).click();
    await expect(page.getByRole("heading").getByText("Cancelled")).toBeVisible();
    expect(await backend.inventory(adminToken, product.id)).toMatchObject({ available_quantity: 3, reserved_quantity: 0 });
  });

  test("payments list with summary and refund action", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 2);
    const customer = await backend.registerCustomer();
    const { paymentId } = await backend.placePaidOrder(customer.token, product.id, 1);

    await signInAsAdmin(page);
    await page.getByRole("link", { name: "Payments", exact: true }).click();
    await page.getByRole("tab", { name: "Captured" }).click();
    await page.getByLabel("Search payments").fill(paymentId.slice(0, 8));
    const row = page.getByTestId("admin-payment-row");
    await expect(row).toHaveCount(1);
    await expect(row).toContainText("Captured");

    // The refund goes to Razorpay. The stubbed checkout produced no real
    // Razorpay payment, so the gateway rejects it and the UI must say so
    // and leave everything unchanged (the API rolls back).
    await row.getByRole("button", { name: "Refund" }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Issue refund" }).click();
    await expect(page.getByText("Refund failed", { exact: true })).toBeVisible();
    await expect(page.getByText("payment gateway refund failed", { exact: false })).toBeVisible();
    await page.getByRole("button", { name: "Keep it" }).click();
    await expect(row).toContainText("Captured");
  });

  test("customer management: roles and account status", async ({ page }) => {
    const customer = await backend.registerCustomer();
    await signInAsAdmin(page);
    await page.getByRole("link", { name: "Customers", exact: true }).click();
    await page.getByLabel("Search users").fill(customer.email);
    const row = page.getByTestId("admin-user-row");
    await expect(row).toHaveCount(1);
    await expect(row).toContainText("Customer");

    const open = () => row.getByRole("button", { name: /Actions for/ }).click();

    await open();
    await page.getByRole("menuitem", { name: "Make admin" }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Make admin" }).click();
    await expect(row).toContainText("Admin");

    await open();
    await page.getByRole("menuitem", { name: "Remove admin" }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Remove admin" }).click();
    await expect(row).toContainText("Customer");

    await open();
    await page.getByRole("menuitem", { name: "Deactivate" }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Deactivate" }).click();
    await expect(row).toContainText("Deactivated");

    await open();
    await page.getByRole("menuitem", { name: "Reactivate" }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Reactivate" }).click();
    await expect(row).toContainText("Active");

    await open();
    await page.getByRole("menuitem", { name: "Delete user" }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Delete user" }).click();
    await expect(row).toHaveCount(0);

    // Admins cannot act on their own account from here.
    await page.getByLabel("Search users").fill(ADMIN.email);
    await expect(page.getByTestId("admin-user-row")).toContainText("(you)");
    await expect(page.getByTestId("admin-user-row").getByRole("button")).toHaveCount(0);
  });

  test("review moderation recalculates the product rating", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 5);
    const happy = await backend.registerCustomer();
    const grumpy = await backend.registerCustomer();
    await backend.review(happy.token, product.id, 5, "Lovely fixture");
    await backend.review(grumpy.token, product.id, 1, `Spam ${product.sku}`);
    expect(await backend.product(product.id)).toMatchObject({ rating_average: 3, rating_count: 2 });

    await signInAsAdmin(page);
    await page.getByRole("link", { name: "Reviews", exact: true }).click();
    await page.getByRole("tab", { name: "1 star" }).click();
    const row = page.getByTestId("admin-review-row").filter({ hasText: `Spam ${product.sku}` });
    await expect(row).toContainText(product.name);
    await row.getByRole("button", { name: /Delete review by/ }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Delete review" }).click();
    await expect(page.getByText("Review deleted", { exact: true })).toBeVisible();
    await expect(row).toHaveCount(0);

    expect(await backend.product(product.id)).toMatchObject({ rating_average: 5, rating_count: 1 });
  });
});
