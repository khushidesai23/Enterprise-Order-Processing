import { lazy, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { RequireAdmin, RequireAuth } from "./components/common";
import { StoreLayout } from "./layouts/StoreLayout";
import { LoginPage } from "./pages/auth/LoginPage";
import { RegisterPage } from "./pages/auth/RegisterPage";
import { CatalogPage } from "./pages/shop/CatalogPage";
import { ProductPage } from "./pages/shop/ProductPage";
import { CartPage } from "./pages/shop/CartPage";
import { CheckoutPage } from "./pages/shop/CheckoutPage";
import { OrdersPage } from "./pages/shop/OrdersPage";
import { OrderDetailPage } from "./pages/shop/OrderDetailPage";
import { AccountPage } from "./pages/shop/AccountPage";
import { NotFoundPage } from "./pages/NotFoundPage";

// The admin console (and its charting library) loads only for admins.
const AdminLayout = lazy(() => import("./layouts/AdminLayout").then((m) => ({ default: m.AdminLayout })));
const OverviewPage = lazy(() => import("./pages/admin/OverviewPage").then((m) => ({ default: m.OverviewPage })));
const ProductsAdminPage = lazy(() => import("./pages/admin/ProductsAdminPage").then((m) => ({ default: m.ProductsAdminPage })));
const CategoriesAdminPage = lazy(() => import("./pages/admin/CategoriesAdminPage").then((m) => ({ default: m.CategoriesAdminPage })));
const InventoryAdminPage = lazy(() => import("./pages/admin/InventoryAdminPage").then((m) => ({ default: m.InventoryAdminPage })));
const OrdersAdminPage = lazy(() => import("./pages/admin/OrdersAdminPage").then((m) => ({ default: m.OrdersAdminPage })));
const OrderAdminDetailPage = lazy(() => import("./pages/admin/OrderAdminDetailPage").then((m) => ({ default: m.OrderAdminDetailPage })));
const PaymentsAdminPage = lazy(() => import("./pages/admin/PaymentsAdminPage").then((m) => ({ default: m.PaymentsAdminPage })));
const CustomersAdminPage = lazy(() => import("./pages/admin/CustomersAdminPage").then((m) => ({ default: m.CustomersAdminPage })));

function Loading() {
  return (
    <div className="flex min-h-[50vh] items-center justify-center text-subtle">
      <Loader2 className="size-6 animate-spin" />
    </div>
  );
}

export function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/shop" replace />} />

      <Route element={<StoreLayout />}>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/shop" element={<CatalogPage />} />
        <Route path="/shop/products/:id" element={<ProductPage />} />
        <Route path="/shop/cart" element={<CartPage />} />
        <Route
          path="/shop/checkout"
          element={
            <RequireAuth>
              <CheckoutPage />
            </RequireAuth>
          }
        />
        <Route
          path="/shop/orders"
          element={
            <RequireAuth>
              <OrdersPage />
            </RequireAuth>
          }
        />
        <Route
          path="/shop/orders/:id"
          element={
            <RequireAuth>
              <OrderDetailPage />
            </RequireAuth>
          }
        />
        <Route
          path="/shop/account"
          element={
            <RequireAuth>
              <AccountPage />
            </RequireAuth>
          }
        />
        <Route path="*" element={<NotFoundPage />} />
      </Route>

      <Route
        path="/admin"
        element={
          <RequireAdmin>
            <Suspense fallback={<Loading />}>
              <AdminLayout />
            </Suspense>
          </RequireAdmin>
        }
      >
        <Route index element={<OverviewPage />} />
        <Route path="products" element={<ProductsAdminPage />} />
        <Route path="categories" element={<CategoriesAdminPage />} />
        <Route path="inventory" element={<InventoryAdminPage />} />
        <Route path="orders" element={<OrdersAdminPage />} />
        <Route path="orders/:id" element={<OrderAdminDetailPage />} />
        <Route path="payments" element={<PaymentsAdminPage />} />
        <Route path="customers" element={<CustomersAdminPage />} />
        <Route path="*" element={<NotFoundPage admin />} />
      </Route>
    </Routes>
  );
}
