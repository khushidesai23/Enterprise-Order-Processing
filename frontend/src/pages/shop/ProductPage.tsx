import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { ChevronLeft, PackageCheck, ShieldCheck, ShoppingCart, Truck } from "lucide-react";
import { productsApi } from "@/lib/endpoints";
import { ApiError, errorMessage } from "@/lib/api";
import { useCart } from "@/lib/cart";
import { money } from "@/lib/format";
import { Button } from "@/components/ui/button";
import { EmptyState, ErrorState, Skeleton } from "@/components/ui/primitives";
import { StockBadge } from "@/components/ui/badge";
import { ProductArt, QuantityStepper } from "@/components/common";

export function ProductPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const { add, lines } = useCart();
  const [quantity, setQuantity] = useState(1);
  const product = useQuery({ queryKey: ["product", id], queryFn: () => productsApi.get(id) });

  if (product.isLoading) {
    return (
      <div className="mx-auto grid max-w-6xl gap-10 px-4 py-10 sm:px-6 lg:grid-cols-2">
        <Skeleton className="aspect-square rounded-2xl" />
        <div className="space-y-4">
          <Skeleton className="h-6 w-32" />
          <Skeleton className="h-10 w-3/4" />
          <Skeleton className="h-24" />
        </div>
      </div>
    );
  }

  if (product.isError || !product.data) {
    const notFound = product.error instanceof ApiError && (product.error.status === 404 || product.error.status === 400);
    return (
      <div className="mx-auto max-w-3xl px-4 py-16">
        {notFound ? (
          <EmptyState
            title="Product not found"
            description="It may have been removed from the catalogue."
            action={
              <Link to="/shop">
                <Button variant="secondary">Back to shop</Button>
              </Link>
            }
          />
        ) : (
          <ErrorState message={errorMessage(product.error)} onRetry={() => product.refetch()} />
        )}
      </div>
    );
  }

  const item = product.data;
  const available = item.inventory?.available_quantity ?? 0;
  const inCart = lines.find((line) => line.productId === item.id)?.quantity ?? 0;
  const maxAddable = Math.max(0, available - inCart);
  const qty = Math.min(quantity, Math.max(1, maxAddable));

  function addToCart(goToCart: boolean) {
    add(
      { productId: item.id, name: item.name, sku: item.sku, price: item.price, category: item.category?.name ?? "" },
      qty,
    );
    setQuantity(1);
    if (goToCart) navigate("/shop/cart");
    else toast.success(`Added ${qty} × ${item.name} to cart`);
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
      <Link to="/shop" className="mb-6 inline-flex items-center gap-1 text-sm text-subtle hover:text-foreground">
        <ChevronLeft className="size-4" /> Back to shop
      </Link>
      <div className="grid gap-10 lg:grid-cols-2">
        <ProductArt id={item.id} name={item.name} size="lg" className="aspect-square rounded-2xl border border-border" />
        <div className="flex flex-col">
          <Link
            to={`/shop?category=${item.category?.id}`}
            className="text-sm font-medium uppercase tracking-wide text-primary hover:underline"
          >
            {item.category?.name}
          </Link>
          <h1 className="mt-2 text-3xl font-semibold tracking-tight">{item.name}</h1>
          <p className="mt-1 text-sm text-faint">SKU {item.sku}</p>
          <p className="tabular mt-5 text-3xl font-semibold">{money(item.price)}</p>
          <p className="text-xs text-subtle">Inclusive of all taxes</p>
          <div className="mt-4">
            <StockBadge available={available} threshold={10} />
          </div>

          <p className="mt-6 whitespace-pre-line leading-relaxed text-subtle">
            {item.description || "No description has been added for this product yet."}
          </p>

          <div className="mt-8 flex flex-wrap items-center gap-3">
            <QuantityStepper value={qty} max={Math.max(1, maxAddable)} onChange={setQuantity} />
            <Button size="lg" disabled={maxAddable <= 0} onClick={() => addToCart(false)}>
              <ShoppingCart /> Add to cart
            </Button>
            <Button size="lg" variant="secondary" disabled={maxAddable <= 0} onClick={() => addToCart(true)}>
              Buy now
            </Button>
          </div>
          {inCart > 0 && (
            <p className="mt-3 text-sm text-subtle">
              {inCart} already in your{" "}
              <Link to="/shop/cart" className="text-primary hover:underline">
                cart
              </Link>
              {maxAddable <= 0 && available > 0 ? " — that's all the stock we have." : "."}
            </p>
          )}

          <div className="mt-10 grid gap-4 border-t border-border pt-6 sm:grid-cols-3">
            {[
              { icon: ShieldCheck, title: "Secure checkout", text: "Razorpay protected" },
              { icon: PackageCheck, title: "Reserved for you", text: "Stock held at order" },
              { icon: Truck, title: "Tracked delivery", text: "Live order status" },
            ].map(({ icon: Icon, title, text }) => (
              <div key={title} className="flex items-start gap-3">
                <Icon className="mt-0.5 size-5 text-primary" />
                <div>
                  <p className="text-sm font-medium">{title}</p>
                  <p className="text-xs text-subtle">{text}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
