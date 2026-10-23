import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { ORDER_STATUS_LABEL, PAYMENT_STATUS_LABEL } from "@/lib/orders";
import type { OrderStatus, PaymentStatus } from "@/lib/types";

export type Tone = "neutral" | "primary" | "success" | "warning" | "danger" | "info";

const TONES: Record<Tone, string> = {
  neutral: "bg-muted text-subtle ring-border",
  primary: "bg-primary-soft text-primary ring-primary/20",
  success: "bg-success-soft text-success ring-success/20",
  warning: "bg-warning-soft text-warning ring-warning/20",
  danger: "bg-danger-soft text-danger ring-danger/20",
  info: "bg-info-soft text-info ring-info/20",
};

export function Badge({ tone = "neutral", children, className }: { tone?: Tone; children: ReactNode; className?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset",
        TONES[tone],
        className,
      )}
    >
      {children}
    </span>
  );
}

const ORDER_TONE: Record<OrderStatus, Tone> = {
  CREATED: "warning",
  PAYMENT_PENDING: "warning",
  PAID: "primary",
  PACKED: "info",
  SHIPPED: "info",
  DELIVERED: "success",
  CANCELLED: "neutral",
};

const PAYMENT_TONE: Record<PaymentStatus, Tone> = {
  PENDING: "warning",
  SUCCESS: "success",
  FAILED: "danger",
  REFUNDED: "neutral",
};

function Dot() {
  return <span className="size-1.5 rounded-full bg-current" />;
}

export function OrderStatusBadge({ status }: { status: OrderStatus }) {
  return (
    <Badge tone={ORDER_TONE[status]}>
      <Dot />
      {ORDER_STATUS_LABEL[status]}
    </Badge>
  );
}

export function PaymentStatusBadge({ status }: { status: PaymentStatus }) {
  return (
    <Badge tone={PAYMENT_TONE[status]}>
      <Dot />
      {PAYMENT_STATUS_LABEL[status]}
    </Badge>
  );
}

export function StockBadge({ available, threshold }: { available: number; threshold: number }) {
  if (available <= 0) return <Badge tone="danger">Out of stock</Badge>;
  if (available <= threshold) return <Badge tone="warning">Low stock · {available}</Badge>;
  return <Badge tone="success">In stock</Badge>;
}
