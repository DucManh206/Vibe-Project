//! Training API Handlers

use actix_web::{web, HttpResponse, HttpRequest};
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};

use crate::AppState;
use crate::error::CaptchaError;

/// Start a new training job
pub async fn start_training(
    state: web::Data<AppState>,
    req: HttpRequest,
    body: web::Json<StartTrainingRequest>,
) -> Result<HttpResponse, CaptchaError> {
    // Get user info from headers
    let user_id = req.headers()
        .get("X-User-ID")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok());

    // Validate request
    if body.name.is_empty() {
        return Err(CaptchaError::BadRequest("Job name is required".to_string()));
    }

    if !["ocr", "cnn", "rnn", "transformer", "ensemble"].contains(&body.model_type.as_str()) {
        return Err(CaptchaError::BadRequest("Invalid model type".to_string()));
    }

    // Create training job in database
    let job = state.db.create_training_job(
        user_id,
        &body.name,
        &body.model_type,
        &serde_json::to_value(&body.config).unwrap_or_default(),
        body.dataset_path.as_deref(),
    ).await?;

    // TODO: Actually start training in background
    // For now, just return the created job
    
    tracing::info!("Training job created: {} ({})", job.id, body.name);

    Ok(HttpResponse::Created().json(TrainingJobResponse::from(job)))
}

/// Get training job status
pub async fn get_training_status(
    state: web::Data<AppState>,
    path: web::Path<u64>,
) -> Result<HttpResponse, CaptchaError> {
    let job_id = path.into_inner();
    
    let job = state.db.get_training_job(job_id).await?
        .ok_or(CaptchaError::BadRequest(format!("Training job {} not found", job_id)))?;

    Ok(HttpResponse::Ok().json(TrainingJobResponse::from(job)))
}

/// List all training jobs
pub async fn list_training_jobs(
    state: web::Data<AppState>,
    req: HttpRequest,
    query: web::Query<ListJobsQuery>,
) -> Result<HttpResponse, CaptchaError> {
    // Get user info from headers
    let user_id = req.headers()
        .get("X-User-ID")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok());

    let jobs = state.db.list_training_jobs(
        user_id,
        query.status.as_deref(),
        query.limit.unwrap_or(20),
        query.offset.unwrap_or(0),
    ).await?;

    let response: Vec<TrainingJobResponse> = jobs.into_iter().map(|j| j.into()).collect();

    Ok(HttpResponse::Ok().json(response))
}

/// Cancel a training job
pub async fn cancel_training(
    state: web::Data<AppState>,
    req: HttpRequest,
    path: web::Path<u64>,
) -> Result<HttpResponse, CaptchaError> {
    let job_id = path.into_inner();
    
    // Get user info from headers
    let user_id = req.headers()
        .get("X-User-ID")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok());

    let job = state.db.get_training_job(job_id).await?
        .ok_or(CaptchaError::BadRequest(format!("Training job {} not found", job_id)))?;

    // Check ownership
    if let Some(uid) = user_id {
        if job.user_id != Some(uid) {
            return Err(CaptchaError::BadRequest("Not authorized to cancel this job".to_string()));
        }
    }

    // Check if job can be cancelled
    if job.status != "pending" && job.status != "running" {
        return Err(CaptchaError::BadRequest(
            format!("Cannot cancel job with status: {}", job.status)
        ));
    }

    // Cancel the job
    state.db.update_training_job_status(job_id, "cancelled", None).await?;

    // TODO: Actually stop the training process if running

    Ok(HttpResponse::Ok().json(serde_json::json!({
        "message": "Training job cancelled"
    })))
}

// Request/Response types

#[derive(Debug, Deserialize)]
pub struct StartTrainingRequest {
    pub name: String,
    pub model_type: String,
    pub config: TrainingConfig,
    pub dataset_path: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct TrainingConfig {
    #[serde(default = "default_epochs")]
    pub epochs: u32,
    #[serde(default = "default_batch_size")]
    pub batch_size: u32,
    #[serde(default = "default_learning_rate")]
    pub learning_rate: f64,
    #[serde(default = "default_validation_split")]
    pub validation_split: f64,
    #[serde(default)]
    pub augmentation: bool,
    #[serde(default)]
    pub early_stopping: bool,
    pub patience: Option<u32>,
}

fn default_epochs() -> u32 { 100 }
fn default_batch_size() -> u32 { 32 }
fn default_learning_rate() -> f64 { 0.001 }
fn default_validation_split() -> f64 { 0.2 }

#[derive(Debug, Deserialize)]
pub struct ListJobsQuery {
    pub status: Option<String>,
    pub limit: Option<u32>,
    pub offset: Option<u32>,
}

#[derive(Debug, Serialize)]
pub struct TrainingJobResponse {
    pub id: u64,
    pub user_id: Option<u64>,
    pub name: String,
    pub status: String,
    pub model_type: String,
    pub config: serde_json::Value,
    pub progress: f64,
    pub current_epoch: Option<u32>,
    pub total_epochs: Option<u32>,
    pub results: Option<serde_json::Value>,
    pub error_message: Option<String>,
    pub started_at: Option<String>,
    pub completed_at: Option<String>,
    pub created_at: String,
}

impl From<crate::models::TrainingJob> for TrainingJobResponse {
    fn from(job: crate::models::TrainingJob) -> Self {
        Self {
            id: job.id,
            user_id: job.user_id,
            name: job.name,
            status: job.status,
            model_type: job.model_type,
            config: job.config,
            progress: job.progress,
            current_epoch: job.current_epoch,
            total_epochs: job.total_epochs,
            results: job.results,
            error_message: job.error_message,
            started_at: job.started_at.map(|t| t.to_rfc3339()),
            completed_at: job.completed_at.map(|t| t.to_rfc3339()),
            created_at: job.created_at.to_rfc3339(),
        }
    }
}