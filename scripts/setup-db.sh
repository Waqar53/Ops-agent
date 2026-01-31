#!/bin/bash

# OpsAgent Database Setup Script

set -e

echo "🗄️  Setting up OpsAgent PostgreSQL database..."

# Database configuration
DB_NAME="${POSTGRES_DB:-opsagent}"
DB_USER="${POSTGRES_USER:-postgres}"
DB_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL is not installed. Installing..."
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        brew install postgresql@15
        brew services start postgresql@15
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        sudo apt-get update
        sudo apt-get install -y postgresql-15
        sudo systemctl start postgresql
    fi
fi

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
until pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; do
    sleep 1
done

# Create database if it doesn't exist
echo "📊 Creating database '$DB_NAME'..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME"

# Run schema migration
echo "🔧 Running database migrations..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f db/schema.sql

# Seed sample data
echo "🌱 Seeding sample data..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME <<EOF
-- Sample project
INSERT INTO projects (id, name, slug, description, language, framework, git_repo, git_branch, status)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440000', 'Next.js E-Commerce', 'nextjs-ecommerce', 'Full-stack e-commerce application', 'Node.js', 'Next.js', 'https://github.com/example/nextjs-ecommerce', 'main', 'active'),
    ('550e8400-e29b-41d4-a716-446655440001', 'Python API', 'python-api', 'FastAPI microservice', 'Python', 'FastAPI', 'https://github.com/example/python-api', 'main', 'active')
ON CONFLICT (id) DO NOTHING;

-- Sample environments
INSERT INTO environments (id, project_id, name, type, status, url)
VALUES 
    ('660e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'production', 'production', 'active', 'https://nextjs-ecommerce.opsagent.dev'),
    ('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'staging', 'staging', 'active', 'https://staging.nextjs-ecommerce.opsagent.dev'),
    ('660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'production', 'production', 'active', 'https://python-api.opsagent.dev')
ON CONFLICT (project_id, name) DO NOTHING;

-- Sample deployments
INSERT INTO deployments (id, project_id, environment_id, version, git_commit, git_branch, strategy, status, deployed_by, completed_at, duration_seconds)
VALUES 
    ('770e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', '660e8400-e29b-41d4-a716-446655440000', 'v1.2.3', 'abc1234', 'main', 'rolling', 'success', 'john@example.com', NOW() - INTERVAL '2 hours', 145),
    ('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', '660e8400-e29b-41d4-a716-446655440000', 'v1.2.2', 'def5678', 'main', 'blue-green', 'success', 'jane@example.com', NOW() - INTERVAL '1 day', 132),
    ('770e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440002', 'v2.0.1', 'ghi9012', 'main', 'canary', 'success', 'bob@example.com', NOW() - INTERVAL '3 hours', 178)
ON CONFLICT (id) DO NOTHING;

-- Sample services
INSERT INTO services (id, project_id, name, type, version, status, endpoint)
VALUES 
    ('880e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'postgresql', 'database', '15', 'running', 'nextjs-ecommerce-db.opsagent.internal:5432'),
    ('880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'redis', 'cache', '7', 'running', 'nextjs-ecommerce-cache.opsagent.internal:6379'),
    ('880e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'postgresql', 'database', '15', 'running', 'python-api-db.opsagent.internal:5432')
ON CONFLICT (id) DO NOTHING;

-- Sample alerts
INSERT INTO alerts (id, project_id, alert_type, severity, title, message, status)
VALUES 
    ('990e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'cpu', 'warning', 'High CPU Usage', 'api-service is using 85% CPU', 'active')
ON CONFLICT (id) DO NOTHING;
EOF

echo "✅ Database setup complete!"
echo ""
echo "Database connection string:"
echo "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"
echo ""
echo "To connect:"
echo "psql postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"
