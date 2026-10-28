import { useRef } from "react";
import { Link } from "react-router-dom";
import { motion } from "motion/react";
import { toast } from "sonner";
import { Heart, Plus } from "lucide-react";
import { useCart } from "@/lib/cart";
import { toSaved, useWishlist, type SavedProduct } from "@/lib/saved";
import { cn } from "@/lib/utils";
import { useBag } from "./bag";
import { Price, ProductShot, Rating, Sticker, discountOf, ease } from "./ui";

export interface TileProduct extends SavedProduct {
  available_quantity?: number;
  sku?: string;
}

export function WishToggle({ product, className }: { product: SavedProduct; className?: string }) {
  const { has, toggle } = useWishlist();
  const saved = has(product.id);
  return (
    <button
      type="button"
      onClick={(event) => {
        event.preventDefault();
        event.stopPropagation();
        const added = toggle(toSaved(product));
        toast(added ? "Added to your wishlist" : "Removed from your wishlist");
      }}
      aria-label={saved ? `Remove ${product.name} from wishlist` : `Add ${product.name} to wishlist`}
      aria-pressed={saved}
      className={cn(
        "flex size-9 items-center justify-center rounded-full bg-bone/90 text-ink backdrop-blur transition-transform hover:scale-110",
        className,
      )}
    >
      <Heart className={cn("size-4", saved && "fill-signal text-signal")} />
    </button>
  );
}

export function ProductTile({
  product,
  index,
  size = "md",
  className,
}: {
  product: TileProduct;
  index?: number;
  size?: "md" | "lg";
  className?: string;
}) {
  const plate = useRef<HTMLDivElement>(null);
  const { lines } = useCart();
  const bag = useBag();
  const href = `/shop/products/${product.id}`;
  const stock = product.available_quantity;
  const soldOut = stock !== undefined && stock <= 0;
  const inBag = lines.find((line) => line.productId === product.id)?.quantity ?? 0;
  const atLimit = stock !== undefined && inBag >= stock;
  const off = discountOf(product.price, product.mrp);

  return (
    <motion.article
      className={cn("group relative flex flex-col", className)}
      data-testid="product-card"
      initial={{ opacity: 0, y: 24 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-5%" }}
      transition={{ duration: 0.6, ease, delay: Math.min((index ?? 0) % 4, 3) * 0.06 }}
    >
      <div ref={plate} className="relative overflow-hidden rounded-[1.75rem] bg-bone-2">
        <Link to={href} className="block" tabIndex={-1} aria-hidden>
          <ProductShot
            id={product.id}
            name={product.name}
            src={product.image_url}
            size={size === "lg" ? "lg" : "md"}
            className={cn("w-full p-6", size === "lg" ? "aspect-[4/5] p-10" : "aspect-square")}
            imgClassName="transition-transform duration-700 ease-[cubic-bezier(.22,1,.36,1)] group-hover:scale-[1.07] group-hover:-rotate-2"
          />
        </Link>

        {index !== undefined && (
          <span className="absolute left-4 top-4 font-mono text-[11px] tracking-widest text-smoke">
            {String(index + 1).padStart(3, "0")}
          </span>
        )}
        {off >= 15 && <Sticker className="absolute -right-1 bottom-4 size-14 text-xs">-{off}%</Sticker>}
        <WishToggle product={product} className="absolute right-3 top-3 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100 aria-pressed:opacity-100" />

        {soldOut ? (
          <span className="absolute bottom-4 left-4 rounded-full bg-ink px-3 py-1 font-mono text-[11px] uppercase tracking-wider text-bone">Sold out</span>
        ) : (
          <button
            type="button"
            disabled={atLimit}
            onClick={() => {
              bag.add(
                { productId: product.id, name: product.name, sku: product.sku ?? "", price: product.price, category: product.category, imageUrl: product.image_url },
                1,
                plate.current,
              );
              toast("Added to bag", { description: product.name });
            }}
            aria-label={`Add ${product.name} to bag`}
            className="absolute bottom-4 left-4 flex h-10 items-center gap-1.5 rounded-full bg-ink pl-3 pr-4 text-sm font-medium text-bone opacity-100 transition-all duration-300 hover:bg-ultra disabled:opacity-40 sm:translate-y-3 sm:opacity-0 sm:group-hover:translate-y-0 sm:group-hover:opacity-100 sm:focus-visible:translate-y-0 sm:focus-visible:opacity-100"
          >
            <Plus className="size-4" /> {inBag ? `In bag (${inBag})` : "Add"}
          </button>
        )}
      </div>

      <div className="mt-3 flex flex-1 flex-col gap-1 px-1">
        <p className="font-mono text-[10px] uppercase tracking-[0.16em] text-smoke">
          {product.brand || "meridian/"} · {product.category}
        </p>
        <Link to={href} className={cn("line-clamp-2 font-medium leading-snug hover:underline", size === "lg" ? "text-lg" : "text-[15px]")}>
          {product.name}
        </Link>
        <div className="mt-auto flex items-end justify-between gap-2 pt-1">
          <Price price={product.price} mrp={product.mrp} size="sm" />
          <Rating average={product.rating_average} count={product.rating_count} />
        </div>
        {stock !== undefined && stock > 0 && stock <= 10 && (
          <p className="font-mono text-[11px] uppercase tracking-wider text-signal">Only {stock} left</p>
        )}
      </div>
    </motion.article>
  );
}
