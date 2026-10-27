import { useMemo } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { ChevronRight, ShoppingCart } from "lucide-react";
import { ORDER_PAGE_SIZE, ordersApi, usersApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { dateTime, fullName, money, number, shortId } from "@/lib/format";
import { ALL_ORDER_STATUSES, ORDER_STATUS_LABEL } from "@/lib/orders";
import type { OrderStatus } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Card, EmptyState, ErrorState, Skeleton, Table, Td, Th } from "@/components/ui/primitives";
import { OrderStatusBadge } from "@/components/ui/badge";
import { PageHeader, Pagination } from "@/components/common";

export function OrdersAdminPage() {
  const [params, setParams] = useSearchParams();
  const status = (params.get("status") as OrderStatus | null) || undefined;
  const page = Math.max(1, Number(params.get("page")) || 1);

  const summary = useQuery({ queryKey: ["orders", "summary"], queryFn: ordersApi.summary });
  const orders = useQuery({
    queryKey: ["orders", "admin", status ?? "all", page],
    queryFn: () => ordersApi.list(page, ORDER_PAGE_SIZE, status),
    placeholderData: keepPreviousData,
  });
  const users = useQuery({ queryKey: ["users"], queryFn: usersApi.list });
  const userById = useMemo(() => new Map((users.data ?? []).map((user) => [user.id, user])), [users.data]);

  function go(next: { status?: OrderStatus; page?: number }) {
    const search = new URLSearchParams();
    if (next.status) search.set("status", next.status);
    if (next.page && next.page > 1) search.set("page", String(next.page));
    setParams(search);
  }

  const tabs: { value?: OrderStatus; label: string; count?: number }[] = [
    { label: "All", count: summary.data?.total_orders },
    ...ALL_ORDER_STATUSES.map((value) => ({ value, label: ORDER_STATUS_LABEL[value], count: summary.data?.by_status[value] })),
  ];

  return (
    <div className="animate-in">
      <PageHeader title="Orders" description="Fulfil paid orders, cancel unpaid ones and issue refunds." />

      <div className="-mx-1 mb-4 flex gap-1 overflow-x-auto px-1 pb-1" role="tablist" aria-label="Filter by status">
        {tabs.map((tab) => {
          const active = tab.value === status;
          return (
            <button
              key={tab.label}
              role="tab"
              aria-selected={active}
              onClick={() => go({ status: tab.value })}
              className={cn(
                "flex shrink-0 items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors",
                active ? "bg-foreground text-white" : "text-subtle hover:bg-muted hover:text-foreground",
              )}
            >
              {tab.label}
              {tab.count !== undefined && (
                <span className={cn("tabular rounded-full px-1.5 text-xs", active ? "bg-white/20" : "bg-muted")}>{number(tab.count)}</span>
              )}
            </button>
          );
        })}
      </div>

      <Card>
        {orders.isError ? (
          <div className="p-5">
            <ErrorState message={errorMessage(orders.error)} onRetry={() => orders.refetch()} />
          </div>
        ) : orders.isLoading ? (
          <div className="p-5">
            <Skeleton className="h-80" />
          </div>
        ) : orders.data?.length === 0 ? (
          <EmptyState icon={<ShoppingCart />} title="No orders here" description={status ? `No orders are ${ORDER_STATUS_LABEL[status].toLowerCase()}.` : undefined} />
        ) : (
          <>
            <Table>
              <thead>
                <tr>
                  <Th>Order</Th>
                  <Th>Customer</Th>
                  <Th>Placed</Th>
                  <Th className="text-right">Items</Th>
                  <Th className="text-right">Total</Th>
                  <Th>Status</Th>
                  <Th />
                </tr>
              </thead>
              <tbody className={cn(orders.isPlaceholderData && "opacity-60")}>
                {orders.data?.map((order) => {
                  const customer = userById.get(order.user_id);
                  return (
                    <tr key={order.id} className="hover:bg-muted/40" data-testid="admin-order-row">
                      <Td>
                        <Link to={`/admin/orders/${order.id}`} className="font-medium hover:text-primary">
                          {shortId(order.id)}
                        </Link>
                      </Td>
                      <Td className="max-w-56">
                        {customer ? (
                          <>
                            <p className="truncate">{fullName(customer)}</p>
                            <p className="truncate text-xs text-subtle">{customer.email}</p>
                          </>
                        ) : (
                          <span className="text-faint">Deleted user</span>
                        )}
                      </Td>
                      <Td className="whitespace-nowrap text-subtle">{dateTime(order.created_at)}</Td>
                      <Td className="tabular text-right">{order.total_items}</Td>
                      <Td className="tabular text-right font-medium">{money(order.total_amount)}</Td>
                      <Td>
                        <OrderStatusBadge status={order.status} />
                      </Td>
                      <Td className="text-right">
                        <Link to={`/admin/orders/${order.id}`} aria-label={`Open order ${shortId(order.id)}`} className="inline-flex text-faint hover:text-foreground">
                          <ChevronRight className="size-4" />
                        </Link>
                      </Td>
                    </tr>
                  );
                })}
              </tbody>
            </Table>
            <Pagination
              page={page}
              hasNext={(orders.data?.length ?? 0) === ORDER_PAGE_SIZE}
              onChange={(next) => go({ status, page: next })}
              loading={orders.isFetching}
            />
          </>
        )}
      </Card>
    </div>
  );
}
