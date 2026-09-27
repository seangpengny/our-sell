"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { usePathname } from "next/navigation";
import { ApiError, authApi, setAccessToken } from "@/lib/api";
import type { LoginResponse, User } from "@/lib/types";

type AuthStatus = "loading" | "authenticated" | "unauthenticated" | "forbidden";
type AuthContextValue = {
  status: AuthStatus;
  user: User | null;
  login: (email: string, password: string) => Promise<LoginResponse>;
  logout: () => Promise<void>;
};
const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: Readonly<{ children: React.ReactNode }>) {
  const pathname = usePathname();
  const [status, setStatus] = useState<AuthStatus>("loading");
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    if (pathname === "/sign-in") {
      setAccessToken(null);
      const timeout = window.setTimeout(() => {
        setUser(null);
        setStatus("unauthenticated");
      }, 0);
      return () => window.clearTimeout(timeout);
    }
    let active = true;
    authApi.refresh()
      .then(async (result) => {
        if (!active) return;
        if (result.user.role !== "admin") {
          try { await authApi.logout(); } catch { /* the access token is still cleared below */ }
          setAccessToken(null);
          setUser(null);
          setStatus("forbidden");
          return;
        }
        setUser(result.user);
        setStatus("authenticated");
      })
      .catch(() => {
        if (!active) return;
        setAccessToken(null);
        setUser(null);
        setStatus("unauthenticated");
      });
    return () => { active = false; };
  }, [pathname]);

  const login = useCallback(async (email: string, password: string) => {
    const result = await authApi.login(email, password);
    if (result.user.role !== "admin") {
      try { await authApi.logout(); } catch { /* access token is cleared below */ }
      setAccessToken(null);
      setUser(null);
      setStatus("forbidden");
      throw new ApiError("This account does not have administrator access.", 403, "FORBIDDEN");
    }
    setUser(result.user);
    setStatus("authenticated");
    return result;
  }, []);

  const logout = useCallback(async () => {
    try { await authApi.logout(); } finally {
      setAccessToken(null);
      setUser(null);
      setStatus("unauthenticated");
    }
  }, []);

  const value = useMemo(() => ({ status, user, login, logout }), [status, user, login, logout]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
