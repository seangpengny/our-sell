"use client";

import Link from "next/link";
import { ArrowRight, Check, KeyRound } from "lucide-react";
import { FormEvent, useState } from "react";
import { AuthShell, AuthFooterLink } from "@/components/auth/auth-shell";
import { Button } from "@/components/ui/button";
import { Field, PasswordField } from "@/components/ui/field";
import { Notice } from "@/components/ui/notice";
import { ApiError, authApi } from "@/lib/api";

export function VerifyEmailForm() {
  const [token, setToken] = useState(() => typeof window === "undefined" ? "" : new URLSearchParams(window.location.search).get("token") ?? ""); const [status, setStatus] = useState<"idle" | "done">("idle"); const [error, setError] = useState(""); const [loading, setLoading] = useState(false);
  async function submit(event: FormEvent) { event.preventDefault(); setError(""); setLoading(true); try { await authApi.verifyEmail(token); setStatus("done"); } catch (cause) { setError(cause instanceof ApiError ? cause.message : "This link is no longer valid. Request a new one and try again."); } finally { setLoading(false); } }
  return <AuthShell eyebrow="Email verification" title={status === "done" ? "You’re verified." : "One last small step."} description={status === "done" ? "Your email address is confirmed and your account is ready." : "Paste the verification token from your email to confirm your address."} footer={<AuthFooterLink prompt="Need another link?" label="Request verification" href="/sign-in" />}>
    {status === "done" ? <div className="success-state"><span className="success-state__icon"><Check size={27} /></span><p>Thanks for keeping your account details up to date.</p><Link className="button button--primary button--wide" href="/sign-in">Continue to sign in <ArrowRight size={17} /></Link></div> : <form className="auth-form" onSubmit={submit}>{error && <Notice>{error}</Notice>}<Field label="Verification token" autoComplete="one-time-code" placeholder="Paste your token" value={token} onChange={(event) => setToken(event.target.value)} required /><Button type="submit" loading={loading} className="button--wide">Verify email <ArrowRight size={17} /></Button><div className="form-note"><KeyRound size={14} /> Tokens are single-use and expire automatically.</div></form>}
  </AuthShell>;
}

export function ResetPasswordForm() {
  const [token, setToken] = useState(() => typeof window === "undefined" ? "" : new URLSearchParams(window.location.search).get("token") ?? ""); const [password, setPassword] = useState(""); const [status, setStatus] = useState<"idle" | "done">("idle"); const [error, setError] = useState(""); const [loading, setLoading] = useState(false);
  async function submit(event: FormEvent) { event.preventDefault(); setError(""); setLoading(true); try { await authApi.resetPassword(token, password); setStatus("done"); } catch (cause) { setError(cause instanceof ApiError ? cause.message : "This reset link is no longer valid. Request a new one and try again."); } finally { setLoading(false); } }
  return <AuthShell eyebrow="Account recovery" title={status === "done" ? "Password updated." : "Choose a new password."} description={status === "done" ? "Your old password no longer works, and active sessions were cleared." : "Use a strong password you haven’t used anywhere else."} footer={<AuthFooterLink prompt="Remembered it?" label="Back to sign in" href="/sign-in" />}>
    {status === "done" ? <div className="success-state"><span className="success-state__icon"><Check size={27} /></span><p>Your account is secure with the new password.</p><Link className="button button--primary button--wide" href="/sign-in">Sign in <ArrowRight size={17} /></Link></div> : <form className="auth-form" onSubmit={submit}>{error && <Notice>{error}</Notice>}<Field label="Reset token" placeholder="Paste your token" value={token} onChange={(event) => setToken(event.target.value)} required /><PasswordField label="New password" autoComplete="new-password" placeholder="At least 12 characters" value={password} onChange={(event) => setPassword(event.target.value)} minLength={12} maxLength={128} required /><Button type="submit" loading={loading} className="button--wide">Update password <ArrowRight size={17} /></Button></form>}
  </AuthShell>;
}
