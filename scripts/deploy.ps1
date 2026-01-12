# ===========================================
# CAPTCHA PLATFORM - Deploy Script for Windows
# ===========================================
# Usage: .\scripts\deploy.ps1 [-SkipBackup] [-SkipBuild] [-Force]

param(
    [switch]$SkipBackup,
    [switch]$SkipBuild,
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$COMPOSE_FILE = "docker-compose.prod.yml"
$BACKUP_DIR = ".\backups"
$TIMESTAMP = Get-Date -Format "yyyyMMdd_HHmmss"

# Colors
function Write-Success { param($msg) Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Warning { param($msg) Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Error { param($msg) Write-Host "[ERROR] $msg" -ForegroundColor Red }
function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }

Write-Host ""
Write-Host "=========================================" -ForegroundColor Blue
Write-Host "  Captcha Platform - Deploy Script" -ForegroundColor Blue
Write-Host "=========================================" -ForegroundColor Blue
Write-Host ""

# Check .env
if (-not (Test-Path ".env")) {
    Write-Error ".env file not found. Please run .\scripts\setup.ps1 first"
    exit 1
}

# Load environment variables
Get-Content ".env" | ForEach-Object {
    if ($_ -match "^([^#][^=]+)=(.*)$") {
        [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}

# Confirmation
if (-not $Force) {
    Write-Warning "You are about to deploy to PRODUCTION"
    Write-Host "Environment: $($env:GATEWAY_ENV)"
    Write-Host ""
    $confirm = Read-Host "Are you sure you want to continue? (y/N)"
    if ($confirm -ne "y" -and $confirm -ne "Y") {
        Write-Host "Deploy cancelled."
        exit 0
    }
}

# Backup database
function Backup-Database {
    if ($SkipBackup) {
        Write-Warning "Skipping database backup..."
        return
    }
    
    Write-Info "Backing up database..."
    
    if (-not (Test-Path $BACKUP_DIR)) {
        New-Item -ItemType Directory -Path $BACKUP_DIR -Force | Out-Null
    }
    
    try {
        $container = docker-compose -f $COMPOSE_FILE ps -q mysql 2>$null
        if ($container) {
            $backupFile = "$BACKUP_DIR\backup_$TIMESTAMP.sql"
            docker exec $container mysqldump -u$($env:DB_USER) -p$($env:DB_PASSWORD) $($env:DB_NAME) > $backupFile
            
            # Compress backup
            Compress-Archive -Path $backupFile -DestinationPath "$backupFile.zip" -Force
            Remove-Item $backupFile
            
            Write-Success "Database backed up to $backupFile.zip"
            
            # Keep only last 7 backups
            Get-ChildItem "$BACKUP_DIR\backup_*.sql.zip" | 
                Sort-Object CreationTime -Descending | 
                Select-Object -Skip 7 | 
                Remove-Item -Force
        } else {
            Write-Warning "MySQL container not running, skipping backup"
        }
    } catch {
        Write-Warning "Backup failed: $_"
    }
}

# Pull latest code
function Pull-Latest {
    Write-Info "Pulling latest code..."
    
    if (Test-Path ".git") {
        git pull origin main
        Write-Success "Code updated"
    } else {
        Write-Warning "Not a git repository, skipping pull"
    }
}

# Build images
function Build-Images {
    if ($SkipBuild) {
        Write-Warning "Skipping image build..."
        return
    }
    
    Write-Info "Building Docker images..."
    docker-compose -f $COMPOSE_FILE build --no-cache
    Write-Success "Images built"
}

# Deploy services
function Deploy-Services {
    Write-Info "Deploying services..."
    
    # Stop old containers
    docker-compose -f $COMPOSE_FILE down --remove-orphans
    
    # Start new containers
    docker-compose -f $COMPOSE_FILE up -d
    
    Write-Success "Services deployed"
}

# Health check
function Test-Health {
    Write-Info "Running health checks..."
    
    $maxRetries = 30
    $retryCount = 0
    
    while ($retryCount -lt $maxRetries) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:$($env:GATEWAY_PORT ?? 8080)/health" -TimeoutSec 5 -UseBasicParsing -ErrorAction SilentlyContinue
            if ($response.StatusCode -eq 200) {
                Write-Success "API Gateway is healthy"
                return
            }
        } catch {
            $retryCount++
            Write-Host "Waiting for services to start... ($retryCount/$maxRetries)"
            Start-Sleep -Seconds 2
        }
    }
    
    Write-Error "Health check failed after $maxRetries attempts"
    Write-Host "Check logs with: docker-compose -f $COMPOSE_FILE logs"
    exit 1
}

# Cleanup
function Invoke-Cleanup {
    Write-Info "Cleaning up..."
    docker image prune -f | Out-Null
    Write-Success "Cleanup completed"
}

# Print status
function Show-Status {
    Write-Host ""
    Write-Host "=========================================" -ForegroundColor Green
    Write-Host "  Deployment Completed Successfully!" -ForegroundColor Green
    Write-Host "=========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Services status:" -ForegroundColor White
    docker-compose -f $COMPOSE_FILE ps
    Write-Host ""
    Write-Host "Useful commands:" -ForegroundColor White
    Write-Host "  View logs:     " -NoNewline; Write-Host "docker-compose -f $COMPOSE_FILE logs -f" -ForegroundColor Yellow
    Write-Host "  Restart:       " -NoNewline; Write-Host "docker-compose -f $COMPOSE_FILE restart" -ForegroundColor Yellow
    Write-Host "  Stop:          " -NoNewline; Write-Host "docker-compose -f $COMPOSE_FILE down" -ForegroundColor Yellow
    Write-Host ""
}

# Main execution
try {
    Backup-Database
    Pull-Latest
    Build-Images
    Deploy-Services
    Start-Sleep -Seconds 10
    Test-Health
    Invoke-Cleanup
    Show-Status
} catch {
    Write-Error "Deployment failed: $_"
    Write-Warning "Rolling back..."
    docker-compose -f $COMPOSE_FILE up -d
    exit 1
}