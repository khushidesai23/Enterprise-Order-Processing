import { forwardRef, type InputHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

export const StoreInput = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(({ className, ...props }, ref) => (
  <input
    ref={ref}
    className={cn(
      "h-12 w-full rounded-2xl border border-ink/15 bg-bone px-4 text-[15px] text-ink transition-colors placeholder:text-smoke/70 focus:border-ink focus:outline-none disabled:bg-bone-2 disabled:text-smoke aria-[invalid=true]:border-signal",
      className,
    )}
    {...props}
  />
));
StoreInput.displayName = "StoreInput";

export function StoreField({
  label,
  htmlFor,
  error,
  hint,
  children,
}: {
  label: string;
  htmlFor: string;
  error?: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label htmlFor={htmlFor} className="mb-1.5 block font-mono text-[11px] uppercase tracking-[0.14em] text-smoke">
        {label}
      </label>
      {children}
      {error ? <p className="mt-1.5 text-xs text-signal">{error}</p> : hint ? <p className="mt-1.5 text-xs text-smoke">{hint}</p> : null}
    </div>
  );
}

export function StoreButton({
  loading,
  children,
  className,
  disabled,
  type = "button",
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & { loading?: boolean; size?: string }) {
  const { size: _size, ...rest } = props as typeof props & { size?: string };
  void _size;
  return (
    <button
      type={type}
      disabled={disabled || loading}
      className={cn(
        "inline-flex h-14 items-center justify-center gap-2 rounded-full bg-ink px-7 text-base font-medium text-bone transition-[background-color,transform] hover:bg-ultra active:scale-[0.98] disabled:pointer-events-none disabled:opacity-40",
        className,
      )}
      {...rest}
    >
      {loading && <span className="size-4 animate-spin rounded-full border-2 border-bone/40 border-t-bone" />}
      {children}
    </button>
  );
}
