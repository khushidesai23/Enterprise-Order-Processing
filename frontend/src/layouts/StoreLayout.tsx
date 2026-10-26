import { ShoppingBag } from "lucide-react";
import { cn } from "@/lib/utils";
import { StoreShell } from "@/store/StoreShell";

// The storefront's chrome lives in src/store; this module keeps the
// routes' import and the logo the admin console uses.
export const StoreLayout = StoreShell;

export function Logo({ dark = false }: { dark?: boolean }) {
  return (
    <span className="flex items-center gap-2">
      <span className="flex size-8 items-center justify-center rounded-lg bg-primary text-white shadow-sm">
        <ShoppingBag className="size-[18px]" />
      </span>
      <span className={cn("text-[15px] font-semibold tracking-tight", dark ? "text-white" : "text-foreground")}>
        meridian<span className={dark ? "text-[#d7ff3a]" : "text-primary"}>/</span>
        <span className={cn("ml-1.5 text-xs font-medium", dark ? "text-slate-400" : "text-subtle")}>ops</span>
      </span>
    </span>
  );
}
