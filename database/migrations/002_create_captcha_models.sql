-- Migration: 002_create_captcha_models
-- Description: Create captcha_models table for ML model management
-- Created: 2024

-- Up Migration
CREATE TABLE IF NOT EXISTS captcha_models (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    type ENUM('ocr', 'cnn', 'rnn', 'transformer', 'ensemble') NOT NULL,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    file_path VARCHAR(500) NOT NULL,
    file_size_bytes BIGINT UNSIGNED NOT NULL DEFAULT 0,
    accuracy DECIMAL(5, 4) NULL COMMENT 'Accuracy from 0.0000 to 1.0000',
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSON NULL COMMENT 'Additional model configuration and stats',
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

-- Down Migration (for rollback)
-- DROP TABLE IF EXISTS captcha_models;