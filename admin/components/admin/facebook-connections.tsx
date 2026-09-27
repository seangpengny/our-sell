"use client";

import { useCallback, useEffect, useState } from "react";
import { Check, ExternalLink, Link2, RefreshCw, ShieldCheck, Trash2 } from "lucide-react";
import { ApiError, authApi } from "@/lib/api";
import type { FacebookConnection, FacebookPage } from "@/lib/types";
import { Button } from "@/components/ui/button";

type PageMap = Record<string, FacebookPage[]>;
type MessageMap = Record<string, string>;

export function FacebookConnections({ onToast }: Readonly<{ onToast: (message: string) => void }>) {
  const [connections, setConnections] = useState<FacebookConnection[]>([]);
  const [loading, setLoading] = useState(true);
  const [connecting, setConnecting] = useState(false);
  const [disconnecting, setDisconnecting] = useState("");
  const [pages, setPages] = useState<PageMap>({});
  const [pagesLoading, setPagesLoading] = useState<MessageMap>({});
  const [pageErrors, setPageErrors] = useState<MessageMap>({});
  const [error, setError] = useState("");

  const loadPages = useCallback(async (connection: FacebookConnection) => {
    setPagesLoading((current) => ({ ...current, [connection.id]: "loading" }));
    setPageErrors((current) => ({ ...current, [connection.id]: "" }));
    try {
      const result = await authApi.facebookPages(connection.id);
      setPages((current) => ({ ...current, [connection.id]: result }));
    } catch (reason) {
      const message = reason instanceof ApiError && reason.code === "FACEBOOK_REAUTH_REQUIRED"
        ? "This connection needs to be reconnected."
        : "Pages could not be loaded. Try again.";
      setPageErrors((current) => ({ ...current, [connection.id]: message }));
    } finally {
      setPagesLoading((current) => ({ ...current, [connection.id]: "" }));
    }
  }, []);

  const loadConnections = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await authApi.facebookConnections();
      setConnections(result);
      result.forEach((connection) => void loadPages(connection));
    } catch {
      setError("Facebook connections could not be loaded. Please try again.");
    } finally {
      setLoading(false);
    }
  }, [loadPages]);

  useEffect(() => {
    const timeout = window.setTimeout(() => void loadConnections(), 0);
    return () => window.clearTimeout(timeout);
  }, [loadConnections]);

  async function connect() {
    setConnecting(true);
    setError("");
    try {
      const { authorization_url: authorizationURL } = await authApi.startFacebookConnect();
      window.location.assign(authorizationURL);
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : "Facebook could not be connected.");
      setConnecting(false);
    }
  }

  async function disconnect(connection: FacebookConnection) {
    if (!window.confirm(`Disconnect ${connection.name} from Facebook?`)) return;
    setDisconnecting(connection.id);
    try {
      await authApi.disconnectFacebook(connection.id);
      setConnections((current) => current.filter((item) => item.id !== connection.id));
      setPages((current) => removeKey(current, connection.id));
      onToast(`${connection.name} was disconnected.`);
    } catch (reason) {
      onToast(reason instanceof ApiError ? reason.message : "That Facebook account could not be disconnected.");
    } finally {
      setDisconnecting("");
    }
  }

  return (
    <section className="dashboard-section animate-fade-up">
      <div className="section-intro">
        <div>
          <span className="eyebrow"><Link2 size={13} /> Integrations · Facebook</span>
          <h1>Facebook accounts</h1>
          <p>Connect multiple Facebook accounts and see how many Pages each account can access.</p>
        </div>
        <Button onClick={() => void connect()} loading={connecting}>
          <ExternalLink size={15} /> Connect with Facebook
        </Button>
      </div>

      {error && <div className="notice notice--error" role="alert"><ShieldCheck size={16} />{error}</div>}
      {loading ? (
        <div className="directory-empty panel"><span className="loading-orb loading-orb--small" /><strong>Loading Facebook connections…</strong></div>
      ) : connections.length === 0 ? (
        <div className="facebook-empty panel">
          <span className="facebook-empty__icon"><Link2 size={21} /></span>
          <strong>Facebook is not connected</strong>
          <p>Connect an account to securely load the Pages it manages.</p>
          <Button onClick={() => void connect()} loading={connecting}>Connect with Facebook</Button>
        </div>
      ) : (
        <div className="facebook-grid">
          {connections.map((connection) => (
            <article className="facebook-connection panel" key={connection.id}>
              <div className="facebook-connection__header">
                <div className="facebook-connection__identity">
                  <span className="facebook-connection__icon"><Link2 size={17} /></span>
                  <span><strong>{connection.name}</strong><small>Account ID: {connection.facebook_account_id}</small></span>
                </div>
                <span className="status-chip status-chip--green"><span /><Check size={12} /> Connected</span>
              </div>
              <div className="facebook-page-summary">
                <div className="facebook-pages-heading"><span><ShieldCheck size={14} /> Pages available</span><strong>{pagesLoading[connection.id] ? "…" : pages[connection.id]?.length ?? 0}</strong></div>
                {pageErrors[connection.id] ? <div className="facebook-page-summary__error"><span>{pageErrors[connection.id]}</span><Button variant="secondary" onClick={() => void loadPages(connection)}>Try again</Button></div> : <small>Open Page management to view individual Pages.</small>}
              </div>
              <div className="facebook-connection__footer"><span><RefreshCw size={13} /> Pages are refreshed from Meta on demand.</span><Button variant="danger" loading={disconnecting === connection.id} onClick={() => void disconnect(connection)}><Trash2 size={14} /> Disconnect</Button></div>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}

function removeKey<T>(value: Record<string, T>, key: string): Record<string, T> {
  const next = { ...value };
  delete next[key];
  return next;
}
