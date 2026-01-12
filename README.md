# 🔐 Captcha Platform

Nền tảng giải và huấn luyện captcha sử dụng kiến trúc microservices hiện đại.

## 📋 Mục Lục

- [Tổng Quan](#tổng-quan)
- [Kiến Trúc](#kiến-trúc)
- [Yêu Cầu Hệ Thống](#yêu-cầu-hệ-thống)
- [Cài Đặt](#cài-đặt)
- [Chạy Dự Án](#chạy-dự-án)
- [Cấu Trúc Thư Mục](#cấu-trúc-thư-mục)
- [API Documentation](#api-documentation)
- [Đóng Góp](#đóng-góp)
- [License](#license)

## 🎯 Tổng Quan

Captcha Platform là một hệ thống hoàn chỉnh để:
- ✅ Giải captcha dạng text tự động
- ✅ Huấn luyện models AI mới
- ✅ Quản lý và theo dõi hiệu suất models
- ✅ Cung cấp API cho bên thứ ba

## 🏗️ Kiến Trúc

```
┌─────────────────┐     ┌─────────────────┐
│   Frontend      │────▶│   API Gateway   │
│   (Next.js)     │     │   (Go)          │
└─────────────────┘     └────────┬────────┘
                                 │
                    ┌────────────┼────────────┐
                    ▼            ▼            ▼
            ┌───────────┐ ┌───────────┐ ┌───────────┐
            │   Auth    │ │  Captcha  │ │  Training │
            │  Service  │ │  Service  │ │  Module   │
            │   (Go)    │ │  (Rust)   │ │  (Rust)   │
            └─────┬─────┘ └─────┬─────┘ └─────┬─────┘
                  │             │             │
                  └─────────────┼─────────────┘
                                ▼
                    ┌─────────────────────┐
                    │   MySQL + Redis     │
                    └─────────────────────┘
```

## 💻 Yêu Cầu Hệ Thống

- **Docker** >= 20.10
- **Docker Compose** >= 2.0
- **Node.js** >= 18 (cho development)
- **Go** >= 1.21 (cho development)
- **Rust** >= 1.70 (cho development)

## 🚀 Cài Đặt

### 1. Clone repository

```bash
git clone https://github.com/your-org/captcha-platform.git
cd captcha-platform
```

### 2. Cấu hình environment

```bash
cp .env.example .env
# Chỉnh sửa file .env với các giá trị phù hợp
```

### 3. Chạy với Docker

```bash
# Development
docker-compose up -d

# Production
docker-compose -f docker-compose.prod.yml up -d
```

## 🏃 Chạy Dự Án

### Development Mode

```bash
# Chạy tất cả services
docker-compose up -d

# Xem logs
docker-compose logs -f

# Chạy frontend riêng (hot reload)
cd frontend && npm run dev

# Chạy backend riêng
cd backend/gateway && go run cmd/main.go
cd backend/auth && go run cmd/main.go
cd backend/captcha && cargo run
```

### Truy cập

- **Frontend**: http://localhost:3000
- **API Gateway**: http://localhost:8080
- **API Docs**: http://localhost:8080/docs

## 📁 Cấu Trúc Thư Mục

```
captcha-platform/
├── frontend/           # Next.js frontend application
├── backend/
│   ├── gateway/       # API Gateway (Go)
│   ├── auth/          # Authentication service (Go)
│   └── captcha/       # Captcha solving service (Rust)
├── database/
│   ├── migrations/    # SQL migrations
│   └── seeds/         # Seed data
├── docker/            # Docker configurations
├── scripts/           # Utility scripts
├── docs/              # Documentation
└── plans/             # Project planning documents
```

## 📚 API Documentation

Chi tiết API documentation có thể xem tại:
- [API Gateway Docs](docs/API.md)
- [Auth Service Docs](backend/auth/docs/AUTH.md)
- [Captcha Service Docs](backend/captcha/docs/CAPTCHA.md)

### Quick API Examples

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"securepassword"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"securepassword"}'

# Solve Captcha
curl -X POST http://localhost:8080/api/v1/captcha/solve \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"image_base64":"YOUR_BASE64_IMAGE"}'
```

## 🔒 Bảo Mật

- Tất cả passwords được hash bằng BCrypt
- JWT tokens cho authentication
- Rate limiting để ngăn chặn abuse
- Input validation và sanitization
- HTTPS enforced trong production
- Prepared statements để chống SQL injection

## 🌍 Internationalization (i18n)

Frontend hỗ trợ đa ngôn ngữ:
- 🇻🇳 Tiếng Việt (mặc định)
- 🇺🇸 English

## 🤝 Đóng Góp

Xem [CONTRIBUTING.md](docs/CONTRIBUTING.md) để biết cách đóng góp cho dự án.

## 📄 License

MIT License - Xem file [LICENSE](LICENSE) để biết thêm chi tiết.

---

**Made with ❤️ by Your Team**