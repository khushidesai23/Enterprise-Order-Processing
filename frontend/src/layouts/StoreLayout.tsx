import { useEffect, useState, type FormEvent } from "react";
import { Link, NavLink, Outlet, useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import * as Dialog from "@radix-ui/react-dialog";
import { ChevronRight, Heart, LayoutDashboard, MapPin, Menu, Search, ShoppingBag, ShoppingCart, UserRound, X } from "lucide-react";
import { categoriesApi } from "@/lib/endpoints";
import { useAuth } from "@/lib/auth";
import { useCart } from "@/lib/cart";
import { useWishlist } from "@/lib/saved";
import { cn } from "@/lib/utils";
import { UserMenu } from "@/components/UserMenu";

export function Logo({ dark = false }: { dark?: boolean }) {
  return (
    <span className="flex items-center gap-2">
      <span className="flex size-8 items-center justify-center rounded-lg bg-primary text-white shadow-sm">
        <ShoppingBag className="size-[18px]" />
      </span>
      <span className={cn("text-[15px] font-semibold tracking-tight", dark ? "text-white" : "text-foreground")}>
        Meridian<span className={dark ? "text-amber-300" : "text-primary"}> Market</span>
      </span>
    </span>
  );
}

function useCategories() {
  return useQuery({ queryKey: ["categories"], queryFn: categoriesApi.list, staleTime: 5 * 60_000 });
}

function HeaderSearch() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();
  const categories = useCategories();
  const onSearch = location.pathname === "/shop/search";
  const [value, setValue] = useState(onSearch ? (params.get("q") ?? "") : "");
  const [category, setCategory] = useState(onSearch ? (params.get("category_id") ?? "") : "");

  useEffect(() => {
    if (onSearch) {
      setValue(params.get("q") ?? "");
      setCategory(params.get("category_id") ?? "");
    }
  }, [params, onSearch]);

  function submit(event: FormEvent) {
    event.preventDefault();
    const next = new URLSearchParams();
    if (value.trim()) next.set("q", value.trim());
    if (category) next.set("category_id", category);
    navigate(`/shop/search?${next.toString()}`);
  }

  return (
    <form onSubmit={submit} className="flex h-10 w-full overflow-hidden rounded-md bg-white focus-within:ring-3 focus-within:ring-amber-400" role="search">
      <select
        value={category}
        onChange={(event) => setCategory(event.target.value)}
        aria-label="Search in department"
        className="hidden max-w-36 cursor-pointer border-r border-border bg-muted px-2 text-xs text-subtle focus:outline-none md:block"
      >
        <option value="">All departments</option>
        {categories.data?.map((item) => (
          <option key={item.id} value={item.id}>
            {item.name}
          </option>
        ))}
      </select>
      <input
        value={value}
        onChange={(event) => setValue(event.target.value)}
        placeholder="Search Meridian Market"
        aria-label="Search products"
        className="min-w-0 flex-1 px-3 text-sm text-foreground placeholder:text-faint focus:outline-none"
      />
      <button type="submit" aria-label="Search" className="flex w-12 items-center justify-center bg-amber-400 text-foreground hover:bg-amber-500">
        <Search className="size-5" />
      </button>
    </form>
  );
}

function DepartmentsDrawer() {
  const categories = useCategories();
  const { user, isAdmin } = useAuth();
  const [open, setOpen] = useState(false);
  const location = useLocation();

  useEffect(() => {
    setOpen(false);
  }, [location.pathname, location.search]);

  const link = "flex items-center justify-between px-6 py-2.5 text-sm hover:bg-muted";

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger asChild>
        <button className="flex shrink-0 items-center gap-1.5 rounded-sm px-2 py-1 font-semibold hover:outline hover:outline-1 hover:outline-white">
          <Menu className="size-5" /> All
        </button>
      </Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-black/60" />
        <Dialog.Content className="fixed inset-y-0 left-0 z-50 flex w-80 max-w-[85vw] flex-col bg-surface shadow-pop focus:outline-none">
          <Dialog.Title className="flex items-center gap-3 bg-nav-2 px-6 py-4 text-lg font-semibold text-white">
            <UserRound className="size-6" /> Hello, {user ? user.first_name : "sign in"}
          </Dialog.Title>
          <Dialog.Description className="sr-only">Browse departments</Dialog.Description>
          <Dialog.Close className="absolute -right-11 top-3 text-white" aria-label="Close menu">
            <X className="size-7" />
          </Dialog.Close>
          <nav className="flex-1 overflow-y-auto py-2">
            <p className="px-6 pb-1 pt-3 text-base font-semibold">Trending</p>
            <Link className={link} to="/shop/search?deals=true&sort=discount">Today's Deals</Link>
            <Link className={link} to="/shop/search?sort=bestselling">Best Sellers</Link>
            <Link className={link} to="/shop/search?sort=reviews&min_rating=4">Top Rated</Link>
            <Link className={link} to="/shop/search?sort=newest">New Arrivals</Link>
            <hr className="my-2" />
            <p className="px-6 pb-1 pt-3 text-base font-semibold">Shop by Department</p>
            {categories.data?.map((category) => (
              <Link key={category.id} className={link} to={`/shop/search?category_id=${category.id}`}>
                {category.name} <ChevronRight className="size-4 text-faint" />
              </Link>
            ))}
            <hr className="my-2" />
            <p className="px-6 pb-1 pt-3 text-base font-semibold">Help & Settings</p>
            {user ? (
              <>
                <Link className={link} to="/shop/account">Your Account</Link>
                <Link className={link} to="/shop/orders">Your Orders</Link>
                <Link className={link} to="/shop/wishlist">Your Wishlist</Link>
                {isAdmin && <Link className={link} to="/admin">Admin console</Link>}
              </>
            ) : (
              <Link className={link} to="/login">Sign in</Link>
            )}
          </nav>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

const NAV_LINKS = [
  { to: "/shop/search?deals=true&sort=discount", label: "Today's Deals" },
  { to: "/shop/search?sort=bestselling", label: "Best Sellers" },
  { to: "/shop/search?sort=reviews&min_rating=4", label: "Top Rated" },
  { to: "/shop/search?sort=newest", label: "New Arrivals" },
];

function Header() {
  const { user, isAdmin } = useAuth();
  const { count } = useCart();
  const wishlist = useWishlist();
  const categories = useCategories();
  const box = "flex shrink-0 flex-col justify-center rounded-sm px-2 py-1 leading-tight hover:outline hover:outline-1 hover:outline-white";

  return (
    <header className="sticky top-0 z-40 text-white">
      <div className="bg-nav">
        <div className="mx-auto flex min-h-16 max-w-[1500px] flex-wrap items-center gap-2 px-3 py-2 sm:flex-nowrap sm:gap-3 sm:px-4 sm:py-0">
          <Link to="/shop" aria-label="Meridian Market home" className="shrink-0 rounded-sm p-1 hover:outline hover:outline-1 hover:outline-white">
            <Logo dark />
          </Link>
          <Link to={user ? "/shop/account" : "/login"} className={cn(box, "hidden lg:flex")}>
            <span className="text-xs text-slate-300">Deliver to {user ? user.first_name : "you"}</span>
            <span className="flex items-center gap-1 text-sm font-semibold">
              <MapPin className="size-4" /> India
            </span>
          </Link>
          {/* Full-width second row on phones, inline from sm up. */}
          <div className="order-last w-full min-w-0 sm:order-none sm:w-auto sm:flex-1">
            <HeaderSearch />
          </div>
          {user ? (
            <div className="hidden sm:block">
              <UserMenu variant="dark" greeting />
            </div>
          ) : (
            <Link to="/login" className={cn(box, "hidden sm:flex")}>
              <span className="text-xs text-slate-300">Hello, sign in</span>
              <span className="text-sm font-semibold">Account & Lists</span>
            </Link>
          )}
          <Link to="/shop/orders" className={cn(box, "hidden md:flex")}>
            <span className="text-xs text-slate-300">Returns</span>
            <span className="text-sm font-semibold">& Orders</span>
          </Link>
          {isAdmin && (
            <Link to="/admin" className={cn(box, "hidden xl:flex")} aria-label="Admin console">
              <span className="text-xs text-slate-300">Manage</span>
              <span className="flex items-center gap-1 text-sm font-semibold">
                <LayoutDashboard className="size-4" /> Admin
              </span>
            </Link>
          )}
          <Link to="/shop/wishlist" className={cn(box, "relative hidden items-center sm:flex")} aria-label={`Wishlist, ${wishlist.items.length} items`}>
            <Heart className="size-6" />
            {wishlist.items.length > 0 && (
              <span className="absolute right-0 top-0 flex h-4 min-w-4 items-center justify-center rounded-full bg-amber-400 px-1 text-[10px] font-bold text-foreground">
                {wishlist.items.length}
              </span>
            )}
          </Link>
          <Link to="/shop/cart" className={cn(box, "relative ml-auto flex-row items-end gap-1 sm:ml-0")} aria-label={`Cart, ${count} items`}>
            <span className="relative">
              <ShoppingCart className="size-8" />
              <span data-testid="cart-count" className="tabular absolute -top-1 left-1/2 -translate-x-1/2 text-sm font-bold text-amber-400">
                {count}
              </span>
            </span>
            <span className="hidden text-sm font-semibold sm:inline">Cart</span>
          </Link>
        </div>
      </div>
      <nav className="bg-nav-2">
        <div className="no-scrollbar mx-auto flex h-10 max-w-[1500px] items-center gap-1 overflow-x-auto px-3 text-sm sm:px-4">
          <DepartmentsDrawer />
          {NAV_LINKS.map((item) => (
            <Link key={item.label} to={item.to} className="shrink-0 rounded-sm px-2 py-1 hover:outline hover:outline-1 hover:outline-white">
              {item.label}
            </Link>
          ))}
          {[...(categories.data ?? [])]
            .sort((a, b) => b.product_count - a.product_count)
            .slice(0, 10)
            .map((category) => (
            <NavLink
              key={category.id}
              to={`/shop/search?category_id=${category.id}`}
              className="shrink-0 rounded-sm px-2 py-1 hover:outline hover:outline-1 hover:outline-white"
            >
              {category.name}
            </NavLink>
          ))}
        </div>
      </nav>
    </header>
  );
}

function Footer() {
  const { user } = useAuth();
  const columns = [
    {
      title: "Shop",
      links: [
        ["Today's Deals", "/shop/search?deals=true&sort=discount"],
        ["Best Sellers", "/shop/search?sort=bestselling"],
        ["Top Rated", "/shop/search?sort=reviews&min_rating=4"],
        ["New Arrivals", "/shop/search?sort=newest"],
      ],
    },
    {
      title: "Your Account",
      links: user
        ? [
            ["Your Account", "/shop/account"],
            ["Your Orders", "/shop/orders"],
            ["Your Wishlist", "/shop/wishlist"],
            ["Your Cart", "/shop/cart"],
          ]
        : [
            ["Sign in", "/login"],
            ["Create an account", "/register"],
            ["Your Cart", "/shop/cart"],
          ],
    },
    {
      title: "Let Us Help You",
      links: [
        ["Track an order", "/shop/orders"],
        ["Payments: UPI, cards, netbanking", "/shop/cart"],
        ["Free delivery over ₹499", "/shop/search?min_price=499"],
      ],
    },
  ];

  return (
    <footer className="mt-10 text-sm">
      <button
        onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
        className="block w-full bg-nav-hover py-3.5 text-center text-white hover:bg-nav-2"
      >
        Back to top
      </button>
      <div className="bg-nav-2 text-slate-300">
        <div className="mx-auto grid max-w-5xl gap-8 px-6 py-10 sm:grid-cols-3">
          {columns.map((column) => (
            <div key={column.title}>
              <p className="mb-3 font-semibold text-white">{column.title}</p>
              <ul className="space-y-2">
                {column.links.map(([label, to]) => (
                  <li key={label}>
                    <Link to={to} className="hover:underline">
                      {label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>
      <div className="bg-nav py-6 text-center text-xs text-slate-400">
        <div className="mb-2 flex justify-center">
          <Logo dark />
        </div>
        Secure payments by Razorpay · Prices in INR, inclusive of taxes · Demo store, product data from DummyJSON
      </div>
    </footer>
  );
}

export function StoreLayout() {
  const location = useLocation();

  // Braces matter: newer browsers return a Promise from scrollTo, which
  // React would otherwise call as the effect cleanup.
  useEffect(() => {
    window.scrollTo(0, 0);
  }, [location.pathname]);

  return (
    <div className="flex min-h-screen flex-col bg-[#eaeded]">
      <Header />
      <main className="flex-1">
        <Outlet />
      </main>
      <Footer />
    </div>
  );
}
