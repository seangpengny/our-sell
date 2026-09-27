"use client";

import {
  Globe2,
  LoaderCircle,
  LogOut,
  MonitorSmartphone,
  RefreshCw,
  ShieldCheck,
  Smartphone,
  Trash2,
} from "lucide-react";
import { useState } from "react";
import type { Session } from "@/lib/types";
import { formatDateTime } from "@/lib/format";
import { Button } from "@/components/ui/button";

export function SessionsSection({
  sessions,
  loading,
  error,
  onRefresh,
  onRevoke,
  onRevokeAll,
}: Readonly<{
  sessions: Session[];
  loading: boolean;
  error?: string;
  onRefresh: () => Promise<void>;
  onRevoke: (id: string) => Promise<void>;
  onRevokeAll: () => Promise<void>;
}>) {
  const [revokingAll, setRevokingAll] = useState(false);
  return (
    <section className="dashboard-section animate-fade-up">
      <div className="section-intro">
        <div>
          <span className="eyebrow">Access management</span>
          <h1>Your active sessions</h1>
          <p>
            Keep an eye on where your account is signed in. Revoke anything you
            don’t recognize.
          </p>
        </div>
        <button
          type="button"
          className="refresh-button"
          disabled={loading}
          aria-busy={loading}
          onClick={() => void onRefresh()}
        >
          <RefreshCw size={15} /> Refresh
        </button>
      </div>
      <div className="session-summary glass-panel">
        <span className="session-summary__icon">
          <ShieldCheck size={21} />
        </span>
        <div>
          <strong>
            {loading
              ? "Loading sessions…"
              : error
                ? "Session data unavailable"
                : `${sessions.length} active ${sessions.length === 1 ? "session" : "sessions"}`}
          </strong>
          <p>
            Sessions expire automatically and rotate their refresh token on use.
          </p>
        </div>
        <div className="session-summary__meta">
          <span className="status-chip status-chip--green">
            <span /> Protected
          </span>
        </div>
      </div>
      <div className="panel sessions-panel">
        <div className="panel__header">
          <div>
            <span className="eyebrow">Trusted access</span>
            <h3>Devices & browsers</h3>
          </div>
          {sessions.length > 1 && (
            <Button
              variant="danger"
              loading={revokingAll}
              onClick={async () => {
                setRevokingAll(true);
                try {
                  await onRevokeAll();
                } finally {
                  setRevokingAll(false);
                }
              }}
            >
              <LogOut size={15} /> Sign out everywhere
            </Button>
          )}
        </div>
        {loading ? (
          <div aria-busy="true" aria-label="Loading sessions">
            <div className="skeleton skeleton-row" />
            <div className="skeleton skeleton-row" />
            <div className="skeleton skeleton-row" />
            <span className="sr-only" role="status">
              Loading sessions…
            </span>
          </div>
        ) : error ? (
          <div className="large-empty" role="alert">
            <MonitorSmartphone size={28} />
            <h4>Sessions unavailable</h4>
            <p>{error}</p>
            <Button variant="secondary" onClick={() => void onRefresh()}>
              Try again
            </Button>
          </div>
        ) : sessions.length === 0 ? (
          <div className="large-empty">
            <span className="large-empty__icon">
              <MonitorSmartphone size={23} />
            </span>
            <h4>No active sessions</h4>
            <p>Sign in on a device and it will appear here.</p>
          </div>
        ) : (
          <div className="session-list">
            {sessions.map((session, index) => (
              <SessionRow
                key={session.id}
                session={session}
                current={index === 0}
                onRevoke={onRevoke}
              />
            ))}
          </div>
        )}
      </div>
      <div className="sessions-note">
        <Globe2 size={15} />
        <span>
          IP addresses are shown only to help you recognize access. They’re
          never shared with other users.
        </span>
      </div>
    </section>
  );
}

function SessionRow({
  session,
  current,
  onRevoke,
}: Readonly<{
  session: Session;
  current: boolean;
  onRevoke: (id: string) => Promise<void>;
}>) {
  const [revoking, setRevoking] = useState(false);
  const isMobile = /mobile|iphone|android/i.test(session.device);
  return (
    <div className="session-row">
      <span className={`session-row__device ${current ? "is-current" : ""}`}>
        {isMobile ? <Smartphone size={21} /> : <MonitorSmartphone size={21} />}
      </span>
      <div className="session-row__main">
        <div className="session-row__title">
          <strong>{session.device || "Trusted browser"}</strong>
          {current && (
            <span className="status-chip status-chip--blue">This device</span>
          )}
        </div>
        <span className="session-row__meta">
          {session.ip} <i>·</i> Last active{" "}
          {formatDateTime(session.last_used_at)}
        </span>
      </div>
      <div className="session-row__expiry">
        <span>Expires</span>
        <strong>{formatDateTime(session.expires_at)}</strong>
      </div>
      <button
        className="icon-button icon-button--danger"
        type="button"
        disabled={revoking}
        aria-busy={revoking}
        onClick={async () => {
          setRevoking(true);
          try {
            await onRevoke(session.id);
          } finally {
            setRevoking(false);
          }
        }}
        aria-label={`Revoke ${session.device || "session"}`}
      >
        {revoking ? (
          <LoaderCircle size={16} className="spin" />
        ) : (
          <Trash2 size={16} />
        )}
      </button>
    </div>
  );
}
