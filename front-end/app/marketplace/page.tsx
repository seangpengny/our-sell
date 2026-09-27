import type { Metadata } from "next";
import { MarketplacePage } from "@/components/marketplace/marketplace-page";

export const metadata: Metadata = {
  title: "Marketplace",
  description: "Browse verified digital pages and properties available for direct handover.",
};

export default function MarketplaceRoute() {
  return <MarketplacePage />;
}
