-- Migration: 005_create_training_jobs
-- Description: Create training_jobs table for ML model training management
-- Created: 2024

-- Up Migration
CREATE TABLE IF NOT EXISTS training_jobs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NULL,
    name VARCHAR(100) NOT NULL,
    status ENUM('pending', 'running', 'completed', 'failed', 'cancelled') NOT NULL DEFAULT 'pending',
    model_type ENUM('ocr', 'cnn', 'rnn', 'transformer', 'ensemble') NOT NULL,
    
    -- Training configuration
    config JSON NOT NULL COMMENT 'Training hyperparameters and settings',
    
    -- Dataset information
    dataset_path VARCHAR(500) NULL,
    dataset_size INT UNSIGNED NULL COMMENT 'Number of training samples',
    
    -- Progress tracking
    progress DECIMAL(5, 2) NOT NULL DEFAULT 0.00 COMMENT 'Progress percentage 0.00 to 100.00',
    current_epoch INT UNSIGNED NULL,
    total_epochs INT UNSIGNED NULL,
    
    -- Results
    results JSON NULL COMMENT 'Training metrics: loss, accuracy, etc.',
    output_model_id BIGINT UNSIGNED NULL COMMENT 'ID of created model after successful training',
    error_message TEXT NULL,
    
    -- Timing
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

-- Down Migration (for rollback)
-- DROP TABLE IF EXISTS training_jobs;