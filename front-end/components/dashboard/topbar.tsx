"use client";

import { Bell, ChevronDown } from "lucide-react";
import { useState } from "react";
import { useAuth } from "@/components/providers/auth-provider";
import { ThemeToggle } from "@/components/providers/theme-provider";
import { Avatar } from "@/components/ui/avatar";
import type { User } from "@/lib/types";

export function Topbar({ view, user }: Readonly<{ view: string; user: User }>) {
  const { logout } = useAuth();
  const [open, setOpen] = useState(false);
  return <header className="topbar"><div className="topbar__mobile-title"><span className="topbar__kicker">Workspace</span><strong>{view}</strong></div><div className="topbar__status"><span className="live-dot" /> All systems operational</div><div className="topbar__actions"><ThemeToggle compact /><button className="icon-button topbar__bell" type="button" aria-label="Notifications"><Bell size={17} /><span /></button><div className="profile-menu"><button type="button" className="profile-menu__trigger" onClick={() => setOpen((current) => !current)} aria-expanded={open}><Avatar name={user.name} size="sm" /><span className="profile-menu__name">{user.name}</span><ChevronDown size={14} /></button>{open && <div className="profile-menu__popover"><span className="profile-menu__email">{user.email}</span><button type="button" onClick={() => void logout()}>Sign out</button></div>}</div></div></header>;
}
