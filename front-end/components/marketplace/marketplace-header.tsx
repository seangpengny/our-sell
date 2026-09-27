import Link from "next/link";
import { ArrowRight, ShieldCheck, WalletCards } from "lucide-react";
import { Logo } from "@/components/ui/logo";
import { ThemeToggle } from "@/components/providers/theme-provider";

export function MarketplaceHeader() {
  return (
    <header className="public-marketplace__header glass-panel-elevated">
      <Logo href="/marketplace" />
      <div className="public-marketplace__actions">
        <span className="public-marketplace__trust">
          <ShieldCheck size={14} /> Trusted handovers
        </span>
        <ThemeToggle compact />
        <Link
          className="button button--secondary glass-pill public-marketplace__wallet"
          href="/app?view=wallet"
        >
          <WalletCards size={14} /> Wallet
        </Link>
        <Link
          className="button button--secondary glass-pill public-marketplace__signin"
          href="/sign-in"
        >
          Sign in <ArrowRight size={14} />
        </Link>
      </div>
    </header>
  );
}
