import { useState } from "react";
import { Link } from "react-router-dom";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { ChevronRight, Package } from "lucide-react";
import { ORDER_PAGE_SIZE, ordersApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { date, money, shortId } from "@/lib/format";
import { isAwaitingPayment } from "@/lib/orders";
import { Button } from "@/components/ui/button";
import { Card, EmptyState, ErrorState, Skeleton } from "@/components/ui/primitives";
import { OrderStatusBadge } from "@/components/ui/badge";
import { PageHeader, Pagination } from "@/components/common";

export function OrdersPage() {
  const [page, setPage] = useState(1);
  const orders = useQuery({
    queryKey: ["orders", "me", page],
    queryFn: () => ordersApi.mine(page, ORDER_PAGE_SIZE),
    placeholderData: keepPreviousData,
  });

  return (
    <div className="mx-auto max-w-4xl px-4 py-10 sm:px-6">
      <PageHeader title="My orders" description="Track, pay for or cancel your orders." />
      {orders.isError ? (
        <ErrorState message={errorMessage(orders.error)} onRetry={() => orders.refetch()} />
      ) : orders.isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} className="h-20 rounded-xl" />
          ))}
        </div>
      ) : orders.data?.length === 0 && page === 1 ? (
        <Card>
          <EmptyState
            icon={<Package />}
            title="No orders yet"
            description="When you place an order it will show up here."
            action={
              <Link to="/shop">
                <Button>Browse products</Button>
              </Link>
            }
          />
        </Card>
      ) : (
        <Card>
          <ul className="divide-y divide-border">
            {orders.data?.map((order) => (
              <li key={order.id}>
                <Link
                  to={`/shop/orders/${order.id}`}
                  className="flex items-center gap-4 px-5 py-4 transition-colors hover:bg-muted/50"
                  data-testid="order-row"
                >
                  <div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary">
                    <Package className="size-5" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="font-medium">Order {shortId(order.id)}</p>
                    <p className="text-sm text-subtle">
                      {date(order.created_at)} · {order.total_items} {order.total_items === 1 ? "item" : "items"}
                    </p>
                  </div>
                  <div className="hidden text-right sm:block">
                    <p className="tabular font-semibold">{money(order.total_amount)}</p>
                    {isAwaitingPayment(order.status) && <p className="text-xs font-medium text-warning">Payment due</p>}
                  </div>
                  <OrderStatusBadge status={order.status} />
                  <ChevronRight className="size-4 text-faint" />
                </Link>
              </li>
            ))}
          </ul>
          <Pagination
            page={page}
            hasNext={(orders.data?.length ?? 0) === ORDER_PAGE_SIZE}
            onChange={setPage}
            loading={orders.isFetching}
          />
        </Card>
      )}
    </div>
  );
}
