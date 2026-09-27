"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { usePathname } from "next/navigation";
import { authApi, ApiError, setAccessToken } from "@/lib/api";
import type { LoginResponse, User } from "@/lib/types";

type AuthStatus = "loading" | "authenticated" | "unauthenticated";
type AuthContextValue = {
  status: AuthStatus;
  user: User | null;
  login: (email: string, password: string) => Promise<LoginResponse>;
  logout: () => Promise<void>;
  logoutAll: () => Promise<void>;
  refreshUser: () => Promise<User | null>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: Readonly<{ children: React.ReactNode }>) {
  const pathname = usePathname();
  const [status, setStatus] = useState<AuthStatus>("loading");
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    const needsAuthBootstrap = pathname.startsWith("/app") || /\/marketplace\/[^/]+\/checkout$/.test(pathname);
    if (!needsAuthBootstrap) {
      setAccessToken(null);
      const timeout = window.setTimeout(() => {
        setUser(null);
        setStatus("unauthenticated");
      }, 0);
      return () => window.clearTimeout(timeout);
    }
    let active = true;
    authApi.refresh()
      .then((result) => {
        if (!active) return;
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

  const logoutAll = useCallback(async () => {
    try { await authApi.logoutAll(); } finally {
      setAccessToken(null);
      setUser(null);
      setStatus("unauthenticated");
    }
  }, []);

  const refreshUser = useCallback(async () => {
    try {
      const nextUser = await authApi.me();
      setUser(nextUser);
      return nextUser;
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        setUser(null);
        setStatus("unauthenticated");
      }
      return null;
    }
  }, []);

  const value = useMemo(() => ({ status, user, login, logout, logoutAll, refreshUser }), [status, user, login, logout, logoutAll, refreshUser]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
