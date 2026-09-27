import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { CheckoutPage } from "@/components/marketplace/checkout-page";
import { getMarketplaceProduct } from "@/lib/marketplace";

type CheckoutRouteProps = Readonly<{ params: Promise<{ id: string }> }>;

export const dynamic = "force-dynamic";

export async function generateMetadata({ params }: CheckoutRouteProps): Promise<Metadata> {
  const { id } = await params;
  const product = await getMarketplaceProduct(id);
  return { title: product ? `Checkout · ${product.title}` : "Checkout" };
}

export default async function CheckoutRoute({ params }: CheckoutRouteProps) {
  const { id } = await params;
  const product = await getMarketplaceProduct(id);
  if (!product || !product.published || product.status !== "available") notFound();
  return <CheckoutPage product={product} />;
}
