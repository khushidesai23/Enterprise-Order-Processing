const inr = new Intl.NumberFormat("en-IN", {
  style: "currency",
  currency: "INR",
  maximumFractionDigits: 2,
});

export function money(amount: number, currency = "INR"): string {
  if (currency === "INR") return inr.format(amount);
  return new Intl.NumberFormat("en-IN", { style: "currency", currency }).format(amount);
}

// Axis-friendly rupees in Indian units: ₹950, ₹12k, ₹3.5L, ₹1.2Cr.
// (Intl's en-IN compact notation writes thousands as "T".)
export function compactMoney(amount: number): string {
  const trim = (value: number) => String(Number(value.toFixed(1)));
  if (amount >= 1e7) return `₹${trim(amount / 1e7)}Cr`;
  if (amount >= 1e5) return `₹${trim(amount / 1e5)}L`;
  if (amount >= 1e3) return `₹${trim(amount / 1e3)}k`;
  return `₹${trim(amount)}`;
}

export function number(value: number): string {
  return new Intl.NumberFormat("en-IN").format(value);
}

export function date(value: string | undefined): string {
  if (!value) return "-";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime()) || parsed.getFullYear() < 2000) return "-";
  return parsed.toLocaleDateString("en-IN", { day: "numeric", month: "short", year: "numeric" });
}

export function dateTime(value: string | undefined): string {
  if (!value) return "-";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime()) || parsed.getFullYear() < 2000) return "-";
  return parsed.toLocaleString("en-IN", {
    day: "numeric",
    month: "short",
    hour: "numeric",
    minute: "2-digit",
  });
}

// Short, human-friendly reference for UUIDs ("#3F2A91C0").
export function shortId(id: string): string {
  return "#" + id.replace(/-/g, "").slice(0, 8).toUpperCase();
}

export function fullName(user: { first_name: string; last_name: string }): string {
  return [user.first_name, user.last_name].filter(Boolean).join(" ");
}

export function initials(name: string): string {
  return (
    name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0]!.toUpperCase())
      .join("") || "?"
  );
}
