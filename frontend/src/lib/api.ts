import type { ApiEnvelope } from "./types";

const TOKEN_KEY = "meridian.token";

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export const tokenStore = {
  get(): string | null {
    try {
      return localStorage.getItem(TOKEN_KEY);
    } catch {
      return null;
    }
  },
  set(token: string | null) {
    try {
      if (token) localStorage.setItem(TOKEN_KEY, token);
      else localStorage.removeItem(TOKEN_KEY);
    } catch {
      // Storage unavailable: the session lasts until the page reloads.
    }
  },
};

// Called when an authenticated request comes back 401 (expired or revoked
// token), so the auth provider can sign the user out.
let onUnauthorized: (() => void) | null = null;

export function setUnauthorizedHandler(handler: (() => void) | null) {
  onUnauthorized = handler;
}

const BASE = "/api/v1";

type Method = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

// request unwraps the API envelope ({ success, message, data, error }) and
// throws ApiError carrying the server's message on any failure.
export async function request<T>(method: Method, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { Accept: "application/json" };
  const token = tokenStore.get();
  if (token) headers.Authorization = `Bearer ${token}`;
  if (body !== undefined) headers["Content-Type"] = "application/json";

  let response: Response;
  try {
    response = await fetch(BASE + path, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new ApiError("Cannot reach the server. Check that the API is running.", 0);
  }

  let envelope: ApiEnvelope<T> | null = null;
  try {
    envelope = (await response.json()) as ApiEnvelope<T>;
  } catch {
    envelope = null;
  }

  if (!response.ok || !envelope || envelope.success === false) {
    if (response.status === 401 && token) onUnauthorized?.();
    const message = envelope?.error || envelope?.message || `Request failed (${response.status})`;
    throw new ApiError(humanize(message), response.status);
  }

  return envelope.data as T;
}

const VALIDATION_REASONS: Record<string, string> = {
  required: "is required",
  min: "is too short",
  max: "is too long",
  email: "must be a valid email address",
  gt: "must be greater than zero",
  gte: "cannot be negative",
};

// Validation errors from Gin and go-playground/validator read like
// "Key: 'CreateUserRequest.Password' Error:Field validation for 'Password'
// failed on the 'min' tag"; turn them into a sentence.
export function humanize(message: string): string {
  const match = message.match(/Field validation for '(\w+)' failed on the '(\w+)' tag/);
  if (!match) return message.charAt(0).toUpperCase() + message.slice(1);

  const [, field, tag] = match;
  const name = field
    .replace(/ID$/, " ID")
    .replace(/([a-z])([A-Z])/g, "$1 $2")
    .split(" ")
    .map((word) => (word === "ID" ? word : word.toLowerCase()))
    .join(" ");
  const reason = VALIDATION_REASONS[tag] ?? "is invalid";

  return `${name.charAt(0).toUpperCase()}${name.slice(1)} ${reason}.`;
}

export function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  return "Something went wrong";
}
