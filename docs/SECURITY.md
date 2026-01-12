# 🔒 Security Guide

## Overview

Tài liệu này mô tả các biện pháp bảo mật được áp dụng trong Captcha Platform và hướng dẫn cấu hình an toàn cho production.

---

## Authentication & Authorization

### Password Security

- **Hashing:** BCrypt với cost factor 12
- **Minimum length:** 8 characters
- **Validation:** Chống brute force với rate limiting

```go
// Password hashing implementation
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
```

### JWT Tokens

- **Algorithm:** HS256 (HMAC-SHA256)
- **Access Token:** 24 hours expiry
- **Refresh Token:** 7 days expiry
- **Claims:** user_id, email, role

**Best Practices:**
- Sử dụng secret key ≥ 32 characters
- Rotate keys định kỳ
- Không lưu tokens trong localStorage cho sensitive apps

### API Key Security

- **Format:** `cp_` prefix + 64 hex characters
- **Storage:** SHA256 hash trong database
- **Rate limiting:** Per-key limits

---

## Input Validation

### Request Validation

Tất cả input được validate trước khi xử lý:

```go
// Example validation
type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email,max=255"`
    Password string `json:"password" binding:"required,min=8,max=128"`
}
```

### SQL Injection Prevention

- **Parameterized queries:** Tất cả queries sử dụng prepared statements
- **ORM:** Không sử dụng raw SQL

```go
// Safe query
db.Where("email = ?", email).First(&user)

// KHÔNG làm
// db.Raw("SELECT * FROM users WHERE email = '" + email + "'")
```

### XSS Prevention

- **Output encoding:** HTML escape tất cả user input
- **Content Security Policy:** Configured trong response headers
- **React:** Mặc định escape output

---

## Network Security

### HTTPS

Production **bắt buộc** sử dụng HTTPS:

```nginx
# Nginx configuration
server {
    listen 443 ssl http2;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    
    # HSTS
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
}
```

### Security Headers

Gateway tự động thêm các security headers:

```go
c.Header("X-Content-Type-Options", "nosniff")
c.Header("X-Frame-Options", "DENY")
c.Header("X-XSS-Protection", "1; mode=block")
c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
```

### CORS Configuration

```go
// Cấu hình CORS an toàn
CORS: CORSConfig{
    AllowedOrigins: []string{
        "https://yourdomain.com",
        "https://app.yourdomain.com",
    },
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders: []string{"Authorization", "Content-Type"},
    AllowCredentials: true,
}
```

---

## Rate Limiting

### Configuration

```env
RATE_LIMIT_REQUESTS=100      # Requests per window
RATE_LIMIT_WINDOW_SECONDS=60 # Window size
```

### Per-Endpoint Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| `/auth/login` | 5 | 1 min |
| `/auth/register` | 3 | 1 min |
| `/captcha/solve` | 100 | 1 min |
| Default | 100 | 1 min |

### Response Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
```

---

## Data Protection

### Sensitive Data

**KHÔNG bao giờ log hoặc expose:**
- Passwords
- JWT tokens
- API keys (full)
- Personal information

### Database Security

```yaml
# MySQL security settings
mysql:
  environment:
    MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
    MYSQL_DATABASE: ${DB_NAME}
    MYSQL_USER: ${DB_USER}
    MYSQL_PASSWORD: ${DB_PASSWORD}
```

**Best Practices:**
- Sử dụng non-root user cho application
- Giới hạn privileges cần thiết
- Enable SSL cho connections

### Encryption at Rest

- Database: MySQL encryption (optional)
- Files: Encrypted volumes
- Backups: Encrypted before storage

---

## Container Security

### Docker Best Practices

```dockerfile
# Run as non-root user
FROM golang:1.21-alpine AS builder
RUN adduser -D -g '' appuser

FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
USER appuser
```

### Resource Limits

```yaml
services:
  gateway:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
```

### Network Isolation

```yaml
networks:
  captcha-network:
    driver: bridge
    internal: true  # No external access
  
  frontend-network:
    driver: bridge
```

---

## Secrets Management

### Environment Variables

**KHÔNG commit secrets vào git:**

```gitignore
# .gitignore
.env
*.pem
*.key
secrets/
```

### Production Secrets

Sử dụng một trong các giải pháp:
- Docker Secrets
- HashiCorp Vault
- AWS Secrets Manager
- Kubernetes Secrets

```yaml
# Docker Secrets example
services:
  auth:
    secrets:
      - db_password
      - jwt_secret

secrets:
  db_password:
    external: true
  jwt_secret:
    external: true
```

---

## Logging & Monitoring

### Audit Logging

Log các sự kiện quan trọng:
- Authentication attempts
- Authorization failures
- Data access
- Configuration changes

```go
log.Info("User login",
    "user_id", user.ID,
    "ip", c.ClientIP(),
    "user_agent", c.GetHeader("User-Agent"),
)
```

### Security Events

Monitor và alert:
- Multiple failed logins
- Unusual API usage patterns
- Unauthorized access attempts
- Rate limit violations

---

## Vulnerability Management

### Dependency Scanning

```bash
# Go
go list -m all | nancy sleuth

# Rust
cargo audit

# Node.js
npm audit
```

### Regular Updates

```bash
# Update dependencies
go get -u ./...
cargo update
npm update
```

### Security Patches

- Subscribe to security advisories
- Apply critical patches within 24h
- Test patches in staging first

---

## Incident Response

### Contact

Security issues: security@yourdomain.com

### Response Process

1. **Identify:** Confirm and assess the issue
2. **Contain:** Isolate affected systems
3. **Eradicate:** Remove the threat
4. **Recover:** Restore normal operations
5. **Learn:** Document and improve

### Responsible Disclosure

If you discover a security vulnerability:
1. Email security@yourdomain.com
2. Provide detailed reproduction steps
3. Allow 90 days for fix before disclosure

---

## Compliance Checklist

### Before Production

- [ ] All passwords are hashed with BCrypt
- [ ] JWT secrets are ≥32 characters
- [ ] HTTPS is enabled and enforced
- [ ] Security headers are configured
- [ ] Rate limiting is enabled
- [ ] CORS is properly configured
- [ ] Database credentials are secure
- [ ] No secrets in code/git
- [ ] Logging excludes sensitive data
- [ ] Dependencies are up to date
- [ ] Container runs as non-root
- [ ] Backups are encrypted

### Regular Audits

- [ ] Monthly dependency updates
- [ ] Quarterly penetration testing
- [ ] Annual security review
- [ ] Continuous vulnerability scanning