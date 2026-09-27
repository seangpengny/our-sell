"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/providers/auth-provider";
import { DashboardShell } from "@/components/dashboard/dashboard-shell";

export function DashboardPage() {
  const router = useRouter();
  const { status } = useAuth();
  useEffect(() => {
    if (status !== "unauthenticated") return;
    const next = `${window.location.pathname}${window.location.search}`;
    router.replace(`/sign-in?next=${encodeURIComponent(next)}`);
  }, [router, status]);
  if (status === "loading" || status === "unauthenticated") return <div className="app-loading"><span className="loading-orb" /><p>Preparing your workspace…</p></div>;
  return <DashboardShell />;
}
