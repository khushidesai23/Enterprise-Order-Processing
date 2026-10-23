import { request } from "./api";
import type {
  Category,
  CategoryListItem,
  CheckoutResult,
  CheckoutSession,
  Inventory,
  InventoryListItem,
  LoginResponse,
  Order,
  OrderListItem,
  OrderStatus,
  OrderStatusSummary,
  Payment,
  PaymentListItem,
  PaymentSummary,
  Product,
  ProductInput,
  ProductListItem,
  Role,
  StockOperation,
  User,
  UserInfo,
} from "./types";

// One function per route registered in internal/api/routes/router.go.

export const authApi = {
  login: (email: string, password: string) =>
    request<LoginResponse>("POST", "/auth/login", { email, password }),
  me: () => request<UserInfo>("GET", "/auth/me"),
};

export const usersApi = {
  register: (input: { first_name: string; last_name: string; email: string; password: string }) =>
    request<User>("POST", "/users", input),
  list: () => request<User[]>("GET", "/users"),
  get: (id: string) => request<User>("GET", `/users/${id}`),
  update: (id: string, input: { first_name?: string; last_name?: string; is_active?: boolean }) =>
    request<User>("PUT", `/users/${id}`, input),
  updateRole: (id: string, role: Role) => request<User>("PATCH", `/users/${id}/role`, { role }),
  remove: (id: string) => request<null>("DELETE", `/users/${id}`),
};

export const categoriesApi = {
  list: () => request<CategoryListItem[]>("GET", "/categories"),
  get: (id: string) => request<Category>("GET", `/categories/${id}`),
  create: (input: { name: string; description: string }) =>
    request<Category>("POST", "/categories", input),
  update: (id: string, input: { name: string; description: string }) =>
    request<Category>("PUT", `/categories/${id}`, input),
  remove: (id: string) => request<null>("DELETE", `/categories/${id}`),
};

export const productsApi = {
  list: () => request<ProductListItem[]>("GET", "/products"),
  get: (id: string) => request<Product>("GET", `/products/${id}`),
  byCategory: (categoryId: string) =>
    request<ProductListItem[]>("GET", `/products/category/${categoryId}`),
  create: (input: ProductInput) => request<Product>("POST", "/products", input),
  update: (id: string, input: ProductInput) => request<Product>("PUT", `/products/${id}`, input),
  remove: (id: string) => request<null>("DELETE", `/products/${id}`),
};

export const inventoryApi = {
  list: () => request<InventoryListItem[]>("GET", "/inventory"),
  get: (productId: string) => request<Inventory>("GET", `/inventory/${productId}`),
  create: (productId: string, availableQuantity: number) =>
    request<Inventory>("POST", "/inventory", {
      product_id: productId,
      available_quantity: availableQuantity,
    }),
  update: (productId: string, available: number, reserved: number) =>
    request<Inventory>("PUT", `/inventory/${productId}`, {
      available_quantity: available,
      reserved_quantity: reserved,
    }),
  operate: (productId: string, operation: StockOperation, quantity: number) =>
    request<Inventory>("PATCH", `/inventory/${productId}/${operation}`, { quantity }),
};

export const ORDER_PAGE_SIZE = 20;

export const ordersApi = {
  create: (items: { product_id: string; quantity: number }[]) =>
    request<Order>("POST", "/orders", { items }),
  // All orders for admins (optionally one status), own orders for customers.
  list: (page = 1, limit = ORDER_PAGE_SIZE, status?: OrderStatus) =>
    request<OrderListItem[]>("GET", `/orders?page=${page}&limit=${limit}${status ? `&status=${status}` : ""}`),
  summary: () => request<OrderStatusSummary>("GET", "/orders/summary"),
  mine: (page = 1, limit = ORDER_PAGE_SIZE) =>
    request<OrderListItem[]>("GET", `/orders/me?page=${page}&limit=${limit}`),
  get: (id: string) => request<Order>("GET", `/orders/${id}`),
  updateStatus: (id: string, status: OrderStatus) =>
    request<Order>("PATCH", `/orders/${id}/status`, { status }),
  cancel: (id: string) => request<Order>("PATCH", `/orders/${id}/cancel`),
};

export const paymentsApi = {
  checkout: (orderId: string) => request<CheckoutSession>("POST", "/payments", { order_id: orderId }),
  verify: (result: CheckoutResult) => request<Payment>("POST", "/payments/verify", result),
  list: () => request<PaymentListItem[]>("GET", "/payments"),
  summary: () => request<PaymentSummary>("GET", "/payments/summary"),
  get: (id: string) => request<Payment>("GET", `/payments/${id}`),
  byOrder: (orderId: string) => request<Payment>("GET", `/payments/order/${orderId}`),
  refund: (id: string) => request<Payment>("POST", `/payments/${id}/refund`),
};

export const healthApi = {
  health: () => request<{ status: string; database: { status: string } }>("GET", "/health"),
};
