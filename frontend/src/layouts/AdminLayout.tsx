import { Suspense, useEffect, useState } from "react";
import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import {
  Boxes,
  CreditCard,
  ExternalLink,
  LayoutDashboard,
  Loader2,
  Menu,
  Package,
  ShoppingCart,
  Tags,
  Users,
  X,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { UserMenu } from "@/components/UserMenu";
import { Logo } from "./StoreLayout";

const NAV = [
  { to: "/admin", label: "Overview", icon: LayoutDashboard, end: true },
  { to: "/admin/orders", label: "Orders", icon: ShoppingCart },
  { to: "/admin/payments", label: "Payments", icon: CreditCard },
  { to: "/admin/products", label: "Products", icon: Package },
  { to: "/admin/categories", label: "Categories", icon: Tags },
  { to: "/admin/inventory", label: "Inventory", icon: Boxes },
  { to: "/admin/customers", label: "Customers", icon: Users },
];

function SidebarNav({ onNavigate }: { onNavigate?: () => void }) {
  return (
    <nav className="flex flex-col gap-0.5 px-3">
      <p className="px-3 pb-2 pt-4 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Operations</p>
      {NAV.map(({ to, label, icon: Icon, end }) => (
        <NavLink
          key={to}
          to={to}
          end={end}
          onClick={onNavigate}
          className={({ isActive }) =>
            cn(
              "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              isActive ? "bg-sidebar-active text-white" : "text-sidebar-foreground hover:bg-white/5 hover:text-white",
            )
          }
        >
          <Icon className="size-[18px]" />
          {label}
        </NavLink>
      ))}
    </nav>
  );
}

export function AdminLayout() {
  const [open, setOpen] = useState(false);
  const location = useLocation();

  useEffect(() => {
    setOpen(false);
  }, [location.pathname]);

  return (
    <div className="min-h-screen bg-background">
      {/* Desktop sidebar */}
      <aside className="fixed inset-y-0 left-0 z-30 hidden w-64 flex-col bg-sidebar lg:flex">
        <div className="flex h-16 items-center px-6">
          <Link to="/admin">
            <Logo dark />
          </Link>
        </div>
        <SidebarNav />
        <div className="mt-auto p-3">
          <Link
            to="/shop"
            className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-sidebar-foreground hover:bg-white/5 hover:text-white"
          >
            <ExternalLink className="size-[18px]" /> View storefront
          </Link>
        </div>
      </aside>

      {/* Mobile sidebar */}
      {open && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <div className="absolute inset-0 bg-slate-950/50" onClick={() => setOpen(false)} />
          <aside className="animate-in absolute inset-y-0 left-0 flex w-72 flex-col bg-sidebar">
            <div className="flex h-16 items-center justify-between px-5">
              <Logo dark />
              <button onClick={() => setOpen(false)} className="text-slate-300" aria-label="Close menu">
                <X className="size-5" />
              </button>
            </div>
            <SidebarNav onNavigate={() => setOpen(false)} />
            <div className="mt-auto p-3">
              <Link to="/shop" className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-sidebar-foreground">
                <ExternalLink className="size-[18px]" /> View storefront
              </Link>
            </div>
          </aside>
        </div>
      )}

      <div className="lg:pl-64">
        <header className="sticky top-0 z-20 flex h-16 items-center gap-4 border-b border-border bg-surface/90 px-4 backdrop-blur sm:px-6">
          <button className="text-subtle lg:hidden" onClick={() => setOpen(true)} aria-label="Open menu">
            <Menu className="size-5" />
          </button>
          <span className="rounded-md bg-primary-soft px-2 py-0.5 text-xs font-semibold text-primary">Admin console</span>
          <div className="ml-auto">
            <UserMenu />
          </div>
        </header>
        <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6">
          <Suspense fallback={<div className="flex min-h-[40vh] items-center justify-center text-subtle"><Loader2 className="size-6 animate-spin" /></div>}>
            <Outlet />
          </Suspense>
        </main>
      </div>
    </div>
  );
}
