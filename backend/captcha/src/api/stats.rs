//! Stats API Handlers

use actix_web::{web, HttpResponse, HttpRequest};
use serde::Serialize;

use crate::AppState;
use crate::error::CaptchaError;

/// Get overall statistics
pub async fn get_stats(
    state: web::Data<AppState>,
    req: HttpRequest,
) -> Result<HttpResponse, CaptchaError> {
    // Get user info from headers (optional - admin sees all, users see their own)
    let user_id = req.headers()
        .get("X-User-ID")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok());

    let stats = state.db.get_stats(user_id).await?;

    Ok(HttpResponse::Ok().json(stats))
}

/// Get stats by model
pub async fn get_model_stats(
    state: web::Data<AppState>,
    _req: HttpRequest,
) -> Result<HttpResponse, CaptchaError> {
    let model_stats = state.db.get_model_stats().await?;
    
    Ok(HttpResponse::Ok().json(model_stats))
}

/// Get stats over time (for charts)
pub async fn get_time_series_stats(
    state: web::Data<AppState>,
    req: HttpRequest,
    query: web::Query<TimeSeriesQuery>,
) -> Result<HttpResponse, CaptchaError> {
    let user_id = req.headers()
        .get("X-User-ID")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok());

    let interval = query.interval.as_deref().unwrap_or("day");
    let days = query.days.unwrap_or(30);

    let time_series = state.db.get_time_series_stats(user_id, interval, days).await?;

    Ok(HttpResponse::Ok().json(time_series))
}

// Query types

#[derive(Debug, serde::Deserialize)]
pub struct TimeSeriesQuery {
    pub interval: Option<String>,  // hour, day, week, month
    pub days: Option<u32>,
}

// Response types

#[derive(Debug, Serialize)]
pub struct StatsResponse {
    pub total_requests: u64,
    pub successful_requests: u64,
    pub failed_requests: u64,
    pub average_processing_time_ms: f64,
    pub accuracy_rate: f64,
    pub models_count: u32,
    pub active_models_count: u32,
}

#[derive(Debug, Serialize)]
pub struct ModelStatsResponse {
    pub model_id: u64,
    pub model_name: String,
    pub total_requests: u64,
    pub correct_predictions: u64,
    pub accuracy: f64,
    pub average_processing_time_ms: f64,
}

#[derive(Debug, Serialize)]
pub struct TimeSeriesPoint {
    pub timestamp: String,
    pub requests: u64,
    pub successful: u64,
    pub accuracy: f64,
    pub avg_processing_time_ms: f64,
}