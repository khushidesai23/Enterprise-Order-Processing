import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, humanize, request, setUnauthorizedHandler, tokenStore } from "./api";
import { addLine, MAX_LINE_QUANTITY, setLineQuantity, type CartLine } from "./cart";
import { compactMoney, shortId } from "./format";
import { adminNextStatuses, fulfilmentStep, isAwaitingPayment } from "./orders";

function mockFetch(status: number, body: unknown) {
  const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(body), { status }));
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

afterEach(() => {
  vi.unstubAllGlobals();
  tokenStore.set(null);
  setUnauthorizedHandler(null);
});

describe("request", () => {
  it("unwraps the data of a success envelope and sends the bearer token", async () => {
    tokenStore.set("tok");
    const fetchMock = mockFetch(200, { success: true, data: { id: "1" }, timestamp: "" });

    await expect(request("POST", "/orders", { a: 1 })).resolves.toEqual({ id: "1" });

    const [url, init] = fetchMock.mock.calls[0]!;
    expect(url).toBe("/api/v1/orders");
    expect(init.headers.Authorization).toBe("Bearer tok");
    expect(init.body).toBe('{"a":1}');
  });

  it("throws ApiError with the server's error message and status", async () => {
    mockFetch(409, { success: false, error: "product SKU already exists", timestamp: "" });

    const error = await request("POST", "/products").catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect((error as ApiError).status).toBe(409);
    expect((error as ApiError).message).toBe("Product SKU already exists");
  });

  it("signs the user out when an authenticated request is rejected", async () => {
    tokenStore.set("expired");
    const onUnauthorized = vi.fn();
    setUnauthorizedHandler(onUnauthorized);
    mockFetch(401, { success: false, error: "token is expired", timestamp: "" });

    await expect(request("GET", "/auth/me")).rejects.toThrow(ApiError);
    expect(onUnauthorized).toHaveBeenCalledOnce();
  });

  it("does not sign out on a failed login (no token yet)", async () => {
    const onUnauthorized = vi.fn();
    setUnauthorizedHandler(onUnauthorized);
    mockFetch(401, { success: false, error: "invalid email or password", timestamp: "" });

    await expect(request("POST", "/auth/login")).rejects.toThrow("Invalid email or password");
    expect(onUnauthorized).not.toHaveBeenCalled();
  });

  it("reports an unreachable server", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")));

    await expect(request("GET", "/products")).rejects.toMatchObject({ status: 0 });
  });
});

describe("humanize", () => {
  it("rewrites validator errors", () => {
    expect(humanize("Key: 'CreateUserRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag")).toBe(
      "Password is too short.",
    );
    expect(humanize("Key: 'X.CategoryID' Error:Field validation for 'CategoryID' failed on the 'required' tag")).toBe(
      "Category ID is required.",
    );
  });
});

describe("cart", () => {
  const lamp = { productId: "p1", name: "Lamp", sku: "L", price: 100, category: "Home" };

  it("merges repeated adds and caps the quantity", () => {
    let lines: CartLine[] = addLine([], lamp, 2);
    lines = addLine(lines, lamp, 3);
    expect(lines).toEqual([{ ...lamp, quantity: 5 }]);
    expect(addLine(lines, lamp, 500)[0]!.quantity).toBe(MAX_LINE_QUANTITY);
  });

  it("removes a line when its quantity drops to zero", () => {
    const lines = addLine([], lamp, 1);
    expect(setLineQuantity(lines, "p1", 4)[0]!.quantity).toBe(4);
    expect(setLineQuantity(lines, "p1", 0)).toEqual([]);
  });
});

describe("order status rules", () => {
  it("mirrors the backend's admin transitions", () => {
    expect(adminNextStatuses("PAID")).toEqual(["PACKED"]);
    expect(adminNextStatuses("PACKED")).toEqual(["SHIPPED"]);
    expect(adminNextStatuses("SHIPPED")).toEqual(["DELIVERED"]);
    expect(adminNextStatuses("CREATED")).toEqual(["CANCELLED"]);
    expect(adminNextStatuses("DELIVERED")).toEqual([]);
    expect(adminNextStatuses("CANCELLED")).toEqual([]);
  });

  it("knows which orders still await payment", () => {
    expect(isAwaitingPayment("CREATED")).toBe(true);
    expect(isAwaitingPayment("PAYMENT_PENDING")).toBe(true);
    expect(isAwaitingPayment("PAID")).toBe(false);
    expect(fulfilmentStep("PAYMENT_PENDING")).toBe(0);
    expect(fulfilmentStep("DELIVERED")).toBe(4);
  });
});

describe("format", () => {
  it("abbreviates rupees in Indian units", () => {
    expect(compactMoney(950)).toBe("₹950");
    expect(compactMoney(12_500)).toBe("₹12.5k");
    expect(compactMoney(350_000)).toBe("₹3.5L");
    expect(compactMoney(12_000_000)).toBe("₹1.2Cr");
  });

  it("shortens UUIDs to a readable reference", () => {
    expect(shortId("b9cbd703-e2ee-4c2e-b98a-1fce66e98962")).toBe("#B9CBD703");
  });
});
