import type { ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { motion } from "motion/react";
import { productsApi } from "@/lib/endpoints";
import { Marquee, ProductShot, distinct, ease } from "@/store/ui";

export function AuthShell({
  title,
  subtitle,
  children,
  footer,
}: {
  title: ReactNode;
  subtitle: string;
  children: ReactNode;
  footer: ReactNode;
}) {
  const products = useQuery({
    queryKey: ["search", "auth-wall"],
    queryFn: () => productsApi.search({ sort: "reviews", min_rating: 4, limit: 60 }),
    staleTime: 5 * 60_000,
  });
  const wall = distinct(products.data?.items, 12);

  return (
    <div className="mx-auto grid max-w-[1600px] gap-6 px-5 pb-10 pt-4 sm:px-8 lg:grid-cols-2">
      <div className="relative hidden min-h-[640px] flex-col justify-between overflow-hidden rounded-[2.5rem] bg-ink p-10 text-bone lg:flex">
        <div>
          <p className="font-mono text-[11px] uppercase tracking-[0.2em] text-bone/50">Members get</p>
          <ul className="mt-4 space-y-2 font-serif text-3xl italic leading-tight">
            <li>Stock held the moment you order.</li>
            <li>Payments verified, never assumed.</li>
            <li>Every order, tracked to the door.</li>
          </ul>
        </div>
        <div className="-mx-10 space-y-4">
          {[0, 1].map((row) => (
            <Marquee key={row} duration={row ? 38 : 30} reverse={row === 1}>
              {wall.slice(row * 6, row * 6 + 6).map((product) => (
                <span key={product.id} className="mx-2 block size-28 rounded-3xl bg-bone/95 p-3">
                  <ProductShot id={product.id} name={product.name} src={product.image_url} size="sm" eager className="size-full" />
                </span>
              ))}
            </Marquee>
          ))}
        </div>
        <p className="font-display text-7xl font-extrabold leading-none tracking-[-0.06em]">
          meridian<span className="text-acid">/</span>
        </p>
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease }}
        className="flex flex-col justify-center rounded-[2.5rem] bg-bone-2/60 px-6 py-12 sm:px-14"
      >
        <h1 className="font-display text-[clamp(2.8rem,5vw,4.5rem)] font-extrabold leading-[0.9] tracking-[-0.05em]">{title}</h1>
        <p className="mt-3 text-smoke">{subtitle}</p>
        <div className="mt-8 max-w-md">{children}</div>
        <p className="mt-8 text-sm text-smoke">{footer}</p>
      </motion.div>
    </div>
  );
}
