import { Check, CircleX } from "lucide-react";
import { FULFILMENT_STEPS, ORDER_STATUS_LABEL, fulfilmentStep } from "@/lib/orders";
import type { OrderStatus } from "@/lib/types";
import { cn } from "@/lib/utils";

const STEP_LABEL: Record<string, string> = {
  CREATED: "Placed",
  PAID: "Paid",
  PACKED: "Packed",
  SHIPPED: "Shipped",
  DELIVERED: "Delivered",
};

export function OrderTimeline({ status }: { status: OrderStatus }) {
  if (status === "CANCELLED") {
    return (
      <div className="flex items-center gap-3 rounded-lg bg-muted px-4 py-3 text-sm text-subtle">
        <CircleX className="size-5 text-faint" />
        This order was cancelled. Any reserved items were returned to stock.
      </div>
    );
  }

  const current = fulfilmentStep(status);

  return (
    <ol className="flex items-start" aria-label={`Order status: ${ORDER_STATUS_LABEL[status]}`}>
      {FULFILMENT_STEPS.map((step, index) => {
        const done = index <= current;
        return (
          <li key={step} className="relative flex flex-1 flex-col items-center text-center">
            {index > 0 && (
              <span
                className={cn("absolute right-1/2 top-3.5 h-0.5 w-full -translate-y-1/2", index <= current ? "bg-primary" : "bg-border")}
              />
            )}
            <span
              className={cn(
                "relative z-10 flex size-7 items-center justify-center rounded-full border-2 text-xs font-semibold",
                done ? "border-primary bg-primary text-white" : "border-border bg-surface text-faint",
                index === current && "ring-4 ring-primary/15",
              )}
            >
              {done ? <Check className="size-3.5" /> : index + 1}
            </span>
            <span className={cn("mt-2 text-xs", done ? "font-medium text-foreground" : "text-subtle")}>{STEP_LABEL[step]}</span>
          </li>
        );
      })}
    </ol>
  );
}
