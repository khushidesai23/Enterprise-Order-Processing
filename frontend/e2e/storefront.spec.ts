import { expect, test, type Page } from "@playwright/test";
import { ADMIN, backend, signIn, stubRazorpay, uniqueEmail } from "./helpers";

test.afterAll(() => backend.cleanup());

// The header search opens a command palette; submitting goes to the index.
async function searchFor(page: Page, text: string) {
  await page.getByRole("button", { name: "Open search" }).click();
  const palette = page.getByRole("dialog");
  await palette.getByLabel("Search products").fill(text);
  await palette.getByRole("button", { name: "Search", exact: true }).click();
  await expect(page).toHaveURL(/\/shop\/search\?/);
}

async function checkoutFromProduct(page: Page, productId: string) {
  await page.goto(`/shop/products/${productId}`);
  await page.getByRole("button", { name: "Buy now" }).click();
  await expect(page).toHaveURL(/\/shop\/checkout$/);
  await page.getByRole("button", { name: /Place order & pay/ }).click();
}

test.describe("storefront", () => {
  test("home page shows live sections and the departments menu", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL(/\/shop$/);
    await expect(page.getByRole("heading", { name: "Objects for the next decade." })).toBeVisible();
    await expect(page.getByText(/objects/).first()).toBeVisible();
    await expect(page.getByRole("heading", { name: "Today's drops." })).toBeVisible();
    await expect(page.getByTestId("product-card").first()).toBeVisible();

    // The departments menu lists live departments and navigates.
    await page.getByRole("button", { name: "Departments" }).click();
    const menu = page.getByRole("dialog");
    await expect(menu.getByText("Departments", { exact: true })).toBeVisible();
    await menu.getByRole("link", { name: "Most wanted" }).click();
    await expect(page).toHaveURL(/sort=bestselling/);
    await expect(page.getByRole("heading", { name: /Most wanted/ })).toBeVisible();
  });

  test("search palette shows live results", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 9, 999);

    await page.goto("/shop");
    await page.keyboard.press("Control+k");
    const palette = page.getByRole("dialog");
    await palette.getByLabel("Search products").fill(product.sku);
    await expect(palette.getByText("1 matches")).toBeVisible();
    await palette.getByRole("button", { name: new RegExp(product.name) }).click();
    await expect(page).toHaveURL(new RegExp(`/shop/products/${product.id}$`));
  });

  test("search with department, brand, price and rating filters and sorting", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const cheap = await backend.createProduct(adminToken, 25, 1234);
    const dear = await backend.createProduct(adminToken, 25, 5678);

    await page.goto("/shop");
    await searchFor(page, cheap.sku);
    await expect(page.getByTestId("product-card")).toHaveCount(1);
    await expect(page.getByTestId("product-card")).toContainText(cheap.name);
    await expect(page.getByTestId("result-summary")).toContainText(`"${cheap.sku}"`);

    // Department facet (GET /products/search?category_id=...).
    await page.getByRole("button", { name: `Remove filter "${cheap.sku}"` }).click();
    await page.getByRole("complementary", { name: "Filters" }).getByRole("button", { name: /^E2E Fixtures/ }).click();
    await expect(page).toHaveURL(/category_id=/);
    await expect(page.getByTestId("result-summary")).toContainText("in E2E Fixtures");

    // Brand facet. The checkbox follows the URL, which updates just after the click.
    await page.getByRole("checkbox", { name: /E2E Brand/ }).click();
    await expect(page.getByRole("checkbox", { name: /E2E Brand/ })).toBeChecked();
    await expect(page).toHaveURL(/brand=E2E\+Brand/);
    for (const card of await page.getByTestId("product-card").all()) {
      await expect(card).toContainText("E2E Brand");
    }

    // Price range.
    await page.getByLabel("Maximum price").fill("2000");
    await page.getByRole("button", { name: "Go" }).click();
    await expect(page.getByText(cheap.name)).toBeVisible();
    await expect(page.getByText(dear.name)).toHaveCount(0);
    await page.getByRole("button", { name: /Remove filter ₹/ }).click();
    await expect(page.getByText(dear.name)).toBeVisible();

    // Sorting.
    await page.getByLabel("Sort results").selectOption("price_desc");
    await expect(page).toHaveURL(/sort=price_desc/);
    await expect(page.getByTestId("product-card").first()).toContainText(dear.name);

    // Rating filter: new fixtures have no reviews.
    await page.getByRole("button", { name: "4 stars and up" }).click();
    await expect(page.getByText(cheap.name)).toHaveCount(0);

    // Product page.
    await page.goto(`/shop/products/${cheap.id}`);
    await expect(page.getByRole("heading", { name: cheap.name })).toBeVisible();
    await expect(page.getByText("-20%").first()).toBeVisible();
    await expect(page.getByRole("navigation", { name: "Breadcrumb" })).toContainText("E2E Brand");
    await expect(page.getByText("In stock", { exact: true })).toBeVisible();
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

  test("new customer registers, buys, pays, reviews and sees the order", async ({ page }) => {
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

    // Add 2 from the product page, then make it 3 in the bag.
    await page.goto(`/shop/products/${product.id}`);
    await page.getByRole("button", { name: "Increase quantity", exact: true }).click();
    await page.getByRole("button", { name: "Add to bag" }).click();
    await expect(page.getByTestId("cart-count")).toHaveText("2");
    await page.goto("/shop/cart");
    await expect(page.getByTestId("cart-line")).toHaveCount(1);
    await page.getByRole("button", { name: /Increase quantity of/i }).click();
    await expect(page.getByTestId("cart-count")).toHaveText("3");
    await expect(page.getByText("₹2,250.00").first()).toBeVisible();

    await page.getByRole("button", { name: "Checkout" }).click();
    await expect(page).toHaveURL(/\/shop\/checkout$/);
    await page.getByRole("button", { name: /Place order & pay/ }).click();

    // POST /orders -> POST /payments (real Razorpay test order) -> verify.
    await expect(page).toHaveURL(/\/shop\/orders\/[0-9a-f-]{36}$/, { timeout: 30_000 });
    await expect(page.getByText("Thank you! Your order is confirmed.", { exact: false })).toBeVisible();
    await expect(page.getByText("Captured")).toBeVisible();
    await expect(page.getByTestId("cart-count")).toHaveText("0");

    const orderId = page.url().split("/").pop()!;
    expect((await backend.order(adminToken, orderId)).status).toBe("PAID");
    expect(await backend.inventory(adminToken, product.id)).toMatchObject({ available_quantity: 7, reserved_quantity: 3 });

    // Order history lists it.
    await page.goto("/shop/orders");
    await expect(page.getByTestId("order-row").first()).toContainText("Paid");
    await expect(page.getByTestId("order-row").first()).toContainText("₹2,250");

    // Having bought it, the review is a verified purchase.
    await page.goto(`/shop/products/${product.id}`);
    await page.getByRole("button", { name: "Write a product review" }).click();
    await page.getByRole("radio", { name: "4 stars" }).click();
    await page.getByLabel("Add a headline").fill("Does the job");
    await page.getByLabel("Add a written review").fill("Sturdy and arrived quickly.");
    await page.getByRole("button", { name: "Submit review" }).click();
    await expect(page.getByText("Thanks! Your verified review is live.")).toBeVisible();
    const review = page.getByTestId("review").filter({ hasText: "Does the job" });
    await expect(review).toContainText("Verified Purchase");
    await expect(review).toContainText("Asha R.");
    await expect(page.getByText("1 global ratings")).toBeVisible();

    // Profile update.
    await page.goto("/shop/account");
    await page.getByLabel("First name").fill("Ashwini");
    await page.getByRole("button", { name: "Save changes" }).click();
    await expect(page.getByText("Profile updated")).toBeVisible();
    await expect(page.getByRole("button", { name: "Account menu" })).toContainText("Ashwini");
  });

  test("bag drawer, wishlist and save for later", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 5, 300);

    await page.goto(`/shop/products/${product.id}`);
    await page.getByRole("button", { name: `Add ${product.name} to wishlist` }).click();
    await expect(page.getByText("Added to your wishlist")).toBeVisible();

    await page.goto("/shop/wishlist");
    await expect(page.getByTestId("wishlist-item")).toContainText(product.name);
    await page.getByRole("button", { name: "Move to bag" }).click();
    await expect(page.getByTestId("wishlist-item")).toHaveCount(0);
    await expect(page.getByTestId("cart-count")).toHaveText("1");

    // The bag drawer shows the line with live price and stepper.
    await page.getByRole("button", { name: /^Bag, 1 items/ }).click();
    const drawer = page.getByRole("dialog");
    await expect(drawer.getByText(product.name)).toBeVisible();
    await drawer.getByRole("link", { name: "View bag" }).click();
    await expect(page).toHaveURL(/\/shop\/cart$/);

    await page.getByRole("button", { name: "Save for later" }).click();
    await expect(page.getByTestId("cart-line")).toHaveCount(0);
    await expect(page.getByRole("heading", { name: "Saved for later" })).toBeVisible();
  });

  test("closed or failed payment can be retried, or the order cancelled", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 5, 300);
    const customer = await backend.registerCustomer();
    const razorpay = await stubRazorpay(page, "dismiss");

    await signIn(page, customer.email, customer.password);
    await expect(page).toHaveURL(/\/shop$/);
    await checkoutFromProduct(page, product.id);

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

    // Retry path: a new order, paid after a dismissal.
    razorpay.setMode("dismiss");
    await checkoutFromProduct(page, product.id);
    await expect(page.getByRole("button", { name: "Pay now" })).toBeVisible({ timeout: 30_000 });
    razorpay.setMode("success");
    await page.getByRole("button", { name: "Pay now" }).click();
    await expect(page.getByText("Payment received")).toBeVisible();
    await expect(page.getByText("Captured")).toBeVisible();
  });

  test("bag blocks checkout when it asks for more than is in stock", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 1, 100);

    await page.goto(`/shop/products/${product.id}`);
    await page.getByRole("button", { name: "Add to bag" }).click();
    await expect(page.getByRole("button", { name: "Add to bag" })).toBeDisabled();

    // Someone else buys the last unit.
    const other = await backend.registerCustomer();
    await backend.placePaidOrder(other.token, product.id, 1);

    await page.goto("/shop/cart");
    await expect(page.getByText("Sold out", { exact: true })).toBeVisible();
    await expect(page.getByRole("button", { name: /check out/i })).toBeDisabled();
  });
});
