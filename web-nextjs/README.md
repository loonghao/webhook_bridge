# Webhook Bridge Dashboard

Next.js dashboard for Webhook Bridge 4.0. It connects to the Rust
`/api/dashboard/*` endpoints and shows route health, worker status, execution
logs, and local hook activity.

## Development

```bash
npm install
npm run dev
npm run type-check
npm run lint
```

The development server proxies API calls to `http://localhost:8080` unless
`NEXT_PUBLIC_API_BASE_URL` is set.
