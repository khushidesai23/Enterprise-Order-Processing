import { motion } from "motion/react";
import { Check, CircleX, CreditCard, Home, Package, PackageCheck, Truck } from "lucide-react";
import { FULFILMENT_STEPS, ORDER_STATUS_LABEL, fulfilmentStep } from "@/lib/orders";
import type { OrderStatus } from "@/lib/types";
import { cn } from "@/lib/utils";
import { ease } from "./ui";

const TONE: Record<OrderStatus, string> = {
  CREATED: "bg-signal/15 text-signal",
  PAYMENT_PENDING: "bg-signal/15 text-signal",
  PAID: "bg-ultra/15 text-ultra",
  PACKED: "bg-ultra/15 text-ultra",
  SHIPPED: "bg-ink text-acid",
  DELIVERED: "bg-acid text-ink",
  CANCELLED: "bg-ink/10 text-smoke",
};

export function StatusChip({ status }: { status: OrderStatus }) {
  return (
    <span className={cn("inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-2.5 py-1 font-mono text-[10px] uppercase tracking-[0.14em]", TONE[status])}>
      <span className="size-1.5 rounded-full bg-current" />
      {ORDER_STATUS_LABEL[status]}
    </span>
  );
}

const STEP_ICON = [Package, CreditCard, PackageCheck, Truck, Home];
const STEP_LABEL: Record<string, string> = { CREATED: "Placed", PAID: "Paid", PACKED: "Packed", SHIPPED: "Shipped", DELIVERED: "Delivered" };

// The order lifecycle from the backend, drawn as a track that fills to
// the current status.
export function OrderTrack({ status }: { status: OrderStatus }) {
  if (status === "CANCELLED") {
    return (
      <div className="flex items-center gap-4 rounded-3xl bg-ink/5 p-6">
        <CircleX className="size-8 shrink-0 text-smoke" />
        <div>
          <p className="font-display text-xl font-bold tracking-tight">This order was cancelled.</p>
          <p className="text-sm text-smoke">Any reserved stock was released back to the index.</p>
        </div>
      </div>
    );
  }

  const current = fulfilmentStep(status);
  const progress = current / (FULFILMENT_STEPS.length - 1);

  return (
    <div aria-label={`Order status: ${ORDER_STATUS_LABEL[status]}`} role="img">
      <div className="relative mx-[10%]">
        <div className="absolute inset-x-0 top-6 h-1 rounded-full bg-ink/10" />
        <motion.div
          className="absolute left-0 top-6 h-1 rounded-full bg-ink"
          initial={{ width: 0 }}
          animate={{ width: `${progress * 100}%` }}
          transition={{ duration: 1.2, ease, delay: 0.2 }}
        />
      </div>
      <ol className="relative grid grid-cols-5">
        {FULFILMENT_STEPS.map((step, index) => {
          const done = index <= current;
          const Icon = done && index < current ? Check : STEP_ICON[index]!;
          return (
            <li key={step} className="flex flex-col items-center text-center">
              <motion.span
                initial={{ scale: 0.6, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ duration: 0.5, ease, delay: 0.15 + index * 0.12 }}
                className={cn(
                  "relative z-10 flex size-12 items-center justify-center rounded-full border-2 transition-colors",
                  done ? "border-ink bg-ink text-acid" : "border-ink/15 bg-bone text-smoke",
                  index === current && "ring-8 ring-acid/60",
                )}
              >
                <Icon className="size-5" />
              </motion.span>
              <span className={cn("mt-3 font-mono text-[10px] uppercase tracking-[0.14em] sm:text-[11px]", done ? "text-ink" : "text-smoke")}>
                {STEP_LABEL[step]}
              </span>
            </li>
          );
        })}
      </ol>
    </div>
  );
}

// Small dots for list rows.
export function MiniTrack({ status }: { status: OrderStatus }) {
  if (status === "CANCELLED") return null;
  const current = fulfilmentStep(status);
  return (
    <span className="flex items-center gap-1" aria-hidden>
      {FULFILMENT_STEPS.map((step, index) => (
        <span key={step} className={cn("h-1.5 w-5 rounded-full", index <= current ? "bg-ink" : "bg-ink/15")} />
      ))}
    </span>
  );
}
