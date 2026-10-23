import { Link, useNavigate } from "react-router-dom";
import * as Dropdown from "@radix-ui/react-dropdown-menu";
import { LayoutDashboard, LogOut, Package, Store, UserRound } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { fullName, initials } from "@/lib/format";
import { cn } from "@/lib/utils";

const itemClass =
  "flex cursor-pointer items-center gap-2.5 rounded-md px-2.5 py-2 text-sm text-foreground outline-none data-[highlighted]:bg-muted [&_svg]:size-4 [&_svg]:text-subtle";

export function UserMenu({ variant = "light" }: { variant?: "light" | "dark" }) {
  const { user, isAdmin, logout } = useAuth();
  const navigate = useNavigate();
  if (!user) return null;
  const name = fullName(user);

  return (
    <Dropdown.Root>
      <Dropdown.Trigger asChild>
        <button
          className={cn(
            "flex items-center gap-2 rounded-full py-1 pl-1 pr-2.5 text-sm font-medium transition-colors",
            variant === "light" ? "hover:bg-muted" : "text-white hover:bg-white/10",
          )}
          aria-label="Account menu"
        >
          <span className="flex size-8 items-center justify-center rounded-full bg-primary text-xs font-semibold text-white">
            {initials(name)}
          </span>
          <span className="hidden max-w-32 truncate sm:inline">{user.first_name}</span>
        </button>
      </Dropdown.Trigger>
      <Dropdown.Portal>
        <Dropdown.Content
          align="end"
          sideOffset={8}
          className="animate-in z-50 w-60 rounded-xl border border-border bg-surface p-1.5 shadow-pop"
        >
          <div className="px-2.5 py-2">
            <p className="truncate text-sm font-semibold">{name}</p>
            <p className="truncate text-xs text-subtle">{user.email}</p>
          </div>
          <Dropdown.Separator className="my-1 h-px bg-border" />
          <Dropdown.Item asChild className={itemClass}>
            <Link to="/shop/account">
              <UserRound /> My account
            </Link>
          </Dropdown.Item>
          <Dropdown.Item asChild className={itemClass}>
            <Link to="/shop/orders">
              <Package /> My orders
            </Link>
          </Dropdown.Item>
          <Dropdown.Item asChild className={itemClass}>
            <Link to="/shop">
              <Store /> Storefront
            </Link>
          </Dropdown.Item>
          {isAdmin && (
            <Dropdown.Item asChild className={itemClass}>
              <Link to="/admin">
                <LayoutDashboard /> Admin console
              </Link>
            </Dropdown.Item>
          )}
          <Dropdown.Separator className="my-1 h-px bg-border" />
          <Dropdown.Item
            className={itemClass}
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
