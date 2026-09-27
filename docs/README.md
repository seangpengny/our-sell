# Our Sell documentation

This folder records implementation work, product decisions, and the remaining marketplace roadmap.

## Documents

- [`IMPLEMENTATION_LOG.md`](./IMPLEMENTATION_LOG.md) — chronological record of completed tasks and verification.
- [`MARKETPLACE_ROADMAP.md`](./MARKETPLACE_ROADMAP.md) — current marketplace lifecycle and next implementation phases.
- [`WALLET_BAKONG.md`](./WALLET_BAKONG.md) — wallet schema, Bakong configuration, payment flow, statuses, and production checklist.

## Current local workflow

1. Connect a Facebook account from the admin console.
2. Sync its Pages from **Page management**.
3. Create a listing and publish it.
4. A customer opens the public listing and submits contact details.
5. A signed-in buyer must have enough USD wallet balance to continue.
6. The API atomically holds the wallet funds, creates the order, and reserves the listing for 15 minutes.
7. If the balance is insufficient, the buyer tops up with Bakong and returns to checkout after confirmation.
8. Direct Bakong/card/USDT order checkout and the final Page transfer still require their provider and operational workflows.
