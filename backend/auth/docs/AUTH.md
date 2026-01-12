# Auth Service Documentation

## Tổng Quan

Auth Service là một microservice xử lý toàn bộ logic authentication và authorization cho Captcha Platform.

## Công Nghệ Sử Dụng

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: MySQL 8.0
- **Password Hashing**: BCrypt
- **Token**: JWT (HS256)

## Cấu Trúc Thư Mục

```
auth/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── config/              # Configuration
│   ├── database/            # Database connection
│   ├── handlers/            # HTTP handlers
│   ├── middleware/          # Auth middleware
│   ├── models/              # Data models
│   ├── repository/          # Database operations
│   └── services/            # Business logic
├── pkg/
│   ├── jwt/                 # JWT utilities
│   └── logger/              # Logging utilities
├── docs/
│   └── AUTH.md
├── Dockerfile
├── go.mod
└── go.sum
```

## API Endpoints

### Public Endpoints

#### POST /auth/register
Đăng ký user mới.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response (201):**
```json
{
  "id": 1,
  "email": "user@example.com",
  "role": "user",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### POST /auth/login
Đăng nhập và nhận JWT token.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response (200):**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "user"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 86400
}
```

#### POST /auth/refresh
Refresh access token.

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Protected Endpoints

Tất cả protected endpoints yêu cầu header:
```
Authorization: Bearer <access_token>
```

#### GET /auth/me
Lấy thông tin user hiện tại.

#### PUT /auth/me
Cập nhật thông tin user.

#### PUT /auth/me/password
Đổi mật khẩu.

**Request Body:**
```json
{
  "current_password": "oldpassword",
  "new_password": "newpassword"
}
```

### API Keys Endpoints

#### GET /api-keys
Danh sách API keys của user.

#### POST /api-keys
Tạo API key mới.

**Request Body:**
```json
{
  "name": "Production API Key",
  "rate_limit": 100,
  "expires_in": 365
}
```

**Response (201):**
```json
{
  "id": 1,
  "name": "Production API Key",
  "key": "cp_abc123...",
  "key_prefix": "cp_abc1234",
  "rate_limit": 100,
  "created_at": "2024-01-01T00:00:00Z"
}
```

⚠️ **Lưu ý**: `key` chỉ được trả về một lần duy nhất khi tạo.

#### DELETE /api-keys/:id
Xóa API key.

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| AUTH_SERVICE_PORT | Server port | 8081 |
| DB_HOST | MySQL host | localhost |
| DB_PORT | MySQL port | 3306 |
| DB_NAME | Database name | captcha_platform |
| DB_USER | Database user | - |
| DB_PASSWORD | Database password | - |
| JWT_SECRET | JWT signing key (min 32 chars) | - |
| JWT_EXPIRES_IN | Access token expiry | 24h |
| JWT_REFRESH_EXPIRES_IN | Refresh token expiry | 7d |
| BCRYPT_COST | BCrypt cost factor | 12 |
| LOG_LEVEL | Log level | debug |

## Security

### Password Hashing
- BCrypt với cost factor 12
- Không bao giờ lưu password plaintext

### JWT Tokens
- Access token: 24h expiry
- Refresh token: 7d expiry
- Sử dụng HS256 algorithm
- Include user ID, email, role trong claims

### API Keys
- SHA256 hash được lưu trong database
- Key prefix để identify (không dùng để authenticate)
- Rate limiting per key
- Expiration date optional

## Error Responses

Tất cả errors trả về format:
```json
{
  "error": "error_code",
  "message": "Human readable message",
  "details": "Optional details"
}
```

### Error Codes
- `validation_error`: Invalid request body
- `invalid_credentials`: Wrong email/password
- `user_exists`: Email already registered
- `user_inactive`: Account deactivated
- `unauthorized`: Missing/invalid token
- `not_found`: Resource not found
- `internal_error`: Server error

## Development

### Run locally
```bash
cd backend/auth
go run cmd/main.go
```

### With hot reload
```bash
air -c .air.toml
```

### Build
```bash
go build -o auth ./cmd/main.go