"use client";

import Link from "next/link";
import { ArrowRight, Check } from "lucide-react";
import { FormEvent, useState } from "react";
import { AuthShell, AuthFooterLink } from "@/components/auth/auth-shell";
import { Button } from "@/components/ui/button";
import { Field, PasswordField } from "@/components/ui/field";
import { Notice } from "@/components/ui/notice";
import { ApiError, authApi } from "@/lib/api";

export function RegisterForm() {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [created, setCreated] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setError(""); setLoading(true);
    try { await authApi.register(name, email, password); setCreated(true); }
    catch (cause) { setError(cause instanceof ApiError ? cause.message : "We couldn’t create your account. Please try again."); }
    finally { setLoading(false); }
  }

  if (created) return <AuthShell eyebrow="You’re all set" title="Check your inbox." description={`We sent a verification link to ${email}. Verify your email, then come back to sign in.`} footer={<AuthFooterLink prompt="Already verified?" label="Sign in" href="/sign-in" />}><div className="success-state"><span className="success-state__icon"><Check size={27} /></span><p>Your account is ready when you are.</p><Link className="button button--primary button--wide" href="/sign-in">Go to sign in <ArrowRight size={17} /></Link></div></AuthShell>;

  return <AuthShell eyebrow="Start simply" title="Create your space." description="A few details, then you’re ready to go." footer={<AuthFooterLink prompt="Already have an account?" label="Sign in" href="/sign-in" />}>
    <form className="auth-form" onSubmit={handleSubmit}>
      {error && <Notice>{error}</Notice>}
      <Field label="Your name" autoComplete="name" placeholder="Alex Morgan" value={name} onChange={(event) => setName(event.target.value)} required maxLength={100} />
      <Field label="Email address" type="email" inputMode="email" autoComplete="email" placeholder="you@company.com" value={email} onChange={(event) => setEmail(event.target.value)} required maxLength={320} />
      <PasswordField label="Password" autoComplete="new-password" placeholder="At least 12 characters" value={password} onChange={(event) => setPassword(event.target.value)} required minLength={12} maxLength={128} />
      <p className="form-legal">By creating an account, you agree to keep your access secure and your information accurate.</p>
      <Button type="submit" loading={loading} className="button--wide">Create account <ArrowRight size={17} /></Button>
    </form>
  </AuthShell>;
}
