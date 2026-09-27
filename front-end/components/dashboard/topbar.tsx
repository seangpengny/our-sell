"use client";

import { Bell, ChevronDown } from "lucide-react";
import { useEffect, useRef } from "react";
import { useAuth } from "@/components/providers/auth-provider";
import { ThemeToggle } from "@/components/providers/theme-provider";
import { Avatar } from "@/components/ui/avatar";
import type { User } from "@/lib/types";

export function Topbar({ view, user }: Readonly<{ view: string; user: User }>) {
  const { logout } = useAuth();
  const actionsRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    function dismiss(event: PointerEvent | KeyboardEvent) {
      const open =
        actionsRef.current?.querySelector<HTMLDetailsElement>("details[open]");
      if (!open) return;
      if (event instanceof KeyboardEvent) {
        if (event.key !== "Escape") return;
        open.open = false;
        open.querySelector("summary")?.focus();
      } else if (!open.contains(event.target as Node)) open.open = false;
    }
    document.addEventListener("pointerdown", dismiss);
    document.addEventListener("keydown", dismiss);
    return () => {
      document.removeEventListener("pointerdown", dismiss);
      document.removeEventListener("keydown", dismiss);
    };
  }, []);
  return (
    <header className="topbar glass-panel-elevated">
      <div className="topbar__mobile-title">
        <span className="topbar__kicker">Your workspace</span>
        <strong>{view}</strong>
      </div>
      <div className="topbar__status">
        Workspace <span>/</span> <strong>{view}</strong>
      </div>
      <div className="topbar__actions" ref={actionsRef}>
        <ThemeToggle compact />
        <details className="profile-menu" name="topbar-menu">
          <summary className="icon-button" aria-label="Notifications">
            <Bell size={18} />
          </summary>
          <div className="profile-menu__popover notification-popover">
            <Bell size={24} />
            <strong>You’re all caught up</strong>
            <p>No notifications to show right now.</p>
          </div>
        </details>
        <details className="profile-menu" name="topbar-menu">
          <summary
            className="profile-menu__trigger"
            aria-label={`Account menu for ${user.name}`}
          >
            <Avatar name={user.name} size="sm" />
            <span className="profile-menu__name">{user.name}</span>
            <ChevronDown size={16} />
          </summary>
          <div className="profile-menu__popover">
            <span className="profile-menu__email">{user.email}</span>
            <button type="button" onClick={() => void logout()}>
              Sign out
            </button>
          </div>
        </details>
      </div>
    </header>
  );
}
