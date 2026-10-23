import { useEffect, useMemo, useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import * as Dropdown from "@radix-ui/react-dropdown-menu";
import { toast } from "sonner";
import { AlertTriangle, Boxes, MoreHorizontal, PackageMinus, PackagePlus, PackageX, Search, SlidersHorizontal } from "lucide-react";
import { inventoryApi, productsApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { number } from "@/lib/format";
import type { InventoryListItem, StockOperation } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, EmptyState, ErrorState, Field, Input, Skeleton, Table, Td, Th } from "@/components/ui/primitives";
import { StockBadge } from "@/components/ui/badge";
import { Dialog } from "@/components/ui/dialog";
import { PageHeader, Pagination, StatCard } from "@/components/common";

const PAGE_SIZE = 50;

export const LOW_STOCK_THRESHOLD = 10;

type Action =
  | { kind: "operation"; item: InventoryListItem; operation: StockOperation }
  | { kind: "adjust"; item: InventoryListItem }
  | { kind: "track"; productId: string; name: string };

const OPERATION_COPY: Record<StockOperation, { title: string; verb: string; help: string }> = {
  "add-stock": { title: "Add stock", verb: "Add", help: "Received goods: increases available units." },
  "remove-stock": { title: "Remove stock", verb: "Remove", help: "Damaged or lost goods: decreases available units." },
  reserve: { title: "Reserve stock", verb: "Reserve", help: "Moves units from available to reserved. Orders do this automatically." },
  release: { title: "Release reservation", verb: "Release", help: "Returns reserved units to available. Cancellations do this automatically." },
  confirm: { title: "Confirm reservation", verb: "Confirm", help: "Consumes reserved units as shipped. Shipping an order does this automatically." },
};

type Filter = "all" | "low" | "out";

export function InventoryAdminPage() {
  const inventory = useQuery({ queryKey: ["inventory"], queryFn: inventoryApi.list });
  const products = useQuery({ queryKey: ["products", { categoryId: "" }], queryFn: productsApi.list });
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<Filter>("all");
  const [action, setAction] = useState<Action | null>(null);
  const [page, setPage] = useState(1);

  useEffect(() => {
    setPage(1);
  }, [search, filter]);

  const items = inventory.data ?? [];
  const untracked = useMemo(() => {
    const tracked = new Set(items.map((item) => item.product_id));
    return (products.data ?? []).filter((product) => !tracked.has(product.id));
  }, [items, products.data]);

  const visible = useMemo(() => {
    const needle = search.trim().toLowerCase();
    return items
      .filter((item) => !needle || item.product_name.toLowerCase().includes(needle) || item.sku.toLowerCase().includes(needle))
      .filter((item) =>
        filter === "low" ? item.available_quantity > 0 && item.available_quantity <= LOW_STOCK_THRESHOLD : filter === "out" ? item.available_quantity <= 0 : true,
      )
      .sort((a, b) => a.available_quantity - b.available_quantity);
  }, [items, search, filter]);

  const lowCount = items.filter((item) => item.available_quantity > 0 && item.available_quantity <= LOW_STOCK_THRESHOLD).length;
  const outCount = items.filter((item) => item.available_quantity <= 0).length;
  const reservedTotal = items.reduce((sum, item) => sum + item.reserved_quantity, 0);

  return (
    <div className="animate-in">
      <PageHeader title="Inventory" description="Stock levels for every product. Orders reserve stock automatically; adjust it here." />

      <div className="mb-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Tracked products" value={number(items.length)} icon={<Boxes />} />
        <StatCard label="Low stock" value={number(lowCount)} hint={`1–${LOW_STOCK_THRESHOLD} units available`} icon={<AlertTriangle />} tone="warning" />
        <StatCard label="Out of stock" value={number(outCount)} icon={<PackageX />} tone="danger" />
        <StatCard label="Units reserved" value={number(reservedTotal)} hint="Held by open orders" icon={<PackageMinus />} tone="info" />
      </div>

      {untracked.length > 0 && (
        <Card className="mb-6 border-warning/30">
          <CardHeader
            title={`${untracked.length} ${untracked.length === 1 ? "product has" : "products have"} no inventory record`}
            description="Customers cannot order these until stock is tracked."
          />
          <ul className="divide-y divide-border">
            {untracked.map((product) => (
              <li key={product.id} className="flex items-center justify-between gap-3 px-5 py-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{product.name}</p>
                  <p className="text-xs text-subtle">SKU {product.sku}</p>
                </div>
                <Button size="sm" variant="secondary" onClick={() => setAction({ kind: "track", productId: product.id, name: product.name })}>
                  Start tracking
                </Button>
              </li>
            ))}
          </ul>
        </Card>
      )}

      <Card>
        <div className="flex flex-col gap-3 border-b border-border p-4 sm:flex-row sm:items-center">
          <div className="relative flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-faint" />
            <Input className="pl-9" placeholder="Search by product or SKU" value={search} onChange={(event) => setSearch(event.target.value)} aria-label="Search inventory" />
          </div>
          <div className="inline-flex rounded-lg border border-border bg-muted/60 p-1" role="tablist" aria-label="Stock filter">
            {(
              [
                ["all", "All"],
                ["low", `Low (${lowCount})`],
                ["out", `Out (${outCount})`],
              ] as [Filter, string][]
            ).map(([value, label]) => (
              <button
                key={value}
                role="tab"
                aria-selected={filter === value}
                onClick={() => setFilter(value)}
                className={cn(
                  "rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                  filter === value ? "bg-surface text-foreground shadow-sm" : "text-subtle hover:text-foreground",
                )}
              >
                {label}
              </button>
            ))}
          </div>
        </div>

        {inventory.isError ? (
          <div className="p-5">
            <ErrorState message={errorMessage(inventory.error)} onRetry={() => inventory.refetch()} />
          </div>
        ) : inventory.isLoading ? (
          <div className="p-5">
            <Skeleton className="h-64" />
          </div>
        ) : visible.length === 0 ? (
          <EmptyState icon={<Boxes />} title="Nothing to show" description="No inventory records match this filter." />
        ) : (
          <>
          <Table>
            <thead>
              <tr>
                <Th>Product</Th>
                <Th className="text-right">Available</Th>
                <Th className="text-right">Reserved</Th>
                <Th>Status</Th>
                <Th className="text-right">Actions</Th>
              </tr>
            </thead>
            <tbody>
              {visible.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE).map((item) => (
                <tr key={item.product_id} className="hover:bg-muted/40" data-testid="inventory-row">
                  <Td>
                    <p className="font-medium">{item.product_name}</p>
                    <p className="text-xs text-subtle">SKU {item.sku}</p>
                  </Td>
                  <Td className="tabular text-right font-semibold">{number(item.available_quantity)}</Td>
                  <Td className="tabular text-right text-subtle">{number(item.reserved_quantity)}</Td>
                  <Td>
                    <StockBadge available={item.available_quantity} threshold={LOW_STOCK_THRESHOLD} />
                  </Td>
                  <Td>
                    <div className="flex justify-end gap-1">
                      <Button size="sm" variant="secondary" onClick={() => setAction({ kind: "operation", item, operation: "add-stock" })}>
                        <PackagePlus /> Add
                      </Button>
                      <Button size="sm" variant="secondary" onClick={() => setAction({ kind: "operation", item, operation: "remove-stock" })}>
                        <PackageMinus /> Remove
                      </Button>
                      <RowMenu onPick={setAction} item={item} />
                    </div>
                  </Td>
                </tr>
              ))}
            </tbody>
          </Table>
          {visible.length > PAGE_SIZE && (
            <div className="flex items-center justify-between border-t border-border">
              <span className="pl-4 text-sm text-subtle">{number(visible.length)} products</span>
              <Pagination page={page} hasNext={page * PAGE_SIZE < visible.length} onChange={setPage} />
            </div>
          )}
          </>
        )}
      </Card>

      {action && <ActionDialog action={action} onClose={() => setAction(null)} />}
    </div>
  );
}

function RowMenu({ item, onPick }: { item: InventoryListItem; onPick: (action: Action) => void }) {
  const itemClass = "flex cursor-pointer items-center gap-2 rounded-md px-2.5 py-2 text-sm outline-none data-[highlighted]:bg-muted";
  return (
    <Dropdown.Root>
      <Dropdown.Trigger asChild>
        <Button variant="ghost" size="icon" aria-label={`More stock actions for ${item.product_name}`}>
          <MoreHorizontal />
        </Button>
      </Dropdown.Trigger>
      <Dropdown.Portal>
        <Dropdown.Content align="end" sideOffset={4} className="animate-in z-50 w-56 rounded-xl border border-border bg-surface p-1.5 shadow-pop">
          <Dropdown.Item className={itemClass} onSelect={() => onPick({ kind: "adjust", item })}>
            <SlidersHorizontal className="size-4 text-subtle" /> Set exact levels
          </Dropdown.Item>
          <Dropdown.Separator className="my-1 h-px bg-border" />
          <Dropdown.Label className="px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wider text-faint">Reservations</Dropdown.Label>
          {(["reserve", "release", "confirm"] as StockOperation[]).map((operation) => (
            <Dropdown.Item key={operation} className={itemClass} onSelect={() => onPick({ kind: "operation", item, operation })}>
              {OPERATION_COPY[operation].title}
            </Dropdown.Item>
          ))}
        </Dropdown.Content>
      </Dropdown.Portal>
    </Dropdown.Root>
  );
}

function ActionDialog({ action, onClose }: { action: Action; onClose: () => void }) {
  const queryClient = useQueryClient();
  const [quantity, setQuantity] = useState(action.kind === "track" ? "0" : "1");
  const [available, setAvailable] = useState(action.kind === "adjust" ? String(action.item.available_quantity) : "0");
  const [reserved, setReserved] = useState(action.kind === "adjust" ? String(action.item.reserved_quantity) : "0");

  const isWhole = (value: string, min: number) => /^\d+$/.test(value) && Number(value) >= min;
  const valid =
    action.kind === "adjust" ? isWhole(available, 0) && isWhole(reserved, 0) : isWhole(quantity, action.kind === "track" ? 0 : 1);

  const run = useMutation({
    mutationFn: () => {
      if (action.kind === "track") return inventoryApi.create(action.productId, Number(quantity));
      if (action.kind === "adjust") return inventoryApi.update(action.item.product_id, Number(available), Number(reserved));
      return inventoryApi.operate(action.item.product_id, action.operation, Number(quantity));
    },
    onSuccess: async () => {
      toast.success("Inventory updated");
      await queryClient.invalidateQueries();
      onClose();
    },
    onError: (error) => toast.error("Could not update inventory", { description: errorMessage(error) }),
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    if (valid) run.mutate();
  }

  const title =
    action.kind === "track" ? "Start tracking stock" : action.kind === "adjust" ? "Set exact stock levels" : OPERATION_COPY[action.operation].title;
  const name = action.kind === "track" ? action.name : action.item.product_name;
  const confirmLabel = action.kind === "operation" ? OPERATION_COPY[action.operation].verb : "Save";

  return (
    <Dialog
      open
      onOpenChange={(open) => !open && onClose()}
      title={title}
      description={name}
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={run.isPending}>
            Cancel
          </Button>
          <Button type="submit" form="stock-form" loading={run.isPending} disabled={!valid}>
            {confirmLabel}
          </Button>
        </>
      }
    >
      <form id="stock-form" onSubmit={submit} className="space-y-4" noValidate>
        {action.kind === "operation" && (
          <div className="flex gap-6 rounded-lg bg-muted px-4 py-3 text-sm">
            <span>
              Available <span className="tabular font-semibold">{number(action.item.available_quantity)}</span>
            </span>
            <span>
              Reserved <span className="tabular font-semibold">{number(action.item.reserved_quantity)}</span>
            </span>
          </div>
        )}
        {action.kind === "adjust" ? (
          <>
            <p className="text-sm text-subtle">Overwrites both counts. Prefer add/remove for routine changes so open orders stay consistent.</p>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label="Available" htmlFor="s-available">
                <Input id="s-available" type="number" min={0} step={1} value={available} onChange={(event) => setAvailable(event.target.value)} autoFocus />
              </Field>
              <Field label="Reserved" htmlFor="s-reserved">
                <Input id="s-reserved" type="number" min={0} step={1} value={reserved} onChange={(event) => setReserved(event.target.value)} />
              </Field>
            </div>
          </>
        ) : (
          <Field
            label={action.kind === "track" ? "Starting stock" : "Quantity"}
            htmlFor="s-quantity"
            hint={action.kind === "operation" ? OPERATION_COPY[action.operation].help : "Units available to sell right away."}
          >
            <Input id="s-quantity" type="number" min={action.kind === "track" ? 0 : 1} step={1} value={quantity} onChange={(event) => setQuantity(event.target.value)} autoFocus />
          </Field>
        )}
      </form>
    </Dialog>
  );
}
