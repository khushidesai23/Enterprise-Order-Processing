import { useRef, useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { ChevronLeft, ChevronRight, Heart, Star } from "lucide-react";
import { toast } from "sonner";
import { useCart } from "@/lib/cart";
import { money, number } from "@/lib/format";
import { toSaved, useWishlist, type SavedProduct } from "@/lib/saved";
import { cn } from "@/lib/utils";
import { ProductArt } from "./common";
import { Skeleton } from "./ui/primitives";

// --- Images -----------------------------------------------------------------

// Product photo on a white tile; falls back to generated art when a
// product has no image or it fails to load (for example offline).
export function ProductImage({
  id,
  name,
  src,
  className,
  imgClassName,
  size = "md",
}: {
  id: string;
  name: string;
  src?: string;
  className?: string;
  imgClassName?: string;
  size?: "sm" | "md" | "lg";
}) {
  const [failed, setFailed] = useState(false);

  if (!src || failed) return <ProductArt id={id} name={name} size={size} className={className} />;

  return (
    <div className={cn("flex items-center justify-center overflow-hidden bg-white", className)}>
      <img
        src={src}
        alt={name}
        loading="lazy"
        decoding="async"
        onError={() => setFailed(true)}
        className={cn("h-full w-full object-contain", imgClassName)}
      />
    </div>
  );
}

// --- Ratings ----------------------------------------------------------------

export function Stars({ value, size = 14, className }: { value: number; size?: number; className?: string }) {
  return (
    <span className={cn("inline-flex items-center", className)} aria-label={`${value.toFixed(1)} out of 5 stars`} role="img">
      {[1, 2, 3, 4, 5].map((star) => {
        const fill = Math.max(0, Math.min(1, value - star + 1));
        return (
          <span key={star} className="relative inline-block" style={{ width: size, height: size }}>
            <Star className="absolute inset-0 text-star/35" style={{ width: size, height: size }} strokeWidth={1.5} />
            <span className="absolute inset-0 overflow-hidden" style={{ width: `${fill * 100}%` }}>
              <Star className="fill-star text-star" style={{ width: size, height: size }} strokeWidth={1.5} />
            </span>
          </span>
        );
      })}
    </span>
  );
}

export function RatingLine({ average, count, href }: { average: number; count: number; href?: string }) {
  if (count === 0) return <span className="text-xs text-subtle">No reviews yet</span>;
  const content = (
    <>
      <span className="text-sm font-medium text-foreground">{average.toFixed(1)}</span>
      <Stars value={average} />
      <span className="text-sm text-link">({number(count)})</span>
    </>
  );
  return href ? (
    <Link to={href} className="inline-flex items-center gap-1.5 hover:underline">
      {content}
    </Link>
  ) : (
    <span className="inline-flex items-center gap-1.5">{content}</span>
  );
}

// --- Prices -----------------------------------------------------------------

export function discountPercent(price: number, mrp: number): number {
  return mrp > price ? Math.round(((mrp - price) / mrp) * 100) : 0;
}

export function PriceTag({ price, mrp, size = "md" }: { price: number; mrp: number; size?: "sm" | "md" | "lg" }) {
  const off = discountPercent(price, mrp);
  return (
    <div>
      <div className="flex flex-wrap items-baseline gap-x-2">
        {off > 0 && size === "lg" && <span className="text-2xl font-light text-deal">-{off}%</span>}
        <span
          className={cn(
            "tabular font-semibold text-foreground",
            size === "sm" && "text-base",
            size === "md" && "text-xl",
            size === "lg" && "text-3xl font-medium",
          )}
        >
          {money(price)}
        </span>
        {off > 0 && size !== "lg" && (
          <span className="text-xs text-subtle">
            M.R.P: <span className="tabular line-through">{money(mrp)}</span> <span className="text-deal">({off}% off)</span>
          </span>
        )}
      </div>
      {off > 0 && size === "lg" && (
        <p className="text-sm text-subtle">
          M.R.P.: <span className="tabular line-through">{money(mrp)}</span>
        </p>
      )}
    </div>
  );
}

export function DealBadge({ price, mrp }: { price: number; mrp: number }) {
  const off = discountPercent(price, mrp);
  if (off < 15) return null;
  return (
    <span className="inline-flex items-center gap-1.5">
      <span className="rounded-sm bg-deal px-1.5 py-0.5 text-[11px] font-semibold text-white">{off}% off</span>
      <span className="text-[11px] font-semibold text-deal">{off >= 35 ? "Great Deal" : "Limited time deal"}</span>
    </span>
  );
}

// Delivery promise shown on cards and the buy box.
export function deliveryDate(days = 3): string {
  const date = new Date();
  date.setDate(date.getDate() + days);
  return date.toLocaleDateString("en-IN", { weekday: "long", day: "numeric", month: "long" });
}

// --- Wishlist ---------------------------------------------------------------

export function WishlistButton({ product, className }: { product: SavedProduct; className?: string }) {
  const { has, toggle } = useWishlist();
  const saved = has(product.id);
  return (
    <button
      type="button"
      onClick={(event) => {
        event.preventDefault();
        event.stopPropagation();
        const added = toggle(toSaved(product));
        toast.success(added ? "Added to your wishlist" : "Removed from your wishlist");
      }}
      aria-label={saved ? `Remove ${product.name} from wishlist` : `Add ${product.name} to wishlist`}
      aria-pressed={saved}
      className={cn(
        "flex size-9 items-center justify-center rounded-full bg-white/90 text-subtle shadow-sm ring-1 ring-border transition-colors hover:text-deal",
        saved && "text-deal",
        className,
      )}
    >
      <Heart className={cn("size-4", saved && "fill-current")} />
    </button>
  );
}

// --- Cards ------------------------------------------------------------------

export interface CardProduct extends SavedProduct {
  available_quantity?: number;
  sku?: string;
}

export function ProductCard({ product, compact = false }: { product: CardProduct; compact?: boolean }) {
  const { add, lines } = useCart();
  const inCart = lines.find((line) => line.productId === product.id)?.quantity ?? 0;
  const stock = product.available_quantity;
  const soldOut = stock !== undefined && stock <= 0;
  const href = `/shop/products/${product.id}`;

  return (
    <article className="group relative flex h-full flex-col rounded-lg border border-border bg-surface p-3 transition-shadow hover:shadow-pop" data-testid="product-card">
      <WishlistButton product={product} className="absolute right-2 top-2 z-10 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100 aria-pressed:opacity-100" />
      <Link to={href} className="block">
        <ProductImage
          id={product.id}
          name={product.name}
          src={product.image_url}
          className={cn("rounded-md", compact ? "aspect-square" : "aspect-[4/4]")}
          imgClassName="transition-transform duration-300 group-hover:scale-[1.03]"
        />
      </Link>
      <div className="mt-3 flex flex-1 flex-col gap-1">
        {product.brand && !compact && <p className="text-xs font-medium uppercase tracking-wide text-subtle">{product.brand}</p>}
        <Link to={href} className={cn("text-sm leading-snug text-foreground hover:text-link", compact ? "line-clamp-2" : "line-clamp-2")}>
          {product.name}
        </Link>
        <RatingLine average={product.rating_average} count={product.rating_count} />
        <DealBadge price={product.price} mrp={product.mrp} />
        <PriceTag price={product.price} mrp={product.mrp} size={compact ? "sm" : "md"} />
        {!compact && (
          <p className="text-xs text-subtle">
            {soldOut ? (
              <span className="font-medium text-danger">Currently unavailable</span>
            ) : stock !== undefined && stock <= 10 ? (
              <span className="font-medium text-deal">Only {stock} left in stock</span>
            ) : (
              <>
                FREE delivery <span className="font-semibold text-foreground">{deliveryDate(product.price >= 499 ? 3 : 5)}</span>
              </>
            )}
          </p>
        )}
        {!compact && (
          <button
            type="button"
            disabled={soldOut || (stock !== undefined && inCart >= stock)}
            onClick={() => {
              add({
                productId: product.id,
                name: product.name,
                sku: product.sku ?? "",
                price: product.price,
                category: product.category,
                imageUrl: product.image_url,
              });
              toast.success("Added to cart", { description: product.name });
            }}
            className="mt-auto w-full rounded-full bg-cart py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-cart-hover disabled:cursor-not-allowed disabled:opacity-50"
            aria-label={`Add ${product.name} to cart`}
          >
            {soldOut ? "Unavailable" : inCart ? `Add to cart (${inCart} in cart)` : "Add to cart"}
          </button>
        )}
      </div>
    </article>
  );
}

// Colour and size variants share their photos; showcases keep one of each.
export function distinctProducts<T extends { id: string; image_url: string }>(items: T[] | undefined, max = Infinity): T[] | undefined {
  if (!items) return undefined;
  const seen = new Set<string>();
  const result: T[] = [];
  for (const item of items) {
    const key = item.image_url || item.id;
    if (seen.has(key)) continue;
    seen.add(key);
    result.push(item);
    if (result.length >= max) break;
  }
  return result;
}

// --- Rails ------------------------------------------------------------------

// A titled horizontal carousel of product cards.
export function ProductRail({
  title,
  action,
  products,
  loading,
  compact = true,
}: {
  title: ReactNode;
  action?: ReactNode;
  products: CardProduct[] | undefined;
  loading?: boolean;
  compact?: boolean;
}) {
  const scroller = useRef<HTMLDivElement>(null);
  const scroll = (direction: number) =>
    scroller.current?.scrollBy({ left: direction * scroller.current.clientWidth * 0.85, behavior: "smooth" });

  if (!loading && (!products || products.length === 0)) return null;

  return (
    <section className="rounded-lg bg-surface p-4 shadow-card">
      <div className="mb-3 flex items-baseline gap-4">
        <h2 className="text-xl font-semibold tracking-tight">{title}</h2>
        {action}
      </div>
      <div className="group/rail relative">
        <div ref={scroller} className="no-scrollbar flex snap-x gap-3 overflow-x-auto scroll-smooth">
          {loading
            ? Array.from({ length: 6 }).map((_, index) => <Skeleton key={index} className="h-72 w-44 shrink-0 rounded-lg" />)
            : products!.map((product) => (
                <div key={product.id} className={cn("shrink-0 snap-start", compact ? "w-44 sm:w-48" : "w-56")}>
                  <ProductCard product={product} compact={compact} />
                </div>
              ))}
        </div>
        {!loading && (
          <>
            <button
              type="button"
              onClick={() => scroll(-1)}
              aria-label="Scroll left"
              className="absolute left-0 top-1/3 hidden h-20 w-10 items-center justify-center rounded-r-md bg-white/95 shadow-pop ring-1 ring-border group-hover/rail:flex"
            >
              <ChevronLeft className="size-6" />
            </button>
            <button
              type="button"
              onClick={() => scroll(1)}
              aria-label="Scroll right"
              className="absolute right-0 top-1/3 hidden h-20 w-10 items-center justify-center rounded-l-md bg-white/95 shadow-pop ring-1 ring-border group-hover/rail:flex"
            >
              <ChevronRight className="size-6" />
            </button>
          </>
        )}
      </div>
    </section>
  );
}
