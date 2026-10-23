import { useEffect, useState, type FormEvent } from "react";
import { Link, useLocation } from "react-router-dom";
import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { BadgeCheck, Star, UserRound } from "lucide-react";
import { reviewsApi } from "@/lib/endpoints";
import { ApiError, errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { date, number } from "@/lib/format";
import type { Product } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/dialog";
import { Field, Input, Skeleton, Textarea } from "@/components/ui/primitives";
import { Stars } from "@/components/shop";

const PAGE_SIZE = 8;

export function ReviewsSection({ product }: { product: Product }) {
  const { user } = useAuth();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [rating, setRating] = useState(0);
  const [sort, setSort] = useState("recent");
  const [page, setPage] = useState(1);
  const [writing, setWriting] = useState(false);
  const [deleting, setDeleting] = useState(false);

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
      toast.success("Your review was deleted");
      setDeleting(false);
      await refresh();
    },
    onError: (error) => toast.error("Could not delete your review", { description: errorMessage(error) }),
  });

  const summary = reviews.data?.summary;
  const pages = reviews.data ? Math.max(1, Math.ceil(reviews.data.total / PAGE_SIZE)) : 1;

  return (
    <section id="reviews" className="grid scroll-mt-32 gap-10 border-t border-border pt-8 lg:grid-cols-[320px_1fr]">
      <div>
        <h2 className="text-2xl font-bold">Customer reviews</h2>
        {summary ? (
          <>
            <div className="mt-2 flex items-center gap-2">
              <Stars value={summary.average} size={20} />
              <span className="text-lg font-medium">{summary.count ? `${summary.average.toFixed(1)} out of 5` : "No ratings yet"}</span>
            </div>
            <p className="mt-1 text-sm text-subtle">{number(summary.count)} global ratings</p>
            <ul className="mt-4 space-y-2">
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
                      className={cn("flex w-full items-center gap-3 text-sm text-link hover:text-deal", rating === stars && "font-bold")}
                      aria-label={`${stars} star reviews: ${percent}%`}
                    >
                      <span className="w-12 shrink-0 text-left">{stars} star</span>
                      <span className="h-5 flex-1 overflow-hidden rounded border border-border bg-muted">
                        <span className="block h-full bg-star" style={{ width: `${percent}%` }} />
                      </span>
                      <span className="w-10 shrink-0 text-right">{percent}%</span>
                    </button>
                  </li>
                );
              })}
            </ul>
          </>
        ) : (
          <Skeleton className="mt-4 h-40" />
        )}

        <hr className="my-6" />
        <h3 className="text-lg font-bold">Review this product</h3>
        <p className="mt-1 text-sm text-subtle">Share your thoughts with other customers.</p>
        {user ? (
          mine.data ? (
            <div className="mt-3 rounded-lg border border-border p-3 text-sm">
              <p className="font-medium">Your review</p>
              <Stars value={mine.data.rating} className="mt-1" />
              {mine.data.title && <p className="mt-1 font-semibold">{mine.data.title}</p>}
              <div className="mt-3 flex gap-2">
                <Button size="sm" variant="secondary" onClick={() => setWriting(true)}>
                  Edit review
                </Button>
                <Button size="sm" variant="ghost" onClick={() => setDeleting(true)}>
                  Delete
                </Button>
              </div>
            </div>
          ) : (
            <Button variant="secondary" className="mt-3 w-full rounded-full" onClick={() => setWriting(true)}>
              Write a product review
            </Button>
          )
        ) : (
          <Link to="/login" state={{ from: location.pathname }} className="mt-3 block">
            <Button variant="secondary" className="w-full rounded-full">
              Sign in to write a review
            </Button>
          </Link>
        )}
      </div>

      <div>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h3 className="text-lg font-bold">
            {rating ? `${rating}-star reviews` : "Top reviews"}{" "}
            {reviews.data && <span className="text-sm font-normal text-subtle">({number(reviews.data.total)})</span>}
          </h3>
          <select
            value={sort}
            onChange={(event) => {
              setSort(event.target.value);
              setPage(1);
            }}
            className="cursor-pointer rounded-lg border border-border bg-muted px-2 py-1 text-sm shadow-sm"
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
              <Skeleton key={i} className="h-28" />
            ))}
          </div>
        ) : reviews.data?.items.length === 0 ? (
          <p className="text-sm text-subtle">{rating ? `No ${rating}-star reviews yet.` : "No reviews yet. Be the first to review this product."}</p>
        ) : (
          <ul className={cn("space-y-6", reviews.isPlaceholderData && "opacity-60")} data-testid="review-list">
            {reviews.data?.items.map((review) => (
              <li key={review.id} data-testid="review">
                <div className="flex items-center gap-2 text-sm">
                  <span className="flex size-8 items-center justify-center rounded-full bg-muted text-subtle">
                    <UserRound className="size-4" />
                  </span>
                  {review.reviewer_name}
                </div>
                <div className="mt-1.5 flex flex-wrap items-center gap-2">
                  <Stars value={review.rating} />
                  {review.title && <span className="text-sm font-bold">{review.title}</span>}
                </div>
                <p className="mt-1 text-xs text-subtle">Reviewed in India on {date(review.created_at)}</p>
                {review.verified_purchase && (
                  <p className="mt-0.5 flex items-center gap-1 text-xs font-semibold text-buy-hover">
                    <BadgeCheck className="size-3.5" /> Verified Purchase
                  </p>
                )}
                {review.body && <p className="mt-2 whitespace-pre-line text-sm leading-relaxed">{review.body}</p>}
              </li>
            ))}
          </ul>
        )}

        {pages > 1 && (
          <div className="mt-6 flex items-center gap-2 text-sm">
            <Button size="sm" variant="secondary" disabled={page <= 1} onClick={() => setPage(page - 1)}>
              ← Previous
            </Button>
            <span className="text-subtle">
              Page {page} of {pages}
            </span>
            <Button size="sm" variant="secondary" disabled={page >= pages} onClick={() => setPage(page + 1)}>
              Next →
            </Button>
          </div>
        )}
      </div>

      {writing && (
        <ReviewDialog
          productId={product.id}
          productName={product.name}
          existing={mine.data ?? null}
          onClose={() => setWriting(false)}
          onSaved={refresh}
        />
      )}

      <ConfirmDialog
        open={deleting}
        onOpenChange={setDeleting}
        title="Delete your review?"
        description="Your rating will no longer count towards this product."
        confirmLabel="Delete review"
        destructive
        loading={remove.isPending}
        onConfirm={() => remove.mutate()}
      />
    </section>
  );
}

function ReviewDialog({
  productId,
  productName,
  existing,
  onClose,
  onSaved,
}: {
  productId: string;
  productName: string;
  existing: { rating: number; title: string; body: string } | null;
  onClose: () => void;
  onSaved: () => Promise<unknown>;
}) {
  const [rating, setRating] = useState(existing?.rating ?? 0);
  const [hover, setHover] = useState(0);
  const [title, setTitle] = useState(existing?.title ?? "");
  const [body, setBody] = useState(existing?.body ?? "");

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => event.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const save = useMutation({
    mutationFn: () => reviewsApi.save(productId, { rating, title: title.trim(), body: body.trim() }),
    onSuccess: async (review) => {
      toast.success(review.verified_purchase ? "Thanks! Your verified review is live." : "Thanks! Your review is live.");
      await onSaved();
      onClose();
    },
    onError: (error) => toast.error("Could not save your review", { description: errorMessage(error) }),
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    if (rating < 1) {
      toast.error("Choose a star rating");
      return;
    }
    save.mutate();
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 p-4" onClick={onClose}>
      <form
        onSubmit={submit}
        onClick={(event) => event.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Write a review"
        className="animate-in w-full max-w-lg rounded-xl bg-surface p-6 shadow-pop"
      >
        <h2 className="text-lg font-semibold">{existing ? "Edit your review" : "Create review"}</h2>
        <p className="mt-0.5 line-clamp-1 text-sm text-subtle">{productName}</p>

        <p className="mt-5 text-sm font-medium">Overall rating</p>
        <div className="mt-1 flex gap-1" role="radiogroup" aria-label="Overall rating" onMouseLeave={() => setHover(0)}>
          {[1, 2, 3, 4, 5].map((stars) => (
            <button
              key={stars}
              type="button"
              role="radio"
              aria-checked={rating === stars}
              aria-label={`${stars} star${stars > 1 ? "s" : ""}`}
              onMouseEnter={() => setHover(stars)}
              onClick={() => setRating(stars)}
            >
              <Star className={cn("size-8", (hover || rating) >= stars ? "fill-star text-star" : "text-faint")} strokeWidth={1.5} />
            </button>
          ))}
        </div>

        <div className="mt-4 space-y-4">
          <Field label="Add a headline" htmlFor="review-title">
            <Input id="review-title" value={title} onChange={(event) => setTitle(event.target.value)} maxLength={150} placeholder="What's most important to know?" />
          </Field>
          <Field label="Add a written review" htmlFor="review-body" hint={`${body.length}/4000`}>
            <Textarea id="review-body" value={body} onChange={(event) => setBody(event.target.value)} maxLength={4000} placeholder="What did you like or dislike? What did you use this product for?" />
          </Field>
        </div>

        <div className="mt-6 flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose} disabled={save.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={save.isPending}>
            Submit review
          </Button>
        </div>
      </form>
    </div>
  );
}
