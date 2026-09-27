"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { ArrowRight, LockKeyhole, ShieldCheck, Sparkles } from "lucide-react";
import { useAuth } from "@/components/providers/auth-provider";
import { AdminSignInForm } from "@/components/auth/admin-sign-in-form";
import { Logo } from "@/components/ui/logo";
import { ThemeToggle } from "@/components/providers/theme-provider";

export function AdminSignInPage() {
  const router = useRouter();
  const { status } = useAuth();
  useEffect(() => {
    if (status === "authenticated") router.replace("/");
  }, [router, status]);

  return (
    <main className="auth-page">
      <div className="auth-page__ambient auth-page__ambient--one" />
      <div className="auth-page__ambient auth-page__ambient--two" />
      <div className="auth-layout">
        <section className="auth-story">
          <div className="auth-story__top">
            <Logo href="/sign-in" />
            <ThemeToggle compact />
          </div>
          <div className="auth-story__content">
            <span className="eyebrow eyebrow--bright"><Sparkles size={13} /> Operations console</span>
            <h1>Keep every account in <em>good hands.</em></h1>
            <p>One calm place to review members, protect access, and keep the Our Sell community moving with intention.</p>
            <div className="proof-list">
              <div className="proof-list__item"><span className="proof-list__icon"><ShieldCheck size={13} /></span> Role-aware access controls</div>
              <div className="proof-list__item"><span className="proof-list__icon"><LockKeyhole size={13} /></span> Short-lived secure sessions</div>
              <div className="proof-list__item"><span className="proof-list__icon"><ArrowRight size={13} /></span> A focused view of what matters</div>
            </div>
          </div>
          <div className="auth-story__footer"><span>Private admin workspace</span><span>Our Sell · 2026</span></div>
        </section>
        <section className="auth-panel">
          <AdminSignInForm />
          <p className="auth-panel__caption">Administrator accounts only</p>
        </section>
      </div>
    </main>
  );
}
