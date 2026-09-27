"use client";

import { ChevronDown, CircleHelp, LayoutDashboard, Link2, LogOut, ShieldCheck, Users, WalletCards } from "lucide-react";
import { useAuth } from "@/components/providers/auth-provider";
import { ThemeToggle } from "@/components/providers/theme-provider";
import { Avatar } from "@/components/ui/avatar";
import { Logo } from "@/components/ui/logo";
import type { AdminView } from "@/components/admin/admin-shell";

const navItems: Array<{ id: AdminView; label: string; icon: typeof LayoutDashboard }> = [
  { id: "overview", label: "Overview", icon: LayoutDashboard },
  { id: "users", label: "User management", icon: Users },
];

export function Sidebar({ activeView, onChange }: Readonly<{ activeView: AdminView; onChange: (view: AdminView) => void }>) {
  const { user, logout } = useAuth();
  if (!user) return null;
  return (
    <aside className="sidebar">
      <div className="sidebar__top"><Logo /><span className="sidebar__version">admin</span></div>
      <div className="sidebar__workspace">
        <span className="sidebar__workspace-label">Operations</span>
        <div className="workspace-switcher"><Avatar name={user.name} size="sm" /><span className="workspace-switcher__text"><strong>{user.name}</strong><small>Administrator</small></span><span className="workspace-switcher__status" /></div>
      </div>
      <nav className="sidebar__nav" aria-label="Admin navigation">
        {navItems.map(({ id, label, icon: Icon }) => <button key={id} type="button" className={`side-nav-item ${activeView === id ? "is-active" : ""}`} aria-current={activeView === id ? "page" : undefined} onClick={() => onChange(id)}><Icon size={17} /><span>{label}</span></button>)}
        <div className="side-nav-group">
          <button type="button" className={`side-nav-item side-nav-item--parent ${activeView === "facebook" || activeView === "facebook-pages" ? "is-active" : ""}`} aria-expanded="true" onClick={() => onChange("facebook")}>
            <Link2 size={17} /><span>Facebook accounts</span><ChevronDown size={15} className="side-nav-item__chevron" />
          </button>
          <div className="side-nav-subnav">
            <button type="button" className={`side-nav-subitem ${activeView === "facebook" ? "is-active" : ""}`} aria-current={activeView === "facebook" ? "page" : undefined} onClick={() => onChange("facebook")}><Link2 size={15} /><span>Accounts</span></button>
            <button type="button" className={`side-nav-subitem ${activeView === "facebook-pages" ? "is-active" : ""}`} aria-current={activeView === "facebook-pages" ? "page" : undefined} onClick={() => onChange("facebook-pages")}><ShieldCheck size={15} /><span>Page management</span></button>
          </div>
        </div>
        <button type="button" className={`side-nav-item ${activeView === "wallet-payments" ? "is-active" : ""}`} aria-current={activeView === "wallet-payments" ? "page" : undefined} onClick={() => onChange("wallet-payments")}><WalletCards size={17} /><span>Wallet payments</span></button>
      </nav>
      <div className="sidebar__bottom">
        <div className="sidebar__help"><CircleHelp size={16} /><span><strong>Admin support</strong><small>Access changes are logged safely.</small></span></div>
        <ThemeToggle />
        <button type="button" className="side-logout" onClick={() => void logout()}><LogOut size={16} /> Sign out</button>
      </div>
    </aside>
  );
}
