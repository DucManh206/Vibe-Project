-- Seed: 001_seed_admin
-- Description: Create default admin user
-- Note: Change the password hash in production!

-- Default admin password is: Admin@123456
-- BCrypt hash with cost 12
INSERT INTO users (email, password_hash, role, is_active, email_verified_at) 
VALUES (
    'admin@captcha-platform.local',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.HQx1k1S1BVVwWy',
    'admin',
    TRUE,
    NOW()
) ON DUPLICATE KEY UPDATE updated_at = NOW();

-- Insert default OCR model placeholder
INSERT INTO captcha_models (name, type, version, file_path, is_active, is_default, description)
VALUES (
    'tesseract-default',
    'ocr',
    '1.0.0',
    '/models/tesseract/default',
    TRUE,
    TRUE,
    'Default Tesseract OCR model for basic text captcha'
) ON DUPLICATE KEY UPDATE updated_at = NOW();