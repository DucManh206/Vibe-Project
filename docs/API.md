# 📚 API Documentation

## Overview

Captcha Platform cung cấp RESTful API để tích hợp dịch vụ giải captcha vào ứng dụng của bạn.

**Base URL:** `http://localhost:8080/api/v1`

## Authentication

### JWT Authentication

Hầu hết các endpoints yêu cầu xác thực bằng JWT token.

```http
Authorization: Bearer <your-jwt-token>
```

### API Key Authentication

Một số endpoints (như solve captcha) hỗ trợ xác thực bằng API key:

```http
X-API-Key: <your-api-key>
```

---

## Endpoints

### Authentication

#### Register

Đăng ký tài khoản mới.

```http
POST /auth/register
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (201):**
```json
{
  "id": 1,
  "email": "user@example.com",
  "role": "user",
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Login

Đăng nhập và nhận JWT tokens.

```http
POST /auth/login
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (200):**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "user",
    "is_active": true
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 86400
}
```

#### Refresh Token

Làm mới access token.

```http
POST /auth/refresh
```

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### Get Current User

Lấy thông tin user hiện tại.

```http
GET /auth/me
Authorization: Bearer <token>
```

#### Change Password

Đổi mật khẩu.

```http
PUT /auth/me/password
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "current_password": "oldpassword",
  "new_password": "newpassword123"
}
```

---

### Captcha Solving

#### Solve Captcha

Giải một captcha.

```http
POST /captcha/solve
Authorization: Bearer <token>
# OR
X-API-Key: <api-key>
```

**Request Body:**
```json
{
  "image_base64": "iVBORw0KGgoAAAANSUhEUgAA...",
  "model": "cnn",
  "preprocess": {
    "grayscale": true,
    "denoise": true,
    "threshold": 128
  }
}
```

**Response (200):**
```json
{
  "text": "ABC123",
  "confidence": 0.95,
  "model": "cnn",
  "processing_time_ms": 125
}
```

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| image_base64 | string | Yes | Base64 encoded image (với hoặc không có data URL prefix) |
| model | string | No | Tên model (default: auto-select) |
| preprocess | object | No | Tùy chọn tiền xử lý |

**Preprocess Options:**
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| grayscale | boolean | true | Chuyển ảnh sang grayscale |
| denoise | boolean | false | Giảm nhiễu ảnh |
| threshold | number | null | Ngưỡng nhị phân (0-255) |
| resize_width | number | null | Resize chiều rộng |
| resize_height | number | null | Resize chiều cao |

#### Batch Solve

Giải nhiều captcha cùng lúc.

```http
POST /captcha/solve/batch
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "images": [
    { "image_base64": "..." },
    { "image_base64": "...", "model": "ocr" }
  ]
}
```

**Response (200):**
```json
{
  "results": [
    {
      "index": 0,
      "success": true,
      "result": {
        "text": "ABC123",
        "confidence": 0.95,
        "model": "cnn",
        "processing_time_ms": 125
      }
    },
    {
      "index": 1,
      "success": true,
      "result": {
        "text": "XYZ789",
        "confidence": 0.88,
        "model": "ocr",
        "processing_time_ms": 89
      }
    }
  ],
  "total_time_ms": 250
}
```

---

### Models

#### List Models

Liệt kê tất cả models.

```http
GET /captcha/models
Authorization: Bearer <token>
```

**Response (200):**
```json
[
  {
    "id": 1,
    "name": "cnn",
    "type": "cnn",
    "version": "1.0.0",
    "accuracy": 0.95,
    "is_active": true,
    "is_default": true,
    "description": "CNN model for text captcha"
  }
]
```

#### Upload Model

Tải lên model mới.

```http
POST /captcha/models/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

---

### API Keys

#### List API Keys

Liệt kê tất cả API keys.

```http
GET /api-keys
Authorization: Bearer <token>
```

**Response (200):**
```json
[
  {
    "id": 1,
    "name": "Production Key",
    "key_prefix": "cp_a1b2c3d",
    "rate_limit": 100,
    "total_requests": 1500,
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

#### Create API Key

Tạo API key mới.

```http
POST /api-keys
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "name": "My API Key",
  "rate_limit": 100,
  "expires_in": 30,
  "scopes": ["captcha:solve"]
}
```

**Response (201):**
```json
{
  "id": 2,
  "name": "My API Key",
  "key": "cp_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0",
  "key_prefix": "cp_a1b2c3d",
  "rate_limit": 100,
  "is_active": true,
  "expires_at": "2024-02-14T10:30:00Z"
}
```

> ⚠️ **Lưu ý:** `key` chỉ được hiển thị một lần. Hãy lưu trữ an toàn!

#### Delete API Key

Xóa API key.

```http
DELETE /api-keys/:id
Authorization: Bearer <token>
```

---

### Statistics

#### Get Stats

Lấy thống kê tổng quan.

```http
GET /captcha/stats
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "total_requests": 10500,
  "successful_requests": 9975,
  "failed_requests": 525,
  "average_processing_time_ms": 145.5,
  "accuracy_rate": 0.95,
  "models_count": 3,
  "active_models_count": 2
}
```

---

### Logs

#### Get Logs

Lấy lịch sử xử lý.

```http
GET /captcha/logs?page=1&limit=20&model_id=1
Authorization: Bearer <token>
```

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| page | number | 1 | Số trang |
| limit | number | 20 | Số kết quả mỗi trang |
| model_id | number | - | Lọc theo model ID |
| is_correct | boolean | - | Lọc theo kết quả đúng/sai |

---

### Training

#### Start Training

Bắt đầu huấn luyện model mới.

```http
POST /captcha/train
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "name": "my-custom-model",
  "model_type": "cnn",
  "config": {
    "epochs": 100,
    "batch_size": 32,
    "learning_rate": 0.001,
    "validation_split": 0.2
  },
  "dataset_path": "/path/to/dataset"
}
```

#### Get Training Status

Lấy trạng thái huấn luyện.

```http
GET /captcha/train/:job_id
Authorization: Bearer <token>
```

---

## Error Responses

Tất cả errors trả về theo format:

```json
{
  "error": "error_code",
  "message": "Human readable message",
  "details": "Optional additional details"
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `validation_error` | 400 | Invalid request body |
| `unauthorized` | 401 | Missing or invalid auth |
| `forbidden` | 403 | Insufficient permissions |
| `not_found` | 404 | Resource not found |
| `rate_limit_exceeded` | 429 | Too many requests |
| `internal_error` | 500 | Server error |

---

## Rate Limiting

- Default: 100 requests/minute
- Custom limits per API key
- Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

---

## SDK Examples

### cURL

```bash
# Solve captcha
curl -X POST http://localhost:8080/api/v1/captcha/solve \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"image_base64":"iVBORw0KGgo..."}'
```

### Python

```python
import requests
import base64

API_URL = "http://localhost:8080/api/v1"
API_KEY = "cp_your_api_key"

# Read image
with open("captcha.png", "rb") as f:
    image_b64 = base64.b64encode(f.read()).decode()

# Solve
response = requests.post(
    f"{API_URL}/captcha/solve",
    headers={"X-API-Key": API_KEY},
    json={"image_base64": image_b64}
)

result = response.json()
print(f"Text: {result['text']}, Confidence: {result['confidence']}")
```

### JavaScript

```javascript
const solveCaptcha = async (imageBase64) => {
  const response = await fetch('http://localhost:8080/api/v1/captcha/solve', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'cp_your_api_key',
    },
    body: JSON.stringify({ image_base64: imageBase64 }),
  });
  
  return response.json();
};