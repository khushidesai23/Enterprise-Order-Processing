// Seeds a demo catalogue (categories, products, stock) and a demo customer
// through the public API. Safe to run repeatedly: existing records are
// left alone.
//
//   ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD=... npm run seed
//
// The admin account must already exist and be promoted:
//   make promote-admin EMAIL=admin@example.com

const API = (process.env.API_URL || "http://localhost:8080").replace(/\/$/, "") + "/api/v1";
const ADMIN_EMAIL = process.env.ADMIN_EMAIL;
const ADMIN_PASSWORD = process.env.ADMIN_PASSWORD;
const CUSTOMER = {
  first_name: "Demo",
  last_name: "Customer",
  email: process.env.CUSTOMER_EMAIL || "customer@meridian.test",
  password: process.env.CUSTOMER_PASSWORD || "customer123",
};

const CATALOGUE = {
  "Home & Kitchen": {
    description: "Cookware, storage and everyday home essentials.",
    products: [
      ["Cast Iron Skillet 26cm", "HK-SKILLET-26", 1899, 40, "Pre-seasoned cast iron that goes from stovetop to oven. Even heat, lifetime durability."],
      ["Borosilicate Storage Jars (Set of 4)", "HK-JARS-4", 1249, 60, "Airtight bamboo lids, heat-safe glass. Ideal for pantry staples."],
      ["Stainless Steel Water Bottle 1L", "HK-BOTTLE-1L", 799, 120, "Double-wall insulated: cold for 24 hours, hot for 12."],
      ["Ceramic Pour-Over Coffee Set", "HK-POUROVER", 2199, 8, "Hand-glazed dripper with carafe and two cups."],
    ],
  },
  Electronics: {
    description: "Audio, charging and smart accessories.",
    products: [
      ["Wireless Earbuds Pro", "EL-EARBUDS-PRO", 4999, 35, "Active noise cancellation, 30-hour battery with the case, IPX4 sweat resistance."],
      ["65W GaN USB-C Charger", "EL-GAN-65W", 2499, 50, "Charges a laptop, phone and earbuds at once from a pocket-sized brick."],
      ["Mechanical Keyboard (Brown Switches)", "EL-KEYBOARD-BRN", 5499, 6, "Hot-swappable switches, PBT keycaps and per-key RGB."],
      ["Smart LED Desk Lamp", "EL-DESKLAMP", 3299, 25, "Adjustable colour temperature, app control and a built-in wireless charger."],
    ],
  },
  "Apparel": {
    description: "Comfortable basics made to last.",
    products: [
      ["Organic Cotton Crew Tee", "AP-TEE-CREW", 699, 150, "Heavyweight 220gsm organic cotton with a relaxed fit."],
      ["Merino Wool Beanie", "AP-BEANIE", 999, 4, "Soft, breathable and naturally temperature regulating."],
      ["Everyday Canvas Tote", "AP-TOTE", 549, 90, "Waxed canvas with an inside zip pocket."],
    ],
  },
  "Health & Wellness": {
    description: "Fitness gear and self-care.",
    products: [
      ["Non-Slip Yoga Mat 6mm", "HW-YOGAMAT", 1499, 45, "Dense TPE foam with an alignment guide and carry strap."],
      ["Resistance Band Set", "HW-BANDS", 899, 0, "Five levels from extra-light to extra-heavy, with door anchor and handles."],
      ["Aromatherapy Diffuser", "HW-DIFFUSER", 1799, 18, "Ultrasonic mist with seven ambient light colours and auto shut-off."],
    ],
  },
};

async function call(method, path, body, token) {
  const response = await fetch(API + path, {
    method,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const json = await response.json().catch(() => ({}));
  return { status: response.status, data: json.data, error: json.error };
}

function fail(message) {
  console.error(`seed: ${message}`);
  process.exit(1);
}

async function main() {
  if (!ADMIN_EMAIL || !ADMIN_PASSWORD) fail("set ADMIN_EMAIL and ADMIN_PASSWORD to an admin account");

  const login = await call("POST", "/auth/login", { email: ADMIN_EMAIL, password: ADMIN_PASSWORD });
  if (login.status !== 200) fail(`admin login failed: ${login.error}`);
  if (login.data.user.role !== "admin") fail(`${ADMIN_EMAIL} is not an admin: run make promote-admin EMAIL=${ADMIN_EMAIL}`);
  const token = login.data.token;

  const existingCategories = new Map((await call("GET", "/categories")).data.map((c) => [c.name, c.id]));
  const existingSkus = new Set((await call("GET", "/products")).data.map((p) => p.sku));

  let createdProducts = 0;
  for (const [name, { description, products }] of Object.entries(CATALOGUE)) {
    let categoryId = existingCategories.get(name);
    if (!categoryId) {
      const created = await call("POST", "/categories", { name, description }, token);
      if (created.status !== 201) fail(`category ${name}: ${created.error}`);
      categoryId = created.data.id;
      console.log(`+ category ${name}`);
    }

    for (const [productName, sku, price, stock, productDescription] of products) {
      if (existingSkus.has(sku)) continue;
      const product = await call(
        "POST",
        "/products",
        { name: productName, sku, price, description: productDescription, category_id: categoryId },
        token,
      );
      if (product.status !== 201) fail(`product ${sku}: ${product.error}`);
      const inventory = await call("POST", "/inventory", { product_id: product.data.id, available_quantity: stock }, token);
      if (inventory.status !== 201) fail(`inventory ${sku}: ${inventory.error}`);
      createdProducts++;
      console.log(`+ product ${productName} (${stock} in stock)`);
    }
  }

  const customer = await call("POST", "/users", CUSTOMER);
  if (customer.status === 201) console.log(`+ customer ${CUSTOMER.email} / ${CUSTOMER.password}`);
  else if (customer.status !== 409) fail(`customer: ${customer.error}`);

  console.log(`seed: done (${createdProducts} new products)`);
}

main().catch((error) => fail(error.message));
