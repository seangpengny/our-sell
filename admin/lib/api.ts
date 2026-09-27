import type {
  AdminUserList,
  ApiEnvelope,
  ApiErrorPayload,
  ApiMessage,
  FacebookConnectResponse,
  FacebookConnection,
  FacebookListingInput,
  FacebookPageInventory,
  FacebookPageListing,
  FacebookPage,
  LoginResponse,
  AdminWalletTopupResult,
  User,
  UserRole,
} from "@/lib/types";

const API_URL = (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080").replace(/\/$/, "");

let accessToken: string | null = null;
let refreshPromise: Promise<LoginResponse> | null = null;

export class ApiError extends Error {
  readonly code: string;
  readonly fields: Record<string, string>;
  readonly status: number;

  constructor(
    message: string,
    status: number,
    code = "REQUEST_FAILED",
    fields: Record<string, string> = {},
  ) {
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

async function parseResponse<T>(response: Response): Promise<T> {
  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json")
    ? ((await response.json()) as ApiErrorPayload & T)
    : null;

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

async function request<T>(
  path: string,
  init: RequestInit = {},
  requiresAuth = false,
  allowRefresh = true,
): Promise<T> {
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

async function refreshAccessToken(): Promise<LoginResponse> {
  if (!refreshPromise) {
    refreshPromise = request<ApiEnvelope<LoginResponse>>(
      "/api/v1/auth/refresh",
      { method: "POST" },
      false,
      false,
    )
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

export const authApi = {
  async login(email: string, password: string) {
    const { data } = await request<ApiEnvelope<LoginResponse>>(
      "/api/v1/auth/login",
      { method: "POST", body: JSON.stringify({ email, password }) },
    );
    setAccessToken(data.access_token);
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

  async users(params: { search: string; role: UserRole | "all"; page: number; pageSize: number }) {
    const query = new URLSearchParams({
      page: String(params.page),
      page_size: String(params.pageSize),
    });
    if (params.search.trim()) query.set("search", params.search.trim());
    if (params.role !== "all") query.set("role", params.role);
    const { data } = await request<ApiEnvelope<AdminUserList>>(
      `/api/v1/admin/users?${query.toString()}`,
      undefined,
      true,
    );
    return data;
  },

  async updateRole(userId: string, role: UserRole) {
    const { data } = await request<ApiEnvelope<User>>(
      `/api/v1/admin/users/${userId}/role`,
      { method: "PATCH", body: JSON.stringify({ role }) },
      true,
    );
    return data;
  },

  async startFacebookConnect() {
    const { data } = await request<ApiEnvelope<FacebookConnectResponse>>(
      "/api/v1/auth/facebook/connect",
      { method: "POST" },
      true,
    );
    return data;
  },

  async facebookConnections() {
    const { data } = await request<ApiEnvelope<FacebookConnection[]>>(
      "/api/v1/facebook/connections",
      undefined,
      true,
    );
    return data;
  },

  async facebookPages(connectionId: string) {
    const query = new URLSearchParams({ connection_id: connectionId });
    const { data } = await request<ApiEnvelope<FacebookPage[]>>(
      `/api/v1/facebook/pages?${query.toString()}`,
      undefined,
      true,
    );
    return data;
  },

  async facebookPageInventory(params: { connectionId?: string; search?: string; page?: number; pageSize?: number }) {
    const query = new URLSearchParams({ page: String(params.page ?? 1), page_size: String(params.pageSize ?? 25) });
    if (params.connectionId) query.set("connection_id", params.connectionId);
    if (params.search?.trim()) query.set("search", params.search.trim());
    const { data } = await request<ApiEnvelope<FacebookPageInventory>>(
      `/api/v1/facebook/page-inventory?${query.toString()}`,
      undefined,
      true,
    );
    return data;
  },

  async syncFacebookConnection(connectionId: string) {
    const { data } = await request<ApiEnvelope<{ page_count: number }>>(
      `/api/v1/facebook/connections/${encodeURIComponent(connectionId)}/sync`,
      { method: "POST" },
      true,
    );
    return data;
  },

  async saveFacebookListing(pageId: string, input: FacebookListingInput) {
    const { data } = await request<ApiEnvelope<FacebookPageListing>>(
      `/api/v1/facebook/pages/${encodeURIComponent(pageId)}/listing`,
      { method: "PUT", body: JSON.stringify(input) },
      true,
    );
    return data;
  },

  async disconnectFacebook(connectionId: string) {
    return request<ApiMessage>(
      `/api/v1/facebook/connections/${encodeURIComponent(connectionId)}`,
      { method: "DELETE" },
      true,
    );
  },

  async walletTopups(params: { search: string; status: string; page: number; pageSize: number }) {
    const query = new URLSearchParams({
      page: String(params.page),
      page_size: String(params.pageSize),
    });
    if (params.search.trim()) query.set("search", params.search.trim());
    if (params.status) query.set("status", params.status);
    const { data } = await request<ApiEnvelope<AdminWalletTopupResult>>(
      `/api/v1/admin/wallet/topups?${query.toString()}`,
      undefined,
      true,
    );
    return data;
  },

  async recheckWalletTopup(topupId: string) {
    const { data } = await request<ApiEnvelope<AdminWalletTopupResult["topups"][number]>>(
      `/api/v1/admin/wallet/topups/${encodeURIComponent(topupId)}/recheck`,
      { method: "POST" },
      true,
    );
    return data;
  },
};
