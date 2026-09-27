import { marketplaceApi, type PublicPageListing } from "@/lib/api";
import type { Product } from "@/lib/types";

const ACCENTS: Product["accent"][] = ["blue", "peach", "lavender", "mint"];

function initials(title: string) {
  const words = title.trim().split(/\s+/).filter(Boolean);
  if (words.length === 0) return "FB";
  return words.slice(0, 2).map((word) => word[0]).join("").toUpperCase();
}

function accentFor(id: string): Product["accent"] {
  let hash = 0;
  for (const character of id) hash = (hash * 31 + character.charCodeAt(0)) | 0;
  return ACCENTS[Math.abs(hash) % ACCENTS.length];
}

export function productFromListing(listing: PublicPageListing): Product {
  return {
    id: listing.id,
    kind: "facebook_page",
    title: listing.title || listing.page_name,
    subtitle: listing.subtitle || listing.page_name,
    description: listing.description || "A verified Facebook Page prepared for a direct handover.",
    deliveryWindow: listing.delivery_window,
    priceUsd: listing.price_usd,
    currency: listing.currency,
    status: "available",
    published: true,
    featured: listing.featured,
    sortOrder: listing.sort_order,
    initials: initials(listing.title || listing.page_name),
    accent: accentFor(listing.id),
    stats: [
      { label: "Facebook Page", value: listing.page_name },
      { label: "Page ID", value: listing.facebook_page_id },
    ],
    features: [
      { label: "Facebook Page access transfer", icon: "key" },
      { label: "Verified listing details", icon: "shield" },
      { label: listing.delivery_window, icon: "clock" },
    ],
  };
}

export async function getMarketplaceListings() {
  const firstPage = await marketplaceApi.listings({ page: 1, pageSize: 100 });
  const remainingPages = await Promise.all(
    Array.from({ length: Math.max(0, firstPage.total_pages - 1) }, (_, index) =>
      marketplaceApi.listings({ page: index + 2, pageSize: 100 }),
    ),
  );
  return [firstPage, ...remainingPages].flatMap((result) =>
    result.pages.map((item) =>
      productFromListing({
        ...item.listing,
        facebook_page_id: item.facebook_page_id,
        page_name: item.name,
      }),
    ),
  );
}

export async function getMarketplaceProduct(id: string) {
  try {
    return productFromListing(await marketplaceApi.listing(id));
  } catch {
    return null;
  }
}
