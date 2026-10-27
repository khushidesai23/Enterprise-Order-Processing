import type { ReactNode } from "react";
import { CheckCircle2 } from "lucide-react";

const POINTS = [
  "Stock is held for you the moment you place an order",
  "Secure checkout with UPI, cards and netbanking via Razorpay",
  "Live order tracking from payment to delivery",
];

export function AuthShell({
  title,
  subtitle,
  children,
  footer,
}: {
  title: string;
  subtitle: string;
  children: ReactNode;
  footer: ReactNode;
}) {
  return (
    <div className="mx-auto grid max-w-5xl gap-10 px-4 py-12 sm:px-6 lg:grid-cols-2 lg:py-20">
      <div className="hidden flex-col justify-center lg:flex">
        <p className="text-sm font-semibold text-primary">Meridian Market</p>
        <h2 className="mt-2 text-3xl font-semibold tracking-tight">Everyday essentials, delivered with care.</h2>
        <ul className="mt-8 space-y-4">
          {POINTS.map((point) => (
            <li key={point} className="flex items-start gap-3 text-sm text-subtle">
              <CheckCircle2 className="mt-0.5 size-5 shrink-0 text-success" />
              {point}
            </li>
          ))}
        </ul>
      </div>
      <div className="animate-in rounded-2xl border border-border bg-surface p-6 shadow-card sm:p-8">
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        <p className="mt-1 text-sm text-subtle">{subtitle}</p>
        <div className="mt-6">{children}</div>
        <p className="mt-6 text-center text-sm text-subtle">{footer}</p>
      </div>
    </div>
  );
}
