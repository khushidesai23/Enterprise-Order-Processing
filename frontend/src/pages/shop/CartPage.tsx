import { Link, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { AlertTriangle, ArrowRight, ShoppingBag, Trash2 } from "lucide-react";
import { productsApi } from "@/lib/endpoints";
import { useAuth } from "@/lib/auth";
import { useCart } from "@/lib/cart";
import { money } from "@/lib/format";
import { Button } from "@/components/ui/button";
import { Card, EmptyState } from "@/components/ui/primitives";
import { ProductArt, QuantityStepper } from "@/components/common";

// Cart lines are checked against live stock from GET /products so problems
// surface here rather than as a failed order.
export function useCartAvailability() {
  const { lines } = useCart();
  const products = useQuery({ queryKey: ["products", { categoryId: "" }], queryFn: productsApi.list });
  const byId = new Map((products.data ?? []).map((product) => [product.id, product]));

  const issues = new Map<string, string>();
  if (products.data) {
    for (const line of lines) {
      const product = byId.get(line.productId);
      if (!product) issues.set(line.productId, "No longer available");
      else if (product.available_quantity <= 0) issues.set(line.productId, "Out of stock");
      else if (line.quantity > product.available_quantity)
        issues.set(line.productId, `Only ${product.available_quantity} available`);
    }
  }
  return { byId, issues, loading: products.isLoading };
}

export function CartPage() {
  const { lines, subtotal, count, setQuantity, remove, clear } = useCart();
  const { user } = useAuth();
  const navigate = useNavigate();
  const { byId, issues } = useCartAvailability();

  if (lines.length === 0) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-16">
        <Card>
          <EmptyState
            icon={<ShoppingBag />}
            title="Your cart is empty"
            description="Browse the catalogue and add something you like."
            action={
              <Link to="/shop">
                <Button>Start shopping</Button>
              </Link>
            }
          />
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-10 sm:px-6">
      <div className="mb-6 flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Shopping cart</h1>
          <p className="mt-1 text-sm text-subtle">{count} {count === 1 ? "item" : "items"}</p>
        </div>
        <Button variant="ghost" size="sm" onClick={clear}>
          <Trash2 /> Clear cart
        </Button>
      </div>

      <div className="grid gap-8 lg:grid-cols-[1fr_360px]">
        <Card className="divide-y divide-border">
          {lines.map((line) => {
            const live = byId.get(line.productId);
            const issue = issues.get(line.productId);
            const price = live?.price ?? line.price;
            return (
              <div key={line.productId} className="flex gap-4 p-4 sm:p-5" data-testid="cart-line">
                <Link to={`/shop/products/${line.productId}`} className="shrink-0">
                  <ProductArt id={line.productId} name={line.name} size="sm" className="size-20 rounded-lg sm:size-24" />
                </Link>
                <div className="flex min-w-0 flex-1 flex-col">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <Link to={`/shop/products/${line.productId}`} className="font-medium hover:text-primary">
                        {line.name}
                      </Link>
                      <p className="text-xs text-faint">SKU {line.sku}</p>
                    </div>
                    <p className="tabular font-semibold">{money(price * line.quantity)}</p>
                  </div>
                  <p className="tabular mt-0.5 text-sm text-subtle">{money(price)} each</p>
                  {issue && (
                    <p className="mt-1 flex items-center gap-1 text-xs font-medium text-warning">
                      <AlertTriangle className="size-3.5" /> {issue}
                    </p>
                  )}
                  <div className="mt-auto flex items-center justify-between pt-3">
                    <QuantityStepper
                      value={line.quantity}
                      max={Math.max(1, live?.available_quantity ?? 99)}
                      onChange={(value) => setQuantity(line.productId, value)}
                    />
                    <Button variant="ghost" size="sm" onClick={() => remove(line.productId)} aria-label={`Remove ${line.name}`}>
                      <Trash2 /> Remove
                    </Button>
                  </div>
                </div>
              </div>
            );
          })}
        </Card>

        <Card className="h-fit p-5 lg:sticky lg:top-24">
          <h2 className="font-semibold">Order summary</h2>
          <dl className="mt-4 space-y-2 text-sm">
            <div className="flex justify-between">
              <dt className="text-subtle">Subtotal</dt>
              <dd className="tabular">{money(subtotal)}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-subtle">Shipping</dt>
              <dd className="text-success">Free</dd>
            </div>
            <div className="flex justify-between border-t border-border pt-3 text-base font-semibold">
              <dt>Total</dt>
              <dd className="tabular">{money(subtotal)}</dd>
            </div>
          </dl>
          <p className="mt-2 text-xs text-subtle">Final total is confirmed with current prices when you place the order.</p>
          <Button
            size="lg"
            className="mt-5 w-full"
            disabled={issues.size > 0}
            onClick={() => navigate(user ? "/shop/checkout" : "/login", { state: { from: "/shop/checkout" } })}
          >
            {user ? "Proceed to checkout" : "Sign in to check out"} <ArrowRight />
          </Button>
          {issues.size > 0 && (
            <p className="mt-2 text-xs text-warning">Fix the highlighted items to continue.</p>
          )}
          <Link to="/shop" className="mt-3 block text-center text-sm text-primary hover:underline">
            Continue shopping
          </Link>
        </Card>
      </div>
    </div>
  );
}
