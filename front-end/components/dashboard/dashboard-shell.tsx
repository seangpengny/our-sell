"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  LayoutDashboard,
  MonitorSmartphone,
  Settings2,
  ShoppingBag,
  WalletCards,
} from "lucide-react";
import { authApi, ApiError } from "@/lib/api";
import type { Session } from "@/lib/types";
import { useAuth } from "@/components/providers/auth-provider";
import { Sidebar } from "@/components/dashboard/sidebar";
import { Topbar } from "@/components/dashboard/topbar";
import { MobileDock } from "@/components/dashboard/mobile-dock";
import { OverviewSection } from "@/components/dashboard/overview-section";
import { SessionsSection } from "@/components/dashboard/sessions-section";
import { AccountSection } from "@/components/dashboard/account-section";
import { MarketplaceSection } from "@/components/dashboard/marketplace-section";
import { WalletSection } from "@/components/dashboard/wallet-section";

export type DashboardView = "overview" | "marketplace" | "wallet" | "sessions" | "account";

function initialView(): DashboardView {
  if (typeof window === "undefined") return "overview";
  const requested = new URLSearchParams(window.location.search).get("view");
  return requested === "marketplace" || requested === "wallet" || requested === "sessions" || requested === "account"
    ? requested
    : "overview";
}

const views: Record<
  DashboardView,
  { label: string; icon: typeof LayoutDashboard }
> = {
  overview: { label: "Overview", icon: LayoutDashboard },
  marketplace: { label: "Marketplace", icon: ShoppingBag },
  wallet: { label: "Wallet", icon: WalletCards },
  sessions: { label: "Sessions", icon: MonitorSmartphone },
  account: { label: "Account", icon: Settings2 },
};

export function DashboardShell() {
  const router = useRouter();
  const { user, refreshUser } = useAuth();
  const [activeView, setActiveView] = useState<DashboardView>(initialView);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [isLoadingSessions, setIsLoadingSessions] = useState(true);
  const [sessionError, setSessionError] = useState("");
  const [toast, setToast] = useState("");

  async function loadSessions() {
    setIsLoadingSessions(true);
    setSessionError("");
    try {
      setSessions(await authApi.sessions());
    } catch (error) {
      if (!(error instanceof ApiError && error.status === 401))
        setSessionError("We couldn’t load your sessions. Please try again.");
    } finally {
      setIsLoadingSessions(false);
    }
  }

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      void loadSessions();
    }, 0);
    return () => window.clearTimeout(timeout);
  }, []);
  useEffect(() => {
    if (!toast) return;
    const timeout = window.setTimeout(() => setToast(""), 3600);
    return () => window.clearTimeout(timeout);
  }, [toast]);

  async function revokeSession(id: string) {
    try {
      await authApi.revokeSession(id);
      setSessions((current) => current.filter((session) => session.id !== id));
      setToast("Session revoked.");
    } catch {
      setToast("That session could not be revoked. Try again.");
    }
  }

  async function revokeAll() {
    try {
      await authApi.logoutAll();
      router.replace("/sign-in");
    } catch {
      setToast("We couldn’t revoke all sessions. Try again.");
    }
  }

  if (!user) return null;
  const view = views[activeView];
  return (
    <main className="app-page">
      <div className="app-page__ambient app-page__ambient--one" />
      <div className="app-page__ambient app-page__ambient--two" />
      <Sidebar
        activeView={activeView}
        onChange={setActiveView}
        sessionsCount={sessions.length}
      />
      <div className="app-content">
        <Topbar view={view.label} user={user} />
        <div className="app-content__body">
          {activeView === "overview" && (
            <OverviewSection
              error={sessionError}
              user={user}
              sessions={sessions}
              loading={isLoadingSessions}
              onChangeView={setActiveView}
              onRefresh={loadSessions}
              onResendVerification={async () => {
                try {
                  await authApi.resendVerification(user.email);
                  setToast(
                    "If your account needs verification, a fresh email is on the way.",
                  );
                } catch {
                  setToast("We couldn’t resend the verification email.");
                }
              }}
            />
          )}
          {activeView === "marketplace" && <MarketplaceSection />}
          {activeView === "wallet" && <WalletSection onToast={setToast} />}
          {activeView === "sessions" && (
            <SessionsSection
              error={sessionError}
              sessions={sessions}
              loading={isLoadingSessions}
              onRefresh={loadSessions}
              onRevoke={revokeSession}
              onRevokeAll={revokeAll}
            />
          )}
          {activeView === "account" && (
            <AccountSection
              user={user}
              onProfileRefresh={refreshUser}
              onToast={setToast}
            />
          )}
        </div>
      </div>
      <MobileDock
        activeView={activeView}
        onChange={setActiveView}
        sessionsCount={sessions.length}
      />
      {toast && (
        <div className="toast animate-spring-in" role="status">
          <span className="toast__dot" />
          {toast}
        </div>
      )}
    </main>
  );
}
