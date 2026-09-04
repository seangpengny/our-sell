"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/providers/auth-provider";
import { DashboardShell } from "@/components/dashboard/dashboard-shell";

export function DashboardPage() {
  const router = useRouter();
  const { status } = useAuth();
  useEffect(() => { if (status === "unauthenticated") router.replace("/sign-in"); }, [router, status]);
  if (status === "loading" || status === "unauthenticated") return <div className="app-loading"><span className="loading-orb" /><p>Preparing your workspace…</p></div>;
  return <DashboardShell />;
}
