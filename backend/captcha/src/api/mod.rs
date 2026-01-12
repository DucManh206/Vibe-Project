//! API Module
//!
//! This module contains all HTTP API handlers for the Captcha Service.

pub mod captcha;
pub mod health;
pub mod logs;
pub mod models;
pub mod stats;
pub mod training;

// Re-export handlers for convenience
pub use captcha::{solve, solve_batch};
pub use health::health_check;
pub use logs::{get_logs, get_log, update_log, export_logs};
pub use models::{list_models, upload_model, get_model, update_model, delete_model, set_default_model};
pub use stats::{get_stats, get_model_stats, get_time_series_stats};
pub use training::{start_training, get_training_status, list_training_jobs, cancel_training};