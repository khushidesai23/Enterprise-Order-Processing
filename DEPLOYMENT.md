# Deploying: Netlify + Render + Neon

Netlify serves static sites, so it hosts the **frontend** only. The Go
**API** is a long-running server with a database pool, which Netlify cannot
run, so it goes on **Render** (Docker). **PostgreSQL** runs on **Neon**. All
three have free tiers.

```text
Browser -> https://your-site.netlify.app          (React app, from frontend/dist)
              /api/*  --Netlify proxy-->  https://your-api.onrender.com/api/*   (Go API, Docker)
                                                      |
                                                      v
                                                Neon PostgreSQL
```

Netlify forwards `/api/*` to Render, so the browser only ever talks to your
Netlify domain. Kafka/Debezium, Prometheus, Grafana and Jaeger are not
deployed: the API runs without them (`KAFKA_CDC_ENABLED=false`,
`OTEL_ENABLED=false`).

The repository already contains everything the platforms need:

| File | Used by |
|---|---|
| `Dockerfile`, `.dockerignore` | Render builds the API image (server + seeder) |
| `render.yaml` | Render blueprint: service, health check, environment variables |
| `netlify.toml` | Netlify build settings |
| `frontend/scripts/netlify-redirects.mjs` | Writes the `/api` proxy and SPA fallback from `API_ORIGIN` |
| `frontend/public/_headers` | Caching and security headers on Netlify |

## 0. Push the code

Render and Netlify deploy from GitHub. Merge `feature/frontend` into the
branch you want to deploy (for example `develop` or `main`) and push it.

## 1. Database: Neon

1. Sign up at https://neon.tech and create a project (pick a region close to
   your Render region, e.g. AWS Singapore or Frankfurt).
2. Open **Connection details**. Turn **off** "Connection pooling", so the
   host does not contain `-pooler`. GORM's prepared statements need a
   direct connection.
3. Note the **host**, **database**, **user** and **password**.

The API creates its tables on first start (`AutoMigrate`), so there is
nothing to run by hand.

## 2. API: Render

1. Sign up at https://render.com with GitHub.
2. Go to **New → Blueprint**, choose this repository and the branch from
   step 0. Render reads `render.yaml` and asks for the values marked
   `sync: false`:

   | Variable | Value |
   |---|---|
   | `DB_HOST` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | From Neon (step 1) |
   | `RAZORPAY_KEY_ID` / `RAZORPAY_KEY_SECRET` | Razorpay dashboard → Account & Settings → API keys (test keys are fine) |
   | `RAZORPAY_WEBHOOK_SECRET` | Any long random string; reuse it in step 5 |
   | `CORS_ALLOWED_ORIGINS` | Leave as `https://example.com` for now; fix it in step 4 |

   `JWT_SECRET` is generated for you.
3. Apply it. The first build takes a few minutes. When it is live, open
   `https://<your-api>.onrender.com/api/v1/health`. It should report
   `"status":"UP"` with the database `"up"`.

## 3. (Optional) Load the demo store into Neon

Run the seeder from your machine against Neon. In PowerShell, from the
repository root:

```powershell
$env:DB_HOST = "<neon host>"; $env:DB_PORT = "5432"; $env:DB_USER = "<user>"
$env:DB_PASSWORD = "<password>"; $env:DB_NAME = "<database>"; $env:DB_SSLMODE = "require"
$env:DB_MAX_OPEN_CONNS = "10"; $env:DB_MAX_IDLE_CONNS = "5"; $env:KAFKA_CDC_ENABLED = "false"
go run ./cmd/seed          # ~840 products, 1,500 customers, 25,000 orders, ~11,700 reviews
```

Environment variables take precedence over your local `.env`. Close the
terminal afterwards so later commands don't point at Neon by accident. For
a lighter store: `go run ./cmd/seed -customers 300 -orders 3000`.

## 4. Frontend: Netlify

1. Sign up at https://netlify.com and choose **Add new site → Import an
   existing project**, then this GitHub repository and branch. Netlify reads
   `netlify.toml` (base `frontend`, build `npm run build && node
   scripts/netlify-redirects.mjs`, publish `dist`), so leave the build
   settings as they are.
2. Before deploying, add an environment variable: **`API_ORIGIN`** =
   `https://<your-api>.onrender.com` (https, no trailing path). The build
   fails with a clear message if it is missing.
3. Deploy. Note the site URL, e.g. `https://meridian-shop.netlify.app`. You
   can rename the site under Site configuration.
4. Back on Render, set **`CORS_ALLOWED_ORIGINS`** to that exact URL. Add
   your custom domain too if you attach one, comma-separated. Save; Render
   redeploys.

   This is required because Netlify's proxy forwards the browser's `Origin`
   header. Without it, sign-in and checkout fail with 403.

## 5. Razorpay webhook

In Razorpay Dashboard → **Webhooks → Add**:

- **URL:** `https://<your-api>.onrender.com/api/v1/payments/webhook`.
  Point it straight at Render: Netlify's proxy gives up after 26 seconds,
  and a sleeping free API can take longer to wake.
- **Secret:** the `RAZORPAY_WEBHOOK_SECRET` from step 2.
- **Events:** `payment.captured`, `payment.failed`, `refund.created`.

Checkout already confirms payments through `POST /payments/verify`, so the
webhook is a backstop for browsers that close before reporting back.

## 6. First admin

Register on your site, then make that account an admin in Neon's **SQL
Editor**:

```sql
UPDATE users SET role = 'admin' WHERE email = 'you@example.com';
```

Sign out and back in, because the role travels in the login token. The admin
console is at `/admin`.

## 7. Check it

- `https://<site>.netlify.app/shop` loads products. The footer should read
  "All systems operational", which is the API health check through the proxy.
- Register, add to bag, check out with a Razorpay
  [test card](https://razorpay.com/docs/payments/payments/test-card-details/)
  and see the order reach **Paid**.
- In `/admin`, fulfil the order and refund it. Refunds of real test payments
  go through Razorpay.

## Things to know

- **Free Render services sleep** after about 15 minutes idle, and the first
  request then takes 30–60 s. Netlify's proxy waits at most 26 s, so the
  very first load after a nap may show an error that clears on retry (the
  app retries failed reads). To avoid it, use Render's paid Starter plan, or
  ping `/api/v1/health` every 10 minutes with a free monitor such as
  UptimeRobot.
- **Neon's free tier** pauses compute when idle and wakes in about a second.
  Storage is limited (0.5 GB), which is plenty for the demo store.
- **Secrets** live only in the Render and Netlify dashboards. `.env` is
  never committed and is not copied into the Docker image.
- **Seeded payments** use made-up Razorpay IDs, so refunding them fails at
  Razorpay. Payments made through checkout refund normally.
- **Other hosts:** the `Dockerfile` runs on any container platform (Railway,
  Fly.io, Koyeb, a VPS). Set the same environment variables as
  `render.yaml`, and point `API_ORIGIN` at that host instead.

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| Netlify build fails: "set API_ORIGIN" | Add `API_ORIGIN` under Site configuration → Environment variables, then redeploy |
| Sign-in or checkout returns 403 | `CORS_ALLOWED_ORIGINS` on Render does not exactly match the site URL (scheme, no trailing slash) |
| 502/504 on the first load | Render was asleep; retry, or keep it warm (see above) |
| Render deploy fails health check | Database settings wrong: check `DB_*`, `DB_SSLMODE=require` and the non-pooler host |
| Payments stay "Payment pending" | `RAZORPAY_KEY_ID`/`SECRET` on Render must be the same key pair as your Razorpay account's test mode |
