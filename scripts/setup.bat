@echo off
REM ===========================================
REM CAPTCHA PLATFORM - Setup Script for Windows
REM ===========================================

echo =========================================
echo   Captcha Platform - Setup Script
echo =========================================
echo.

REM Check Docker
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker is not installed. Please install Docker Desktop first.
    pause
    exit /b 1
)
echo [OK] Docker installed

REM Check Docker Compose
docker compose version >nul 2>&1
if %errorlevel% neq 0 (
    docker-compose --version >nul 2>&1
    if %errorlevel% neq 0 (
        echo [ERROR] Docker Compose is not installed.
        pause
        exit /b 1
    )
)
echo [OK] Docker Compose installed

REM Copy .env if not exists
if not exist .env (
    copy .env.example .env >nul
    echo [OK] Created .env file from .env.example
    echo [WARN] Please review and update .env with your configuration
) else (
    echo [OK] .env file already exists
)

REM Create directories
if not exist backend\captcha\models mkdir backend\captcha\models
if not exist logs mkdir logs
if not exist backups mkdir backups
echo [OK] Directories created

REM Build and start
echo.
echo Building and starting services...
echo.

if "%1"=="--build" (
    docker-compose build --no-cache
)

docker-compose up -d

echo.
echo Waiting for services to start...
timeout /t 10 /nobreak >nul

REM Health check
echo.
echo Checking services...
curl -s http://localhost:8080/health >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] API Gateway is running
) else (
    echo [WAIT] API Gateway is starting up...
)

curl -s http://localhost:3000 >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Frontend is running
) else (
    echo [WAIT] Frontend is starting up...
)

echo.
echo =========================================
echo   Setup Complete!
echo =========================================
echo.
echo Access the application at:
echo   Frontend:    http://localhost:3000
echo   API Gateway: http://localhost:8080
echo.
echo Useful commands:
echo   View logs:       docker-compose logs -f
echo   Stop services:   docker-compose down
echo   Restart:         docker-compose restart
echo.

pause