# Example Next.js Application for OpsAgent Testing

This is a production-ready Next.js application with PostgreSQL, Redis, and Stripe integration to test the OpsAgent platform.

## Features

- Next.js 14 with App Router
- PostgreSQL database
- Redis caching
- Stripe payments
- TypeScript
- Tailwind CSS

## Local Development

```bash
npm install
npm run dev
```

## Deploy with OpsAgent

```bash
# Initialize OpsAgent
ops init

# Deploy to production
ops deploy
```

## Environment Variables

```env
DATABASE_URL=postgresql://user:pass@localhost:5432/myapp
REDIS_URL=redis://localhost:6379
STRIPE_SECRET_KEY=sk_test_...
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_...
```
