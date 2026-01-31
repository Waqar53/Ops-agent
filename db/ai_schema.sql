-- AI Cost Optimization schema

-- Cost recommendations
CREATE TABLE IF NOT EXISTS cost_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- rightsize, spot, reserved, cleanup, schedule
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    estimated_savings DECIMAL(10, 2) NOT NULL, -- USD per month
    confidence DECIMAL(3, 2) NOT NULL, -- 0.00 to 1.00
    priority VARCHAR(20) DEFAULT 'medium', -- low, medium, high
    action TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'pending', -- pending, applied, dismissed
    created_at TIMESTAMP DEFAULT NOW(),
    applied_at TIMESTAMP,
    dismissed_at TIMESTAMP
);

-- Resource optimization history
CREATE TABLE IF NOT EXISTS optimization_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    recommendation_id UUID REFERENCES cost_recommendations(id) ON DELETE SET NULL,
    action_type VARCHAR(50) NOT NULL,
    before_state JSONB NOT NULL,
    after_state JSONB NOT NULL,
    actual_savings DECIMAL(10, 2), -- Measured after optimization
    applied_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cost_recs_project ON cost_recommendations(project_id);
CREATE INDEX IF NOT EXISTS idx_cost_recs_status ON cost_recommendations(status);
CREATE INDEX IF NOT EXISTS idx_cost_recs_savings ON cost_recommendations(estimated_savings DESC);
CREATE INDEX IF NOT EXISTS idx_optimization_history_project ON optimization_history(project_id);
