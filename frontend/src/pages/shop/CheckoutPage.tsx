import { useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { AnimatePresence, motion } from "motion/react";
import { toast } from "sonner";
import { Check, Loader2, Lock, ShieldCheck } from "lucide-react";
import { ordersApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useCart } from "@/lib/cart";
import { fullName, money, number } from "@/lib/format";
import { usePayOrder } from "@/lib/usePayOrder";
import type { Order } from "@/lib/types";
import { cn } from "@/lib/utils";
import { useBagAvailability } from "@/store/bag";
import { Eyebrow, MaskText, ProductShot, ease, pill } from "@/store/ui";

const STEPS = ["Bag", "Review", "Pay", "Done"];

function Steps({ current }: { current: number }) {
  return (
    <ol className="flex items-center gap-2 font-mono text-[11px] uppercase tracking-[0.16em]">
      {STEPS.map((step, index) => (
        <li key={step} className="flex items-center gap-2">
          <span
            className={cn(
              "flex size-7 items-center justify-center rounded-full border text-[10px]",
              index < current ? "border-ink bg-ink text-acid" : index === current ? "border-ink text-ink" : "border-ink/20 text-smoke",
            )}
          >
            {index < current ? <Check className="size-3.5" /> : index + 1}
          </span>
          <span className={index <= current ? "text-ink" : "text-smoke"}>{step}</span>
          {index < STEPS.length - 1 && <span className="mx-1 h-px w-6 bg-ink/20 sm:w-10" />}
        </li>
      ))}
    </ol>
  );
}

export function CheckoutPage() {
  const { user } = useAuth();
  const { lines, count, clear } = useCart();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { pay, paying } = usePayOrder();
  const { byId, issues, subtotal, loading } = useBagAvailability();
  const [placed, setPlaced] = useState<Order | null>(null);

  // POST /orders reserves the stock; usePayOrder then runs
  // POST /payments -> Razorpay Checkout -> POST /payments/verify.
  const placeOrder = useMutation({
    mutationFn: () => ordersApi.create(lines.map((line) => ({ product_id: line.productId, quantity: line.quantity }))),
    onSuccess: async (order) => {
      setPlaced(order);
      clear();
      await queryClient.invalidateQueries();
      const paid = await pay(order.id);
      navigate(`/shop/orders/${order.id}`, { replace: true, state: { justPlaced: true, paid } });
    },
    onError: (error) => toast.error("Could not place your order", { description: errorMessage(error) }),
  });

  if (lines.length === 0 && !placed) return <Navigate to="/shop/cart" replace />;

  const busy = placeOrder.isPending || paying || !!placed;
  const status = placeOrder.isPending ? "Reserving your stock…" : paying || placed ? "Waiting for secure payment…" : null;

  return (
    <div className="mx-auto max-w-[1400px] px-5 pt-6 sm:px-8">
      <Steps current={placed ? 2 : 1} />
      <h1 className="mt-8 font-display text-[clamp(3rem,7vw,6rem)] font-extrabold leading-[0.88] tracking-[-0.055em]">
        <MaskText text="Almost" /> <MaskText text="yours." delay={0.1} className="font-serif font-normal italic tracking-[-0.02em] text-ultra" />
      </h1>

      <div className="mt-10 grid items-start gap-10 lg:grid-cols-[1fr_420px]">
        <div className="space-y-6">
          <section className="rounded-[2rem] bg-bone-2/70 p-6 sm:p-8">
            <Eyebrow>01 — Contact</Eyebrow>
            <p className="mt-3 text-lg font-medium">{user ? fullName(user) : ""}</p>
            <p className="font-mono text-sm text-smoke">{user?.email}</p>
            <p className="mt-3 text-sm text-smoke">Order and delivery updates go to this account.</p>
          </section>

          <section className="rounded-[2rem] bg-bone-2/70 p-6 sm:p-8">
            <div className="flex items-baseline justify-between">
              <Eyebrow>02 — Items · {number(count)}</Eyebrow>
              <Link to="/shop/cart" className="font-mono text-[11px] uppercase tracking-wider underline underline-offset-4">
                Edit bag
              </Link>
            </div>
            <ul className="mt-4 divide-y divide-ink/10">
              {(placed ? [] : lines).map((line) => {
                const live = byId.get(line.productId);
                return (
                  <li key={line.productId} className="flex items-center gap-4 py-3">
                    <ProductShot id={line.productId} name={line.name} src={live?.image_url ?? line.imageUrl} size="sm" className="size-16 shrink-0 rounded-2xl bg-bone p-1.5" />
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-medium">{line.name}</p>
                      <p className="font-mono text-xs text-smoke">Qty {line.quantity}</p>
                      {issues.get(line.productId) && <p className="font-mono text-xs text-signal">{issues.get(line.productId)}</p>}
                    </div>
                    <p className="tabular font-mono text-sm">{money((live?.price ?? line.price) * line.quantity)}</p>
                  </li>
                );
              })}
              {placed && <li className="py-3 text-sm text-smoke">Order placed. Your items are reserved.</li>}
            </ul>
          </section>

          <section className="rounded-[2rem] bg-bone-2/70 p-6 sm:p-8">
            <Eyebrow>03 — Payment</Eyebrow>
            <p className="mt-3 max-w-xl text-[15px] leading-relaxed">
              Placing the order reserves your items straight away. You then pay in Razorpay's secure window with UPI, cards, netbanking or wallets.
              The order is only confirmed once the payment is verified on our side. Close the window and you can pay later from your orders.
            </p>
          </section>
        </div>

        <aside className="rounded-[2rem] bg-ink p-7 text-bone lg:sticky lg:top-28" aria-label="Order summary">
          <p className="font-mono text-[11px] uppercase tracking-[0.18em] text-bone/50">Total due</p>
          <p className="tabular mt-2 break-words font-display text-[clamp(1.8rem,3vw,2.8rem)] font-extrabold leading-tight tracking-tight">
            {money(placed?.total_amount ?? subtotal)}
          </p>
          <p className="mt-1 text-sm text-bone/60">Free delivery · taxes included · confirmed with current prices</p>
          <button
            onClick={() => placeOrder.mutate()}
            disabled={busy || issues.size > 0 || loading || lines.length === 0}
            className={cn(pill("acid", "lg"), "mt-7 w-full")}
          >
            {busy ? <Loader2 className="size-5 animate-spin" /> : <Lock className="size-5" />}
            {placeOrder.isPending ? "Placing order…" : busy ? "Waiting for payment…" : "Place order & pay"}
          </button>
          <AnimatePresence>
            {status && (
              <motion.p
                key={status}
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.3, ease }}
                className="mt-3 text-center font-mono text-xs text-acid"
                role="status"
              >
                {status}
              </motion.p>
            )}
          </AnimatePresence>
          <p className="mt-5 flex items-center justify-center gap-1.5 font-mono text-[10px] uppercase tracking-wider text-bone/50">
            <ShieldCheck className="size-3.5" /> Verified server-side before confirmation
          </p>
        </aside>
      </div>
    </div>
  );
}
