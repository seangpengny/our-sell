import Link from "next/link";
import {
  ArrowLeft,
  ArrowRight,
  Check,
  CheckCircle2,
  Clock3,
  KeyRound,
  LockKeyhole,
  MessageCircle,
  ShieldCheck,
} from "lucide-react";
import { MarketplaceHeader } from "@/components/marketplace/marketplace-header";
import { PRODUCT_KIND_LABELS, formatProductPrice } from "@/lib/products";
import type { Product } from "@/lib/types";

import { FeatureIcon } from "@/components/marketplace/product-features";

export function ProductDetailPage({ product }: Readonly<{ product: Product }>) {
  return (
    <main className="public-marketplace product-detail-page">
      <div className="app-page__ambient app-page__ambient--one" />
      <div className="app-page__ambient app-page__ambient--two" />
      <MarketplaceHeader />
      <div className="product-subpage__content">
        <Link className="back-link" href="/marketplace">
          <ArrowLeft size={14} /> Back to marketplace
        </Link>
        <div className="product-detail-layout">
          <div className="product-detail-main">
            <section
              className={`product-detail-hero product-detail-hero--${product.accent}`}
            >
              <div className="product-detail-hero__visual">
                <span className="product-card__shine" />
                <span className="product-detail-hero__type">
                  {PRODUCT_KIND_LABELS[product.kind]}
                </span>
                <span className="product-card__avatar product-detail-hero__avatar">
                  {product.initials}
                </span>
                <span className="product-card__verified">
                  <Check size={12} /> Verified listing
                </span>
              </div>
              <div className="product-detail-hero__copy">
                <span className="eyebrow eyebrow--bright">
                  <CheckCircle2 size={13} /> Available now
                </span>
                <h1>{product.title}</h1>
                <p>{product.description}</p>
                <div className="product-detail-hero__meta">
                  <span>
                    <ShieldCheck size={14} /> 100% clean
                  </span>
                  <span>
                    <KeyRound size={14} /> Admin transfer
                  </span>
                  <span>
                    <Clock3 size={14} /> {product.deliveryWindow}
                  </span>
                </div>
              </div>
            </section>

            <section className="detail-panel glass-panel">
              <div className="detail-panel__heading">
                <div>
                  <span className="eyebrow">At a glance</span>
                  <h2>What you’re getting</h2>
                </div>
                <span className="detail-panel__tag">
                  <ShieldCheck size={13} /> Reviewed
                </span>
              </div>
              <div className="detail-stats">
                {product.stats.map((stat) => (
                  <div key={stat.label}>
                    <span>{stat.label}</span>
                    <strong
                      className={
                        stat.label === "Page quality" ? "is-quality" : ""
                      }
                    >
                      {stat.value}
                    </strong>
                  </div>
                ))}
                <div>
                  <span>Handover</span>
                  <strong>
                    {product.deliveryWindow.replace(" handover", "")}
                  </strong>
                </div>
              </div>
            </section>

            <section className="detail-panel glass-panel">
              <div className="detail-panel__heading">
                <div>
                  <span className="eyebrow">Included with this listing</span>
                  <h2>Simple, direct ownership</h2>
                </div>
              </div>
              <ul className="detail-feature-list">
                {product.features.map((feature) => (
                  <li key={feature.label}>
                    <span className="detail-feature-list__icon">
                      <FeatureIcon icon={feature.icon} />
                    </span>
                    <span>
                      <strong>{feature.label}</strong>
                      <small>Included in the purchase handover</small>
                    </span>
                    <CheckCircle2 size={16} />
                  </li>
                ))}
              </ul>
            </section>

            <section className="detail-panel glass-panel">
              <div className="detail-panel__heading">
                <div>
                  <span className="eyebrow">How it works</span>
                  <h2>From listing to ownership</h2>
                </div>
              </div>
              <div className="handover-steps">
                <HandoverStep
                  number="01"
                  title="Review the listing"
                  detail="Confirm the page, price, stats, and included features."
                />
                <HandoverStep
                  number="02"
                  title="Choose your payment"
                  detail="Use the mock checkout to select KHQR/Bakong, card, or USDT."
                />
                <HandoverStep
                  number="03"
                  title="Receive admin access"
                  detail="After payment confirmation, the seller completes the direct handover."
                />
              </div>
            </section>

            <div className="detail-support">
              <span className="detail-support__icon">
                <MessageCircle size={17} />
              </span>
              <span>
                <strong>Need more information?</strong>
                <small>
                  Seller messaging will be available with the live marketplace.
                </small>
              </span>
              <span className="detail-support__status">Coming soon</span>
            </div>
          </div>

          <aside className="product-purchase-card glass-panel-elevated">
            <span className="eyebrow">
              <LockKeyhole size={13} /> Checkout preview
            </span>
            <h2>Ready to make it yours?</h2>
            <p>
              Review the total and explore your preferred payment method. This
              demo does not reserve the listing.
            </p>
            <div className="product-purchase-card__price">
              <span>Total price</span>
              <strong>
                {formatProductPrice(product.priceUsd, product.currency)}
              </strong>
            </div>
            <Link
              className="button button--primary button--wide"
              href={`/marketplace/${product.id}/checkout`}
            >
              Continue to checkout <ArrowRight size={15} />
            </Link>
            <div className="product-purchase-card__note">
              <CheckCircle2 size={14} /> No payment is captured in this demo.
            </div>
            <div className="product-purchase-card__trust">
              <span>
                <ShieldCheck size={14} /> Reviewed listing
              </span>
              <span>
                <KeyRound size={14} /> Direct admin transfer
              </span>
              <span>
                <Clock3 size={14} /> {product.deliveryWindow}
              </span>
            </div>
            <Link className="product-purchase-card__back" href="/marketplace">
              <ArrowLeft size={13} /> Browse other listings
            </Link>
          </aside>
        </div>
      </div>
      <div className="mobile-purchase-bar">
        <span>
          <small>Total price</small>
          <strong>
                {formatProductPrice(product.priceUsd, product.currency)}
          </strong>
        </span>
        <Link
          className="button button--primary"
          href={`/marketplace/${product.id}/checkout`}
        >
          Checkout <ArrowRight size={16} />
        </Link>
      </div>
    </main>
  );
}

function HandoverStep({
  number,
  title,
  detail,
}: Readonly<{ number: string; title: string; detail: string }>) {
  return (
    <div className="handover-step">
      <span>{number}</span>
      <div>
        <strong>{title}</strong>
        <small>{detail}</small>
      </div>
    </div>
  );
}
