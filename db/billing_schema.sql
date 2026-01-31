-- Billing and subscription schema

-- Subscription plans table
CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL, -- free, starter, pro, team, enterprise
    display_name VARCHAR(100) NOT NULL,
    price_monthly DECIMAL(10, 2) NOT NULL,
    price_yearly DECIMAL(10, 2),
    stripe_price_id_monthly VARCHAR(255),
    stripe_price_id_yearly VARCHAR(255),
    features JSONB DEFAULT '{}',
    limits JSONB DEFAULT '{}', -- {projects: 5, environments: 10, deployments_per_month: 100}
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Usage tracking
CREATE TABLE IF NOT EXISTS usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL, -- compute, storage, bandwidth, builds
    amount DECIMAL(15, 6) NOT NULL,
    unit VARCHAR(20) NOT NULL, -- hours, gb, gb-transfer, count
    cost DECIMAL(10, 4) DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    recorded_at TIMESTAMP DEFAULT NOW()
);

-- Invoices
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    stripe_invoice_id VARCHAR(255),
    amount_due DECIMAL(10, 2) NOT NULL,
    amount_paid DECIMAL(10, 2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'draft', -- draft, open, paid, void, uncollectible
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    due_date TIMESTAMP,
    paid_at TIMESTAMP,
    invoice_pdf_url TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Payment methods
CREATE TABLE IF NOT EXISTS payment_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    stripe_payment_method_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- card, bank_account
    is_default BOOLEAN DEFAULT FALSE,
    card_brand VARCHAR(50),
    card_last4 VARCHAR(4),
    card_exp_month INTEGER,
    card_exp_year INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Billing events (for audit trail)
CREATE TABLE IF NOT EXISTS billing_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL, -- subscription.created, payment.succeeded, etc.
    stripe_event_id VARCHAR(255),
    data JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_usage_org ON usage_records(organization_id);
CREATE INDEX IF NOT EXISTS idx_usage_recorded_at ON usage_records(recorded_at);
CREATE INDEX IF NOT EXISTS idx_invoices_org ON invoices(organization_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_payment_methods_org ON payment_methods(organization_id);
CREATE INDEX IF NOT EXISTS idx_billing_events_org ON billing_events(organization_id);

-- Seed subscription plans
INSERT INTO subscription_plans (name, display_name, price_monthly, price_yearly, features, limits) VALUES
('free', 'Free', 0, 0, 
 '{"projects": 1, "environments": 1, "support": "community"}',
 '{"projects": 1, "environments": 1, "deployments_per_month": 10, "storage_gb": 1, "bandwidth_gb": 10}'),
('starter', 'Starter', 29, 290,
 '{"projects": 5, "environments": "unlimited", "support": "email", "preview_envs": true}',
 '{"projects": 5, "deployments_per_month": 100, "storage_gb": 10, "bandwidth_gb": 100}'),
('pro', 'Pro', 99, 990,
 '{"projects": 20, "environments": "unlimited", "support": "priority", "preview_envs": true, "advanced_metrics": true}',
 '{"projects": 20, "deployments_per_month": 500, "storage_gb": 50, "bandwidth_gb": 500}'),
('team', 'Team', 299, 2990,
 '{"projects": "unlimited", "environments": "unlimited", "support": "priority", "preview_envs": true, "advanced_metrics": true, "sso": true, "rbac": true}',
 '{"projects": -1, "deployments_per_month": 2000, "storage_gb": 200, "bandwidth_gb": 2000}'),
('enterprise', 'Enterprise', 999, 9990,
 '{"projects": "unlimited", "environments": "unlimited", "support": "dedicated", "preview_envs": true, "advanced_metrics": true, "sso": true, "rbac": true, "multi_cloud": true, "compliance": true, "sla": "99.99%"}',
 '{"projects": -1, "deployments_per_month": -1, "storage_gb": 1000, "bandwidth_gb": 10000}')
ON CONFLICT (name) DO NOTHING;
