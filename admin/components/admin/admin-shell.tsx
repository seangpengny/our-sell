"use client";

import { useEffect, useState } from "react";
import { LayoutDashboard, Link2, ShieldCheck, Users, WalletCards } from "lucide-react";
import { useAuth } from "@/components/providers/auth-provider";
import { Sidebar } from "@/components/admin/sidebar";
import { Topbar } from "@/components/admin/topbar";
import { UserManagement } from "@/components/admin/user-management";
import { Overview } from "@/components/admin/overview";
import { FacebookConnections } from "@/components/admin/facebook-connections";
import { FacebookPages } from "@/components/admin/facebook-pages";
import { WalletPayments } from "@/components/admin/wallet-payments";

export type AdminView = "overview" | "users" | "facebook" | "facebook-pages" | "wallet-payments";

const views: Record<AdminView, { label: string; icon: typeof LayoutDashboard }> = {
  overview: { label: "Overview", icon: LayoutDashboard },
  users: { label: "Users", icon: Users },
  facebook: { label: "Facebook accounts", icon: Link2 },
  "facebook-pages": { label: "Page management", icon: ShieldCheck },
  "wallet-payments": { label: "Wallet payments", icon: WalletCards },
};

export function AdminShell() {
  const { user } = useAuth();
  const [activeView, setActiveView] = useState<AdminView>("overview");
  const [toast, setToast] = useState("");

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      const params = new URLSearchParams(window.location.search);
      if (params.get("view") === "facebook") setActiveView("facebook");
      if (params.get("view") === "facebook-pages") setActiveView("facebook-pages");
      if (params.get("view") === "wallet-payments") setActiveView("wallet-payments");
      const status = params.get("status");
      if (status === "connected") setToast("Facebook account connected. Loading Pages…");
      if (status === "cancelled") setToast("Facebook connection was cancelled.");
      if (status === "error" || status === "invalid_state") setToast("Facebook could not be connected. Please try again.");
      if (status) window.history.replaceState({}, "", window.location.pathname);
    }, 0);
    return () => window.clearTimeout(timeout);
  }, []);

  useEffect(() => {
    if (!toast) return;
    const timeout = window.setTimeout(() => setToast(""), 5000);
    return () => window.clearTimeout(timeout);
  }, [toast]);
  if (!user) return null;
  const view = views[activeView];

  return (
    <main className="app-page">
      <div className="app-page__ambient app-page__ambient--one" />
      <div className="app-page__ambient app-page__ambient--two" />
      <Sidebar activeView={activeView} onChange={setActiveView} />
      <div className="app-content">
        <Topbar view={view.label} user={user} />
        <div className="app-content__body">
          {activeView === "overview" && <Overview user={user} onChangeView={setActiveView} />}
          {activeView === "users" && <UserManagement onToast={setToast} />}
          {activeView === "facebook" && <FacebookConnections onToast={setToast} />}
          {activeView === "facebook-pages" && <FacebookPages onOpenAccounts={() => setActiveView("facebook")} onToast={setToast} />}
          {activeView === "wallet-payments" && <WalletPayments />}
        </div>
      </div>
      <nav className="mobile-dock" aria-label="Mobile navigation">
        {(Object.entries(views) as Array<[AdminView, (typeof views)[AdminView]]>).map(([id, item]) => {
          const Icon = item.icon;
          const label = id === "facebook" ? "Accounts" : id === "facebook-pages" ? "Pages" : item.label;
          return <button key={id} type="button" className={activeView === id ? "is-active" : ""} onClick={() => setActiveView(id)}><Icon size={18} /><span>{label}</span></button>;
        })}
      </nav>
      {toast && <div className="toast animate-spring-in" role="status"><span className="toast__dot" />{toast}</div>}
    </main>
  );
}
