import { useRef, useState, type MouseEvent, type ReactNode } from "react";
import { motion, useMotionValue, useReducedMotion, useSpring } from "motion/react";
import { Star } from "lucide-react";
import { money } from "@/lib/format";
import { cn } from "@/lib/utils";
import { ProductArt } from "@/components/common";

export const ease = [0.22, 1, 0.36, 1] as const;

// --- Identity -----------------------------------------------------------------

// The mark: a globe cut by a single meridian, the slash of the wordmark.
export function Mark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 32 32" className={className} aria-hidden fill="none">
      <circle cx="16" cy="16" r="13" stroke="currentColor" strokeWidth="2.2" />
      <ellipse cx="16" cy="16" rx="5.2" ry="13" stroke="currentColor" strokeWidth="1.6" opacity="0.55" />
      <path d="M21 3.5 11 28.5" stroke="var(--color-acid)" strokeWidth="3.2" strokeLinecap="round" />
    </svg>
  );
}

export function Wordmark({ className, inverse = false }: { className?: string; inverse?: boolean }) {
  return (
    <span className={cn("inline-flex items-center gap-2 font-display text-[1.35rem] font-extrabold tracking-[-0.04em]", className)}>
      <Mark className={cn("size-7", inverse ? "text-bone" : "text-ink")} />
      <span>
        meridian<span className={inverse ? "text-acid" : "text-ultra"}>/</span>
      </span>
    </span>
  );
}

// --- Motion primitives --------------------------------------------------------

// Words slide up from behind a mask, one after another. The reveal is
// triggered by the line itself: the words start outside their masks, so
// they would never be seen as "in view" on their own.
export function MaskText({
  text,
  as: Tag = "span",
  className,
  delay = 0,
}: {
  text: string;
  as?: "span" | "p" | "h1" | "h2" | "div";
  className?: string;
  delay?: number;
}) {
  const words = text.split(" ");
  const MotionTag = motion[Tag];
  return (
    <MotionTag className={className} initial="hidden" whileInView="shown" viewport={{ once: true, margin: "-5%" }}>
      {/* Screen readers get the sentence once; the animated words are decorative. */}
      <span className="sr-only">{text}</span>
      {words.map((word, index) => (
        <span key={index} aria-hidden className="inline-block overflow-hidden pb-[0.08em] align-bottom">
          <motion.span
            className="inline-block"
            variants={{
              hidden: { y: "110%" },
              shown: { y: "0%", transition: { duration: 0.9, ease, delay: delay + index * 0.07 } },
            }}
          >
            {word}
            {index < words.length - 1 ? " " : ""}
          </motion.span>
        </span>
      ))}
    </MotionTag>
  );
}

export function Reveal({
  children,
  className,
  delay = 0,
  y = 28,
}: {
  children: ReactNode;
  className?: string;
  delay?: number;
  y?: number;
}) {
  return (
    <motion.div
      className={className}
      initial={{ opacity: 0, y }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-8%" }}
      transition={{ duration: 0.8, ease, delay }}
    >
      {children}
    </motion.div>
  );
}

// Children drift towards the pointer while it is near them.
export function Magnetic({ children, strength = 0.28, className }: { children: ReactNode; strength?: number; className?: string }) {
  const reduce = useReducedMotion();
  const ref = useRef<HTMLDivElement>(null);
  const x = useSpring(useMotionValue(0), { stiffness: 220, damping: 15, mass: 0.3 });
  const y = useSpring(useMotionValue(0), { stiffness: 220, damping: 15, mass: 0.3 });

  function onMove(event: MouseEvent) {
    if (reduce || !ref.current) return;
    const rect = ref.current.getBoundingClientRect();
    x.set((event.clientX - rect.left - rect.width / 2) * strength);
    y.set((event.clientY - rect.top - rect.height / 2) * strength);
  }

  return (
    <motion.div
      ref={ref}
      className={cn("inline-block", className)}
      style={{ x, y }}
      onMouseMove={onMove}
      onMouseLeave={() => {
        x.set(0);
        y.set(0);
      }}
    >
      {children}
    </motion.div>
  );
}

export function Marquee({
  children,
  duration = 40,
  reverse = false,
  className,
}: {
  children: ReactNode;
  duration?: number;
  reverse?: boolean;
  className?: string;
}) {
  return (
    <div className={cn("marquee flex overflow-hidden", className)} aria-hidden>
      <div className="marquee-track flex shrink-0" data-reverse={reverse} style={{ ["--marquee-duration" as string]: `${duration}s` }}>
        <div className="flex shrink-0 items-center">{children}</div>
        <div className="flex shrink-0 items-center">{children}</div>
      </div>
    </div>
  );
}

// --- Buttons ------------------------------------------------------------------

type Tone = "acid" | "ink" | "ghost" | "bone" | "line";

const TONES: Record<Tone, string> = {
  acid: "bg-acid text-ink hover:bg-acid-2",
  ink: "bg-ink text-bone hover:bg-ink-3",
  bone: "bg-bone text-ink hover:bg-white",
  ghost: "text-ink hover:bg-ink/5",
  line: "border border-ink/80 text-ink hover:bg-ink hover:text-bone",
};

export function pill(tone: Tone = "ink", size: "sm" | "md" | "lg" = "md") {
  return cn(
    "group/pill relative inline-flex items-center justify-center gap-2 overflow-hidden rounded-full font-medium transition-[background-color,color,transform] duration-200 active:scale-[0.97] disabled:pointer-events-none disabled:opacity-40 cursor-pointer",
    TONES[tone],
    size === "sm" && "h-9 px-4 text-sm",
    size === "md" && "h-11 px-5 text-[15px]",
    size === "lg" && "h-14 px-7 text-base",
  );
}

// --- Data display -------------------------------------------------------------

export function Eyebrow({ children, className }: { children: ReactNode; className?: string }) {
  return <p className={cn("font-mono text-[11px] uppercase tracking-[0.18em] text-smoke", className)}>{children}</p>;
}

export function discountOf(price: number, mrp: number) {
  return mrp > price ? Math.round(((mrp - price) / mrp) * 100) : 0;
}

export function Price({
  price,
  mrp,
  size = "md",
  className,
}: {
  price: number;
  mrp: number;
  size?: "sm" | "md" | "lg";
  className?: string;
}) {
  const off = discountOf(price, mrp);
  return (
    <div className={cn("flex flex-wrap items-baseline gap-x-2 gap-y-0.5", className)}>
      <span
        className={cn(
          "tabular font-display font-bold tracking-tight",
          size === "sm" && "text-[15px]",
          size === "md" && "text-lg",
          size === "lg" && "text-4xl",
        )}
      >
        {money(price)}
      </span>
      {off > 0 && (
        <>
          <span className={cn("tabular font-mono text-smoke line-through", size === "lg" ? "text-sm" : "text-[11px]")}>{money(mrp)}</span>
          <span className={cn("font-mono font-medium text-signal", size === "lg" ? "text-sm" : "text-[11px]")}>-{off}%</span>
        </>
      )}
    </div>
  );
}

// A round, slightly rotated sticker, used for discounts.
export function Sticker({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <span
      className={cn(
        "inline-flex size-16 rotate-[-12deg] items-center justify-center rounded-full bg-acid text-center font-display text-sm font-extrabold leading-none text-ink shadow-[0_6px_20px_-6px_rgb(0_0_0/0.35)]",
        className,
      )}
    >
      {children}
    </span>
  );
}

export function Rating({ average, count, className }: { average: number; count: number; className?: string }) {
  if (count === 0) return <span className={cn("font-mono text-[11px] text-smoke", className)}>NO REVIEWS YET</span>;
  return (
    <span className={cn("inline-flex items-center gap-1.5 font-mono text-[11px] text-smoke", className)} aria-label={`Rated ${average.toFixed(1)} out of 5 from ${count} reviews`}>
      <Star className="size-3.5 fill-ink text-ink" />
      <span className="text-ink">{average.toFixed(1)}</span>
      <span>({count})</span>
    </span>
  );
}

export function StarRow({ value, size = 16, className }: { value: number; size?: number; className?: string }) {
  return (
    <span className={cn("inline-flex items-center gap-0.5", className)} role="img" aria-label={`${value.toFixed(1)} out of 5 stars`}>
      {[1, 2, 3, 4, 5].map((star) => {
        const fill = Math.max(0, Math.min(1, value - star + 1));
        return (
          <span key={star} className="relative inline-block" style={{ width: size, height: size }}>
            <Star className="absolute inset-0 text-ink/25" style={{ width: size, height: size }} strokeWidth={1.5} />
            <span className="absolute inset-0 overflow-hidden" style={{ width: `${fill * 100}%` }}>
              <Star className="fill-ink text-ink" style={{ width: size, height: size }} strokeWidth={1.5} />
            </span>
          </span>
        );
      })}
    </span>
  );
}

// Stock as a short bar: honest about scarcity without fake urgency.
export function StockMeter({ available, className }: { available: number; className?: string }) {
  const level = available <= 0 ? 0 : available <= 10 ? 1 : available <= 50 ? 2 : 3;
  const label = available <= 0 ? "Sold out" : available <= 10 ? `Only ${available} left` : "In stock";
  return (
    <div className={cn("flex items-center gap-2", className)}>
      <span className="flex gap-0.5" aria-hidden>
        {[1, 2, 3].map((bar) => (
          <span key={bar} className={cn("h-3 w-1.5 rounded-sm", bar <= level ? (level === 1 ? "bg-signal" : "bg-ink") : "bg-ink/15")} />
        ))}
      </span>
      <span className={cn("font-mono text-xs uppercase tracking-wider", level === 0 && "text-signal", level === 1 && "text-signal")}>{label}</span>
    </div>
  );
}

// --- Imagery ------------------------------------------------------------------

// Product photo on a tinted plate. Photos are transparent cut-outs, so
// they sit on the plate; missing or broken images fall back to generated
// art.
export function ProductShot({
  id,
  name,
  src,
  className,
  imgClassName,
  size = "md",
  eager = false,
}: {
  id: string;
  name: string;
  src?: string;
  className?: string;
  imgClassName?: string;
  size?: "sm" | "md" | "lg";
  eager?: boolean;
}) {
  const [failed, setFailed] = useState(false);
  const [loaded, setLoaded] = useState(false);

  if (!src || failed) return <ProductArt id={id} name={name} size={size} className={className} />;

  return (
    <div className={cn("relative flex items-center justify-center overflow-hidden", className)}>
      {!loaded && <div className="absolute inset-0 animate-pulse bg-ink/[0.04]" />}
      <img
        src={src}
        alt={name}
        loading={eager ? "eager" : "lazy"}
        decoding="async"
        onLoad={() => setLoaded(true)}
        onError={() => setFailed(true)}
        className={cn("h-full w-full object-contain transition-opacity duration-500", loaded ? "opacity-100" : "opacity-0", imgClassName)}
      />
    </div>
  );
}

// Colour, storage and pack variants share photos; showcases keep one each.
export function distinct<T extends { id: string; image_url: string }>(items: T[] | undefined, max = Infinity): T[] {
  const seen = new Set<string>();
  const result: T[] = [];
  for (const item of items ?? []) {
    const key = item.image_url || item.id;
    if (seen.has(key)) continue;
    seen.add(key);
    result.push(item);
    if (result.length >= max) break;
  }
  return result;
}

// "Realme C35 (256 GB, Glacier Blue)" -> "Realme C35".
export function baseName(name: string) {
  return name.replace(/\s*\([^)]*\)\s*$/, "");
}

// "Realme C35 (256 GB, Glacier Blue)" -> "256 GB, Glacier Blue".
export function variantLabel(name: string) {
  return name.match(/\(([^)]*)\)\s*$/)?.[1] ?? "Standard";
}

export function Skel({ className }: { className?: string }) {
  return <div className={cn("animate-pulse rounded-2xl bg-ink/[0.06]", className)} />;
}
