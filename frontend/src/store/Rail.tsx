import { useRef, type ReactNode } from "react";
import { ArrowLeft, ArrowRight } from "lucide-react";
import { Eyebrow, Skel } from "./ui";
import { ProductTile, type TileProduct } from "./ProductTile";

// A titled, swipeable row of product tiles.
export function Rail({
  eyebrow,
  title,
  products,
  loading,
  action,
}: {
  eyebrow?: string;
  title: ReactNode;
  products: TileProduct[];
  loading?: boolean;
  action?: ReactNode;
}) {
  const scroller = useRef<HTMLDivElement>(null);
  const scroll = (direction: number) =>
    scroller.current?.scrollBy({ left: direction * scroller.current.clientWidth * 0.8, behavior: "smooth" });

  if (!loading && products.length === 0) return null;

  return (
    <section className="py-12">
      <div className="mb-6 flex items-end justify-between gap-4">
        <div>
          {eyebrow && <Eyebrow>{eyebrow}</Eyebrow>}
          <h2 className="mt-2 font-display text-3xl font-extrabold tracking-[-0.04em] sm:text-4xl">{title}</h2>
        </div>
        <div className="flex items-center gap-2">
          {action}
          <button onClick={() => scroll(-1)} aria-label="Scroll back" className="hidden size-10 items-center justify-center rounded-full border border-ink/15 hover:bg-ink hover:text-bone sm:flex">
            <ArrowLeft className="size-4" />
          </button>
          <button onClick={() => scroll(1)} aria-label="Scroll forward" className="hidden size-10 items-center justify-center rounded-full border border-ink/15 hover:bg-ink hover:text-bone sm:flex">
            <ArrowRight className="size-4" />
          </button>
        </div>
      </div>
      <div ref={scroller} className="no-scrollbar -mx-5 flex snap-x gap-5 overflow-x-auto px-5 pb-2 sm:-mx-8 sm:px-8">
        {loading
          ? Array.from({ length: 5 }).map((_, i) => <Skel key={i} className="aspect-[3/4] w-56 shrink-0" />)
          : products.map((product, index) => (
              <div key={product.id} className="w-56 shrink-0 snap-start sm:w-64">
                <ProductTile product={product} index={index} />
              </div>
            ))}
      </div>
    </section>
  );
}
