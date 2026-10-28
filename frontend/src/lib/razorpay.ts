import type { CheckoutResult, CheckoutSession } from "./types";

const SCRIPT_URL = "https://checkout.razorpay.com/v1/checkout.js";

interface RazorpayFailure {
  error?: { description?: string; reason?: string };
}

interface RazorpayInstance {
  open(): void;
  on(event: "payment.failed", handler: (response: RazorpayFailure) => void): void;
}

declare global {
  interface Window {
    Razorpay?: new (options: Record<string, unknown>) => RazorpayInstance;
  }
}

let scriptPromise: Promise<void> | null = null;

function loadScript(): Promise<void> {
  if (window.Razorpay) return Promise.resolve();
  if (!scriptPromise) {
    scriptPromise = new Promise<void>((resolve, reject) => {
      const script = document.createElement("script");
      script.src = SCRIPT_URL;
      script.async = true;
      script.onload = () => resolve();
      script.onerror = () => {
        scriptPromise = null;
        reject(new Error("Could not load Razorpay Checkout. Check your internet connection."));
      };
      document.body.appendChild(script);
    });
  }
  return scriptPromise;
}

export type CheckoutOutcome =
  | { kind: "paid"; result: CheckoutResult }
  | { kind: "dismissed" }
  | { kind: "failed"; reason: string };

// Opens Razorpay Checkout for a payment created by POST /payments and
// resolves once the customer pays, closes the window, or a payment
// attempt fails. Razorpay lets the customer retry inside the same window
// after a failure, so "failed" is only reported when the window closes.
export async function openCheckout(
  session: CheckoutSession,
  customer: { name: string; email: string },
): Promise<CheckoutOutcome> {
  await loadScript();
  const Razorpay = window.Razorpay;
  if (!Razorpay) throw new Error("Razorpay Checkout is unavailable.");

  return new Promise<CheckoutOutcome>((resolve) => {
    let lastFailure: string | null = null;

    const checkout = new Razorpay({
      key: session.key_id,
      order_id: session.gateway_order_id,
      amount: Math.round(session.amount * 100),
      currency: session.currency,
      name: "meridian/",
      description: `Order ${session.order_id.slice(0, 8).toUpperCase()}`,
      prefill: { name: customer.name, email: customer.email },
      notes: { order_id: session.order_id },
      theme: { color: "#0c0c0d" },
      handler: (response: CheckoutResult) => resolve({ kind: "paid", result: response }),
      modal: {
        ondismiss: () =>
          resolve(lastFailure ? { kind: "failed", reason: lastFailure } : { kind: "dismissed" }),
      },
    });

    checkout.on("payment.failed", (response) => {
      lastFailure = response.error?.description || response.error?.reason || "Payment failed";
    });

    checkout.open();
  });
}
