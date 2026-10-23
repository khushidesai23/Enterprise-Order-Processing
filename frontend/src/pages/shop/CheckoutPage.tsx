import { useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { CreditCard, Lock, ShieldCheck } from "lucide-react";
import { ordersApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useCart } from "@/lib/cart";
import { fullName, money } from "@/lib/format";
import { usePayOrder } from "@/lib/usePayOrder";
import type { Order } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card, CardHeader } from "@/components/ui/primitives";
import { ProductArt } from "@/components/common";
import { useCartAvailability } from "./CartPage";

const STEPS = ["Cart", "Review", "Payment", "Confirmation"];

export function CheckoutPage() {
  const { user } = useAuth();
  const { lines, subtotal, clear } = useCart();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { pay, paying } = usePayOrder();
  const { byId, issues } = useCartAvailability();
  const [placed, setPlaced] = useState<Order | null>(null);

  const placeOrder = useMutation({
    mutationFn: () => ordersApi.create(lines.map((line) => ({ product_id: line.productId, quantity: line.quantity }))),
    onSuccess: async (order) => {
      // The order now holds the stock; the cart has done its job.
      setPlaced(order);
      clear();
      await queryClient.invalidateQueries();
      const paid = await pay(order.id);
      navigate(`/shop/orders/${order.id}`, { replace: true, state: { justPlaced: true, paid } });
    },
    onError: (error) => toast.error("Could not place your order", { description: errorMessage(error) }),
  });

  if (lines.length === 0 && !placed) return <Navigate to="/shop/cart" replace />;

  const busy = placeOrder.isPending || paying;
  const estimated = lines.reduce((sum, line) => sum + (byId.get(line.productId)?.price ?? line.price) * line.quantity, 0);

  return (
    <div className="mx-auto max-w-5xl px-4 py-10 sm:px-6">
      <ol className="mb-8 flex items-center gap-2 text-sm">
        {STEPS.map((step, index) => (
          <li key={step} className="flex items-center gap-2">
            <span
              className={
                index <= 1
                  ? "flex size-6 items-center justify-center rounded-full bg-primary text-xs font-semibold text-white"
                  : "flex size-6 items-center justify-center rounded-full bg-muted text-xs font-semibold text-subtle"
              }
            >
              {index + 1}
            </span>
            <span className={index <= 1 ? "font-medium" : "text-subtle"}>{step}</span>
            {index < STEPS.length - 1 && <span className="mx-1 h-px w-6 bg-border sm:w-10" />}
          </li>
        ))}
      </ol>

      <div className="grid gap-8 lg:grid-cols-[1fr_360px]">
        <div className="space-y-6">
          <Card>
            <CardHeader title="Contact" description="Order updates are sent to this account." />
            <div className="px-5 py-4 text-sm">
              <p className="font-medium">{user ? fullName(user) : ""}</p>
              <p className="text-subtle">{user?.email}</p>
            </div>
          </Card>

          <Card>
            <CardHeader
              title="Items"
              action={
                <Link to="/shop/cart" className="text-sm text-primary hover:underline">
                  Edit cart
                </Link>
              }
            />
            <ul className="divide-y divide-border">
              {lines.map((line) => (
                <li key={line.productId} className="flex items-center gap-4 px-5 py-3">
                  <ProductArt id={line.productId} name={line.name} size="sm" className="size-12 rounded-md" />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">{line.name}</p>
                    <p className="text-xs text-subtle">Qty {line.quantity}</p>
                    {issues.get(line.productId) && <p className="text-xs text-warning">{issues.get(line.productId)}</p>}
                  </div>
                  <p className="tabular text-sm font-medium">
                    {money((byId.get(line.productId)?.price ?? line.price) * line.quantity)}
                  </p>
                </li>
              ))}
            </ul>
          </Card>

          <Card>
            <CardHeader title="Payment" description="You'll pay securely with Razorpay after placing the order." />
            <div className="flex items-start gap-3 px-5 py-4 text-sm text-subtle">
              <CreditCard className="mt-0.5 size-5 text-primary" />
              <p>
                UPI, cards, netbanking and wallets are supported. Your items are reserved as soon as the order is placed;
                if you close the payment window you can pay later from <span className="font-medium text-foreground">My orders</span>.
              </p>
            </div>
          </Card>
        </div>

        <Card className="h-fit p-5 lg:sticky lg:top-24">
          <h2 className="font-semibold">Summary</h2>
          <dl className="mt-4 space-y-2 text-sm">
            <div className="flex justify-between">
              <dt className="text-subtle">Items</dt>
              <dd className="tabular">{money(estimated || subtotal)}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-subtle">Shipping</dt>
              <dd className="text-success">Free</dd>
            </div>
            <div className="flex justify-between border-t border-border pt-3 text-base font-semibold">
              <dt>Total</dt>
              <dd className="tabular">{money(estimated || subtotal)}</dd>
            </div>
          </dl>
          <Button
            size="lg"
            className="mt-5 w-full"
            loading={busy}
            disabled={issues.size > 0 || lines.length === 0}
            onClick={() => placeOrder.mutate()}
          >
            <Lock /> {placeOrder.isPending ? "Placing order…" : paying ? "Waiting for payment…" : "Place order & pay"}
          </Button>
          <p className="mt-3 flex items-center justify-center gap-1.5 text-xs text-subtle">
            <ShieldCheck className="size-3.5 text-success" /> Payments secured by Razorpay
          </p>
        </Card>
      </div>
    </div>
  );
}
