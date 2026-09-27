"use client";

import { FormEvent, useState } from "react";
import { Eye, EyeOff, LockKeyhole, Mail, ShieldCheck } from "lucide-react";
import { ApiError } from "@/lib/api";
import { useAuth } from "@/components/providers/auth-provider";
import { Button } from "@/components/ui/button";

export function AdminSignInForm() {
  const { login, status } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      await login(email, password);
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "We couldn’t sign you in. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="auth-card animate-spring-in">
      <div className="auth-card__heading">
        <span className="eyebrow"><ShieldCheck size={13} /> Secure sign in</span>
        <h2>Welcome back.</h2>
        <p>Use your administrator account to enter the operations console.</p>
      </div>
      <form className="auth-form" onSubmit={submit}>
        <label className="field">
          <span className="field__label-row"><span>Work email</span></span>
          <span className="field__control-wrap">
            <Mail className="field__leading" size={17} />
            <input className="field__control field__control--with-leading" type="email" value={email} onChange={(event) => setEmail(event.target.value)} placeholder="admin@oursell.com" autoComplete="email" required />
          </span>
        </label>
        <label className="field">
          <span className="field__label-row"><span>Password</span></span>
          <span className="field__control-wrap">
            <LockKeyhole className="field__leading" size={17} />
            <input className="field__control field__control--with-leading field__control--with-trailing" type={showPassword ? "text" : "password"} value={password} onChange={(event) => setPassword(event.target.value)} placeholder="Enter your password" autoComplete="current-password" required />
            <button className="input-icon-button" type="button" onClick={() => setShowPassword((value) => !value)} aria-label={showPassword ? "Hide password" : "Show password"}>{showPassword ? <EyeOff size={17} /> : <Eye size={17} />}</button>
          </span>
        </label>
        {error && <div className="notice notice--error" role="alert">{error}</div>}
        <Button type="submit" className="button--wide" loading={loading} disabled={status === "loading"}>Enter console <span aria-hidden="true">→</span></Button>
      </form>
      <p className="form-note"><LockKeyhole size={13} /> Your session is encrypted and protected.</p>
    </div>
  );
}
