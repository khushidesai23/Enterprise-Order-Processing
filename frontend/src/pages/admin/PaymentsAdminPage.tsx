import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { CheckCircle2, Clock, CreditCard, IndianRupee, RotateCcw, Search, XCircle } from "lucide-react";
import { paymentsApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { dateTime, money, number, shortId } from "@/lib/format";
import { PAYMENT_STATUS_LABEL } from "@/lib/orders";
import type { PaymentListItem, PaymentStatus } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, EmptyState, ErrorState, Input, Skeleton, Table, Td, Th } from "@/components/ui/primitives";
import { PaymentStatusBadge } from "@/components/ui/badge";
import { ConfirmDialog } from "@/components/ui/dialog";
import { PageHeader, Pagination, StatCard } from "@/components/common";

const PAGE_SIZE = 25;

export function PaymentsAdminPage() {
  const queryClient = useQueryClient();
  const summary = useQuery({ queryKey: ["payments", "summary"], queryFn: paymentsApi.summary });
  const payments = useQuery({ queryKey: ["payments"], queryFn: paymentsApi.list });
  const [status, setStatus] = useState<PaymentStatus | "">("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [refunding, setRefunding] = useState<PaymentListItem | null>(null);

  const filtered = useMemo(() => {
    const needle = search.trim().toLowerCase().replace(/^#/, "");
    return (payments.data ?? []).filter(
      (payment) =>
        (!status || payment.status === status) &&
        (!needle || payment.id.replace(/-/g, "").includes(needle) || payment.order_id.replace(/-/g, "").includes(needle)),
    );
  }, [payments.data, status, search]);
  const visible = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  const refund = useMutation({
    mutationFn: (id: string) => paymentsApi.refund(id),
    onSuccess: async () => {
      toast.success("Refund issued");
      setRefunding(null);
      await queryClient.invalidateQueries();
    },
    onError: (error) => toast.error("Refund failed", { description: errorMessage(error) }),
  });

  const filters: { value: PaymentStatus | ""; label: string }[] = [
    { value: "", label: "All" },
    ...(["SUCCESS", "PENDING", "FAILED", "REFUNDED"] as PaymentStatus[]).map((value) => ({ value, label: PAYMENT_STATUS_LABEL[value] })),
  ];

  return (
    <div className="animate-in">
      <PageHeader title="Payments" description="Razorpay payments for every order. Refund captured payments here." />

      <div className="mb-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
        <StatCard label="Revenue" value={summary.data ? money(summary.data.revenue) : <Skeleton className="h-8 w-24" />} icon={<IndianRupee />} tone="success" />
        <StatCard label="Captured" value={summary.data ? number(summary.data.successful_payments) : "–"} icon={<CheckCircle2 />} tone="success" />
        <StatCard label="Pending" value={summary.data ? number(summary.data.pending_payments) : "–"} icon={<Clock />} tone="warning" />
        <StatCard label="Failed" value={summary.data ? number(summary.data.failed_payments) : "–"} icon={<XCircle />} tone="danger" />
        <StatCard label="Refunded" value={summary.data ? number(summary.data.refunded_payments) : "–"} icon={<RotateCcw />} tone="info" />
      </div>

      <Card>
        <div className="flex flex-col gap-3 border-b border-border p-4 lg:flex-row lg:items-center">
          <div className="relative flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-faint" />
            <Input
              className="pl-9"
              placeholder="Search by payment or order reference"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value);
                setPage(1);
              }}
              aria-label="Search payments"
            />
          </div>
          <div className="inline-flex rounded-lg border border-border bg-muted/60 p-1" role="tablist" aria-label="Payment status">
            {filters.map((filter) => (
              <button
                key={filter.label}
                role="tab"
                aria-selected={status === filter.value}
                onClick={() => {
                  setStatus(filter.value);
                  setPage(1);
                }}
                className={cn(
                  "rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                  status === filter.value ? "bg-surface text-foreground shadow-sm" : "text-subtle hover:text-foreground",
                )}
              >
                {filter.label}
              </button>
            ))}
          </div>
        </div>

        {payments.isError ? (
          <div className="p-5">
            <ErrorState message={errorMessage(payments.error)} onRetry={() => payments.refetch()} />
          </div>
        ) : payments.isLoading ? (
          <div className="p-5">
            <Skeleton className="h-64" />
          </div>
        ) : visible.length === 0 ? (
          <EmptyState icon={<CreditCard />} title="No payments" description={payments.data?.length ? "Nothing matches these filters." : "Payments appear once customers check out."} />
        ) : (
          <>
            <Table>
              <thead>
                <tr>
                  <Th>Payment</Th>
                  <Th>Order</Th>
                  <Th>Created</Th>
                  <Th>Gateway</Th>
                  <Th className="text-right">Amount</Th>
                  <Th>Status</Th>
                  <Th className="text-right">Actions</Th>
                </tr>
              </thead>
              <tbody>
                {visible.map((payment) => (
                  <tr key={payment.id} className="hover:bg-muted/40" data-testid="admin-payment-row">
                    <Td className="font-medium">{shortId(payment.id)}</Td>
                    <Td>
                      <Link to={`/admin/orders/${payment.order_id}`} className="text-primary hover:underline">
                        {shortId(payment.order_id)}
                      </Link>
                    </Td>
                    <Td className="whitespace-nowrap text-subtle">{dateTime(payment.created_at)}</Td>
                    <Td className="text-subtle">{payment.gateway}</Td>
                    <Td className="tabular text-right font-medium">{money(payment.amount, payment.currency)}</Td>
                    <Td>
                      <PaymentStatusBadge status={payment.status} />
                    </Td>
                    <Td className="text-right">
                      {payment.status === "SUCCESS" ? (
                        <Button size="sm" variant="danger-outline" onClick={() => setRefunding(payment)}>
                          <RotateCcw /> Refund
                        </Button>
                      ) : (
                        <span className="text-faint">—</span>
                      )}
                    </Td>
                  </tr>
                ))}
              </tbody>
            </Table>
            <Pagination page={page} hasNext={page * PAGE_SIZE < filtered.length} onChange={setPage} />
          </>
        )}
      </Card>

      <ConfirmDialog
        open={!!refunding}
        onOpenChange={(open) => !open && setRefunding(null)}
        title="Refund this payment?"
        description={`${money(refunding?.amount ?? 0)} for order ${refunding ? shortId(refunding.order_id) : ""} is returned through Razorpay. Unshipped orders are also cancelled and their stock released.`}
        confirmLabel="Issue refund"
        destructive
        loading={refund.isPending}
        onConfirm={() => refunding && refund.mutate(refunding.id)}
      />
    </div>
  );
}
