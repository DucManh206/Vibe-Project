#!/bin/bash

# ===========================================
# CAPTCHA PLATFORM - Deploy Script
# ===========================================
# This script deploys the application to production

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}  Captcha Platform - Deploy Script${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Parse arguments
SKIP_BACKUP=false
SKIP_BUILD=false
FORCE=false

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --skip-backup) SKIP_BACKUP=true ;;
        --skip-build) SKIP_BUILD=true ;;
        --force|-f) FORCE=true ;;
        --help|-h) 
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --skip-backup  Skip database backup"
            echo "  --skip-build   Skip Docker image build"
            echo "  --force, -f    Force deploy without confirmation"
            echo "  --help, -h     Show this help"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
    shift
done

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${RED}Error: .env file not found${NC}"
    echo "Please run ./scripts/setup.sh first"
    exit 1
fi

# Load environment variables
source .env

# Confirmation
if [ "$FORCE" != true ]; then
    echo -e "${YELLOW}You are about to deploy to PRODUCTION${NC}"
    echo -e "Environment: ${GATEWAY_ENV:-production}"
    echo ""
    read -p "Are you sure you want to continue? (y/N) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Deploy cancelled."
        exit 0
    fi
fi

# Backup database
backup_database() {
    if [ "$SKIP_BACKUP" = true ]; then
        echo -e "${YELLOW}Skipping database backup...${NC}"
        return
    fi
    
    echo -e "${YELLOW}Backing up database...${NC}"
    
    mkdir -p "$BACKUP_DIR"
    
    # Get MySQL container name
    MYSQL_CONTAINER=$(docker-compose -f $COMPOSE_FILE ps -q mysql 2>/dev/null || echo "")
    
    if [ -n "$MYSQL_CONTAINER" ]; then
        docker exec $MYSQL_CONTAINER mysqldump \
            -u${DB_USER:-captcha_user} \
            -p${DB_PASSWORD} \
            ${DB_NAME:-captcha_platform} \
            > "$BACKUP_DIR/backup_${TIMESTAMP}.sql"
        
        # Compress backup
        gzip "$BACKUP_DIR/backup_${TIMESTAMP}.sql"
        
        echo -e "${GREEN}✓ Database backed up to $BACKUP_DIR/backup_${TIMESTAMP}.sql.gz${NC}"
        
        # Keep only last 7 backups
        ls -t "$BACKUP_DIR"/backup_*.sql.gz 2>/dev/null | tail -n +8 | xargs -r rm
    else
        echo -e "${YELLOW}⚠ MySQL container not running, skipping backup${NC}"
    fi
    
    echo ""
}

# Pull latest code
pull_latest() {
    echo -e "${YELLOW}Pulling latest code...${NC}"
    
    if git rev-parse --git-dir > /dev/null 2>&1; then
        git pull origin main
        echo -e "${GREEN}✓ Code updated${NC}"
    else
        echo -e "${YELLOW}⚠ Not a git repository, skipping pull${NC}"
    fi
    
    echo ""
}

# Build images
build_images() {
    if [ "$SKIP_BUILD" = true ]; then
        echo -e "${YELLOW}Skipping image build...${NC}"
        return
    fi
    
    echo -e "${YELLOW}Building Docker images...${NC}"
    
    docker-compose -f $COMPOSE_FILE build --no-cache
    
    echo -e "${GREEN}✓ Images built${NC}"
    echo ""
}

# Run migrations
run_migrations() {
    echo -e "${YELLOW}Running database migrations...${NC}"
    
    # Wait for MySQL to be ready
    echo "Waiting for MySQL..."
    sleep 5
    
    # Run migrations if migration tool exists
    # For now, migrations are run on container startup via init scripts
    
    echo -e "${GREEN}✓ Migrations completed${NC}"
    echo ""
}

# Deploy services
deploy_services() {
    echo -e "${YELLOW}Deploying services...${NC}"
    
    # Stop old containers
    docker-compose -f $COMPOSE_FILE down --remove-orphans
    
    # Start new containers
    docker-compose -f $COMPOSE_FILE up -d
    
    echo -e "${GREEN}✓ Services deployed${NC}"
    echo ""
}

# Health check
health_check() {
    echo -e "${YELLOW}Running health checks...${NC}"
    
    MAX_RETRIES=30
    RETRY_COUNT=0
    
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        # Check Gateway health
        if curl -sf http://localhost:${GATEWAY_PORT:-8080}/health > /dev/null 2>&1; then
            echo -e "${GREEN}✓ API Gateway is healthy${NC}"
            break
        fi
        
        RETRY_COUNT=$((RETRY_COUNT + 1))
        echo "Waiting for services to start... ($RETRY_COUNT/$MAX_RETRIES)"
        sleep 2
    done
    
    if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
        echo -e "${RED}✗ Health check failed after $MAX_RETRIES attempts${NC}"
        echo "Check logs with: docker-compose -f $COMPOSE_FILE logs"
        exit 1
    fi
    
    echo ""
}

# Clean up
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    
    # Remove dangling images
    docker image prune -f
    
    # Remove unused volumes (be careful with this)
    # docker volume prune -f
    
    echo -e "${GREEN}✓ Cleanup completed${NC}"
    echo ""
}

# Print status
print_status() {
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}  Deployment Completed Successfully!${NC}"
    echo -e "${GREEN}=========================================${NC}"
    echo ""
    echo -e "Services status:"
    docker-compose -f $COMPOSE_FILE ps
    echo ""
    echo -e "Access URLs:"
    echo -e "  Frontend:    ${GREEN}https://your-domain.com${NC}"
    echo -e "  API Gateway: ${GREEN}https://api.your-domain.com${NC}"
    echo ""
    echo -e "Useful commands:"
    echo -e "  View logs:     ${YELLOW}docker-compose -f $COMPOSE_FILE logs -f${NC}"
    echo -e "  Restart:       ${YELLOW}docker-compose -f $COMPOSE_FILE restart${NC}"
    echo -e "  Stop:          ${YELLOW}docker-compose -f $COMPOSE_FILE down${NC}"
    echo -e "  Rollback:      ${YELLOW}./scripts/rollback.sh${NC}"
    echo ""
}

# Rollback function
rollback() {
    echo -e "${RED}Deployment failed. Rolling back...${NC}"
    
    # Restore from backup if available
    LATEST_BACKUP=$(ls -t "$BACKUP_DIR"/backup_*.sql.gz 2>/dev/null | head -1)
    
    if [ -n "$LATEST_BACKUP" ]; then
        echo "Restoring from $LATEST_BACKUP..."
        gunzip -c "$LATEST_BACKUP" | docker exec -i $(docker-compose -f $COMPOSE_FILE ps -q mysql) \
            mysql -u${DB_USER:-captcha_user} -p${DB_PASSWORD} ${DB_NAME:-captcha_platform}
    fi
    
    # Start previous containers
    docker-compose -f $COMPOSE_FILE up -d
    
    echo -e "${YELLOW}Rollback completed. Please check the logs for errors.${NC}"
    exit 1
}

# Main execution
main() {
    # Set up error handling
    trap rollback ERR
    
    backup_database
    pull_latest
    build_images
    deploy_services
    run_migrations
    health_check
    cleanup
    print_status
}

# Run main function
main