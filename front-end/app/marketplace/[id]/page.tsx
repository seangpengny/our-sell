import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { ProductDetailPage } from "@/components/marketplace/product-detail-page";
import { getMarketplaceProduct } from "@/lib/marketplace";

type ProductRouteProps = Readonly<{ params: Promise<{ id: string }> }>;

export const dynamic = "force-dynamic";

export async function generateMetadata({ params }: ProductRouteProps): Promise<Metadata> {
  const { id } = await params;
  const product = await getMarketplaceProduct(id);
  return { title: product ? `${product.title} · Marketplace` : "Listing not found" };
}

export default async function ProductRoute({ params }: ProductRouteProps) {
  const { id } = await params;
  const product = await getMarketplaceProduct(id);
  if (!product || !product.published || product.status !== "available") notFound();
  return <ProductDetailPage product={product} />;
}
