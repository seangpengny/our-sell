"use client";

import Link from "next/link";
import { ArrowRight, LockKeyhole } from "lucide-react";
import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { AuthShell, AuthFooterLink } from "@/components/auth/auth-shell";
import { Button } from "@/components/ui/button";
import { Field, PasswordField } from "@/components/ui/field";
import { Notice } from "@/components/ui/notice";
import { ApiError } from "@/lib/api";
import { useAuth } from "@/components/providers/auth-provider";

export function SignInForm() {
  const router = useRouter();
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      await login(email, password);
      const next = new URLSearchParams(window.location.search).get("next");
      router.push(next?.startsWith("/") ? next : "/app");
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "We couldn’t sign you in. Check your details and try again.");
    } finally { setLoading(false); }
  }

  return <AuthShell eyebrow="Welcome back" title="Good to see you." description="Sign in to pick up where you left off." footer={<AuthFooterLink prompt="New to Our Sell?" label="Create an account" href="/register" />}>
    <form className="auth-form" onSubmit={handleSubmit}>
      {error && <Notice>{error}</Notice>}
      <Field label="Email address" type="email" inputMode="email" autoComplete="email" placeholder="you@company.com" value={email} onChange={(event) => setEmail(event.target.value)} required />
      <div className="field-group"><PasswordField label="Password" autoComplete="current-password" placeholder="Enter your password" value={password} onChange={(event) => setPassword(event.target.value)} required /><div className="field-group__link"><Link href="/forgot-password">Forgot password?</Link></div></div>
      <Button type="submit" loading={loading} className="button--wide">Continue <ArrowRight size={17} /></Button>
      <div className="form-note"><LockKeyhole size={14} /> Your session is encrypted and private.</div>
    </form>
  </AuthShell>;
}
