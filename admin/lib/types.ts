export type UserRole = "user" | "admin";

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  email_verified_at?: string | null;
  created_at: string;
}

export interface LoginResponse {
  access_token: string;
  token_type: "Bearer" | string;
  expires_in: number;
  user: User;
}

export interface AdminUserList {
  users: User[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ApiErrorPayload {
  error?: {
    code?: string;
    message?: string;
    fields?: Record<string, string>;
  };
}

export interface ApiEnvelope<T> {
  data: T;
}

export interface ApiMessage {
  message: string;
}

export type WalletTopupStatus = "pending_payment" | "confirmed" | "failed" | "expired" | "reversed";

export interface AdminWalletTopup {
  id: string;
  wallet_id: string;
  user_id: string;
  user_name: string;
  user_email: string;
  reference: string;
  amount_usd: number;
  currency: string;
  qr_md5: string;
  provider_transaction_id?: string;
  status: WalletTopupStatus;
  failure_reason?: string;
  resolution_note?: string;
  expires_at: string;
  paid_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface AdminWalletTopupResult {
  topups: AdminWalletTopup[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface FacebookConnection {
  id: string;
  facebook_account_id: string;
  name: string;
  connected_at: string;
  token_expires_at?: string | null;
}

export interface FacebookPage {
  id: string;
  name: string;
}

export type FacebookListingStatus = "draft" | "pending_review" | "published" | "unlisted" | "reserved" | "sold" | "transfer_pending" | "transferred" | "cancelled";

export interface FacebookPageListing {
  id: string;
  page_id: string;
  title: string;
  subtitle: string;
  description: string;
  price_usd: number;
  currency: string;
  delivery_window: string;
  status: FacebookListingStatus;
  featured: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
  published_at?: string | null;
}

export interface FacebookPageInventoryItem {
  id: string;
  facebook_page_id: string;
  name: string;
  connection_id: string;
  connection_name: string;
  last_synced_at?: string | null;
  listing?: FacebookPageListing | null;
}

export interface FacebookPageInventory {
  pages: FacebookPageInventoryItem[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface FacebookListingInput {
  title: string;
  subtitle: string;
  description: string;
  price_usd: number;
  currency: string;
  delivery_window: string;
  status: FacebookListingStatus;
  featured: boolean;
  sort_order: number;
}

export interface FacebookConnectResponse {
  authorization_url: string;
}
