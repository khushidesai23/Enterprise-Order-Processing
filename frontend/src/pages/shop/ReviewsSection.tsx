import { useState, type FormEvent } from "react";
import { Link, useLocation } from "react-router-dom";
import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import * as Dialog from "@radix-ui/react-dialog";
import { AnimatePresence, motion } from "motion/react";
import { toast } from "sonner";
import { BadgeCheck, Star, X } from "lucide-react";
import { reviewsApi } from "@/lib/endpoints";
import { ApiError, errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { date, number } from "@/lib/format";
import type { Product } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Eyebrow, Skel, StarRow, ease, pill } from "@/store/ui";

const PAGE_SIZE = 6;

export function ReviewsSection({ product }: { product: Product }) {
  const { user } = useAuth();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [rating, setRating] = useState(0);
  const [sort, setSort] = useState("recent");
  const [page, setPage] = useState(1);
  const [writing, setWriting] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);

  const reviews = useQuery({
    queryKey: ["reviews", product.id, rating, sort, page],
    queryFn: () => reviewsApi.forProduct(product.id, { rating: rating || undefined, sort, page, limit: PAGE_SIZE }),
    placeholderData: keepPreviousData,
  });

  // 404 just means the user has not reviewed this product.
  const mine = useQuery({
    queryKey: ["reviews", product.id, "mine", user?.id],
    queryFn: async () => {
      try {
        return await reviewsApi.mine(product.id);
      } catch (error) {
        if (error instanceof ApiError && error.status === 404) return null;
        throw error;
      }
    },
    enabled: !!user,
  });

  const refresh = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: ["reviews", product.id] }),
      queryClient.invalidateQueries({ queryKey: ["product", product.id] }),
    ]);

  const remove = useMutation({
    mutationFn: () => reviewsApi.remove(mine.data!.id),
    onSuccess: async () => {
      toast("Your review was deleted");
      setConfirmDelete(false);
      await refresh();
    },
    onError: (error) => toast.error("Could not delete your review", { description: errorMessage(error) }),
  });

  const summary = reviews.data?.summary;
  const pages = reviews.data ? Math.max(1, Math.ceil(reviews.data.total / PAGE_SIZE)) : 1;

  return (
    <section id="reviews" className="scroll-mt-24 border-t border-ink/10 py-16">
      <div className="grid gap-12 lg:grid-cols-[360px_1fr]">
        <div className="lg:sticky lg:top-28 lg:self-start">
          <Eyebrow>Reviews</Eyebrow>
          {summary ? (
            <>
              <p className="mt-3 font-display text-8xl font-extrabold leading-none tracking-[-0.06em]">
                {summary.count ? summary.average.toFixed(1) : "—"}
              </p>
              <StarRow value={summary.average} size={20} className="mt-3" />
              <p className="mt-2 font-mono text-xs text-smoke">{number(summary.count)} global ratings</p>
              <ul className="mt-6 space-y-1.5">
                {[5, 4, 3, 2, 1].map((stars) => {
                  const count = summary.histogram[String(stars)] ?? 0;
                  const percent = summary.count ? Math.round((count / summary.count) * 100) : 0;
                  return (
                    <li key={stars}>
                      <button
                        onClick={() => {
                          setRating(rating === stars ? 0 : stars);
                          setPage(1);
                        }}
                        aria-pressed={rating === stars}
                        aria-label={`${stars} star reviews: ${percent}%`}
                        className={cn("group flex w-full items-center gap-3 rounded-lg px-1 py-0.5 text-sm", rating === stars && "bg-ink text-bone")}
                      >
                        <span className="w-6 font-mono text-xs">{stars}★</span>
                        <span className="h-2 flex-1 overflow-hidden rounded-full bg-ink/10">
                          <motion.span
                            className={cn("block h-full rounded-full", rating === stars ? "bg-acid" : "bg-ink group-hover:bg-ultra")}
                            initial={{ width: 0 }}
                            whileInView={{ width: `${percent}%` }}
                            viewport={{ once: true }}
                            transition={{ duration: 0.9, ease, delay: (5 - stars) * 0.06 }}
                          />
                        </span>
                        <span className="w-9 text-right font-mono text-xs">{percent}%</span>
                      </button>
                    </li>
                  );
                })}
              </ul>
            </>
          ) : (
            <Skel className="mt-4 h-60" />
          )}

          <div className="mt-8 rounded-3xl bg-bone-2 p-5">
            <p className="font-display text-xl font-bold tracking-tight">Own this?</p>
            <p className="mt-1 text-sm text-smoke">Reviews from people who bought it are marked verified.</p>
            {user ? (
              mine.data ? (
                <div className="mt-4">
                  <StarRow value={mine.data.rating} />
                  {mine.data.title && <p className="mt-1 text-sm font-medium">{mine.data.title}</p>}
                  <div className="mt-3 flex gap-2">
                    <button className={pill("ink", "sm")} onClick={() => setWriting(true)}>
                      Edit review
                    </button>
                    <button className={pill("ghost", "sm")} onClick={() => setConfirmDelete(true)}>
                      Delete
                    </button>
                  </div>
                </div>
              ) : (
                <button className={cn(pill("ink"), "mt-4 w-full")} onClick={() => setWriting(true)}>
                  Write a product review
                </button>
              )
            ) : (
              <Link to="/login" state={{ from: location.pathname }} className={cn(pill("ink"), "mt-4 w-full")}>
                Sign in to write a review
              </Link>
            )}
          </div>
        </div>

        <div>
          <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
            <h2 className="font-display text-3xl font-extrabold tracking-[-0.04em]">
              {rating ? `${rating}-star reviews` : "What people say"}{" "}
              {reviews.data && <span className="font-mono text-sm font-normal text-smoke">({number(reviews.data.total)})</span>}
            </h2>
            <select
              value={sort}
              onChange={(event) => {
                setSort(event.target.value);
                setPage(1);
              }}
              className="h-9 cursor-pointer rounded-full border border-ink/15 bg-bone px-3 text-sm focus:border-ink focus:outline-none"
              aria-label="Sort reviews"
            >
              <option value="recent">Most recent</option>
              <option value="highest">Highest rated</option>
              <option value="lowest">Lowest rated</option>
            </select>
          </div>

          {reviews.isLoading ? (
            <div className="space-y-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skel key={i} className="h-32" />
              ))}
            </div>
          ) : reviews.data?.items.length === 0 ? (
            <p className="rounded-3xl border border-dashed border-ink/20 p-10 text-center text-smoke">
              {rating ? `No ${rating}-star reviews yet.` : "No reviews yet. Be the first."}
            </p>
          ) : (
            <ul className={cn("grid gap-4 md:grid-cols-2", reviews.isPlaceholderData && "opacity-50")} data-testid="review-list">
              {reviews.data?.items.map((review, index) => (
                <motion.li
                  key={review.id}
                  data-testid="review"
                  initial={{ opacity: 0, y: 16 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.4, ease, delay: index * 0.04 }}
                  className="flex flex-col rounded-3xl bg-bone-2/70 p-6"
                >
                  <div className="flex items-center justify-between gap-3">
                    <StarRow value={review.rating} size={14} />
                    <span className="font-mono text-[11px] text-smoke">{date(review.created_at)}</span>
                  </div>
                  {review.title && <p className="mt-3 font-display text-lg font-bold leading-tight tracking-tight">{review.title}</p>}
                  {review.body && <p className="mt-2 flex-1 text-[15px] leading-relaxed">{review.body}</p>}
                  <p className="mt-4 flex flex-wrap items-center gap-2 text-sm">
                    <span className="flex size-7 items-center justify-center rounded-full bg-ink font-display text-[11px] font-bold text-acid">
                      {review.reviewer_name.slice(0, 1)}
                    </span>
                    {review.reviewer_name}
                    {review.verified_purchase && (
                      <span className="inline-flex items-center gap-1 rounded-full bg-acid px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider">
                        <BadgeCheck className="size-3" /> Verified Purchase
                      </span>
                    )}
                  </p>
                </motion.li>
              ))}
            </ul>
          )}

          {pages > 1 && (
            <div className="mt-8 flex items-center gap-3">
              <button className={pill("line", "sm")} disabled={page <= 1} onClick={() => setPage(page - 1)}>
                Previous
              </button>
              <span className="font-mono text-xs text-smoke">
                {page} / {pages}
              </span>
              <button className={pill("line", "sm")} disabled={page >= pages} onClick={() => setPage(page + 1)}>
                Next
              </button>
            </div>
          )}
        </div>
      </div>

      <ReviewDialog
        open={writing}
        onOpenChange={setWriting}
        productId={product.id}
        productName={product.name}
        existing={mine.data ?? null}
        onSaved={refresh}
      />

      <Dialog.Root open={confirmDelete} onOpenChange={setConfirmDelete}>
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-50 bg-ink/50" />
          <Dialog.Content className="store fixed left-1/2 top-1/2 z-50 w-[min(420px,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 rounded-3xl bg-bone p-7 focus:outline-none">
            <Dialog.Title className="font-display text-2xl font-bold tracking-tight">Delete your review?</Dialog.Title>
            <Dialog.Description className="mt-2 text-smoke">Your rating will no longer count towards this object.</Dialog.Description>
            <div className="mt-6 flex justify-end gap-2">
              <Dialog.Close className={pill("ghost", "sm")}>Keep it</Dialog.Close>
              <button className={pill("ink", "sm")} disabled={remove.isPending} onClick={() => remove.mutate()}>
                Delete review
              </button>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </section>
  );
}

function ReviewDialog({
  open,
  onOpenChange,
  productId,
  productName,
  existing,
  onSaved,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  productId: string;
  productName: string;
  existing: { rating: number; title: string; body: string } | null;
  onSaved: () => Promise<unknown>;
}) {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open && (
          <Dialog.Portal forceMount>
            <Dialog.Overlay asChild forceMount>
              <motion.div className="fixed inset-0 z-50 bg-ink/50 backdrop-blur-sm" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} />
            </Dialog.Overlay>
            <Dialog.Content asChild forceMount aria-describedby={undefined}>
              <motion.div
                className="store fixed left-1/2 top-1/2 z-50 w-[min(560px,calc(100vw-1.5rem))] -translate-x-1/2 -translate-y-1/2 rounded-[2rem] bg-bone p-7 focus:outline-none"
                initial={{ opacity: 0, scale: 0.96, y: 20 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.96 }}
                transition={{ duration: 0.35, ease }}
              >
                <ReviewForm productId={productId} productName={productName} existing={existing} onDone={() => onOpenChange(false)} onSaved={onSaved} />
              </motion.div>
            </Dialog.Content>
          </Dialog.Portal>
        )}
      </AnimatePresence>
    </Dialog.Root>
  );
}

function ReviewForm({
  productId,
  productName,
  existing,
  onDone,
  onSaved,
}: {
  productId: string;
  productName: string;
  existing: { rating: number; title: string; body: string } | null;
  onDone: () => void;
  onSaved: () => Promise<unknown>;
}) {
  const [rating, setRating] = useState(existing?.rating ?? 0);
  const [hover, setHover] = useState(0);
  const [title, setTitle] = useState(existing?.title ?? "");
  const [body, setBody] = useState(existing?.body ?? "");
  const [error, setError] = useState<string | null>(null);

  const save = useMutation({
    mutationFn: () => reviewsApi.save(productId, { rating, title: title.trim(), body: body.trim() }),
    onSuccess: async (review) => {
      toast(review.verified_purchase ? "Thanks! Your verified review is live." : "Thanks! Your review is live.");
      await onSaved();
      onDone();
    },
    onError: (err) => setError(errorMessage(err)),
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    if (rating < 1) {
      setError("Choose a star rating.");
      return;
    }
    setError(null);
    save.mutate();
  }

  const field = "w-full rounded-2xl border border-ink/15 bg-transparent px-4 py-3 text-[15px] focus:border-ink focus:outline-none";

  return (
    <form onSubmit={submit} noValidate>
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <Dialog.Title className="font-display text-3xl font-extrabold tracking-tight">{existing ? "Edit your review" : "Your review"}</Dialog.Title>
          <p className="mt-1 truncate font-mono text-[11px] uppercase tracking-wider text-smoke">{productName}</p>
        </div>
        <Dialog.Close className="flex size-9 shrink-0 items-center justify-center rounded-full hover:bg-ink/5" aria-label="Close">
          <X className="size-5" />
        </Dialog.Close>
      </div>

      <div className="mt-6 flex gap-1" role="radiogroup" aria-label="Overall rating" onMouseLeave={() => setHover(0)}>
        {[1, 2, 3, 4, 5].map((stars) => (
          <motion.button
            key={stars}
            type="button"
            role="radio"
            aria-checked={rating === stars}
            aria-label={`${stars} star${stars > 1 ? "s" : ""}`}
            whileTap={{ scale: 0.8 }}
            onMouseEnter={() => setHover(stars)}
            onClick={() => setRating(stars)}
          >
            <Star className={cn("size-10 transition-colors", (hover || rating) >= stars ? "fill-ink text-ink" : "text-ink/25")} strokeWidth={1.5} />
          </motion.button>
        ))}
      </div>

      <div className="mt-6 space-y-4">
        <label className="block">
          <span className="mb-1.5 block font-mono text-[11px] uppercase tracking-wider text-smoke">Add a headline</span>
          <input value={title} onChange={(event) => setTitle(event.target.value)} maxLength={150} placeholder="The one thing to know" className={field} />
        </label>
        <label className="block">
          <span className="mb-1.5 block font-mono text-[11px] uppercase tracking-wider text-smoke">Add a written review</span>
          <textarea value={body} onChange={(event) => setBody(event.target.value)} maxLength={4000} rows={4} placeholder="What did you like or dislike?" className={field} />
        </label>
      </div>
      {error && (
        <p role="alert" className="mt-3 text-sm text-signal">
          {error}
        </p>
      )}
      <div className="mt-6 flex justify-end gap-2">
        <Dialog.Close className={pill("ghost")} type="button">
          Cancel
        </Dialog.Close>
        <button type="submit" disabled={save.isPending} className={pill("ink")}>
          {save.isPending ? "Saving…" : "Submit review"}
        </button>
      </div>
    </form>
  );
}
