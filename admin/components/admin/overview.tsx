"use client";

import { ArrowUpRight, ShieldCheck, UserRound, UsersRound } from "lucide-react";
import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import type { AdminView } from "@/components/admin/admin-shell";
import type { User } from "@/lib/types";

export function Overview({ user, onChangeView }: Readonly<{ user: User; onChangeView: (view: AdminView) => void }>) {
  return (
    <section className="dashboard-section animate-fade-up">
      <div className="section-intro"><div><span className="eyebrow">Operations · today</span><h1>Good morning, {user.name.split(" ")[0]}.</h1><p>A focused view of your marketplace accounts and access controls.</p></div><span className="status-chip status-chip--green"><span /> Systems healthy</span></div>
      <div className="hero-card glass-panel-elevated admin-hero"><div className="hero-card__glow" /><div className="hero-card__copy"><span className="eyebrow eyebrow--bright"><ShieldCheck size={13} /> Admin control center</span><h2>Quiet control. Clear decisions.</h2><p>Review who belongs in the workspace, keep roles intentional, and make access changes without losing the human context behind every account.</p><div className="hero-card__meta"><span><UsersRound size={14} /> User directory</span><span><ShieldCheck size={14} /> Role protection</span></div></div><div className="security-score"><div className="security-score__ring"><ShieldCheck size={30} /></div><span>Protected</span></div></div>
      <div className="stats-grid"><Stat label="User directory" value="Manage" note="Search and filter accounts" icon={<UsersRound size={17} />} tone="blue" onClick={() => onChangeView("users")} /><Stat label="Role policy" value="2 roles" note="User and administrator" icon={<ShieldCheck size={17} />} tone="green" /><Stat label="Your access" value="Admin" note="Full role management" icon={<UserRound size={17} />} tone="violet" /></div>
      <div className="dashboard-grid"><div className="panel"><div className="panel__header"><div><span className="eyebrow">Access principles</span><h3>Keep the workspace considered</h3></div><span className="panel-header-icon"><ShieldCheck size={16} /></span></div><div className="health-list"><div className="health-row"><span className="health-row__check is-complete"><ShieldCheck size={13} /></span><span><strong>Least privilege by default</strong><small>New accounts begin as standard users.</small></span></div><div className="health-row"><span className="health-row__check is-complete"><ShieldCheck size={13} /></span><span><strong>Explicit role changes</strong><small>Promotions and demotions are intentional.</small></span></div><div className="health-row"><span className="health-row__check is-complete"><ShieldCheck size={13} /></span><span><strong>Self-protection</strong><small>Admins cannot remove their own access.</small></span></div></div><button className="panel-link" type="button" onClick={() => onChangeView("users")}>Open user management <ArrowUpRight size={15} /></button></div><div className="panel activity-panel"><div className="panel__header"><div><span className="eyebrow">Current operator</span><h3>Your admin identity</h3></div><Avatar name={user.name} size="md" /></div><div className="operator-card"><strong>{user.name}</strong><span>{user.email}</span><small>Administrator since {new Intl.DateTimeFormat("en", { month: "short", year: "numeric" }).format(new Date(user.created_at))}</small></div><Button variant="secondary" onClick={() => onChangeView("users")}>View all accounts <ArrowUpRight size={15} /></Button></div></div>
      <p className="privacy-line"><ShieldCheck size={14} /> Role changes are checked by the API and never trusted from the browser alone.</p>
    </section>
  );
}

function Stat({ label, value, note, icon, tone, onClick }: Readonly<{ label: string; value: string; note: string; icon: React.ReactNode; tone: string; onClick?: () => void }>) {
  const content = <><span className="stat-card__icon">{icon}</span><span className="stat-card__label">{label}</span><strong>{value}</strong><span className="stat-card__note">{note}{onClick && <ArrowUpRight size={12} />}</span></>;
  return onClick ? <button className={`stat-card stat-card--${tone} is-clickable`} type="button" onClick={onClick}>{content}</button> : <div className={`stat-card stat-card--${tone}`}>{content}</div>;
}
