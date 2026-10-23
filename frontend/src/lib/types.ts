// Shapes of the Go API's responses (internal/modules/*/dto.go).

export type Role = "customer" | "admin";

export interface ApiEnvelope<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
  timestamp: string;
}

export interface UserInfo {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  is_active: boolean;
  role: Role;
}

export interface User extends UserInfo {
  created_at: string;
  updated_at: string;
}

export interface LoginResponse {
  token: string;
  token_type: string;
  expires_in: string;
  user: UserInfo;
}

export interface Category {
  id: string;
  name: string;
  description: string;
}

export interface CategoryListItem extends Category {
  product_count: number;
}

export interface ProductListItem {
  id: string;
  name: string;
  sku: string;
  price: number;
  brand: string;
  mrp: number;
  image_url: string;
  rating_average: number;
  rating_count: number;
  category_id: string;
  category: string;
  available_quantity: number;
  reserved_quantity: number;
}

export interface ProductSpec {
  name: string;
  value: string;
}

export interface Product {
  id: string;
  name: string;
  description: string;
  sku: string;
  price: number;
  brand: string;
  mrp: number;
  image_url: string;
  images: string[];
  specs: ProductSpec[];
  rating_average: number;
  rating_count: number;
  category: { id: string; name: string };
  inventory: { available_quantity: number; reserved_quantity: number };
}

export interface ProductInput {
  name: string;
  description: string;
  sku: string;
  price: number;
  category_id: string;
  brand?: string;
  mrp?: number;
  image_url?: string;
  images?: string[];
}

export type SearchSort =
  | "relevance"
  | "featured"
  | "price_asc"
  | "price_desc"
  | "rating"
  | "reviews"
  | "newest"
  | "discount"
  | "bestselling";

export interface SearchParams {
  q?: string;
  category_id?: string;
  brand?: string[];
  min_price?: number;
  max_price?: number;
  min_rating?: number;
  in_stock?: boolean;
  deals?: boolean;
  sort?: SearchSort;
  page?: number;
  limit?: number;
}

export interface ProductSearchResult {
  items: ProductListItem[];
  total: number;
  page: number;
  limit: number;
  facets: {
    categories: { id: string; name: string; count: number }[];
    brands: { name: string; count: number }[];
    price_min: number;
    price_max: number;
  };
}

export interface Review {
  id: string;
  product_id: string;
  user_id: string;
  reviewer_name: string;
  rating: number;
  title: string;
  body: string;
  verified_purchase: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProductReviews {
  items: Review[];
  total: number;
  page: number;
  limit: number;
  summary: { average: number; count: number; histogram: Record<string, number> };
}

export interface AdminReview extends Review {
  product_name: string;
  reviewer_email: string;
}

export interface Inventory {
  product_id: string;
  available_quantity: number;
  reserved_quantity: number;
}

export interface InventoryListItem extends Inventory {
  product_name: string;
  sku: string;
}

export type StockOperation = "add-stock" | "remove-stock" | "reserve" | "release" | "confirm";

export type OrderStatus =
  | "CREATED"
  | "PAYMENT_PENDING"
  | "PAID"
  | "PACKED"
  | "SHIPPED"
  | "DELIVERED"
  | "CANCELLED";

export interface OrderItem {
  product_id: string;
  product_name: string;
  sku: string;
  image_url: string;
  quantity: number;
  price: number;
  sub_total: number;
}

export interface Order {
  id: string;
  user_id: string;
  status: OrderStatus;
  total_amount: number;
  items: OrderItem[];
  created_at: string;
}

export interface OrderListItem {
  id: string;
  user_id: string;
  status: OrderStatus;
  total_amount: number;
  total_items: number;
  created_at: string;
}

export interface OrderStatusSummary {
  total_orders: number;
  by_status: Record<OrderStatus, number>;
}

export type PaymentStatus = "PENDING" | "SUCCESS" | "FAILED" | "REFUNDED";

export interface Payment {
  id: string;
  order_id: string;
  gateway_order_id: string;
  transaction_id: string;
  status: PaymentStatus;
  amount: number;
  currency: string;
  gateway: string;
  created_at: string;
}

export interface PaymentListItem {
  id: string;
  order_id: string;
  status: PaymentStatus;
  amount: number;
  currency: string;
  gateway: string;
  created_at: string;
}

export interface PaymentSummary {
  total_payments: number;
  successful_payments: number;
  failed_payments: number;
  pending_payments: number;
  refunded_payments: number;
  revenue: number;
}

export interface CheckoutSession {
  payment_id: string;
  order_id: string;
  amount: number;
  currency: string;
  key_id: string;
  gateway: string;
  gateway_order_id: string;
}

export interface CheckoutResult {
  razorpay_order_id: string;
  razorpay_payment_id: string;
  razorpay_signature: string;
}
