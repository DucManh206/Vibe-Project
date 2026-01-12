# API Gateway Documentation

## Tổng Quan

API Gateway là entry point cho toàn bộ Captcha Platform. Nó xử lý routing, authentication, rate limiting, và proxy requests đến các backend services.

## Công Nghệ Sử Dụng

- **Language**: Go 1.21+
- **Framework**: Gin
- **Cache**: Redis (rate limiting)

## Cấu Trúc Thư Mục

```
gateway/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── config/              # Configuration
│   ├── handlers/            # Proxy handlers
│   ├── middleware/          # Middleware (auth, rate limit, cors)
│   └── proxy/               # Service proxies
├── pkg/
│   └── logger/              # Logging utilities
├── docs/
│   └── GATEWAY.md
├── Dockerfile
├── go.mod
└── go.sum
```

## Chức Năng

### 1. Request Routing

Gateway route requests đến các backend services:

| Path Pattern | Target Service |
|--------------|----------------|
| /api/v1/auth/* | Auth Service (port 8081) |
| /api/v1/api-keys/* | Auth Service |
| /api/v1/captcha/* | Captcha Service (port 8082) |

### 2. Authentication

Gateway hỗ trợ 2 phương thức authentication:

#### JWT Token
```
Authorization: Bearer <jwt_token>
```

#### API Key
```
X-API-Key: cp_xxxxxxxxxxxx
```

Khi validate thành công, Gateway forward thông tin user qua headers:
- `X-User-ID`: User ID
- `X-User-Email`: User email
- `X-User-Role`: User role

### 3. Rate Limiting

Rate limiting dựa trên client IP:
- Default: 100 requests/minute
- Configurable via environment variables
- Redis-based (fallback to in-memory)

Headers trả về:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 60
```

### 4. CORS

Tự động handle CORS preflight requests:
- Configurable allowed origins
- All standard methods (GET, POST, PUT, DELETE, OPTIONS)
- Custom headers support

### 5. Security Headers

Response tự động include:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security` (production only)

### 6. Request ID

Mỗi request được gán unique ID:
- Header: `X-Request-ID`
- Auto-generated nếu không có
- Forwarded đến backend services

## API Endpoints

### Health Check

```
GET /health
```

Response:
```json
{
  "status": "healthy",
  "service": "gateway",
  "time": "2024-01-01T00:00:00Z"
}
```

### Auth Routes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /api/v1/auth/register | No | Register user |
| POST | /api/v1/auth/login | No | Login |
| POST | /api/v1/auth/refresh | No | Refresh token |
| POST | /api/v1/auth/logout | No | Logout |
| GET | /api/v1/auth/me | JWT | Get current user |
| PUT | /api/v1/auth/me | JWT | Update user |
| PUT | /api/v1/auth/me/password | JWT | Change password |

### API Keys Routes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /api/v1/api-keys | JWT | List API keys |
| POST | /api/v1/api-keys | JWT | Create API key |
| DELETE | /api/v1/api-keys/:id | JWT | Delete API key |

### Captcha Routes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /api/v1/captcha/solve | JWT/API Key | Solve captcha |
| POST | /api/v1/captcha/solve/batch | JWT/API Key | Batch solve |
| GET | /api/v1/captcha/models | JWT | List models |
| POST | /api/v1/captcha/models/upload | JWT | Upload model |
| POST | /api/v1/captcha/train | JWT | Start training |
| GET | /api/v1/captcha/train/:job_id | JWT | Training status |
| GET | /api/v1/captcha/logs | JWT | Get logs |
| GET | /api/v1/captcha/stats | JWT | Get statistics |

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| GATEWAY_PORT | Server port | 8080 |
| GATEWAY_ENV | Environment | development |
| AUTH_SERVICE_URL | Auth service URL | http://localhost:8081 |
| CAPTCHA_SERVICE_URL | Captcha service URL | http://localhost:8082 |
| REDIS_HOST | Redis host | localhost |
| REDIS_PORT | Redis port | 6379 |
| REDIS_PASSWORD | Redis password | - |
| JWT_SECRET | JWT signing key | - |
| CORS_ORIGINS | Allowed origins | http://localhost:3000 |
| RATE_LIMIT_REQUESTS | Max requests | 100 |
| RATE_LIMIT_WINDOW_SECONDS | Window duration | 60 |
| LOG_LEVEL | Log level | debug |

## Error Responses

### 401 Unauthorized
```json
{
  "error": "unauthorized",
  "message": "Authorization header is required"
}
```

### 429 Too Many Requests
```json
{
  "error": "rate_limit_exceeded",
  "message": "Too many requests, please try again later"
}
```

### 502 Bad Gateway
```json
{
  "error": "service_unavailable",
  "message": "Backend service is unavailable"
}
```

## Development

### Run locally
```bash
cd backend/gateway
go run cmd/main.go
```

### With hot reload
```bash
air -c .air.toml
```

### Build
```bash
go build -o gateway ./cmd/main.go
```

## Logging

All requests được log với format:
```json
{
  "timestamp": "2024-01-01T00:00:00.000Z",
  "level": "info",
  "message": "HTTP Request",
  "method": "POST",
  "path": "/api/v1/captcha/solve",
  "status": 200,
  "latency": "50ms",
  "client_ip": "192.168.1.1",
  "request_id": "uuid-xxx"
}
```

## High Availability

Để chạy multiple instances:
1. Sử dụng Redis cho rate limiting (shared state)
2. Load balancer phía trước
3. Sticky sessions không cần thiết (stateless)