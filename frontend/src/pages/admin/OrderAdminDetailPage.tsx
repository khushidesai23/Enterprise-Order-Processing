import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ChevronLeft, PackageCheck, RotateCcw, Truck, XCircle } from "lucide-react";
import { ordersApi, paymentsApi, usersApi } from "@/lib/endpoints";
import { ApiError, errorMessage } from "@/lib/api";
import { dateTime, fullName, money, shortId } from "@/lib/format";
import { ACTION_LABEL, adminNextStatuses } from "@/lib/orders";
import type { OrderStatus } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, EmptyState, ErrorState, Skeleton, Table, Td, Th } from "@/components/ui/primitives";
import { OrderStatusBadge, PaymentStatusBadge } from "@/components/ui/badge";
import { ConfirmDialog } from "@/components/ui/dialog";
import { PageHeader } from "@/components/common";
import { OrderTimeline } from "@/components/OrderTimeline";

const ACTION_ICON: Partial<Record<OrderStatus, React.ReactNode>> = {
  PACKED: <PackageCheck />,
  SHIPPED: <Truck />,
  DELIVERED: <PackageCheck />,
  CANCELLED: <XCircle />,
};

export function OrderAdminDetailPage() {
  const { id = "" } = useParams();
  const queryClient = useQueryClient();
  const [pendingStatus, setPendingStatus] = useState<OrderStatus | null>(null);
  const [confirmRefund, setConfirmRefund] = useState(false);

  const order = useQuery({ queryKey: ["order", id], queryFn: () => ordersApi.get(id) });
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
  const customer = useQuery({
    queryKey: ["user", order.data?.user_id],
    queryFn: () => usersApi.get(order.data!.user_id),
    enabled: !!order.data,
    retry: false,
  });

  const updateStatus = useMutation({
    mutationFn: (status: OrderStatus) => ordersApi.updateStatus(id, status),
    onSuccess: async (updated) => {
      toast.success(`Order ${shortId(updated.id)} updated`);
      setPendingStatus(null);
      await queryClient.invalidateQueries();
    },
    onError: (error) => toast.error("Could not update the order", { description: errorMessage(error) }),
  });

  const refund = useMutation({
    mutationFn: () => paymentsApi.refund(payment.data!.id),
    onSuccess: async () => {
      toast.success("Refund issued");
      setConfirmRefund(false);
      await queryClient.invalidateQueries();
    },
    onError: (error) => toast.error("Refund failed", { description: errorMessage(error) }),
  });

  if (order.isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-28 rounded-xl" />
        <Skeleton className="h-64 rounded-xl" />
      </div>
    );
  }

  if (order.isError || !order.data) {
    const missing = order.error instanceof ApiError && (order.error.status === 404 || order.error.status === 400);
    return missing ? (
      <EmptyState
        title="Order not found"
        action={
          <Link to="/admin/orders">
            <Button variant="secondary">Back to orders</Button>
          </Link>
        }
      />
    ) : (
      <ErrorState message={errorMessage(order.error)} onRetry={() => order.refetch()} />
    );
  }

  const data = order.data;
  const next = adminNextStatuses(data.status);
  const refundable = payment.data?.status === "SUCCESS";
  // Before shipment a refund also cancels the order and releases its stock.
  const refundCancels = data.status === "PAID" || data.status === "PACKED";

  return (
    <div className="animate-in">
      <Link to="/admin/orders" className="mb-4 inline-flex items-center gap-1 text-sm text-subtle hover:text-foreground">
        <ChevronLeft className="size-4" /> Orders
      </Link>
      <PageHeader
        title={
          <span className="flex flex-wrap items-center gap-3">
            Order {shortId(data.id)} <OrderStatusBadge status={data.status} />
          </span>
        }
        description={`Placed ${dateTime(data.created_at)} · ${data.id}`}
        actions={
          <>
            {next
              .filter((status) => status !== "CANCELLED")
              .map((status) => (
                <Button key={status} onClick={() => setPendingStatus(status)}>
                  {ACTION_ICON[status]} {ACTION_LABEL[status]}
                </Button>
              ))}
            {next.includes("CANCELLED") && (
              <Button variant="danger-outline" onClick={() => setPendingStatus("CANCELLED")}>
                <XCircle /> Cancel order
              </Button>
            )}
            {refundable && (
              <Button variant="danger-outline" onClick={() => setConfirmRefund(true)}>
                <RotateCcw /> {refundCancels ? "Refund & cancel" : "Refund payment"}
              </Button>
            )}
          </>
        }
      />

      <Card className="mb-6 p-6">
        <OrderTimeline status={data.status} />
      </Card>

      <div className="grid items-start gap-6 xl:grid-cols-[1fr_340px]">
        <Card>
          <CardHeader title="Line items" />
          <Table>
            <thead>
              <tr>
                <Th>Product</Th>
                <Th className="text-right">Qty</Th>
                <Th className="text-right">Unit price</Th>
                <Th className="text-right">Subtotal</Th>
              </tr>
            </thead>
            <tbody>
              {data.items.map((item) => (
                <tr key={item.product_id}>
                  <Td>
                    <p className="font-medium">{item.product_name}</p>
                    <p className="text-xs text-subtle">SKU {item.sku}</p>
                  </Td>
                  <Td className="tabular text-right">{item.quantity}</Td>
                  <Td className="tabular text-right">{money(item.price)}</Td>
                  <Td className="tabular text-right font-medium">{money(item.sub_total)}</Td>
                </tr>
              ))}
              <tr>
                <Td colSpan={3} className="text-right font-semibold">
                  Total
                </Td>
                <Td className="tabular text-right font-semibold">{money(data.total_amount)}</Td>
              </tr>
            </tbody>
          </Table>
        </Card>

        <div className="space-y-6">
          <Card>
            <CardHeader title="Customer" />
            <div className="px-5 py-4 text-sm">
              {customer.data ? (
                <>
                  <p className="font-medium">{fullName(customer.data)}</p>
                  <p className="text-subtle">{customer.data.email}</p>
                </>
              ) : customer.isLoading ? (
                <Skeleton className="h-10" />
              ) : (
                <p className="text-subtle">Customer account no longer exists.</p>
              )}
            </div>
          </Card>

          <Card>
            <CardHeader title="Payment" />
            <div className="space-y-3 px-5 py-4 text-sm">
              {payment.isLoading ? (
                <Skeleton className="h-20" />
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
                  <div className="flex justify-between gap-3">
                    <span className="text-subtle">Gateway order</span>
                    <span className="truncate font-mono text-xs">{payment.data.gateway_order_id}</span>
                  </div>
                  <div className="flex justify-between gap-3">
                    <span className="text-subtle">Transaction</span>
                    <span className="truncate font-mono text-xs">{payment.data.transaction_id || "—"}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-subtle">Created</span>
                    <span>{dateTime(payment.data.created_at)}</span>
                  </div>
                </>
              ) : (
                <p className="text-subtle">The customer has not started a payment.</p>
              )}
            </div>
          </Card>

          {(data.status === "CREATED" || data.status === "PAYMENT_PENDING") && (
            <p className="rounded-lg bg-muted px-4 py-3 text-xs text-subtle">
              Payment status is set by Razorpay. Orders move to <strong>Paid</strong> once the payment is captured.
            </p>
          )}
        </div>
      </div>

      <ConfirmDialog
        open={!!pendingStatus}
        onOpenChange={(open) => !open && setPendingStatus(null)}
        title={pendingStatus ? `${ACTION_LABEL[pendingStatus]}?` : ""}
        description={
          pendingStatus === "CANCELLED"
            ? "The order is cancelled and its reserved stock is released. This cannot be undone."
            : pendingStatus === "SHIPPED"
              ? "Shipping consumes the reserved stock. The order can no longer be cancelled afterwards."
              : "The customer will see the new status on their order."
        }
        confirmLabel={pendingStatus ? ACTION_LABEL[pendingStatus]! : ""}
        destructive={pendingStatus === "CANCELLED"}
        loading={updateStatus.isPending}
        onConfirm={() => pendingStatus && updateStatus.mutate(pendingStatus)}
      />

      <ConfirmDialog
        open={confirmRefund}
        onOpenChange={setConfirmRefund}
        title="Refund this payment?"
        description={
          refundCancels
            ? `${money(payment.data?.amount ?? 0)} is returned through Razorpay, the order is cancelled and its stock released.`
            : `${money(payment.data?.amount ?? 0)} is returned through Razorpay. The order keeps its fulfilment status.`
        }
        confirmLabel="Issue refund"
        destructive
        loading={refund.isPending}
        onConfirm={() => refund.mutate()}
      />
    </div>
  );
}
