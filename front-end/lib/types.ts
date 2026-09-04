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
