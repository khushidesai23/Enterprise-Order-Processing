import { Link, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { AnimatePresence, motion } from "motion/react";
import { toast } from "sonner";
import { ArrowRight, Lock } from "lucide-react";
import { productsApi } from "@/lib/endpoints";
import { useAuth } from "@/lib/auth";
import { MAX_LINE_QUANTITY, useCart } from "@/lib/cart";
import { money, number } from "@/lib/format";
import { useWishlist } from "@/lib/saved";
import { Stepper, useBagAvailability } from "@/store/bag";
import { Rail } from "@/store/Rail";
import { Eyebrow, MaskText, ProductShot, distinct, ease, pill } from "@/store/ui";

export function CartPage() {
  const { lines, count, setQuantity, remove } = useCart();
  const wishlist = useWishlist();
  const { user } = useAuth();
  const navigate = useNavigate();
  const { byId, issues, subtotal, loading } = useBagAvailability();
  const suggestions = useQuery({
    queryKey: ["search", "bag-suggestions"],
    queryFn: () => productsApi.search({ sort: "bestselling", in_stock: true, limit: 60 }),
    staleTime: 60_000,
  });
  const mrpTotal = lines.reduce((sum, line) => {
    const live = byId.get(line.productId);
    return sum + (live && live.mrp > live.price ? live.mrp : (live?.price ?? line.price)) * line.quantity;
  }, 0);
  const savings = Math.max(0, mrpTotal - subtotal);

  function saveForLater(productId: string) {
    const line = lines.find((item) => item.productId === productId);
    const product = byId.get(productId);
    if (!line) return;
    if (!wishlist.has(productId)) {
      wishlist.toggle({
        id: productId,
        name: line.name,
        brand: product?.brand ?? "",
        price: product?.price ?? line.price,
        mrp: product?.mrp ?? 0,
        image_url: product?.image_url ?? line.imageUrl ?? "",
        rating_average: product?.rating_average ?? 0,
        rating_count: product?.rating_count ?? 0,
        category: line.category,
      });
    }
    remove(productId);
    toast("Saved for later in your wishlist");
  }

  return (
    <div className="mx-auto max-w-[1600px] px-5 pt-6 sm:px-8">
      <Eyebrow>Your bag · {number(count)} {count === 1 ? "item" : "items"}</Eyebrow>
      <h1 className="mt-3 font-display text-[clamp(3rem,8vw,7rem)] font-extrabold leading-[0.88] tracking-[-0.055em]">
        <MaskText text="The" /> <MaskText text="bag." delay={0.1} className="font-serif font-normal italic tracking-[-0.02em] text-ultra" />
      </h1>

      {lines.length === 0 ? (
        <div className="mt-12 flex flex-col items-start gap-5 rounded-[2.5rem] bg-bone-2 p-10">
          <p className="font-display text-4xl font-bold tracking-tight">
            Empty, <span className="font-serif font-normal italic">for now.</span>
          </p>
          <p className="max-w-md text-smoke">Add objects from the index; prices and stock stay live until checkout.</p>
          <div className="flex flex-wrap gap-2">
            <Link to="/shop/search" className={pill("ink")}>
              Browse the index <ArrowRight className="size-4" />
            </Link>
            {!user && (
              <Link to="/login" className={pill("line")}>
                Sign in
              </Link>
            )}
          </div>
        </div>
      ) : (
        <div className="mt-10 grid items-start gap-10 lg:grid-cols-[1fr_400px]">
          <ul className="border-t border-ink/10">
            <AnimatePresence initial={false}>
              {lines.map((line) => {
                const live = byId.get(line.productId);
                const issue = issues.get(line.productId);
                const price = live?.price ?? line.price;
                const available = live?.inventory?.available_quantity;
                return (
                  <motion.li
                    key={line.productId}
                    layout
                    exit={{ opacity: 0, height: 0, transition: { duration: 0.3 } }}
                    className="flex gap-5 border-b border-ink/10 py-6"
                    data-testid="cart-line"
                  >
                    <Link to={`/shop/products/${line.productId}`} className="shrink-0">
                      <ProductShot id={line.productId} name={line.name} src={live?.image_url ?? line.imageUrl} size="sm" className="size-28 rounded-3xl bg-bone-2 p-3 sm:size-40" />
                    </Link>
                    <div className="flex min-w-0 flex-1 flex-col">
                      <div className="flex items-start justify-between gap-4">
                        <div className="min-w-0">
                          <p className="font-mono text-[10px] uppercase tracking-[0.16em] text-smoke">{live?.brand || line.category}</p>
                          <Link to={`/shop/products/${line.productId}`} className="mt-1 line-clamp-2 text-lg font-medium hover:underline">
                            {line.name}
                          </Link>
                        </div>
                        <p className="tabular shrink-0 font-display text-xl font-bold">{money(price * line.quantity)}</p>
                      </div>
                      <p className="mt-1 font-mono text-xs text-smoke">{money(price)} each</p>
                      {issue ? (
                        <p className="mt-2 font-mono text-xs uppercase tracking-wider text-signal">{issue}</p>
                      ) : (
                        <p className="mt-2 font-mono text-xs uppercase tracking-wider text-ink/70">In stock · reserved when you order</p>
                      )}
                      <div className="mt-auto flex flex-wrap items-center gap-4 pt-4">
                        <Stepper
                          size="sm"
                          value={line.quantity}
                          max={Math.max(1, Math.min(available ?? MAX_LINE_QUANTITY, MAX_LINE_QUANTITY))}
                          onChange={(value) => setQuantity(line.productId, value)}
                          label={`Quantity of ${line.name}`}
                        />
                        <button onClick={() => saveForLater(line.productId)} className="text-sm underline-offset-4 hover:underline">
                          Save for later
                        </button>
                        <button onClick={() => remove(line.productId)} className="text-sm text-smoke underline-offset-4 hover:text-signal hover:underline" aria-label={`Remove ${line.name}`}>
                          Remove
                        </button>
                      </div>
                    </div>
                  </motion.li>
                );
              })}
            </AnimatePresence>
          </ul>

          <motion.aside
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, ease }}
            className="rounded-[2rem] bg-ink p-7 text-bone lg:sticky lg:top-28"
            aria-label="Order summary"
          >
            <p className="font-mono text-[11px] uppercase tracking-[0.18em] text-bone/50">Summary</p>
            <dl className="mt-5 space-y-3 text-sm">
              <div className="flex justify-between">
                <dt className="text-bone/70">Items ({number(count)})</dt>
                <dd className="tabular font-mono">{money(subtotal)}</dd>
              </div>
              {savings > 0 && (
                <div className="flex justify-between text-acid">
                  <dt>You save vs. list price</dt>
                  <dd className="tabular font-mono">−{money(savings)}</dd>
                </div>
              )}
              <div className="flex justify-between">
                <dt className="text-bone/70">Delivery</dt>
                <dd className="font-mono">Free</dd>
              </div>
            </dl>
            <div className="mt-6 border-t border-bone/15 pt-5">
              <span className="font-mono text-[11px] uppercase tracking-[0.18em] text-bone/50">Total</span>
              <p className="tabular mt-1 break-words font-display text-[clamp(1.6rem,2.6vw,2.4rem)] font-extrabold leading-tight tracking-tight">{money(subtotal)}</p>
            </div>
            <button
              disabled={issues.size > 0 || loading}
              onClick={() => navigate(user ? "/shop/checkout" : "/login", { state: { from: "/shop/checkout" } })}
              className={`${pill("acid", "lg")} mt-6 w-full`}
            >
              {user ? "Checkout" : "Sign in to check out"} <ArrowRight className="size-5" />
            </button>
            {issues.size > 0 && <p className="mt-3 text-center font-mono text-xs text-signal">Fix the flagged items to continue.</p>}
            <p className="mt-4 flex items-center justify-center gap-1.5 font-mono text-[10px] uppercase tracking-wider text-bone/50">
              <Lock className="size-3" /> Payments verified by Razorpay
            </p>
          </motion.aside>
        </div>
      )}

      {wishlist.items.length > 0 && (
        <Rail
          eyebrow="Wishlist"
          title={
            <>
              Saved for <span className="font-serif font-normal italic text-ultra">later</span>
            </>
          }
          products={wishlist.items}
        />
      )}
      <Rail
        eyebrow="Most wanted"
        title={
          <>
            Others are <span className="font-serif font-normal italic text-ultra">ordering</span>
          </>
        }
        products={distinct(suggestions.data?.items, 14)}
        loading={suggestions.isLoading}
      />
    </div>
  );
}
