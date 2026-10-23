import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useQueries, useQuery } from "@tanstack/react-query";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { categoriesApi, productsApi } from "@/lib/endpoints";
import { recentlyViewed } from "@/lib/saved";
import type { CategoryListItem, SearchParams } from "@/lib/types";
import { cn } from "@/lib/utils";
import { ProductImage, ProductRail, distinctProducts } from "@/components/shop";
import { Skeleton } from "@/components/ui/primitives";

function useSearch(key: string, params: SearchParams) {
  return useQuery({ queryKey: ["search", key, params], queryFn: () => productsApi.search(params), staleTime: 60_000 });
}

const SLIDES = [
  {
    title: "Mega Deals Week",
    text: "Up to 50% off across smartphones, laptops and accessories.",
    cta: "Shop deals",
    to: "/shop/search?deals=true&sort=discount",
    tone: "from-indigo-700 via-violet-600 to-fuchsia-500",
    params: { deals: true, sort: "discount", limit: 24 } as SearchParams,
  },
  {
    title: "Best sellers of the season",
    text: "What thousands of customers are buying right now.",
    cta: "See best sellers",
    to: "/shop/search?sort=bestselling",
    tone: "from-emerald-700 via-teal-600 to-cyan-500",
    params: { sort: "bestselling", limit: 24 } as SearchParams,
  },
  {
    title: "Loved by customers",
    text: "Highly rated picks with hundreds of verified reviews.",
    cta: "Explore top rated",
    to: "/shop/search?sort=reviews&min_rating=4",
    tone: "from-amber-600 via-orange-500 to-rose-500",
    params: { sort: "reviews", min_rating: 4, limit: 24 } as SearchParams,
  },
];

function Hero() {
  const [index, setIndex] = useState(0);
  const slides = useQueries({
    queries: SLIDES.map((slide, i) => ({
      queryKey: ["search", "hero", i],
      queryFn: () => productsApi.search(slide.params),
      staleTime: 60_000,
    })),
  });

  useEffect(() => {
    const timer = window.setInterval(() => setIndex((current) => (current + 1) % SLIDES.length), 6000);
    return () => window.clearInterval(timer);
  }, []);

  const slide = SLIDES[index]!;
  const products = distinctProducts(slides[index]?.data?.items, 3) ?? [];

  return (
    <div className="relative">
      <div className={cn("bg-gradient-to-r transition-colors duration-700", slide.tone)}>
        <div className="mx-auto flex h-72 max-w-[1500px] items-center gap-8 px-6 sm:h-80 sm:px-16">
          <div className="max-w-md text-white">
            <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">{slide.title}</h1>
            <p className="mt-3 text-white/90">{slide.text}</p>
            <Link
              to={slide.to}
              className="mt-6 inline-flex rounded-full bg-white px-5 py-2.5 text-sm font-semibold text-foreground shadow hover:bg-amber-100"
            >
              {slide.cta}
            </Link>
          </div>
          <div className="ml-auto hidden gap-4 md:flex">
            {products.map((product, i) => (
              <Link
                key={product.id}
                to={`/shop/products/${product.id}`}
                className={cn("block rounded-xl bg-white p-3 shadow-pop transition-transform hover:-translate-y-1", i === 1 && "-translate-y-4")}
              >
                <ProductImage id={product.id} name={product.name} src={product.image_url} className="size-36 lg:size-44" />
              </Link>
            ))}
          </div>
        </div>
      </div>
      <div className="absolute inset-x-0 bottom-0 h-24 bg-gradient-to-t from-[#eaeded] to-transparent" />
      <button
        onClick={() => setIndex((index + SLIDES.length - 1) % SLIDES.length)}
        className="absolute left-2 top-24 hidden rounded-md p-3 sm:block text-white/90 hover:bg-white/10"
        aria-label="Previous banner"
      >
        <ChevronLeft className="size-8" />
      </button>
      <button
        onClick={() => setIndex((index + 1) % SLIDES.length)}
        className="absolute right-2 top-24 hidden rounded-md p-3 sm:block text-white/90 hover:bg-white/10"
        aria-label="Next banner"
      >
        <ChevronRight className="size-8" />
      </button>
    </div>
  );
}

// An Amazon-style card: a department with four of its best sellers.
function DepartmentCard({ category }: { category: CategoryListItem }) {
  const products = useSearch(`dept-${category.id}`, { category_id: category.id, sort: "bestselling", limit: 24 });
  const href = `/shop/search?category_id=${category.id}`;

  return (
    <div className="flex flex-col bg-surface p-5 shadow-card">
      <h2 className="mb-3 text-lg font-semibold leading-tight">{category.name}</h2>
      <div className="grid flex-1 grid-cols-2 gap-3">
        {products.isLoading
          ? Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="aspect-square" />)
          : distinctProducts(products.data?.items, 4)?.map((product) => (
              <Link key={product.id} to={`/shop/products/${product.id}`} className="group">
                <ProductImage id={product.id} name={product.name} src={product.image_url} size="sm" className="aspect-square bg-muted/40" imgClassName="p-1 transition-transform group-hover:scale-105" />
                <p className="mt-1 truncate text-xs text-subtle">{product.name}</p>
              </Link>
            ))}
      </div>
      <Link to={href} className="mt-4 text-sm text-link hover:text-deal hover:underline">
        See more in {category.name}
      </Link>
    </div>
  );
}

export function HomePage() {
  const categories = useQuery({ queryKey: ["categories"], queryFn: categoriesApi.list, staleTime: 5 * 60_000 });
  const deals = useSearch("deals", { deals: true, sort: "discount", in_stock: true, limit: 60 });
  const bestsellers = useSearch("bestsellers", { sort: "bestselling", limit: 60 });
  const topRated = useSearch("top-rated", { sort: "reviews", min_rating: 4.2, limit: 60 });
  const newest = useSearch("newest", { sort: "newest", limit: 60 });
  const [recent] = useState(recentlyViewed);

  // Departments with the most products lead the home page.
  const featured = useMemo(
    () =>
      [...(categories.data ?? [])]
        .filter((category) => category.product_count >= 4)
        .sort((a, b) => b.product_count - a.product_count),
    [categories.data],
  );
  const phones = featured.find((category) => category.name === "Smartphones");
  const phoneRail = useQuery({
    queryKey: ["search", "phones", phones?.id],
    queryFn: () => productsApi.search({ category_id: phones!.id, sort: "bestselling", limit: 60 }),
    enabled: !!phones,
    staleTime: 60_000,
  });

  return (
    <div>
      <Hero />
      <div className="relative mx-auto -mt-16 max-w-[1500px] space-y-5 px-3 pb-6 sm:px-4">
        <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
          {categories.isLoading
            ? Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-96" />)
            : featured.slice(0, 4).map((category) => <DepartmentCard key={category.id} category={category} />)}
        </div>

        <ProductRail
          title="Today's Deals"
          action={<Link to="/shop/search?deals=true&sort=discount" className="text-sm text-link hover:underline">See all deals</Link>}
          products={distinctProducts(deals.data?.items, 18)}
          loading={deals.isLoading}
        />

        <ProductRail
          title="Best Sellers"
          action={<Link to="/shop/search?sort=bestselling" className="text-sm text-link hover:underline">See more</Link>}
          products={distinctProducts(bestsellers.data?.items, 18)}
          loading={bestsellers.isLoading}
        />

        {recent.length > 0 && <ProductRail title="Your browsing history" products={recent} />}

        <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
          {featured.slice(4, 8).map((category) => (
            <DepartmentCard key={category.id} category={category} />
          ))}
        </div>

        <ProductRail
          title="Top rated by customers"
          action={<Link to="/shop/search?sort=reviews&min_rating=4" className="text-sm text-link hover:underline">See more</Link>}
          products={distinctProducts(topRated.data?.items, 18)}
          loading={topRated.isLoading}
        />

        {phones && (
          <ProductRail
            title="Best sellers in Smartphones"
            action={<Link to={`/shop/search?category_id=${phones.id}`} className="text-sm text-link hover:underline">Shop Smartphones</Link>}
            products={distinctProducts(phoneRail.data?.items, 18)}
            loading={phoneRail.isLoading}
          />
        )}

        <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
          {featured.slice(8, 12).map((category) => (
            <DepartmentCard key={category.id} category={category} />
          ))}
        </div>

        <ProductRail
          title="New arrivals"
          action={<Link to="/shop/search?sort=newest" className="text-sm text-link hover:underline">See more</Link>}
          products={distinctProducts(newest.data?.items, 18)}
          loading={newest.isLoading}
        />
      </div>
    </div>
  );
}
