# Our Sell frontend

The frontend for the Our Sell authentication and session-management API. It uses Next.js App Router, React 19, TypeScript, Tailwind CSS v4, and Lucide.

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
- `components/dashboard/` contains the responsive account workspace.
- `components/providers/` owns auth restoration and theme state.
- `components/ui/` contains small reusable controls.
- `lib/api.ts` is the typed API boundary. Access tokens stay in memory; refresh tokens remain in the backend’s HttpOnly cookie.

## Verify

```bash
npm run typecheck
npm run lint
npm run build
```
