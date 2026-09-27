"use client";

import {
  CircleHelp,
  LayoutDashboard,
  LogOut,
  MonitorSmartphone,
  Settings2,
  ShoppingBag,
  WalletCards,
} from "lucide-react";
import { Logo } from "@/components/ui/logo";
import { Avatar } from "@/components/ui/avatar";
import { ThemeToggle } from "@/components/providers/theme-provider";
import { useAuth } from "@/components/providers/auth-provider";
import type { DashboardView } from "@/components/dashboard/dashboard-shell";

const navItems: Array<{
  id: DashboardView;
  label: string;
  icon: typeof LayoutDashboard;
}> = [
  { id: "overview", label: "Overview", icon: LayoutDashboard },
  { id: "marketplace", label: "Marketplace", icon: ShoppingBag },
  { id: "wallet", label: "Wallet", icon: WalletCards },
  { id: "sessions", label: "Sessions", icon: MonitorSmartphone },
  { id: "account", label: "Account", icon: Settings2 },
];

export function Sidebar({
  activeView,
  onChange,
  sessionsCount,
}: Readonly<{
  activeView: DashboardView;
  onChange: (view: DashboardView) => void;
  sessionsCount: number;
}>) {
  const { user, logout } = useAuth();
  if (!user) return null;
  return (
    <aside className="sidebar">
      <div className="sidebar__top">
        <Logo href="/app" />
        <span className="sidebar__version">v1.0</span>
      </div>
      <div className="sidebar__workspace">
        <span className="sidebar__workspace-label">Your space</span>
        <div className="workspace-switcher">
          <Avatar name={user.name} size="sm" />
          <span className="workspace-switcher__text">
            <strong>{user.name}</strong>
            <small>
              {user.role === "admin" ? "Admin account" : "Personal account"}
            </small>
          </span>
          <span className="workspace-switcher__status" />
        </div>
      </div>
      <nav className="sidebar__nav" aria-label="Workspace navigation">
        {navItems.map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            className={`side-nav-item ${activeView === id ? "is-active" : ""}`}
            type="button"
            aria-current={activeView === id ? "page" : undefined}
            onClick={() => onChange(id)}
          >
            <Icon size={17} />
            <span>{label}</span>
            {id === "sessions" && sessionsCount > 0 && <b>{sessionsCount}</b>}
          </button>
        ))}
      </nav>
      <div className="sidebar__bottom">
        <div className="sidebar__help">
          <CircleHelp size={16} />
          <span>
            <strong>Need a hand?</strong>
            <small>We’re here when you need us.</small>
          </span>
        </div>
        <ThemeToggle />
        <button
          type="button"
          className="side-logout"
          onClick={() => void logout()}
        >
          <LogOut size={16} /> Sign out
        </button>
      </div>
    </aside>
  );
}
