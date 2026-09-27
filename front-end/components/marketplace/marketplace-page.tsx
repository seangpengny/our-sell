"use client";

import { MarketplaceSection } from "@/components/dashboard/marketplace-section";
import { MarketplaceHeader } from "@/components/marketplace/marketplace-header";

export function MarketplacePage() {
  return <main className="public-marketplace">
    <div className="app-page__ambient app-page__ambient--one" /><div className="app-page__ambient app-page__ambient--two" />
    <MarketplaceHeader />
    <div className="public-marketplace__content"><MarketplaceSection /></div>
  </main>;
}
