-- OpsAgent Database Schema
-- PostgreSQL 15+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    language VARCHAR(50) NOT NULL,
    framework VARCHAR(100) NOT NULL,
    git_repo VARCHAR(500),
    git_branch VARCHAR(100) DEFAULT 'main',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by UUID,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Environments table
CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- production, staging, development, preview
    status VARCHAR(50) DEFAULT 'active',
    url VARCHAR(500),
    config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- Deployments table
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id) ON DELETE CASCADE,
    version VARCHAR(100) NOT NULL,
    git_commit VARCHAR(40),
    git_branch VARCHAR(100),
    strategy VARCHAR(50) NOT NULL, -- rolling, blue-green, canary, direct
    status VARCHAR(50) DEFAULT 'pending', -- pending, building, deploying, success, failed, rolled_back
    deployed_by VARCHAR(255),
    deployed_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    duration_seconds INTEGER,
    build_logs TEXT,
    deploy_logs TEXT,
    rollback_from UUID REFERENCES deployments(id),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Services table (detected services like PostgreSQL, Redis, etc.)
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- database, cache, queue, storage
    version VARCHAR(50),
    config JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(50) DEFAULT 'provisioning',
    endpoint VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Resources table (infrastructure resources)
CREATE TABLE resources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id) ON DELETE CASCADE,
    resource_type VARCHAR(100) NOT NULL, -- ec2, rds, elasticache, alb, etc.
    resource_id VARCHAR(255) NOT NULL,
    region VARCHAR(50),
    status VARCHAR(50) DEFAULT 'creating',
    config JSONB DEFAULT '{}'::jsonb,
    cost_per_hour DECIMAL(10, 4),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Metrics table (aggregated metrics)
CREATE TABLE metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15, 4) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    tags JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Alerts table
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id),
    alert_type VARCHAR(50) NOT NULL, -- error_rate, latency, cpu, memory, cost
    severity VARCHAR(20) NOT NULL, -- info, warning, critical
    title VARCHAR(255) NOT NULL,
    message TEXT,
    status VARCHAR(50) DEFAULT 'active', -- active, acknowledged, resolved
    triggered_at TIMESTAMP DEFAULT NOW(),
    resolved_at TIMESTAMP,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Cost tracking table
CREATE TABLE costs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id),
    resource_id UUID REFERENCES resources(id),
    date DATE NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    breakdown JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, environment_id, resource_id, date)
);

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    password_hash VARCHAR(255),
    role VARCHAR(50) DEFAULT 'developer',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- API Keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_projects_slug ON projects(slug);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_deployments_project ON deployments(project_id);
CREATE INDEX idx_deployments_status ON deployments(status);
CREATE INDEX idx_deployments_created ON deployments(created_at DESC);
CREATE INDEX idx_metrics_project_time ON metrics(project_id, timestamp DESC);
CREATE INDEX idx_metrics_name ON metrics(metric_name);
CREATE INDEX idx_alerts_project ON alerts(project_id);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_costs_project_date ON costs(project_id, date DESC);

-- Functions
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers
CREATE TRIGGER projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER environments_updated_at BEFORE UPDATE ON environments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER services_updated_at BEFORE UPDATE ON services
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER resources_updated_at BEFORE UPDATE ON resources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
