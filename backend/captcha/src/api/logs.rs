//! Logs API Handlers

use actix_web::{web, HttpResponse, HttpRequest};
use serde::{Deserialize, Serialize};

use crate::AppState;
use crate::error::CaptchaError;

/// Get captcha processing logs
pub async fn get_logs(
    state: web::Data<AppState>,
    req: HttpRequest,
    query: web::Query<LogsQuery>,
) -> Result<HttpResponse, CaptchaError> {
    // Get user info from headers
    let user_id = req.headers()
        .get("X-User-ID")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok());

    let logs = state.db.get_logs(
        user_id,
        query.model_id,
        query.is_correct,
        query.limit.unwrap_or(50),
        query.offset.unwrap_or(0),
    ).await?;

    let total = state.db.count_logs(user_id, query.model_id, query.is_correct).await?;

    let response = LogsResponse {
        logs: logs.into_iter().map(|l| l.into()).collect(),
        total,
        limit: query.limit.unwrap_or(50),
        offset: query.offset.unwrap_or(0),
    };

    Ok(HttpResponse::Ok().json(response))
}

/// Get a single log entry
pub async fn get_log(
    state: web::Data<AppState>,
    path: web::Path<u64>,
) -> Result<HttpResponse, CaptchaError> {
    let log_id = path.into_inner();
    
    let log = state.db.get_log_by_id(log_id).await?
        .ok_or(CaptchaError::BadRequest(format!("Log {} not found", log_id)))?;

    Ok(HttpResponse::Ok().json(LogResponse::from(log)))
}

/// Update log with actual text (for feedback/training)
pub async fn update_log(
    state: web::Data<AppState>,
    path: web::Path<u64>,
    body: web::Json<UpdateLogRequest>,
) -> Result<HttpResponse, CaptchaError> {
    let log_id = path.into_inner();
    
    // Check if log exists
    let existing = state.db.get_log_by_id(log_id).await?
        .ok_or(CaptchaError::BadRequest(format!("Log {} not found", log_id)))?;

    // Calculate if correct
    let is_correct = body.actual_text.as_ref()
        .map(|actual| existing.predicted_text.as_ref() == Some(actual));

    // Update the log
    state.db.update_log(log_id, body.actual_text.clone(), is_correct).await?;

    // Fetch updated log
    let log = state.db.get_log_by_id(log_id).await?
        .ok_or(CaptchaError::BadRequest(format!("Log {} not found", log_id)))?;

    Ok(HttpResponse::Ok().json(LogResponse::from(log)))
}

/// Export logs as CSV
pub async fn export_logs(
    state: web::Data<AppState>,
    req: HttpRequest,
    query: web::Query<ExportQuery>,
) -> Result<HttpResponse, CaptchaError> {
    // Get user info from headers
    let user_id = req.headers()
        .get("X-User-ID")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok());

    let logs = state.db.get_logs(
        user_id,
        query.model_id,
        None,
        query.limit.unwrap_or(1000),
        0,
    ).await?;

    // Generate CSV
    let mut csv = String::from("id,image_hash,predicted_text,actual_text,is_correct,confidence,processing_time_ms,model_id,created_at\n");
    
    for log in logs {
        csv.push_str(&format!(
            "{},{},{},{},{},{},{},{},{}\n",
            log.id,
            log.image_hash,
            log.predicted_text.unwrap_or_default(),
            log.actual_text.unwrap_or_default(),
            log.is_correct.map(|b| b.to_string()).unwrap_or_default(),
            log.confidence.map(|c| c.to_string()).unwrap_or_default(),
            log.processing_time_ms,
            log.model_id.map(|id| id.to_string()).unwrap_or_default(),
            log.created_at.to_rfc3339(),
        ));
    }

    Ok(HttpResponse::Ok()
        .content_type("text/csv")
        .insert_header(("Content-Disposition", "attachment; filename=\"captcha_logs.csv\""))
        .body(csv))
}

// Request/Response types

#[derive(Debug, Deserialize)]
pub struct LogsQuery {
    pub model_id: Option<u64>,
    pub is_correct: Option<bool>,
    pub limit: Option<u32>,
    pub offset: Option<u32>,
}

#[derive(Debug, Deserialize)]
pub struct ExportQuery {
    pub model_id: Option<u64>,
    pub limit: Option<u32>,
}

#[derive(Debug, Deserialize)]
pub struct UpdateLogRequest {
    pub actual_text: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct LogsResponse {
    pub logs: Vec<LogResponse>,
    pub total: u64,
    pub limit: u32,
    pub offset: u32,
}

#[derive(Debug, Serialize)]
pub struct LogResponse {
    pub id: u64,
    pub user_id: Option<u64>,
    pub model_id: Option<u64>,
    pub image_hash: String,
    pub predicted_text: Option<String>,
    pub actual_text: Option<String>,
    pub confidence: Option<f64>,
    pub is_correct: Option<bool>,
    pub processing_time_ms: u32,
    pub created_at: String,
}

impl From<crate::models::CaptchaLog> for LogResponse {
    fn from(log: crate::models::CaptchaLog) -> Self {
        Self {
            id: log.id,
            user_id: log.user_id,
            model_id: log.model_id,
            image_hash: log.image_hash,
            predicted_text: log.predicted_text,
            actual_text: log.actual_text,
            confidence: log.confidence,
            is_correct: log.is_correct,
            processing_time_ms: log.processing_time_ms,
            created_at: log.created_at.to_rfc3339(),
        }
    }
}