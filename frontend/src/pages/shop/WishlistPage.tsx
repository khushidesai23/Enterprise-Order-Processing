import { Link } from "react-router-dom";
import { toast } from "sonner";
import { Heart, Trash2 } from "lucide-react";
import { useCart } from "@/lib/cart";
import { useWishlist } from "@/lib/saved";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/primitives";
import { DealBadge, PriceTag, ProductImage, RatingLine } from "@/components/shop";

export function WishlistPage() {
  const { items, remove } = useWishlist();
  const { add } = useCart();

  return (
    <div className="mx-auto max-w-5xl px-3 py-6 sm:px-4">
      <div className="rounded-lg bg-surface p-5 shadow-card">
        <h1 className="border-b border-border pb-3 text-2xl font-semibold">Your Wishlist</h1>
        {items.length === 0 ? (
          <EmptyState
            icon={<Heart />}
            title="Your wishlist is empty"
            description="Tap the heart on any product to save it here."
            action={
              <Link to="/shop">
                <Button variant="secondary">Continue shopping</Button>
              </Link>
            }
          />
        ) : (
          <ul className="divide-y divide-border">
            {items.map((item) => (
              <li key={item.id} className="flex gap-4 py-4" data-testid="wishlist-item">
                <Link to={`/shop/products/${item.id}`} className="shrink-0">
                  <ProductImage id={item.id} name={item.name} src={item.image_url} size="sm" className="size-32" />
                </Link>
                <div className="min-w-0 flex-1">
                  <Link to={`/shop/products/${item.id}`} className="line-clamp-2 text-base hover:text-link">
                    {item.name}
                  </Link>
                  {item.brand && <p className="text-xs text-subtle">by {item.brand}</p>}
                  <div className="mt-1">
                    <RatingLine average={item.rating_average} count={item.rating_count} />
                  </div>
                  <DealBadge price={item.price} mrp={item.mrp} />
                  <PriceTag price={item.price} mrp={item.mrp} />
                </div>
                <div className="flex shrink-0 flex-col gap-2">
                  <button
                    onClick={() => {
                      add({ productId: item.id, name: item.name, sku: "", price: item.price, category: item.category, imageUrl: item.image_url });
                      remove(item.id);
                      toast.success("Moved to cart", { description: item.name });
                    }}
                    className="rounded-full bg-cart px-4 py-1.5 text-sm font-medium hover:bg-cart-hover"
                  >
                    Move to cart
                  </button>
                  <Button variant="ghost" size="sm" onClick={() => remove(item.id)} aria-label={`Remove ${item.name} from wishlist`}>
                    <Trash2 /> Remove
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
