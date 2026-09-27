import Link from "next/link";
import { ShoppingBag } from "lucide-react";
import { MarketplaceHeader } from "@/components/marketplace/marketplace-header";

export function StorefrontState({
  title,
  description,
  children,
}: Readonly<{
  title: string;
  description: string;
  children?: React.ReactNode;
}>) {
  return (
    <main className="public-marketplace">
      <MarketplaceHeader />
      <div className="public-marketplace__content">
        <section className="marketplace-empty glass-panel">
          <ShoppingBag size={32} />
          <h1>{title}</h1>
          <p>{description}</p>
          {children}
          <Link className="button button--secondary" href="/marketplace">
            Back to marketplace
          </Link>
        </section>
      </div>
    </main>
  );
}
