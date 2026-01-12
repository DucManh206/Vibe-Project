//! Models API Handlers

use actix_web::{web, HttpResponse, HttpRequest};
use serde::{Deserialize, Serialize};

use crate::AppState;
use crate::error::CaptchaError;
use crate::models::CaptchaModel;

/// List all available models
pub async fn list_models(
    state: web::Data<AppState>,
    _req: HttpRequest,
) -> Result<HttpResponse, CaptchaError> {
    let models = state.db.get_all_models().await?;
    
    let response: Vec<ModelResponse> = models.into_iter().map(|m| m.into()).collect();
    
    Ok(HttpResponse::Ok().json(response))
}

/// Upload a new model
pub async fn upload_model(
    state: web::Data<AppState>,
    req: HttpRequest,
    body: web::Json<UploadModelRequest>,
) -> Result<HttpResponse, CaptchaError> {
    // Get user info from headers (forwarded by gateway)
    let user_id = req.headers()
        .get("X-User-ID")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok());

    // Validate request
    if body.name.is_empty() {
        return Err(CaptchaError::BadRequest("Model name is required".to_string()));
    }

    if !["ocr", "cnn", "rnn", "transformer", "ensemble"].contains(&body.model_type.as_str()) {
        return Err(CaptchaError::BadRequest("Invalid model type".to_string()));
    }

    // TODO: Actually save the model file
    // For now, just save metadata to database
    
    let model = state.db.create_model(
        &body.name,
        &body.model_type,
        &body.version.clone().unwrap_or_else(|| "1.0.0".to_string()),
        &format!("{}/{}.onnx", state.config.models.path, body.name),
        0, // file size
        user_id,
        body.description.clone(),
    ).await?;

    Ok(HttpResponse::Created().json(ModelResponse::from(model)))
}

/// Get model by ID
pub async fn get_model(
    state: web::Data<AppState>,
    path: web::Path<u64>,
) -> Result<HttpResponse, CaptchaError> {
    let model_id = path.into_inner();
    
    let model = state.db.get_model_by_id(model_id).await?
        .ok_or(CaptchaError::ModelNotFound(format!("Model {} not found", model_id)))?;

    Ok(HttpResponse::Ok().json(ModelResponse::from(model)))
}

/// Update model
pub async fn update_model(
    state: web::Data<AppState>,
    path: web::Path<u64>,
    body: web::Json<UpdateModelRequest>,
) -> Result<HttpResponse, CaptchaError> {
    let model_id = path.into_inner();
    
    // Check if model exists
    let _existing = state.db.get_model_by_id(model_id).await?
        .ok_or(CaptchaError::ModelNotFound(format!("Model {} not found", model_id)))?;

    // Update model
    state.db.update_model(
        model_id,
        body.is_active,
        body.is_default,
        body.description.clone(),
    ).await?;

    // Fetch updated model
    let model = state.db.get_model_by_id(model_id).await?
        .ok_or(CaptchaError::ModelNotFound(format!("Model {} not found", model_id)))?;

    Ok(HttpResponse::Ok().json(ModelResponse::from(model)))
}

/// Delete model
pub async fn delete_model(
    state: web::Data<AppState>,
    path: web::Path<u64>,
) -> Result<HttpResponse, CaptchaError> {
    let model_id = path.into_inner();
    
    // Check if model exists
    let _existing = state.db.get_model_by_id(model_id).await?
        .ok_or(CaptchaError::ModelNotFound(format!("Model {} not found", model_id)))?;

    // Delete model
    state.db.delete_model(model_id).await?;

    Ok(HttpResponse::Ok().json(serde_json::json!({
        "message": "Model deleted successfully"
    })))
}

/// Set model as default
pub async fn set_default_model(
    state: web::Data<AppState>,
    path: web::Path<u64>,
) -> Result<HttpResponse, CaptchaError> {
    let model_id = path.into_inner();
    
    // Check if model exists
    let existing = state.db.get_model_by_id(model_id).await?
        .ok_or(CaptchaError::ModelNotFound(format!("Model {} not found", model_id)))?;

    if !existing.is_active {
        return Err(CaptchaError::BadRequest("Cannot set inactive model as default".to_string()));
    }

    // Set as default
    state.db.set_default_model(model_id).await?;

    // Fetch updated model
    let model = state.db.get_model_by_id(model_id).await?
        .ok_or(CaptchaError::ModelNotFound(format!("Model {} not found", model_id)))?;

    Ok(HttpResponse::Ok().json(ModelResponse::from(model)))
}

// Request/Response types

#[derive(Debug, Deserialize)]
pub struct UploadModelRequest {
    pub name: String,
    pub model_type: String,
    pub version: Option<String>,
    pub description: Option<String>,
    #[serde(default)]
    pub model_data: Option<String>, // Base64 encoded model file
}

#[derive(Debug, Deserialize)]
pub struct UpdateModelRequest {
    pub is_active: Option<bool>,
    pub is_default: Option<bool>,
    pub description: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct ModelResponse {
    pub id: u64,
    pub name: String,
    #[serde(rename = "type")]
    pub model_type: String,
    pub version: String,
    pub accuracy: Option<f64>,
    pub is_active: bool,
    pub is_default: bool,
    pub description: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

impl From<CaptchaModel> for ModelResponse {
    fn from(model: CaptchaModel) -> Self {
        Self {
            id: model.id,
            name: model.name,
            model_type: model.model_type,
            version: model.version,
            accuracy: model.accuracy,
            is_active: model.is_active,
            is_default: model.is_default,
            description: model.description,
            created_at: model.created_at.to_rfc3339(),
            updated_at: model.updated_at.to_rfc3339(),
        }
    }
}