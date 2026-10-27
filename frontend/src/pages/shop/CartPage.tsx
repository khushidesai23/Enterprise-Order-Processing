import { Link, useNavigate } from "react-router-dom";
import { useQueries, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { AlertTriangle, CheckCircle2, ShoppingCart } from "lucide-react";
import { productsApi } from "@/lib/endpoints";
import { useAuth } from "@/lib/auth";
import { MAX_LINE_QUANTITY, useCart } from "@/lib/cart";
import { money, number } from "@/lib/format";
import { useWishlist } from "@/lib/saved";
import type { Product } from "@/lib/types";
import { ProductImage, ProductRail, distinctProducts } from "@/components/shop";

// Each cart line is checked against the live product so price and stock
// problems surface here rather than as a failed order.
export function useCartAvailability() {
  const { lines } = useCart();
  const products = useQueries({
    queries: lines.map((line) => ({
      queryKey: ["product", line.productId],
      queryFn: () => productsApi.get(line.productId),
      retry: false,
      staleTime: 15_000,
    })),
  });

  const byId = new Map<string, Product>();
  const issues = new Map<string, string>();
  lines.forEach((line, index) => {
    const query = products[index];
    if (!query) return;
    if (query.data) {
      byId.set(line.productId, query.data);
      const available = query.data.inventory?.available_quantity ?? 0;
      if (available <= 0) issues.set(line.productId, "Out of stock");
      else if (line.quantity > available) issues.set(line.productId, `Only ${available} available`);
    } else if (query.isError) {
      issues.set(line.productId, "No longer available");
    }
  });

  return { byId, issues, loading: products.some((query) => query.isLoading) };
}

export function CartPage() {
  const { lines, count, setQuantity, remove } = useCart();
  const wishlist = useWishlist();
  const { user } = useAuth();
  const navigate = useNavigate();
  const { byId, issues, loading } = useCartAvailability();
  const suggestions = useQuery({
    queryKey: ["search", "cart-suggestions"],
    queryFn: () => productsApi.search({ sort: "bestselling", in_stock: true, limit: 60 }),
    staleTime: 60_000,
  });

  const subtotal = lines.reduce((sum, line) => sum + (byId.get(line.productId)?.price ?? line.price) * line.quantity, 0);

  function saveForLater(productId: string) {
    const line = lines.find((item) => item.productId === productId);
    const product = byId.get(productId);
    if (!line) return;
    if (!wishlist.has(productId)) {
      wishlist.toggle({
        id: productId,
        name: line.name,
        brand: product?.brand ?? "",
        price: product?.price ?? line.price,
        mrp: product?.mrp ?? 0,
        image_url: product?.image_url ?? line.imageUrl ?? "",
        rating_average: product?.rating_average ?? 0,
        rating_count: product?.rating_count ?? 0,
        category: line.category,
      });
    }
    remove(productId);
    toast.success("Saved for later in your wishlist");
  }

  return (
    <div className="mx-auto max-w-[1500px] space-y-5 px-3 py-5 sm:px-4">
      <div className="grid items-start gap-5 lg:grid-cols-[1fr_300px]">
        <section className="rounded-lg bg-surface p-5 shadow-card">
          {lines.length === 0 ? (
            <div className="flex flex-col items-start gap-6 py-6 sm:flex-row sm:items-center">
              <ShoppingCart className="size-24 text-faint" />
              <div>
                <h1 className="text-2xl font-semibold">Your Meridian Cart is empty</h1>
                <Link to="/shop/search?deals=true&sort=discount" className="mt-1 block text-sm text-link hover:underline">
                  Shop today's deals
                </Link>
                {!user && (
                  <div className="mt-4 flex gap-2">
                    <Link to="/login" className="rounded-full bg-cart px-5 py-1.5 text-sm font-medium hover:bg-cart-hover">
                      Sign in to your account
                    </Link>
                    <Link to="/register" className="rounded-full border border-border px-5 py-1.5 text-sm hover:bg-muted">
                      Sign up now
                    </Link>
                  </div>
                )}
              </div>
            </div>
          ) : (
            <>
              <div className="flex items-end justify-between border-b border-border pb-2">
                <h1 className="text-2xl font-semibold">Shopping Cart</h1>
                <span className="text-sm text-subtle">Price</span>
              </div>
              <ul className="divide-y divide-border">
                {lines.map((line) => {
                  const live = byId.get(line.productId);
                  const issue = issues.get(line.productId);
                  const price = live?.price ?? line.price;
                  const available = live?.inventory?.available_quantity;
                  const max = Math.max(1, Math.min(available ?? MAX_LINE_QUANTITY, 30));
                  return (
                    <li key={line.productId} className="flex gap-4 py-4" data-testid="cart-line">
                      <Link to={`/shop/products/${line.productId}`} className="shrink-0">
                        <ProductImage id={line.productId} name={line.name} src={live?.image_url ?? line.imageUrl} size="sm" className="size-28 sm:size-44" />
                      </Link>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-start justify-between gap-4">
                          <Link to={`/shop/products/${line.productId}`} className="line-clamp-2 text-base hover:text-link sm:text-lg">
                            {line.name}
                          </Link>
                          <span className="tabular shrink-0 text-lg font-semibold">{money(price * line.quantity)}</span>
                        </div>
                        {issue ? (
                          <p className="mt-1 flex items-center gap-1 text-sm font-medium text-deal">
                            <AlertTriangle className="size-4" /> {issue}
                          </p>
                        ) : (
                          <p className="mt-1 text-sm text-success">In stock</p>
                        )}
                        <p className="text-xs text-subtle">
                          {money(price)} each{live?.brand ? ` · ${live.brand}` : ""}
                        </p>
                        <div className="mt-3 flex flex-wrap items-center gap-3 text-sm">
                          <select
                            value={Math.min(line.quantity, max)}
                            onChange={(event) => setQuantity(line.productId, Number(event.target.value))}
                            className="cursor-pointer rounded-lg border border-border bg-muted px-2 py-1 shadow-sm"
                            aria-label={`Quantity of ${line.name}`}
                          >
                            {Array.from({ length: Math.max(max, line.quantity) }, (_, i) => i + 1).map((n) => (
                              <option key={n} value={n}>
                                Qty: {n}
                              </option>
                            ))}
                          </select>
                          <span className="h-4 w-px bg-border" />
                          <button className="text-link hover:underline" onClick={() => remove(line.productId)} aria-label={`Delete ${line.name}`}>
                            Delete
                          </button>
                          <span className="h-4 w-px bg-border" />
                          <button className="text-link hover:underline" onClick={() => saveForLater(line.productId)}>
                            Save for later
                          </button>
                        </div>
                      </div>
                    </li>
                  );
                })}
              </ul>
              <p className="border-t border-border pt-3 text-right text-lg">
                Subtotal ({number(count)} {count === 1 ? "item" : "items"}): <span className="tabular font-semibold">{money(subtotal)}</span>
              </p>
            </>
          )}
        </section>

        {lines.length > 0 && (
          <aside className="rounded-lg bg-surface p-5 shadow-card lg:sticky lg:top-32">
            {subtotal >= 499 && (
              <p className="flex items-start gap-1.5 text-xs text-success">
                <CheckCircle2 className="size-4 shrink-0" /> Your order is eligible for FREE Delivery.
              </p>
            )}
            <p className="mt-2 text-lg">
              Subtotal ({number(count)} {count === 1 ? "item" : "items"}): <span className="tabular font-semibold">{money(subtotal)}</span>
            </p>
            <button
              disabled={issues.size > 0 || loading}
              onClick={() => navigate(user ? "/shop/checkout" : "/login", { state: { from: "/shop/checkout" } })}
              className="mt-4 w-full rounded-full bg-cart py-2 text-sm font-medium hover:bg-cart-hover disabled:opacity-50"
            >
              {user ? "Proceed to Buy" : "Sign in to check out"}
            </button>
            {issues.size > 0 && <p className="mt-2 text-xs text-deal">Fix the highlighted items to continue.</p>}
          </aside>
        )}
      </div>

      {wishlist.items.length > 0 && (
        <ProductRail
          title="Saved for later"
          action={<Link to="/shop/wishlist" className="text-sm text-link hover:underline">View wishlist</Link>}
          products={wishlist.items}
        />
      )}
      <ProductRail title="Customers also bought" products={distinctProducts(suggestions.data?.items, 18)} loading={suggestions.isLoading} />
    </div>
  );
}
