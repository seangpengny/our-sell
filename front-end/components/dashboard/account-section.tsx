"use client";

import { FormEvent, useState } from "react";
import { ArrowRight, Check, KeyRound, Mail, ShieldCheck, UserRound } from "lucide-react";
import type { User } from "@/lib/types";
import { authApi, ApiError } from "@/lib/api";
import { formatDate } from "@/lib/format";
import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { PasswordField } from "@/components/ui/field";
import { Notice } from "@/components/ui/notice";

export function AccountSection({ user, onProfileRefresh, onToast }: Readonly<{ user: User; onProfileRefresh: () => Promise<User | null>; onToast: (message: string) => void }>) {
  const [current, setCurrent] = useState(""); const [next, setNext] = useState(""); const [loading, setLoading] = useState(false); const [error, setError] = useState(""); const [changed, setChanged] = useState(false);
  async function submit(event: FormEvent) { event.preventDefault(); setError(""); setChanged(false); setLoading(true); try { await authApi.changePassword(current, next); setCurrent(""); setNext(""); setChanged(true); onToast("Password changed successfully."); await onProfileRefresh(); } catch (cause) { setError(cause instanceof ApiError ? cause.message : "We couldn’t change your password."); } finally { setLoading(false); } }
  return <section className="dashboard-section animate-fade-up"><div className="section-intro"><div><span className="eyebrow">Personal settings</span><h1>Your account</h1><p>Manage the details that keep your workspace yours.</p></div></div><div className="account-grid"><div className="panel profile-panel"><div className="panel__header"><div><span className="eyebrow">Profile</span><h3>About you</h3></div><span className="status-chip status-chip--green"><span /> Active</span></div><div className="profile-hero"><Avatar name={user.name} size="lg" /><div><h2>{user.name}</h2><p>{user.role === "admin" ? "Administrator" : "Member"} · since {formatDate(user.created_at, { month: "long", year: "numeric" })}</p></div></div><div className="profile-details"><Detail icon={<Mail size={16} />} label="Email address" value={user.email} verified={Boolean(user.email_verified_at)} /><Detail icon={<UserRound size={16} />} label="Account ID" value={`${user.id.slice(0, 8)}…`} /><Detail icon={<ShieldCheck size={16} />} label="Role" value={user.role === "admin" ? "Administrator" : "Standard member"} /></div></div><div className="panel password-panel"><div className="panel__header"><div><span className="eyebrow">Security</span><h3>Change password</h3></div><span className="panel-header-icon"><KeyRound size={17} /></span></div><p className="panel-description">Choose a strong password of at least 12 characters. Other active sessions will be revoked after you change it.</p><form className="settings-form" onSubmit={submit}>{error && <Notice>{error}</Notice>}{changed && <Notice variant="success">Your password has been changed.</Notice>}<PasswordField label="Current password" autoComplete="current-password" value={current} onChange={(event) => setCurrent(event.target.value)} required /><PasswordField label="New password" autoComplete="new-password" value={next} onChange={(event) => setNext(event.target.value)} minLength={12} maxLength={128} required /><Button type="submit" loading={loading} className="button--wide">Update password <ArrowRight size={16} /></Button></form></div></div><p className="privacy-line"><ShieldCheck size={14} /> We never store raw passwords. Your password is hashed before it reaches the database.</p></section>;
}

function Detail({ icon, label, value, verified }: Readonly<{ icon: React.ReactNode; label: string; value: string; verified?: boolean }>) {
  return <div className="detail-row"><span className="detail-row__icon">{icon}</span><span><small>{label}</small><strong>{value}</strong></span>{verified && <Check className="detail-row__verified" size={15} />}</div>;
}
