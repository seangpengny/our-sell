"use client";

import Link from "next/link";
import { ArrowLeft, ArrowRight, Mail, ShieldCheck } from "lucide-react";
import { FormEvent, useState } from "react";
import { AuthShell, AuthFooterLink } from "@/components/auth/auth-shell";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { Notice } from "@/components/ui/notice";
import { ApiError, authApi } from "@/lib/api";

export function ForgotPasswordForm() {
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [sent, setSent] = useState(false);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setError(""); setLoading(true);
    try { await authApi.forgotPassword(email); setSent(true); }
    catch (cause) { setError(cause instanceof ApiError ? cause.message : "We couldn’t send the reset email. Please try again."); }
    finally { setLoading(false); }
  }

  return <AuthShell eyebrow="Account recovery" title={sent ? "Check your inbox." : "Let’s get you back in."} description={sent ? "If an account exists for that address, password reset instructions are on the way." : "Enter your email and we’ll send a secure reset link."} footer={<AuthFooterLink prompt="Remembered your password?" label="Sign in" href="/sign-in" />}>
    {sent ? <div className="success-state"><span className="success-state__icon"><Mail size={25} /></span><p>Nothing else is needed here. The link will expire after a short time for your security.</p><Link className="button button--secondary button--wide" href="/sign-in"><ArrowLeft size={16} /> Back to sign in</Link></div> : <form className="auth-form" onSubmit={handleSubmit}>{error && <Notice>{error}</Notice>}<Field label="Email address" type="email" inputMode="email" autoComplete="email" placeholder="you@company.com" value={email} onChange={(event) => setEmail(event.target.value)} required /><Button type="submit" loading={loading} className="button--wide">Send reset link <ArrowRight size={17} /></Button><div className="form-note"><ShieldCheck size={14} /> We never reveal whether an account exists.</div></form>}
  </AuthShell>;
}
