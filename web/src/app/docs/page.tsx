'use client';

import Link from 'next/link';
import { useState } from 'react';

export default function DocsPage() {
    const [activeSection, setActiveSection] = useState('quickstart');

    const sections = [
        { id: 'quickstart', title: 'Quick Start' },
        { id: 'installation', title: 'Installation' },
        { id: 'configuration', title: 'Configuration' },
        { id: 'deploy', title: 'Deploying' },
        { id: 'monitoring', title: 'Monitoring' },
        { id: 'security', title: 'Security' },
        { id: 'api', title: 'API Reference' },
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
                        <Link href="/pricing" style={styles.navLink}>Pricing</Link>
                        <Link href="/docs" style={styles.navLinkActive}>Docs</Link>
                        <Link href="/login" style={styles.navLink}>Sign In</Link>
                    </div>
                </div>
            </nav>

            <div style={styles.container}>
                <aside style={styles.sidebar}>
                    <h3 style={styles.sidebarTitle}>Documentation</h3>
                    {sections.map((section) => (
                        <a
                            key={section.id}
                            href={`#${section.id}`}
                            onClick={() => setActiveSection(section.id)}
                            style={activeSection === section.id ? styles.sidebarLinkActive : styles.sidebarLink}
                        >
                            {section.title}
                        </a>
                    ))}
                </aside>

                <main style={styles.content}>
                    <section id="quickstart" style={styles.section}>
                        <h1 style={styles.h1}>Quick Start Guide</h1>
                        <p style={styles.intro}>
                            Get your first application deployed in under 5 minutes with OpsAgent.
                        </p>

                        <div style={styles.codeBlock}>
                            <div style={styles.codeHeader}>Terminal</div>
                            <pre style={styles.pre}>
                                {`# Install OpsAgent CLI
npm install -g @opsagent/cli

# Navigate to your project
cd my-project

# Initialize OpsAgent (AI analyzes your code)
ops init

# Deploy to production
ops deploy`}
                            </pre>
                        </div>

                        <p style={styles.text}>
                            That's it! OpsAgent automatically detects your language, framework,
                            dependencies, and configures everything for optimal deployment.
                        </p>
                    </section>

                    <section id="installation" style={styles.section}>
                        <h2 style={styles.h2}>Installation</h2>

                        <h3 style={styles.h3}>Prerequisites</h3>
                        <ul style={styles.list}>
                            <li>Node.js 18+ or npm 8+</li>
                            <li>A supported cloud account (AWS, GCP, or Azure)</li>
                            <li>OpsAgent account (free tier available)</li>
                        </ul>

                        <h3 style={styles.h3}>Install via npm</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>npm install -g @opsagent/cli</pre>
                        </div>

                        <h3 style={styles.h3}>Verify Installation</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>ops --version</pre>
                        </div>

                        <h3 style={styles.h3}>Login to OpsAgent</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>ops login</pre>
                        </div>
                    </section>

                    <section id="configuration" style={styles.section}>
                        <h2 style={styles.h2}>Configuration</h2>

                        <p style={styles.text}>
                            When you run <code style={styles.code}>ops init</code>, OpsAgent's AI
                            analyzes your codebase and generates an <code style={styles.code}>opsagent.yml</code>
                            configuration file.
                        </p>

                        <h3 style={styles.h3}>AI-Detected Configuration</h3>
                        <div style={styles.codeBlock}>
                            <div style={styles.codeHeader}>opsagent.yml</div>
                            <pre style={styles.pre}>
                                {`version: "1"
app:
  name: my-saas-app
  language: Node.js
  framework: Next.js

build:
  base_image: node:20-alpine
  port: 3000
  health_check: /api/health
  multi_stage: true

environments:
  production:
    replicas: 2
    resources:
      cpu: 500m
      memory: 1Gi
    scaling:
      enabled: true
      min: 1
      max: 10
      target_cpu: 70

services:
  postgresql:
    version: "15"
    backup:
      enabled: true
      frequency: daily
  redis:
    version: "7"

monitoring:
  metrics: true
  logging: true
  alerts:
    - cpu_usage > 80%
    - memory_usage > 85%
    - error_rate > 1%
    - latency_p99 > 500ms

security:
  ssl: auto
  secrets: encrypted
  scanning: true`}
                            </pre>
                        </div>

                        <h3 style={styles.h3}>What Gets Detected</h3>
                        <table style={styles.table}>
                            <thead>
                                <tr>
                                    <th style={styles.th}>Feature</th>
                                    <th style={styles.th}>Detection Method</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td style={styles.td}>Language</td>
                                    <td style={styles.td}>package.json, requirements.txt, go.mod</td>
                                </tr>
                                <tr>
                                    <td style={styles.td}>Framework</td>
                                    <td style={styles.td}>Dependency analysis (Next.js, Express, Django, etc.)</td>
                                </tr>
                                <tr>
                                    <td style={styles.td}>Services</td>
                                    <td style={styles.td}>Database drivers, cache clients, API integrations</td>
                                </tr>
                                <tr>
                                    <td style={styles.td}>Resources</td>
                                    <td style={styles.td}>Dependency count, framework requirements</td>
                                </tr>
                            </tbody>
                        </table>
                    </section>

                    <section id="deploy" style={styles.section}>
                        <h2 style={styles.h2}>Deploying Your Application</h2>

                        <h3 style={styles.h3}>Basic Deploy</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>ops deploy</pre>
                        </div>

                        <h3 style={styles.h3}>Deploy to Specific Environment</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>{`ops deploy --env staging
ops deploy --env production`}</pre>
                        </div>

                        <h3 style={styles.h3}>Dry Run (Preview Changes)</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>ops deploy --dry-run</pre>
                        </div>

                        <h3 style={styles.h3}>Deployment Strategies</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>{`# Rolling deployment (default)
ops deploy --strategy rolling

# Blue-green deployment
ops deploy --strategy blue-green

# Canary deployment (10% traffic first)
ops deploy --strategy canary --canary-percent 10`}</pre>
                        </div>

                        <div style={styles.alert}>
                            <strong>💡 Pro Tip:</strong> Use <code style={styles.code}>--dry-run</code> to
                            preview what OpsAgent will deploy before making changes.
                        </div>
                    </section>

                    <section id="monitoring" style={styles.section}>
                        <h2 style={styles.h2}>Monitoring & Observability</h2>

                        <p style={styles.text}>
                            OpsAgent automatically configures monitoring for every deployment:
                        </p>

                        <h3 style={styles.h3}>Default Alert Rules</h3>
                        <ul style={styles.list}>
                            <li><strong>CPU Usage &gt; 80%</strong> - Triggers scale-up or alert</li>
                            <li><strong>Memory Usage &gt; 85%</strong> - Potential memory leak warning</li>
                            <li><strong>Error Rate &gt; 1%</strong> - Application health issue</li>
                            <li><strong>P99 Latency &gt; 500ms</strong> - Performance degradation</li>
                        </ul>

                        <h3 style={styles.h3}>View Logs</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>{`ops logs
ops logs --follow
ops logs --since 1h`}</pre>
                        </div>

                        <h3 style={styles.h3}>View Metrics</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>ops metrics</pre>
                        </div>
                    </section>

                    <section id="security" style={styles.section}>
                        <h2 style={styles.h2}>Security</h2>

                        <h3 style={styles.h3}>Automatic Security Features</h3>
                        <ul style={styles.list}>
                            <li><strong>SSL/TLS:</strong> Automatic certificate provisioning</li>
                            <li><strong>Secrets:</strong> Encrypted storage, never in plain text</li>
                            <li><strong>Scanning:</strong> Vulnerability detection in dependencies</li>
                            <li><strong>Compliance:</strong> SOC2, HIPAA-ready configurations</li>
                        </ul>

                        <h3 style={styles.h3}>Managing Secrets</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>{`# Set a secret
ops secrets set DATABASE_URL "postgresql://..."

# List secrets (values hidden)
ops secrets list

# Remove a secret
ops secrets delete API_KEY`}</pre>
                        </div>
                    </section>

                    <section id="api" style={styles.section}>
                        <h2 style={styles.h2}>API Reference</h2>

                        <p style={styles.text}>
                            OpsAgent provides a REST API for programmatic access to all features.
                        </p>

                        <h3 style={styles.h3}>Authentication</h3>
                        <div style={styles.codeBlock}>
                            <pre style={styles.pre}>{`curl -H "Authorization: Bearer YOUR_API_KEY" \\
  https://api.opsagent.dev/v1/projects`}</pre>
                        </div>

                        <h3 style={styles.h3}>Endpoints</h3>
                        <table style={styles.table}>
                            <thead>
                                <tr>
                                    <th style={styles.th}>Method</th>
                                    <th style={styles.th}>Endpoint</th>
                                    <th style={styles.th}>Description</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr><td style={styles.td}>GET</td><td style={styles.td}>/v1/projects</td><td style={styles.td}>List all projects</td></tr>
                                <tr><td style={styles.td}>POST</td><td style={styles.td}>/v1/projects</td><td style={styles.td}>Create new project</td></tr>
                                <tr><td style={styles.td}>POST</td><td style={styles.td}>/v1/projects/:id/deploy</td><td style={styles.td}>Trigger deployment</td></tr>
                                <tr><td style={styles.td}>GET</td><td style={styles.td}>/v1/deployments</td><td style={styles.td}>List deployments</td></tr>
                                <tr><td style={styles.td}>GET</td><td style={styles.td}>/v1/metrics</td><td style={styles.td}>Get metrics</td></tr>
                                <tr><td style={styles.td}>GET</td><td style={styles.td}>/v1/logs</td><td style={styles.td}>Stream logs</td></tr>
                            </tbody>
                        </table>
                    </section>

                    <div style={styles.cta}>
                        <h3 style={styles.ctaTitle}>Ready to Deploy?</h3>
                        <p style={styles.ctaText}>Start your 7-day free trial and deploy your first app in minutes.</p>
                        <Link href="/signup" style={styles.btnPrimary}>Start Free Trial →</Link>
                    </div>
                </main>
            </div>

            <footer style={styles.footer}>
                <p style={styles.footerText}>© 2024 OpsAgent by Mohd Waqar Azim. All rights reserved.</p>
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
        maxWidth: '1400px',
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
    container: {
        display: 'flex',
        paddingTop: '64px',
        maxWidth: '1400px',
        margin: '0 auto',
    },
    sidebar: {
        width: '260px',
        padding: '32px 24px',
        borderRight: '1px solid #222',
        position: 'sticky',
        top: '64px',
        height: 'calc(100vh - 64px)',
        overflowY: 'auto',
    },
    sidebarTitle: {
        fontSize: '12px',
        fontWeight: 600,
        color: '#888',
        textTransform: 'uppercase',
        letterSpacing: '1px',
        marginBottom: '16px',
    },
    sidebarLink: {
        display: 'block',
        color: '#888',
        textDecoration: 'none',
        padding: '8px 12px',
        borderRadius: '6px',
        marginBottom: '4px',
        fontSize: '14px',
    },
    sidebarLinkActive: {
        display: 'block',
        color: '#fff',
        textDecoration: 'none',
        padding: '8px 12px',
        borderRadius: '6px',
        marginBottom: '4px',
        fontSize: '14px',
        background: 'rgba(0, 112, 243, 0.1)',
        borderLeft: '2px solid #0070f3',
    },
    content: {
        flex: 1,
        padding: '32px 64px',
        maxWidth: '900px',
    },
    section: {
        marginBottom: '64px',
        paddingTop: '32px',
    },
    h1: {
        fontSize: '36px',
        fontWeight: 800,
        marginBottom: '16px',
    },
    h2: {
        fontSize: '28px',
        fontWeight: 700,
        marginBottom: '24px',
        paddingTop: '16px',
        borderTop: '1px solid #222',
    },
    h3: {
        fontSize: '18px',
        fontWeight: 600,
        marginTop: '32px',
        marginBottom: '16px',
    },
    intro: {
        fontSize: '18px',
        color: '#888',
        lineHeight: 1.6,
        marginBottom: '32px',
    },
    text: {
        fontSize: '16px',
        color: '#ccc',
        lineHeight: 1.7,
        marginBottom: '16px',
    },
    list: {
        paddingLeft: '24px',
        marginBottom: '24px',
    },
    codeBlock: {
        background: '#0a0a0a',
        border: '1px solid #333',
        borderRadius: '8px',
        overflow: 'hidden',
        marginBottom: '24px',
    },
    codeHeader: {
        background: '#1a1a1a',
        padding: '8px 16px',
        fontSize: '12px',
        color: '#888',
        borderBottom: '1px solid #333',
    },
    pre: {
        padding: '16px',
        margin: 0,
        fontSize: '14px',
        lineHeight: 1.6,
        color: '#50fa7b',
        overflow: 'auto',
    },
    code: {
        background: '#1a1a1a',
        padding: '2px 6px',
        borderRadius: '4px',
        fontSize: '14px',
        color: '#f472b6',
    },
    table: {
        width: '100%',
        borderCollapse: 'collapse',
        marginBottom: '24px',
    },
    th: {
        textAlign: 'left',
        padding: '12px',
        background: '#111',
        borderBottom: '1px solid #333',
        fontSize: '14px',
        fontWeight: 600,
    },
    td: {
        padding: '12px',
        borderBottom: '1px solid #222',
        fontSize: '14px',
        color: '#ccc',
    },
    alert: {
        background: 'rgba(0, 112, 243, 0.1)',
        border: '1px solid rgba(0, 112, 243, 0.3)',
        borderRadius: '8px',
        padding: '16px',
        fontSize: '14px',
        marginTop: '24px',
    },
    cta: {
        background: 'linear-gradient(135deg, rgba(0, 112, 243, 0.1), rgba(121, 40, 202, 0.1))',
        border: '1px solid rgba(255, 255, 255, 0.1)',
        borderRadius: '16px',
        padding: '32px',
        textAlign: 'center',
        marginTop: '64px',
    },
    ctaTitle: {
        fontSize: '24px',
        fontWeight: 700,
        marginBottom: '8px',
    },
    ctaText: {
        color: '#888',
        marginBottom: '24px',
    },
    btnPrimary: {
        display: 'inline-block',
        padding: '14px 28px',
        background: 'linear-gradient(135deg, #0070f3, #7928ca)',
        color: '#fff',
        borderRadius: '8px',
        fontWeight: 600,
        textDecoration: 'none',
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
