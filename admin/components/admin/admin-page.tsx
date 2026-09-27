"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/providers/auth-provider";
import { AdminShell } from "@/components/admin/admin-shell";

export function AdminPage() {
  const router = useRouter();
  const { status } = useAuth();

  useEffect(() => {
    if (status === "unauthenticated" || status === "forbidden") router.replace("/sign-in");
  }, [router, status]);

  if (status === "loading" || status === "unauthenticated" || status === "forbidden") {
    return <div className="app-loading"><span className="loading-orb" /><p>Preparing the console…</p></div>;
  }
  return <AdminShell />;
}
