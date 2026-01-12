-- Migration: 004_create_api_keys
-- Description: Create api_keys table for external API access
-- Created: 2024

-- Up Migration
CREATE TABLE IF NOT EXISTS api_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    key_prefix VARCHAR(8) NOT NULL COMMENT 'First 8 chars for identification',
    key_hash VARCHAR(255) NOT NULL COMMENT 'SHA256 hash of full key',
    scopes JSON NULL COMMENT 'Array of allowed scopes: ["captcha:solve", "captcha:train", "models:read"]',
    rate_limit INT UNSIGNED NOT NULL DEFAULT 100 COMMENT 'Requests per minute',
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

-- Down Migration (for rollback)
-- DROP TABLE IF EXISTS api_keys;