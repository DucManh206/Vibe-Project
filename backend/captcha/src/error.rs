//! Error types for Captcha Service

use actix_web::{HttpResponse, ResponseError};
use std::fmt;

/// Custom error types for the captcha service
#[derive(Debug)]
pub enum CaptchaError {
    /// Invalid image data
    InvalidImage(String),
    /// Image too large
    ImageTooLarge,
    /// Model not found
    ModelNotFound(String),
    /// Model loading failed
    ModelLoadError(String),
    /// Processing timeout
    Timeout,
    /// Database error
    DatabaseError(String),
    /// Internal processing error
    ProcessingError(String),
    /// Invalid request
    BadRequest(String),
    /// Unauthorized
    Unauthorized,
    /// Not found
    NotFound(String),
}

impl fmt::Display for CaptchaError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            CaptchaError::InvalidImage(msg) => write!(f, "Invalid image: {}", msg),
            CaptchaError::ImageTooLarge => write!(f, "Image exceeds maximum allowed size"),
            CaptchaError::ModelNotFound(name) => write!(f, "Model not found: {}", name),
            CaptchaError::ModelLoadError(msg) => write!(f, "Failed to load model: {}", msg),
            CaptchaError::Timeout => write!(f, "Processing timeout"),
            CaptchaError::DatabaseError(msg) => write!(f, "Database error: {}", msg),
            CaptchaError::ProcessingError(msg) => write!(f, "Processing error: {}", msg),
            CaptchaError::BadRequest(msg) => write!(f, "Bad request: {}", msg),
            CaptchaError::Unauthorized => write!(f, "Unauthorized"),
            CaptchaError::NotFound(msg) => write!(f, "Not found: {}", msg),
        }
    }
}

impl std::error::Error for CaptchaError {}

impl ResponseError for CaptchaError {
    fn error_response(&self) -> HttpResponse {
        let (status, error_code, message) = match self {
            CaptchaError::InvalidImage(msg) => {
                (actix_web::http::StatusCode::BAD_REQUEST, "invalid_image", msg.clone())
            }
            CaptchaError::ImageTooLarge => {
                (actix_web::http::StatusCode::BAD_REQUEST, "image_too_large", "Image exceeds maximum allowed size".to_string())
            }
            CaptchaError::ModelNotFound(name) => {
                (actix_web::http::StatusCode::NOT_FOUND, "model_not_found", format!("Model '{}' not found", name))
            }
            CaptchaError::ModelLoadError(msg) => {
                (actix_web::http::StatusCode::INTERNAL_SERVER_ERROR, "model_load_error", msg.clone())
            }
            CaptchaError::Timeout => {
                (actix_web::http::StatusCode::REQUEST_TIMEOUT, "timeout", "Processing timeout".to_string())
            }
            CaptchaError::DatabaseError(msg) => {
                (actix_web::http::StatusCode::INTERNAL_SERVER_ERROR, "database_error", msg.clone())
            }
            CaptchaError::ProcessingError(msg) => {
                (actix_web::http::StatusCode::INTERNAL_SERVER_ERROR, "processing_error", msg.clone())
            }
            CaptchaError::BadRequest(msg) => {
                (actix_web::http::StatusCode::BAD_REQUEST, "bad_request", msg.clone())
            }
            CaptchaError::Unauthorized => {
                (actix_web::http::StatusCode::UNAUTHORIZED, "unauthorized", "Unauthorized".to_string())
            }
            CaptchaError::NotFound(msg) => {
                (actix_web::http::StatusCode::NOT_FOUND, "not_found", msg.clone())
            }
        };

        HttpResponse::build(status).json(serde_json::json!({
            "error": error_code,
            "message": message
        }))
    }
}

/// Result type alias for captcha operations
pub type CaptchaResult<T> = Result<T, CaptchaError>;

// Implement From traits for common error types
impl From<sqlx::Error> for CaptchaError {
    fn from(err: sqlx::Error) -> Self {
        CaptchaError::DatabaseError(err.to_string())
    }
}

impl From<image::ImageError> for CaptchaError {
    fn from(err: image::ImageError) -> Self {
        CaptchaError::InvalidImage(err.to_string())
    }
}

impl From<base64::DecodeError> for CaptchaError {
    fn from(err: base64::DecodeError) -> Self {
        CaptchaError::InvalidImage(format!("Invalid base64: {}", err))
    }
}

impl From<std::io::Error> for CaptchaError {
    fn from(err: std::io::Error) -> Self {
        CaptchaError::ProcessingError(err.to_string())
    }
}