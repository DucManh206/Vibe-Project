-- Migration: 003_create_captcha_logs
-- Description: Create captcha_logs table for tracking captcha solving history
-- Created: 2024

-- Up Migration
CREATE TABLE IF NOT EXISTS captcha_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NULL,
    model_id BIGINT UNSIGNED NULL,
    image_hash VARCHAR(64) NOT NULL COMMENT 'SHA256 hash of original image',
    image_base64 LONGTEXT NULL COMMENT 'Base64 encoded image (optional storage)',
    predicted_text VARCHAR(255) NULL,
    actual_text VARCHAR(255) NULL COMMENT 'For training feedback',
    confidence DECIMAL(5, 4) NULL COMMENT 'Model confidence from 0.0000 to 1.0000',
    is_correct BOOLEAN NULL COMMENT 'NULL if not verified, TRUE/FALSE after feedback',
    processing_time_ms INT UNSIGNED NOT NULL DEFAULT 0,
    request_ip VARCHAR(45) NULL COMMENT 'IPv4 or IPv6 address',
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

-- Down Migration (for rollback)
-- DROP TABLE IF EXISTS captcha_logs;