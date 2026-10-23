import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { Lock, MapPin, RotateCcw, ShieldCheck, Truck } from "lucide-react";
import { productsApi } from "@/lib/endpoints";
import { ApiError, errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { MAX_LINE_QUANTITY, useCart } from "@/lib/cart";
import { rememberViewed } from "@/lib/saved";
import type { Product } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { EmptyState, ErrorState, Skeleton } from "@/components/ui/primitives";
import { DealBadge, PriceTag, ProductImage, ProductRail, RatingLine, WishlistButton, deliveryDate, distinctProducts } from "@/components/shop";
import { ReviewsSection } from "./ReviewsSection";

function Gallery({ product }: { product: Product }) {
  const images = product.images.length > 0 ? product.images : product.image_url ? [product.image_url] : [];
  const [active, setActive] = useState(0);
  const current = images[active] ?? product.image_url;

  return (
    <div className="flex gap-3 lg:sticky lg:top-32">
      {images.length > 1 && (
        <div className="flex flex-col gap-2" role="tablist" aria-label="Product images">
          {images.map((image, index) => (
            <button
              key={image}
              role="tab"
              aria-selected={index === active}
              aria-label={`Image ${index + 1}`}
              onMouseEnter={() => setActive(index)}
              onClick={() => setActive(index)}
              className={cn(
                "size-14 overflow-hidden rounded-md border-2 bg-white p-0.5",
                index === active ? "border-amber-500 shadow" : "border-border hover:border-amber-300",
              )}
            >
              <img src={image} alt="" className="h-full w-full object-contain" />
            </button>
          ))}
        </div>
      )}
      <div className="relative flex-1">
        <ProductImage id={product.id} name={product.name} src={current} size="lg" className="aspect-square rounded-lg" />
        <WishlistButton
          product={{ ...product, category: product.category?.name ?? "" }}
          className="absolute right-3 top-3"
        />
      </div>
    </div>
  );
}

export function ProductPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { add, lines } = useCart();
  const [quantity, setQuantity] = useState(1);
  const product = useQuery({ queryKey: ["product", id], queryFn: () => productsApi.get(id) });
  const related = useQuery({
    queryKey: ["search", "related", product.data?.category.id],
    queryFn: () => productsApi.search({ category_id: product.data!.category.id, sort: "bestselling", limit: 60 }),
    enabled: !!product.data,
    staleTime: 60_000,
  });
  const brandMore = useQuery({
    queryKey: ["search", "brand", product.data?.brand],
    queryFn: () => productsApi.search({ brand: [product.data!.brand], sort: "reviews", limit: 60 }),
    enabled: !!product.data?.brand,
    staleTime: 60_000,
  });

  useEffect(() => {
    setQuantity(1);
  }, [id]);

  useEffect(() => {
    const item = product.data;
    if (item) {
      rememberViewed({
        id: item.id,
        name: item.name,
        brand: item.brand,
        price: item.price,
        mrp: item.mrp,
        image_url: item.image_url,
        rating_average: item.rating_average,
        rating_count: item.rating_count,
        category: item.category?.name ?? "",
      });
    }
  }, [product.data]);

  if (product.isLoading) {
    return (
      <div className="mx-auto grid max-w-[1500px] gap-8 bg-surface px-4 py-8 lg:grid-cols-[1.1fr_1.2fr_300px]">
        <Skeleton className="aspect-square" />
        <div className="space-y-4">
          <Skeleton className="h-8 w-3/4" />
          <Skeleton className="h-5 w-40" />
          <Skeleton className="h-32" />
        </div>
        <Skeleton className="h-80" />
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
                <Button variant="secondary">Back to the store</Button>
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
  const maxAddable = Math.max(0, Math.min(available, MAX_LINE_QUANTITY) - inCart);
  const qty = Math.min(quantity, Math.max(1, maxAddable));
  const highlights = item.specs.filter((spec) => !["Model number", "Delivery", "Returns", "Warranty"].includes(spec.name)).slice(0, 6);
  const warranty = item.specs.find((spec) => spec.name === "Warranty")?.value;
  const returns = item.specs.find((spec) => spec.name === "Returns")?.value;

  function addToCart(buyNow: boolean) {
    add(
      {
        productId: item.id,
        name: item.name,
        sku: item.sku,
        price: item.price,
        category: item.category?.name ?? "",
        imageUrl: item.image_url,
      },
      qty,
    );
    setQuantity(1);
    if (buyNow) navigate(user ? "/shop/checkout" : "/login", { state: { from: "/shop/checkout" } });
    else toast.success(`Added ${qty} × ${item.name} to cart`);
  }

  return (
    <div className="bg-surface">
      <nav className="mx-auto max-w-[1500px] px-4 pt-3 text-xs text-subtle" aria-label="Breadcrumb">
        <Link to="/shop" className="hover:underline">Home</Link>
        <span className="mx-1.5">›</span>
        <Link to={`/shop/search?category_id=${item.category?.id}`} className="hover:underline">
          {item.category?.name}
        </Link>
        {item.brand && (
          <>
            <span className="mx-1.5">›</span>
            <Link to={`/shop/search?brand=${encodeURIComponent(item.brand)}`} className="hover:underline">
              {item.brand}
            </Link>
          </>
        )}
      </nav>

      <div className="mx-auto grid max-w-[1500px] items-start gap-8 px-4 py-5 lg:grid-cols-[1.1fr_1.2fr_300px]">
        <Gallery key={item.id} product={item} />

        <div className="min-w-0">
          <h1 className="text-2xl font-medium leading-snug">{item.name}</h1>
          {item.brand && (
            <Link to={`/shop/search?brand=${encodeURIComponent(item.brand)}`} className="mt-1 inline-block text-sm text-link hover:text-deal hover:underline">
              Visit the {item.brand} Store
            </Link>
          )}
          <div className="mt-1.5">
            <RatingLine average={item.rating_average} count={item.rating_count} href="#reviews" />
          </div>
          <hr className="my-3" />
          <DealBadge price={item.price} mrp={item.mrp} />
          <div className="mt-2">
            <PriceTag price={item.price} mrp={item.mrp} size="lg" />
          </div>
          <p className="text-sm">Inclusive of all taxes</p>

          <div className="mt-4 grid grid-cols-3 gap-2 border-y border-border py-4 text-center text-xs text-link">
            <div className="flex flex-col items-center gap-1">
              <Truck className="size-6 text-subtle" /> Free Delivery
            </div>
            <div className="flex flex-col items-center gap-1">
              <RotateCcw className="size-6 text-subtle" /> {returns ?? "Easy returns"}
            </div>
            <div className="flex flex-col items-center gap-1">
              <ShieldCheck className="size-6 text-subtle" /> {warranty ?? "Secure transaction"}
            </div>
          </div>

          {highlights.length > 0 && (
            <table className="mt-4 text-sm">
              <tbody>
                {highlights.map((spec) => (
                  <tr key={spec.name}>
                    <th className="w-36 py-1 pr-4 text-left align-top font-semibold">{spec.name}</th>
                    <td className="py-1">{spec.value}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}

          <hr className="my-4" />
          <h2 className="mb-2 text-base font-bold">About this item</h2>
          <ul className="list-disc space-y-1.5 pl-5 text-sm leading-relaxed">
            {(item.description || "No description has been added for this product yet.")
              .split(/(?<=\.)\s+/)
              .filter(Boolean)
              .map((sentence, index) => (
                <li key={index}>{sentence}</li>
              ))}
          </ul>
        </div>

        <aside className="rounded-lg border border-border p-4" aria-label="Buy box">
          <PriceTag price={item.price} mrp={0} size="md" />
          <p className="mt-2 text-sm">
            <span className="text-link">FREE delivery</span>{" "}
            <span className="font-semibold">{deliveryDate(item.price >= 499 ? 3 : 5)}</span>
          </p>
          <p className="mt-1 flex items-center gap-1 text-xs text-link">
            <MapPin className="size-3.5" /> Deliver to {user ? user.first_name : "you"} – India
          </p>

          <p className={cn("mt-3 text-lg font-medium", available > 0 ? "text-success" : "text-danger")}>
            {available <= 0 ? "Currently unavailable" : available <= 10 ? `Only ${available} left in stock.` : "In stock"}
          </p>

          {available > 0 && (
            <>
              <label className="mt-3 flex items-center gap-2 text-sm">
                Quantity:
                <select
                  value={qty}
                  onChange={(event) => setQuantity(Number(event.target.value))}
                  disabled={maxAddable <= 0}
                  className="cursor-pointer rounded-lg border border-border bg-muted px-2 py-1 shadow-sm"
                  aria-label="Quantity"
                >
                  {Array.from({ length: Math.max(1, Math.min(maxAddable, 10)) }, (_, i) => i + 1).map((n) => (
                    <option key={n} value={n}>
                      {n}
                    </option>
                  ))}
                </select>
              </label>
              <button
                onClick={() => addToCart(false)}
                disabled={maxAddable <= 0}
                className="mt-4 w-full rounded-full bg-cart py-2 text-sm font-medium hover:bg-cart-hover disabled:opacity-50"
              >
                Add to Cart
              </button>
              <button
                onClick={() => addToCart(true)}
                disabled={maxAddable <= 0}
                className="mt-2 w-full rounded-full bg-buy py-2 text-sm font-medium hover:bg-buy-hover disabled:opacity-50"
              >
                Buy Now
              </button>
              {inCart > 0 && (
                <p className="mt-2 text-xs text-subtle">
                  {inCart} in your{" "}
                  <Link to="/shop/cart" className="text-link hover:underline">
                    cart
                  </Link>
                  {maxAddable <= 0 ? " — that's all the stock we have." : "."}
                </p>
              )}
            </>
          )}

          <dl className="mt-4 grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-xs">
            <dt className="text-subtle">Ships from</dt>
            <dd>Meridian Market</dd>
            <dt className="text-subtle">Sold by</dt>
            <dd>{item.brand || "Meridian Market"}</dd>
            <dt className="text-subtle">Payment</dt>
            <dd className="flex items-center gap-1">
              <Lock className="size-3" /> Secure transaction
            </dd>
          </dl>
        </aside>
      </div>

      <div className="mx-auto max-w-[1500px] space-y-6 px-4 pb-8">
        <ProductRail
          title={`More in ${item.category?.name ?? "this department"}`}
          products={distinctProducts(related.data?.items.filter((other) => other.image_url !== item.image_url), 18)}
          loading={related.isLoading}
        />
        {item.brand && (
          <ProductRail
            title={`More from ${item.brand}`}
            products={distinctProducts(brandMore.data?.items.filter((other) => other.image_url !== item.image_url), 18)}
            loading={brandMore.isLoading}
          />
        )}

        {item.specs.length > 0 && (
          <section className="rounded-lg border border-border">
            <h2 className="border-b border-border px-5 py-3 text-lg font-bold">Product information</h2>
            <table className="w-full text-sm">
              <tbody>
                {item.specs.map((spec) => (
                  <tr key={spec.name} className="border-b border-border last:border-0">
                    <th className="w-1/3 bg-muted/60 px-5 py-2 text-left font-semibold">{spec.name}</th>
                    <td className="px-5 py-2">{spec.value}</td>
                  </tr>
                ))}
                <tr>
                  <th className="w-1/3 bg-muted/60 px-5 py-2 text-left font-semibold">SKU</th>
                  <td className="px-5 py-2">{item.sku}</td>
                </tr>
              </tbody>
            </table>
          </section>
        )}

        <ReviewsSection product={item} />
      </div>
    </div>
  );
}
