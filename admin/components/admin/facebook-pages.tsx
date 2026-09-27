"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import { Check, ChevronLeft, ChevronRight, Link2, RefreshCw, Search, ShieldCheck } from "lucide-react";
import { ApiError, authApi } from "@/lib/api";
import type {
  FacebookConnection,
  FacebookListingInput,
  FacebookListingStatus,
  FacebookPageInventory,
  FacebookPageInventoryItem,
} from "@/lib/types";
import { Button } from "@/components/ui/button";

const PAGE_SIZES = [25, 50, 100];
const EDITABLE_STATUSES: FacebookListingStatus[] = ["draft", "pending_review", "published", "unlisted"];

type ListingForm = FacebookListingInput & { price: string };

function emptyListing(page: FacebookPageInventoryItem): ListingForm {
  const listing = page.listing;
  return {
    title: listing?.title ?? page.name,
    subtitle: listing?.subtitle ?? "Facebook Page available for transfer",
    description: listing?.description ?? "",
    price: listing ? listing.price_usd.toFixed(2) : "0.00",
    price_usd: listing?.price_usd ?? 0,
    currency: listing?.currency ?? "USD",
    delivery_window: listing?.delivery_window ?? "Transfer after payment and verification",
    status: listing && EDITABLE_STATUSES.includes(listing.status) ? listing.status : "draft",
    featured: listing?.featured ?? false,
    sort_order: listing?.sort_order ?? 0,
  };
}

function statusLabel(status?: FacebookListingStatus) {
  if (!status) return "Not listed";
  return status.replaceAll("_", " ");
}

export function FacebookPages({
  onOpenAccounts,
  onToast,
}: Readonly<{ onOpenAccounts: () => void; onToast: (message: string) => void }>) {
  const [connections, setConnections] = useState<FacebookConnection[]>([]);
  const [inventory, setInventory] = useState<FacebookPageInventory>({ pages: [], total: 0, page: 1, page_size: 25, total_pages: 1 });
  const [search, setSearch] = useState("");
  const [selectedConnection, setSelectedConnection] = useState("all");
  const [pageNumber, setPageNumber] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);
  const [error, setError] = useState("");
  const [editingPage, setEditingPage] = useState<FacebookPageInventoryItem | null>(null);
  const [form, setForm] = useState<ListingForm | null>(null);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [nextConnections, nextInventory] = await Promise.all([
        authApi.facebookConnections(),
        authApi.facebookPageInventory({
          connectionId: selectedConnection === "all" ? undefined : selectedConnection,
          search,
          page: pageNumber,
          pageSize,
        }),
      ]);
      setConnections(nextConnections);
      setInventory(nextInventory);
      if (pageNumber > nextInventory.total_pages) setPageNumber(Math.max(nextInventory.total_pages, 1));
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "Facebook Pages could not be loaded.");
    } finally {
      setLoading(false);
    }
  }, [pageNumber, pageSize, search, selectedConnection]);

  useEffect(() => {
    const timeout = window.setTimeout(() => void load(), 180);
    return () => window.clearTimeout(timeout);
  }, [load]);

  async function syncPages() {
    setSyncing(true);
    setError("");
    try {
      const targets = selectedConnection === "all"
        ? connections
        : connections.filter((connection) => connection.id === selectedConnection);
      await Promise.all(targets.map((connection) => authApi.syncFacebookConnection(connection.id)));
      onToast(targets.length ? "Facebook Pages synchronized." : "Connect a Facebook account first.");
      await load();
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "Facebook Pages could not be synchronized.");
    } finally {
      setSyncing(false);
    }
  }

  function openEditor(page: FacebookPageInventoryItem) {
    setEditingPage(page);
    setForm(emptyListing(page));
  }

  function updateForm<Key extends keyof ListingForm>(key: Key, value: ListingForm[Key]) {
    setForm((current) => current ? { ...current, [key]: value } : current);
  }

  async function saveListing(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editingPage || !form) return;
    const price = Number.parseFloat(form.price);
    if (!Number.isFinite(price) || price < 0) {
      setError("Price must be zero or greater.");
      return;
    }
    setSaving(true);
    setError("");
    try {
      const input: FacebookListingInput = {
        title: form.title,
        subtitle: form.subtitle,
        description: form.description,
        price_usd: Math.round(price * 100) / 100,
        currency: "USD",
        delivery_window: form.delivery_window,
        status: form.status,
        featured: form.featured,
        sort_order: form.sort_order,
      };
      await authApi.saveFacebookListing(editingPage.id, input);
      setEditingPage(null);
      setForm(null);
      onToast(`Listing saved for ${editingPage.name}.`);
      await load();
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "The listing could not be saved.");
    } finally {
      setSaving(false);
    }
  }

  const firstVisible = inventory.total === 0 ? 0 : (inventory.page - 1) * inventory.page_size + 1;
  const lastVisible = inventory.total === 0 ? 0 : Math.min(inventory.page * inventory.page_size, inventory.total);
  const publishedCount = inventory.pages.filter((page) => page.listing?.status === "published").length;

  return (
    <section className="dashboard-section animate-fade-up">
      <div className="section-intro">
        <div>
          <span className="eyebrow"><ShieldCheck size={13} /> Facebook · Pages</span>
          <h1>Page management</h1>
          <p>Keep your synced Pages in inventory and prepare the ones you want to publish in the marketplace.</p>
        </div>
        <Button variant="secondary" onClick={() => void syncPages()} loading={syncing}>
          <RefreshCw size={15} /> Sync Pages
        </Button>
      </div>

      {error && <div className="notice notice--error" role="alert"><ShieldCheck size={16} />{error}</div>}
      {loading && connections.length === 0 ? (
        <div className="directory-empty panel"><span className="loading-orb loading-orb--small" /><strong>Loading Page management…</strong></div>
      ) : connections.length === 0 ? (
        <div className="facebook-empty panel">
          <span className="facebook-empty__icon"><Link2 size={21} /></span>
          <strong>No Facebook accounts connected</strong>
          <p>Connect a Facebook account first, then its managed Pages will be stored here.</p>
          <Button onClick={onOpenAccounts}>Open Facebook accounts</Button>
        </div>
      ) : (
        <>
          <div className="directory-stats">
            <div className="directory-stat"><span className="directory-stat__icon directory-stat__icon--blue"><Link2 size={17} /></span><span><small>Connected accounts</small><strong>{connections.length}</strong></span></div>
            <div className="directory-stat"><span className="directory-stat__icon directory-stat__icon--green"><ShieldCheck size={17} /></span><span><small>Stored Pages</small><strong>{inventory.total}</strong></span></div>
            <div className="directory-stat"><span className="directory-stat__icon directory-stat__icon--violet"><Check size={17} /></span><span><small>Published on page</small><strong>{publishedCount}</strong></span></div>
          </div>

          {editingPage && form && <ListingEditor page={editingPage} form={form} saving={saving} onChange={updateForm} onCancel={() => { setEditingPage(null); setForm(null); }} onSubmit={saveListing} />}

          <div className="panel directory-panel facebook-page-management-panel">
            <div className="directory-toolbar">
              <label className="directory-search"><Search size={16} /><span className="sr-only">Search Pages</span><input value={search} onChange={(event) => { setSearch(event.target.value); setPageNumber(1); }} placeholder="Search Pages or account…" /></label>
              <label className="filter-control"><span className="sr-only">Filter by Facebook account</span><select value={selectedConnection} onChange={(event) => { setSelectedConnection(event.target.value); setPageNumber(1); }}><option value="all">All accounts</option>{connections.map((connection) => <option key={connection.id} value={connection.id}>{connection.name}</option>)}</select></label>
              <label className="filter-control filter-control--page-size"><span className="sr-only">Pages per view</span><select value={pageSize} onChange={(event) => { setPageSize(Number(event.target.value)); setPageNumber(1); }}>{PAGE_SIZES.map((size) => <option key={size} value={size}>{size} / page</option>)}</select></label>
            </div>
            {inventory.total === 0 ? (
              <div className="directory-empty"><span className="empty-icon"><Search size={18} /></span><strong>No stored Pages match your filters.</strong><span>Sync a connected account or try a different search.</span></div>
            ) : (
              <div className="facebook-page-table-wrap">
                <table className="facebook-page-table">
                  <thead><tr><th>Page</th><th>Connected account</th><th>Last synced</th><th>Listing</th><th /></tr></thead>
                  <tbody>{inventory.pages.map((page) => <PageRow key={page.id} page={page} onEdit={() => openEditor(page)} />)}</tbody>
                </table>
              </div>
            )}
            {inventory.total > 0 && <div className="facebook-page-pagination"><span>Showing {firstVisible}–{lastVisible} of {inventory.total} Pages</span><div><button type="button" className="pagination__button" disabled={inventory.page <= 1} onClick={() => setPageNumber((current) => Math.max(1, current - 1))} aria-label="Previous page"><ChevronLeft size={16} /></button><span>Page {inventory.page} of {inventory.total_pages}</span><button type="button" className="pagination__button" disabled={inventory.page >= inventory.total_pages} onClick={() => setPageNumber((current) => Math.min(inventory.total_pages, current + 1))} aria-label="Next page"><ChevronRight size={16} /></button></div></div>}
          </div>
        </>
      )}
    </section>
  );
}

function PageRow({ page, onEdit }: Readonly<{ page: FacebookPageInventoryItem; onEdit: () => void }>) {
  const listing = page.listing;
  return (
    <tr>
      <td><span className="facebook-page-table__name"><span className="facebook-page-management-card__icon"><ShieldCheck size={16} /></span><span><strong>{page.name}</strong><small>{page.facebook_page_id}</small></span></span></td>
      <td><span className="facebook-page-table__account"><Link2 size={14} />{page.connection_name}</span></td>
      <td>{page.last_synced_at ? new Date(page.last_synced_at).toLocaleString() : "Not synced"}</td>
      <td><span className={`listing-status listing-status--${listing?.status ?? "none"}`}><span />{statusLabel(listing?.status)}</span></td>
      <td><Button variant="ghost" onClick={onEdit}>{listing ? "Edit listing" : "Create listing"}</Button></td>
    </tr>
  );
}

function ListingEditor({
  page,
  form,
  saving,
  onChange,
  onCancel,
  onSubmit,
}: Readonly<{
  page: FacebookPageInventoryItem;
  form: ListingForm;
  saving: boolean;
  onChange: <Key extends keyof ListingForm>(key: Key, value: ListingForm[Key]) => void;
  onCancel: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}>) {
  return (
    <div className="panel facebook-listing-editor">
      <div className="panel-heading"><div><span className="eyebrow"><ShieldCheck size={13} /> Marketplace listing</span><h2>{page.name}</h2><p>Save this Page as a draft, send it for review, or publish it.</p></div></div>
      <form onSubmit={onSubmit}>
        <div className="facebook-listing-editor__grid">
          <label className="field"><span>Listing title</span><input value={form.title} onChange={(event) => onChange("title", event.target.value)} required maxLength={255} /></label>
          <label className="field"><span>Subtitle</span><input value={form.subtitle} onChange={(event) => onChange("subtitle", event.target.value)} required maxLength={255} /></label>
          <label className="field"><span>Price (USD)</span><input type="number" min="0" step="0.01" value={form.price} onChange={(event) => onChange("price", event.target.value)} required /></label>
          <label className="field"><span>Delivery window</span><input value={form.delivery_window} onChange={(event) => onChange("delivery_window", event.target.value)} required maxLength={120} /></label>
          <label className="field"><span>Status</span><select value={form.status} onChange={(event) => onChange("status", event.target.value as FacebookListingStatus)}>{EDITABLE_STATUSES.map((status) => <option key={status} value={status}>{statusLabel(status)}</option>)}</select></label>
        </div>
        <label className="field"><span>Description</span><textarea className="facebook-listing-editor__textarea" value={form.description} onChange={(event) => onChange("description", event.target.value)} maxLength={5000} rows={4} /></label>
        <label className="facebook-listing-editor__checkbox"><input type="checkbox" checked={form.featured} onChange={(event) => onChange("featured", event.target.checked)} /> Feature this listing near the top of the marketplace</label>
        <div className="facebook-listing-editor__actions"><Button variant="ghost" onClick={onCancel}>Cancel</Button><Button type="submit" loading={saving}>Save listing</Button></div>
      </form>
    </div>
  );
}
