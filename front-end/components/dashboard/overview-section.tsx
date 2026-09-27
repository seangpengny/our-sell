"use client";

import {
  ArrowUpRight,
  Check,
  ChevronRight,
  Clock3,
  MailCheck,
  RefreshCw,
  ShieldCheck,
  Sparkles,
  Wifi,
} from "lucide-react";
import type { Session, User } from "@/lib/types";
import { formatDate, relativeTime } from "@/lib/format";
import type { DashboardView } from "@/components/dashboard/dashboard-shell";
import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";

export function OverviewSection({
  user,
  sessions,
  loading,
  error,
  onChangeView,
  onRefresh,
  onResendVerification,
}: Readonly<{
  user: User;
  sessions: Session[];
  loading: boolean;
  error?: string;
  onChangeView: (view: DashboardView) => void;
  onRefresh: () => Promise<void>;
  onResendVerification: () => Promise<void>;
}>) {
  const isVerified = Boolean(user.email_verified_at);
  return (
    <section className="dashboard-section animate-fade-up">
      <div className="section-intro">
        <div>
          <span className="eyebrow">Your workspace · today</span>
          <h1>Welcome back, {user.name.split(" ")[0]} </h1>
          <p>
            Here’s a clear view of your account and what needs your attention.
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
      <div className="hero-card glass-panel-elevated">
        <div className="hero-card__glow" />
        <div className="hero-card__copy">
          <span className="eyebrow eyebrow--bright">
            <Sparkles size={13} /> Your security snapshot
          </span>
          <h2>Quietly in control.</h2>
          <p>
            Your account is protected by short-lived access tokens, rotating
            sessions, and a private-by-default refresh flow.
          </p>
          <div className="hero-card__meta">
            <span>
              <Wifi size={14} /> Session management
            </span>
            <span>
              <Clock3 size={14} /> Automatic session expiry
            </span>
          </div>
        </div>
        <div className="security-score">
          <div className="security-score__ring">
            <ShieldCheck size={30} />
          </div>
          <span>{isVerified ? "Email verified" : "Verify your email"}</span>
        </div>
      </div>
      <div className="stats-grid">
        <StatCard
          label="Active sessions"
          value={loading || error ? "—" : String(sessions.length)}
          note={
            error
              ? "Refresh to retry"
              : loading
                ? "Loading devices…"
                : sessions.length === 1
                  ? "1 trusted device"
                  : `${sessions.length} trusted devices`
          }
          icon={<Wifi size={17} />}
          tone="blue"
          onClick={() => onChangeView("sessions")}
        />
        <StatCard
          label="Email status"
          value={isVerified ? "Verified" : "Action needed"}
          note={isVerified ? "Confirmed address" : "Verify your email"}
          icon={<MailCheck size={17} />}
          tone={isVerified ? "green" : "amber"}
          onClick={isVerified ? undefined : () => void onResendVerification()}
        />
        <StatCard
          label="Member since"
          value={formatDate(user.created_at, {
            month: "short",
            day: "numeric",
          })}
          note={formatDate(user.created_at, { year: "numeric" })}
          icon={<ShieldCheck size={17} />}
          tone="violet"
        />
      </div>
      {!isVerified && (
        <div className="attention-card">
          <span className="attention-card__icon">
            <MailCheck size={18} />
          </span>
          <div>
            <strong>Your email still needs a quick check.</strong>
            <p>
              We’ll keep your workspace safe, but verified email unlocks the
              full experience.
            </p>
          </div>
          <Button
            variant="secondary"
            onClick={() => void onResendVerification()}
          >
            Resend email <ArrowUpRight size={15} />
          </Button>
        </div>
      )}
      <div className="dashboard-grid">
        <div className="panel security-panel">
          <div className="panel__header">
            <div>
              <span className="eyebrow">Protection layer</span>
              <h3>Account health</h3>
            </div>
            <span className="status-chip status-chip--green">
              <span /> Healthy
            </span>
          </div>
          <div className="health-list">
            <HealthRow
              label="Strong password"
              detail="Argon2id protected"
              complete
            />
            <HealthRow
              label="Email verification"
              detail={
                isVerified ? "Address confirmed" : "Waiting for confirmation"
              }
              complete={isVerified}
            />
            <HealthRow
              label="Session hygiene"
              detail="Rotating refresh tokens"
              complete
            />
          </div>
          <button
            className="panel-link"
            type="button"
            onClick={() => onChangeView("account")}
          >
            Review account settings <ChevronRight size={15} />
          </button>
        </div>
        <div className="panel activity-panel">
          <div className="panel__header">
            <div>
              <span className="eyebrow">Latest activity</span>
              <h3>Recent sessions</h3>
            </div>
            <button
              className="text-button"
              type="button"
              onClick={() => onChangeView("sessions")}
            >
              See all <ArrowUpRight size={14} />
            </button>
          </div>
          {loading ? (
            <div className="panel-empty">
              <span className="loading-orb loading-orb--small" /> Loading
              sessions…
            </div>
          ) : error ? (
            <div className="panel-empty" role="alert">
              <span>{error}</span>
              <Button variant="secondary" onClick={() => void onRefresh()}>
                Try again
              </Button>
            </div>
          ) : sessions.length === 0 ? (
            <div className="panel-empty">
              <MonitorIcon />
              <span>No active sessions yet.</span>
            </div>
          ) : (
            <div className="activity-list">
              {sessions.slice(0, 3).map((session, index) => (
                <div className="activity-item" key={session.id}>
                  <Avatar name={session.device} size="sm" />
                  <div>
                    <strong>{session.device || "Trusted device"}</strong>
                    <span>
                      {index === 0
                        ? "Current session"
                        : `${session.ip} · ${relativeTime(session.last_used_at)}`}
                    </span>
                  </div>
                  <span
                    className={`activity-item__indicator ${index === 0 ? "is-live" : ""}`}
                  />
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
      <p className="privacy-line">
        <ShieldCheck size={14} /> We only show safe device metadata. Passwords
        and raw tokens never leave the security boundary.
      </p>
    </section>
  );
}

function StatCard({
  label,
  value,
  note,
  icon,
  tone,
  onClick,
}: Readonly<{
  label: string;
  value: string;
  note: string;
  icon: React.ReactNode;
  tone: string;
  onClick?: () => void;
}>) {
  const content = (
    <>
      <span className="stat-card__icon">{icon}</span>
      <span className="stat-card__label">{label}</span>
      <strong>{value}</strong>
      <span className="stat-card__note">
        {note}
        {onClick && <ArrowUpRight size={12} />}
      </span>
    </>
  );
  return onClick ? (
    <button
      type="button"
      className={`stat-card stat-card--${tone} is-clickable`}
      onClick={onClick}
    >
      {content}
    </button>
  ) : (
    <div className={`stat-card stat-card--${tone}`}>{content}</div>
  );
}

function HealthRow({
  label,
  detail,
  complete,
}: Readonly<{ label: string; detail: string; complete: boolean }>) {
  return (
    <div className="health-row">
      <span className={`health-row__check ${complete ? "is-complete" : ""}`}>
        {complete ? <Check size={13} /> : <span />}
      </span>
      <span>
        <strong>{label}</strong>
        <small>{detail}</small>
      </span>
      <ChevronRight size={14} />
    </div>
  );
}

function MonitorIcon() {
  return (
    <span className="empty-icon">
      <Wifi size={18} />
    </span>
  );
}
