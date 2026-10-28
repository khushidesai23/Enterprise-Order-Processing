import { useEffect, useRef, useState, type MouseEvent, type ReactNode } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { AnimatePresence, motion, useReducedMotion } from "motion/react";
import { toast } from "sonner";
import { ArrowRight, Plus, RotateCcw, ShieldCheck, Truck } from "lucide-react";
import { productsApi } from "@/lib/endpoints";
import { ApiError, errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { MAX_LINE_QUANTITY, useCart } from "@/lib/cart";
import { rememberViewed } from "@/lib/saved";
import type { Product } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Stepper, useBag } from "@/store/bag";
import { WishToggle } from "@/store/ProductTile";
import { Rail } from "@/store/Rail";
import { Eyebrow, Price, ProductShot, Skel, StarRow, StockMeter, Sticker, baseName, discountOf, distinct, ease, pill, variantLabel } from "@/store/ui";
import { ReviewsSection } from "./ReviewsSection";

function deliveryDate(days: number) {
  const date = new Date();
  date.setDate(date.getDate() + days);
  return date.toLocaleDateString("en-IN", { weekday: "short", day: "numeric", month: "short" });
}

// Main image with crossfade between gallery images and a pointer zoom.
function Gallery({ product }: { product: Product }) {
  const images = product.images.length > 0 ? product.images : product.image_url ? [product.image_url] : [];
  const [active, setActive] = useState(0);
  const [zoom, setZoom] = useState<{ x: number; y: number } | null>(null);
  const reduce = useReducedMotion();
  const off = discountOf(product.price, product.mrp);
  const current = images[active] ?? product.image_url;

  function onMove(event: MouseEvent<HTMLDivElement>) {
    if (reduce) return;
    const rect = event.currentTarget.getBoundingClientRect();
    setZoom({ x: ((event.clientX - rect.left) / rect.width) * 100, y: ((event.clientY - rect.top) / rect.height) * 100 });
  }

  return (
    <div className="lg:sticky lg:top-24">
      <div
        className="relative aspect-square cursor-zoom-in overflow-hidden rounded-[2.5rem] bg-bone-2"
        onMouseMove={onMove}
        onMouseLeave={() => setZoom(null)}
      >
        <AnimatePresence mode="wait" initial={false}>
          <motion.div
            key={current}
            className="absolute inset-0 p-10 sm:p-16"
            initial={{ opacity: 0, scale: 0.96 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 1.02 }}
            transition={{ duration: 0.45, ease }}
          >
            <div
              className="size-full transition-transform duration-200 ease-out"
              style={zoom ? { transform: "scale(1.8)", transformOrigin: `${zoom.x}% ${zoom.y}%` } : undefined}
            >
              <ProductShot id={product.id} name={product.name} src={current} size="lg" eager className="size-full" />
            </div>
          </motion.div>
        </AnimatePresence>
        {off >= 5 && <Sticker className="absolute left-6 top-6 size-20 text-base">-{off}%</Sticker>}
        <WishToggle product={{ ...product, category: product.category?.name ?? "" }} className="absolute right-5 top-5 size-11" />
        {images.length > 1 && (
          <span className="absolute bottom-5 right-6 font-mono text-[11px] tracking-widest text-smoke">
            {String(active + 1).padStart(2, "0")} / {String(images.length).padStart(2, "0")}
          </span>
        )}
      </div>
      {images.length > 1 && (
        <div className="mt-3 flex gap-2" role="tablist" aria-label="Product images">
          {images.map((image, index) => (
            <button
              key={image}
              role="tab"
              aria-selected={index === active}
              aria-label={`Image ${index + 1}`}
              onClick={() => setActive(index)}
              onMouseEnter={() => setActive(index)}
              className={cn(
                "size-20 overflow-hidden rounded-2xl bg-bone-2 p-2 ring-2 ring-offset-2 ring-offset-bone transition-all",
                index === active ? "ring-ink" : "ring-transparent opacity-60 hover:opacity-100",
              )}
            >
              <img src={image} alt="" className="size-full object-contain" />
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

// Sibling listings share the base name: "Realme C35 (256 GB, Glacier Blue)".
function Variants({ product }: { product: Product }) {
  const base = baseName(product.name);
  const siblings = useQuery({
    queryKey: ["search", "variants", base, product.brand],
    queryFn: () => productsApi.search({ q: base, brand: product.brand ? [product.brand] : undefined, sort: "price_asc", limit: 30 }),
    staleTime: 60_000,
  });
  const options = (siblings.data?.items ?? []).filter((item) => baseName(item.name) === base);
  if (options.length < 2) return null;

  return (
    <div>
      <p className="mb-2 font-mono text-[11px] uppercase tracking-[0.18em] text-smoke">Options · {options.length}</p>
      <div className="flex flex-wrap gap-2">
        {options.map((option) => {
          const current = option.id === product.id;
          const soldOut = option.available_quantity <= 0;
          return (
            <Link
              key={option.id}
              to={`/shop/products/${option.id}`}
              replace
              aria-current={current ? "true" : undefined}
              className={cn(
                "rounded-2xl border px-3.5 py-2 text-sm transition-colors",
                current ? "border-ink bg-ink text-bone" : "border-ink/15 hover:border-ink",
                soldOut && !current && "text-smoke line-through decoration-signal",
              )}
            >
              {variantLabel(option.name)}
            </Link>
          );
        })}
      </div>
    </div>
  );
}

function Fold({ title, children, open = false }: { title: string; children: ReactNode; open?: boolean }) {
  return (
    <details open={open} className="group border-t border-ink/10 py-1 [&_summary::-webkit-details-marker]:hidden">
      <summary className="flex cursor-pointer list-none items-center justify-between py-4 font-display text-lg font-bold tracking-tight">
        {title}
        <Plus className="size-5 transition-transform duration-300 group-open:rotate-45" />
      </summary>
      <div className="pb-6">{children}</div>
    </details>
  );
}

export function ProductPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { lines } = useCart();
  const bag = useBag();
  const buyBox = useRef<HTMLDivElement>(null);
  const [quantity, setQuantity] = useState(1);
  const product = useQuery({ queryKey: ["product", id], queryFn: () => productsApi.get(id) });
  const item = product.data;

  const related = useQuery({
    queryKey: ["search", "related", item?.category.id],
    queryFn: () => productsApi.search({ category_id: item!.category.id, sort: "bestselling", limit: 60 }),
    enabled: !!item,
    staleTime: 60_000,
  });
  const brandMore = useQuery({
    queryKey: ["search", "brand", item?.brand],
    queryFn: () => productsApi.search({ brand: [item!.brand], sort: "reviews", limit: 60 }),
    enabled: !!item?.brand,
    staleTime: 60_000,
  });

  useEffect(() => {
    setQuantity(1);
  }, [id]);

  useEffect(() => {
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
  }, [item]);

  if (product.isLoading) {
    return (
      <div className="mx-auto grid max-w-[1600px] gap-10 px-5 py-8 sm:px-8 lg:grid-cols-[1.15fr_1fr]">
        <Skel className="aspect-square rounded-[2.5rem]" />
        <div className="space-y-4">
          <Skel className="h-4 w-40" />
          <Skel className="h-24 w-full" />
          <Skel className="h-10 w-48" />
          <Skel className="h-40" />
        </div>
      </div>
    );
  }

  if (product.isError || !item) {
    const notFound = product.error instanceof ApiError && (product.error.status === 404 || product.error.status === 400);
    return (
      <div className="mx-auto flex max-w-3xl flex-col items-start gap-4 px-5 py-24">
        <p className="font-display text-6xl font-extrabold tracking-tight">{notFound ? "Not in the index." : "Something broke."}</p>
        <p className="text-smoke">{notFound ? "This object may have been retired from the catalogue." : errorMessage(product.error)}</p>
        <Link to="/shop/search" className={pill("ink")}>
          Back to the index
        </Link>
      </div>
    );
  }

  const available = item.inventory?.available_quantity ?? 0;
  const inBag = lines.find((line) => line.productId === item.id)?.quantity ?? 0;
  const maxAddable = Math.max(0, Math.min(available, MAX_LINE_QUANTITY) - inBag);
  const qty = Math.min(quantity, Math.max(1, maxAddable));
  const spec = (name: string) => item.specs.find((entry) => entry.name === name)?.value;
  const sentences = (item.description || "No description has been added for this object yet.").split(/(?<=\.)\s+/).filter(Boolean);

  function add(goToCheckout: boolean) {
    bag.add(
      { productId: item!.id, name: item!.name, sku: item!.sku, price: item!.price, category: item!.category?.name ?? "", imageUrl: item!.image_url },
      qty,
      goToCheckout ? null : (document.querySelector("[data-gallery]") as HTMLElement | null),
    );
    setQuantity(1);
    if (goToCheckout) navigate(user ? "/shop/checkout" : "/login", { state: { from: "/shop/checkout" } });
    else toast("Added to bag", { description: `${qty} × ${item!.name}` });
  }

  return (
    <div className="mx-auto max-w-[1600px] px-5 sm:px-8">
      <nav aria-label="Breadcrumb" className="py-4 font-mono text-[11px] uppercase tracking-[0.16em] text-smoke">
        <Link to="/shop/search" className="hover:text-ink">Index</Link>
        <span className="mx-2">/</span>
        <Link to={`/shop/search?category_id=${item.category?.id}`} className="hover:text-ink">{item.category?.name}</Link>
        {item.brand && (
          <>
            <span className="mx-2">/</span>
            <Link to={`/shop/search?brand=${encodeURIComponent(item.brand)}`} className="hover:text-ink">{item.brand}</Link>
          </>
        )}
      </nav>

      <div className="grid items-start gap-10 lg:grid-cols-[1.15fr_1fr] lg:gap-16">
        <div data-gallery>
          <Gallery key={item.id} product={item} />
        </div>

        <motion.div ref={buyBox} initial={{ opacity: 0, y: 24 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.7, ease }} className="min-w-0">
          <Eyebrow>
            {item.brand || "meridian/"} · {item.category?.name} · {item.sku}
          </Eyebrow>
          <h1 className="mt-3 font-display text-[clamp(2.2rem,4.2vw,3.8rem)] font-extrabold leading-[0.95] tracking-[-0.045em]">{item.name}</h1>
          <a href="#reviews" className="mt-4 inline-flex items-center gap-2 text-sm hover:underline">
            <StarRow value={item.rating_average} />
            <span className="font-mono text-xs text-smoke">
              {item.rating_count > 0 ? `${item.rating_average.toFixed(1)} · ${item.rating_count} reviews` : "No reviews yet"}
            </span>
          </a>

          <div className="mt-8 flex flex-wrap items-end gap-4">
            <Price price={item.price} mrp={item.mrp} size="lg" />
          </div>
          <p className="mt-1 text-sm text-smoke">Inclusive of all taxes</p>

          <div className="mt-8 space-y-6">
            <Variants product={item} />
            <StockMeter available={available} />

            {available > 0 ? (
              <div className="flex flex-wrap items-center gap-3">
                <Stepper value={qty} max={Math.max(1, maxAddable)} onChange={setQuantity} label="Quantity" />
                <button onClick={() => add(false)} disabled={maxAddable <= 0} className={cn(pill("acid", "lg"), "flex-1")}>
                  Add to bag
                </button>
                <button onClick={() => add(true)} disabled={maxAddable <= 0} className={cn(pill("ink", "lg"), "flex-1")}>
                  Buy now <ArrowRight className="size-5" />
                </button>
                {inBag > 0 && (
                  <p className="w-full font-mono text-xs text-smoke">
                    {inBag} already in your bag{maxAddable <= 0 ? " — that's all the stock there is." : "."}
                  </p>
                )}
              </div>
            ) : (
              <p className="rounded-2xl bg-bone-2 p-4 text-sm">
                This object is sold out. Save it to your wishlist and check back: stock is updated live.
              </p>
            )}

            <dl className="grid gap-px overflow-hidden rounded-3xl bg-ink/10 sm:grid-cols-3">
              {[
                { icon: Truck, label: "Delivery", value: `${item.price >= 499 ? "Free" : "Standard"} · ${deliveryDate(item.price >= 499 ? 3 : 5)}` },
                { icon: RotateCcw, label: "Returns", value: spec("Returns") ?? "Easy returns" },
                { icon: ShieldCheck, label: "Warranty", value: spec("Warranty") ?? "Seller warranty" },
              ].map(({ icon: Icon, label, value }) => (
                <div key={label} className="bg-bone p-4">
                  <Icon className="size-5" />
                  <dt className="mt-3 font-mono text-[10px] uppercase tracking-[0.16em] text-smoke">{label}</dt>
                  <dd className="mt-0.5 text-sm">{value}</dd>
                </div>
              ))}
            </dl>
          </div>

          <div className="mt-10">
            <Fold title="About this object" open>
              <ul className="space-y-2 text-[15px] leading-relaxed">
                {sentences.map((sentence, index) => (
                  <li key={index} className="flex gap-3">
                    <span className="mt-2.5 size-1.5 shrink-0 rounded-full bg-ultra" />
                    {sentence}
                  </li>
                ))}
              </ul>
            </Fold>
            {item.specs.length > 0 && (
              <Fold title="Specifications">
                <table className="w-full text-sm">
                  <tbody>
                    {item.specs.map((entry) => (
                      <tr key={entry.name} className="border-b border-ink/10 last:border-0">
                        <th className="w-2/5 py-2.5 pr-4 text-left font-mono text-[11px] font-normal uppercase tracking-wider text-smoke">{entry.name}</th>
                        <td className="py-2.5">{entry.value}</td>
                      </tr>
                    ))}
                    <tr>
                      <th className="w-2/5 py-2.5 pr-4 text-left font-mono text-[11px] font-normal uppercase tracking-wider text-smoke">SKU</th>
                      <td className="py-2.5 font-mono">{item.sku}</td>
                    </tr>
                  </tbody>
                </table>
              </Fold>
            )}
            <Fold title="Delivery & orders">
              <p className="text-[15px] leading-relaxed text-smoke">
                Stock is reserved the moment you place your order and held while you pay. Payment is verified with Razorpay before the order is
                confirmed, and you can follow it from packed to delivered in your orders.
              </p>
            </Fold>
          </div>
        </motion.div>
      </div>

      <Rail
        eyebrow={`More in ${item.category?.name ?? "this department"}`}
        title={
          <>
            Pairs <span className="font-serif font-normal italic text-ultra">well</span> with
          </>
        }
        products={distinct(related.data?.items.filter((other) => other.image_url !== item.image_url), 14)}
        loading={related.isLoading}
      />
      {item.brand && (
        <Rail
          eyebrow="Same maker"
          title={<>More from {item.brand}</>}
          products={distinct(brandMore.data?.items.filter((other) => other.image_url !== item.image_url), 14)}
          loading={brandMore.isLoading}
        />
      )}

      <ReviewsSection product={item} />
    </div>
  );
}
