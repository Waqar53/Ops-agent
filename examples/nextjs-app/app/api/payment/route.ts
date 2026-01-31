import { NextResponse } from 'next/server';
import { stripe } from '@/lib/db';
export async function POST(request: Request) {
    try {
        const body = await request.json();
        const { amount, currency = 'usd' } = body;
        const paymentIntent = await stripe.paymentIntents.create({
            amount,
            currency,
            automatic_payment_methods: {
                enabled: true,
            },
        });
        return NextResponse.json({
            clientSecret: paymentIntent.client_secret,
        });
    } catch (error) {
        return NextResponse.json({ error: String(error) }, { status: 500 });
    }
}
