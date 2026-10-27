import { useState, type FormEvent, type ReactNode } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { ChevronLeft, ChevronRight, PackageSearch, X } from "lucide-react";
import { productsApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { money, number } from "@/lib/format";
import type { SearchParams, SearchSort } from "@/lib/types";
import { cn } from "@/lib/utils";
import { ProductCard, Stars } from "@/components/shop";
import { Button } from "@/components/ui/button";
import { EmptyState, ErrorState, Skeleton } from "@/components/ui/primitives";

const PAGE_SIZE = 24;

const SORTS: { value: SearchSort; label: string }[] = [
  { value: "featured", label: "Featured" },
  { value: "price_asc", label: "Price: Low to High" },
  { value: "price_desc", label: "Price: High to Low" },
  { value: "rating", label: "Avg. Customer Review" },
  { value: "reviews", label: "Most Reviewed" },
  { value: "bestselling", label: "Best Sellers" },
  { value: "discount", label: "Biggest Discount" },
  { value: "newest", label: "Newest Arrivals" },
];

const PRICE_RANGES: { label: string; min?: number; max?: number }[] = [
  { label: "Under ₹500", max: 500 },
  { label: "₹500 – ₹1,000", min: 500, max: 1000 },
  { label: "₹1,000 – ₹5,000", min: 1000, max: 5000 },
  { label: "₹5,000 – ₹20,000", min: 5000, max: 20000 },
  { label: "₹20,000 – ₹1,00,000", min: 20000, max: 100000 },
  { label: "Over ₹1,00,000", min: 100000 },
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

function FilterSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="border-b border-border pb-4">
      <h3 className="mb-2 text-sm font-bold">{title}</h3>
      {children}
    </div>
  );
}

export function SearchPage() {
  const [params, setParams] = useSearchParams();
  const search = readParams(params);
  const [showAllBrands, setShowAllBrands] = useState(false);
  const [minDraft, setMinDraft] = useState(search.min_price?.toString() ?? "");
  const [maxDraft, setMaxDraft] = useState(search.max_price?.toString() ?? "");

  const result = useQuery({
    queryKey: ["search", "page", params.toString()],
    queryFn: () => productsApi.search(search),
    placeholderData: keepPreviousData,
  });

  // Every filter change returns to page 1.
  function update(changes: Record<string, string | string[] | undefined>) {
    const next = new URLSearchParams(params);
    for (const [key, value] of Object.entries(changes)) {
      next.delete(key);
      if (Array.isArray(value)) value.forEach((item) => next.append(key, item));
      else if (value !== undefined && value !== "") next.set(key, value);
    }
    if (!("page" in changes)) next.delete("page");
    setParams(next);
  }

  function toggleBrand(brand: string) {
    const brands = search.brand ?? [];
    update({ brand: brands.includes(brand) ? brands.filter((item) => item !== brand) : [...brands, brand] });
  }

  function applyPrice(event: FormEvent) {
    event.preventDefault();
    update({ min_price: minDraft, max_price: maxDraft });
  }

  const data = result.data;
  const facets = data?.facets;
  const category = facets?.categories.find((item) => item.id === search.category_id);
  const brands = facets?.brands ?? [];
  const visibleBrands = showAllBrands ? brands : brands.slice(0, 8);
  const pages = data ? Math.max(1, Math.ceil(data.total / PAGE_SIZE)) : 1;
  const page = search.page ?? 1;
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
  if (search.deals) chips.push({ label: "Deals", clear: { deals: undefined } });
  if (search.in_stock) chips.push({ label: "In stock", clear: { in_stock: undefined } });

  const heading = search.q ? (
    <>
      for <span className="font-semibold text-deal">"{search.q}"</span>
    </>
  ) : category ? (
    <>in {category.name}</>
  ) : search.deals ? (
    <>in Today's Deals</>
  ) : null;

  return (
    <div className="bg-surface">
      <div className="border-b border-border shadow-sm">
        <div className="mx-auto flex max-w-[1500px] flex-wrap items-center justify-between gap-3 px-4 py-2.5 text-sm">
          <p data-testid="result-summary">
            {data ? (
              <>
                {number(first)}-{number(last)} of {number(data.total)} results {heading}
              </>
            ) : (
              "Searching…"
            )}
          </p>
          <label className="flex items-center gap-2">
            <span className="text-subtle">Sort by:</span>
            <select
              value={search.sort ?? (search.q ? "relevance" : "featured")}
              onChange={(event) => update({ sort: event.target.value })}
              className="cursor-pointer rounded-lg border border-border bg-muted px-2 py-1 text-sm shadow-sm"
              aria-label="Sort results"
            >
              {search.q && <option value="relevance">Relevance</option>}
              {SORTS.map((sort) => (
                <option key={sort.value} value={sort.value}>
                  {sort.label}
                </option>
              ))}
            </select>
          </label>
        </div>
      </div>

      <div className="mx-auto grid max-w-[1500px] gap-6 px-4 py-5 lg:grid-cols-[250px_1fr]">
        <aside className="space-y-4 text-sm" aria-label="Filters">
          <FilterSection title="Department">
            {search.category_id && (
              <button onClick={() => update({ category_id: undefined })} className="mb-1 flex items-center gap-1 text-subtle hover:text-deal">
                <ChevronLeft className="size-3.5" /> Any department
              </button>
            )}
            <ul className="space-y-1">
              {(facets?.categories ?? []).map((item) => (
                <li key={item.id}>
                  <button
                    onClick={() => update({ category_id: item.id })}
                    className={cn("text-left hover:text-deal", item.id === search.category_id && "font-bold")}
                  >
                    {item.name} <span className="text-faint">({number(item.count)})</span>
                  </button>
                </li>
              ))}
            </ul>
          </FilterSection>

          <FilterSection title="Customer Reviews">
            <ul className="space-y-1.5">
              {[4, 3, 2, 1].map((stars) => (
                <li key={stars}>
                  <button
                    onClick={() => update({ min_rating: search.min_rating === stars ? undefined : String(stars) })}
                    className={cn("flex items-center gap-1.5 hover:text-deal", search.min_rating === stars && "font-bold")}
                    aria-label={`${stars} stars and up`}
                  >
                    <Stars value={stars} size={16} /> & Up
                  </button>
                </li>
              ))}
            </ul>
          </FilterSection>

          {brands.length > 0 && (
            <FilterSection title="Brands">
              <ul className="space-y-1.5">
                {visibleBrands.map((brand) => (
                  <li key={brand.name}>
                    <label className="flex cursor-pointer items-center gap-2 hover:text-deal">
                      <input
                        type="checkbox"
                        className="size-4 accent-[var(--color-link)]"
                        checked={(search.brand ?? []).includes(brand.name)}
                        onChange={() => toggleBrand(brand.name)}
                      />
                      {brand.name} <span className="text-faint">({brand.count})</span>
                    </label>
                  </li>
                ))}
              </ul>
              {brands.length > 8 && (
                <button onClick={() => setShowAllBrands(!showAllBrands)} className="mt-2 text-link hover:underline">
                  {showAllBrands ? "See less" : `See all ${brands.length} brands`}
                </button>
              )}
            </FilterSection>
          )}

          <FilterSection title="Price">
            <ul className="space-y-1">
              {PRICE_RANGES.map((range) => {
                const active = search.min_price === range.min && search.max_price === range.max;
                return (
                  <li key={range.label}>
                    <button
                      onClick={() =>
                        active
                          ? update({ min_price: undefined, max_price: undefined })
                          : update({ min_price: range.min?.toString(), max_price: range.max?.toString() })
                      }
                      className={cn("hover:text-deal", active && "font-bold")}
                    >
                      {range.label}
                    </button>
                  </li>
                );
              })}
            </ul>
            <form onSubmit={applyPrice} className="mt-2 flex items-center gap-1.5">
              <input
                value={minDraft}
                onChange={(event) => setMinDraft(event.target.value)}
                inputMode="numeric"
                placeholder="₹ Min"
                aria-label="Minimum price"
                className="h-8 w-20 rounded-md border border-border px-2"
              />
              <input
                value={maxDraft}
                onChange={(event) => setMaxDraft(event.target.value)}
                inputMode="numeric"
                placeholder="₹ Max"
                aria-label="Maximum price"
                className="h-8 w-20 rounded-md border border-border px-2"
              />
              <Button type="submit" size="sm" variant="secondary">
                Go
              </Button>
            </form>
            {facets && facets.price_max > 0 && (
              <p className="mt-1 text-xs text-faint">
                {money(facets.price_min)} – {money(facets.price_max)} in these results
              </p>
            )}
          </FilterSection>

          <FilterSection title="Deals & Availability">
            <label className="flex cursor-pointer items-center gap-2 py-0.5">
              <input
                type="checkbox"
                className="size-4 accent-[var(--color-link)]"
                checked={!!search.deals}
                onChange={(event) => update({ deals: event.target.checked ? "true" : undefined })}
              />
              Today's Deals
            </label>
            <label className="flex cursor-pointer items-center gap-2 py-0.5">
              <input
                type="checkbox"
                className="size-4 accent-[var(--color-link)]"
                checked={!!search.in_stock}
                onChange={(event) => update({ in_stock: event.target.checked ? "true" : undefined })}
              />
              Exclude out of stock
            </label>
          </FilterSection>
        </aside>

        <section>
          {chips.length > 0 && (
            <div className="mb-4 flex flex-wrap items-center gap-2">
              {chips.map((chip) => (
                <button
                  key={chip.label}
                  onClick={() => update(chip.clear)}
                  className="flex items-center gap-1 rounded-full border border-border bg-muted px-3 py-1 text-xs hover:border-faint"
                  aria-label={`Remove filter ${chip.label}`}
                >
                  {chip.label} <X className="size-3" />
                </button>
              ))}
              <Link to="/shop/search" className="text-xs text-link hover:underline">
                Clear all
              </Link>
            </div>
          )}

          <h1 className="mb-1 text-xl font-bold">Results</h1>
          <p className="mb-4 text-sm text-subtle">Check each product page for other buying options.</p>

          {result.isError ? (
            <ErrorState message={errorMessage(result.error)} onRetry={() => result.refetch()} />
          ) : result.isLoading ? (
            <div className="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-4">
              {Array.from({ length: 8 }).map((_, i) => (
                <Skeleton key={i} className="h-96" />
              ))}
            </div>
          ) : data && data.items.length === 0 ? (
            <EmptyState
              icon={<PackageSearch />}
              title="No results"
              description="Try checking your spelling, using more general words, or removing filters."
              action={
                <Link to="/shop/search">
                  <Button variant="secondary">Clear all filters</Button>
                </Link>
              }
            />
          ) : (
            <>
              <div className={cn("grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-4", result.isPlaceholderData && "opacity-60")}>
                {data?.items.map((product) => <ProductCard key={product.id} product={product} />)}
              </div>
              <Pager page={page} pages={pages} onChange={(next) => update({ page: String(next) })} />
            </>
          )}
        </section>
      </div>
    </div>
  );
}

function Pager({ page, pages, onChange }: { page: number; pages: number; onChange: (page: number) => void }) {
  if (pages <= 1) return null;

  const shown = new Set([1, pages, page - 1, page, page + 1].filter((p) => p >= 1 && p <= pages));
  const sorted = [...shown].sort((a, b) => a - b);
  const go = (next: number) => {
    onChange(next);
    window.scrollTo({ top: 0 });
  };

  return (
    <nav className="mt-8 flex justify-center" aria-label="Pagination">
      <div className="flex items-center overflow-hidden rounded-lg border border-border bg-surface text-sm shadow-sm">
        <button disabled={page <= 1} onClick={() => go(page - 1)} className="flex items-center gap-1 px-3 py-2 hover:bg-muted disabled:opacity-40">
          <ChevronLeft className="size-4" /> Previous
        </button>
        {sorted.map((p, i) => (
          <span key={p} className="flex">
            {i > 0 && p - sorted[i - 1]! > 1 && <span className="px-3 py-2 text-faint">…</span>}
            <button
              onClick={() => go(p)}
              aria-current={p === page ? "page" : undefined}
              className={cn("min-w-10 px-3 py-2", p === page ? "border-x border-foreground font-semibold" : "hover:bg-muted")}
            >
              {p}
            </button>
          </span>
        ))}
        <button disabled={page >= pages} onClick={() => go(page + 1)} className="flex items-center gap-1 px-3 py-2 hover:bg-muted disabled:opacity-40">
          Next <ChevronRight className="size-4" />
        </button>
      </div>
    </nav>
  );
}
