# ===========================================
# Captcha Service Dockerfile (Rust)
# ===========================================

# Build stage
FROM rust:1.75-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache \
    musl-dev \
    pkgconfig \
    openssl-dev \
    openssl-libs-static \
    tesseract-ocr-dev \
    leptonica-dev \
    clang \
    llvm \
    cmake \
    make \
    g++

# Create a new empty shell project
RUN USER=root cargo new --bin captcha-service
WORKDIR /app/captcha-service

# Copy manifests
COPY Cargo.toml ./

# Build dependencies only (for caching)
RUN cargo build --release 2>/dev/null || true
RUN rm src/*.rs

# Copy source code
COPY src ./src

# Build the application
RUN cargo build --release

# ===========================================
# Development stage
# ===========================================
FROM rust:1.75-alpine AS development

WORKDIR /app

RUN apk add --no-cache \
    musl-dev \
    pkgconfig \
    openssl-dev \
    tesseract-ocr-dev \
    tesseract-ocr-data-eng \
    leptonica-dev

# Install cargo-watch for hot reload
RUN cargo install cargo-watch

COPY . .

CMD ["cargo", "watch", "-x", "run"]

# ===========================================
# Production stage
# ===========================================
FROM alpine:3.19 AS production

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    leptonica \
    libgcc

# Create non-root user
RUN adduser -D -g '' appuser

# Create models directory
RUN mkdir -p /app/models && chown -R appuser:appuser /app/models

# Copy binary from builder
COPY --from=builder /app/captcha-service/target/release/captcha-service /app/captcha-service

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Set environment variables
ENV TESSDATA_PREFIX=/usr/share/tessdata

# Run the application
CMD ["/app/captcha-service"]