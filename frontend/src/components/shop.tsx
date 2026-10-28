import { useState } from "react";
import { Star } from "lucide-react";
import { cn } from "@/lib/utils";
import { ProductArt } from "./common";

// Product photo for admin tables; falls back to generated art when a
// product has no image or it fails to load (for example offline).
export function ProductImage({
  id,
  name,
  src,
  className,
  imgClassName,
  size = "md",
}: {
  id: string;
  name: string;
  src?: string;
  className?: string;
  imgClassName?: string;
  size?: "sm" | "md" | "lg";
}) {
  const [failed, setFailed] = useState(false);

  if (!src || failed) return <ProductArt id={id} name={name} size={size} className={className} />;

  return (
    <div className={cn("flex items-center justify-center overflow-hidden bg-white", className)}>
      <img
        src={src}
        alt={name}
        loading="lazy"
        decoding="async"
        onError={() => setFailed(true)}
        className={cn("h-full w-full object-contain", imgClassName)}
      />
    </div>
  );
}

export function Stars({ value, size = 14, className }: { value: number; size?: number; className?: string }) {
  return (
    <span className={cn("inline-flex items-center", className)} aria-label={`${value.toFixed(1)} out of 5 stars`} role="img">
      {[1, 2, 3, 4, 5].map((star) => {
        const fill = Math.max(0, Math.min(1, value - star + 1));
        return (
          <span key={star} className="relative inline-block" style={{ width: size, height: size }}>
            <Star className="absolute inset-0 text-amber-500/35" style={{ width: size, height: size }} strokeWidth={1.5} />
            <span className="absolute inset-0 overflow-hidden" style={{ width: `${fill * 100}%` }}>
              <Star className="fill-amber-500 text-amber-500" style={{ width: size, height: size }} strokeWidth={1.5} />
            </span>
          </span>
        );
      })}
    </span>
  );
}
