"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ArrowDownToLine, Clock3, QrCode, ShieldCheck, WalletCards } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { ApiError, walletApi } from "@/lib/api";
import type { Wallet, WalletLedgerEntry, WalletTopup } from "@/lib/types";
import { Button } from "@/components/ui/button";

const money = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 2 });

export function WalletSection({ onToast }: Readonly<{ onToast: (message: string) => void }>) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const returnTo = safeReturnPath(searchParams.get("return_to"));
  const [wallet, setWallet] = useState<Wallet | null>(null);
  const [entries, setEntries] = useState<WalletLedgerEntry[]>([]);
  const [topup, setTopup] = useState<WalletTopup | null>(null);
  const [amount, setAmount] = useState("10.00");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [nextWallet, nextLedger] = await Promise.all([walletApi.wallet(), walletApi.ledger()]);
      setWallet(nextWallet);
      setEntries(nextLedger.entries);
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "Your wallet could not be loaded.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const timeout = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timeout);
  }, [load]);

  useEffect(() => {
    if (!topup || topup.status !== "pending_payment") return;
    const interval = window.setInterval(async () => {
      try {
        const refreshed = await walletApi.topup(topup.id);
        setTopup(refreshed);
        if (refreshed.status === "confirmed") {
          onToast(`${money.format(refreshed.amount_usd)} added to your wallet.`);
          await load();
          if (returnTo) router.push(returnTo);
        }
        if (refreshed.status === "expired" || refreshed.status === "failed") {
          onToast("This wallet top-up is no longer payable.");
        }
      } catch {
        // Keep the QR visible; the next poll can recover from a transient error.
      }
    }, 4000);
    return () => window.clearInterval(interval);
  }, [load, onToast, returnTo, router, topup]);

  async function createTopup(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const value = Number.parseFloat(amount);
    if (!Number.isFinite(value) || value <= 0) {
      setError("Enter a valid USD amount.");
      return;
    }
    setSaving(true);
    setError("");
    try {
      setTopup(await walletApi.createTopup(Math.round(value * 100) / 100));
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "The top-up could not be created.");
    } finally {
      setSaving(false);
    }
  }

  if (loading && !wallet) {
    return <section className="dashboard-section animate-fade-up"><div className="directory-empty panel"><span className="loading-orb loading-orb--small" /><strong>Loading your wallet…</strong></div></section>;
  }

  return (
    <section className="dashboard-section animate-fade-up wallet-section">
      <div className="section-intro">
        <div>
          <span className="eyebrow"><WalletCards size={13} /> Wallet</span>
          <h1>Keep your balance ready.</h1>
          <p>Top up with Bakong, then use your available USD balance for Page purchases.</p>
        </div>
        <span className="status-chip status-chip--green"><ShieldCheck size={13} /> Ledger protected</span>
      </div>
      {error && <div className="notice notice--error" role="alert">{error}</div>}
      <div className="wallet-summary-grid">
        <div className="wallet-balance-card glass-panel-elevated"><span className="eyebrow eyebrow--bright">Available balance</span><strong>{money.format(wallet?.available_amount ?? 0)}</strong><small>USD ready to spend</small></div>
        <div className="wallet-balance-card glass-panel"><span className="eyebrow">Held for orders</span><strong>{money.format(wallet?.held_amount ?? 0)}</strong><small>Released or spent by order status</small></div>
      </div>
      <div className="wallet-grid">
        <div className="panel wallet-topup-panel">
          <div className="panel__header"><div><span className="eyebrow"><ArrowDownToLine size={13} /> Add funds</span><h2>Top up with Bakong</h2></div><QrCode size={20} /></div>
          {!topup || topup.status !== "pending_payment" ? (
            <form onSubmit={createTopup}>
              <label className="field"><span>Amount (USD)</span><input type="number" min="1" step="0.01" value={amount} onChange={(event) => setAmount(event.target.value)} /></label>
              <p className="wallet-help">A unique KHQR is created for every top-up. Your balance changes only after the server verifies the Bakong payment.</p>
              <Button type="submit" loading={saving}><QrCode size={15} /> Create KHQR</Button>
            </form>
          ) : (
            <div className="wallet-qr-card">
              <div className="wallet-qr-card__code"><QRCodeSVG value={topup.qr_payload} size={220} includeMargin bgColor="#ffffff" fgColor="#101827" /></div>
              <div className="wallet-qr-card__copy"><span className="eyebrow"><Clock3 size={13} /> Waiting for payment</span><h3>{money.format(topup.amount_usd)}</h3><p>Scan this KHQR with Bakong or a supported Cambodian banking app.</p><small>Reference: {topup.reference}</small><small>Expires: {new Date(topup.expires_at).toLocaleTimeString()}</small><Button variant="ghost" type="button" onClick={() => setTopup(null)}>Cancel QR</Button></div>
            </div>
          )}
          {topup && topup.status !== "pending_payment" && <div className={`wallet-topup-result wallet-topup-result--${topup.status}`}><strong>{topup.status === "confirmed" ? "Top-up confirmed" : `Top-up ${topup.status.replaceAll("_", " ")}`}</strong><span>{topup.status === "confirmed" ? `${money.format(topup.amount_usd)} is now available.` : topup.failure_reason || "You can create another top-up."}</span><Button variant="ghost" type="button" onClick={() => setTopup(null)}>Create another</Button></div>}
        </div>
        <div className="panel wallet-ledger-panel">
          <div className="panel__header"><div><span className="eyebrow">History</span><h2>Wallet ledger</h2></div><span className="panel-header-icon"><WalletCards size={16} /></span></div>
          {entries.length === 0 ? <div className="directory-empty"><strong>No wallet activity yet.</strong><span>Your confirmed top-ups and order holds will appear here.</span></div> : <div className="wallet-ledger-list">{entries.map((entry) => <LedgerRow entry={entry} key={entry.id} />)}</div>}
        </div>
      </div>
    </section>
  );
}

function safeReturnPath(value: string | null) {
  if (!value || !value.startsWith("/") || value.startsWith("//")) return "";
  return value;
}

function LedgerRow({ entry }: Readonly<{ entry: WalletLedgerEntry }>) {
  // Holds move money between available and held balances, so the available
  // delta is the clearest customer-facing movement for the ledger row.
  const delta = entry.available_delta_usd !== 0 ? entry.available_delta_usd : entry.held_delta_usd;
  return <div className="wallet-ledger-row"><span className={`wallet-ledger-row__icon ${delta >= 0 ? "is-positive" : "is-negative"}`}><WalletCards size={14} /></span><span><strong>{entry.description || entry.entry_type.replaceAll("_", " ")}</strong><small>{entry.reference} · {new Date(entry.created_at).toLocaleString()}</small></span><b className={delta >= 0 ? "is-positive" : "is-negative"}>{delta >= 0 ? "+" : ""}{money.format(delta)}</b></div>;
}
