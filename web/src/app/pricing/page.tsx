'use client';

import Link from 'next/link';
import { useState } from 'react';

export default function PricingPage() {
    const [annual, setAnnual] = useState(true);

    const plans = [
        {
            name: 'Starter',
            desc: 'Perfect for side projects and learning',
            price: annual ? 0 : 0,
            features: [
                '1 project',
                '100 deployments/month',
                'Community support',
                'Basic monitoring',
                'Shared resources',
            ],
            cta: 'Start Free',
            highlight: false,
        },
        {
            name: 'Pro',
            desc: 'For professional developers and teams',
            price: annual ? 29 : 39,
            features: [
                '10 projects',
                'Unlimited deployments',
                'Priority support',
                'Advanced monitoring',
                'Custom domains',
                'Auto-scaling',
                'Team collaboration',
            ],
            cta: 'Start 7-Day Trial',
            highlight: true,
            trial: true,
        },
        {
            name: 'Enterprise',
            desc: 'For large organizations',
            price: annual ? 199 : 249,
            features: [
                'Unlimited projects',
                'Unlimited deployments',
                '24/7 dedicated support',
                'SLA guarantee',
                'SSO / SAML',
                'Audit logs',
                'On-premise option',
                'Custom integrations',
            ],
            cta: 'Contact Sales',
            highlight: false,
        },
    ];

    return (
        <div style={styles.page}>
            <nav style={styles.nav}>
                <div style={styles.navInner}>
                    <Link href="/" style={styles.logo}>
                        <span style={styles.logoIcon}>◈</span>
                        OpsAgent
                    </Link>
                    <div style={styles.navLinks}>
                        <Link href="/#features" style={styles.navLink}>Features</Link>
                        <Link href="/pricing" style={styles.navLinkActive}>Pricing</Link>
                        <Link href="/docs" style={styles.navLink}>Docs</Link>
                        <Link href="/login" style={styles.navLink}>Sign In</Link>
                    </div>
                </div>
            </nav>

            <section style={styles.hero}>
                <h1 style={styles.h1}>Simple, transparent pricing</h1>
                <p style={styles.subtitle}>
                    Start free, scale as you grow. No hidden fees, cancel anytime.
                </p>

                <div style={styles.toggle}>
                    <span style={!annual ? styles.toggleActive : styles.toggleInactive}>Monthly</span>
                    <button
                        onClick={() => setAnnual(!annual)}
                        style={styles.toggleSwitch}
                    >
                        <span style={{
                            ...styles.toggleKnob,
                            transform: annual ? 'translateX(24px)' : 'translateX(0)',
                        }}></span>
                    </button>
                    <span style={annual ? styles.toggleActive : styles.toggleInactive}>
                        Annual <span style={styles.saveBadge}>Save 25%</span>
                    </span>
                </div>
            </section>

            <section style={styles.plans}>
                {plans.map((plan, i) => (
                    <div
                        key={i}
                        style={{
                            ...styles.planCard,
                            ...(plan.highlight ? styles.planHighlight : {}),
                        }}
                    >
                        {plan.trial && (
                            <div style={styles.trialBadge}>7-DAY FREE TRIAL</div>
                        )}
                        <h3 style={styles.planName}>{plan.name}</h3>
                        <p style={styles.planDesc}>{plan.desc}</p>
                        <div style={styles.planPrice}>
                            <span style={styles.currency}>$</span>
                            <span style={styles.amount}>{plan.price}</span>
                            <span style={styles.period}>/month</span>
                        </div>
                        <ul style={styles.features}>
                            {plan.features.map((feature, j) => (
                                <li key={j} style={styles.feature}>
                                    <span style={styles.check}>✓</span> {feature}
                                </li>
                            ))}
                        </ul>
                        <Link
                            href="/signup"
                            style={plan.highlight ? styles.btnPrimary : styles.btnSecondary}
                        >
                            {plan.cta}
                        </Link>
                    </div>
                ))}
            </section>

            <section style={styles.faq}>
                <h2 style={styles.faqTitle}>Frequently Asked Questions</h2>
                <div style={styles.faqGrid}>
                    <div style={styles.faqItem}>
                        <h4 style={styles.faqQuestion}>How does the 7-day trial work?</h4>
                        <p style={styles.faqAnswer}>
                            Start using all Pro features immediately. No credit card required for trial.
                            Cancel or downgrade anytime before trial ends.
                        </p>
                    </div>
                    <div style={styles.faqItem}>
                        <h4 style={styles.faqQuestion}>Can I change plans anytime?</h4>
                        <p style={styles.faqAnswer}>
                            Yes! Upgrade or downgrade at any time. Changes apply immediately,
                            and we'll prorate your billing.
                        </p>
                    </div>
                    <div style={styles.faqItem}>
                        <h4 style={styles.faqQuestion}>What clouds do you support?</h4>
                        <p style={styles.faqAnswer}>
                            AWS is fully supported in v1.0. GCP and Azure support coming soon.
                            We use standard containerization for portability.
                        </p>
                    </div>
                    <div style={styles.faqItem}>
                        <h4 style={styles.faqQuestion}>Need enterprise features?</h4>
                        <p style={styles.faqAnswer}>
                            Contact our sales team for custom pricing, SLAs, on-premise deployment,
                            and dedicated support.
                        </p>
                    </div>
                </div>
            </section>

            <footer style={styles.footer}>
                <p style={styles.footerText}>© 2024 OpsAgent. All rights reserved.</p>
            </footer>
        </div>
    );
}

const styles: { [key: string]: React.CSSProperties } = {
    page: {
        minHeight: '100vh',
        background: '#000',
        color: '#fafafa',
        fontFamily: 'Inter, -apple-system, sans-serif',
    },
    nav: {
        position: 'fixed',
        top: 0,
        width: '100%',
        zIndex: 100,
        background: 'rgba(0, 0, 0, 0.85)',
        backdropFilter: 'blur(12px)',
        borderBottom: '1px solid rgba(255,255,255,0.1)',
    },
    navInner: {
        maxWidth: '1200px',
        margin: '0 auto',
        padding: '16px 24px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
    },
    logo: {
        display: 'flex',
        alignItems: 'center',
        gap: '10px',
        fontWeight: 700,
        fontSize: '20px',
        textDecoration: 'none',
        color: '#fff',
    },
    logoIcon: {
        fontSize: '24px',
        background: 'linear-gradient(135deg, #0070f3, #7928ca)',
        WebkitBackgroundClip: 'text',
        WebkitTextFillColor: 'transparent',
    },
    navLinks: {
        display: 'flex',
        gap: '32px',
        alignItems: 'center',
    },
    navLink: {
        color: '#888',
        textDecoration: 'none',
        fontSize: '14px',
    },
    navLinkActive: {
        color: '#fff',
        textDecoration: 'none',
        fontSize: '14px',
        fontWeight: 600,
    },
    hero: {
        paddingTop: '140px',
        textAlign: 'center',
        maxWidth: '600px',
        margin: '0 auto',
        padding: '140px 24px 48px',
    },
    h1: {
        fontSize: '48px',
        fontWeight: 800,
        margin: '0 0 16px',
    },
    subtitle: {
        fontSize: '18px',
        color: '#888',
        marginBottom: '32px',
    },
    toggle: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '12px',
    },
    toggleActive: {
        color: '#fff',
        fontWeight: 600,
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
    },
    toggleInactive: {
        color: '#666',
    },
    toggleSwitch: {
        width: '56px',
        height: '32px',
        background: '#333',
        borderRadius: '16px',
        border: 'none',
        cursor: 'pointer',
        position: 'relative',
        padding: '4px',
    },
    toggleKnob: {
        width: '24px',
        height: '24px',
        background: '#0070f3',
        borderRadius: '50%',
        display: 'block',
        transition: 'transform 0.2s',
    },
    saveBadge: {
        background: 'linear-gradient(135deg, #0070f3, #7928ca)',
        color: '#fff',
        padding: '4px 8px',
        borderRadius: '4px',
        fontSize: '12px',
    },
    plans: {
        display: 'flex',
        justifyContent: 'center',
        gap: '24px',
        padding: '48px 24px',
        flexWrap: 'wrap',
        maxWidth: '1200px',
        margin: '0 auto',
    },
    planCard: {
        background: 'rgba(255,255,255,0.02)',
        border: '1px solid #222',
        borderRadius: '16px',
        padding: '32px',
        width: '320px',
        position: 'relative',
    },
    planHighlight: {
        border: '2px solid #0070f3',
        background: 'rgba(0, 112, 243, 0.05)',
    },
    trialBadge: {
        position: 'absolute',
        top: '-12px',
        left: '50%',
        transform: 'translateX(-50%)',
        background: 'linear-gradient(135deg, #0070f3, #7928ca)',
        color: '#fff',
        padding: '6px 16px',
        borderRadius: '20px',
        fontSize: '12px',
        fontWeight: 600,
    },
    planName: {
        fontSize: '24px',
        fontWeight: 700,
        marginBottom: '8px',
    },
    planDesc: {
        fontSize: '14px',
        color: '#888',
        marginBottom: '24px',
    },
    planPrice: {
        marginBottom: '24px',
    },
    currency: {
        fontSize: '24px',
        verticalAlign: 'top',
    },
    amount: {
        fontSize: '48px',
        fontWeight: 800,
    },
    period: {
        fontSize: '16px',
        color: '#888',
    },
    features: {
        listStyle: 'none',
        padding: 0,
        marginBottom: '24px',
    },
    feature: {
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        marginBottom: '12px',
        fontSize: '14px',
        color: '#ccc',
    },
    check: {
        color: '#0070f3',
        fontWeight: 'bold',
    },
    btnPrimary: {
        display: 'block',
        textAlign: 'center',
        padding: '14px 24px',
        background: 'linear-gradient(135deg, #0070f3, #7928ca)',
        color: '#fff',
        borderRadius: '8px',
        fontWeight: 600,
        textDecoration: 'none',
    },
    btnSecondary: {
        display: 'block',
        textAlign: 'center',
        padding: '14px 24px',
        background: 'transparent',
        color: '#fff',
        borderRadius: '8px',
        fontWeight: 600,
        textDecoration: 'none',
        border: '1px solid #333',
    },
    faq: {
        maxWidth: '900px',
        margin: '0 auto',
        padding: '80px 24px',
    },
    faqTitle: {
        fontSize: '32px',
        fontWeight: 700,
        textAlign: 'center',
        marginBottom: '48px',
    },
    faqGrid: {
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(350px, 1fr))',
        gap: '32px',
    },
    faqItem: {
        background: 'rgba(255,255,255,0.02)',
        border: '1px solid #222',
        borderRadius: '12px',
        padding: '24px',
    },
    faqQuestion: {
        fontSize: '16px',
        fontWeight: 600,
        marginBottom: '8px',
    },
    faqAnswer: {
        fontSize: '14px',
        color: '#888',
        lineHeight: 1.6,
        margin: 0,
    },
    footer: {
        padding: '32px 24px',
        borderTop: '1px solid #222',
        textAlign: 'center',
    },
    footerText: {
        fontSize: '14px',
        color: '#666',
    },
};
