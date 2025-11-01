#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Territory Service Bot - Quick Start  ${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}Creating .env file from template...${NC}"
    cat > .env << 'EOF'
# Log Configuration
TS_LOG_LEVEL=debug

# PostgreSQL Configuration
TS_POSTGRESQL_USER=admin
TS_POSTGRESQL_PASSWORD=postgres
TS_POSTGRESQL_HOST=localhost
TS_POSTGRESQL_DATABASE=api

# Telegram Configuration
TS_TELEGRAM_BOT_TOKEN=your_bot_token_here
EOF
    echo -e "${GREEN}✓ Created .env file${NC}"
    echo -e "${YELLOW}⚠ Please update TS_TELEGRAM_BOT_TOKEN in .env with your actual bot token${NC}"
    echo ""
fi

# Start PostgreSQL
echo -e "${BLUE}[1/3] Starting PostgreSQL container...${NC}"
docker-compose up -d
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ PostgreSQL started successfully${NC}"
else
    echo -e "${RED}✗ Failed to start PostgreSQL${NC}"
    exit 1
fi

# Wait for PostgreSQL to be ready
echo -e "${BLUE}Waiting for PostgreSQL to be ready...${NC}"
sleep 8

# Run database seeding
echo -e "${BLUE}[2/3] Seeding database with test data...${NC}"
go run scripts/seed.go
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database seeded successfully${NC}"
else
    echo -e "${RED}✗ Failed to seed database${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Setup Complete! 🎉${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo -e "  1. Update your Telegram bot token in .env file"
echo -e "  2. Run the application: ${YELLOW}make run${NC} or ${YELLOW}go run cmd/main.go${NC}"
echo -e "  3. Access database: ${YELLOW}make psql${NC}"
echo -e "  4. View logs: ${YELLOW}make logs${NC}"
echo ""
echo -e "${BLUE}Available commands:${NC}"
echo -e "  ${YELLOW}make help${NC}        - Show all available commands"
echo -e "  ${YELLOW}make dev${NC}         - Start development environment"
echo -e "  ${YELLOW}make seed-clean${NC}  - Reset and reseed database"
echo ""
echo -e "${BLUE}Test Data Summary:${NC}"
echo -e "  - 3 Congregations"
echo -e "  - 5 Territory Groups"
echo -e "  - 4 Users (2 admins, 2 publishers)"
echo -e "  - 6 Territories"
echo -e "  - 3 Territory Notes"
echo ""

