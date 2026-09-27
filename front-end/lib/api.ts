import type {
  ApiEnvelope,
  ApiErrorPayload,
  ApiMessage,
  LoginResponse,
  Session,
  User,
  Wallet,
  WalletLedgerResult,
  WalletTopup,
} from "@/lib/types";

const API_URL = (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080").replace(/\/$/, "");

let accessToken: string | null = null;
let refreshPromise: Promise<LoginResponse> | null = null;

export class ApiError extends Error {
  readonly code: string;
  readonly fields: Record<string, string>;
  readonly status: number;

  constructor(message: string, status: number, code = "REQUEST_FAILED", fields: Record<string, string> = {}) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.fields = fields;
    this.status = status;
  }
}

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken() {
  return accessToken;
}

async function parseResponse<T>(response: Response): Promise<T> {
  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json") ? ((await response.json()) as ApiErrorPayload & T) : null;

  if (!response.ok) {
    const error = payload as ApiErrorPayload | null;
    throw new ApiError(
      error?.error?.message ?? "Something went wrong. Please try again.",
      response.status,
      error?.error?.code,
      error?.error?.fields ?? {},
    );
  }

  return payload as T;
}

async function request<T>(path: string, init: RequestInit = {}, requiresAuth = false, allowRefresh = true): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");
  if (init.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  if (requiresAuth && accessToken) headers.set("Authorization", `Bearer ${accessToken}`);

  const response = await fetch(`${API_URL}${path}`, {
    ...init,
    headers,
    credentials: "include",
    cache: "no-store",
  });

  if (response.status === 401 && requiresAuth && allowRefresh && path !== "/api/v1/auth/refresh") {
    try {
      await refreshAccessToken();
      return request<T>(path, init, requiresAuth, false);
    } catch {
      setAccessToken(null);
    }
  }

  return parseResponse<T>(response);
}

export async function refreshAccessToken(): Promise<LoginResponse> {
  if (!refreshPromise) {
    refreshPromise = request<ApiEnvelope<LoginResponse>>("/api/v1/auth/refresh", { method: "POST" }, false, false)
      .then(({ data }) => {
        setAccessToken(data.access_token);
        return data;
      })
      .finally(() => {
        refreshPromise = null;
      });
  }
  return refreshPromise;
}

export interface PublicPageListingBase {
  id: string;
  page_id: string;
  title: string;
  subtitle: string;
  description: string;
  price_usd: number;
  currency: string;
  delivery_window: string;
  status: "published";
  featured: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
  published_at?: string | null;
}

export interface PublicPageListing extends PublicPageListingBase {
  facebook_page_id: string;
  page_name: string;
}

export interface PublicPageListingResult {
  pages: Array<{
    id: string;
    facebook_page_id: string;
    name: string;
    listing: PublicPageListingBase;
  }>;
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface MarketplaceOrderInput {
  buyer_note: string;
  payment_method: "wallet" | "bakong" | "card" | "usdt";
  terms_accepted: boolean;
}

export interface MarketplaceOrder {
  id: string;
  reference: string;
  listing_id: string;
  buyer_user_id: string;
  buyer_note?: string;
  amount_usd: number;
  currency: string;
  payment_method?: string;
  status: "pending_payment" | "payment_failed" | "paid" | "transfer_pending" | "completed" | "cancelled" | "expired";
  reservation_expires_at: string;
  created_at: string;
  updated_at: string;
  paid_at?: string | null;
  completed_at?: string | null;
}

export const marketplaceApi = {
  async listings(params: { search?: string; page?: number; pageSize?: number } = {}) {
    const query = new URLSearchParams({
      page: String(params.page ?? 1),
      page_size: String(params.pageSize ?? 100),
    });
    if (params.search?.trim()) query.set("search", params.search.trim());
    const { data } = await request<ApiEnvelope<PublicPageListingResult>>(
      `/api/v1/marketplace/pages?${query.toString()}`,
    );
    return data;
  },

  async listing(id: string) {
    const { data } = await request<ApiEnvelope<PublicPageListing>>(
      `/api/v1/marketplace/pages/${encodeURIComponent(id)}`,
    );
    return data;
  },

  async createOrder(listingId: string, input: MarketplaceOrderInput) {
    const { data } = await request<ApiEnvelope<MarketplaceOrder>>(
      `/api/v1/marketplace/pages/${encodeURIComponent(listingId)}/orders`,
      { method: "POST", body: JSON.stringify(input) },
      true,
    );
    return data;
  },
};

export const walletApi = {
  async wallet() {
    const { data } = await request<ApiEnvelope<Wallet>>("/api/v1/wallet", undefined, true);
    return data;
  },

  async ledger(page = 1, pageSize = 20) {
    const query = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
    const { data } = await request<ApiEnvelope<WalletLedgerResult>>(
      `/api/v1/wallet/ledger?${query.toString()}`,
      undefined,
      true,
    );
    return data;
  },

  async createTopup(amountUsd: number) {
    const { data } = await request<ApiEnvelope<WalletTopup>>(
      "/api/v1/wallet/topups",
      { method: "POST", body: JSON.stringify({ amount_usd: amountUsd }) },
      true,
    );
    return data;
  },

  async topup(id: string) {
    const { data } = await request<ApiEnvelope<WalletTopup>>(
      `/api/v1/wallet/topups/${encodeURIComponent(id)}`,
      undefined,
      true,
    );
    return data;
  },
};

export const authApi = {
  async login(email: string, password: string) {
    const { data } = await request<ApiEnvelope<LoginResponse>>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }, false);
    setAccessToken(data.access_token);
    return data;
  },

  async register(name: string, email: string, password: string) {
    const { data } = await request<ApiEnvelope<User>>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({ name, email, password }),
    }, false);
    return data;
  },

  async refresh() {
    return refreshAccessToken();
  },

  async logout() {
    const result = await request<ApiMessage>("/api/v1/auth/logout", { method: "POST" }, true);
    setAccessToken(null);
    return result;
  },

  async logoutAll() {
    const result = await request<ApiMessage>("/api/v1/auth/logout-all", { method: "POST" }, true);
    setAccessToken(null);
    return result;
  },

  async forgotPassword(email: string) {
    return request<ApiMessage>("/api/v1/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify({ email }),
    }, false);
  },

  async resetPassword(token: string, newPassword: string) {
    return request<ApiMessage>("/api/v1/auth/reset-password", {
      method: "POST",
      body: JSON.stringify({ token, new_password: newPassword }),
    }, false);
  },

  async verifyEmail(token: string) {
    return request<ApiMessage>("/api/v1/auth/verify-email", {
      method: "POST",
      body: JSON.stringify({ token }),
    }, false);
  },

  async resendVerification(email: string) {
    return request<ApiMessage>("/api/v1/auth/resend-verification", {
      method: "POST",
      body: JSON.stringify({ email }),
    }, false);
  },

  async changePassword(currentPassword: string, newPassword: string) {
    return request<ApiMessage>("/api/v1/auth/change-password", {
      method: "POST",
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    }, true);
  },

  async me() {
    const { data } = await request<ApiEnvelope<User>>("/api/v1/auth/me", undefined, true);
    return data;
  },

  async sessions() {
    const { data } = await request<ApiEnvelope<Session[]>>("/api/v1/me/sessions", undefined, true);
    return data;
  },

  async revokeSession(sessionId: string) {
    return request<ApiMessage>(`/api/v1/me/sessions/${sessionId}`, { method: "DELETE" }, true);
  },
};
