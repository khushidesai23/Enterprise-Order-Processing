import { useEffect, useLayoutEffect, useRef, useState, type MouseEvent } from "react";
import { Link } from "react-router-dom";
import { useQueries, useQuery } from "@tanstack/react-query";
import { motion, useMotionValue, useReducedMotion, useScroll, useSpring, useTransform } from "motion/react";
import { ArrowDown, ArrowRight, ArrowUpRight, BadgeCheck, Quote } from "lucide-react";
import { categoriesApi, productsApi, reviewsApi } from "@/lib/endpoints";
import { number } from "@/lib/format";
import type { ProductListItem, SearchParams } from "@/lib/types";
import { cn } from "@/lib/utils";
import { DepartmentIndex } from "@/store/StoreShell";
import { ProductTile } from "@/store/ProductTile";
import {
  Eyebrow,
  Magnetic,
  MaskText,
  Marquee,
  Price,
  ProductShot,
  Reveal,
  Skel,
  StarRow,
  Sticker,
  discountOf,
  distinct,
  ease,
  pill,
} from "@/store/ui";

function useSearch(key: string, params: SearchParams, enabled = true) {
  return useQuery({ queryKey: ["search", "home", key, params], queryFn: () => productsApi.search(params), staleTime: 60_000, enabled });
}

function useMedia(query: string) {
  const [matches, setMatches] = useState(() => typeof window !== "undefined" && window.matchMedia(query).matches);
  useEffect(() => {
    const media = window.matchMedia(query);
    const update = () => setMatches(media.matches);
    update();
    media.addEventListener("change", update);
    return () => media.removeEventListener("change", update);
  }, [query]);
  return matches;
}

// The best seller of each of the largest departments (one small query each,
// cached and shared by the orbit and the most-wanted gallery).
function useDepartmentLeaders(count: number) {
  const departments = useQuery({ queryKey: ["categories"], queryFn: categoriesApi.list, staleTime: 5 * 60_000 });
  const largest = [...(departments.data ?? [])]
    .filter((department) => department.product_count > 0)
    .sort((a, b) => b.product_count - a.product_count)
    .slice(0, count);
  const leaders = useQueries({
    queries: largest.map((department) => ({
      queryKey: ["search", "dept-preview", department.id],
      queryFn: () => productsApi.search({ category_id: department.id, sort: "bestselling", limit: 1 }),
      staleTime: 5 * 60_000,
    })),
  });
  const items = leaders.map((query) => query.data?.items[0]).filter((item): item is ProductListItem => !!item);
  return { items, loading: departments.isLoading || leaders.some((query) => query.isLoading) };
}

// --- Hero: kinetic type and an orbit of best sellers -----------------------------

function Orbit({ products, featured }: { products: ProductListItem[]; featured?: ProductListItem }) {
  const reduce = useReducedMotion();
  const small = useMedia("(max-width: 639px)");
  const tiltX = useSpring(useMotionValue(0), { stiffness: 80, damping: 20 });
  const tiltY = useSpring(useMotionValue(0), { stiffness: 80, damping: 20 });
  const radius = small ? 150 : 300;
  const item = small ? 74 : 108;

  function onMove(event: MouseEvent<HTMLDivElement>) {
    if (reduce) return;
    const rect = event.currentTarget.getBoundingClientRect();
    tiltY.set(((event.clientX - rect.left) / rect.width - 0.5) * 16);
    tiltX.set(-((event.clientY - rect.top) / rect.height - 0.5) * 12);
  }

  return (
    <div
      className="relative h-[380px] sm:h-[600px]"
      style={{ perspective: 1400 }}
      onMouseMove={onMove}
      onMouseLeave={() => {
        tiltX.set(0);
        tiltY.set(0);
      }}
    >
      <motion.div className="absolute inset-0" style={{ rotateX: tiltX, rotateY: tiltY, transformStyle: "preserve-3d" }}>
        {/* Orbit path */}
        <div
          aria-hidden
          className="absolute left-1/2 top-1/2 rounded-full border border-ink/15"
          style={{ width: radius * 2, height: radius * 0.55, marginLeft: -radius, marginTop: -radius * 0.275 }}
        />
        <div className="orbit-ring absolute left-1/2 top-1/2" style={{ width: 0, height: 0 }}>
          {products.map((product, index) => {
            const angle = (index * 360) / products.length;
            return (
              <div
                key={product.id}
                className="absolute"
                style={{
                  width: item,
                  height: item,
                  left: -item / 2,
                  top: -item / 2,
                  transform: `rotateY(${angle}deg) translateZ(${radius}px)`,
                  transformStyle: "preserve-3d",
                }}
              >
                <div style={{ transform: `rotateY(${-angle}deg)`, transformStyle: "preserve-3d" }}>
                  <div className="orbit-face">
                    <Link
                      to={`/shop/products/${product.id}`}
                      className="block rounded-[1.4rem] bg-bone-2/90 p-2 shadow-[0_18px_40px_-18px_rgb(0_0_0/0.4)] transition-transform hover:scale-110"
                      style={{ width: item, height: item }}
                      aria-label={product.name}
                    >
                      <ProductShot id={product.id} name={product.name} src={product.image_url} size="sm" eager className="size-full" />
                    </Link>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
        {featured && (
          <Link
            to={`/shop/products/${featured.id}`}
            className="float absolute left-1/2 top-1/2 block -translate-x-1/2 -translate-y-1/2"
            style={{ transform: "translateZ(40px)" }}
            aria-label={`Featured: ${featured.name}`}
          >
            <span className="relative block size-48 sm:size-72">
              <span aria-hidden className="absolute inset-[12%] rounded-full bg-acid blur-3xl opacity-70" />
              <ProductShot id={featured.id} name={featured.name} src={featured.image_url} size="lg" eager className="relative size-full drop-shadow-[0_30px_40px_rgb(0_0_0/0.25)]" />
              {discountOf(featured.price, featured.mrp) > 0 && (
                <Sticker className="absolute -right-2 top-2 size-20 text-base">-{discountOf(featured.price, featured.mrp)}%</Sticker>
              )}
            </span>
          </Link>
        )}
      </motion.div>
    </div>
  );
}

function Hero() {
  const total = useSearch("total", { limit: 1 });
  const inStock = useSearch("in-stock", { in_stock: true, limit: 1 });
  const departments = useQuery({ queryKey: ["categories"], queryFn: categoriesApi.list, staleTime: 5 * 60_000 });
  const leaders = useDepartmentLeaders(9);
  const deals = useSearch("hero-deal", { deals: true, sort: "discount", in_stock: true, limit: 10 });
  // One best seller per department, so the orbit spans the catalogue.
  const orbit = leaders.loading ? [] : leaders.items;
  const featured = deals.data?.items[0];

  return (
    <section className="relative mx-auto grid max-w-[1600px] items-center gap-4 px-5 pb-10 pt-4 sm:px-8 lg:min-h-[calc(100vh-72px)] lg:grid-cols-[1.05fr_1fr]">
      <div className="relative z-10">
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.8, delay: 0.1 }}>
          <Eyebrow className="flex flex-wrap items-center gap-x-3 gap-y-1">
            <span className="inline-flex items-center gap-1.5">
              <span className="size-1.5 animate-pulse rounded-full bg-signal" /> Live index
            </span>
            <span>{total.data ? `${number(total.data.total)} objects` : "— objects"}</span>
            <span>{inStock.data ? `${number(inStock.data.total)} in stock now` : ""}</span>
            <span>{departments.data ? `${departments.data.length} departments` : ""}</span>
          </Eyebrow>
        </motion.div>
        <h1 className="mt-6 font-display text-[clamp(3rem,7.2vw,7rem)] font-extrabold leading-[0.88] tracking-[-0.055em]">
          <MaskText text="Objects for" className="block" />
          <span className="block">
            <MaskText text="the" delay={0.15} />{" "}
            <span className="relative inline-block">
              <MaskText text="next" delay={0.22} className="font-serif font-normal italic tracking-[-0.02em]" />
              <motion.span
                aria-hidden
                className="absolute -bottom-1 left-0 h-[0.14em] w-full origin-left rounded-full bg-acid"
                initial={{ scaleX: 0 }}
                animate={{ scaleX: 1 }}
                transition={{ delay: 0.9, duration: 0.8, ease }}
              />
            </span>
          </span>
          <MaskText text="decade." delay={0.3} className="block" />
        </h1>
        <Reveal delay={0.5}>
          <p className="mt-8 max-w-md text-lg leading-relaxed text-smoke">
            A living index of things worth owning, from pocket tech to pantry staples. Live stock, honest prices, and orders tracked from bag to door.
          </p>
        </Reveal>
        <Reveal delay={0.65} className="mt-8 flex flex-wrap items-center gap-3">
          <Magnetic>
            <Link to="/shop/search" className={pill("acid", "lg")}>
              Enter the index <ArrowRight className="size-5 transition-transform group-hover/pill:translate-x-1" />
            </Link>
          </Magnetic>
          <Magnetic>
            <Link to="/shop/search?deals=true&sort=discount" className={pill("line", "lg")}>
              Today's drops
            </Link>
          </Magnetic>
        </Reveal>
        <div className="mt-14 hidden items-center gap-3 font-mono text-[11px] uppercase tracking-[0.2em] text-smoke lg:flex">
          <ArrowDown className="size-4 animate-bounce" /> Scroll to explore
        </div>
      </div>
      <motion.div initial={{ opacity: 0, scale: 0.92 }} animate={{ opacity: 1, scale: 1 }} transition={{ duration: 1.2, ease, delay: 0.2 }}>
        {orbit.length > 0 ? <Orbit products={orbit} featured={featured} /> : <Skel className="mx-auto h-[380px] w-[90%] rounded-full sm:h-[600px]" />}
      </motion.div>
    </section>
  );
}

// --- Sections ---------------------------------------------------------------------

function SectionHead({ index, title, accent, aside }: { index: string; title: string; accent?: string; aside?: React.ReactNode }) {
  return (
    <div className="mb-10 flex flex-wrap items-end justify-between gap-6">
      <div>
        <Eyebrow>{index}</Eyebrow>
        <h2 className="mt-3 font-display text-[clamp(2.4rem,5vw,4.6rem)] font-extrabold leading-[0.9] tracking-[-0.045em]">
          <MaskText text={title} />
          {accent && (
            <>
              {" "}
              <MaskText text={accent} delay={0.12} className="font-serif font-normal italic tracking-[-0.02em] text-ultra" />
            </>
          )}
        </h2>
      </div>
      {aside}
    </div>
  );
}

function Drops() {
  const deals = useSearch("drops", { deals: true, sort: "discount", in_stock: true, limit: 60 });
  const items = distinct(deals.data?.items, 5);
  const [feature, ...rest] = items;

  return (
    <section className="mx-auto max-w-[1600px] px-5 py-20 sm:px-8">
      <SectionHead
        index="001 — The drop"
        title="Today's"
        accent="drops."
        aside={
          <div className="flex items-center gap-4">
            <p className="max-w-56 text-sm text-smoke">
              {deals.data ? `${number(deals.data.total)} objects below list price right now.` : "Loading live deals…"}
            </p>
            <Link to="/shop/search?deals=true&sort=discount" className={pill("ink", "md")}>
              All drops <ArrowUpRight className="size-4" />
            </Link>
          </div>
        }
      />
      {deals.isLoading ? (
        <div className="grid gap-6 lg:grid-cols-12">
          <Skel className="h-[560px] lg:col-span-6" />
          <div className="grid grid-cols-2 gap-6 lg:col-span-6">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skel key={i} className="h-64" />
            ))}
          </div>
        </div>
      ) : (
        <div className="grid gap-x-6 gap-y-10 lg:grid-cols-12">
          {feature && <ProductTile product={feature} size="lg" index={0} className="lg:col-span-6" />}
          <div className="grid grid-cols-2 gap-x-6 gap-y-10 lg:col-span-6">
            {rest.map((product, index) => (
              <ProductTile key={product.id} product={product} index={index + 1} />
            ))}
          </div>
        </div>
      )}
    </section>
  );
}

function PromiseBand() {
  const lines = ["Stock reserved at order", "Payments verified", "Tracked to your door", "Free delivery over ₹499", "Real reviews only"];
  return (
    <section aria-label="Our promises" className="relative -rotate-[1.5deg] overflow-hidden bg-ink py-5 text-bone">
      <Marquee duration={32}>
        {lines.map((line) => (
          <span key={line} className="flex items-center gap-8 px-8 font-display text-3xl font-bold tracking-tight sm:text-5xl">
            {line} <span className="text-acid">✳</span>
          </span>
        ))}
      </Marquee>
    </section>
  );
}

function Departments() {
  return (
    <section className="mx-auto grid max-w-[1600px] gap-12 px-5 py-24 sm:px-8 lg:grid-cols-[1fr_2fr]">
      <div className="lg:sticky lg:top-28 lg:self-start">
        <Eyebrow>002 — Departments</Eyebrow>
        <h2 className="mt-3 font-display text-[clamp(2.4rem,5vw,4.6rem)] font-extrabold leading-[0.9] tracking-[-0.045em]">
          <MaskText text="Every corner," />
          <MaskText text="indexed." delay={0.1} className="block font-serif font-normal italic tracking-[-0.02em] text-ultra" />
        </h2>
        <p className="mt-6 max-w-sm text-smoke">Hover a department to preview its best seller. Counts are live from the catalogue.</p>
      </div>
      <DepartmentIndex inverse={false} />
    </section>
  );
}

// Horizontal gallery driven by vertical scroll; a plain swipe row on
// small screens and with reduced motion.
function MostWanted() {
  const best = useDepartmentLeaders(10);
  const items = best.items;
  const reduce = useReducedMotion();
  const wide = useMedia("(min-width: 1024px)");
  const pinned = wide && !reduce;
  const section = useRef<HTMLElement>(null);
  const track = useRef<HTMLDivElement>(null);
  const [distance, setDistance] = useState(0);
  const { scrollYProgress } = useScroll({ target: section, offset: ["start start", "end end"] });
  const x = useTransform(scrollYProgress, [0, 1], [0, -distance]);

  useLayoutEffect(() => {
    const measure = () => setDistance(Math.max(0, (track.current?.scrollWidth ?? 0) - window.innerWidth + 64));
    measure();
    window.addEventListener("resize", measure);
    return () => window.removeEventListener("resize", measure);
  }, [items.length, pinned]);

  const cards = items.map((product, index) => (
    <Link
      key={product.id}
      to={`/shop/products/${product.id}`}
      className="group relative flex w-[78vw] shrink-0 flex-col sm:w-[360px]"
      data-testid="wanted-card"
    >
      <span
        aria-hidden
        className="pointer-events-none absolute -top-10 left-0 z-10 font-display text-[7rem] font-extrabold leading-none tracking-[-0.06em] text-transparent [-webkit-text-stroke:1.5px_var(--color-ink)]"
      >
        {String(index + 1).padStart(2, "0")}
      </span>
      <div className="relative overflow-hidden rounded-[2rem] bg-bone-2 p-10">
        <ProductShot id={product.id} name={product.name} src={product.image_url} size="lg" eager className="aspect-square" imgClassName="transition-transform duration-700 group-hover:scale-110 group-hover:rotate-3" />
      </div>
      <div className="mt-4 flex items-start justify-between gap-4 px-1">
        <div className="min-w-0">
          <p className="font-mono text-[10px] uppercase tracking-[0.16em] text-smoke">{product.brand || product.category}</p>
          <p className="mt-1 line-clamp-1 text-lg font-medium group-hover:underline">{product.name}</p>
        </div>
        <Price price={product.price} mrp={product.mrp} size="sm" className="shrink-0 justify-end" />
      </div>
    </Link>
  ));

  const head = (
    <div className="mx-auto mb-20 flex max-w-[1600px] flex-wrap items-end justify-between gap-6 px-5 sm:px-8">
      <div>
        <Eyebrow>003 — Most wanted</Eyebrow>
        <h2 className="mt-3 font-display text-[clamp(2.4rem,5vw,4.6rem)] font-extrabold leading-[0.9] tracking-[-0.045em]">
          What everyone's <span className="font-serif font-normal italic tracking-[-0.02em] text-ultra">ordering.</span>
        </h2>
      </div>
      <p className="max-w-xs text-sm text-smoke">The best seller in each of our biggest departments, by units sold through paid and delivered orders.</p>
    </div>
  );

  if (!pinned) {
    return (
      <section className="py-20">
        {head}
        <div className="no-scrollbar flex snap-x gap-6 overflow-x-auto px-5 pt-10 sm:px-8">
          {best.loading ? Array.from({ length: 3 }).map((_, i) => <Skel key={i} className="h-96 w-[78vw] shrink-0 sm:w-[360px]" />) : cards.map((card, i) => <div key={i} className="snap-start">{card}</div>)}
        </div>
      </section>
    );
  }

  return (
    <section ref={section} className="relative" style={{ height: `calc(100vh + ${distance}px)` }}>
      <div className="sticky top-0 flex h-screen flex-col justify-center overflow-hidden">
        {head}
        <motion.div ref={track} style={{ x }} className="flex gap-8 px-8 pt-10">
          {best.loading ? Array.from({ length: 4 }).map((_, i) => <Skel key={i} className="h-96 w-[360px] shrink-0" />) : cards}
        </motion.div>
      </div>
    </section>
  );
}

function Voices() {
  const popular = useSearch("voices", { sort: "reviews", min_rating: 4, limit: 30 });
  const products = distinct(popular.data?.items, 3);
  const reviews = useQueries({
    queries: products.map((product) => ({
      queryKey: ["reviews", product.id, "voice"],
      queryFn: () => reviewsApi.forProduct(product.id, { sort: "highest", limit: 3 }),
      staleTime: 5 * 60_000,
    })),
  });

  const quotes = products
    .map((product, index) => ({ product, review: reviews[index]?.data?.items.find((review) => review.body.length > 30) }))
    .filter((entry) => entry.review);

  if (products.length === 0) return null;

  return (
    <section className="bg-ultra py-24 text-bone">
      <div className="mx-auto max-w-[1600px] px-5 sm:px-8">
        <Eyebrow className="text-bone/60">004 — Said by humans</Eyebrow>
        <h2 className="mt-3 max-w-4xl font-display text-[clamp(2.4rem,5vw,4.6rem)] font-extrabold leading-[0.9] tracking-[-0.045em]">
          <MaskText text="Reviews from people" /> <MaskText text="who bought it." delay={0.1} className="font-serif font-normal italic tracking-[-0.02em] text-acid" />
        </h2>
        <div className="mt-14 grid gap-6 md:grid-cols-3">
          {quotes.map(({ product, review }, index) => (
            <Reveal key={product.id} delay={index * 0.1}>
              <figure className={cn("flex h-full flex-col rounded-[2rem] bg-bone p-7 text-ink", index === 1 && "md:translate-y-10")}>
                <Quote className="size-8 text-ultra" />
                <StarRow value={review!.rating} className="mt-4" />
                <blockquote className="mt-4 flex-1 font-serif text-2xl leading-snug">“{review!.body}”</blockquote>
                <figcaption className="mt-6 flex items-center gap-3 border-t border-ink/10 pt-5">
                  <ProductShot id={product.id} name={product.name} src={product.image_url} size="sm" className="size-14 shrink-0 rounded-2xl bg-bone-2 p-1" />
                  <div className="min-w-0">
                    <p className="flex items-center gap-1 text-sm font-medium">
                      {review!.reviewer_name}
                      {review!.verified_purchase && <BadgeCheck className="size-4 text-ultra" aria-label="Verified purchase" />}
                    </p>
                    <Link to={`/shop/products/${product.id}`} className="block truncate font-mono text-[11px] uppercase tracking-wider text-smoke hover:text-ink">
                      on {product.name}
                    </Link>
                  </div>
                </figcaption>
              </figure>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}

function NewIn() {
  const newest = useSearch("new", { sort: "newest", in_stock: true, limit: 60 });
  const items = distinct(newest.data?.items, 7);

  return (
    <section className="mx-auto max-w-[1600px] px-5 py-24 sm:px-8">
      <SectionHead
        index="005 — New in"
        title="Fresh"
        accent="arrivals."
        aside={
          <Link to="/shop/search?sort=newest" className={pill("line")}>
            Everything new <ArrowUpRight className="size-4" />
          </Link>
        }
      />
      {newest.isLoading ? (
        <Skel className="h-[600px]" />
      ) : (
        <div className="grid grid-cols-2 gap-x-6 gap-y-10 md:grid-cols-4">
          {items.map((product, index) => (
            <ProductTile key={product.id} product={product} index={index} size={index === 0 ? "lg" : "md"} className={index === 0 ? "col-span-2 row-span-2" : undefined} />
          ))}
        </div>
      )}
    </section>
  );
}

function Closer() {
  return (
    <section className="mx-auto max-w-[1600px] px-5 sm:px-8">
      <div className="relative overflow-hidden rounded-[2.5rem] bg-acid px-8 py-20 sm:px-16">
        <svg aria-hidden viewBox="0 0 400 400" className="absolute -right-20 -top-20 size-[420px] text-ink/10">
          <circle cx="200" cy="200" r="180" fill="none" stroke="currentColor" strokeWidth="2" />
          <ellipse cx="200" cy="200" rx="70" ry="180" fill="none" stroke="currentColor" strokeWidth="2" />
          <ellipse cx="200" cy="200" rx="180" ry="60" fill="none" stroke="currentColor" strokeWidth="2" />
        </svg>
        <h2 className="relative max-w-4xl font-display text-[clamp(2.6rem,6vw,5.6rem)] font-extrabold leading-[0.9] tracking-[-0.05em]">
          Your next object is <span className="font-serif font-normal italic">one search</span> away.
        </h2>
        <div className="relative mt-10 flex flex-wrap gap-3">
          <Magnetic>
            <Link to="/shop/search" className={pill("ink", "lg")}>
              Open the index <ArrowRight className="size-5" />
            </Link>
          </Magnetic>
          <Link to="/register" className={pill("line", "lg")}>
            Create an account
          </Link>
        </div>
      </div>
    </section>
  );
}

export function HomePage() {
  return (
    <div>
      <Hero />
      <PromiseBand />
      <Drops />
      <Departments />
      <MostWanted />
      <Voices />
      <NewIn />
      <Closer />
    </div>
  );
}
