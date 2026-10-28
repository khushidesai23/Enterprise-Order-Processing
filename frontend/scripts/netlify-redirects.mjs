// Writes dist/_redirects for Netlify after `npm run build`:
//
//   /api/*  -> the Go API (proxied, so the browser stays on one origin)
//   /*      -> index.html (client-side routes such as /shop/orders/123)
//
// The API's address comes from the API_ORIGIN environment variable, set in
// Netlify's site settings, so it is not hard-coded in the repository.
import { writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

const raw = (process.env.API_ORIGIN || "").trim().replace(/\/+$/, "");

let origin;
try {
  origin = new URL(raw);
} catch {
  console.error("netlify-redirects: set API_ORIGIN to the API's public URL, e.g. https://meridian-api.onrender.com");
  process.exit(1);
}
if (origin.protocol !== "https:" || origin.pathname !== "/") {
  console.error(`netlify-redirects: API_ORIGIN must be an https origin without a path, got "${raw}"`);
  process.exit(1);
}

const rules = [`/api/*  ${origin.origin}/api/:splat  200`, `/*  /index.html  200`, ""].join("\n");
const target = fileURLToPath(new URL("../dist/_redirects", import.meta.url));
writeFileSync(target, rules);
console.log(`netlify-redirects: /api/* -> ${origin.origin}/api/*`);
