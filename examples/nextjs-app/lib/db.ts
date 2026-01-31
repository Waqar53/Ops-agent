import { Pool } from 'pg';
import { createClient } from 'redis';
import Stripe from 'stripe';
export const db = new Pool({
    connectionString: process.env.DATABASE_URL,
    max: 20,
    idleTimeoutMillis: 30000,
    connectionTimeoutMillis: 2000,
});
export const redis = createClient({
    url: process.env.REDIS_URL,
});
redis.on('error', (err) => console.log('Redis Client Error', err));
redis.connect();
export const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!, {
    apiVersion: '2023-10-16',
});
export async function healthCheck() {
    try {
        await db.query('SELECT 1');
        await redis.ping();
        return { status: 'healthy', timestamp: new Date().toISOString() };
    } catch (error) {
        return { status: 'unhealthy', error: String(error) };
    }
}
