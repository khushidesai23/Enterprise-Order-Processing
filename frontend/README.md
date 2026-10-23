# Meridian Market: storefront and admin console

React frontend for the Enterprise Order Processing API. It is one app with
two areas:

| Area | Path | Who |
|---|---|---|
| Storefront | `/shop` | Anyone. Signed-in customers can check out, pay and track orders. |
| Admin console | `/admin` | Users with the `admin` role. |

Stack: React 18, TypeScript, Vite 6, Tailwind CSS 4, Radix UI primitives
(shadcn-style components in `src/components/ui`), React Router 6, TanStack
Query 5, Recharts, Vitest and Playwright.

## Running it

Prerequisites: Node 22+, and the Go API running (`make run` from the repo
root, with `docker compose up -d`).

```powershell
cd frontend
npm install
npm run dev            # http://localhost:5173
```

The dev server forwards `/api` to the Go API, so the browser only talks to
its own origin and no CORS setup is needed. Point it at another API with
`API_PROXY_TARGET` (see `.env.example`), for example
`$env:API_PROXY_TARGET = "http://localhost:8081"; npm run dev`.

### First admin and demo data

Registration always creates customers. To get an admin and a demo
catalogue:

```powershell
# 1. Register in the storefront (or POST /api/v1/users), then promote:
make promote-admin EMAIL=admin@meridian.test
#    (or: UPDATE users SET role = 'admin' WHERE email = '...')

# 2. Seed categories, products with stock, and a demo customer
#    (customer@meridian.test / customer123). Safe to re-run.
$env:ADMIN_EMAIL = "admin@meridian.test"; $env:ADMIN_PASSWORD = "..."
npm run seed           # API_URL defaults to http://localhost:8080
```

Sign out and back in after a role change: the role travels in the JWT.

### Production build

```powershell
npm run build          # type-checks, then writes dist/
npm run preview        # serves dist/ on :4173 with the same /api proxy
```

In production, serve `dist/` from any static host and route `/api` to the
Go API on the same origin (for example an nginx `location /api/` proxy),
with a fallback to `index.html` for client-side routes.

## Features

**Storefront**

- Catalogue with category chips (`GET /products/category/:id`), search by
  name/SKU/category, price range, in-stock filter and sorting.
- Product pages with live stock, low-stock and sold-out badges.
- Cart stored in the browser, checked against live stock before checkout.
- Sign-up and sign-in. An expired session returns you to the page you were
  on after signing in again.
- Checkout: `POST /orders` reserves the stock, `POST /payments` creates the
  Razorpay order, Razorpay Checkout takes the payment, and
  `POST /payments/verify` confirms it.
- Orders: history with pagination, a status timeline, payment details,
  **Pay now / Retry payment** for unpaid orders, and **Cancel order**
  (releases the stock).
- Account page to edit your name.

**Admin console**

- Overview: revenue, order, low-stock and customer KPIs, a 14-day captured
  revenue chart, orders by status (from `GET /orders/summary`), recent
  orders and low-stock warnings.
- Products: create (with starting stock), edit and delete, with search and
  a category filter.
- Categories: create, edit, delete (blocked while a category has products).
- Inventory: add and remove stock, set exact levels, reserve, release and
  confirm reservations, low and out-of-stock filters, and a "start
  tracking" action for products without an inventory record.
- Orders: status tabs with server-side filtering (`GET /orders?status=`),
  order detail, fulfilment (packed → shipped → delivered), cancellation of
  unpaid orders, and refunds (before shipment a refund also cancels the
  order).
- Payments: summary, status filter, search and refunds.
- Customers: search, role filter, grant or remove admin, deactivate or
  reactivate, and delete. Admins cannot act on their own account.

Admin pages are hidden from customers in the UI, but the protection is the
API's own `RequireRole(admin)` on every admin route.

## Payments without a public webhook URL

Razorpay cannot reach `localhost`, so locally its webhooks never arrive. When
Razorpay Checkout succeeds, the browser receives
`razorpay_order_id`, `razorpay_payment_id` and `razorpay_signature`, and the
frontend sends them to `POST /api/v1/payments/verify`. The API checks the
HMAC signature with the key secret and the caller's ownership, then completes
the payment the same way a `payment.captured` webhook does. Whichever of the
two arrives second is a no-op, so the webhook still works when configured.

Use Razorpay's [test cards or netbanking](https://razorpay.com/docs/payments/payments/test-card-details/)
with your `rzp_test_` keys.

## Tests

```powershell
npm run typecheck      # TypeScript, including the e2e specs
npm test               # Vitest unit tests (API client, cart, status rules, formatting)
npm run test:e2e       # Playwright browser tests (first time: npx playwright install chromium)
```

The Playwright suite drives the real UI against the real API and database:
registration, browsing and filters, cart, checkout, payment, retry and
cancellation, the admin catalogue, inventory operations, fulfilment,
refunds, customer management and access control. It needs:

- the API running with its `.env` (the suite reads `RAZORPAY_KEY_SECRET`
  from `../.env`, or from the environment);
- an admin account: `ADMIN_EMAIL` / `ADMIN_PASSWORD`, default
  `admin@meridian.test` / `admin12345`;
- internet access, because `POST /payments` creates real Razorpay test-mode
  orders.

Razorpay's checkout widget is a third-party iframe, so the suite swaps
`checkout.js` for a stand-in that returns what the real widget would,
including a correctly signed success response. Everything else is real,
including the backend's signature check. Because no real Razorpay payment
exists, the refund test checks that a gateway rejection is reported and
rolled back. To check a real capture and refund, pay once by hand with a
Razorpay test card and refund it from **Admin → Payments**.

Tests create their own products, orders and `@e2e.test` accounts, and
delete the products and accounts afterwards. Orders stay, as order history.
