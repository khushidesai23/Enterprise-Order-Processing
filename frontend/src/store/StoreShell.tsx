import { useEffect, useMemo, useRef, useState, type FormEvent, type KeyboardEvent, type MouseEvent } from "react";
import { Link, useLocation, useNavigate, useOutlet } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import * as Dialog from "@radix-ui/react-dialog";
import * as Dropdown from "@radix-ui/react-dropdown-menu";
import { AnimatePresence, motion, useMotionValue, useReducedMotion, useSpring } from "motion/react";
import { ArrowRight, ArrowUpRight, Heart, LayoutDashboard, LogOut, Menu, Package, Search, ShoppingBag, UserRound, X } from "lucide-react";
import { categoriesApi, healthApi, productsApi } from "@/lib/endpoints";
import { useAuth } from "@/lib/auth";
import { useCart } from "@/lib/cart";
import { money, number } from "@/lib/format";
import { useWishlist } from "@/lib/saved";
import type { CategoryListItem } from "@/lib/types";
import { cn } from "@/lib/utils";
import { BagProvider, useBag } from "./bag";
import { Marquee, ProductShot, Wordmark, distinct, ease, pill } from "./ui";

function useDepartments() {
  return useQuery({
    queryKey: ["categories"],
    queryFn: categoriesApi.list,
    staleTime: 5 * 60_000,
    select: (categories: CategoryListItem[]) =>
      [...categories].filter((category) => category.product_count > 0).sort((a, b) => b.product_count - a.product_count),
  });
}

// --- Search palette -------------------------------------------------------------------

function useDebounced<T>(value: T, delay = 220) {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const timer = window.setTimeout(() => setDebounced(value), delay);
    return () => window.clearTimeout(timer);
  }, [value, delay]);
  return debounced;
}

function SearchPalette({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const navigate = useNavigate();
  const departments = useDepartments();
  const [value, setValue] = useState("");
  const [active, setActive] = useState(-1);
  const query = useDebounced(value.trim());
  const results = useQuery({
    queryKey: ["search", "palette", query],
    queryFn: () => productsApi.search({ q: query, limit: 12 }),
    enabled: open && query.length > 0,
    staleTime: 30_000,
  });
  const trending = useQuery({
    queryKey: ["search", "palette-trending"],
    queryFn: () => productsApi.search({ sort: "bestselling", limit: 24 }),
    enabled: open,
    staleTime: 60_000,
  });
  const items = distinct(query ? results.data?.items : trending.data?.items, 6);

  useEffect(() => {
    if (!open) {
      setValue("");
      setActive(-1);
    }
  }, [open]);

  function go(path: string) {
    onOpenChange(false);
    navigate(path);
  }

  function submit(event: FormEvent) {
    event.preventDefault();
    if (active >= 0 && items[active]) return go(`/shop/products/${items[active].id}`);
    go(value.trim() ? `/shop/search?q=${encodeURIComponent(value.trim())}` : "/shop/search");
  }

  function onKeyDown(event: KeyboardEvent) {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActive((index) => Math.min(index + 1, items.length - 1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActive((index) => Math.max(index - 1, -1));
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open && (
          <Dialog.Portal forceMount>
            <Dialog.Overlay asChild forceMount>
              <motion.div className="fixed inset-0 z-50 bg-ink/50 backdrop-blur-sm" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} />
            </Dialog.Overlay>
            <Dialog.Content asChild forceMount aria-describedby={undefined}>
              <motion.div
                className="store fixed left-1/2 top-[8vh] z-50 w-[min(720px,calc(100vw-1.5rem))] -translate-x-1/2 overflow-hidden rounded-3xl bg-bone shadow-2xl focus:outline-none"
                initial={{ opacity: 0, y: -16, scale: 0.98 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -10, scale: 0.98 }}
                transition={{ duration: 0.35, ease }}
              >
                <Dialog.Title className="sr-only">Search the index</Dialog.Title>
                <form role="search" onSubmit={submit} className="flex items-center gap-3 border-b border-ink/10 px-5">
                  <Search className="size-5 shrink-0 text-smoke" />
                  <input
                    autoFocus
                    value={value}
                    onChange={(event) => {
                      setValue(event.target.value);
                      setActive(-1);
                    }}
                    onKeyDown={onKeyDown}
                    placeholder="Search 800+ objects, brands, departments"
                    aria-label="Search products"
                    className="h-16 min-w-0 flex-1 bg-transparent focus-visible:outline-none font-display text-xl font-semibold tracking-tight placeholder:font-store placeholder:text-base placeholder:font-normal placeholder:text-smoke focus:outline-none"
                  />
                  <button type="submit" aria-label="Search" className={pill("ink", "sm")}>
                    Search <ArrowRight className="size-4" />
                  </button>
                </form>

                <div className="max-h-[65vh] overflow-y-auto p-3">
                  <p className="px-3 pb-2 pt-1 font-mono text-[11px] uppercase tracking-[0.18em] text-smoke">
                    {query ? (results.isFetching ? "Searching…" : `${number(results.data?.total ?? 0)} matches`) : "Trending now"}
                  </p>
                  <ul>
                    {items.map((item, index) => (
                      <li key={item.id}>
                        <button
                          type="button"
                          onMouseEnter={() => setActive(index)}
                          onClick={() => go(`/shop/products/${item.id}`)}
                          className={cn(
                            "flex w-full items-center gap-4 rounded-2xl px-3 py-2 text-left transition-colors",
                            active === index ? "bg-ink text-bone" : "hover:bg-ink/5",
                          )}
                        >
                          <ProductShot id={item.id} name={item.name} src={item.image_url} size="sm" className="size-12 shrink-0 rounded-xl bg-bone-2 p-1" />
                          <span className="min-w-0 flex-1">
                            <span className="block truncate text-sm font-medium">{item.name}</span>
                            <span className={cn("font-mono text-[11px] uppercase tracking-wider", active === index ? "text-bone/60" : "text-smoke")}>
                              {item.brand || item.category}
                            </span>
                          </span>
                          <span className="tabular font-mono text-sm">{money(item.price)}</span>
                        </button>
                      </li>
                    ))}
                  </ul>
                  {query && results.data?.total === 0 && (
                    <p className="px-3 py-6 text-smoke">
                      Nothing matches <span className="font-serif text-lg italic text-ink">“{query}”</span>. Try a brand or department.
                    </p>
                  )}
                  <p className="mt-3 px-3 pb-2 font-mono text-[11px] uppercase tracking-[0.18em] text-smoke">Departments</p>
                  <div className="flex flex-wrap gap-2 px-3 pb-2">
                    {departments.data?.slice(0, 12).map((department) => (
                      <button
                        key={department.id}
                        type="button"
                        onClick={() => go(`/shop/search?category_id=${department.id}`)}
                        className="rounded-full border border-ink/15 px-3 py-1.5 text-sm transition-colors hover:border-ink hover:bg-ink hover:text-bone"
                      >
                        {department.name}
                      </button>
                    ))}
                  </div>
                </div>
              </motion.div>
            </Dialog.Content>
          </Dialog.Portal>
        )}
      </AnimatePresence>
    </Dialog.Root>
  );
}

// --- Departments index (full-screen menu) -------------------------------------------

// The best seller of a department, for the hover preview.
function DepartmentPreview({ id }: { id: string | null }) {
  const preview = useQuery({
    queryKey: ["search", "dept-preview", id],
    queryFn: () => productsApi.search({ category_id: id!, sort: "bestselling", limit: 1 }),
    enabled: !!id,
    staleTime: 5 * 60_000,
  });
  const product = preview.data?.items[0];
  if (!product) return null;
  return <ProductShot key={product.id} id={product.id} name={product.name} src={product.image_url} eager className="size-full" />;
}

export function DepartmentIndex({ onNavigate, inverse = true }: { onNavigate?: () => void; inverse?: boolean }) {
  const departments = useDepartments();
  const reduce = useReducedMotion();
  const [hovered, setHovered] = useState<string | null>(null);
  const x = useSpring(useMotionValue(0), { stiffness: 260, damping: 28 });
  const y = useSpring(useMotionValue(0), { stiffness: 260, damping: 28 });
  const list = useRef<HTMLDivElement>(null);

  function track(event: MouseEvent) {
    const rect = list.current?.getBoundingClientRect();
    if (!rect) return;
    x.set(event.clientX - rect.left);
    y.set(event.clientY - rect.top);
  }

  return (
    <div ref={list} className="relative" onMouseMove={track} onMouseLeave={() => setHovered(null)}>
      <ol>
        {departments.data?.map((department, index) => (
          <li key={department.id} className={cn("border-b", inverse ? "border-bone/15" : "border-ink/15")}>
            <Link
              to={`/shop/search?category_id=${department.id}`}
              onClick={onNavigate}
              onMouseEnter={() => setHovered(department.id)}
              onFocus={() => setHovered(department.id)}
              className="group flex items-baseline gap-4 py-3 sm:gap-8 sm:py-4"
            >
              <span className={cn("w-8 font-mono text-xs", inverse ? "text-bone/50" : "text-smoke")}>{String(index + 1).padStart(2, "0")}</span>
              <span
                className={cn(
                  "font-display text-3xl font-bold tracking-tight transition-[transform,color] duration-300 group-hover:translate-x-3 sm:text-5xl",
                  inverse ? "group-hover:text-acid" : "group-hover:text-ultra",
                )}
              >
                {department.name}
              </span>
              <span className={cn("ml-auto font-mono text-xs", inverse ? "text-bone/50" : "text-smoke")}>{department.product_count}</span>
              <ArrowUpRight className={cn("size-5 opacity-0 transition-opacity group-hover:opacity-100", inverse ? "text-acid" : "text-ultra")} />
            </Link>
          </li>
        ))}
      </ol>
      {!reduce && (
        <motion.div
          aria-hidden
          className="pointer-events-none absolute left-0 top-0 z-10 hidden size-44 -translate-x-1/2 -translate-y-1/2 overflow-hidden rounded-3xl bg-bone p-3 shadow-2xl lg:block"
          style={{ x, y }}
          animate={{ opacity: hovered ? 1 : 0, scale: hovered ? 1 : 0.6, rotate: hovered ? -4 : 0 }}
          transition={{ duration: 0.3, ease }}
        >
          <DepartmentPreview id={hovered} />
        </motion.div>
      )}
    </div>
  );
}

function MenuOverlay({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const { user, isAdmin } = useAuth();
  const close = () => onOpenChange(false);
  const links = [
    ["The index", "/shop/search"],
    ["Drops", "/shop/search?deals=true&sort=discount"],
    ["Most wanted", "/shop/search?sort=bestselling"],
    ["Highly rated", "/shop/search?sort=reviews&min_rating=4"],
  ];

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open && (
          <Dialog.Portal forceMount>
            <Dialog.Content asChild forceMount aria-describedby={undefined}>
              <motion.div
                className="store fixed inset-0 z-50 overflow-y-auto bg-ink text-bone focus:outline-none"
                initial={{ clipPath: "inset(0 0 100% 0)" }}
                animate={{ clipPath: "inset(0 0 0% 0)" }}
                exit={{ clipPath: "inset(0 0 100% 0)" }}
                transition={{ duration: 0.65, ease }}
              >
                <div className="mx-auto flex max-w-[1600px] items-center justify-between px-5 py-5 sm:px-8">
                  <Link to="/shop" onClick={close} aria-label="meridian/ home">
                    <Wordmark inverse />
                  </Link>
                  <Dialog.Title className="sr-only">Menu</Dialog.Title>
                  <Dialog.Close className={pill("bone", "sm")} aria-label="Close menu">
                    Close <X className="size-4" />
                  </Dialog.Close>
                </div>
                <div className="mx-auto grid max-w-[1600px] gap-10 px-5 pb-16 pt-6 sm:px-8 lg:grid-cols-[1fr_2.2fr]">
                  <motion.nav
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.25, duration: 0.6, ease }}
                    className="space-y-8"
                  >
                    <div>
                      <p className="mb-3 font-mono text-[11px] uppercase tracking-[0.18em] text-bone/50">Explore</p>
                      <ul className="space-y-1">
                        {links.map(([label, to]) => (
                          <li key={label}>
                            <Link to={to} onClick={close} className="font-serif text-4xl italic transition-colors hover:text-acid">
                              {label}
                            </Link>
                          </li>
                        ))}
                      </ul>
                    </div>
                    <div>
                      <p className="mb-3 font-mono text-[11px] uppercase tracking-[0.18em] text-bone/50">You</p>
                      <ul className="space-y-2 text-lg">
                        {user ? (
                          <>
                            <li><Link to="/shop/account" onClick={close} className="hover:text-acid">Account</Link></li>
                            <li><Link to="/shop/orders" onClick={close} className="hover:text-acid">Orders</Link></li>
                            <li><Link to="/shop/wishlist" onClick={close} className="hover:text-acid">Wishlist</Link></li>
                            {isAdmin && <li><Link to="/admin" onClick={close} className="hover:text-acid">Admin console</Link></li>}
                          </>
                        ) : (
                          <>
                            <li><Link to="/login" onClick={close} className="hover:text-acid">Sign in</Link></li>
                            <li><Link to="/register" onClick={close} className="hover:text-acid">Create an account</Link></li>
                            <li><Link to="/shop/wishlist" onClick={close} className="hover:text-acid">Wishlist</Link></li>
                          </>
                        )}
                      </ul>
                    </div>
                  </motion.nav>
                  <motion.div initial={{ opacity: 0, y: 30 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.35, duration: 0.7, ease }}>
                    <p className="mb-2 font-mono text-[11px] uppercase tracking-[0.18em] text-bone/50">Departments</p>
                    <DepartmentIndex onNavigate={close} />
                  </motion.div>
                </div>
              </motion.div>
            </Dialog.Content>
          </Dialog.Portal>
        )}
      </AnimatePresence>
    </Dialog.Root>
  );
}

// --- Header -------------------------------------------------------------------------

function AccountMenu() {
  const { user, isAdmin, logout } = useAuth();
  const navigate = useNavigate();
  if (!user) return null;
  const item = "flex cursor-pointer items-center gap-3 rounded-xl px-3 py-2.5 text-sm outline-none data-[highlighted]:bg-ink data-[highlighted]:text-bone [&_svg]:size-4";

  return (
    <Dropdown.Root modal={false}>
      <Dropdown.Trigger asChild>
        <button className="flex h-10 items-center gap-2 rounded-full pl-1 pr-3 text-sm transition-colors hover:bg-ink/5" aria-label="Account menu">
          <span className="flex size-8 items-center justify-center rounded-full bg-ink font-display text-xs font-bold text-acid">
            {user.first_name.slice(0, 1).toUpperCase()}
          </span>
          <span className="hidden max-w-28 truncate font-medium md:inline">{user.first_name}</span>
        </button>
      </Dropdown.Trigger>
      <Dropdown.Portal>
        <Dropdown.Content align="end" sideOffset={10} className="store animate-in z-50 w-60 rounded-2xl border border-ink/10 bg-bone p-1.5 shadow-2xl">
          <div className="px-3 py-2">
            <p className="truncate font-medium">{user.first_name} {user.last_name}</p>
            <p className="truncate font-mono text-[11px] text-smoke">{user.email}</p>
          </div>
          <Dropdown.Separator className="my-1 h-px bg-ink/10" />
          <Dropdown.Item asChild className={item}><Link to="/shop/account"><UserRound /> Account</Link></Dropdown.Item>
          <Dropdown.Item asChild className={item}><Link to="/shop/orders"><Package /> Orders</Link></Dropdown.Item>
          <Dropdown.Item asChild className={item}><Link to="/shop/wishlist"><Heart /> Wishlist</Link></Dropdown.Item>
          {isAdmin && (
            <Dropdown.Item asChild className={item}><Link to="/admin"><LayoutDashboard /> Admin console</Link></Dropdown.Item>
          )}
          <Dropdown.Separator className="my-1 h-px bg-ink/10" />
          <Dropdown.Item
            className={item}
            onSelect={() => {
              logout();
              navigate("/shop");
            }}
          >
            <LogOut /> Sign out
          </Dropdown.Item>
        </Dropdown.Content>
      </Dropdown.Portal>
    </Dropdown.Root>
  );
}

function BagButton() {
  const { count } = useCart();
  const bag = useBag();
  return (
    <button
      ref={bag.targetRef}
      onClick={() => bag.setOpen(true)}
      aria-label={`Bag, ${count} items`}
      className="relative flex h-10 items-center gap-2 rounded-full bg-ink pl-3 pr-4 text-sm font-medium text-bone transition-colors hover:bg-ink-3"
    >
      <ShoppingBag className="size-4" />
      <span key={bag.bumpKey} data-testid="cart-count" className={cn("tabular font-mono", bag.bumpKey > 0 && "bump inline-block")}>
        {count}
      </span>
    </button>
  );
}

function Header({ onSearch, onMenu }: { onSearch: () => void; onMenu: () => void }) {
  const { user } = useAuth();
  const wishlist = useWishlist();
  const location = useLocation();
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 12);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  const nav = [
    { label: "Index", to: "/shop/search" },
    { label: "Drops", to: "/shop/search?deals=true&sort=discount" },
    { label: "Most wanted", to: "/shop/search?sort=bestselling" },
  ];
  const current = location.pathname + location.search;

  return (
    <header
      className={cn(
        "sticky top-0 z-40 transition-[background-color,box-shadow,backdrop-filter] duration-300",
        scrolled ? "bg-bone/80 shadow-[0_1px_0_rgb(12_12_13/0.08)] backdrop-blur-xl" : "bg-transparent",
      )}
    >
      <div className="mx-auto flex h-[72px] max-w-[1600px] items-center gap-3 px-4 sm:px-8">
        <Link to="/shop" aria-label="meridian/ home" className="shrink-0">
          <Wordmark />
        </Link>

        <nav className="ml-6 hidden items-center gap-1 lg:flex" aria-label="Primary">
          {nav.map((item) => (
            <Link
              key={item.label}
              to={item.to}
              className={cn(
                "relative rounded-full px-3.5 py-2 text-sm transition-colors hover:bg-ink/5",
                current === item.to && "bg-ink text-bone hover:bg-ink",
              )}
            >
              {item.label}
            </Link>
          ))}
          <button onClick={onMenu} className="rounded-full px-3.5 py-2 text-sm transition-colors hover:bg-ink/5">
            Departments
          </button>
        </nav>

        <div className="ml-auto flex items-center gap-1.5 sm:gap-2">
          <button
            onClick={onSearch}
            aria-label="Open search"
            className="flex h-10 items-center gap-2 rounded-full border border-ink/15 px-3 text-sm text-smoke transition-colors hover:border-ink/40 sm:w-64 sm:px-4"
          >
            <Search className="size-4 text-ink" />
            <span className="hidden flex-1 text-left sm:inline">Search the index</span>
            <kbd className="hidden rounded-md border border-ink/15 px-1.5 font-mono text-[10px] sm:inline">⌘K</kbd>
          </button>
          <Link
            to="/shop/wishlist"
            aria-label={`Wishlist, ${wishlist.items.length} items`}
            className="relative hidden size-10 items-center justify-center rounded-full transition-colors hover:bg-ink/5 sm:flex"
          >
            <Heart className="size-[18px]" />
            {wishlist.items.length > 0 && (
              <span className="absolute right-1 top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-signal px-1 font-mono text-[9px] text-white">
                {wishlist.items.length}
              </span>
            )}
          </Link>
          {user ? (
            <AccountMenu />
          ) : (
            <Link to="/login" className="hidden h-10 items-center rounded-full px-3 text-sm transition-colors hover:bg-ink/5 sm:flex">
              Sign in
            </Link>
          )}
          <BagButton />
          <button onClick={onMenu} aria-label="Open menu" className="flex size-10 items-center justify-center rounded-full transition-colors hover:bg-ink/5 lg:hidden">
            <Menu className="size-5" />
          </button>
        </div>
      </div>
    </header>
  );
}

// --- Footer -------------------------------------------------------------------------

function Footer() {
  const health = useQuery({ queryKey: ["health"], queryFn: healthApi.health, refetchInterval: 60_000, retry: false });
  const departments = useDepartments();
  const up = health.data?.status === "UP" && health.data.database.status === "up";

  return (
    <footer className="store mt-24 overflow-hidden bg-ink text-bone">
      <Marquee duration={50} className="border-b border-bone/10 py-4">
        {(departments.data ?? []).map((department) => (
          <span key={department.id} className="flex items-center gap-6 px-6 font-display text-2xl font-bold tracking-tight">
            {department.name} <span className="text-acid">✳</span>
          </span>
        ))}
      </Marquee>
      <div className="mx-auto grid max-w-[1600px] gap-10 px-5 py-16 sm:px-8 md:grid-cols-4">
        <div className="md:col-span-2">
          <p className="max-w-md font-serif text-3xl italic leading-tight">
            An index of objects, run on a real order-processing engine.
          </p>
          <p className="mt-4 max-w-md text-sm text-bone/60">
            Stock is reserved the moment you order, payments are verified with Razorpay, and every order moves through a tracked lifecycle.
          </p>
          <p className="mt-6 inline-flex items-center gap-2 rounded-full border border-bone/15 px-3 py-1.5 font-mono text-[11px] uppercase tracking-wider text-bone/70" role="status">
            <span className={cn("size-2 rounded-full", health.isLoading ? "bg-bone/40" : up ? "bg-acid" : "bg-signal")} />
            {health.isLoading ? "Checking systems" : up ? "All systems operational" : "Systems degraded"}
          </p>
        </div>
        <div>
          <p className="mb-3 font-mono text-[11px] uppercase tracking-[0.18em] text-bone/50">Shop</p>
          <ul className="space-y-2 text-sm">
            <li><Link to="/shop/search" className="hover:text-acid">The index</Link></li>
            <li><Link to="/shop/search?deals=true&sort=discount" className="hover:text-acid">Drops</Link></li>
            <li><Link to="/shop/search?sort=bestselling" className="hover:text-acid">Most wanted</Link></li>
            <li><Link to="/shop/search?sort=newest" className="hover:text-acid">New in</Link></li>
          </ul>
        </div>
        <div>
          <p className="mb-3 font-mono text-[11px] uppercase tracking-[0.18em] text-bone/50">You</p>
          <ul className="space-y-2 text-sm">
            <li><Link to="/shop/orders" className="hover:text-acid">Orders & tracking</Link></li>
            <li><Link to="/shop/wishlist" className="hover:text-acid">Wishlist</Link></li>
            <li><Link to="/shop/cart" className="hover:text-acid">Bag</Link></li>
            <li><Link to="/shop/account" className="hover:text-acid">Account</Link></li>
          </ul>
        </div>
      </div>
      <div className="mx-auto max-w-[1600px] px-5 sm:px-8">
        <p className="select-none font-display text-[clamp(4rem,19vw,19rem)] font-extrabold leading-[0.8] tracking-[-0.06em] text-bone">
          meridian<span className="text-acid">/</span>
        </p>
      </div>
      <div className="mx-auto flex max-w-[1600px] flex-wrap justify-between gap-2 px-5 py-6 font-mono text-[11px] uppercase tracking-wider text-bone/40 sm:px-8">
        <span>Prices in INR, taxes included</span>
        <span>Demo store · product data from DummyJSON</span>
      </div>
    </footer>
  );
}

// --- Shell --------------------------------------------------------------------------

export function StoreShell() {
  const location = useLocation();
  const outlet = useOutlet();
  const [searchOpen, setSearchOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);

  // ⌘K / Ctrl+K and "/" open search, as in most tools.
  useEffect(() => {
    const onKey = (event: globalThis.KeyboardEvent) => {
      const typing = event.target instanceof HTMLElement && /input|textarea|select/i.test(event.target.tagName);
      if ((event.key === "k" && (event.metaKey || event.ctrlKey)) || (event.key === "/" && !typing)) {
        event.preventDefault();
        setSearchOpen(true);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  useEffect(() => {
    setMenuOpen(false);
    setSearchOpen(false);
  }, [location.pathname, location.search]);

  useEffect(() => {
    window.scrollTo(0, 0);
  }, [location.pathname]);

  const pageKey = useMemo(() => location.pathname, [location.pathname]);

  return (
    <BagProvider>
      <div className="store grain flex min-h-screen flex-col">
        <a href="#main" className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[60] focus:rounded-full focus:bg-ink focus:px-4 focus:py-2 focus:text-bone">
          Skip to content
        </a>
        <Header onSearch={() => setSearchOpen(true)} onMenu={() => setMenuOpen(true)} />
        <main id="main" className="flex-1">
          {/* Pages fade in; there is deliberately no exit animation, so the old
              page (and its route guards) unmount at once and never act on the
              new URL. */}
          <motion.div key={pageKey} initial={{ opacity: 0, y: 14 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.4, ease }}>
            {outlet}
          </motion.div>
        </main>
        <Footer />
      </div>
      <SearchPalette open={searchOpen} onOpenChange={setSearchOpen} />
      <MenuOverlay open={menuOpen} onOpenChange={setMenuOpen} />
    </BagProvider>
  );
}
