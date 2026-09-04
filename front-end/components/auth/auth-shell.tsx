"use client";

import Link from "next/link";
import { ArrowUpRight, Check, ShieldCheck, Sparkles } from "lucide-react";
import { Logo } from "@/components/ui/logo";
import { ThemeToggle } from "@/components/providers/theme-provider";

const proofPoints = ["One calm place for your account", "Session-level control, always visible", "Security that stays out of your way"];

export function AuthShell({ children, eyebrow, title, description, footer }: Readonly<{ children: React.ReactNode; eyebrow: string; title: string; description: string; footer?: React.ReactNode }>) {
  return (
    <main className="auth-page">
      <div className="auth-page__ambient auth-page__ambient--one" />
      <div className="auth-page__ambient auth-page__ambient--two" />
      <div className="auth-layout">
        <section className="auth-story">
          <div className="auth-story__top"><Logo dark /><ThemeToggle compact /></div>
          <div className="auth-story__content">
            <span className="eyebrow eyebrow--inverse"><Sparkles size={13} /> thoughtful selling starts here</span>
            <h1>Make room for the work that <em>matters.</em></h1>
            <p>Our Sell brings your account, access, and everyday momentum into one clear, quiet workspace.</p>
            <div className="proof-list">{proofPoints.map((point) => <div className="proof-list__item" key={point}><span className="proof-list__icon"><Check size={13} /></span><span>{point}</span></div>)}</div>
          </div>
          <div className="auth-story__footer"><span><ShieldCheck size={14} /> Protected by design</span><span>© 2026 Our Sell</span></div>
        </section>

        <section className="auth-panel">
          <div className="auth-card animate-spring-in">
            <div className="auth-card__heading"><span className="eyebrow">{eyebrow}</span><h2>{title}</h2><p>{description}</p></div>
            {children}
            {footer && <div className="auth-card__footer">{footer}</div>}
          </div>
          <p className="auth-panel__caption">Private by default. Built for focus.</p>
        </section>
      </div>
    </main>
  );
}

export function AuthFooterLink({ prompt, label, href }: Readonly<{ prompt: string; label: string; href: string }>) {
  return <span>{prompt} <Link href={href}>{label} <ArrowUpRight size={13} /></Link></span>;
}
