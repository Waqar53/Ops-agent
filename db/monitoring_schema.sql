-- Monitoring and alerting schema

-- Metrics table (time-series data)
CREATE TABLE IF NOT EXISTS metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL, -- cpu, memory, disk, network, requests, latency, errors, custom
    name VARCHAR(255) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(50) NOT NULL, -- percent, bytes, ms, count, etc.
    tags JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Alert configurations
CREATE TABLE IF NOT EXISTS alert_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    condition VARCHAR(10) NOT NULL, -- >, <, >=, <=, ==, !=
    threshold DOUBLE PRECISION NOT NULL,
    duration INTEGER DEFAULT 60, -- seconds
    severity VARCHAR(20) DEFAULT 'warning', -- info, warning, critical
    enabled BOOLEAN DEFAULT TRUE,
    channels JSONB DEFAULT '[]', -- ["email", "slack", "pagerduty"]
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Notification channels
CREATE TABLE IF NOT EXISTS notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- email, slack, pagerduty, webhook
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL, -- {webhook_url: "...", api_key: "..."}
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_metrics_project ON metrics(project_id);
CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_project_type_time ON metrics(project_id, metric_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_alert_configs_project ON alert_configs(project_id);
CREATE INDEX IF NOT EXISTS idx_alert_configs_enabled ON alert_configs(enabled);
CREATE INDEX IF NOT EXISTS idx_notification_channels_org ON notification_channels(organization_id);

-- Hypertable for time-series (if using TimescaleDB)
-- SELECT create_hypertable('metrics', 'timestamp', if_not_exists => TRUE);

-- Retention policy (auto-delete old metrics after 90 days)
CREATE OR REPLACE FUNCTION delete_old_metrics() RETURNS void AS $$
BEGIN
    DELETE FROM metrics WHERE timestamp < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- Trigger for updated_at
CREATE TRIGGER update_alert_configs_updated_at BEFORE UPDATE ON alert_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
