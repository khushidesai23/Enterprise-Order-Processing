import { Fragment, useState, type FormEvent, type ReactNode } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import * as Dialog from "@radix-ui/react-dialog";
import { AnimatePresence, motion } from "motion/react";
import { ArrowLeft, ArrowRight, SlidersHorizontal, X } from "lucide-react";
import { productsApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { money, number } from "@/lib/format";
import type { ProductSearchResult, SearchParams, SearchSort } from "@/lib/types";
import { cn } from "@/lib/utils";
import { ProductTile } from "@/store/ProductTile";
import { Eyebrow, Skel, StarRow, ease, pill } from "@/store/ui";

const PAGE_SIZE = 24;

const SORTS: { value: SearchSort; label: string }[] = [
  { value: "featured", label: "Featured" },
  { value: "bestselling", label: "Most wanted" },
  { value: "rating", label: "Highest rated" },
  { value: "reviews", label: "Most reviewed" },
  { value: "discount", label: "Biggest drop" },
  { value: "newest", label: "Newest" },
  { value: "price_asc", label: "Price: low to high" },
  { value: "price_desc", label: "Price: high to low" },
];

const PRICE_RANGES: { label: string; min?: number; max?: number }[] = [
  { label: "Under ₹500", max: 500 },
  { label: "₹500 – 1k", min: 500, max: 1000 },
  { label: "₹1k – 5k", min: 1000, max: 5000 },
  { label: "₹5k – 20k", min: 5000, max: 20000 },
  { label: "₹20k – 1L", min: 20000, max: 100000 },
  { label: "Over ₹1L", min: 100000 },
];

function readParams(params: URLSearchParams): SearchParams {
  const num = (key: string) => {
    const value = params.get(key);
    return value === null || value === "" || Number.isNaN(Number(value)) ? undefined : Number(value);
  };
  return {
    q: params.get("q") || undefined,
    category_id: params.get("category_id") || undefined,
    brand: params.getAll("brand"),
    min_price: num("min_price"),
    max_price: num("max_price"),
    min_rating: num("min_rating"),
    in_stock: params.get("in_stock") === "true",
    deals: params.get("deals") === "true",
    sort: (params.get("sort") as SearchSort) || undefined,
    page: num("page") ?? 1,
    limit: PAGE_SIZE,
  };
}

type Update = (changes: Record<string, string | string[] | undefined>) => void;

function FilterGroup({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="border-t border-ink/10 py-5">
      <p className="mb-3 font-mono text-[11px] uppercase tracking-[0.18em] text-smoke">{title}</p>
      {children}
    </div>
  );
}

function FilterPanel({ search, facets, update }: { search: SearchParams; facets?: ProductSearchResult["facets"]; update: Update }) {
  const [showAllBrands, setShowAllBrands] = useState(false);
  const [minDraft, setMinDraft] = useState(search.min_price?.toString() ?? "");
  const [maxDraft, setMaxDraft] = useState(search.max_price?.toString() ?? "");
  const brands = facets?.brands ?? [];
  const toggleBrand = (brand: string) => {
    const current = search.brand ?? [];
    update({ brand: current.includes(brand) ? current.filter((item) => item !== brand) : [...current, brand] });
  };
  const applyPrice = (event: FormEvent) => {
    event.preventDefault();
    update({ min_price: minDraft, max_price: maxDraft });
  };
  const chip = (active: boolean) =>
    cn(
      "rounded-full border px-3 py-1.5 text-left text-sm transition-colors",
      active ? "border-ink bg-ink text-bone" : "border-ink/15 hover:border-ink/50",
    );

  return (
    <div>
      <FilterGroup title="Department">
        <ul className="space-y-0.5">
          {search.category_id && (
            <li>
              <button onClick={() => update({ category_id: undefined })} className="flex items-center gap-1 py-1 text-sm text-smoke hover:text-ink">
                <ArrowLeft className="size-3.5" /> All departments
              </button>
            </li>
          )}
          {(facets?.categories ?? []).map((item) => (
            <li key={item.id}>
              <button
                onClick={() => update({ category_id: item.id })}
                className={cn(
                  "flex w-full items-baseline justify-between gap-2 rounded-lg px-2 py-1 text-left text-sm transition-colors hover:bg-ink/5",
                  item.id === search.category_id && "bg-ink text-bone hover:bg-ink",
                )}
              >
                {item.name} <span className="font-mono text-[11px] opacity-60">{number(item.count)}</span>
              </button>
            </li>
          ))}
        </ul>
      </FilterGroup>

      {brands.length > 0 && (
        <FilterGroup title="Brand">
          <ul className="space-y-1.5">
            {(showAllBrands ? brands : brands.slice(0, 8)).map((brand) => (
              <li key={brand.name}>
                <label className="flex cursor-pointer items-center gap-2.5 text-sm">
                  <input
                    type="checkbox"
                    className="size-4 rounded accent-[var(--color-ink)]"
                    checked={(search.brand ?? []).includes(brand.name)}
                    onChange={() => toggleBrand(brand.name)}
                  />
                  {brand.name} <span className="font-mono text-[11px] text-smoke">{brand.count}</span>
                </label>
              </li>
            ))}
          </ul>
          {brands.length > 8 && (
            <button onClick={() => setShowAllBrands(!showAllBrands)} className="mt-3 font-mono text-[11px] uppercase tracking-wider underline underline-offset-4">
              {showAllBrands ? "Fewer brands" : `All ${brands.length} brands`}
            </button>
          )}
        </FilterGroup>
      )}

      <FilterGroup title="Price">
        <div className="flex flex-wrap gap-1.5">
          {PRICE_RANGES.map((range) => {
            const active = search.min_price === range.min && search.max_price === range.max;
            return (
              <button
                key={range.label}
                className={chip(active)}
                onClick={() =>
                  active
                    ? update({ min_price: undefined, max_price: undefined })
                    : update({ min_price: range.min?.toString(), max_price: range.max?.toString() })
                }
              >
                {range.label}
              </button>
            );
          })}
        </div>
        <form onSubmit={applyPrice} className="mt-3 flex items-center gap-1.5">
          <input
            value={minDraft}
            onChange={(event) => setMinDraft(event.target.value)}
            inputMode="numeric"
            placeholder="Min ₹"
            aria-label="Minimum price"
            className="h-9 w-full min-w-0 rounded-full border border-ink/15 bg-transparent px-3 font-mono text-sm focus:border-ink focus:outline-none"
          />
          <input
            value={maxDraft}
            onChange={(event) => setMaxDraft(event.target.value)}
            inputMode="numeric"
            placeholder="Max ₹"
            aria-label="Maximum price"
            className="h-9 w-full min-w-0 rounded-full border border-ink/15 bg-transparent px-3 font-mono text-sm focus:border-ink focus:outline-none"
          />
          <button type="submit" className={pill("ink", "sm")}>
            Go
          </button>
        </form>
        {facets && facets.price_max > 0 && (
          <p className="mt-2 font-mono text-[11px] text-smoke">
            {money(facets.price_min)} – {money(facets.price_max)} here
          </p>
        )}
      </FilterGroup>

      <FilterGroup title="Customer rating">
        <div className="space-y-1">
          {[4, 3, 2].map((stars) => (
            <button
              key={stars}
              onClick={() => update({ min_rating: search.min_rating === stars ? undefined : String(stars) })}
              className={cn("flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-sm hover:bg-ink/5", search.min_rating === stars && "bg-ink text-bone hover:bg-ink")}
              aria-label={`${stars} stars and up`}
              aria-pressed={search.min_rating === stars}
            >
              <StarRow value={stars} size={14} className={search.min_rating === stars ? "[&_svg]:!fill-bone [&_svg]:!text-bone" : undefined} />& up
            </button>
          ))}
        </div>
      </FilterGroup>

      <FilterGroup title="Availability">
        <label className="flex cursor-pointer items-center gap-2.5 py-1 text-sm">
          <input
            type="checkbox"
            className="size-4 accent-[var(--color-ink)]"
            checked={!!search.deals}
            onChange={(event) => update({ deals: event.target.checked ? "true" : undefined })}
          />
          Drops only (below list price)
        </label>
        <label className="flex cursor-pointer items-center gap-2.5 py-1 text-sm">
          <input
            type="checkbox"
            className="size-4 accent-[var(--color-ink)]"
            checked={!!search.in_stock}
            onChange={(event) => update({ in_stock: event.target.checked ? "true" : undefined })}
          />
          In stock only
        </label>
      </FilterGroup>
    </div>
  );
}

export function SearchPage() {
  const [params, setParams] = useSearchParams();
  const search = readParams(params);
  const [filtersOpen, setFiltersOpen] = useState(false);

  const result = useQuery({
    queryKey: ["search", "page", params.toString()],
    queryFn: () => productsApi.search(search),
    placeholderData: keepPreviousData,
  });

  // Every filter change returns to page 1.
  const update: Update = (changes) => {
    const next = new URLSearchParams(params);
    for (const [key, value] of Object.entries(changes)) {
      next.delete(key);
      if (Array.isArray(value)) value.forEach((item) => next.append(key, item));
      else if (value !== undefined && value !== "") next.set(key, value);
    }
    if (!("page" in changes)) next.delete("page");
    setParams(next);
  };

  const data = result.data;
  const facets = data?.facets;
  const category = facets?.categories.find((item) => item.id === search.category_id);
  const page = search.page ?? 1;
  const pages = data ? Math.max(1, Math.ceil(data.total / PAGE_SIZE)) : 1;
  const first = data && data.total > 0 ? (page - 1) * PAGE_SIZE + 1 : 0;
  const last = data ? Math.min(page * PAGE_SIZE, data.total) : 0;

  const chips: { label: string; clear: Record<string, undefined | string[]> }[] = [];
  if (search.q) chips.push({ label: `"${search.q}"`, clear: { q: undefined } });
  if (category) chips.push({ label: category.name, clear: { category_id: undefined } });
  for (const brand of search.brand ?? []) chips.push({ label: brand, clear: { brand: (search.brand ?? []).filter((b) => b !== brand) } });
  if (search.min_price !== undefined || search.max_price !== undefined)
    chips.push({
      label: `${search.min_price !== undefined ? money(search.min_price) : "₹0"} – ${search.max_price !== undefined ? money(search.max_price) : "any"}`,
      clear: { min_price: undefined, max_price: undefined },
    });
  if (search.min_rating) chips.push({ label: `${search.min_rating}★ & up`, clear: { min_rating: undefined } });
  if (search.deals) chips.push({ label: "Drops", clear: { deals: undefined } });
  if (search.in_stock) chips.push({ label: "In stock", clear: { in_stock: undefined } });

  const title = search.q ? (
    <>
      <span className="font-serif font-normal italic tracking-[-0.02em]">“{search.q}”</span>
    </>
  ) : category ? (
    category.name
  ) : search.deals ? (
    <>
      Today's <span className="font-serif font-normal italic tracking-[-0.02em] text-ultra">drops</span>
    </>
  ) : search.sort === "bestselling" ? (
    <>
      Most <span className="font-serif font-normal italic tracking-[-0.02em] text-ultra">wanted</span>
    </>
  ) : (
    <>
      Every<span className="font-serif font-normal italic tracking-[-0.02em] text-ultra">thing.</span>
    </>
  );

  return (
    <div className="mx-auto max-w-[1600px] px-5 pb-10 pt-6 sm:px-8">
      <header className="border-b border-ink/10 pb-8">
        <Eyebrow>The index{search.q ? " — search" : category ? " — department" : ""}</Eyebrow>
        <motion.h1
          key={`${search.q}-${search.category_id}-${search.deals}-${search.sort}`}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease }}
          className="mt-3 break-words font-display text-[clamp(2.6rem,7vw,6.5rem)] font-extrabold leading-[0.9] tracking-[-0.05em]"
        >
          {title}
        </motion.h1>
      </header>

      <div className="sticky top-[72px] z-20 -mx-5 flex flex-wrap items-center gap-2 bg-bone/85 px-5 py-3 backdrop-blur-xl sm:-mx-8 sm:px-8">
        <p data-testid="result-summary" className="mr-auto font-mono text-xs text-smoke">
          {data ? (
            <>
              {number(first)}–{number(last)} of {number(data.total)}
              {search.q ? <> for <span className="text-ink">"{search.q}"</span></> : null}
              {category ? <> in {category.name}</> : null}
            </>
          ) : (
            "Searching…"
          )}
        </p>
        <button onClick={() => setFiltersOpen(true)} className={cn(pill("line", "sm"), "lg:hidden")}>
          <SlidersHorizontal className="size-4" /> Filters{chips.length ? ` (${chips.length})` : ""}
        </button>
        <label className="flex items-center gap-2">
          <span className="sr-only">Sort by</span>
          <select
            value={search.sort ?? (search.q ? "relevance" : "featured")}
            onChange={(event) => update({ sort: event.target.value })}
            className="h-9 cursor-pointer rounded-full border border-ink/15 bg-bone px-3 text-sm focus:border-ink focus:outline-none"
            aria-label="Sort results"
          >
            {search.q && <option value="relevance">Most relevant</option>}
            {SORTS.map((sort) => (
              <option key={sort.value} value={sort.value}>
                {sort.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      {chips.length > 0 && (
        <div className="mt-3 flex flex-wrap items-center gap-2">
          <AnimatePresence initial={false}>
            {chips.map((chip) => (
              <motion.button
                key={chip.label}
                layout
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.8 }}
                onClick={() => update(chip.clear)}
                className="flex items-center gap-1.5 rounded-full bg-ink px-3 py-1.5 text-xs text-bone hover:bg-ultra"
                aria-label={`Remove filter ${chip.label}`}
              >
                {chip.label} <X className="size-3" />
              </motion.button>
            ))}
          </AnimatePresence>
          <Link to="/shop/search" className="font-mono text-[11px] uppercase tracking-wider text-smoke underline-offset-4 hover:text-ink hover:underline">
            Clear all
          </Link>
        </div>
      )}

      <div className="mt-6 grid gap-10 lg:grid-cols-[250px_1fr]">
        <aside aria-label="Filters" className="hidden lg:block">
          <FilterPanel key={params.toString()} search={search} facets={facets} update={update} />
        </aside>

        <section aria-label="Results">
          {result.isError ? (
            <div className="rounded-3xl border border-signal/30 p-8">
              <p className="font-medium text-signal">{errorMessage(result.error)}</p>
              <button onClick={() => result.refetch()} className={cn(pill("ink", "sm"), "mt-4")}>
                Try again
              </button>
            </div>
          ) : result.isLoading ? (
            <div className="grid grid-cols-2 gap-x-6 gap-y-10 md:grid-cols-3 xl:grid-cols-4">
              {Array.from({ length: 8 }).map((_, i) => (
                <Skel key={i} className="aspect-[3/4]" />
              ))}
            </div>
          ) : data && data.items.length === 0 ? (
            <div className="flex flex-col items-start gap-4 py-16">
              <p className="font-display text-5xl font-extrabold tracking-tight">
                No <span className="font-serif font-normal italic">matches.</span>
              </p>
              <p className="max-w-md text-smoke">Try a broader word, a different department, or loosen the filters.</p>
              <Link to="/shop/search" className={pill("ink")}>
                Reset the index
              </Link>
            </div>
          ) : (
            <>
              <div className={cn("grid grid-cols-2 gap-x-5 gap-y-10 md:grid-cols-3 xl:grid-cols-4", result.isPlaceholderData && "opacity-50 transition-opacity")}>
                {data?.items.map((product, index) => (
                  <Fragment key={product.id}>
                    <ProductTile product={product} index={(page - 1) * PAGE_SIZE + index} />
                    {index === 7 && search.sort !== "bestselling" && data.items.length > 12 && (
                      <Link
                        to={`/shop/search?${new URLSearchParams({ ...(search.category_id ? { category_id: search.category_id } : {}), sort: "bestselling" })}`}
                        className="group col-span-2 flex flex-col justify-between overflow-hidden rounded-[1.75rem] bg-ink p-8 text-bone md:col-span-3 xl:col-span-4 xl:flex-row xl:items-end"
                      >
                        <p className="font-display text-4xl font-extrabold leading-[0.95] tracking-tight sm:text-5xl">
                          Can't decide? <span className="font-serif font-normal italic text-acid">See what sells.</span>
                        </p>
                        <span className="mt-6 inline-flex items-center gap-2 font-mono text-xs uppercase tracking-widest text-bone/70 group-hover:text-acid">
                          Sort by most wanted <ArrowRight className="size-4 transition-transform group-hover:translate-x-1" />
                        </span>
                      </Link>
                    )}
                  </Fragment>
                ))}
              </div>
              <Pager page={page} pages={pages} onChange={(next) => update({ page: String(next) })} />
            </>
          )}
        </section>
      </div>

      <Dialog.Root open={filtersOpen} onOpenChange={setFiltersOpen}>
        <AnimatePresence>
          {filtersOpen && (
            <Dialog.Portal forceMount>
              <Dialog.Overlay asChild forceMount>
                <motion.div className="fixed inset-0 z-50 bg-ink/40" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} />
              </Dialog.Overlay>
              <Dialog.Content asChild forceMount aria-describedby={undefined}>
                <motion.div
                  className="store fixed inset-x-0 bottom-0 z-50 max-h-[85vh] overflow-y-auto rounded-t-[2rem] bg-bone px-6 pb-8 pt-5 focus:outline-none"
                  initial={{ y: "100%" }}
                  animate={{ y: 0 }}
                  exit={{ y: "100%" }}
                  transition={{ duration: 0.45, ease }}
                >
                  <div className="mb-2 flex items-center justify-between">
                    <Dialog.Title className="font-display text-2xl font-bold">Filters</Dialog.Title>
                    <Dialog.Close className={pill("ink", "sm")}>Show {data ? number(data.total) : ""} results</Dialog.Close>
                  </div>
                  <FilterPanel key={params.toString()} search={search} facets={facets} update={update} />
                </motion.div>
              </Dialog.Content>
            </Dialog.Portal>
          )}
        </AnimatePresence>
      </Dialog.Root>
    </div>
  );
}

function Pager({ page, pages, onChange }: { page: number; pages: number; onChange: (page: number) => void }) {
  if (pages <= 1) return null;
  const shown = [...new Set([1, pages, page - 1, page, page + 1].filter((p) => p >= 1 && p <= pages))].sort((a, b) => a - b);
  const go = (next: number) => {
    onChange(next);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  return (
    <nav className="mt-16 flex items-center justify-center gap-2" aria-label="Pagination">
      <button disabled={page <= 1} onClick={() => go(page - 1)} className={pill("line", "sm")} aria-label="Previous page">
        <ArrowLeft className="size-4" />
      </button>
      {shown.map((p, i) => (
        <Fragment key={p}>
          {i > 0 && p - shown[i - 1]! > 1 && <span className="font-mono text-smoke">…</span>}
          <button
            onClick={() => go(p)}
            aria-current={p === page ? "page" : undefined}
            className={cn("size-9 rounded-full font-mono text-sm transition-colors", p === page ? "bg-ink text-bone" : "hover:bg-ink/5")}
          >
            {p}
          </button>
        </Fragment>
      ))}
      <button disabled={page >= pages} onClick={() => go(page + 1)} className={pill("line", "sm")} aria-label="Next page">
        <ArrowRight className="size-4" />
      </button>
    </nav>
  );
}
