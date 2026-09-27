"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Check, ChevronLeft, ChevronRight, Filter, RefreshCw, Search, ShieldCheck, UserRound, UsersRound } from "lucide-react";
import { ApiError, authApi } from "@/lib/api";
import { formatDate } from "@/lib/format";
import type { User, UserRole } from "@/lib/types";
import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/components/providers/auth-provider";

const PAGE_SIZE = 8;

export function UserManagement({ onToast }: Readonly<{ onToast: (message: string) => void }>) {
  const { user: currentUser } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [search, setSearch] = useState("");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [role, setRole] = useState<UserRole | "all">("all");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [updatingId, setUpdatingId] = useState("");

  const loadUsers = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await authApi.users({ search: appliedSearch, role, page, pageSize: PAGE_SIZE });
      setUsers(result.users);
      setTotal(result.total);
      setTotalPages(result.total_pages);
    } catch (reason) {
      if (reason instanceof ApiError && reason.status === 403) setError("Your administrator access is required to view users.");
      else setError("We couldn’t load the user directory. Please try again.");
    } finally { setLoading(false); }
  }, [appliedSearch, page, role]);

  useEffect(() => {
    const timeout = window.setTimeout(() => { void loadUsers(); }, 0);
    return () => window.clearTimeout(timeout);
  }, [loadUsers]);

  function submitSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPage(1);
    setAppliedSearch(search.trim());
  }

  function changeRoleFilter(nextRole: UserRole | "all") {
    setRole(nextRole);
    setPage(1);
  }

  async function updateRole(account: User, nextRole: UserRole) {
    if (account.role === nextRole) return;
    setUpdatingId(account.id);
    try {
      const updated = await authApi.updateRole(account.id, nextRole);
      setUsers((current) => current.map((item) => item.id === updated.id ? updated : item));
      onToast(`${updated.name} is now ${nextRole === "admin" ? "an administrator" : "a standard user"}.`);
    } catch (reason) {
      onToast(reason instanceof ApiError ? reason.message : "That role change could not be saved.");
    } finally { setUpdatingId(""); }
  }

  const adminCount = useMemo(() => users.filter((account) => account.role === "admin").length, [users]);
  const verifiedCount = useMemo(() => users.filter((account) => Boolean(account.email_verified_at)).length, [users]);

  return (
    <section className="dashboard-section animate-fade-up">
      <div className="section-intro"><div><span className="eyebrow">Directory · access control</span><h1>User management</h1><p>Search every account and keep administrator access exactly where it belongs.</p></div><button type="button" className="refresh-button" disabled={loading} onClick={() => void loadUsers()}><RefreshCw size={15} className={loading ? "spin" : ""} /> Refresh</button></div>
      <div className="directory-stats"><div className="directory-stat"><span className="directory-stat__icon directory-stat__icon--blue"><UsersRound size={16} /></span><span><small>Total accounts</small><strong>{loading ? "—" : total}</strong></span></div><div className="directory-stat"><span className="directory-stat__icon directory-stat__icon--violet"><ShieldCheck size={16} /></span><span><small>Admins on page</small><strong>{loading ? "—" : adminCount}</strong></span></div><div className="directory-stat"><span className="directory-stat__icon directory-stat__icon--green"><Check size={16} /></span><span><small>Verified on page</small><strong>{loading ? "—" : verifiedCount}</strong></span></div></div>
      <div className="directory-panel panel"><div className="directory-toolbar"><form className="directory-search" onSubmit={submitSearch}><Search size={17} /><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search by name or email" aria-label="Search users" /><button type="submit">Search</button></form><label className="filter-control"><Filter size={15} /><span className="sr-only">Filter by role</span><select value={role} onChange={(event) => changeRoleFilter(event.target.value as UserRole | "all")}><option value="all">All roles</option><option value="admin">Administrators</option><option value="user">Standard users</option></select></label></div>
        {error ? <div className="directory-empty" role="alert"><span className="empty-icon"><ShieldCheck size={18} /></span><strong>{error}</strong><Button variant="secondary" onClick={() => void loadUsers()}>Try again</Button></div> : loading ? <div className="directory-empty"><span className="loading-orb loading-orb--small" /><strong>Loading user directory…</strong></div> : users.length === 0 ? <div className="directory-empty"><span className="empty-icon"><UserRound size={18} /></span><strong>No users match those filters.</strong><span>Try a different name, email, or role.</span></div> : <><div className="user-table-wrap"><table className="user-table"><thead><tr><th>Account</th><th>Joined</th><th>Verification</th><th>Role</th></tr></thead><tbody>{users.map((account) => <UserRow key={account.id} account={account} isCurrent={account.id === currentUser?.id} updating={updatingId === account.id} onRoleChange={(nextRole) => void updateRole(account, nextRole)} />)}</tbody></table></div><div className="user-cards">{users.map((account) => <UserCard key={account.id} account={account} isCurrent={account.id === currentUser?.id} updating={updatingId === account.id} onRoleChange={(nextRole) => void updateRole(account, nextRole)} />)}</div><div className="pagination"><span>Showing {Math.min((page - 1) * PAGE_SIZE + 1, total)}–{Math.min(page * PAGE_SIZE, total)} of {total}</span><div><button type="button" className="pagination__button" disabled={page <= 1} onClick={() => setPage((current) => current - 1)} aria-label="Previous page"><ChevronLeft size={16} /></button><span>Page {page} of {Math.max(totalPages, 1)}</span><button type="button" className="pagination__button" disabled={page >= totalPages} onClick={() => setPage((current) => current + 1)} aria-label="Next page"><ChevronRight size={16} /></button></div></div></>}
      </div>
    </section>
  );
}

function UserRow({ account, isCurrent, updating, onRoleChange }: Readonly<{ account: User; isCurrent: boolean; updating: boolean; onRoleChange: (role: UserRole) => void }>) {
  return <tr><td><div className="account-cell"><Avatar name={account.name} size="sm" /><span><strong>{account.name}{isCurrent && <em>You</em>}</strong><small>{account.email}</small></span></div></td><td><span className="table-muted">{formatDate(account.created_at)}</span></td><td><span className={`verification-chip ${account.email_verified_at ? "is-verified" : ""}`}><span />{account.email_verified_at ? "Verified" : "Pending"}</span></td><td><RoleSelect account={account} updating={updating} onChange={onRoleChange} /></td></tr>;
}

function UserCard({ account, isCurrent, updating, onRoleChange }: Readonly<{ account: User; isCurrent: boolean; updating: boolean; onRoleChange: (role: UserRole) => void }>) {
  return <article className="user-card"><div className="account-cell"><Avatar name={account.name} size="sm" /><span><strong>{account.name}{isCurrent && <em>You</em>}</strong><small>{account.email}</small></span></div><div className="user-card__meta"><span><small>Joined</small>{formatDate(account.created_at)}</span><span><small>Status</small><span className={`verification-chip ${account.email_verified_at ? "is-verified" : ""}`}><span />{account.email_verified_at ? "Verified" : "Pending"}</span></span></div><RoleSelect account={account} updating={updating} onChange={onRoleChange} /></article>;
}

function RoleSelect({ account, updating, onChange }: Readonly<{ account: User; updating: boolean; onChange: (role: UserRole) => void }>) {
  return <label className={`role-select role-select--${account.role} ${updating ? "is-updating" : ""}`}><span className="sr-only">Role for {account.name}</span><select value={account.role} disabled={updating} onChange={(event) => onChange(event.target.value as UserRole)}><option value="user">Standard user</option><option value="admin">Administrator</option></select>{updating ? <RefreshCw size={14} className="spin" /> : account.role === "admin" ? <ShieldCheck size={14} /> : <UserRound size={14} />}</label>;
}
