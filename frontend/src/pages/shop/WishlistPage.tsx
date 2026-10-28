import { useRef } from "react";
import { Link } from "react-router-dom";
import { AnimatePresence, motion } from "motion/react";
import { toast } from "sonner";
import { ArrowRight, X } from "lucide-react";
import { useWishlist, type SavedProduct } from "@/lib/saved";
import { useBag } from "@/store/bag";
import { Eyebrow, MaskText, Price, ProductShot, Rating, ease, pill } from "@/store/ui";

function WishCard({ item }: { item: SavedProduct }) {
  const { remove } = useWishlist();
  const bag = useBag();
  const plate = useRef<HTMLDivElement>(null);

  return (
    <motion.li
      layout
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.9 }}
      transition={{ duration: 0.4, ease }}
      className="flex flex-col"
      data-testid="wishlist-item"
    >
      <div ref={plate} className="relative overflow-hidden rounded-[1.75rem] bg-bone-2">
        <Link to={`/shop/products/${item.id}`} className="block">
          <ProductShot id={item.id} name={item.name} src={item.image_url} className="aspect-square p-6" />
        </Link>
        <button
          onClick={() => remove(item.id)}
          aria-label={`Remove ${item.name} from wishlist`}
          className="absolute right-3 top-3 flex size-9 items-center justify-center rounded-full bg-bone/90 hover:bg-ink hover:text-bone"
        >
          <X className="size-4" />
        </button>
      </div>
      <p className="mt-3 font-mono text-[10px] uppercase tracking-[0.16em] text-smoke">{item.brand || item.category}</p>
      <Link to={`/shop/products/${item.id}`} className="mt-1 line-clamp-2 font-medium hover:underline">
        {item.name}
      </Link>
      <div className="mt-2 flex items-center justify-between gap-2">
        <Price price={item.price} mrp={item.mrp} size="sm" />
        <Rating average={item.rating_average} count={item.rating_count} />
      </div>
      <button
        onClick={() => {
          bag.add({ productId: item.id, name: item.name, sku: "", price: item.price, category: item.category, imageUrl: item.image_url }, 1, plate.current);
          remove(item.id);
          toast("Moved to bag", { description: item.name });
        }}
        className={`${pill("ink", "sm")} mt-3`}
      >
        Move to bag
      </button>
    </motion.li>
  );
}

export function WishlistPage() {
  const { items } = useWishlist();

  return (
    <div className="mx-auto max-w-[1600px] px-5 pt-6 sm:px-8">
      <Eyebrow>Saved on this device · {items.length}</Eyebrow>
      <h1 className="mt-3 font-display text-[clamp(3rem,8vw,7rem)] font-extrabold leading-[0.88] tracking-[-0.055em]">
        <MaskText text="Wish" /> <MaskText text="list." delay={0.1} className="font-serif font-normal italic tracking-[-0.02em] text-ultra" />
      </h1>

      {items.length === 0 ? (
        <div className="mt-12 flex flex-col items-start gap-5 rounded-[2.5rem] bg-bone-2 p-10">
          <p className="font-display text-4xl font-bold tracking-tight">
            Nothing saved <span className="font-serif font-normal italic">yet.</span>
          </p>
          <p className="max-w-md text-smoke">Tap the heart on any object to keep it here.</p>
          <Link to="/shop/search" className={pill("ink")}>
            Browse the index <ArrowRight className="size-4" />
          </Link>
        </div>
      ) : (
        <ul className="mt-10 grid grid-cols-2 gap-x-6 gap-y-10 md:grid-cols-3 xl:grid-cols-5">
          <AnimatePresence initial={false}>
            {items.map((item) => (
              <WishCard key={item.id} item={item} />
            ))}
          </AnimatePresence>
        </ul>
      )}
    </div>
  );
}
