import { useState } from "react";
import { Link, useLocation, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { CheckCircle2, ChevronLeft, Clock, CreditCard, XCircle } from "lucide-react";
import { ordersApi, paymentsApi } from "@/lib/endpoints";
import { ApiError, errorMessage } from "@/lib/api";
import { dateTime, money, shortId } from "@/lib/format";
import { isAwaitingPayment } from "@/lib/orders";
import { usePayOrder } from "@/lib/usePayOrder";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, EmptyState, ErrorState, Skeleton } from "@/components/ui/primitives";
import { OrderStatusBadge, PaymentStatusBadge } from "@/components/ui/badge";
import { ConfirmDialog } from "@/components/ui/dialog";
import { ProductImage } from "@/components/shop";
import { OrderTimeline } from "@/components/OrderTimeline";

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
      toast.success("Order cancelled", { description: "Reserved items were returned to stock." });
      setConfirmCancel(false);
      await queryClient.invalidateQueries();
    },
    onError: (error) => toast.error("Could not cancel the order", { description: errorMessage(error) }),
  });

  if (order.isLoading) {
    return (
      <div className="mx-auto max-w-4xl space-y-4 px-4 py-10 sm:px-6">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-32 rounded-xl" />
        <Skeleton className="h-64 rounded-xl" />
      </div>
    );
  }

  if (order.isError || !order.data) {
    const missing = order.error instanceof ApiError && (order.error.status === 404 || order.error.status === 400);
    return (
      <div className="mx-auto max-w-3xl px-4 py-16">
        {missing ? (
          <EmptyState
            title="Order not found"
            action={
              <Link to="/shop/orders">
                <Button variant="secondary">Back to orders</Button>
              </Link>
            }
          />
        ) : (
          <ErrorState message={errorMessage(order.error)} onRetry={() => order.refetch()} />
        )}
      </div>
    );
  }

  const data = order.data;
  const awaitingPayment = isAwaitingPayment(data.status);
  const paymentFailed = payment.data?.status === "FAILED";

  return (
    <div className="mx-auto max-w-4xl px-4 py-10 sm:px-6">
      <Link to="/shop/orders" className="mb-6 inline-flex items-center gap-1 text-sm text-subtle hover:text-foreground">
        <ChevronLeft className="size-4" /> All orders
      </Link>

      {justPlaced && data.status === "PAID" && (
        <div className="animate-in mb-6 flex items-start gap-3 rounded-xl border border-success/20 bg-success-soft p-4">
          <CheckCircle2 className="mt-0.5 size-5 text-success" />
          <div>
            <p className="font-semibold text-success">Thank you! Your order is confirmed.</p>
            <p className="text-sm text-subtle">We've received your payment and will start packing your items.</p>
          </div>
        </div>
      )}

      {awaitingPayment && (
        <div className="mb-6 flex flex-col gap-4 rounded-xl border border-warning/25 bg-warning-soft p-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-start gap-3">
            {paymentFailed ? <XCircle className="mt-0.5 size-5 text-danger" /> : <Clock className="mt-0.5 size-5 text-warning" />}
            <div>
              <p className="font-semibold">{paymentFailed ? "Your last payment attempt failed" : "Payment pending"}</p>
              <p className="text-sm text-subtle">Your items are reserved. Complete payment to confirm the order.</p>
            </div>
          </div>
          <div className="flex gap-2">
            <Button variant="danger-outline" onClick={() => setConfirmCancel(true)} disabled={paying}>
              Cancel order
            </Button>
            <Button onClick={() => pay(data.id)} loading={paying}>
              <CreditCard /> {paymentFailed ? "Retry payment" : "Pay now"}
            </Button>
          </div>
        </div>
      )}

      <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Order {shortId(data.id)}</h1>
          <p className="text-sm text-subtle">Placed {dateTime(data.created_at)}</p>
        </div>
        <OrderStatusBadge status={data.status} />
      </div>

      <Card className="mb-6 p-6">
        <OrderTimeline status={data.status} />
      </Card>

      <div className="grid items-start gap-6 lg:grid-cols-[1fr_300px]">
        <Card>
          <CardHeader title="Items" description={`${data.items.length} ${data.items.length === 1 ? "product" : "products"}`} />
          <ul className="divide-y divide-border">
            {data.items.map((item) => (
              <li key={item.product_id} className="flex items-center gap-4 px-5 py-3">
                <ProductImage id={item.product_id} name={item.product_name} src={item.image_url} size="sm" className="size-12 rounded-md" />
                <div className="min-w-0 flex-1">
                  <Link to={`/shop/products/${item.product_id}`} className="block truncate text-sm font-medium hover:text-primary">
                    {item.product_name}
                  </Link>
                  <p className="tabular text-xs text-subtle">
                    {item.quantity} × {money(item.price)} · SKU {item.sku}
                  </p>
                </div>
                <p className="tabular text-sm font-semibold">{money(item.sub_total)}</p>
              </li>
            ))}
          </ul>
          <div className="flex justify-between border-t border-border px-5 py-4 font-semibold">
            <span>Total</span>
            <span className="tabular">{money(data.total_amount)}</span>
          </div>
        </Card>

        <Card className="h-fit">
          <CardHeader title="Payment" />
          <div className="space-y-3 px-5 py-4 text-sm">
            {payment.isLoading ? (
              <Skeleton className="h-16" />
            ) : payment.data ? (
              <>
                <div className="flex items-center justify-between">
                  <span className="text-subtle">Status</span>
                  <PaymentStatusBadge status={payment.data.status} />
                </div>
                <div className="flex justify-between">
                  <span className="text-subtle">Amount</span>
                  <span className="tabular font-medium">{money(payment.data.amount, payment.data.currency)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-subtle">Method</span>
                  <span>{payment.data.gateway}</span>
                </div>
                {payment.data.transaction_id && (
                  <div className="flex justify-between gap-3">
                    <span className="text-subtle">Reference</span>
                    <span className="truncate font-mono text-xs">{payment.data.transaction_id}</span>
                  </div>
                )}
                {payment.data.status === "REFUNDED" && (
                  <p className="rounded-md bg-muted px-3 py-2 text-xs text-subtle">
                    This payment was refunded to your original payment method.
                  </p>
                )}
              </>
            ) : (
              <p className="text-subtle">No payment has been started for this order.</p>
            )}
            {!awaitingPayment && data.status !== "CANCELLED" && payment.data?.status === "SUCCESS" && (
              <p className="text-xs text-subtle">Need a refund? Contact support with your order number.</p>
            )}
          </div>
        </Card>
      </div>

      <ConfirmDialog
        open={confirmCancel}
        onOpenChange={setConfirmCancel}
        title="Cancel this order?"
        description="The reserved items will be released back to stock. This cannot be undone."
        confirmLabel="Cancel order"
        destructive
        loading={cancel.isPending}
        onConfirm={() => cancel.mutate()}
      />
    </div>
  );
}
