-- ===========================================
-- CAPTCHA PLATFORM - Complete Database Schema
-- ===========================================
-- This file contains the complete schema for reference
-- Use migrations for actual database setup

-- Drop tables in reverse order of creation (due to foreign keys)
DROP TABLE IF EXISTS training_jobs;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS captcha_logs;
DROP TABLE IF EXISTS captcha_models;
DROP TABLE IF EXISTS users;

-- ===========================================
-- Table: users
-- ===========================================
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('user', 'admin', 'api_user') NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    email_verified_at TIMESTAMP NULL,
    last_login_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_users_email (email),
    INDEX idx_users_role (role),
    INDEX idx_users_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===========================================
-- Table: captcha_models
-- ===========================================
CREATE TABLE captcha_models (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    type ENUM('ocr', 'cnn', 'rnn', 'transformer', 'ensemble') NOT NULL,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    file_path VARCHAR(500) NOT NULL,
    file_size_bytes BIGINT UNSIGNED NOT NULL DEFAULT 0,
    accuracy DECIMAL(5, 4) NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSON NULL,
    description TEXT NULL,
    created_by BIGINT UNSIGNED NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_captcha_models_type (type),
    INDEX idx_captcha_models_is_active (is_active),
    INDEX idx_captcha_models_is_default (is_default),
    INDEX idx_captcha_models_accuracy (accuracy),
    
    CONSTRAINT fk_captcha_models_created_by 
        FOREIGN KEY (created_by) REFERENCES users(id) 
        ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===========================================
-- Table: captcha_logs
-- ===========================================
CREATE TABLE captcha_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NULL,
    model_id BIGINT UNSIGNED NULL,
    image_hash VARCHAR(64) NOT NULL,
    image_base64 LONGTEXT NULL,
    predicted_text VARCHAR(255) NULL,
    actual_text VARCHAR(255) NULL,
    confidence DECIMAL(5, 4) NULL,
    is_correct BOOLEAN NULL,
    processing_time_ms INT UNSIGNED NOT NULL DEFAULT 0,
    request_ip VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    error_message TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_captcha_logs_user_id (user_id),
    INDEX idx_captcha_logs_model_id (model_id),
    INDEX idx_captcha_logs_image_hash (image_hash),
    INDEX idx_captcha_logs_is_correct (is_correct),
    INDEX idx_captcha_logs_created_at (created_at),
    INDEX idx_captcha_logs_confidence (confidence),
    
    CONSTRAINT fk_captcha_logs_user_id 
        FOREIGN KEY (user_id) REFERENCES users(id) 
        ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_captcha_logs_model_id 
        FOREIGN KEY (model_id) REFERENCES captcha_models(id) 
        ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===========================================
-- Table: api_keys
-- ===========================================
CREATE TABLE api_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    key_prefix VARCHAR(8) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    scopes JSON NULL,
    rate_limit INT UNSIGNED NOT NULL DEFAULT 100,
    total_requests BIGINT UNSIGNED NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_api_keys_user_id (user_id),
    INDEX idx_api_keys_key_prefix (key_prefix),
    INDEX idx_api_keys_key_hash (key_hash),
    INDEX idx_api_keys_is_active (is_active),
    INDEX idx_api_keys_expires_at (expires_at),
    
    CONSTRAINT fk_api_keys_user_id 
        FOREIGN KEY (user_id) REFERENCES users(id) 
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ===========================================
-- Table: training_jobs
-- ===========================================
CREATE TABLE training_jobs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NULL,
    name VARCHAR(100) NOT NULL,
    status ENUM('pending', 'running', 'completed', 'failed', 'cancelled') NOT NULL DEFAULT 'pending',
    model_type ENUM('ocr', 'cnn', 'rnn', 'transformer', 'ensemble') NOT NULL,
    config JSON NOT NULL,
    dataset_path VARCHAR(500) NULL,
    dataset_size INT UNSIGNED NULL,
    progress DECIMAL(5, 2) NOT NULL DEFAULT 0.00,
    current_epoch INT UNSIGNED NULL,
    total_epochs INT UNSIGNED NULL,
    results JSON NULL,
    output_model_id BIGINT UNSIGNED NULL,
    error_message TEXT NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_training_jobs_user_id (user_id),
    INDEX idx_training_jobs_status (status),
    INDEX idx_training_jobs_model_type (model_type),
    INDEX idx_training_jobs_created_at (created_at),
    
    CONSTRAINT fk_training_jobs_user_id 
        FOREIGN KEY (user_id) REFERENCES users(id) 
        ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT fk_training_jobs_output_model_id 
        FOREIGN KEY (output_model_id) REFERENCES captcha_models(id) 
        ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;