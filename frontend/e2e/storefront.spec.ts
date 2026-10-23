import { expect, test, type Page } from "@playwright/test";
import { ADMIN, backend, signIn, stubRazorpay, uniqueEmail } from "./helpers";

test.afterAll(() => backend.cleanup());

async function searchFor(page: Page, text: string) {
  await page.getByRole("search").getByLabel("Search products").fill(text);
  await page.getByRole("search").getByRole("button", { name: "Search" }).click();
  await expect(page).toHaveURL(/\/shop\/search\?/);
}

async function checkoutFromProduct(page: Page, productId: string) {
  await page.goto(`/shop/products/${productId}`);
  await page.getByRole("button", { name: "Buy Now" }).click();
  await expect(page).toHaveURL(/\/shop\/checkout$/);
  await page.getByRole("button", { name: /Place order & pay/ }).click();
}

test.describe("storefront", () => {
  test("home page shows departments and product rails", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL(/\/shop$/);
    await expect(page.getByRole("heading", { name: "Today's Deals" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Best Sellers", exact: true })).toBeVisible();
    await expect(page.getByTestId("product-card").first()).toBeVisible();
    await expect(page.getByRole("link", { name: /^See more in / }).first()).toBeVisible();

    // The "All" drawer lists every department.
    await page.getByRole("button", { name: "All", exact: true }).click();
    await expect(page.getByRole("dialog").getByText("Shop by Department")).toBeVisible();
    await page.getByRole("dialog").getByRole("link", { name: "Best Sellers" }).click();
    await expect(page).toHaveURL(/sort=bestselling/);
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

    // Brand facet.
    // The checkbox follows the URL, which updates just after the click.
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
    await expect(page.getByText("-20%")).toBeVisible();
    await expect(page.getByText("Visit the E2E Brand Store")).toBeVisible();
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

    // Add 2 from the product page, then make it 3 in the cart.
    await page.goto(`/shop/products/${product.id}`);
    await page.getByLabel("Quantity").selectOption("2");
    await page.getByRole("button", { name: "Add to Cart" }).click();
    await expect(page.getByTestId("cart-count")).toHaveText("2");
    await page.goto("/shop/cart");
    await expect(page.getByTestId("cart-line")).toHaveCount(1);
    await page.getByLabel(`Quantity of ${product.name}`).selectOption("3");
    await expect(page.getByTestId("cart-count")).toHaveText("3");
    await expect(page.getByText("₹2,250.00").first()).toBeVisible();

    await page.getByRole("button", { name: "Proceed to Buy" }).click();
    await expect(page).toHaveURL(/\/shop\/checkout$/);
    await page.getByRole("button", { name: /Place order & pay/ }).click();

    // POST /orders -> POST /payments (real Razorpay test order) -> verify.
    await expect(page).toHaveURL(/\/shop\/orders\/[0-9a-f-]{36}$/, { timeout: 30_000 });
    await expect(page.getByText("Thank you! Your order is confirmed.")).toBeVisible();
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

  test("wishlist: save, move to cart and save for later", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 5, 300);

    await page.goto(`/shop/products/${product.id}`);
    await page.getByRole("button", { name: `Add ${product.name} to wishlist` }).click();
    await expect(page.getByText("Added to your wishlist")).toBeVisible();

    await page.goto("/shop/wishlist");
    await expect(page.getByTestId("wishlist-item")).toContainText(product.name);
    await page.getByRole("button", { name: "Move to cart" }).click();
    await expect(page.getByTestId("wishlist-item")).toHaveCount(0);
    await expect(page.getByTestId("cart-count")).toHaveText("1");

    await page.goto("/shop/cart");
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

  test("cart blocks checkout when it asks for more than is in stock", async ({ page }) => {
    const adminToken = await backend.login(ADMIN.email, ADMIN.password);
    const product = await backend.createProduct(adminToken, 1, 100);

    await page.goto(`/shop/products/${product.id}`);
    await page.getByRole("button", { name: "Add to Cart" }).click();
    await expect(page.getByRole("button", { name: "Add to Cart" })).toBeDisabled();

    // Someone else buys the last unit.
    const other = await backend.registerCustomer();
    await backend.placePaidOrder(other.token, product.id, 1);

    await page.goto("/shop/cart");
    await expect(page.getByText("Out of stock")).toBeVisible();
    await expect(page.getByRole("button", { name: /check out/i })).toBeDisabled();
  });
});
