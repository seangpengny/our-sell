export type UserRole = "user" | "admin";

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  email_verified_at?: string | null;
  created_at: string;
}

export interface Session {
  id: string;
  device: string;
  ip: string;
  expires_at: string;
  last_used_at: string;
  created_at: string;
}

export interface Wallet {
  id: string;
  user_id: string;
  currency: "USD" | string;
  available_amount: number;
  held_amount: number;
  created_at: string;
  updated_at: string;
}

export interface WalletLedgerEntry {
  id: string;
  wallet_id: string;
  user_id: string;
  entry_type: string;
  reference: string;
  available_delta_usd: number;
  held_delta_usd: number;
  available_balance_usd: number;
  held_balance_usd: number;
  description: string;
  created_at: string;
}

export interface WalletLedgerResult {
  entries: WalletLedgerEntry[];
  page: number;
  page_size: number;
  has_more: boolean;
}

export interface WalletTopup {
  id: string;
  wallet_id: string;
  user_id: string;
  reference: string;
  amount_usd: number;
  currency: string;
  qr_payload: string;
  qr_md5: string;
  provider_transaction_id?: string;
  status: "pending_payment" | "confirmed" | "failed" | "expired" | "reversed";
  failure_reason?: string;
  resolution_note?: string;
  expires_at: string;
  paid_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface LoginResponse {
  access_token: string;
  token_type: "Bearer" | string;
  expires_in: number;
  user: User;
}

export interface ApiErrorPayload {
  error?: {
    code?: string;
    message?: string;
    fields?: Record<string, string>;
  };
}

export interface ApiMessage {
  message: string;
}

export interface ApiEnvelope<T> {
  data: T;
}

export type ProductKind = "facebook_page" | "youtube_channel" | "instagram_account" | "service";
export type ProductStatus = "available" | "reserved" | "sold" | "coming_soon";

export interface ProductFeature {
  label: string;
  icon: "key" | "flame" | "clock" | "shield" | "sparkles";
}

export interface ProductStat {
  label: string;
  value: string;
}

/**
 * The catalog shape is intentionally product-agnostic. The admin project can
 * later manage these fields without changing the storefront component.
 */
export interface Product {
  id: string;
  kind: ProductKind;
  title: string;
  subtitle: string;
  description: string;
  deliveryWindow: string;
  priceUsd: number;
  currency: string;
  status: ProductStatus;
  published: boolean;
  featured: boolean;
  sortOrder: number;
  initials: string;
  accent: "blue" | "peach" | "lavender" | "mint";
  stats: ProductStat[];
  features: ProductFeature[];
}
