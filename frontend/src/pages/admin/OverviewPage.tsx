import { useMemo } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { AlertTriangle, ArrowRight, Boxes, CreditCard, IndianRupee, ShoppingCart, Users } from "lucide-react";
import { Bar, BarChart, CartesianGrid, Cell, LabelList, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { inventoryApi, ordersApi, paymentsApi, productsApi, usersApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { compactMoney, dateTime, fullName, money, number, shortId } from "@/lib/format";
import { ALL_ORDER_STATUSES, ORDER_STATUS_LABEL } from "@/lib/orders";
import { Card, CardHeader, EmptyState, ErrorState, Skeleton, Table, Td, Th } from "@/components/ui/primitives";
import { OrderStatusBadge } from "@/components/ui/badge";
import { PageHeader, StatCard } from "@/components/common";
import { LOW_STOCK_THRESHOLD } from "./InventoryAdminPage";

const DAYS = 14;
const SERIES = "#4f46e5";
const AXIS = { fontSize: 12, fill: "#667085" };

function dayKey(value: Date) {
  return value.toISOString().slice(0, 10);
}

function ChartTooltip({ active, payload, label, format }: { active?: boolean; payload?: { value: number }[]; label?: string; format: (value: number) => string }) {
  if (!active || !payload?.length) return null;
  return (
    <div className="rounded-lg border border-border bg-surface px-3 py-2 text-xs shadow-pop">
      <p className="text-subtle">{label}</p>
      <p className="tabular mt-0.5 text-sm font-semibold text-foreground">{format(payload[0]!.value)}</p>
    </div>
  );
}

export function OverviewPage() {
  const summary = useQuery({ queryKey: ["payments", "summary"], queryFn: paymentsApi.summary });
  const payments = useQuery({ queryKey: ["payments"], queryFn: paymentsApi.list });
  const orderSummary = useQuery({ queryKey: ["orders", "summary"], queryFn: ordersApi.summary });
  const recentOrders = useQuery({ queryKey: ["orders", "recent"], queryFn: () => ordersApi.list(1, 8) });
  const inventory = useQuery({ queryKey: ["inventory"], queryFn: inventoryApi.list });
  const products = useQuery({ queryKey: ["products", { categoryId: "" }], queryFn: productsApi.list });
  const users = useQuery({ queryKey: ["users"], queryFn: usersApi.list });

  const userById = useMemo(() => new Map((users.data ?? []).map((user) => [user.id, user])), [users.data]);

  // Captured revenue per day over the last DAYS days.
  const revenueByDay = useMemo(() => {
    const today = new Date();
    const buckets = new Map<string, number>();
    for (let i = DAYS - 1; i >= 0; i--) {
      const day = new Date(today);
      day.setDate(today.getDate() - i);
      buckets.set(dayKey(day), 0);
    }
    for (const payment of payments.data ?? []) {
      if (payment.status !== "SUCCESS") continue;
      const key = payment.created_at?.slice(0, 10);
      if (key && buckets.has(key)) buckets.set(key, buckets.get(key)! + payment.amount);
    }
    return [...buckets].map(([key, revenue]) => ({
      day: new Date(key + "T00:00:00").toLocaleDateString("en-IN", { day: "numeric", month: "short" }),
      revenue: Math.round(revenue * 100) / 100,
    }));
  }, [payments.data]);

  const ordersByStatus = useMemo(
    () =>
      ALL_ORDER_STATUSES.map((status) => ({
        status,
        label: ORDER_STATUS_LABEL[status],
        count: orderSummary.data?.by_status[status] ?? 0,
      })),
    [orderSummary.data],
  );

  const lowStock = useMemo(
    () =>
      (inventory.data ?? [])
        .filter((item) => item.available_quantity <= LOW_STOCK_THRESHOLD)
        .sort((a, b) => a.available_quantity - b.available_quantity),
    [inventory.data],
  );

  const missingInventory = useMemo(() => {
    const tracked = new Set((inventory.data ?? []).map((item) => item.product_id));
    return (products.data ?? []).filter((product) => !tracked.has(product.id)).length;
  }, [inventory.data, products.data]);

  const openOrders = (orderSummary.data?.by_status.PAID ?? 0) + (orderSummary.data?.by_status.PACKED ?? 0);
  const customers = (users.data ?? []).filter((user) => user.role === "customer").length;
  const revenueWindow = revenueByDay.reduce((sum, day) => sum + day.revenue, 0);

  const firstError = [summary, orderSummary, inventory].find((query) => query.isError)?.error;

  return (
    <div className="animate-in">
      <PageHeader title="Overview" description="Orders, payments and stock across the store." />

      {firstError && (
        <div className="mb-6">
          <ErrorState message={errorMessage(firstError)} />
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Captured revenue"
          value={summary.data ? money(summary.data.revenue) : <Skeleton className="h-8 w-28" />}
          hint={summary.data ? `${number(summary.data.successful_payments)} successful payments` : undefined}
          icon={<IndianRupee />}
          tone="success"
        />
        <StatCard
          label="Total orders"
          value={orderSummary.data ? number(orderSummary.data.total_orders) : <Skeleton className="h-8 w-20" />}
          hint={orderSummary.data ? `${number(openOrders)} awaiting fulfilment` : undefined}
          icon={<ShoppingCart />}
        />
        <StatCard
          label="Low-stock products"
          value={inventory.data ? number(lowStock.length) : <Skeleton className="h-8 w-16" />}
          hint={inventory.data ? `At or below ${LOW_STOCK_THRESHOLD} units${missingInventory ? ` · ${missingInventory} untracked` : ""}` : undefined}
          icon={<AlertTriangle />}
          tone={lowStock.length ? "warning" : "info"}
        />
        <StatCard
          label="Customers"
          value={users.data ? number(customers) : <Skeleton className="h-8 w-16" />}
          hint={products.data ? `${number(products.data.length)} products in catalogue` : undefined}
          icon={<Users />}
          tone="info"
        />
      </div>

      <div className="mt-6 grid gap-6 xl:grid-cols-[1.6fr_1fr]">
        <Card>
          <CardHeader
            title="Captured revenue"
            description={`Last ${DAYS} days · ${money(revenueWindow)}`}
          />
          <div className="h-72 px-2 pb-4 pt-4">
            {payments.isLoading ? (
              <Skeleton className="mx-4 h-full" />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={revenueByDay} margin={{ top: 8, right: 16, left: 0, bottom: 0 }}>
                  <CartesianGrid vertical={false} stroke="#eaecf0" />
                  <XAxis dataKey="day" tick={AXIS} tickLine={false} axisLine={{ stroke: "#d0d5dd" }} interval="preserveStartEnd" minTickGap={16} />
                  <YAxis tick={AXIS} tickLine={false} axisLine={false} width={64} tickFormatter={(value: number) => compactMoney(value)} />
                  <Tooltip cursor={{ fill: "#f2f4f7" }} content={<ChartTooltip format={money} />} />
                  <Bar dataKey="revenue" fill={SERIES} radius={[4, 4, 0, 0]} maxBarSize={28} animationDuration={400} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </Card>

        <Card>
          <CardHeader title="Orders by status" description={orderSummary.data ? `${number(orderSummary.data.total_orders)} orders` : undefined} />
          <div className="h-72 px-2 pb-4 pt-2">
            {orderSummary.isLoading ? (
              <Skeleton className="mx-4 h-full" />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={ordersByStatus} layout="vertical" margin={{ top: 8, right: 56, left: 8, bottom: 0 }} barCategoryGap={6}>
                  <CartesianGrid horizontal={false} stroke="#eaecf0" />
                  <XAxis type="number" allowDecimals={false} tick={AXIS} tickLine={false} axisLine={false} tickFormatter={(value: number) => (value >= 1000 ? `${number(value / 1000)}k` : number(value))} />
                  <YAxis type="category" dataKey="label" tick={AXIS} tickLine={false} axisLine={false} width={116} />
                  <Tooltip cursor={{ fill: "#f2f4f7" }} content={<ChartTooltip format={(value) => `${number(value)} orders`} />} />
                  <Bar dataKey="count" radius={[0, 4, 4, 0]} maxBarSize={22} animationDuration={400}>
                    {/* Counts are often skewed by one status, so small bars carry their value as a label. */}
                    <LabelList dataKey="count" position="right" offset={8} formatter={(value: number) => number(value)} style={{ fontSize: 12, fill: "#344054" }} />
                    {ordersByStatus.map((entry) => (
                      <Cell key={entry.status} fill={entry.status === "CANCELLED" ? "#98a2b3" : SERIES} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </Card>
      </div>

      {summary.data && (
        <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[
            { label: "Captured", value: summary.data.successful_payments, tone: "text-success" },
            { label: "Pending", value: summary.data.pending_payments, tone: "text-warning" },
            { label: "Failed", value: summary.data.failed_payments, tone: "text-danger" },
            { label: "Refunded", value: summary.data.refunded_payments, tone: "text-subtle" },
          ].map((item) => (
            <Link key={item.label} to="/admin/payments" className="flex items-center justify-between rounded-xl border border-border bg-surface px-5 py-4 shadow-card hover:border-faint">
              <span className="flex items-center gap-2 text-sm text-subtle">
                <CreditCard className={`size-4 ${item.tone}`} /> {item.label} payments
              </span>
              <span className="tabular text-lg font-semibold">{number(item.value)}</span>
            </Link>
          ))}
        </div>
      )}

      <div className="mt-6 grid gap-6 xl:grid-cols-[1.6fr_1fr]">
        <Card>
          <CardHeader
            title="Recent orders"
            action={
              <Link to="/admin/orders" className="flex items-center gap-1 text-sm font-medium text-primary hover:underline">
                View all <ArrowRight className="size-4" />
              </Link>
            }
          />
          {recentOrders.isLoading ? (
            <div className="p-5">
              <Skeleton className="h-40" />
            </div>
          ) : (recentOrders.data ?? []).length === 0 ? (
            <EmptyState icon={<ShoppingCart />} title="No orders yet" />
          ) : (
            <Table>
              <thead>
                <tr>
                  <Th>Order</Th>
                  <Th>Customer</Th>
                  <Th>Status</Th>
                  <Th className="text-right">Total</Th>
                </tr>
              </thead>
              <tbody>
                {(recentOrders.data ?? []).map((order) => {
                  const customer = userById.get(order.user_id);
                  return (
                    <tr key={order.id} className="hover:bg-muted/40">
                      <Td>
                        <Link to={`/admin/orders/${order.id}`} className="font-medium hover:text-primary">
                          {shortId(order.id)}
                        </Link>
                        <p className="text-xs text-subtle">{dateTime(order.created_at)}</p>
                      </Td>
                      <Td className="max-w-48 truncate">{customer ? fullName(customer) : <span className="text-faint">Unknown</span>}</Td>
                      <Td>
                        <OrderStatusBadge status={order.status} />
                      </Td>
                      <Td className="tabular text-right font-medium">{money(order.total_amount)}</Td>
                    </tr>
                  );
                })}
              </tbody>
            </Table>
          )}
        </Card>

        <Card>
          <CardHeader
            title="Low-stock warnings"
            description={`Products at or below ${LOW_STOCK_THRESHOLD} available units`}
            action={
              <Link to="/admin/inventory" className="flex items-center gap-1 text-sm font-medium text-primary hover:underline">
                Manage <ArrowRight className="size-4" />
              </Link>
            }
          />
          {inventory.isLoading ? (
            <div className="p-5">
              <Skeleton className="h-40" />
            </div>
          ) : lowStock.length === 0 ? (
            <EmptyState icon={<Boxes />} title="Stock levels are healthy" description="No product is running low." />
          ) : (
            <ul className="divide-y divide-border">
              {lowStock.slice(0, 7).map((item) => (
                <li key={item.product_id} className="flex items-center justify-between gap-3 px-5 py-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{item.product_name}</p>
                    <p className="text-xs text-subtle">SKU {item.sku}</p>
                  </div>
                  <span
                    className={`tabular flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                      item.available_quantity <= 0 ? "bg-danger-soft text-danger" : "bg-warning-soft text-warning"
                    }`}
                  >
                    <AlertTriangle className="size-3" />
                    {item.available_quantity <= 0 ? "Out of stock" : `${item.available_quantity} left`}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </Card>
      </div>
    </div>
  );
}
