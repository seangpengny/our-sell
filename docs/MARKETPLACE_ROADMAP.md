# Marketplace roadmap

## Current lifecycle

```text
Facebook OAuth
    -> Page synchronization
    -> Admin listing draft
    -> Published listing
    -> Signed-in buyer wallet check
    -> Wallet funds held + order created
    -> Customer reservation (15 minutes)
    -> Transfer pending
    -> Completed
```

The current implementation completes the flow through the **wallet funds held + order created** step. A Bakong top-up is verified before the order is created; direct order payment and Page transfer are intentionally not automated yet.

## Recommended next phases

### Phase 1 — Operations dashboard

- Add an admin Orders view.
- Search by order reference, buyer account, Page name, and status.
- Add manual status transitions with audit history.
- Add a manual “release reservation” action.
- Add a background cleanup job for expired reservations.

### Phase 2 — Wallet and payment operations

- Add an admin order view for wallet holds, expirations, releases, and final spending.
- Add a scheduled cleanup worker so expired wallet holds are released promptly.
- Add reconciliation for Bakong top-ups and provider transaction references.
- If direct order payment is needed later, add it as a separate payment record and verify it server-side before moving an order to `paid`.

### Phase 3 — Handover

- Create a transfer task after payment succeeds.
- Record operator notes and evidence.
- Move the order through `transfer_pending` and `completed`.
- Use Meta’s official Page access workflow; never collect Facebook passwords or share raw access tokens.

### Phase 4 — Production hardening

- Add buyer and seller notifications.
- Add audit logs and support tooling.
- Add reconciliation for payments and transfers.
- Add database backups, monitoring, error alerts, and retention rules.
- Complete Meta production configuration, privacy documentation, and permission review.
