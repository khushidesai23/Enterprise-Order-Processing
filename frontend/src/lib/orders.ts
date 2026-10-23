import type { OrderStatus, PaymentStatus } from "./types";

// Mirrors validateStatusTransition in internal/modules/order/service.go.
// PAYMENT_PENDING and PAID are set by the payment flow, never by hand, and
// a paid order is cancelled by refunding its payment.
const ADMIN_TRANSITIONS: Record<OrderStatus, OrderStatus[]> = {
  CREATED: ["CANCELLED"],
  PAYMENT_PENDING: ["CANCELLED"],
  PAID: ["PACKED"],
  PACKED: ["SHIPPED"],
  SHIPPED: ["DELIVERED"],
  DELIVERED: [],
  CANCELLED: [],
};

export function adminNextStatuses(status: OrderStatus): OrderStatus[] {
  return ADMIN_TRANSITIONS[status];
}

// Orders a customer can still pay for or cancel themselves.
export function isAwaitingPayment(status: OrderStatus): boolean {
  return status === "CREATED" || status === "PAYMENT_PENDING";
}

export const FULFILMENT_STEPS: OrderStatus[] = ["CREATED", "PAID", "PACKED", "SHIPPED", "DELIVERED"];

// Index of the furthest fulfilment step reached; PAYMENT_PENDING counts as
// still at CREATED.
export function fulfilmentStep(status: OrderStatus): number {
  if (status === "PAYMENT_PENDING") return 0;
  return FULFILMENT_STEPS.indexOf(status);
}

export const ORDER_STATUS_LABEL: Record<OrderStatus, string> = {
  CREATED: "Awaiting payment",
  PAYMENT_PENDING: "Payment pending",
  PAID: "Paid",
  PACKED: "Packed",
  SHIPPED: "Shipped",
  DELIVERED: "Delivered",
  CANCELLED: "Cancelled",
};

export const PAYMENT_STATUS_LABEL: Record<PaymentStatus, string> = {
  PENDING: "Pending",
  SUCCESS: "Captured",
  FAILED: "Failed",
  REFUNDED: "Refunded",
};

export const ALL_ORDER_STATUSES: OrderStatus[] = [
  "CREATED",
  "PAYMENT_PENDING",
  "PAID",
  "PACKED",
  "SHIPPED",
  "DELIVERED",
  "CANCELLED",
];

export const ACTION_LABEL: Partial<Record<OrderStatus, string>> = {
  PACKED: "Mark as packed",
  SHIPPED: "Mark as shipped",
  DELIVERED: "Mark as delivered",
  CANCELLED: "Cancel order",
};
