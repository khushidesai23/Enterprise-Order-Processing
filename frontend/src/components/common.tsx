import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { ChevronLeft, ChevronRight, Loader2, Minus, Plus } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";
import { Button } from "./ui/button";

// --- Product artwork --------------------------------------------------------
// The API stores no product images, so each product gets a stable,
// generated cover derived from its ID, so a product looks the same everywhere.

const PALETTES = [
  ["#eef2ff", "#c7d2fe", "#4f46e5"],
  ["#ecfdf3", "#a6f4c5", "#039855"],
  ["#fff6ed", "#fed7aa", "#c4320a"],
  ["#f0f9ff", "#b9e6fe", "#0086c9"],
  ["#fdf2fa", "#fcceee", "#c11574"],
  ["#fefbe8", "#fde68a", "#a15c07"],
  ["#f4f3ff", "#d9d6fe", "#6938ef"],
  ["#f0fdf9", "#99f6e0", "#0e9384"],
];

function hash(value: string): number {
  let result = 0;
  for (let i = 0; i < value.length; i++) result = (result * 31 + value.charCodeAt(i)) >>> 0;
  return result;
}

export function ProductArt({
  id,
  name,
  className,
  size = "md",
}: {
  id: string;
  name: string;
  className?: string;
  size?: "sm" | "md" | "lg";
}) {
  const seed = hash(id);
  const [bg, mid, ink] = PALETTES[seed % PALETTES.length]!;
  const variant = (seed >>> 4) % 3;
  const letters = name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0]!.toUpperCase())
    .join("");

  return (
    <div
      className={cn("relative flex items-center justify-center overflow-hidden", className)}
      style={{ background: bg }}
      aria-hidden
    >
      <svg className="absolute inset-0 h-full w-full" viewBox="0 0 200 200" preserveAspectRatio="xMidYMid slice">
        {variant === 0 && (
          <>
            <circle cx="160" cy="40" r="70" fill={mid} opacity="0.7" />
            <circle cx="30" cy="180" r="50" fill={mid} opacity="0.5" />
          </>
        )}
        {variant === 1 && (
          <>
            <rect x="-20" y="120" width="260" height="120" rx="40" fill={mid} opacity="0.6" transform="rotate(-12 100 160)" />
            <circle cx="40" cy="40" r="26" fill={mid} opacity="0.6" />
          </>
        )}
        {variant === 2 && (
          <>
            <path d="M0 140 C 60 100, 120 180, 200 120 L200 200 L0 200 Z" fill={mid} opacity="0.7" />
            <circle cx="150" cy="50" r="34" fill={mid} opacity="0.55" />
          </>
        )}
      </svg>
      <span
        className={cn(
          "relative font-semibold tracking-tight",
          size === "sm" && "text-sm",
          size === "md" && "text-3xl",
          size === "lg" && "text-6xl",
        )}
        style={{ color: ink }}
      >
        {letters}
      </span>
    </div>
  );
}

// --- Route guards -----------------------------------------------------------

function FullPageSpinner() {
  return (
    <div className="flex min-h-[50vh] items-center justify-center text-subtle">
      <Loader2 className="size-6 animate-spin" />
    </div>
  );
}

export function RequireAuth({ children }: { children: ReactNode }) {
  const { user, loading, signedOut } = useAuth();
  const location = useLocation();
  if (loading) return <FullPageSpinner />;
  if (!user && signedOut) return <Navigate to="/shop" replace />;
  if (!user) return <Navigate to="/login" replace state={{ from: location.pathname + location.search }} />;
  return <>{children}</>;
}

// Hiding admin pages is a convenience only: every admin route is enforced
// by RequireRole(admin) on the API.
export function RequireAdmin({ children }: { children: ReactNode }) {
  const { user, loading, isAdmin, signedOut } = useAuth();
  const location = useLocation();
  if (loading) return <FullPageSpinner />;
  if (!user && signedOut) return <Navigate to="/shop" replace />;
  if (!user) return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  if (!isAdmin) return <Navigate to="/shop" replace />;
  return <>{children}</>;
}

// --- Layout bits ------------------------------------------------------------

export function PageHeader({
  title,
  description,
  actions,
  eyebrow,
}: {
  title: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  eyebrow?: ReactNode;
}) {
  return (
    <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        {eyebrow && <div className="mb-1 text-[13px] text-subtle">{eyebrow}</div>}
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        {description && <p className="mt-1 text-sm text-subtle">{description}</p>}
      </div>
      {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
    </div>
  );
}

export function Pagination({
  page,
  hasNext,
  onChange,
  loading,
}: {
  page: number;
  hasNext: boolean;
  onChange: (page: number) => void;
  loading?: boolean;
}) {
  if (page === 1 && !hasNext) return null;
  return (
    <div className="flex items-center justify-between px-4 py-3 text-sm text-subtle">
      <span>Page {page}</span>
      <div className="flex gap-2">
        <Button variant="secondary" size="sm" disabled={page <= 1 || loading} onClick={() => onChange(page - 1)}>
          <ChevronLeft /> Previous
        </Button>
        <Button variant="secondary" size="sm" disabled={!hasNext || loading} onClick={() => onChange(page + 1)}>
          Next <ChevronRight />
        </Button>
      </div>
    </div>
  );
}

export function StatCard({
  label,
  value,
  hint,
  icon,
  tone = "primary",
}: {
  label: string;
  value: ReactNode;
  hint?: ReactNode;
  icon: ReactNode;
  tone?: "primary" | "success" | "warning" | "danger" | "info";
}) {
  const tones = {
    primary: "bg-primary-soft text-primary",
    success: "bg-success-soft text-success",
    warning: "bg-warning-soft text-warning",
    danger: "bg-danger-soft text-danger",
    info: "bg-info-soft text-info",
  };
  return (
    <div className="rounded-xl border border-border bg-surface p-5 shadow-card">
      <div className="flex items-start justify-between">
        <p className="text-[13px] font-medium text-subtle">{label}</p>
        <div className={cn("flex size-9 items-center justify-center rounded-lg [&_svg]:size-[18px]", tones[tone])}>
          {icon}
        </div>
      </div>
      <div className="tabular mt-2 text-2xl font-semibold tracking-tight">{value}</div>
      {hint && <p className="mt-1 text-xs text-subtle">{hint}</p>}
    </div>
  );
}

export function QuantityStepper({
  value,
  max,
  onChange,
  label = "Quantity",
}: {
  value: number;
  max: number;
  onChange: (value: number) => void;
  label?: string;
}) {
  return (
    <div className="inline-flex h-10 items-center rounded-lg border border-border bg-surface shadow-sm" role="group" aria-label={label}>
      <button
        type="button"
        className="flex h-full w-9 items-center justify-center text-subtle hover:text-foreground disabled:opacity-40"
        onClick={() => onChange(value - 1)}
        disabled={value <= 1}
        aria-label="Decrease quantity"
      >
        <Minus className="size-4" />
      </button>
      <span className="tabular w-8 text-center text-sm font-medium" aria-live="polite">
        {value}
      </span>
      <button
        type="button"
        className="flex h-full w-9 items-center justify-center text-subtle hover:text-foreground disabled:opacity-40"
        onClick={() => onChange(value + 1)}
        disabled={value >= max}
        aria-label="Increase quantity"
      >
        <Plus className="size-4" />
      </button>
    </div>
  );
}
