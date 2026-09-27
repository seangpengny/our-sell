"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import { Check, ChevronLeft, ChevronRight, Clock3, Filter, RefreshCw, Search, WalletCards, XCircle } from "lucide-react";
import { ApiError, authApi } from "@/lib/api";
import { formatDate } from "@/lib/format";
import type { AdminWalletTopup, WalletTopupStatus } from "@/lib/types";
import { Button } from "@/components/ui/button";

const PAGE_SIZE = 20;
const statuses: Array<{ value: "" | WalletTopupStatus; label: string }> = [
  { value: "", label: "All statuses" },
  { value: "pending_payment", label: "Pending payment" },
  { value: "confirmed", label: "Confirmed" },
  { value: "failed", label: "Failed" },
  { value: "expired", label: "Expired" },
  { value: "reversed", label: "Reversed" },
];

const money = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 2 });

export function WalletPayments() {
  const [topups, setTopups] = useState<AdminWalletTopup[]>([]);
  const [search, setSearch] = useState("");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [status, setStatus] = useState<"" | WalletTopupStatus>("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [recheckingId, setRecheckingId] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await authApi.walletTopups({ search: appliedSearch, status, page, pageSize: PAGE_SIZE });
      setTopups(result.topups);
      setTotal(result.total);
      setTotalPages(result.total_pages);
      if (page > result.total_pages && result.total_pages > 0) setPage(result.total_pages);
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "Wallet top-ups could not be loaded.");
    } finally {
      setLoading(false);
    }
  }, [appliedSearch, page, status]);

  useEffect(() => {
    const timeout = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timeout);
  }, [load]);

  function submitSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPage(1);
    setAppliedSearch(search.trim());
  }

  async function recheck(topup: AdminWalletTopup) {
    setRecheckingId(topup.id);
    setError("");
    try {
      await authApi.recheckWalletTopup(topup.id);
      await load();
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "The Bakong payment could not be rechecked.");
    } finally {
      setRecheckingId("");
    }
  }

  const confirmed = topups.filter((topup) => topup.status === "confirmed").length;
  const pending = topups.filter((topup) => topup.status === "pending_payment").length;
  const visibleAmount = topups.reduce((sum, topup) => sum + topup.amount_usd, 0);

  return (
    <section className="dashboard-section animate-fade-up">
      <div className="section-intro">
        <div>
          <span className="eyebrow"><WalletCards size={13} /> Wallet · payments</span>
          <h1>Wallet top-ups</h1>
          <p>Review Bakong payment verification, settlement references, and wallet credit history.</p>
        </div>
        <button type="button" className="refresh-button" disabled={loading} onClick={() => void load()}>
          <RefreshCw size={15} className={loading ? "spin" : ""} /> Refresh
        </button>
      </div>

      <div className="directory-stats">
        <div className="directory-stat"><span className="directory-stat__icon directory-stat__icon--blue"><WalletCards size={16} /></span><span><small>Top-ups on page</small><strong>{loading ? "—" : topups.length}</strong></span></div>
        <div className="directory-stat"><span className="directory-stat__icon directory-stat__icon--green"><Check size={16} /></span><span><small>Confirmed on page</small><strong>{loading ? "—" : confirmed}</strong></span></div>
        <div className="directory-stat"><span className="directory-stat__icon directory-stat__icon--violet"><Clock3 size={16} /></span><span><small>Visible amount</small><strong>{loading ? "—" : money.format(visibleAmount)}</strong></span></div>
      </div>

      <div className="directory-panel panel">
        <div className="directory-toolbar">
          <form className="directory-search" onSubmit={submitSearch}>
            <Search size={17} />
            <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search reference or user email" aria-label="Search wallet top-ups" />
            <button type="submit">Search</button>
          </form>
          <label className="filter-control"><Filter size={15} /><span className="sr-only">Filter top-ups by status</span><select value={status} onChange={(event) => { setStatus(event.target.value as "" | WalletTopupStatus); setPage(1); }}>{statuses.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select></label>
        </div>
        {error ? (
          <div className="directory-empty" role="alert"><span className="empty-icon"><XCircle size={18} /></span><strong>{error}</strong><Button variant="secondary" onClick={() => void load()}>Try again</Button></div>
        ) : loading ? (
          <div className="directory-empty"><span className="loading-orb loading-orb--small" /><strong>Loading wallet top-ups…</strong></div>
        ) : topups.length === 0 ? (
          <div className="directory-empty"><span className="empty-icon"><WalletCards size={18} /></span><strong>No wallet top-ups match those filters.</strong><span>Confirmed and pending Bakong payments will appear here.</span></div>
        ) : (
          <>
            <div className="wallet-admin-table-wrap">
              <table className="wallet-admin-table">
                <thead><tr><th>Customer</th><th>Reference</th><th>Amount</th><th>Status</th><th>Bakong transaction</th><th>Created</th><th /></tr></thead>
                <tbody>{topups.map((topup) => <TopupRow key={topup.id} topup={topup} rechecking={recheckingId === topup.id} onRecheck={() => void recheck(topup)} />)}</tbody>
              </table>
            </div>
            <div className="pagination"><span>Showing {Math.min((page - 1) * PAGE_SIZE + 1, total)}–{Math.min(page * PAGE_SIZE, total)} of {total}</span><div><button type="button" className="pagination__button" disabled={page <= 1} onClick={() => setPage((current) => current - 1)} aria-label="Previous page"><ChevronLeft size={16} /></button><span>Page {page} of {Math.max(totalPages, 1)}</span><button type="button" className="pagination__button" disabled={page >= totalPages} onClick={() => setPage((current) => current + 1)} aria-label="Next page"><ChevronRight size={16} /></button></div></div>
          </>
        )}
      </div>
      {pending > 0 && <p className="privacy-line"><Clock3 size={14} /> Pending top-ups are rechecked by the customer flow or a scheduled reconciliation worker.</p>}
    </section>
  );
}

function TopupRow({ topup, rechecking, onRecheck }: Readonly<{ topup: AdminWalletTopup; rechecking: boolean; onRecheck: () => void }>) {
  const statusLabel = topup.status.replaceAll("_", " ");
  return (
    <tr>
      <td><span className="wallet-admin-customer"><strong>{topup.user_name}</strong><small>{topup.user_email}</small></span></td>
      <td><code>{topup.reference}</code></td>
      <td><strong>{money.format(topup.amount_usd)}</strong></td>
      <td><span className={`wallet-admin-status wallet-admin-status--${topup.status}`}><span />{statusLabel}</span></td>
      <td><code>{topup.provider_transaction_id || "Not verified"}</code></td>
      <td><span className="table-muted">{formatDate(topup.created_at)}</span></td>
      <td>{topup.status === "pending_payment" && <Button variant="ghost" onClick={onRecheck} loading={rechecking}>Recheck</Button>}</td>
    </tr>
  );
}
