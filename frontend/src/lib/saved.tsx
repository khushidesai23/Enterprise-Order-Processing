import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import type { ProductListItem } from "./types";

// A product as remembered in the browser: enough to render a card without
// a request. Prices and stock are refreshed from the API where it matters.
export type SavedProduct = Pick<
  ProductListItem,
  "id" | "name" | "brand" | "price" | "mrp" | "image_url" | "rating_average" | "rating_count" | "category"
>;

export function toSaved(product: SavedProduct): SavedProduct {
  const { id, name, brand, price, mrp, image_url, rating_average, rating_count, category } = product;
  return { id, name, brand, price, mrp, image_url, rating_average, rating_count, category };
}

function read(key: string): SavedProduct[] {
  try {
    const parsed = JSON.parse(localStorage.getItem(key) ?? "[]");
    return Array.isArray(parsed) ? parsed.filter((item) => item && typeof item.id === "string") : [];
  } catch {
    return [];
  }
}

function write(key: string, items: SavedProduct[]) {
  try {
    localStorage.setItem(key, JSON.stringify(items));
  } catch {
    // Storage unavailable: the list lasts until the page reloads.
  }
}

// --- Wishlist ---------------------------------------------------------------

const WISHLIST_KEY = "meridian.wishlist";

interface WishlistState {
  items: SavedProduct[];
  has: (id: string) => boolean;
  toggle: (product: SavedProduct) => boolean;
  remove: (id: string) => void;
}

const WishlistContext = createContext<WishlistState | null>(null);

export function WishlistProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<SavedProduct[]>(() => read(WISHLIST_KEY));

  useEffect(() => write(WISHLIST_KEY, items), [items]);

  const has = useCallback((id: string) => items.some((item) => item.id === id), [items]);

  // Returns true when the product was added, false when removed.
  const toggle = useCallback(
    (product: SavedProduct) => {
      const present = items.some((item) => item.id === product.id);
      setItems((current) =>
        present ? current.filter((item) => item.id !== product.id) : [toSaved(product), ...current].slice(0, 200),
      );
      return !present;
    },
    [items],
  );

  const remove = useCallback((id: string) => setItems((current) => current.filter((item) => item.id !== id)), []);

  const value = useMemo(() => ({ items, has, toggle, remove }), [items, has, toggle, remove]);
  return <WishlistContext.Provider value={value}>{children}</WishlistContext.Provider>;
}

export function useWishlist(): WishlistState {
  const context = useContext(WishlistContext);
  if (!context) throw new Error("useWishlist must be used inside WishlistProvider");
  return context;
}

// --- Recently viewed --------------------------------------------------------

const RECENT_KEY = "meridian.recent";

export function rememberViewed(product: SavedProduct) {
  const items = read(RECENT_KEY).filter((item) => item.id !== product.id);
  write(RECENT_KEY, [toSaved(product), ...items].slice(0, 20));
}

export function recentlyViewed(): SavedProduct[] {
  return read(RECENT_KEY);
}
