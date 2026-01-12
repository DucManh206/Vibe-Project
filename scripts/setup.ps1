# ===========================================
# CAPTCHA PLATFORM - Setup Script for Windows
# ===========================================
# Usage: .\scripts\setup.ps1 [-Build] [-NoBuild]

param(
    [switch]$Build,
    [switch]$NoBuild
)

$ErrorActionPreference = "Stop"

# Colors
function Write-Success { param($msg) Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Warning { param($msg) Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Error { param($msg) Write-Host "[ERROR] $msg" -ForegroundColor Red }
function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }

Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  Captcha Platform - Setup Script" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Check prerequisites
Write-Info "Checking prerequisites..."

# Check Docker
try {
    $dockerVersion = docker --version 2>$null
    if ($dockerVersion) {
        Write-Success "Docker installed: $dockerVersion"
    }
} catch {
    Write-Error "Docker is not installed. Please install Docker Desktop first."
    Write-Host "Download: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}

# Check if Docker is running
try {
    docker info 2>$null | Out-Null
    Write-Success "Docker is running"
} catch {
    Write-Error "Docker is not running. Please start Docker Desktop."
    exit 1
}

# Check Docker Compose
try {
    $composeVersion = docker compose version 2>$null
    if ($composeVersion) {
        Write-Success "Docker Compose installed"
    }
} catch {
    try {
        $composeVersion = docker-compose --version 2>$null
        if ($composeVersion) {
            Write-Success "Docker Compose installed (legacy)"
        }
    } catch {
        Write-Error "Docker Compose is not installed."
        exit 1
    }
}

Write-Host ""

# Setup environment
Write-Info "Setting up environment..."

# Copy .env if not exists
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Success "Created .env file from .env.example"
    Write-Warning "Please review and update .env with your configuration"
    
    # Generate secure secrets
    $jwtSecret = -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 32 | ForEach-Object {[char]$_})
    $dbPassword = -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 16 | ForEach-Object {[char]$_})
    $redisPassword = -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 16 | ForEach-Object {[char]$_})
    
    # Update .env file
    $envContent = Get-Content ".env"
    $envContent = $envContent -replace "JWT_SECRET=.*", "JWT_SECRET=$jwtSecret"
    $envContent = $envContent -replace "DB_PASSWORD=your_secure_password_here", "DB_PASSWORD=$dbPassword"
    $envContent = $envContent -replace "REDIS_PASSWORD=your_redis_password_here", "REDIS_PASSWORD=$redisPassword"
    $envContent | Set-Content ".env"
    
    Write-Success "Generated secure secrets"
} else {
    Write-Success ".env file already exists"
}

Write-Host ""

# Create directories
Write-Info "Creating required directories..."

$directories = @(
    "backend\captcha\models",
    "logs",
    "backups",
    "data\mysql",
    "data\redis"
)

foreach ($dir in $directories) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }
}
Write-Success "Directories created"

Write-Host ""

# Build images if requested
if ($Build -and -not $NoBuild) {
    Write-Info "Building Docker images (this may take a while)..."
    docker-compose build --no-cache
    Write-Success "Docker images built"
    Write-Host ""
}

# Start services
if (-not $NoBuild) {
    Write-Info "Starting services..."
    docker-compose up -d
    Write-Success "Services started"
    Write-Host ""
    
    # Wait for services
    Write-Info "Waiting for services to be ready..."
    Start-Sleep -Seconds 15
    
    # Health checks
    Write-Info "Checking service health..."
    
    $services = @{
        "MySQL" = $null  # Internal check
        "Redis" = $null  # Internal check
        "API Gateway" = "http://localhost:8080/health"
        "Auth Service" = "http://localhost:8081/health"
        "Captcha Service" = "http://localhost:8082/health"
        "Frontend" = "http://localhost:3000"
    }
    
    foreach ($service in @("API Gateway", "Auth Service", "Captcha Service", "Frontend")) {
        try {
            $response = Invoke-WebRequest -Uri $services[$service] -TimeoutSec 5 -UseBasicParsing -ErrorAction SilentlyContinue
            if ($response.StatusCode -eq 200) {
                Write-Success "$service is running"
            }
        } catch {
            Write-Warning "$service is starting up..."
        }
    }
}

Write-Host ""
Write-Host "=========================================" -ForegroundColor Green
Write-Host "  Setup Complete!" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Access the application at:" -ForegroundColor White
Write-Host "  Frontend:    " -NoNewline; Write-Host "http://localhost:3000" -ForegroundColor Cyan
Write-Host "  API Gateway: " -NoNewline; Write-Host "http://localhost:8080" -ForegroundColor Cyan
Write-Host "  API Docs:    " -NoNewline; Write-Host "http://localhost:8080/docs" -ForegroundColor Cyan
Write-Host ""
Write-Host "Useful commands:" -ForegroundColor White
Write-Host "  View logs:       " -NoNewline; Write-Host "docker-compose logs -f" -ForegroundColor Yellow
Write-Host "  Stop services:   " -NoNewline; Write-Host "docker-compose down" -ForegroundColor Yellow
Write-Host "  Restart:         " -NoNewline; Write-Host "docker-compose restart" -ForegroundColor Yellow
Write-Host "  Rebuild:         " -NoNewline; Write-Host ".\scripts\setup.ps1 -Build" -ForegroundColor Yellow
Write-Host ""