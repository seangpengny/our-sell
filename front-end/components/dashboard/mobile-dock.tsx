"use client";

import { LayoutDashboard, MonitorSmartphone, Settings2 } from "lucide-react";
import type { DashboardView } from "@/components/dashboard/dashboard-shell";

const items: Array<{ id: DashboardView; short: string; label: string; icon: typeof LayoutDashboard }> = [
  { id: "overview", short: "Home", label: "Overview", icon: LayoutDashboard },
  { id: "sessions", short: "Access", label: "Sessions", icon: MonitorSmartphone },
  { id: "account", short: "You", label: "Account", icon: Settings2 },
];

export function MobileDock({ activeView, onChange, sessionsCount }: Readonly<{ activeView: DashboardView; onChange: (view: DashboardView) => void; sessionsCount: number }>) {
  return <div className="mobile-dock"><nav>{items.map(({ id, short, label, icon: Icon }) => <button key={id} type="button" className={activeView === id ? "is-active" : ""} onClick={() => onChange(id)}><span className="mobile-dock__icon"><Icon size={17} />{id === "sessions" && sessionsCount > 0 && <b>{sessionsCount}</b>}</span><span className="mobile-dock__short">{short}</span><span className="mobile-dock__long">{label}</span></button>)}</nav></div>;
}
