import { createContext, useCallback, useContext, useMemo, useRef, useState, type MutableRefObject, type ReactNode } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useQueries } from "@tanstack/react-query";
import * as Dialog from "@radix-ui/react-dialog";
import { AnimatePresence, animate, motion, useReducedMotion } from "motion/react";
import { ArrowRight, Minus, Plus, X } from "lucide-react";
import { productsApi } from "@/lib/endpoints";
import { useAuth } from "@/lib/auth";
import { MAX_LINE_QUANTITY, useCart, type CartLine } from "@/lib/cart";
import { money } from "@/lib/format";
import type { Product } from "@/lib/types";
import { cn } from "@/lib/utils";
import { ProductShot, ease, pill } from "./ui";

// --- Live stock for bag lines ----------------------------------------------------

// Each line is checked against the live product so price and stock
// problems show before checkout instead of as a failed order.
export function useBagAvailability() {
  const { lines } = useCart();
  const products = useQueries({
    queries: lines.map((line) => ({
      queryKey: ["product", line.productId],
      queryFn: () => productsApi.get(line.productId),
      retry: false,
      staleTime: 15_000,
    })),
  });

  const byId = new Map<string, Product>();
  const issues = new Map<string, string>();
  lines.forEach((line, index) => {
    const query = products[index];
    if (!query) return;
    if (query.data) {
      byId.set(line.productId, query.data);
      const available = query.data.inventory?.available_quantity ?? 0;
      if (available <= 0) issues.set(line.productId, "Sold out");
      else if (line.quantity > available) issues.set(line.productId, `Only ${available} available`);
    } else if (query.isError) {
      issues.set(line.productId, "No longer available");
    }
  });

  const subtotal = lines.reduce((sum, line) => sum + (byId.get(line.productId)?.price ?? line.price) * line.quantity, 0);
  return { byId, issues, subtotal, loading: products.some((query) => query.isLoading) };
}

// --- Bag UI: drawer and fly-to-bag -------------------------------------------------

interface BagUi {
  open: boolean;
  setOpen: (open: boolean) => void;
  targetRef: MutableRefObject<HTMLButtonElement | null>;
  bumpKey: number;
  add: (line: Omit<CartLine, "quantity">, quantity: number, source?: HTMLElement | null) => void;
}

const BagContext = createContext<BagUi | null>(null);

export function BagProvider({ children }: { children: ReactNode }) {
  const cart = useCart();
  const reduce = useReducedMotion();
  const [open, setOpen] = useState(false);
  const [bumpKey, setBumpKey] = useState(0);
  const targetRef = useRef<HTMLButtonElement | null>(null);

  // A copy of the product image flies from where it was added into the
  // bag button, which then bumps.
  const fly = useCallback(
    (source: HTMLElement | null | undefined, imageUrl: string | undefined) => {
      const target = targetRef.current;
      const image = source?.querySelector("img") ?? null;
      if (reduce || !target || !image || !imageUrl) {
        setBumpKey((key) => key + 1);
        return;
      }
      const from = image.getBoundingClientRect();
      const to = target.getBoundingClientRect();
      const ghost = document.createElement("img");
      ghost.src = imageUrl;
      ghost.alt = "";
      Object.assign(ghost.style, {
        position: "fixed",
        left: `${from.left}px`,
        top: `${from.top}px`,
        width: `${from.width}px`,
        height: `${from.height}px`,
        objectFit: "contain",
        zIndex: "80",
        pointerEvents: "none",
      });
      document.body.appendChild(ghost);
      const dx = to.left + to.width / 2 - (from.left + from.width / 2);
      const dy = to.top + to.height / 2 - (from.top + from.height / 2);
      animate(
        ghost,
        { x: [0, dx * 0.35, dx], y: [0, dy - 120, dy], scale: [1, 0.7, 0.08], opacity: [1, 1, 0.2] },
        { duration: 0.75, ease },
      ).then(() => {
        ghost.remove();
        setBumpKey((key) => key + 1);
      });
    },
    [reduce],
  );

  const add = useCallback(
    (line: Omit<CartLine, "quantity">, quantity: number, source?: HTMLElement | null) => {
      cart.add(line, quantity);
      fly(source, line.imageUrl);
    },
    [cart, fly],
  );

  const value = useMemo(() => ({ open, setOpen, targetRef, bumpKey, add }), [open, bumpKey, add]);
  return (
    <BagContext.Provider value={value}>
      {children}
      <BagDrawer open={open} onOpenChange={setOpen} />
    </BagContext.Provider>
  );
}

export function useBag(): BagUi {
  const context = useContext(BagContext);
  if (!context) throw new Error("useBag must be used inside BagProvider");
  return context;
}

// --- Quantity stepper --------------------------------------------------------------

export function Stepper({
  value,
  max,
  onChange,
  label,
  size = "md",
}: {
  value: number;
  max: number;
  onChange: (value: number) => void;
  label: string;
  size?: "sm" | "md";
}) {
  const box = size === "sm" ? "h-8" : "h-11";
  return (
    <div className={cn("inline-flex items-center rounded-full border border-ink/20 bg-bone", box)} role="group" aria-label={label}>
      <button
        type="button"
        onClick={() => onChange(value - 1)}
        disabled={value <= 1}
        aria-label={`Decrease ${label.toLowerCase()}`}
        className={cn("flex h-full items-center justify-center rounded-l-full text-ink transition-colors hover:bg-ink/5 disabled:opacity-30", size === "sm" ? "w-8" : "w-11")}
      >
        <Minus className="size-3.5" />
      </button>
      <span className="tabular w-7 text-center font-mono text-sm" aria-live="polite">
        {value}
      </span>
      <button
        type="button"
        onClick={() => onChange(value + 1)}
        disabled={value >= max}
        aria-label={`Increase ${label.toLowerCase()}`}
        className={cn("flex h-full items-center justify-center rounded-r-full text-ink transition-colors hover:bg-ink/5 disabled:opacity-30", size === "sm" ? "w-8" : "w-11")}
      >
        <Plus className="size-3.5" />
      </button>
    </div>
  );
}

// --- Drawer -------------------------------------------------------------------------

function BagDrawer({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const { lines, count, setQuantity, remove } = useCart();
  const { user } = useAuth();
  const navigate = useNavigate();
  const { byId, issues, subtotal } = useBagAvailability();

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open && (
          <Dialog.Portal forceMount>
            <Dialog.Overlay asChild forceMount>
              <motion.div
                className="fixed inset-0 z-50 bg-ink/40 backdrop-blur-[2px]"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
              />
            </Dialog.Overlay>
            <Dialog.Content asChild forceMount>
              <motion.aside
                className="store fixed inset-y-2 right-2 z-50 flex w-[min(440px,calc(100vw-1rem))] flex-col overflow-hidden rounded-3xl bg-bone shadow-2xl focus:outline-none"
                initial={{ x: "105%" }}
                animate={{ x: 0 }}
                exit={{ x: "105%" }}
                transition={{ duration: 0.5, ease }}
              >
                <div className="flex items-center justify-between border-b border-ink/10 px-6 py-5">
                  <Dialog.Title className="font-display text-2xl font-bold tracking-tight">
                    Bag <span className="font-mono text-sm font-normal text-smoke">({count})</span>
                  </Dialog.Title>
                  <Dialog.Description className="sr-only">Items in your bag</Dialog.Description>
                  <Dialog.Close className="flex size-10 items-center justify-center rounded-full hover:bg-ink/5" aria-label="Close bag">
                    <X className="size-5" />
                  </Dialog.Close>
                </div>

                {lines.length === 0 ? (
                  <div className="flex flex-1 flex-col items-start justify-center gap-4 px-6">
                    <p className="font-display text-4xl font-bold leading-none tracking-tight">
                      Nothing here <span className="font-serif font-normal italic">yet.</span>
                    </p>
                    <p className="text-smoke">Objects you add appear here, held at live prices.</p>
                    <Link to="/shop/search" onClick={() => onOpenChange(false)} className={pill("ink")}>
                      Browse the index <ArrowRight className="size-4" />
                    </Link>
                  </div>
                ) : (
                  <>
                    <ul className="flex-1 divide-y divide-ink/10 overflow-y-auto px-6">
                      <AnimatePresence initial={false}>
                        {lines.map((line) => {
                          const live = byId.get(line.productId);
                          const price = live?.price ?? line.price;
                          const available = live?.inventory?.available_quantity;
                          const issue = issues.get(line.productId);
                          return (
                            <motion.li
                              key={line.productId}
                              layout
                              initial={{ opacity: 0, x: 24 }}
                              animate={{ opacity: 1, x: 0 }}
                              exit={{ opacity: 0, x: 40, height: 0 }}
                              transition={{ duration: 0.35, ease }}
                              className="flex gap-4 py-4"
                            >
                              <ProductShot id={line.productId} name={line.name} src={live?.image_url ?? line.imageUrl} size="sm" className="size-20 shrink-0 rounded-2xl bg-bone-2 p-2" />
                              <div className="min-w-0 flex-1">
                                <Link
                                  to={`/shop/products/${line.productId}`}
                                  onClick={() => onOpenChange(false)}
                                  className="line-clamp-2 text-sm font-medium hover:underline"
                                >
                                  {line.name}
                                </Link>
                                <p className="tabular mt-0.5 font-mono text-xs text-smoke">{money(price)}</p>
                                {issue && <p className="mt-1 font-mono text-xs text-signal">{issue}</p>}
                                <div className="mt-2 flex items-center justify-between">
                                  <Stepper
                                    size="sm"
                                    value={line.quantity}
                                    max={Math.max(1, Math.min(available ?? MAX_LINE_QUANTITY, MAX_LINE_QUANTITY))}
                                    onChange={(value) => setQuantity(line.productId, value)}
                                    label={`Quantity of ${line.name}`}
                                  />
                                  <button onClick={() => remove(line.productId)} className="font-mono text-xs text-smoke underline-offset-4 hover:text-ink hover:underline">
                                    Remove
                                  </button>
                                </div>
                              </div>
                            </motion.li>
                          );
                        })}
                      </AnimatePresence>
                    </ul>
                    <div className="border-t border-ink/10 bg-bone-2/60 px-6 py-5">
                      <div className="flex items-baseline justify-between">
                        <span className="font-mono text-xs uppercase tracking-widest text-smoke">Subtotal</span>
                        <span className="tabular font-display text-xl font-bold">{money(subtotal)}</span>
                      </div>
                      <p className="mt-1 text-xs text-smoke">Free delivery over ₹499. Final total confirmed at checkout.</p>
                      <div className="mt-4 grid grid-cols-2 gap-2">
                        <Link to="/shop/cart" onClick={() => onOpenChange(false)} className={pill("line")}>
                          View bag
                        </Link>
                        <button
                          className={pill("acid")}
                          disabled={issues.size > 0}
                          onClick={() => {
                            onOpenChange(false);
                            navigate(user ? "/shop/checkout" : "/login", { state: { from: "/shop/checkout" } });
                          }}
                        >
                          Checkout <ArrowRight className="size-4" />
                        </button>
                      </div>
                    </div>
                  </>
                )}
              </motion.aside>
            </Dialog.Content>
          </Dialog.Portal>
        )}
      </AnimatePresence>
    </Dialog.Root>
  );
}
