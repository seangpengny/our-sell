import type { Product, ProductKind } from "@/lib/types";

export const PRODUCT_KIND_LABELS: Record<ProductKind, string> = {
  facebook_page: "Facebook Pages",
  youtube_channel: "YouTube Channels",
  instagram_account: "Instagram Accounts",
  service: "Digital Services",
};

export const PRODUCT_ROADMAP = [
  { kind: "youtube_channel" as const, label: "YouTube Channels", detail: "Audience-ready channels" },
  { kind: "instagram_account" as const, label: "Instagram Accounts", detail: "Curated niche profiles" },
  { kind: "service" as const, label: "Digital Services", detail: "Growth and setup support" },
];

export function formatProductPrice(priceUsd: number, currency: Product["currency"] = "USD") {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    maximumFractionDigits: 2,
  }).format(priceUsd);
}
