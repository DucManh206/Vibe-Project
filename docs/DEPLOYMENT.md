# 🚀 Deployment Guide

## Mục lục

- [Prerequisites](#prerequisites)
- [Development Setup](#development-setup)
- [Production Deployment](#production-deployment)
- [Docker Configuration](#docker-configuration)
- [Environment Variables](#environment-variables)
- [SSL/TLS Configuration](#ssltls-configuration)
- [Monitoring & Logging](#monitoring--logging)
- [Scaling](#scaling)
- [Backup & Recovery](#backup--recovery)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### System Requirements

- **OS:** Linux (Ubuntu 20.04+), macOS, Windows (WSL2)
- **CPU:** 2+ cores
- **RAM:** 4GB minimum, 8GB recommended
- **Disk:** 20GB+ free space

### Software Requirements

- Docker >= 20.10
- Docker Compose >= 2.0
- Git

### Optional (for local development)

- Node.js >= 18
- Go >= 1.21
- Rust >= 1.70

---

## Development Setup

### Quick Start

```bash
# Clone repository
git clone https://github.com/your-org/captcha-platform.git
cd captcha-platform

# Run setup script
chmod +x scripts/setup.sh
./scripts/setup.sh --build

# Access
# Frontend: http://localhost:3000
# API: http://localhost:8080
```

### Manual Setup

1. **Clone và cấu hình:**
```bash
git clone https://github.com/your-org/captcha-platform.git
cd captcha-platform
cp .env.example .env
```

2. **Chỉnh sửa .env:**
```bash
# Generate secure secrets
openssl rand -base64 32  # JWT_SECRET
openssl rand -base64 16  # DB_PASSWORD
```

3. **Start services:**
```bash
docker-compose up -d
```

4. **Verify:**
```bash
docker-compose ps
curl http://localhost:8080/health
```

### Local Development (without Docker)

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

**Auth Service:**
```bash
cd backend/auth
go mod download
go run cmd/main.go
```

**Captcha Service:**
```bash
cd backend/captcha
cargo run
```

---

## Production Deployment

### Pre-deployment Checklist

- [ ] Đã update tất cả dependencies
- [ ] Đã test trên staging environment
- [ ] Đã backup database
- [ ] Đã cấu hình SSL certificates
- [ ] Đã review security settings
- [ ] Đã setup monitoring

### Deploy Steps

```bash
# 1. SSH to server
ssh user@your-server

# 2. Clone/update code
git clone https://github.com/your-org/captcha-platform.git
# OR
git pull origin main

# 3. Configure environment
cp .env.example .env
vim .env  # Update production values

# 4. Deploy
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

### Docker Compose Production

```bash
# Use production compose file
docker-compose -f docker-compose.prod.yml up -d
```

---

## Docker Configuration

### Image Build

```bash
# Build all images
docker-compose build

# Build specific service
docker-compose build auth

# No cache build
docker-compose build --no-cache
```

### Container Management

```bash
# Start
docker-compose up -d

# Stop
docker-compose down

# Restart
docker-compose restart

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f gateway
```

### Resource Limits

Trong `docker-compose.prod.yml`:

```yaml
services:
  gateway:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
```

---

## Environment Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_HOST` | MySQL host | `mysql` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_NAME` | Database name | `captcha_platform` |
| `DB_USER` | Database user | `captcha_user` |
| `DB_PASSWORD` | Database password | `secure_password` |
| `JWT_SECRET` | JWT signing key (32+ chars) | `your-secret-key` |
| `REDIS_HOST` | Redis host | `redis` |
| `REDIS_PASSWORD` | Redis password | `redis_password` |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GATEWAY_PORT` | `8080` | API Gateway port |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `CORS_ORIGINS` | `*` | Allowed CORS origins |
| `RATE_LIMIT_REQUESTS` | `100` | Max requests per minute |

---

## SSL/TLS Configuration

### Using Nginx (Recommended)

1. **Install Nginx:**
```bash
sudo apt install nginx
```

2. **Configure SSL:**
```nginx
# /etc/nginx/sites-available/captcha-platform
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

3. **Get SSL Certificate (Let's Encrypt):**
```bash
sudo certbot --nginx -d api.yourdomain.com
```

### Using Traefik

```yaml
# docker-compose.prod.yml
services:
  traefik:
    image: traefik:v2.10
    command:
      - "--providers.docker=true"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.tlschallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.email=admin@yourdomain.com"
    ports:
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock

  gateway:
    labels:
      - "traefik.http.routers.gateway.rule=Host(`api.yourdomain.com`)"
      - "traefik.http.routers.gateway.tls=true"
      - "traefik.http.routers.gateway.tls.certresolver=letsencrypt"
```

---

## Monitoring & Logging

### Logging

Logs được output theo JSON format cho dễ parse:

```bash
# View all logs
docker-compose logs -f

# Filter by service
docker-compose logs -f gateway auth

# Export logs
docker-compose logs > logs/all_$(date +%Y%m%d).log
```

### Health Checks

```bash
# Check all services
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
```

### Prometheus Metrics (Optional)

```yaml
# docker-compose.monitoring.yml
services:
  prometheus:
    image: prom/prometheus
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana
    ports:
      - "3001:3000"
```

---

## Scaling

### Horizontal Scaling

```bash
# Scale gateway service
docker-compose up -d --scale gateway=3

# With load balancer
# Configure nginx/traefik upstream
```

### Database Scaling

1. **Read Replicas:** Setup MySQL replication
2. **Connection Pooling:** Use PgBouncer/ProxySQL
3. **Sharding:** For very large scale

---

## Backup & Recovery

### Database Backup

```bash
# Manual backup
docker exec captcha-mysql mysqldump -u root -p captcha_platform > backup.sql

# Automated backup (add to cron)
0 2 * * * /path/to/scripts/backup.sh
```

### Backup Script

```bash
#!/bin/bash
BACKUP_DIR=/backups
DATE=$(date +%Y%m%d_%H%M%S)

# Database
docker exec captcha-mysql mysqldump -u root -p$DB_ROOT_PASSWORD $DB_NAME | gzip > $BACKUP_DIR/db_$DATE.sql.gz

# Keep last 7 days
find $BACKUP_DIR -name "db_*.sql.gz" -mtime +7 -delete
```

### Recovery

```bash
# Restore database
gunzip < backup.sql.gz | docker exec -i captcha-mysql mysql -u root -p captcha_platform
```

---

## Troubleshooting

### Common Issues

#### 1. Container won't start

```bash
# Check logs
docker-compose logs service_name

# Check container status
docker-compose ps
```

#### 2. Database connection failed

```bash
# Verify MySQL is running
docker exec captcha-mysql mysqladmin ping -h localhost

# Check credentials
docker exec -it captcha-mysql mysql -u captcha_user -p
```

#### 3. Port conflicts

```bash
# Find what's using the port
lsof -i :8080
netstat -tuln | grep 8080
```

#### 4. Out of disk space

```bash
# Clean Docker resources
docker system prune -a
docker volume prune
```

### Debug Mode

```bash
# Run with debug logging
LOG_LEVEL=debug docker-compose up
```

### Support

- GitHub Issues: https://github.com/your-org/captcha-platform/issues
- Documentation: https://docs.captcha-platform.com