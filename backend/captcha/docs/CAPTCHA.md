# Captcha Service Documentation

## Tổng Quan

Captcha Service là một microservice xử lý việc giải và huấn luyện captcha, được viết bằng Rust để tối ưu performance.

## Công Nghệ Sử Dụng

- **Language**: Rust 1.75+
- **Framework**: Actix-web
- **Database**: MySQL 8.0 (via SQLx)
- **OCR Engine**: Tesseract
- **ML Runtime**: tract-onnx

## Cấu Trúc Thư Mục

```
captcha/
├── src/
│   ├── main.rs              # Entry point
│   ├── config.rs            # Configuration
│   ├── db.rs                # Database operations
│   ├── error.rs             # Error types
│   ├── models.rs            # Data models
│   ├── api/
│   │   ├── mod.rs
│   │   ├── captcha.rs       # Solve endpoints
│   │   ├── models.rs        # Model management
│   │   ├── training.rs      # Training endpoints
│   │   ├── logs.rs          # Logs endpoints
│   │   ├── stats.rs         # Statistics
│   │   └── health.rs        # Health check
│   └── solvers/
│       ├── mod.rs           # Solver manager
│       ├── ocr.rs           # Tesseract OCR
│       ├── cnn.rs           # CNN model
│       └── preprocessor.rs  # Image preprocessing
├── models/                  # Pre-trained models
├── docs/
│   └── CAPTCHA.md
├── Dockerfile
├── Cargo.toml
└── Cargo.lock
```

## Solvers

### 1. OCR Solver (Tesseract)

Sử dụng Tesseract OCR cho captcha đơn giản:
- Grayscale conversion
- Thresholding
- Noise reduction
- Text extraction

**Ưu điểm**: Nhanh, không cần GPU
**Nhược điểm**: Độ chính xác thấp với captcha phức tạp

### 2. CNN Solver (Deep Learning)

Sử dụng Convolutional Neural Network:
- ONNX model format
- CTC decoding
- Support batch processing

**Ưu điểm**: Độ chính xác cao
**Nhược điểm**: Cần model pre-trained

### 3. Ensemble Solver

Kết hợp nhiều solvers và chọn kết quả confidence cao nhất.

## API Endpoints

### Solve Captcha

#### POST /captcha/solve
Giải một captcha.

**Request Body:**
```json
{
  "image_base64": "iVBORw0KGgo...",
  "model": "cnn",
  "preprocess": {
    "grayscale": true,
    "threshold": 128,
    "denoise": true
  }
}
```

**Response (200):**
```json
{
  "text": "AB12CD",
  "confidence": 0.95,
  "model": "cnn",
  "processing_time_ms": 45
}
```

#### POST /captcha/solve/batch
Giải nhiều captcha cùng lúc.

**Request Body:**
```json
{
  "images": [
    { "image_base64": "iVBORw0KGgo..." },
    { "image_base64": "iVBORw0KGgo..." }
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
        "text": "AB12CD",
        "confidence": 0.95,
        "model": "cnn",
        "processing_time_ms": 45
      }
    },
    {
      "index": 1,
      "success": false,
      "error": "Invalid image format"
    }
  ],
  "total_time_ms": 120
}
```

### Models

#### GET /captcha/models
Danh sách models.

**Response:**
```json
{
  "models": [
    {
      "id": 1,
      "name": "cnn-v1",
      "model_type": "cnn",
      "version": "1.0.0",
      "accuracy": 0.95,
      "is_active": true,
      "is_default": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

#### POST /captcha/models/upload
Upload model mới.

**Request Body:**
```json
{
  "name": "my-model",
  "model_type": "cnn",
  "version": "1.0.0",
  "description": "Custom trained model",
  "model_data": "base64_encoded_onnx_file"
}
```

### Training

#### POST /captcha/train
Bắt đầu training job.

**Request Body:**
```json
{
  "name": "my-training-job",
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

**Response (201):**
```json
{
  "id": 1,
  "name": "my-training-job",
  "status": "pending",
  "message": "Training job created successfully"
}
```

#### GET /captcha/train/:job_id
Lấy trạng thái training.

**Response:**
```json
{
  "id": 1,
  "name": "my-training-job",
  "status": "running",
  "model_type": "cnn",
  "progress": 45.5,
  "current_epoch": 45,
  "total_epochs": 100,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Logs

#### GET /captcha/logs
Lấy danh sách logs.

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)
- `model_id`: Filter by model
- `is_correct`: Filter by correctness

**Response:**
```json
{
  "logs": [
    {
      "id": 1,
      "user_id": 1,
      "model_id": 1,
      "image_hash": "abc123...",
      "predicted_text": "AB12CD",
      "is_correct": true,
      "processing_time_ms": 45,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 20,
  "total_pages": 5
}
```

### Statistics

#### GET /captcha/stats
Thống kê tổng quan.

**Response:**
```json
{
  "total_requests": 10000,
  "successful_requests": 9500,
  "failed_requests": 500,
  "average_processing_time_ms": 50.5,
  "accuracy_rate": 0.92,
  "models_count": 5,
  "active_models_count": 3
}
```

### Health Check

#### GET /health
Kiểm tra service health.

**Response:**
```json
{
  "status": "healthy",
  "service": "captcha",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "0.1.0",
  "models_loaded": 2
}
```

## Image Preprocessing

### PreprocessOptions

```json
{
  "grayscale": true,      // Convert to grayscale
  "threshold": 128,       // Binary threshold (0-255)
  "denoise": true,        // Apply median filter
  "resize_width": 200,    // Resize width
  "resize_height": 50     // Resize height
}
```

### Preprocessing Pipeline

1. **Grayscale**: Chuyển ảnh sang grayscale
2. **Resize**: Resize về kích thước chuẩn của model
3. **Denoise**: Áp dụng median filter để giảm noise
4. **Threshold**: Binary threshold để tách background
5. **Contrast**: Tăng contrast nếu cần

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| CAPTCHA_SERVICE_PORT | Server port | 8082 |
| DB_HOST | MySQL host | localhost |
| DB_PORT | MySQL port | 3306 |
| DB_NAME | Database name | captcha_platform |
| DB_USER | Database user | - |
| DB_PASSWORD | Database password | - |
| MODELS_PATH | Path to models | /app/models |
| MAX_IMAGE_SIZE_MB | Max image size | 10 |
| LOG_LEVEL | Log level | info |
| TESSDATA_PREFIX | Tesseract data path | /usr/share/tessdata |

## Error Responses

### Error Format
```json
{
  "error": "error_code",
  "message": "Human readable message"
}
```

### Error Codes
- `invalid_image`: Invalid or corrupted image
- `image_too_large`: Image exceeds size limit
- `model_not_found`: Requested model not found
- `model_load_error`: Failed to load model
- `processing_error`: Error during processing
- `timeout`: Processing timeout
- `database_error`: Database error
- `bad_request`: Invalid request

## Model Training

### Dataset Format

Training dataset structure:
```
dataset/
├── images/
│   ├── image001.png
│   ├── image002.png
│   └── ...
└── labels.csv
```

labels.csv format:
```csv
filename,label
image001.png,AB12CD
image002.png,XY98ZW
```

### Training Configuration

```json
{
  "epochs": 100,
  "batch_size": 32,
  "learning_rate": 0.001,
  "validation_split": 0.2,
  "options": {
    "augmentation": true,
    "early_stopping": true,
    "patience": 10
  }
}
```

## Development

### Run locally
```bash
cd backend/captcha
cargo run
```

### With hot reload
```bash
cargo watch -x run
```

### Build release
```bash
cargo build --release
```

### Run tests
```bash
cargo test
```

## Performance Tips

1. **Batch Processing**: Sử dụng `/solve/batch` cho nhiều ảnh
2. **Model Selection**: Chọn model phù hợp với loại captcha
3. **Preprocessing**: Điều chỉnh preprocess options
4. **Caching**: Kết quả được hash để có thể cache
5. **Connection Pool**: SQLx tự động manage pool

## Adding New Solvers

1. Implement `CaptchaSolver` trait:
```rust
#[async_trait]
pub trait CaptchaSolver: Send + Sync {
    async fn solve(&self, image: &DynamicImage, options: Option<&PreprocessOptions>) -> CaptchaResult<SolveResult>;
    fn name(&self) -> &str;
    fn is_ready(&self) -> bool;
}
```

2. Register trong `SolverManager`
3. Enable trong config