import { useState } from "react";
import { Link } from "react-router-dom";
import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { BadgeCheck, MessageSquareText, Trash2 } from "lucide-react";
import { reviewsApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { dateTime, number } from "@/lib/format";
import type { AdminReview } from "@/lib/types";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, EmptyState, ErrorState, Skeleton } from "@/components/ui/primitives";
import { ConfirmDialog } from "@/components/ui/dialog";
import { PageHeader, Pagination } from "@/components/common";
import { Stars } from "@/components/shop";

const PAGE_SIZE = 20;

export function ReviewsAdminPage() {
  const queryClient = useQueryClient();
  const [rating, setRating] = useState(0);
  const [page, setPage] = useState(1);
  const [deleting, setDeleting] = useState<AdminReview | null>(null);

  const reviews = useQuery({
    queryKey: ["admin-reviews", rating, page],
    queryFn: () => reviewsApi.list({ rating: rating || undefined, page, limit: PAGE_SIZE }),
    placeholderData: keepPreviousData,
  });

  const remove = useMutation({
    mutationFn: (id: string) => reviewsApi.remove(id),
    onSuccess: async () => {
      toast.success("Review deleted", { description: "The product's rating was recalculated." });
      setDeleting(null);
      await queryClient.invalidateQueries();
    },
    onError: (error) => toast.error("Could not delete review", { description: errorMessage(error) }),
  });

  return (
    <div className="animate-in">
      <PageHeader title="Reviews" description="Moderate customer reviews. Deleting a review updates the product's rating." />

      <div className="-mx-1 mb-4 flex gap-1 overflow-x-auto px-1" role="tablist" aria-label="Filter by rating">
        {[0, 5, 4, 3, 2, 1].map((stars) => (
          <button
            key={stars}
            role="tab"
            aria-selected={rating === stars}
            onClick={() => {
              setRating(stars);
              setPage(1);
            }}
            className={cn(
              "shrink-0 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors",
              rating === stars ? "bg-foreground text-white" : "text-subtle hover:bg-muted hover:text-foreground",
            )}
          >
            {stars ? `${stars} star` : "All"}
          </button>
        ))}
      </div>

      <Card>
        {reviews.isError ? (
          <div className="p-5">
            <ErrorState message={errorMessage(reviews.error)} onRetry={() => reviews.refetch()} />
          </div>
        ) : reviews.isLoading ? (
          <div className="p-5">
            <Skeleton className="h-80" />
          </div>
        ) : reviews.data?.items.length === 0 ? (
          <EmptyState icon={<MessageSquareText />} title="No reviews here" />
        ) : (
          <>
            <ul className={cn("divide-y divide-border", reviews.isPlaceholderData && "opacity-60")}>
              {reviews.data?.items.map((review) => (
                <li key={review.id} className="flex gap-4 px-5 py-4" data-testid="admin-review-row">
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <Stars value={review.rating} />
                      {review.title && <span className="text-sm font-semibold">{review.title}</span>}
                      {review.verified_purchase && (
                        <span className="flex items-center gap-1 text-xs font-medium text-success">
                          <BadgeCheck className="size-3.5" /> Verified
                        </span>
                      )}
                    </div>
                    {review.body && <p className="mt-1 line-clamp-2 text-sm text-subtle">{review.body}</p>}
                    <p className="mt-1.5 text-xs text-faint">
                      {review.reviewer_name} ({review.reviewer_email || "deleted account"}) on{" "}
                      <Link to={`/shop/products/${review.product_id}`} target="_blank" className="text-link hover:underline">
                        {review.product_name || "a deleted product"}
                      </Link>{" "}
                      · {dateTime(review.created_at)}
                    </p>
                  </div>
                  <Button variant="ghost" size="icon" onClick={() => setDeleting(review)} aria-label={`Delete review by ${review.reviewer_name}`}>
                    <Trash2 className="text-danger" />
                  </Button>
                </li>
              ))}
            </ul>
            <div className="flex items-center justify-between border-t border-border">
              <span className="pl-4 text-sm text-subtle">{number(reviews.data?.total ?? 0)} reviews</span>
              <Pagination
                page={page}
                hasNext={page * PAGE_SIZE < (reviews.data?.total ?? 0)}
                onChange={setPage}
                loading={reviews.isFetching}
              />
            </div>
          </>
        )}
      </Card>

      <ConfirmDialog
        open={!!deleting}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Delete this review?"
        description={`The ${deleting?.rating}-star review by ${deleting?.reviewer_name} is removed and the product's rating recalculated.`}
        confirmLabel="Delete review"
        destructive
        loading={remove.isPending}
        onConfirm={() => deleting && remove.mutate(deleting.id)}
      />
    </div>
  );
}
