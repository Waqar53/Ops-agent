'use client';
import Link from 'next/link';
import { useState } from 'react';
export default function SignupPage() {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [name, setName] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError('');
        try {
            const res = await fetch('http:
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, email, password }),
            });
            if (!res.ok) {
                const data = await res.text();
                throw new Error(data || 'Registration failed');
            }
            const { token, user } = await res.json();
            localStorage.setItem('token', token);
            localStorage.setItem('user', JSON.stringify(user));
            window.location.href = '/dashboard';
        } catch (err: any) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };
    return (
        <div style={styles.page}>
            <Link href="/" style={styles.backLink}>
                ← Back to home
            </Link>
            <div style={styles.container}>
                <div style={styles.header}>
                    <span style={styles.logoIcon}>◈</span>
                    <h1 style={styles.title}>Start your 7-day free trial</h1>
                    <p style={styles.subtitle}>No credit card required. Cancel anytime.</p>
                </div>
                <form onSubmit={handleSubmit} style={styles.form}>
                    <div style={styles.field}>
                        <label style={styles.label}>Full Name</label>
                        <input
                            type="text"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="John Doe"
                            style={styles.input}
                            required
                        />
                    </div>
                    <div style={styles.field}>
                        <label style={styles.label}>Email</label>
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            placeholder="you@company.com"
                            style={styles.input}
                            required
                        />
                    </div>
                    <div style={styles.field}>
                        <label style={styles.label}>Password</label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="••••••••"
                            style={styles.input}
                            required
                            minLength={8}
                        />
                        <span style={styles.hint}>Minimum 8 characters</span>
                    </div>
                    {error && <div style={styles.error}>{error}</div>}
                    <button type="submit" style={styles.button} disabled={loading}>
                        {loading ? 'Creating account...' : 'Start Free Trial →'}
                    </button>
                    <p style={styles.terms}>
                        By signing up, you agree to our{' '}
                        <Link href="/terms" style={styles.link}>Terms of Service</Link> and{' '}
                        <Link href="/privacy" style={styles.link}>Privacy Policy</Link>.
                    </p>
                </form>
                <div style={styles.divider}>
                    <span>or continue with</span>
                </div>
                <div style={styles.oauth}>
                    <button style={styles.oauthBtn}>
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                        </svg>
                        GitHub
                    </button>
                    <button style={styles.oauthBtn}>
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
                            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
                        </svg>
                        Google
                    </button>
                </div>
                <p style={styles.loginLink}>
                    Already have an account? <Link href="/login" style={styles.link}>Sign in</Link>
                </p>
            </div>
        </div>
    );
}
const styles: { [key: string]: React.CSSProperties } = {
    page: {
        minHeight: '100vh',
        background: '#000',
        color: '#fafafa',
        fontFamily: 'Inter, -apple-system, sans-serif',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px',
    },
    backLink: {
        position: 'absolute',
        top: '24px',
        left: '24px',
        color: '#888',
        textDecoration: 'none',
        fontSize: '14px',
    },
    container: {
        width: '100%',
        maxWidth: '400px',
    },
    header: {
        textAlign: 'center',
        marginBottom: '32px',
    },
    logoIcon: {
        fontSize: '40px',
        background: 'linear-gradient(135deg, #0070f3, #7928ca)',
        WebkitBackgroundClip: 'text',
        WebkitTextFillColor: 'transparent',
    },
    title: {
        fontSize: '24px',
        fontWeight: 700,
        margin: '16px 0 8px',
    },
    subtitle: {
        fontSize: '14px',
        color: '#888',
        margin: 0,
    },
    form: {
        background: 'rgba(255,255,255,0.02)',
        border: '1px solid #222',
        borderRadius: '16px',
        padding: '32px',
    },
    field: {
        marginBottom: '20px',
    },
    label: {
        display: 'block',
        fontSize: '14px',
        fontWeight: 500,
        marginBottom: '8px',
    },
    input: {
        width: '100%',
        padding: '12px 16px',
        background: '#0a0a0a',
        border: '1px solid #333',
        borderRadius: '8px',
        color: '#fff',
        fontSize: '16px',
        outline: 'none',
        boxSizing: 'border-box',
    },
    hint: {
        fontSize: '12px',
        color: '#666',
        marginTop: '4px',
        display: 'block',
    },
    error: {
        background: 'rgba(255, 0, 0, 0.1)',
        border: '1px solid rgba(255, 0, 0, 0.3)',
        borderRadius: '8px',
        padding: '12px',
        fontSize: '14px',
        color: '#ff6b6b',
        marginBottom: '20px',
    },
    button: {
        width: '100%',
        padding: '14px',
        background: 'linear-gradient(135deg, #0070f3, #7928ca)',
        color: '#fff',
        border: 'none',
        borderRadius: '8px',
        fontSize: '16px',
        fontWeight: 600,
        cursor: 'pointer',
    },
    terms: {
        fontSize: '12px',
        color: '#666',
        textAlign: 'center',
        marginTop: '16px',
    },
    link: {
        color: '#0070f3',
        textDecoration: 'none',
    },
    divider: {
        display: 'flex',
        alignItems: 'center',
        margin: '24px 0',
        color: '#666',
        fontSize: '14px',
    },
    oauth: {
        display: 'flex',
        gap: '12px',
    },
    oauthBtn: {
        flex: 1,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '8px',
        padding: '12px',
        background: 'transparent',
        border: '1px solid #333',
        borderRadius: '8px',
        color: '#fff',
        fontSize: '14px',
        cursor: 'pointer',
    },
    loginLink: {
        textAlign: 'center',
        fontSize: '14px',
        color: '#888',
        marginTop: '24px',
    },
};
