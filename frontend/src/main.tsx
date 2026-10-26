import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "sonner";
import { MotionConfig } from "motion/react";
import { ApiError } from "./lib/api";
import { AuthProvider } from "./lib/auth";
import { CartProvider } from "./lib/cart";
import { WishlistProvider } from "./lib/saved";
import { App } from "./App";
import "./index.css";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 15_000,
      refetchOnWindowFocus: false,
      // Client errors (403, 404, ...) will not succeed on retry.
      retry: (count, error) => !(error instanceof ApiError && error.status >= 400 && error.status < 500) && count < 2,
    },
  },
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <MotionConfig reducedMotion="user">
    <QueryClientProvider client={queryClient}>
      <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
        <AuthProvider>
          <CartProvider>
            <WishlistProvider>
              <App />
              <Toaster
                position="bottom-center"
                toastOptions={{
                  classNames: {
                    toast: "!rounded-2xl !border-0 !bg-[#0c0c0d] !text-[#f2efe8] !shadow-2xl !font-[Geist,Inter,sans-serif]",
                    description: "!text-[#f2efe8]/60",
                    error: "!bg-[#ff5a1f] !text-white",
                  },
                }}
              />
            </WishlistProvider>
          </CartProvider>
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
    </MotionConfig>
  </StrictMode>,
);
