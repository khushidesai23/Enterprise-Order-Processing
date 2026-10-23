import { useEffect, useState, type FormEvent } from "react";
import { Link, NavLink, Outlet, useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { LayoutDashboard, Search, ShoppingBag } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { useCart } from "@/lib/cart";
import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";
import { UserMenu } from "@/components/UserMenu";

export function Logo({ dark = false }: { dark?: boolean }) {
  return (
    <span className="flex items-center gap-2">
      <span className="flex size-8 items-center justify-center rounded-lg bg-primary text-white shadow-sm">
        <ShoppingBag className="size-[18px]" />
      </span>
      <span className={cn("text-[15px] font-semibold tracking-tight", dark ? "text-white" : "text-foreground")}>
        Meridian<span className={dark ? "text-indigo-300" : "text-primary"}> Market</span>
      </span>
    </span>
  );
}

function HeaderSearch() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();
  const [value, setValue] = useState(params.get("q") ?? "");

  useEffect(() => {
    if (location.pathname === "/shop") setValue(params.get("q") ?? "");
  }, [params, location.pathname]);

  function submit(event: FormEvent) {
    event.preventDefault();
    const next = new URLSearchParams(location.pathname === "/shop" ? params : undefined);
    if (value.trim()) next.set("q", value.trim());
    else next.delete("q");
    navigate(`/shop?${next.toString()}`);
  }

  return (
    <form onSubmit={submit} className="relative hidden w-full max-w-md md:block" role="search">
      <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-faint" />
      <input
        value={value}
        onChange={(event) => setValue(event.target.value)}
        placeholder="Search products or SKUs"
        aria-label="Search products"
        className="h-10 w-full rounded-full border border-border bg-muted/60 pl-9 pr-4 text-sm placeholder:text-faint focus:border-primary focus:bg-surface focus:outline-none focus:ring-3 focus:ring-primary/15"
      />
    </form>
  );
}

export function StoreLayout() {
  const { user, isAdmin } = useAuth();
  const { count } = useCart();
  const location = useLocation();

  // Braces matter: newer browsers return a Promise from scrollTo, which
  // React would otherwise call as the effect cleanup.
  useEffect(() => {
    window.scrollTo(0, 0);
  }, [location.pathname]);

  const navClass = ({ isActive }: { isActive: boolean }) =>
    cn(
      "rounded-md px-3 py-2 text-sm font-medium transition-colors",
      isActive ? "text-foreground" : "text-subtle hover:text-foreground",
    );

  return (
    <div className="flex min-h-screen flex-col">
      <header className="sticky top-0 z-40 border-b border-border bg-surface/90 backdrop-blur">
        <div className="mx-auto flex h-16 max-w-7xl items-center gap-6 px-4 sm:px-6">
          <Link to="/shop" aria-label="Meridian Market home">
            <Logo />
          </Link>
          <nav className="hidden items-center gap-1 sm:flex">
            <NavLink to="/shop" end className={navClass}>
              Shop
            </NavLink>
            {user && (
              <NavLink to="/shop/orders" className={navClass}>
                Orders
              </NavLink>
            )}
          </nav>
          <div className="flex flex-1 justify-center">
            <HeaderSearch />
          </div>
          <div className="flex items-center gap-2">
            {isAdmin && (
              <Link to="/admin" className={cn(buttonVariants({ variant: "secondary", size: "sm" }), "hidden lg:inline-flex")}>
                <LayoutDashboard /> Admin
              </Link>
            )}
            <Link
              to="/shop/cart"
              className="relative flex size-10 items-center justify-center rounded-full text-subtle transition-colors hover:bg-muted hover:text-foreground"
              aria-label={`Cart, ${count} items`}
            >
              <ShoppingBag className="size-5" />
              {count > 0 && (
                <span
                  data-testid="cart-count"
                  className="tabular absolute -right-0.5 -top-0.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1 text-[11px] font-semibold text-white"
                >
                  {count}
                </span>
              )}
            </Link>
            {user ? (
              <UserMenu />
            ) : (
              <>
                <Link to="/login" className={buttonVariants({ variant: "ghost", size: "sm" })}>
                  Sign in
                </Link>
                <Link to="/register" className={cn(buttonVariants({ size: "sm" }), "hidden sm:inline-flex")}>
                  Create account
                </Link>
              </>
            )}
          </div>
        </div>
      </header>

      <main className="flex-1">
        <Outlet />
      </main>

      <footer className="border-t border-border bg-surface">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 px-4 py-8 text-sm text-subtle sm:flex-row sm:items-center sm:justify-between sm:px-6">
          <Logo />
          <p>Secure payments by Razorpay · Prices in INR, inclusive of taxes</p>
        </div>
      </footer>
    </div>
  );
}
