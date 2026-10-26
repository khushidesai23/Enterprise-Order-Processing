# meridian/: storefront and admin console

React frontend for the Enterprise Order Processing API. It is one app with
two areas:

| Area | Path | Who |
|---|---|---|
| Storefront | `/shop` | Anyone. Signed-in customers can check out, pay and track orders. |
| Admin console | `/admin` | Users with the `admin` role. |

Stack: React 18, TypeScript, Vite 6, Tailwind CSS 4, Motion for React, Radix UI primitives
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

### Demo data and the first admin

Fill the database with a large demo store, from the repository root:

```powershell
go run ./cmd/seed      # or: make seed
```

| Seeded | Amount |
|---|---|
| Departments | 24 |
| Products | ~840 listings with photos, brands, MRP and specs: the [DummyJSON](https://dummyjson.com) catalogue (`cmd/seed/data/products.json`) expanded into colour, storage and pack variants |
| Stock | Every product, about 6% sold out and 10% low |
| Customers | 1,500 (`*@shopper.example.com`, password `password123`) |
| Orders | 25,000 over the past year, in every status, with matching payments and stock reservations |
| Reviews | ~11,700, most from verified purchases |

The seeder is deterministic and safe to re-run: products and customers are
matched by SKU and email, and orders and reviews are only generated once
(`-force-orders` adds more). Scale it with flags, for example
`go run ./cmd/seed -customers 5000 -orders 100000`.

Seeded payments are written straight to the database, so refunding one in
the admin console fails at Razorpay (its payment IDs are fake). Real
payments made through checkout refund normally.

Product photos load from DummyJSON's CDN. Offline, products fall back to
generated tiles.

Registration always creates customers. To get an admin, register in the
storefront, then promote the account:

```powershell
make promote-admin EMAIL=admin@meridian.test
#    (or: UPDATE users SET role = 'admin' WHERE email = '...')
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

**Storefront: meridian/**

The storefront has its own identity, separate from the admin console: a
neo-editorial "index of objects".

- **Visual language.** Warm bone paper, deep ink, an acid-lime signal
  colour and an ultraviolet accent. Type is Syne (display), Instrument
  Serif italics (editorial accents), Geist (text) and Geist Mono (data
  such as SKUs, prices and counts). The identity is a globe-and-meridian
  mark with a slashed wordmark.
- **Code.** Tokens live in `src/index.css` (`--color-bone`, `--color-ink`,
  `--color-acid`, ... and `--font-display`, ...). The store's own
  components are in `src/store/` (shell, tiles, bag, rails, order
  tracking, motion primitives).
- **Motion.** Built with [Motion for React](https://motion.dev):
  - masked word reveals and staggered entrances
  - a CSS 3D orbit of each department's best seller that tilts with the
    pointer
  - a horizontal "most wanted" gallery driven by vertical scroll
  - magnetic buttons, image zoom and crossfading galleries
  - a fly-to-bag animation, animated order tracking, and page fades
  - a cursor-following preview in the departments index

  Everything respects `prefers-reduced-motion` (`MotionConfig
  reducedMotion="user"` plus CSS fallbacks). There is no Three.js: CSS 3D
  covers the orbit without adding hundreds of kilobytes.

Pages and features, all on live API data:

- **Home:** live catalogue counts, a hero orbit, today's drops (biggest
  real discounts), a promise marquee, a department index with live
  counts, most wanted (each department's best seller by units sold), quotes
  from real reviews, and new arrivals.
- **Search:** a command palette (⌘K, Ctrl+K or `/`) with live results.
- **The index** (`/shop/search`, `GET /products/search`): word search;
  department, brand, rating, price, drops and in-stock filters with counts;
  sorting; pages. Filters live in the URL. On phones they open in a bottom
  sheet.
- **Product pages:**
  - a pointer-zoom gallery
  - real variant options: sibling listings such as storage, colour or
    pack, each with live stock
  - a live stock meter, a quantity stepper, Add to bag and Buy now
  - delivery, returns and warranty, plus collapsible details and specs
  - related rails and reviews
- **Reviews:** rating histogram filter, sorting and pages, and write, edit
  or delete your own review. Buyers' reviews are marked **Verified
  Purchase**.
- **Bag:** a slide-in drawer and a full bag page, both checked against
  live stock and price. Save for later moves items to the wishlist.
- **Wishlist and recently viewed:** stored in the browser.
- **Checkout:** `POST /orders` reserves the stock, `POST /payments` opens
  Razorpay, and `POST /payments/verify` confirms it. The UI says "It's
  yours" only after the backend reports the order `PAID`. Duplicate
  submissions are blocked while a request is in flight.
- **Orders:** history with lifecycle dots, plus an order page with an
  animated tracker (placed, paid, packed, shipped, delivered), payment
  details, and **Pay now / Retry payment** and **Cancel order** for unpaid
  orders.
- **Account:** sign-up, sign-in and the account page. An expired session
  returns you to the page you were on after signing in again.
- **Footer:** a live API and database health indicator
  (`GET /api/v1/health`).

**Admin console**

- Overview: revenue, order, low-stock and customer KPIs, a 14-day captured
  revenue chart, orders by status (from `GET /orders/summary`), recent
  orders and low-stock warnings.
- Products: create (with starting stock), edit and delete, including
  brand, MRP and images; server-side search, category filter and pages.
- Categories: create, edit, delete (blocked while a category has products).
- Inventory: add and remove stock, set exact levels, reserve, release and
  confirm reservations, low and out-of-stock filters, and a "start
  tracking" action for products without an inventory record.
- Orders: status tabs with server-side filtering (`GET /orders?status=`),
  order detail, fulfilment (packed → shipped → delivered), cancellation of
  unpaid orders, and refunds (before shipment a refund also cancels the
  order).
- Payments: summary, status filter, search and refunds.
- Reviews: all reviews newest first, filter by stars, delete (the
  product's rating is recalculated).
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
the home page, search filters and sorting, registration, cart, checkout,
payment, retry and cancellation, verified reviews, the wishlist, the admin
catalogue, inventory operations, fulfilment, refunds, review moderation,
customer management and access control. It needs:

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
