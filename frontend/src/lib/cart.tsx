import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";

export interface CartLine {
  productId: string;
  name: string;
  sku: string;
  price: number;
  category: string;
  imageUrl?: string;
  quantity: number;
}

interface CartState {
  lines: CartLine[];
  count: number;
  subtotal: number;
  add: (line: Omit<CartLine, "quantity">, quantity?: number) => void;
  setQuantity: (productId: string, quantity: number) => void;
  remove: (productId: string) => void;
  clear: () => void;
}

const CART_KEY = "meridian.cart";

export const MAX_LINE_QUANTITY = 99;

function load(): CartLine[] {
  try {
    const raw = localStorage.getItem(CART_KEY);
    const parsed = raw ? (JSON.parse(raw) as CartLine[]) : [];
    return Array.isArray(parsed) ? parsed.filter((line) => line.productId && line.quantity > 0) : [];
  } catch {
    return [];
  }
}

// Pure cart operations, exported for tests.
export function addLine(lines: CartLine[], line: Omit<CartLine, "quantity">, quantity: number): CartLine[] {
  const existing = lines.find((item) => item.productId === line.productId);
  if (existing) {
    return lines.map((item) =>
      item.productId === line.productId
        ? { ...item, ...line, quantity: Math.min(MAX_LINE_QUANTITY, item.quantity + quantity) }
        : item,
    );
  }
  return [...lines, { ...line, quantity: Math.min(MAX_LINE_QUANTITY, quantity) }];
}

export function setLineQuantity(lines: CartLine[], productId: string, quantity: number): CartLine[] {
  if (quantity <= 0) return lines.filter((item) => item.productId !== productId);
  return lines.map((item) =>
    item.productId === productId ? { ...item, quantity: Math.min(MAX_LINE_QUANTITY, quantity) } : item,
  );
}

const CartContext = createContext<CartState | null>(null);

export function CartProvider({ children }: { children: ReactNode }) {
  const [lines, setLines] = useState<CartLine[]>(load);

  useEffect(() => {
    try {
      localStorage.setItem(CART_KEY, JSON.stringify(lines));
    } catch {
      // Storage unavailable: the cart lasts until the page reloads.
    }
  }, [lines]);

  const add = useCallback((line: Omit<CartLine, "quantity">, quantity = 1) => {
    setLines((current) => addLine(current, line, quantity));
  }, []);

  const setQuantity = useCallback((productId: string, quantity: number) => {
    setLines((current) => setLineQuantity(current, productId, quantity));
  }, []);

  const remove = useCallback((productId: string) => {
    setLines((current) => current.filter((item) => item.productId !== productId));
  }, []);

  const clear = useCallback(() => setLines([]), []);

  const value = useMemo<CartState>(
    () => ({
      lines,
      count: lines.reduce((sum, line) => sum + line.quantity, 0),
      subtotal: lines.reduce((sum, line) => sum + line.price * line.quantity, 0),
      add,
      setQuantity,
      remove,
      clear,
    }),
    [lines, add, setQuantity, remove, clear],
  );

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart(): CartState {
  const context = useContext(CartContext);
  if (!context) throw new Error("useCart must be used inside CartProvider");
  return context;
}
