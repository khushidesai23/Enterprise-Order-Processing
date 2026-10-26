import { useState } from "react";
import { Link } from "react-router-dom";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { motion } from "motion/react";
import { ArrowRight, ArrowUpRight } from "lucide-react";
import { ORDER_PAGE_SIZE, ordersApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { date, money, shortId } from "@/lib/format";
import { isAwaitingPayment } from "@/lib/orders";
import { cn } from "@/lib/utils";
import { MiniTrack, StatusChip } from "@/store/orders";
import { Eyebrow, MaskText, Skel, ease, pill } from "@/store/ui";

export function OrdersPage() {
  const [page, setPage] = useState(1);
  const orders = useQuery({
    queryKey: ["orders", "me", page],
    queryFn: () => ordersApi.mine(page, ORDER_PAGE_SIZE),
    placeholderData: keepPreviousData,
  });

  return (
    <div className="mx-auto max-w-[1200px] px-5 pt-6 sm:px-8">
      <Eyebrow>Orders & tracking</Eyebrow>
      <h1 className="mt-3 font-display text-[clamp(3rem,8vw,7rem)] font-extrabold leading-[0.88] tracking-[-0.055em]">
        <MaskText text="Your" /> <MaskText text="orders." delay={0.1} className="font-serif font-normal italic tracking-[-0.02em] text-ultra" />
      </h1>

      <div className="mt-10">
        {orders.isError ? (
          <p className="text-signal">{errorMessage(orders.error)}</p>
        ) : orders.isLoading ? (
          <div className="space-y-3">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skel key={i} className="h-24" />
            ))}
          </div>
        ) : orders.data?.length === 0 && page === 1 ? (
          <div className="flex flex-col items-start gap-5 rounded-[2.5rem] bg-bone-2 p-10">
            <p className="font-display text-4xl font-bold tracking-tight">
              Nothing ordered <span className="font-serif font-normal italic">yet.</span>
            </p>
            <p className="text-smoke">Orders you place appear here, tracked from payment to delivery.</p>
            <Link to="/shop/search" className={pill("ink")}>
              Browse the index <ArrowRight className="size-4" />
            </Link>
          </div>
        ) : (
          <>
            <ul className={cn("border-t border-ink/10", orders.isPlaceholderData && "opacity-50")}>
              {orders.data?.map((order, index) => (
                <motion.li
                  key={order.id}
                  initial={{ opacity: 0, y: 12 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.4, ease, delay: index * 0.03 }}
                >
                  <Link
                    to={`/shop/orders/${order.id}`}
                    data-testid="order-row"
                    className="group grid grid-cols-[1fr_auto] items-center gap-x-6 gap-y-2 border-b border-ink/10 py-6 transition-colors hover:bg-bone-2/60 sm:grid-cols-[140px_1fr_auto_auto] sm:px-4"
                  >
                    <span className="font-mono text-sm">{shortId(order.id)}</span>
                    <span className="order-3 col-span-2 flex flex-wrap items-center gap-3 sm:order-none sm:col-span-1">
                      <span className="text-sm text-smoke">
                        {date(order.created_at)} · {order.total_items} {order.total_items === 1 ? "item" : "items"}
                      </span>
                      <MiniTrack status={order.status} />
                      {isAwaitingPayment(order.status) && <span className="font-mono text-[11px] uppercase tracking-wider text-signal">Payment due</span>}
                    </span>
                    <span className="hidden sm:block">
                      <StatusChip status={order.status} />
                    </span>
                    <span className="flex items-center gap-3">
                      <span className="tabular font-display text-xl font-bold">{money(order.total_amount)}</span>
                      <ArrowUpRight className="size-5 transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
                    </span>
                    <span className="sm:hidden">
                      <StatusChip status={order.status} />
                    </span>
                  </Link>
                </motion.li>
              ))}
            </ul>
            {(page > 1 || (orders.data?.length ?? 0) === ORDER_PAGE_SIZE) && (
              <div className="mt-8 flex items-center gap-3">
                <button className={pill("line", "sm")} disabled={page <= 1} onClick={() => setPage(page - 1)}>
                  Newer
                </button>
                <span className="font-mono text-xs text-smoke">Page {page}</span>
                <button className={pill("line", "sm")} disabled={(orders.data?.length ?? 0) < ORDER_PAGE_SIZE} onClick={() => setPage(page + 1)}>
                  Older
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
