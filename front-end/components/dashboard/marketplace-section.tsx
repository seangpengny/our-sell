"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import {
  ArrowRight,
  Check,
  CheckCircle2,
  Clock3,
  Camera,
  KeyRound,
  Search,
  ShieldCheck,
  ShoppingBag,
  Sparkles,
  X,
  Video,
} from "lucide-react";
import {
  PRODUCT_KIND_LABELS,
  PRODUCT_ROADMAP,
  formatProductPrice,
} from "@/lib/products";
import { getMarketplaceListings } from "@/lib/marketplace";
import type { Product, ProductKind } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import { ProductFeatures } from "@/components/marketplace/product-features";

type MarketplaceFilter = "all" | ProductKind;
type SortOrder = "recommended" | "price-asc" | "price-desc";

export function MarketplaceSection() {
  const [filter, setFilter] = useState<MarketplaceFilter>("all");
  const [search, setSearch] = useState("");
  const [sort, setSort] = useState<SortOrder>("recommended");
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");

  useEffect(() => {
    let active = true;
    void getMarketplaceListings()
      .then((nextProducts) => {
        if (!active) return;
        setProducts(nextProducts);
        setLoadError("");
      })
      .catch(() => {
        if (active) setLoadError("Listings are temporarily unavailable. Please try again shortly.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  const liveProducts = useMemo(
    () =>
      products.filter(
        (product) => product.published && product.status === "available",
      ),
    [products],
  );
  const filters = useMemo<Array<{ id: MarketplaceFilter; label: string }>>(
    () => [
      { id: "all", label: "All listings" },
      ...Array.from(new Set(liveProducts.map((product) => product.kind))).map(
        (kind) => ({ id: kind, label: PRODUCT_KIND_LABELS[kind] }),
      ),
    ],
    [liveProducts],
  );
  const visibleProducts = useMemo(() => {
    const query = search.trim().toLowerCase();
    return liveProducts
      .filter(
        (product) =>
          (filter === "all" || product.kind === filter) &&
          (!query ||
            `${product.title} ${product.subtitle} ${product.id}`
              .toLowerCase()
              .includes(query)),
      )
      .sort((a, b) =>
        sort === "price-asc"
          ? a.priceUsd - b.priceUsd
          : sort === "price-desc"
            ? b.priceUsd - a.priceUsd
            : a.sortOrder - b.sortOrder,
      );
  }, [filter, liveProducts, search, sort]);

  return (
    <section className="dashboard-section marketplace-section animate-fade-up">
      <div className="section-intro marketplace-intro">
        <div>
          <span className="eyebrow">
            <ShoppingBag size={14} /> The Our Sell marketplace
          </span>
          <h1>A new start. Already built.</h1>
          <p>
            Discover digital pages and properties, with clear pricing and a
            direct handover.
          </p>
        </div>
        <span className="catalog-status">
          <span /> {loading ? "Loading listings…" : `${liveProducts.length} available listings`}
        </span>
      </div>
      <div className="marketplace-benefits" aria-label="Buying benefits">
        <span>
          <ShieldCheck size={18} />
          <span>
            <strong>Clear listing details</strong>
            <small>Know what’s included</small>
          </span>
        </span>
        <span>
          <KeyRound size={18} />
          <span>
            <strong>Direct handover</strong>
            <small>A simple ownership transfer</small>
          </span>
        </span>
        <span>
          <Clock3 size={18} />
          <span>
            <strong>Your way to pay</strong>
            <small>Preview KHQR, card & USDT</small>
          </span>
        </span>
      </div>
      <div className="marketplace-toolbar">
        <div className="filter-pills" aria-label="Product filters">
          {filters.map((item) => (
            <button
              key={item.id}
              type="button"
              className={`filter-pill ${filter === item.id ? "is-active" : ""}`}
              aria-pressed={filter === item.id}
              onClick={() => setFilter(item.id)}
            >
              {item.label}
              <b>
                {item.id === "all"
                  ? liveProducts.length
                  : liveProducts.filter((product) => product.kind === item.id)
                      .length}
              </b>
            </button>
          ))}
        </div>
        <label className="catalog-search">
          <Search size={18} />
          <span className="sr-only">Search listings</span>
          <input
            type="search"
            placeholder="Search name or listing ID"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </label>
      </div>
      <div className="marketplace-listing-header">
        <h2>
          Explore listings{" "}
          <span aria-live="polite">{visibleProducts.length}</span>
        </h2>
        <label className="catalog-sort">
          <span>Sort by</span>
          <select
            aria-label="Sort listings"
            value={sort}
            onChange={(event) => setSort(event.target.value as SortOrder)}
          >
            <option value="recommended">Recommended</option>
            <option value="price-asc">Price: low to high</option>
            <option value="price-desc">Price: high to low</option>
          </select>
        </label>
      </div>
      {loadError ? (
        <div className="marketplace-empty glass-panel" role="alert">
          <Search size={28} />
          <strong>Marketplace unavailable</strong>
          <span>{loadError}</span>
        </div>
      ) : loading ? (
        <div className="marketplace-empty glass-panel" role="status">
          <span className="loading-orb loading-orb--small" />
          <strong>Loading available listings…</strong>
        </div>
      ) : visibleProducts.length > 0 ? (
        <div className="product-grid">
          {visibleProducts.map((product) => (
            <ProductCard
              key={product.id}
              product={product}
              onQuickView={setSelectedProduct}
            />
          ))}
        </div>
      ) : (
        <div className="marketplace-empty glass-panel" role="status">
          <Search size={28} />
          <strong>No matching listings</strong>
          <span>
            Try another name or clear your filters to see all available
            products.
          </span>
          <Button
            variant="secondary"
            onClick={() => {
              setSearch("");
              setFilter("all");
            }}
          >
            Clear filters
          </Button>
        </div>
      )}
      <div className="roadmap-panel glass-panel">
        <div className="roadmap-panel__heading">
          <span className="eyebrow">
            <Sparkles size={14} /> Coming next
          </span>
          <h3>Room for more possibilities.</h3>
          <p>
            More types of digital products are on the way. Explore what’s next
            for the marketplace.
          </p>
        </div>
        <div className="roadmap-list">
          {PRODUCT_ROADMAP.map((item) => (
            <div className="roadmap-item" key={item.kind}>
              <span className="roadmap-item__icon">
                <RoadmapIcon kind={item.kind} />
              </span>
              <span>
                <strong>{item.label}</strong>
                <small>{item.detail}</small>
              </span>
              <span className="roadmap-item__status">Coming soon</span>
            </div>
          ))}
        </div>
      </div>
      <p className="storefront-disclaimer">
        You’re exploring a demo storefront. Listings and checkout are previews;
        no money is collected.
      </p>
      {selectedProduct && (
        <QuickViewDialog
          product={selectedProduct}
          onClose={() => setSelectedProduct(null)}
        />
      )}
    </section>
  );
}

function ProductCard({
  product,
  onQuickView,
}: Readonly<{ product: Product; onQuickView: (product: Product) => void }>) {
  return (
    <article className="product-card glass-panel">
      <div
        className={`product-card__visual product-card__visual--${product.accent}`}
      >
        <span className="product-card__avatar">{product.initials}</span>
        <span className="product-card__kind">
          {PRODUCT_KIND_LABELS[product.kind]}
        </span>
        <span className="product-card__verified">
          <Check size={13} /> Available
        </span>
      </div>
      <div className="product-card__body">
        <div className="product-card__heading">
          <div>
            <h3>
              <Link href={`/marketplace/${product.id}`}>{product.title}</Link>
              <CheckCircle2 size={16} />
            </h3>
            <p>{product.subtitle}</p>
          </div>
          <strong className="product-card__price">
            {formatProductPrice(product.priceUsd, product.currency)}
          </strong>
        </div>
        <div className="product-card__stats">
          {product.stats.map((stat) => (
            <div className="product-stat" key={stat.label}>
              <span>{stat.label}</span>
              <strong
                className={stat.label === "Page quality" ? "is-quality" : ""}
              >
                {stat.value}
              </strong>
            </div>
          ))}
        </div>
        <ProductFeatures features={product.features} />
        <div className="product-card__footer">
          <span className="product-card__id">
            #{product.id.replace("page-", "")}
          </span>
          <div className="product-card__actions">
            <Button
              variant="ghost"
              aria-label={`Quick view ${product.title}`}
              onClick={() => onQuickView(product)}
            >
              Quick view
            </Button>
            <Link
              className="button button--primary"
              href={`/marketplace/${product.id}`}
            >
              View details <ArrowRight size={15} />
            </Link>
          </div>
        </div>
      </div>
    </article>
  );
}

function QuickViewDialog({
  product,
  onClose,
}: Readonly<{ product: Product; onClose: () => void }>) {
  return (
    <Modal
      className="quick-view-dialog"
      labelledBy="quick-view-dialog-title"
      onClose={onClose}
    >
      <div className="quick-view-dialog__top">
        <span className="eyebrow">Listing preview</span>
        <button
          type="button"
          className="icon-button"
          aria-label="Close product preview"
          onClick={onClose}
          autoFocus
        >
          <X size={20} />
        </button>
      </div>
      <div className="quick-view-dialog__grid">
        <div
          className={`quick-view-dialog__visual product-card__visual--${product.accent}`}
        >
          <span className="quick-view-dialog__visual-label">
            {PRODUCT_KIND_LABELS[product.kind]}
          </span>
          <span className="product-card__avatar quick-view-dialog__avatar">
            {product.initials}
          </span>
          <span className="product-card__verified">
            <Check size={13} /> Available
          </span>
        </div>
        <div className="quick-view-dialog__content">
          <div className="quick-view-dialog__heading">
            <div>
              <h2 id="quick-view-dialog-title">{product.title}</h2>
              <span>{product.subtitle}</span>
            </div>
            <strong>
              {formatProductPrice(product.priceUsd, product.currency)}
            </strong>
          </div>
          <p className="quick-view-dialog__description">
            {product.description}
          </p>
          <div className="quick-view-dialog__stats">
            {product.stats.map((stat) => (
              <div key={stat.label}>
                <span>{stat.label}</span>
                <strong
                  className={stat.label === "Page quality" ? "is-quality" : ""}
                >
                  {stat.value}
                </strong>
              </div>
            ))}
          </div>
          <ProductFeatures
            features={product.features}
            className="quick-view-dialog__features"
          />
          <div className="quick-view-dialog__handover">
            <ShieldCheck size={18} />
            <span>
              <strong>Direct ownership transfer</strong>
              <small>{product.deliveryWindow}</small>
            </span>
          </div>
        </div>
      </div>
      <div className="quick-view-dialog__actions">
        <Link
          className="button button--secondary"
          href={`/marketplace/${product.id}`}
          onClick={onClose}
        >
          View full details
        </Link>
        <Link
          className="button button--primary"
          href={`/marketplace/${product.id}/checkout`}
          onClick={onClose}
        >
          Continue to checkout <ArrowRight size={16} />
        </Link>
      </div>
      <small className="quick-view-dialog__fineprint">
        Demo checkout · No payment will be collected
      </small>
    </Modal>
  );
}

function RoadmapIcon({ kind }: Readonly<{ kind: ProductKind }>) {
  if (kind === "youtube_channel") return <Video size={18} />;
  if (kind === "instagram_account") return <Camera size={18} />;
  return <Sparkles size={18} />;
}
