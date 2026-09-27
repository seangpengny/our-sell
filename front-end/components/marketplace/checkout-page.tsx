"use client";

import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import {
  ArrowLeft,
  ArrowRight,
  Check,
  CheckCircle2,
  ChevronDown,
  CircleHelp,
  CreditCard,
  Landmark,
  LockKeyhole,
  QrCode,
  ShieldCheck,
  WalletCards,
} from "lucide-react";
import { MarketplaceHeader } from "@/components/marketplace/marketplace-header";
import { PRODUCT_KIND_LABELS, formatProductPrice } from "@/lib/products";
import { ApiError, marketplaceApi, walletApi, type MarketplaceOrder } from "@/lib/api";
import type { Product, Wallet } from "@/lib/types";

import { Field } from "@/components/ui/field";
import { useAuth } from "@/components/providers/auth-provider";

type CheckoutStep = 1 | 2 | 3;
type PaymentMethod = "wallet" | "bakong" | "card" | "usdt";
type CustomerForm = {
  buyerNote: string;
  termsAccepted: boolean;
};

export function CheckoutPage({ product }: Readonly<{ product: Product }>) {
  const [step, setStep] = useState<CheckoutStep>(1);
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>("wallet");
  const [customer, setCustomer] = useState<CustomerForm>({
    buyerNote: "",
    termsAccepted: false,
  });
  const { status: authStatus, user } = useAuth();
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [reservationError, setReservationError] = useState("");
  const [creatingOrder, setCreatingOrder] = useState(false);
  const [order, setOrder] = useState<MarketplaceOrder | null>(null);
  const [orderCreated, setOrderCreated] = useState(false);
  const [wallet, setWallet] = useState<Wallet | null>(null);
  const [walletLoading, setWalletLoading] = useState(true);
  const [walletError, setWalletError] = useState("");
  const usdTotal = useMemo(() => product.priceUsd, [product.priceUsd]);
  const contentRef = useRef<HTMLDivElement>(null);
  const previousStep = useRef(step);
  useEffect(() => {
    if (previousStep.current === step) return;
    previousStep.current = step;
    const heading = contentRef.current?.querySelector("h1");
    if (heading) {
      heading.tabIndex = -1;
      heading.focus();
      heading.scrollIntoView({ block: "start" });
    }
  }, [step]);
  const orderReference = order?.reference ?? `OS-DEMO-${product.id.replace("page-", "")}`;
  const topUpHref = `/app?view=wallet&return_to=${encodeURIComponent(`/marketplace/${product.id}/checkout`)}`;

  const loadWallet = useCallback(async () => {
    setWalletLoading(true);
    setWalletError("");
    try {
      setWallet(await walletApi.wallet());
    } catch (reason) {
      setWalletError(reason instanceof ApiError ? reason.message : "Your wallet balance could not be loaded.");
    } finally {
      setWalletLoading(false);
    }
  }, []);

  useEffect(() => {
    if (authStatus !== "authenticated") return;
    const timeout = window.setTimeout(() => void loadWallet(), 0);
    return () => window.clearTimeout(timeout);
  }, [authStatus, loadWallet]);

  function updateCustomer(field: keyof CustomerForm, value: string | boolean) {
    setCustomer((current) => ({ ...current, [field]: value }));
    if (errors[field]) setErrors((current) => ({ ...current, [field]: "" }));
  }

  async function continueToPayment(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextErrors: Record<string, string> = {};
    if (!customer.termsAccepted)
      nextErrors.termsAccepted = "Accept the reservation terms to continue.";
    setErrors(nextErrors);
    if (Object.keys(nextErrors).length === 0) {
      setReservationError("");
      setStep(2);
    } else {
      const form = event.currentTarget;
      requestAnimationFrame(() =>
        form
          .querySelector<HTMLInputElement>(
            `[name="${Object.keys(nextErrors)[0]}"]`,
          )
          ?.focus(),
      );
    }
  }

  if (authStatus === "loading") {
    return <MarketplaceAuthGate productId={product.id} loading />;
  }
  if (authStatus === "unauthenticated") {
    return <MarketplaceAuthGate productId={product.id} />;
  }

  async function confirmPayment() {
    if (order) {
      setOrderCreated(true);
      setStep(3);
      return;
    }
    if (!wallet || wallet.available_amount < usdTotal) {
      setReservationError("Your wallet balance is not sufficient. Top up your wallet before creating this order.");
      return;
    }
    setCreatingOrder(true);
    setReservationError("");
    try {
      const created = await marketplaceApi.createOrder(product.id, {
        buyer_note: customer.buyerNote,
        payment_method: paymentMethod,
        terms_accepted: customer.termsAccepted,
      });
      setOrder(created);
      setOrderCreated(true);
      setStep(3);
    } catch (reason) {
      if (reason instanceof ApiError && reason.code === "INSUFFICIENT_FUNDS") {
        setReservationError("Your available balance changed. Top up your wallet before creating this order.");
        void loadWallet();
      } else {
        setReservationError(reason instanceof ApiError ? reason.message : "This listing could not be reserved.");
      }
    } finally {
      setCreatingOrder(false);
    }
  }

  return (
    <main className="public-marketplace checkout-page">
      <div className="app-page__ambient app-page__ambient--one" />
      <div className="app-page__ambient app-page__ambient--two" />
      <MarketplaceHeader />
      <div className="checkout-content" ref={contentRef}>
        <div className="checkout-breadcrumb">
          <Link href={`/marketplace/${product.id}`}>
            <ArrowLeft size={14} /> Back to {product.title}
          </Link>
          <span>Checkout</span>
        </div>
        <div className="demo-banner">
          <CircleHelp size={16} />
          <span>
            Demo checkout · Use test details only. No payment will be collected.
          </span>
        </div>
        <CheckoutProgress step={step} />
        {step !== 3 && <MobileOrderSummary product={product} />}
        {step === 3 && orderCreated ? (
          <CheckoutSuccess
            product={product}
            orderReference={orderReference}
            paymentMethod={paymentMethod}
            order={order}
          />
        ) : (
          <div className="checkout-layout">
            <section className="checkout-main">
              {step === 1 ? (
                <CustomerStep
                  customer={customer}
                  errors={errors}
                  onUpdate={updateCustomer}
                  onSubmit={continueToPayment}
                  product={product}
                  loading={false}
                  error={reservationError}
                  userEmail={user?.email ?? ""}
                />
              ) : (
                <PaymentStep
                  paymentMethod={paymentMethod}
                  onPaymentMethodChange={setPaymentMethod}
                  usdTotal={usdTotal}
                  onBack={() => setStep(1)}
                  onConfirm={confirmPayment}
                  loading={creatingOrder}
                  error={reservationError}
                  wallet={wallet}
                  walletLoading={walletLoading}
                  walletError={walletError}
                  topUpHref={topUpHref}
                />
              )}
            </section>
            <CheckoutSummary
              product={product}
              step={step}
              paymentMethod={paymentMethod}
              customer={customer}
            />
          </div>
        )}
      </div>
    </main>
  );
}

function CheckoutProgress({ step }: Readonly<{ step: CheckoutStep }>) {
  return (
    <div className="checkout-progress" aria-label="Checkout progress">
      <ProgressStep
        number="01"
        label="Your details"
        active={step === 1}
        complete={step > 1}
      />
      <span className="checkout-progress__line" />
      <ProgressStep
        number="02"
        label="Payment"
        active={step === 2}
        complete={step > 2}
      />
      <span className="checkout-progress__line" />
      <ProgressStep
        number="03"
        label="Complete"
        active={step === 3}
        complete={false}
      />
    </div>
  );
}

function ProgressStep({
  number,
  label,
  active,
  complete,
}: Readonly<{
  number: string;
  label: string;
  active: boolean;
  complete: boolean;
}>) {
  return (
    <div
      aria-current={active ? "step" : undefined}
      className={`checkout-progress__step ${active ? "is-active" : ""} ${complete ? "is-complete" : ""}`}
    >
      <span>{complete ? <Check size={12} /> : number}</span>
      <small>{label}</small>
    </div>
  );
}

function MarketplaceAuthGate({
  productId,
  loading = false,
}: Readonly<{ productId: string; loading?: boolean }>) {
  const checkoutPath = `/marketplace/${productId}/checkout`;
  return (
    <main className="public-marketplace checkout-page">
      <div className="app-page__ambient app-page__ambient--one" />
      <div className="app-page__ambient app-page__ambient--two" />
      <MarketplaceHeader />
      <div className="checkout-content">
        <div className="checkout-step-card glass-panel checkout-auth-gate">
          <span className="eyebrow"><LockKeyhole size={13} /> Secure checkout</span>
          <h1>{loading ? "Checking your session…" : "Sign in to continue"}</h1>
          <p>
            You must be signed in before reserving a listing. Your order will
            be linked to your Our Sell account.
          </p>
          {!loading && (
            <Link
              className="button button--primary"
              href={`/sign-in?next=${encodeURIComponent(checkoutPath)}`}
            >
              Sign in to continue <ArrowRight size={15} />
            </Link>
          )}
        </div>
      </div>
    </main>
  );
}

function CustomerStep({
  customer,
  errors,
  onUpdate,
  onSubmit,
  product,
  loading,
  error,
  userEmail,
}: Readonly<{
  customer: CustomerForm;
  errors: Record<string, string>;
  onUpdate: (field: keyof CustomerForm, value: string | boolean) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void | Promise<void>;
  product: Product;
  loading: boolean;
  error: string;
  userEmail: string;
}>) {
  return (
    <div className="checkout-step-card glass-panel">
      <div className="checkout-step-card__heading">
        <span className="eyebrow">
          <LockKeyhole size={13} /> Step 1 of 2
        </span>
        <h1>Handover details</h1>
        <p>
          Your account is attached to this order. Add any information the
          seller or admin should know for the handover.
        </p>
      </div>
      <form className="checkout-form" onSubmit={onSubmit} noValidate>
        <div className="checkout-account-note">
          Signed in as <strong>{userEmail}</strong>
        </div>
        <label className="checkout-textarea-field">
          <span>Buyer note <small>Optional</small></span>
          <textarea
            name="buyerNote"
            maxLength={2000}
            placeholder="For example: your Facebook profile URL, Meta Business ID, or handover instructions."
            value={customer.buyerNote}
            onChange={(event) => onUpdate("buyerNote", event.target.value)}
          />
          {errors.buyerNote && <small className="checkout-error">{errors.buyerNote}</small>}
        </label>
        <label
          className={`checkout-checkbox ${errors.termsAccepted ? "has-error" : ""}`}
        >
          <input
            type="checkbox"
            name="termsAccepted"
            aria-invalid={Boolean(errors.termsAccepted)}
            aria-describedby={errors.termsAccepted ? "terms-error" : undefined}
            checked={customer.termsAccepted}
            onChange={(event) =>
              onUpdate("termsAccepted", event.target.checked)
            }
          />
          <span>
            I understand the order is created only after the server confirms
            enough wallet funds, and the listing is then reserved for handover.
          </span>
        </label>
        {errors.termsAccepted && (
          <small
            id="terms-error"
            className="checkout-error checkout-checkbox__error"
          >
            {errors.termsAccepted}
          </small>
        )}
        {error && <div className="checkout-form__notice checkout-form__notice--error" role="alert"><CircleHelp size={16} /><span>{error}</span></div>}
        <div className="checkout-form__notice">
          <ShieldCheck size={16} />
          <span>
            <strong>Reservation preview.</strong>
            <small>
              The listing is reserved for 15 minutes after your wallet funds are held.
            </small>
          </span>
        </div>
        <div className="checkout-step-actions">
          <Link
            className="button button--secondary"
            href={`/marketplace/${product.id}`}
          >
            <ArrowLeft size={14} /> Back to details
          </Link>
          <button className="button button--primary" type="submit" disabled={loading} aria-busy={loading}>
            {loading ? "Continuing…" : <>Continue to payment <ArrowRight size={15} /></>}
          </button>
        </div>
      </form>
    </div>
  );
}

function CheckoutField({
  label,
  name,
  type = "text",
  placeholder,
  value,
  hint,
  error,
  onChange,
}: Readonly<{
  label: string;
  name: string;
  type?: string;
  placeholder: string;
  value: string;
  hint?: string;
  error?: string;
  onChange: (value: string) => void;
}>) {
  return (
    <Field
      label={label}
      name={name}
      type={type}
      placeholder={placeholder}
      value={value}
      hint={hint}
      error={error}
      autoComplete={
        name === "email" ? "email" : name === "name" ? "name" : "off"
      }
      onChange={(event) => onChange(event.target.value)}
    />
  );
}

function PaymentStep({
  paymentMethod,
  onPaymentMethodChange,
  usdTotal,
  onBack,
  onConfirm,
  loading,
  error,
  wallet,
  walletLoading,
  walletError,
  topUpHref,
}: Readonly<{
  paymentMethod: PaymentMethod;
  onPaymentMethodChange: (method: PaymentMethod) => void;
  usdTotal: number;
  onBack: () => void;
  onConfirm: () => void;
  loading: boolean;
  error: string;
  wallet: Wallet | null;
  walletLoading: boolean;
  walletError: string;
  topUpHref: string;
}>) {
  return (
    <div className="checkout-step-card glass-panel">
      <div className="checkout-step-card__heading">
        <span className="eyebrow">
          <WalletCards size={13} /> Step 2 of 2
        </span>
        <h1>Choose how to pay</h1>
        <p>
          Page orders currently use your USD wallet balance. Top up with
          Bakong first if your available balance is too low.
        </p>
      </div>
      <div
        className="payment-methods"
        role="radiogroup"
        aria-label="Payment methods"
      >
        <PaymentOption
          method="wallet"
          selected={paymentMethod}
          onSelect={onPaymentMethodChange}
          icon={<WalletCards size={18} />}
          title="Wallet balance"
          description="Use your available USD balance"
        />
        <PaymentOption
          method="bakong"
          selected={paymentMethod}
          onSelect={onPaymentMethodChange}
          icon={<QrCode size={18} />}
          title="Bakong / KHQR"
          description="Top up your wallet first"
          badge="Top up only"
          disabled
        />
        <PaymentOption
          method="card"
          selected={paymentMethod}
          onSelect={onPaymentMethodChange}
          icon={<CreditCard size={18} />}
          title="Card payment"
          description="Available in a future release"
          badge="Coming soon"
          disabled
        />
        <PaymentOption
          method="usdt"
          selected={paymentMethod}
          onSelect={onPaymentMethodChange}
          icon={<Landmark size={18} />}
          title="USDT"
          description="Available in a future release"
          badge="Coming soon"
          disabled
        />
      </div>
      {error && <div className="notice notice--error" role="alert">{error}</div>}
      {paymentMethod === "wallet" && <WalletPayment usdTotal={usdTotal} wallet={wallet} loading={walletLoading} error={walletError} topUpHref={topUpHref} />}
      {paymentMethod === "bakong" && <BakongPayment usdTotal={usdTotal} />}
      {paymentMethod === "card" && <CardPayment />}
      {paymentMethod === "usdt" && <UsdtPayment />}
      <div className="checkout-step-actions">
        <button
          className="button button--secondary"
          type="button"
          onClick={onBack}
        >
          <ArrowLeft size={14} /> Back to details
        </button>
        <button
          className="button button--primary"
          type="button"
          onClick={onConfirm}
          disabled={loading || walletLoading || Boolean(walletError) || !wallet || wallet.available_amount < usdTotal}
        >
          {loading ? "Creating order…" : "Create order with wallet"} <ArrowRight size={15} />
        </button>
      </div>
    </div>
  );
}

function WalletPayment({ usdTotal, wallet, loading, error, topUpHref }: Readonly<{ usdTotal: number; wallet: Wallet | null; loading: boolean; error: string; topUpHref: string }>) {
  const available = wallet?.available_amount ?? 0;
  const enough = available >= usdTotal;
  return (
    <div className="payment-panel">
      <div className="payment-panel__header"><span><WalletCards size={16} /> Wallet payment</span><small>Server verified</small></div>
      {loading ? <p>Checking your available USD balance…</p> : error ? <div className="payment-panel__notice"><CircleHelp size={14} /><span>{error}</span></div> : <>
        <div className="wallet-checkout-balance"><span>Available balance</span><strong>{formatProductPrice(available, "USD")}</strong><small>Required: {formatProductPrice(usdTotal, "USD")}</small></div>
        {enough ? <p>When you create the order, {formatProductPrice(usdTotal, "USD")} moves to held funds while the Page handover is completed.</p> : <div className="payment-panel__notice payment-panel__notice--warning"><CircleHelp size={14} /><span>Your balance is short by {formatProductPrice(usdTotal - available, "USD")}.</span><Link className="button button--secondary" href={topUpHref}>Top up wallet <ArrowRight size={14} /></Link></div>}
      </>}
    </div>
  );
}

function PaymentOption({
  method,
  selected,
  onSelect,
  icon,
  title,
  description,
  badge,
  disabled = false,
}: Readonly<{
  method: PaymentMethod;
  selected: PaymentMethod;
  onSelect: (method: PaymentMethod) => void;
  icon: React.ReactNode;
  title: string;
  description: string;
  badge?: string;
  disabled?: boolean;
}>) {
  return (
    <label
      className={`payment-option ${selected === method ? "is-selected" : ""}`}
    >
      <input
        type="radio"
        name="payment-method"
        value={method}
        checked={selected === method}
        disabled={disabled}
        onChange={() => onSelect(method)}
      />
      <span className="payment-option__icon">{icon}</span>
      <span className="payment-option__copy">
        <strong>{title}</strong>
        <small>{description}</small>
      </span>
      {badge && <span className="payment-option__badge">{badge}</span>}
      <span className="payment-option__radio" />
    </label>
  );
}

function BakongPayment({ usdTotal }: Readonly<{ usdTotal: number }>) {
  return (
    <div className="payment-panel payment-panel--bakong">
      <div className="payment-panel__header">
        <span>
          <QrCode size={16} /> KHQR payment
        </span>
        <small>Mock QR</small>
      </div>
      <div className="bakong-payment">
        <div
          className="khqr-placeholder"
          aria-label="Demo QR placeholder, not scannable"
        >
          <QrCode size={80} strokeWidth={1.25} aria-hidden="true" />
          <strong>Demo QR</strong>
          <small>Not scannable</small>
        </div>
        <div className="bakong-payment__copy">
          <span className="eyebrow">Amount to pay</span>
          <strong>{formatProductPrice(usdTotal, "USD")}</strong>
          <small>Charged in USD; the final KHQR amount will be generated at payment time.</small>
          <p>
            In production, scan this KHQR with Bakong or a supported Cambodian
            banking app.
          </p>
        </div>
      </div>
      <div className="payment-panel__notice">
        <CircleHelp size={14} />
        <span>
          Wallet top-ups use the real Bakong verification flow from the Wallet
          page. This checkout option is reserved for a future direct payment.
        </span>
      </div>
    </div>
  );
}

function CardPayment() {
  const [card, setCard] = useState({
    name: "",
    number: "",
    expiry: "",
    cvc: "",
  });
  function updateCard(field: keyof typeof card, value: string) {
    setCard((current) => ({ ...current, [field]: value }));
  }

  return (
    <div className="payment-panel">
      <div className="payment-panel__header">
        <span>
          <CreditCard size={16} /> Card details
        </span>
        <small>Mock form</small>
      </div>
      <div className="card-payment-form">
        <CheckoutField
          label="Name on card"
          name="card-name"
          placeholder="Cardholder name"
          value={card.name}
          onChange={(value) => updateCard("name", value)}
        />
        <CheckoutField
          label="Card number"
          name="card-number"
          placeholder="1234 5678 9012 3456"
          value={card.number}
          onChange={(value) => updateCard("number", value)}
        />
        <div className="checkout-form__grid">
          <CheckoutField
            label="Expiry"
            name="card-expiry"
            placeholder="MM / YY"
            value={card.expiry}
            onChange={(value) => updateCard("expiry", value)}
          />
          <CheckoutField
            label="CVC"
            name="card-cvc"
            placeholder="123"
            value={card.cvc}
            onChange={(value) => updateCard("cvc", value)}
          />
        </div>
      </div>
      <div className="payment-panel__notice">
        <LockKeyhole size={14} />
        <span>
          Use test details only. Nothing is submitted or saved. Live card
          payments will use a payment provider’s secure form.
        </span>
      </div>
    </div>
  );
}

function UsdtPayment() {
  return (
    <div className="payment-panel">
      <div className="payment-panel__header">
        <span>
          <Landmark size={16} /> USDT transfer
        </span>
        <small>Mock wallet</small>
      </div>
      <div className="usdt-payment">
        <div>
          <span className="eyebrow">Network</span>
          <strong>TRC20</strong>
          <small>Use the exact network shown after the live integration.</small>
        </div>
        <div className="usdt-payment__address">
          <span>Wallet address</span>
          <code>TX8m...OURSELLDEMO</code>
        </div>
      </div>
      <div className="payment-panel__notice">
        <CircleHelp size={14} />
        <span>
          Wallet address and blockchain confirmation will be supplied by the
          payment service later.
        </span>
      </div>
    </div>
  );
}

function CheckoutSummary({
  product,
  step,
  paymentMethod,
  customer,
}: Readonly<{
  product: Product;
  step: CheckoutStep;
  paymentMethod: PaymentMethod;
  customer: CustomerForm;
}>) {
  return (
    <aside className="checkout-summary glass-panel">
      <div className="checkout-summary__heading">
        <span className="eyebrow">Order summary</span>
        <span className="checkout-summary__secure">
          <LockKeyhole size={12} /> Demo
        </span>
      </div>
      <div className="checkout-summary__product">
        <span
          className={`product-card__avatar product-card__avatar--small product-card__avatar--${product.accent}`}
        >
          {product.initials}
        </span>
        <span>
          <strong>{product.title}</strong>
          <small>
            {PRODUCT_KIND_LABELS[product.kind]} · ID{" "}
            {product.id.replace("page-", "")}
          </small>
        </span>
      </div>
      <div className="checkout-summary__line">
        <span>Listing price</span>
        <strong>
          {formatProductPrice(product.priceUsd, product.currency)}
        </strong>
      </div>
      <div className="checkout-summary__line">
        <span>Service fee</span>
        <strong>Included</strong>
      </div>
      <div className="checkout-summary__total">
        <span>Total</span>
        <strong>
          {formatProductPrice(product.priceUsd, product.currency)}
        </strong>
      </div>
      {step === 2 && (
        <div className="checkout-summary__method">
          <span>Selected payment</span>
          <strong>
            {paymentMethod === "wallet"
              ? "Wallet balance"
              : paymentMethod === "bakong"
              ? "Bakong / KHQR"
              : paymentMethod === "card"
                ? "Card payment"
                : "USDT"}
          </strong>
        </div>
      )}
      {customer.buyerNote && (
        <div className="checkout-summary__contact">
          <CheckCircle2 size={13} /> Buyer note added
        </div>
      )}
      <div className="checkout-summary__trust">
        <span>
          <ShieldCheck size={14} /> Reviewed listing
        </span>
        <span>
          <CheckCircle2 size={14} /> Direct admin transfer
        </span>
      </div>
    </aside>
  );
}

function CheckoutSuccess({
  product,
  orderReference,
  paymentMethod,
  order,
}: Readonly<{
  product: Product;
  orderReference: string;
  paymentMethod: PaymentMethod;
  order: MarketplaceOrder | null;
}>) {
  return (
    <section className="checkout-success glass-panel-elevated">
      <span className="checkout-success__icon">
        <CheckCircle2 size={30} />
      </span>
      <span className="eyebrow">
        <CheckCircle2 size={13} /> Listing reserved
      </span>
      <h1>Your reservation is ready.</h1>
      <p>
        Your order has been recorded and the listing is reserved temporarily.
        {paymentMethod === "wallet"
          ? " Your wallet funds are held until the Page handover is completed."
          : " Payment and Page transfer will continue after the provider integration is enabled."}
      </p>
      <div className="checkout-success__reference">
        <span>Order reference</span>
        <strong>{orderReference}</strong>
        <small>
          Payment method:{" "}
          {paymentMethod === "wallet"
            ? "Wallet balance"
            : paymentMethod === "bakong"
            ? "Bakong / KHQR"
            : paymentMethod === "card"
              ? "Card payment"
              : "USDT"}
        </small>
        {order && <small>Reservation expires: {new Date(order.reservation_expires_at).toLocaleString()}</small>}
      </div>
      <div className="checkout-success__timeline">
        <span>
          <b>01</b>
          <strong>Details received</strong>
          <small>Complete</small>
        </span>
        <span>
          <b>02</b>
          <strong>Payment confirmation</strong>
          <small>{paymentMethod === "wallet" ? "Wallet funds held" : "Pending payment integration"}</small>
        </span>
        <span>
          <b>03</b>
          <strong>Admin handover</strong>
          <small>Starts after confirmation</small>
        </span>
      </div>
      <div className="checkout-success__actions">
        <Link
          className="button button--primary"
          href={`/marketplace/${product.id}`}
        >
          View listing <ArrowRight size={15} />
        </Link>
        <Link className="button button--secondary" href="/marketplace">
          Browse marketplace
        </Link>
      </div>
    </section>
  );
}

function MobileOrderSummary({ product }: Readonly<{ product: Product }>) {
  return (
    <details className="checkout-mobile-summary">
      <summary>
        <span>
          Order summary <ChevronDown size={16} />
        </span>
        <strong>
          {formatProductPrice(product.priceUsd, product.currency)}
        </strong>
      </summary>
      <div>
        <p>
          <strong>{product.title}</strong>
          <small>
            {PRODUCT_KIND_LABELS[product.kind]} · #
            {product.id.replace("page-", "")}
          </small>
        </p>
        <div className="checkout-summary__line">
          <span>Listing price</span>
          <strong>
            {formatProductPrice(product.priceUsd, product.currency)}
          </strong>
        </div>
        <div className="checkout-summary__line">
          <span>Service fee</span>
          <strong>Included</strong>
        </div>
      </div>
    </details>
  );
}
