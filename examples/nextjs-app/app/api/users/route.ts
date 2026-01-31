import { NextResponse } from 'next/server';
import { db, redis } from '@/lib/db';

export async function GET() {
    try {
        // Try to get from cache first
        const cached = await redis.get('users');
        if (cached) {
            return NextResponse.json(JSON.parse(cached));
        }

        // Query database
        const result = await db.query('SELECT id, name, email FROM users LIMIT 10');
        const users = result.rows;

        // Cache for 5 minutes
        await redis.setEx('users', 300, JSON.stringify(users));

        return NextResponse.json(users);
    } catch (error) {
        return NextResponse.json({ error: String(error) }, { status: 500 });
    }
}

export async function POST(request: Request) {
    try {
        const body = await request.json();
        const { name, email } = body;

        const result = await db.query(
            'INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *',
            [name, email]
        );

        // Invalidate cache
        await redis.del('users');

        return NextResponse.json(result.rows[0], { status: 201 });
    } catch (error) {
        return NextResponse.json({ error: String(error) }, { status: 500 });
    }
}
