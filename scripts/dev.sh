#!/bin/bash

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory (assuming script is in scripts/ folder)
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"
echo -e "${BLUE}📂 Project root: ${PROJECT_ROOT}${NC}"

# Source environment variables
if [[ -f "cmd/api/.env" ]]; then
    source cmd/api/.env
    echo -e "${GREEN}✓${NC} Environment variables loaded from cmd/api/.env"
else
    echo -e "${RED}✗${NC} cmd/api/.env file not found!"
    exit 1
fi

echo -e "${BLUE}🚀 Starting Musical-Zoe Development Environment${NC}"
echo "=================================================="

# Function to check if PostgreSQL container is ready
check_postgres_ready() {
    echo -e "${YELLOW}⏳${NC} Waiting for PostgreSQL to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if PGPASSWORD="$POSTGRES_PASSWORD" psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q' 2>/dev/null; then
            echo -e "${GREEN}✓${NC} PostgreSQL is ready!"
            return 0
        fi
        
        echo -n "."
        sleep 2
        ((attempt++))
    done
    
    echo -e "${RED}✗${NC} PostgreSQL failed to start after $((max_attempts * 2)) seconds"
    docker-compose logs postgres
    exit 1
}

# Start Docker services
echo -e "${YELLOW}🐳${NC} Starting Docker services..."
docker-compose up -d

if [[ $? -eq 0 ]]; then
    echo -e "${GREEN}✓${NC} Docker services started successfully"
else
    echo -e "${RED}✗${NC} Failed to start Docker services"
    exit 1
fi

# Wait for PostgreSQL to be ready
check_postgres_ready

# Run migrations
echo -e "${YELLOW}📊${NC} Running database migrations..."
if goose -dir internal/sql/schema postgres "$MUSICALZOE_DB_DSN" up; then
    echo -e "${GREEN}✓${NC} Migrations completed successfully"
else
    echo -e "${RED}✗${NC} Migration failed"
    exit 1
fi

# Show status
echo ""
echo -e "${GREEN}🎉 Development environment is ready!${NC}"
echo "=================================================="
echo -e "Database: ${BLUE}$POSTGRES_DB${NC} on ${BLUE}localhost:$POSTGRES_PORT${NC}"
echo -e "User: ${BLUE}$POSTGRES_USER${NC}"
echo ""
echo "Available commands:"
echo -e "  ${YELLOW}make run/api${NC}          - Start the API server"
echo -e "  ${YELLOW}docker-compose logs${NC}   - View container logs"
echo -e "  ${YELLOW}docker-compose down${NC}   - Stop services"
echo -e "  ${YELLOW}docker-compose down -v${NC} - Stop services and remove data"
echo ""
echo "Database connection:"
echo -e "  ${YELLOW}PGPASSWORD=$POSTGRES_PASSWORD psql -U $POSTGRES_USER -d $POSTGRES_DB -h localhost${NC}"