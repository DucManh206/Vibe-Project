# ===========================================
# Captcha Service Dockerfile (Rust)
# ===========================================

# Build stage
FROM rust:1.91-slim-bookworm AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    libtesseract-dev \
    libleptonica-dev \
    clang \
    libclang-dev \
    cmake \
    make \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Create a new empty shell project
RUN USER=root cargo new --bin captcha-service
WORKDIR /app/captcha-service

# Copy manifests
COPY Cargo.toml Cargo.lock ./

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
FROM rust:1.91-slim-bookworm AS development

WORKDIR /app

RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    libtesseract-dev \
    libleptonica-dev \
    tesseract-ocr \
    tesseract-ocr-eng \
    && rm -rf /var/lib/apt/lists/*

# Install cargo-watch for hot reload
RUN cargo install cargo-watch

COPY . .

CMD ["cargo", "watch", "-x", "run"]

# ===========================================
# Production stage
# ===========================================
FROM debian:bookworm-slim AS production

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    wget \
    tesseract-ocr \
    tesseract-ocr-eng \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN adduser --disabled-password --gecos "" appuser

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