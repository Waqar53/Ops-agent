-- CI/CD schema

-- Builds table
CREATE TABLE IF NOT EXISTS builds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    git_commit VARCHAR(255) NOT NULL,
    git_branch VARCHAR(255) NOT NULL,
    git_author VARCHAR(255),
    git_message TEXT,
    status VARCHAR(50) DEFAULT 'pending', -- pending, running, success, failed, canceled
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration INTEGER, -- seconds
    log_url TEXT,
    artifact_url TEXT, -- Docker image URL
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Preview environments
CREATE TABLE IF NOT EXISTS preview_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    pull_request_id VARCHAR(255) NOT NULL,
    git_branch VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'creating', -- creating, ready, destroying, destroyed
    created_at TIMESTAMP DEFAULT NOW(),
    destroyed_at TIMESTAMP
);

-- Webhooks configuration
CREATE TABLE IF NOT EXISTS webhook_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL, -- github, gitlab, bitbucket
    secret VARCHAR(255) NOT NULL,
    events JSONB DEFAULT '[]', -- ["push", "pull_request", "tag"]
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Webhook deliveries (for debugging)
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_config_id UUID REFERENCES webhook_configs(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'pending', -- pending, processing, success, failed
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    processed_at TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_builds_project ON builds(project_id);
CREATE INDEX IF NOT EXISTS idx_builds_status ON builds(status);
CREATE INDEX IF NOT EXISTS idx_builds_created_at ON builds(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_preview_envs_project ON preview_environments(project_id);
CREATE INDEX IF NOT EXISTS idx_preview_envs_pr ON preview_environments(pull_request_id);
CREATE INDEX IF NOT EXISTS idx_webhook_configs_project ON webhook_configs(project_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_config ON webhook_deliveries(webhook_config_id);
