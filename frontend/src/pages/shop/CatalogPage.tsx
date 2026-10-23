import { useMemo } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { PackageSearch, ShieldCheck, ShoppingCart, Truck, X } from "lucide-react";
import { categoriesApi, productsApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { useCart } from "@/lib/cart";
import { money } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { ProductListItem } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { EmptyState, ErrorState, Input, Select, Skeleton } from "@/components/ui/primitives";
import { ProductArt } from "@/components/common";

const LOW_STOCK = 10;

type Sort = "newest" | "price-asc" | "price-desc" | "name";

export function CatalogPage() {
  const [params, setParams] = useSearchParams();
  const query = params.get("q") ?? "";
  const categoryId = params.get("category") ?? "";
  const sort = (params.get("sort") as Sort) || "newest";
  const inStockOnly = params.get("stock") === "1";
  const minPrice = params.get("min") ?? "";
  const maxPrice = params.get("max") ?? "";

  const categories = useQuery({ queryKey: ["categories"], queryFn: categoriesApi.list });
  // A selected category is filtered server-side through
  // GET /products/category/:id; search, price and stock filters apply locally.
  const products = useQuery({
    queryKey: ["products", { categoryId }],
    queryFn: () => (categoryId ? productsApi.byCategory(categoryId) : productsApi.list()),
  });

  function update(key: string, value: string) {
    const next = new URLSearchParams(params);
    if (value) next.set(key, value);
    else next.delete(key);
    setParams(next, { replace: true });
  }

  const visible = useMemo(() => {
    let list = products.data ?? [];
    const needle = query.trim().toLowerCase();
    if (needle) {
      list = list.filter(
        (product) =>
          product.name.toLowerCase().includes(needle) ||
          product.sku.toLowerCase().includes(needle) ||
          product.category.toLowerCase().includes(needle),
      );
    }
    if (inStockOnly) list = list.filter((product) => product.available_quantity > 0);
    const min = Number(minPrice);
    const max = Number(maxPrice);
    if (minPrice && !Number.isNaN(min)) list = list.filter((product) => product.price >= min);
    if (maxPrice && !Number.isNaN(max)) list = list.filter((product) => product.price <= max);

    const sorted = [...list];
    if (sort === "price-asc") sorted.sort((a, b) => a.price - b.price);
    if (sort === "price-desc") sorted.sort((a, b) => b.price - a.price);
    if (sort === "name") sorted.sort((a, b) => a.name.localeCompare(b.name));
    return sorted;
  }, [products.data, query, inStockOnly, minPrice, maxPrice, sort]);

  const activeCategory = categories.data?.find((category) => category.id === categoryId);
  const hasFilters = query || categoryId || inStockOnly || minPrice || maxPrice;

  return (
    <div>
      <section className="border-b border-border bg-gradient-to-br from-indigo-50 via-white to-sky-50">
        <div className="mx-auto grid max-w-7xl gap-8 px-4 py-12 sm:px-6 lg:grid-cols-[1.4fr_1fr] lg:items-center">
          <div>
            <p className="text-sm font-semibold text-primary">New season · Fresh stock daily</p>
            <h1 className="mt-2 text-4xl font-semibold tracking-tight sm:text-5xl">Shop smarter with Meridian.</h1>
            <p className="mt-4 max-w-xl text-base text-subtle">
              Live inventory on every product, secure Razorpay checkout and order tracking from payment to your door.
            </p>
            <div className="mt-6 flex flex-wrap gap-6 text-sm text-subtle">
              <span className="flex items-center gap-2">
                <ShieldCheck className="size-4 text-success" /> Secure payments
              </span>
              <span className="flex items-center gap-2">
                <Truck className="size-4 text-info" /> Tracked delivery
              </span>
              <span className="flex items-center gap-2">
                <PackageSearch className="size-4 text-primary" /> Real-time stock
              </span>
            </div>
          </div>
          <div className="hidden grid-cols-3 gap-3 lg:grid">
            {(products.data ?? []).slice(0, 6).map((product, index) => (
              <ProductArt
                key={product.id}
                id={product.id}
                name={product.name}
               
                className={cn("aspect-square rounded-2xl shadow-card", index % 2 === 1 && "translate-y-4")}
              />
            ))}
          </div>
        </div>
      </section>

      <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6">
        {/* Category chips */}
        <div className="-mx-1 mb-6 flex gap-2 overflow-x-auto px-1 pb-1">
          <CategoryChip active={!categoryId} onClick={() => update("category", "")}>
            All products
          </CategoryChip>
          {categories.data?.map((category) => (
            <CategoryChip key={category.id} active={category.id === categoryId} onClick={() => update("category", category.id)}>
              {category.name}
              <span className="ml-1.5 text-xs opacity-60">{category.product_count}</span>
            </CategoryChip>
          ))}
        </div>

        <div className="grid gap-8 lg:grid-cols-[240px_1fr]">
          {/* Filters */}
          <aside className="space-y-6">
            <div>
              <h2 className="mb-3 text-sm font-semibold">Search</h2>
              <Input
                placeholder="Name, SKU or category"
                value={query}
                onChange={(event) => update("q", event.target.value)}
                aria-label="Filter products"
              />
            </div>
            <div>
              <h2 className="mb-3 text-sm font-semibold">Price (₹)</h2>
              <div className="flex items-center gap-2">
                <Input type="number" min={0} placeholder="Min" value={minPrice} onChange={(event) => update("min", event.target.value)} aria-label="Minimum price" />
                <span className="text-faint">–</span>
                <Input type="number" min={0} placeholder="Max" value={maxPrice} onChange={(event) => update("max", event.target.value)} aria-label="Maximum price" />
              </div>
            </div>
            <label className="flex cursor-pointer items-center gap-2.5 text-sm">
              <input
                type="checkbox"
                className="size-4 accent-[var(--color-primary)]"
                checked={inStockOnly}
                onChange={(event) => update("stock", event.target.checked ? "1" : "")}
              />
              In stock only
            </label>
            {hasFilters && (
              <Button variant="secondary" size="sm" onClick={() => setParams({}, { replace: true })}>
                <X /> Clear filters
              </Button>
            )}
          </aside>

          {/* Results */}
          <section>
            <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 className="text-lg font-semibold">{activeCategory ? activeCategory.name : "All products"}</h2>
                <p className="text-sm text-subtle">
                  {products.isLoading ? "Loading…" : `${visible.length} ${visible.length === 1 ? "product" : "products"}`}
                  {activeCategory?.description ? ` · ${activeCategory.description}` : ""}
                </p>
              </div>
              <Select value={sort} onChange={(event) => update("sort", event.target.value)} className="w-48" aria-label="Sort products">
                <option value="newest">Newest first</option>
                <option value="price-asc">Price: low to high</option>
                <option value="price-desc">Price: high to low</option>
                <option value="name">Name A–Z</option>
              </Select>
            </div>

            {products.isError ? (
              <ErrorState message={errorMessage(products.error)} onRetry={() => products.refetch()} />
            ) : products.isLoading ? (
              <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
                {Array.from({ length: 6 }).map((_, index) => (
                  <Skeleton key={index} className="h-80 rounded-xl" />
                ))}
              </div>
            ) : visible.length === 0 ? (
              <EmptyState
                icon={<PackageSearch />}
                title="No products match"
                description="Try a different search or clear the filters."
                action={
                  hasFilters ? (
                    <Button variant="secondary" onClick={() => setParams({}, { replace: true })}>
                      Clear filters
                    </Button>
                  ) : undefined
                }
              />
            ) : (
              <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
                {visible.map((product) => (
                  <ProductCard key={product.id} product={product} />
                ))}
              </div>
            )}
          </section>
        </div>
      </div>
    </div>
  );
}

function CategoryChip({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "flex shrink-0 items-center rounded-full border px-4 py-1.5 text-sm font-medium transition-colors",
        active ? "border-foreground bg-foreground text-white" : "border-border bg-surface text-subtle hover:border-faint hover:text-foreground",
      )}
    >
      {children}
    </button>
  );
}

function ProductCard({ product }: { product: ProductListItem }) {
  const { add, lines } = useCart();
  const inCart = lines.find((line) => line.productId === product.id)?.quantity ?? 0;
  const soldOut = product.available_quantity <= 0;
  const canAdd = inCart < product.available_quantity;

  function addToCart() {
    try {
      add({ productId: product.id, name: product.name, sku: product.sku, price: product.price, category: product.category });
      toast.success("Added to cart", { description: product.name });
    } catch (error) {
      toast.error(errorMessage(error));
    }
  }

  return (
    <article className="group flex flex-col overflow-hidden rounded-xl border border-border bg-surface shadow-card transition-shadow hover:shadow-pop" data-testid="product-card">
      <Link to={`/shop/products/${product.id}`} className="relative block">
        <ProductArt id={product.id} name={product.name} className="aspect-[4/3] transition-transform duration-300 group-hover:scale-[1.02]" />
        {soldOut ? (
          <span className="absolute left-3 top-3 rounded-full bg-foreground/85 px-2.5 py-1 text-xs font-medium text-white">Sold out</span>
        ) : product.available_quantity <= LOW_STOCK ? (
          <span className="absolute left-3 top-3 rounded-full bg-warning px-2.5 py-1 text-xs font-medium text-white">
            Only {product.available_quantity} left
          </span>
        ) : null}
      </Link>
      <div className="flex flex-1 flex-col p-4">
        <p className="text-xs font-medium uppercase tracking-wide text-subtle">{product.category || "Uncategorised"}</p>
        <Link to={`/shop/products/${product.id}`} className="mt-1 line-clamp-2 font-semibold hover:text-primary">
          {product.name}
        </Link>
        <p className="mt-0.5 text-xs text-faint">SKU {product.sku}</p>
        <div className="mt-auto flex items-center justify-between pt-4">
          <span className="tabular text-lg font-semibold">{money(product.price)}</span>
          <Button size="sm" variant={inCart ? "secondary" : "primary"} disabled={soldOut || !canAdd} onClick={addToCart} aria-label={`Add ${product.name} to cart`}>
            <ShoppingCart />
            {soldOut ? "Sold out" : inCart ? `In cart (${inCart})` : "Add"}
          </Button>
        </div>
      </div>
    </article>
  );
}
