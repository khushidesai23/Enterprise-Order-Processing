import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { errorMessage } from "./api";
import { useAuth } from "./auth";
import { paymentsApi } from "./endpoints";
import { fullName } from "./format";
import { openCheckout } from "./razorpay";

// Runs the full payment flow for an order:
// POST /payments -> Razorpay Checkout -> POST /payments/verify.
// Returns true when the order is paid.
export function usePayOrder() {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const [paying, setPaying] = useState(false);

  async function pay(orderId: string): Promise<boolean> {
    setPaying(true);
    try {
      const session = await paymentsApi.checkout(orderId);
      const outcome = await openCheckout(session, {
        name: user ? fullName(user) : "",
        email: user?.email ?? "",
      });

      if (outcome.kind === "dismissed") {
        toast.info("Payment not completed", {
          description: "Your order is saved and its items are held. You can pay from your orders.",
        });
        return false;
      }

      if (outcome.kind === "failed") {
        toast.error("Payment failed", { description: `${outcome.reason}. You can try again.` });
        return false;
      }

      await paymentsApi.verify(outcome.result);
      toast.success("Payment received", { description: "Your order is confirmed." });
      return true;
    } catch (error) {
      toast.error("Payment could not be completed", { description: errorMessage(error) });
      return false;
    } finally {
      setPaying(false);
      await queryClient.invalidateQueries();
    }
  }

  return { pay, paying };
}
