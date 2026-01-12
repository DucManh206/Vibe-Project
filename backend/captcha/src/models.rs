//! Data Models
//!
//! This module contains all data structures used throughout the application.

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

// =============================================================================
// Request/Response Models for API
// =============================================================================

/// Request for solving a single captcha
#[derive(Debug, Deserialize)]
pub struct SolveRequest {
    pub image_base64: String,
    pub model: Option<String>,
    pub preprocess: Option<PreprocessOptions>,
}

/// Response from solving a captcha
#[derive(Debug, Serialize)]
pub struct SolveResponse {
    pub text: String,
    pub confidence: f32,
    pub model: String,
    pub processing_time_ms: u64,
}

/// Request for batch solving
#[derive(Debug, Deserialize)]
pub struct BatchSolveRequest {
    pub images: Vec<SolveRequest>,
}

/// Response from batch solving
#[derive(Debug, Serialize)]
pub struct BatchSolveResponse {
    pub results: Vec<BatchResult>,
    pub total_time_ms: u64,
}

/// Result for a single image in batch
#[derive(Debug, Serialize)]
pub struct BatchResult {
    pub index: usize,
    pub success: bool,
    pub result: Option<SolveResponse>,
    pub error: Option<String>,
}

/// Image preprocessing options
#[derive(Debug, Clone, Deserialize, Serialize, Default)]
pub struct PreprocessOptions {
    pub grayscale: Option<bool>,
    pub threshold: Option<u8>,
    pub denoise: Option<bool>,
    pub resize_width: Option<u32>,
    pub resize_height: Option<u32>,
}

// =============================================================================
// Database Models
// =============================================================================

/// Captcha model stored in database
#[derive(Debug, Clone, Serialize)]
pub struct CaptchaModel {
    pub id: u64,
    pub name: String,
    pub model_type: String,
    pub version: String,
    pub file_path: String,
    pub file_size_bytes: u64,
    pub accuracy: Option<f64>,
    pub is_active: bool,
    pub is_default: bool,
    pub metadata: Option<serde_json::Value>,
    pub description: Option<String>,
    pub created_by: Option<u64>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// Captcha processing log entry
#[derive(Debug, Clone, Serialize)]
pub struct CaptchaLog {
    pub id: u64,
    pub user_id: Option<u64>,
    pub model_id: Option<u64>,
    pub image_hash: String,
    pub image_base64: Option<String>,
    pub predicted_text: Option<String>,
    pub actual_text: Option<String>,
    pub confidence: Option<f64>,
    pub is_correct: Option<bool>,
    pub processing_time_ms: u32,
    pub request_ip: Option<String>,
    pub user_agent: Option<String>,
    pub error_message: Option<String>,
    pub created_at: DateTime<Utc>,
}

/// Training job
#[derive(Debug, Clone, Serialize)]
pub struct TrainingJob {
    pub id: u64,
    pub user_id: Option<u64>,
    pub name: String,
    pub status: String,
    pub model_type: String,
    pub config: serde_json::Value,
    pub dataset_path: Option<String>,
    pub dataset_size: Option<u32>,
    pub progress: f64,
    pub current_epoch: Option<u32>,
    pub total_epochs: Option<u32>,
    pub results: Option<serde_json::Value>,
    pub output_model_id: Option<u64>,
    pub error_message: Option<String>,
    pub started_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// API Key
#[derive(Debug, Clone, Serialize)]
pub struct ApiKey {
    pub id: u64,
    pub user_id: u64,
    pub name: String,
    pub key_prefix: String,
    pub key_hash: String,
    pub scopes: Option<serde_json::Value>,
    pub rate_limit: u32,
    pub total_requests: u64,
    pub last_used_at: Option<DateTime<Utc>>,
    pub is_active: bool,
    pub expires_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// User
#[derive(Debug, Clone, Serialize)]
pub struct User {
    pub id: u64,
    pub email: String,
    pub role: String,
    pub is_active: bool,
    pub created_at: DateTime<Utc>,
}

// =============================================================================
// Statistics Models
// =============================================================================

/// Overall statistics
#[derive(Debug, Serialize)]
pub struct Stats {
    pub total_requests: u64,
    pub successful_requests: u64,
    pub failed_requests: u64,
    pub average_processing_time_ms: f64,
    pub accuracy_rate: f64,
    pub models_count: u32,
    pub active_models_count: u32,
}

/// Statistics by model
#[derive(Debug, Serialize)]
pub struct ModelStats {
    pub model_id: u64,
    pub model_name: String,
    pub total_requests: u64,
    pub correct_predictions: u64,
    pub accuracy: f64,
    pub average_processing_time_ms: f64,
}

/// Time series data point
#[derive(Debug, Serialize)]
pub struct TimeSeriesPoint {
    pub timestamp: String,
    pub requests: u64,
    pub successful: u64,
    pub accuracy: f64,
    pub avg_processing_time_ms: f64,
}

// =============================================================================
// Training Models
// =============================================================================

/// Training configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrainingConfig {
    pub epochs: u32,
    pub batch_size: u32,
    pub learning_rate: f64,
    pub validation_split: f64,
    pub augmentation: bool,
    pub early_stopping: bool,
    pub patience: Option<u32>,
}

impl Default for TrainingConfig {
    fn default() -> Self {
        Self {
            epochs: 100,
            batch_size: 32,
            learning_rate: 0.001,
            validation_split: 0.2,
            augmentation: false,
            early_stopping: true,
            patience: Some(10),
        }
    }
}

/// Training results
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrainingResults {
    pub final_accuracy: f64,
    pub final_loss: f64,
    pub validation_accuracy: f64,
    pub validation_loss: f64,
    pub epochs_trained: u32,
    pub training_time_seconds: u64,
    pub model_path: String,
}