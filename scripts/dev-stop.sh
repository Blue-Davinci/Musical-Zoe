#!/bin/bash

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo -e "${BLUE}🛑 Stopping Musical-Zoe Development Environment${NC}"
echo "=================================================="

# Stop services
echo -e "${YELLOW}🐳${NC} Stopping Docker services..."
docker-compose down

if [[ $? -eq 0 ]]; then
    echo -e "${GREEN}✓${NC} Services stopped successfully"
else
    echo -e "${RED}✗${NC} Failed to stop services"
    exit 1
fi

echo -e "${GREEN}🏁 Development environment stopped${NC}"
echo ""
echo "To start again, run:"
echo -e "  ${YELLOW}make dev/setup${NC}"
