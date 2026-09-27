# Our Sell admin

The administrator console for the Our Sell marketplace. It is a separate Next.js app so customer-facing navigation and administrator access controls stay isolated.

## Run locally

```bash
npm install
cp .env.example .env.local
npm run dev
```

The app runs on the default Next.js port (`http://localhost:3000`) and expects the Go API at `http://localhost:8080`. Set `NEXT_PUBLIC_API_URL` when the API is hosted elsewhere.

Only accounts with the `admin` role can enter the console. The initial admin account can be promoted through PostgreSQL after registration:

```sql
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
```

## User management

The User management view supports name/email search, role filtering, pagination, email-verification status, and changing a user between `user` and `admin`. Role changes are protected again in the API with admin middleware, and an administrator cannot remove their own admin access.

## Facebook Page management

The Page management view stores Pages returned from connected Facebook accounts in PostgreSQL. It supports account filtering, search, pagination, manual synchronization, and a listing editor. Listings can remain drafts, be sent for review, or be published to the customer marketplace. Only published listings are returned by the public marketplace API.

## Verify

```bash
npm run typecheck
npm run lint
npm run build
```
