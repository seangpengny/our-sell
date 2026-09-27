"use client";

import Link from "next/link";
import { ArrowUpRight, Check, ShieldCheck, Sparkles } from "lucide-react";
import { Logo } from "@/components/ui/logo";
import { ThemeToggle } from "@/components/providers/theme-provider";

const proofPoints = [
  "Explore digital products in one place",
  "Manage your profile and account",
  "Stay in control of your active sessions",
];

export function AuthShell({
  children,
  eyebrow,
  title,
  description,
  footer,
}: Readonly<{
  children: React.ReactNode;
  eyebrow: string;
  title: string;
  description: string;
  footer?: React.ReactNode;
}>) {
  return (
    <main className="auth-page">
      <div className="auth-page__ambient auth-page__ambient--one" />
      <div className="auth-page__ambient auth-page__ambient--two" />
      <div className="auth-layout">
        <section className="auth-story">
          <div className="auth-story__top">
            <Logo dark href="/marketplace" />
            <ThemeToggle compact />
          </div>
          <div className="auth-story__content">
            <span className="eyebrow eyebrow--inverse">
              <Sparkles size={13} /> Your marketplace, connected
            </span>
            <h1>
              Your next chapter
              <br />
              starts <em>here.</em>
            </h1>
            <p>
              Discover new possibilities with Our Sell. One account for your
              marketplace and a workspace that keeps things simple.
            </p>
            <div className="proof-list">
              {proofPoints.map((point) => (
                <div className="proof-list__item" key={point}>
                  <span className="proof-list__icon">
                    <Check size={13} />
                  </span>
                  <span>{point}</span>
                </div>
              ))}
            </div>
          </div>
          <div className="auth-story__footer">
            <span>
              <ShieldCheck size={14} /> Protected by design
            </span>
            <span>© 2026 Our Sell</span>
          </div>
        </section>

        <section className="auth-panel">
          <div className="auth-card animate-spring-in">
            <div className="auth-card__heading">
              <span className="eyebrow">{eyebrow}</span>
              <h2>{title}</h2>
              <p>{description}</p>
            </div>
            {children}
            {footer && <div className="auth-card__footer">{footer}</div>}
          </div>
          <Link className="auth-panel__caption" href="/marketplace">
            ← Back to marketplace
          </Link>
        </section>
      </div>
    </main>
  );
}

export function AuthFooterLink({
  prompt,
  label,
  href,
}: Readonly<{ prompt: string; label: string; href: string }>) {
  return (
    <span>
      {prompt}{" "}
      <Link href={href}>
        {label} <ArrowUpRight size={13} />
      </Link>
    </span>
  );
}
