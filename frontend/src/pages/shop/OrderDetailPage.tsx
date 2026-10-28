import { useState } from "react";
import { Link, useLocation, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import * as Dialog from "@radix-ui/react-dialog";
import { motion, useReducedMotion } from "motion/react";
import { toast } from "sonner";
import { ArrowLeft, CreditCard } from "lucide-react";
import { ordersApi, paymentsApi } from "@/lib/endpoints";
import { ApiError, errorMessage } from "@/lib/api";
import { dateTime, money, shortId } from "@/lib/format";
import { PAYMENT_STATUS_LABEL, isAwaitingPayment } from "@/lib/orders";
import { usePayOrder } from "@/lib/usePayOrder";
import { cn } from "@/lib/utils";
import { OrderTrack, StatusChip } from "@/store/orders";
import { Eyebrow, ProductShot, Skel, ease, pill } from "@/store/ui";

// A short burst of acid dots behind the confirmation headline.
function Burst() {
  const reduce = useReducedMotion();
  if (reduce) return null;
  return (
    <div aria-hidden className="pointer-events-none absolute left-1/2 top-1/2">
      {Array.from({ length: 18 }).map((_, i) => {
        const angle = (i / 18) * Math.PI * 2;
        const distance = 160 + (i % 3) * 70;
        return (
          <motion.span
            key={i}
            className={cn("absolute size-3 rounded-full", i % 3 === 0 ? "bg-ultra" : i % 3 === 1 ? "bg-acid" : "bg-ink")}
            initial={{ x: 0, y: 0, scale: 0, opacity: 1 }}
            animate={{ x: Math.cos(angle) * distance, y: Math.sin(angle) * distance, scale: [0, 1.4, 0.6], opacity: [1, 1, 0] }}
            transition={{ duration: 1.4, ease, delay: 0.2 + (i % 4) * 0.04 }}
          />
        );
      })}
    </div>
  );
}

export function OrderDetailPage() {
  const { id = "" } = useParams();
  const location = useLocation();
  const queryClient = useQueryClient();
  const justPlaced = (location.state as { justPlaced?: boolean } | null)?.justPlaced;
  const [confirmCancel, setConfirmCancel] = useState(false);
  const { pay, paying } = usePayOrder();

  const order = useQuery({ queryKey: ["order", id], queryFn: () => ordersApi.get(id) });
  // 404 simply means no payment has been started for this order yet.
  const payment = useQuery({
    queryKey: ["payment", "order", id],
    queryFn: async () => {
      try {
        return await paymentsApi.byOrder(id);
      } catch (error) {
        if (error instanceof ApiError && error.status === 404) return null;
        throw error;
      }
    },
    enabled: order.isSuccess,
  });

  const cancel = useMutation({
    mutationFn: () => ordersApi.cancel(id),
    onSuccess: async () => {
      toast("Order cancelled", { description: "Reserved items were returned to stock." });
      setConfirmCancel(false);
      await queryClient.invalidateQueries();
    },
    onError: (error) => toast.error("Could not cancel the order", { description: errorMessage(error) }),
  });

  if (order.isLoading) {
    return (
      <div className="mx-auto max-w-[1200px] space-y-6 px-5 py-10 sm:px-8">
        <Skel className="h-24 w-2/3" />
        <Skel className="h-36" />
        <Skel className="h-72" />
      </div>
    );
  }

  if (order.isError || !order.data) {
    const missing = order.error instanceof ApiError && (order.error.status === 404 || order.error.status === 400);
    return (
      <div className="mx-auto flex max-w-3xl flex-col items-start gap-4 px-5 py-24">
        <p className="font-display text-6xl font-extrabold tracking-tight">{missing ? "Order not found." : "Something broke."}</p>
        {!missing && <p className="text-smoke">{errorMessage(order.error)}</p>}
        <Link to="/shop/orders" className={pill("ink")}>
          Back to orders
        </Link>
      </div>
    );
  }

  const data = order.data;
  const awaitingPayment = isAwaitingPayment(data.status);
  const paymentFailed = payment.data?.status === "FAILED";
  const confirmed = justPlaced && data.status === "PAID";

  return (
    <div className="mx-auto max-w-[1200px] px-5 pt-6 sm:px-8">
      <Link to="/shop/orders" className="inline-flex items-center gap-1.5 font-mono text-[11px] uppercase tracking-[0.16em] text-smoke hover:text-ink">
        <ArrowLeft className="size-3.5" /> All orders
      </Link>

      {confirmed ? (
        <section className="relative mt-6 overflow-hidden rounded-[2.5rem] bg-acid px-8 py-16 text-center sm:py-20">
          <Burst />
          <motion.p initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="relative font-mono text-[11px] uppercase tracking-[0.2em]">
            Payment verified · {shortId(data.id)}
          </motion.p>
          <motion.h1
            initial={{ opacity: 0, y: 30, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            transition={{ duration: 0.8, ease }}
            className="relative mt-4 font-display text-[clamp(3rem,9vw,7.5rem)] font-extrabold leading-[0.85] tracking-[-0.06em]"
          >
            It's <span className="font-serif font-normal italic">yours.</span>
          </motion.h1>
          <p className="relative mx-auto mt-5 max-w-md">Thank you! Your order is confirmed. We'll pack it next; follow every step right here.</p>
        </section>
      ) : (
        <header className="mt-6 flex flex-wrap items-end justify-between gap-4">
          <div>
            <Eyebrow>Placed {dateTime(data.created_at)}</Eyebrow>
            <h1 className="mt-2 font-display text-[clamp(2.6rem,6vw,5rem)] font-extrabold leading-[0.9] tracking-[-0.05em]">
              Order <span className="font-mono text-[0.6em] font-medium tracking-tight">{shortId(data.id)}</span>
            </h1>
          </div>
          <StatusChip status={data.status} />
        </header>
      )}

      {awaitingPayment && (
        <motion.section
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease }}
          className="mt-8 flex flex-col gap-5 rounded-[2rem] bg-ink p-7 text-bone sm:flex-row sm:items-center sm:justify-between"
        >
          <div>
            <p className="font-display text-2xl font-bold tracking-tight">{paymentFailed ? "Your last payment attempt failed" : "Payment pending"}</p>
            <p className="mt-1 text-sm text-bone/70">Your items are reserved. Complete payment to confirm the order.</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <button onClick={() => setConfirmCancel(true)} disabled={paying} className={pill("line", "md") + " border-bone/40 text-bone hover:bg-bone hover:text-ink"}>
              Cancel order
            </button>
            <button onClick={() => pay(data.id)} disabled={paying} className={pill("acid", "md")}>
              <CreditCard className="size-4" /> {paying ? "Opening…" : paymentFailed ? "Retry payment" : "Pay now"}
            </button>
          </div>
        </motion.section>
      )}

      <section className="mt-8 rounded-[2rem] bg-bone-2/70 px-4 py-8 sm:px-8">
        <OrderTrack status={data.status} />
      </section>

      <div className="mt-8 grid items-start gap-8 lg:grid-cols-[1fr_340px]">
        <section>
          <Eyebrow>Items · {data.items.length}</Eyebrow>
          <ul className="mt-4 divide-y divide-ink/10 border-y border-ink/10">
            {data.items.map((item, index) => (
              <motion.li
                key={item.product_id}
                initial={{ opacity: 0, x: -12 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.4, ease, delay: index * 0.05 }}
                className="flex items-center gap-4 py-4"
              >
                <ProductShot id={item.product_id} name={item.product_name} src={item.image_url} size="sm" className="size-20 shrink-0 rounded-2xl bg-bone-2 p-2" />
                <div className="min-w-0 flex-1">
                  <Link to={`/shop/products/${item.product_id}`} className="line-clamp-2 font-medium hover:underline">
                    {item.product_name}
                  </Link>
                  <p className="tabular font-mono text-xs text-smoke">
                    {item.quantity} × {money(item.price)} · {item.sku}
                  </p>
                </div>
                <p className="tabular font-mono">{money(item.sub_total)}</p>
              </motion.li>
            ))}
          </ul>
          <div className="flex items-baseline justify-between py-5">
            <span className="font-mono text-[11px] uppercase tracking-[0.18em] text-smoke">Total</span>
            <span className="tabular font-display text-3xl font-extrabold tracking-tight">{money(data.total_amount)}</span>
          </div>
        </section>

        <aside className="rounded-[2rem] border border-ink/10 p-6" aria-label="Payment">
          <Eyebrow>Payment</Eyebrow>
          {payment.isLoading ? (
            <Skel className="mt-4 h-24" />
          ) : payment.data ? (
            <dl className="mt-4 space-y-3 text-sm">
              <div className="flex items-center justify-between">
                <dt className="text-smoke">Status</dt>
                <dd
                  className={cn(
                    "rounded-full px-2.5 py-0.5 font-mono text-[11px] uppercase tracking-wider",
                    payment.data.status === "SUCCESS" ? "bg-acid" : payment.data.status === "FAILED" ? "bg-signal/15 text-signal" : "bg-ink/10",
                  )}
                >
                  {PAYMENT_STATUS_LABEL[payment.data.status]}
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-smoke">Amount</dt>
                <dd className="tabular font-mono">{money(payment.data.amount, payment.data.currency)}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-smoke">Method</dt>
                <dd className="font-mono">{payment.data.gateway}</dd>
              </div>
              {payment.data.transaction_id && (
                <div className="flex justify-between gap-3">
                  <dt className="text-smoke">Reference</dt>
                  <dd className="truncate font-mono text-xs">{payment.data.transaction_id}</dd>
                </div>
              )}
              {payment.data.status === "REFUNDED" && <p className="rounded-2xl bg-ink/5 p-3 text-xs">Refunded to your original payment method.</p>}
            </dl>
          ) : (
            <p className="mt-4 text-sm text-smoke">No payment has been started for this order.</p>
          )}
          {!awaitingPayment && data.status !== "CANCELLED" && payment.data?.status === "SUCCESS" && (
            <p className="mt-4 text-xs text-smoke">Need a refund? Contact support with your order number.</p>
          )}
        </aside>
      </div>

      <Dialog.Root open={confirmCancel} onOpenChange={setConfirmCancel}>
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-50 bg-ink/50" />
          <Dialog.Content className="store fixed left-1/2 top-1/2 z-50 w-[min(440px,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 rounded-3xl bg-bone p-7 focus:outline-none">
            <Dialog.Title className="font-display text-2xl font-bold tracking-tight">Cancel this order?</Dialog.Title>
            <Dialog.Description className="mt-2 text-smoke">The reserved items go back to stock. This cannot be undone.</Dialog.Description>
            <div className="mt-6 flex justify-end gap-2">
              <Dialog.Close className={pill("ghost", "sm")}>Keep it</Dialog.Close>
              <button className={pill("ink", "sm")} disabled={cancel.isPending} onClick={() => cancel.mutate()}>
                {cancel.isPending ? "Cancelling…" : "Cancel order"}
              </button>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </div>
  );
}
