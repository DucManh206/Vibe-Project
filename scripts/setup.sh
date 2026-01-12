#!/bin/bash

# ===========================================
# CAPTCHA PLATFORM - Setup Script
# ===========================================
# This script sets up the development environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}  Captcha Platform - Setup Script${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""

# Check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}Checking prerequisites...${NC}"
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}Docker is not installed. Please install Docker first.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker installed${NC}"
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        echo -e "${RED}Docker Compose is not installed. Please install Docker Compose first.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker Compose installed${NC}"
    
    # Check Node.js (optional for local development)
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node -v)
        echo -e "${GREEN}✓ Node.js installed ($NODE_VERSION)${NC}"
    else
        echo -e "${YELLOW}⚠ Node.js not found (optional for local frontend development)${NC}"
    fi
    
    # Check Go (optional for local development)
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        echo -e "${GREEN}✓ Go installed ($GO_VERSION)${NC}"
    else
        echo -e "${YELLOW}⚠ Go not found (optional for local backend development)${NC}"
    fi
    
    # Check Rust (optional for local development)
    if command -v cargo &> /dev/null; then
        RUST_VERSION=$(rustc --version | awk '{print $2}')
        echo -e "${GREEN}✓ Rust installed ($RUST_VERSION)${NC}"
    else
        echo -e "${YELLOW}⚠ Rust not found (optional for local captcha service development)${NC}"
    fi
    
    echo ""
}

# Setup environment
setup_environment() {
    echo -e "${YELLOW}Setting up environment...${NC}"
    
    # Copy .env.example to .env if not exists
    if [ ! -f .env ]; then
        cp .env.example .env
        echo -e "${GREEN}✓ Created .env file from .env.example${NC}"
        echo -e "${YELLOW}⚠ Please review and update .env with your configuration${NC}"
    else
        echo -e "${GREEN}✓ .env file already exists${NC}"
    fi
    
    # Generate secure secrets if not set
    if grep -q "your_super_secret_jwt_key_min_32_chars" .env; then
        JWT_SECRET=$(openssl rand -base64 32 | tr -d '\n')
        sed -i.bak "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env 2>/dev/null || \
        sed -i '' "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env
        echo -e "${GREEN}✓ Generated secure JWT_SECRET${NC}"
    fi
    
    if grep -q "your_secure_password_here" .env; then
        DB_PASSWORD=$(openssl rand -base64 16 | tr -d '\n' | tr -d '/')
        sed -i.bak "s/DB_PASSWORD=.*/DB_PASSWORD=$DB_PASSWORD/" .env 2>/dev/null || \
        sed -i '' "s/DB_PASSWORD=.*/DB_PASSWORD=$DB_PASSWORD/" .env
        echo -e "${GREEN}✓ Generated secure DB_PASSWORD${NC}"
    fi
    
    if grep -q "your_redis_password_here" .env; then
        REDIS_PASSWORD=$(openssl rand -base64 16 | tr -d '\n' | tr -d '/')
        sed -i.bak "s/REDIS_PASSWORD=.*/REDIS_PASSWORD=$REDIS_PASSWORD/" .env 2>/dev/null || \
        sed -i '' "s/REDIS_PASSWORD=.*/REDIS_PASSWORD=$REDIS_PASSWORD/" .env
        echo -e "${GREEN}✓ Generated secure REDIS_PASSWORD${NC}"
    fi
    
    # Remove backup files
    rm -f .env.bak
    
    echo ""
}

# Create required directories
create_directories() {
    echo -e "${YELLOW}Creating required directories...${NC}"
    
    mkdir -p backend/captcha/models
    mkdir -p logs
    mkdir -p data/mysql
    mkdir -p data/redis
    
    echo -e "${GREEN}✓ Directories created${NC}"
    echo ""
}

# Build Docker images
build_images() {
    echo -e "${YELLOW}Building Docker images...${NC}"
    
    docker-compose build --no-cache
    
    echo -e "${GREEN}✓ Docker images built${NC}"
    echo ""
}

# Start services
start_services() {
    echo -e "${YELLOW}Starting services...${NC}"
    
    docker-compose up -d
    
    echo -e "${GREEN}✓ Services started${NC}"
    echo ""
    
    # Wait for services to be ready
    echo -e "${YELLOW}Waiting for services to be ready...${NC}"
    sleep 10
    
    # Check service health
    check_services
}

# Check services health
check_services() {
    echo -e "${YELLOW}Checking service health...${NC}"
    
    # Check MySQL
    if docker-compose exec -T mysql mysqladmin ping -h localhost &> /dev/null; then
        echo -e "${GREEN}✓ MySQL is running${NC}"
    else
        echo -e "${RED}✗ MySQL is not responding${NC}"
    fi
    
    # Check Redis
    if docker-compose exec -T redis redis-cli ping &> /dev/null; then
        echo -e "${GREEN}✓ Redis is running${NC}"
    else
        echo -e "${RED}✗ Redis is not responding${NC}"
    fi
    
    # Check Gateway
    if curl -s http://localhost:8080/health &> /dev/null; then
        echo -e "${GREEN}✓ API Gateway is running${NC}"
    else
        echo -e "${YELLOW}⚠ API Gateway is starting up...${NC}"
    fi
    
    # Check Auth Service
    if curl -s http://localhost:8081/health &> /dev/null; then
        echo -e "${GREEN}✓ Auth Service is running${NC}"
    else
        echo -e "${YELLOW}⚠ Auth Service is starting up...${NC}"
    fi
    
    # Check Captcha Service
    if curl -s http://localhost:8082/health &> /dev/null; then
        echo -e "${GREEN}✓ Captcha Service is running${NC}"
    else
        echo -e "${YELLOW}⚠ Captcha Service is starting up...${NC}"
    fi
    
    # Check Frontend
    if curl -s http://localhost:3000 &> /dev/null; then
        echo -e "${GREEN}✓ Frontend is running${NC}"
    else
        echo -e "${YELLOW}⚠ Frontend is starting up...${NC}"
    fi
    
    echo ""
}

# Print access information
print_info() {
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}  Setup Complete!${NC}"
    echo -e "${GREEN}=========================================${NC}"
    echo ""
    echo -e "Access the application at:"
    echo -e "  Frontend:    ${GREEN}http://localhost:3000${NC}"
    echo -e "  API Gateway: ${GREEN}http://localhost:8080${NC}"
    echo -e "  API Docs:    ${GREEN}http://localhost:8080/docs${NC}"
    echo ""
    echo -e "Database credentials (check .env for details):"
    echo -e "  MySQL:  localhost:3306"
    echo -e "  Redis:  localhost:6379"
    echo ""
    echo -e "Useful commands:"
    echo -e "  View logs:       ${YELLOW}docker-compose logs -f${NC}"
    echo -e "  Stop services:   ${YELLOW}docker-compose down${NC}"
    echo -e "  Restart:         ${YELLOW}docker-compose restart${NC}"
    echo ""
}

# Main execution
main() {
    check_prerequisites
    setup_environment
    create_directories
    
    if [ "$1" == "--build" ] || [ "$1" == "-b" ]; then
        build_images
    fi
    
    if [ "$1" != "--no-start" ]; then
        start_services
    fi
    
    print_info
}

# Run main function
main "$@"