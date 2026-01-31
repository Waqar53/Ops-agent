-- Seed sample data for OpsAgent

-- Sample projects
INSERT INTO projects (id, name, slug, description, language, framework, git_repo, git_branch, status)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440000', 'Next.js E-Commerce', 'nextjs-ecommerce', 'Full-stack e-commerce application with PostgreSQL and Redis', 'Node.js', 'Next.js', 'https://github.com/example/nextjs-ecommerce', 'main', 'active'),
    ('550e8400-e29b-41d4-a716-446655440001', 'Python API', 'python-api', 'FastAPI microservice with PostgreSQL', 'Python', 'FastAPI', 'https://github.com/example/python-api', 'main', 'active'),
    ('550e8400-e29b-41d4-a716-446655440002', 'Go Microservice', 'go-microservice', 'High-performance Go API', 'Go', 'Gin', 'https://github.com/example/go-api', 'main', 'active')
ON CONFLICT (id) DO NOTHING;

-- Sample environments
INSERT INTO environments (id, project_id, name, type, status, url)
VALUES 
    ('660e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'production', 'production', 'active', 'https://nextjs-ecommerce.opsagent.dev'),
    ('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'staging', 'staging', 'active', 'https://staging.nextjs-ecommerce.opsagent.dev'),
    ('660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'production', 'production', 'active', 'https://python-api.opsagent.dev'),
    ('660e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', 'production', 'production', 'active', 'https://go-api.opsagent.dev')
ON CONFLICT (project_id, name) DO NOTHING;

-- Sample deployments
INSERT INTO deployments (id, project_id, environment_id, version, git_commit, git_branch, strategy, status, deployed_by, deployed_at, completed_at, duration_seconds)
VALUES 
    ('770e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', '660e8400-e29b-41d4-a716-446655440000', 'v1.2.3', 'abc1234', 'main', 'rolling', 'success', 'john@example.com', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours' + INTERVAL '145 seconds', 145),
    ('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', '660e8400-e29b-41d4-a716-446655440000', 'v1.2.2', 'def5678', 'main', 'blue-green', 'success', 'jane@example.com', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '132 seconds', 132),
    ('770e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440002', 'v2.0.1', 'ghi9012', 'main', 'canary', 'success', 'bob@example.com', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours' + INTERVAL '178 seconds', 178),
    ('770e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440000', '660e8400-e29b-41d4-a716-446655440001', 'v1.3.0', 'jkl3456', 'develop', 'rolling', 'success', 'alice@example.com', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes' + INTERVAL '98 seconds', 98),
    ('770e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440003', 'v3.1.0', 'mno7890', 'main', 'direct', 'success', 'charlie@example.com', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours' + INTERVAL '67 seconds', 67)
ON CONFLICT (id) DO NOTHING;

-- Sample services
INSERT INTO services (id, project_id, name, type, version, status, endpoint)
VALUES 
    ('880e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'postgresql', 'database', '15', 'running', 'nextjs-ecommerce-db.opsagent.internal:5432'),
    ('880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'redis', 'cache', '7', 'running', 'nextjs-ecommerce-cache.opsagent.internal:6379'),
    ('880e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'postgresql', 'database', '15', 'running', 'python-api-db.opsagent.internal:5432'),
    ('880e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', 'postgresql', 'database', '15', 'running', 'go-api-db.opsagent.internal:5432')
ON CONFLICT (id) DO NOTHING;

-- Sample alerts
INSERT INTO alerts (id, project_id, environment_id, alert_type, severity, title, message, status)
VALUES 
    ('990e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', '660e8400-e29b-41d4-a716-446655440000', 'cpu', 'warning', 'High CPU Usage', 'api-service is using 85% CPU', 'active'),
    ('990e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440002', 'latency', 'critical', 'High Latency Detected', 'P99 latency is 1.2s (threshold: 500ms)', 'active')
ON CONFLICT (id) DO NOTHING;

-- Sample cost data
INSERT INTO costs (project_id, environment_id, date, amount, breakdown)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440000', '660e8400-e29b-41d4-a716-446655440000', CURRENT_DATE, 42.50, '{"compute": 25.00, "database": 12.50, "cache": 5.00}'),
    ('550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440002', CURRENT_DATE, 28.75, '{"compute": 18.75, "database": 10.00}'),
    ('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440003', CURRENT_DATE, 15.20, '{"compute": 12.00, "database": 3.20}')
ON CONFLICT (project_id, environment_id, date) DO UPDATE SET amount = EXCLUDED.amount;
