import { Pool } from 'pg';
import { createClient } from 'redis';
import Stripe from 'stripe';

// Database connection
export const db = new Pool({
    connectionString: process.env.DATABASE_URL,
    max: 20,
    idleTimeoutMillis: 30000,
    connectionTimeoutMillis: 2000,
});

// Redis connection
export const redis = createClient({
    url: process.env.REDIS_URL,
});

redis.on('error', (err) => console.log('Redis Client Error', err));
redis.connect();

// Stripe
export const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!, {
    apiVersion: '2023-10-16',
});

// Health check
export async function healthCheck() {
    try {
        // Check database
        await db.query('SELECT 1');

        // Check Redis
        await redis.ping();

        return { status: 'healthy', timestamp: new Date().toISOString() };
    } catch (error) {
        return { status: 'unhealthy', error: String(error) };
    }
}
