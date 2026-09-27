# Our Sell frontend

The frontend for the Our Sell marketplace and authentication API. It uses Next.js App Router, React 19, TypeScript, Tailwind CSS v4, and Lucide.

## Run locally

```bash
npm install
cp .env.example .env.local
npm run dev
```

The Go API should be running at `http://localhost:8080`. `NEXT_PUBLIC_API_URL` can point to another API origin when needed.

## Architecture

- `app/` contains route entry points and the root layout.
- `components/auth/` contains reusable auth shell and auth forms.
- `components/dashboard/` contains the responsive account workspace and authenticated Marketplace view.
- `components/marketplace/` contains the public storefront shell.
- `components/providers/` owns auth restoration and theme state.
- `components/ui/` contains small reusable controls.
- `lib/marketplace.ts` maps published Facebook Page listings from the API into the storefront product contract. `lib/products.ts` contains shared product labels, pricing, and the future product-type roadmap.
- `lib/api.ts` is the typed API boundary. Access tokens stay in memory; refresh tokens remain in the backend’s HttpOnly cookie.

The public storefront is available at `/` and `/marketplace`. It reads published Page listings from the Go API; quick view leads to `/marketplace/[id]` and `/marketplace/[id]/checkout`. Checkout is wallet-first: a signed-in buyer must have enough available USD, then the server atomically holds the funds and creates a real 15-minute listing reservation. If the balance is insufficient, the buyer is sent to Wallet to top up with Bakong and returned to checkout after confirmation. Direct Bakong/KHQR order checkout, card, USDT, and ownership transfer remain provider/operations work.

## UI design system

`IOS26_UI_DESIGN_GUIDE.md` is the visual reference. Shared tokens and glass primitives live in `app/globals.css`; storefront, detail, and checkout layouts live in `app/marketplace.css`. Fonts are self-hosted through `next/font`: Plus Jakarta Sans and JetBrains Mono. Keep Safari-prefixed backdrop filters **before** their standard declarations so the CSS optimizer preserves Chromium support.

Floating headers and the mobile dock use elevated glass. Quick view uses a native modal dialog with a scrollable body and persistent actions. Readable dark-theme captions, 44px controls, reduced-motion support, and solid-surface fallbacks take priority over decorative transparency.

## Verify

```bash
npm run typecheck
npm run lint
npm run build
```
