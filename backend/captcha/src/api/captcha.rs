//! Captcha Solving Handlers

use actix_web::{web, HttpResponse, HttpRequest};
use base64::{Engine as _, engine::general_purpose::STANDARD as BASE64};
use image::io::Reader as ImageReader;
use sha2::{Sha256, Digest};
use std::io::Cursor;
use std::time::Instant;

use crate::AppState;
use crate::error::{CaptchaError, CaptchaResult};
use crate::models::{
    SolveRequest, SolveResponse, BatchSolveRequest, 
    BatchSolveResponse, BatchResult, PreprocessOptions
};

/// Solve a single captcha
pub async fn solve(
    state: web::Data<AppState>,
    req: HttpRequest,
    body: web::Json<SolveRequest>,
) -> Result<HttpResponse, CaptchaError> {
    let start = Instant::now();

    // Decode base64 image
    let image_data = decode_base64_image(&body.image_base64)?;
    
    // Calculate image hash for logging
    let image_hash = calculate_hash(&image_data);

    // Load image
    let image = load_image(&image_data)?;

    // Get preprocessing options
    let preprocess_opts = body.preprocess.clone();

    // Solve captcha
    let result = state.solver_manager.solve(
        &image,
        body.model.as_deref(),
        preprocess_opts.as_ref(),
    ).await?;

    let processing_time = start.elapsed().as_millis() as u64;

    // Get user info from headers (forwarded by gateway)
    let user_id = req.headers()
        .get("X-User-ID")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.parse::<u64>().ok());

    let request_ip = req.headers()
        .get("X-Forwarded-For")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.to_string())
        .or_else(|| req.peer_addr().map(|a| a.ip().to_string()));

    // Log the request
    let model_id = state.db.get_model_by_name(&result.solver_name).await?
        .map(|m| m.id);

    state.db.create_log(
        user_id,
        model_id,
        &image_hash,
        Some(&result.text),
        Some(result.confidence as f64),
        processing_time as u32,
        request_ip.as_deref(),
    ).await?;

    Ok(HttpResponse::Ok().json(SolveResponse {
        text: result.text,
        confidence: result.confidence,
        model: result.solver_name,
        processing_time_ms: processing_time,
    }))
}

/// Solve multiple captchas in batch
pub async fn solve_batch(
    state: web::Data<AppState>,
    req: HttpRequest,
    body: web::Json<BatchSolveRequest>,
) -> Result<HttpResponse, CaptchaError> {
    let start = Instant::now();
    let batch_size = state.config.processing.batch_size;

    // Limit batch size
    if body.images.len() > batch_size {
        return Err(CaptchaError::BadRequest(
            format!("Batch size exceeds limit of {}", batch_size)
        ));
    }

    let mut results: Vec<BatchResult> = Vec::with_capacity(body.images.len());

    for (index, solve_req) in body.images.iter().enumerate() {
        let result = process_single_image(&state, solve_req).await;

        match result {
            Ok(response) => {
                results.push(BatchResult {
                    index,
                    success: true,
                    result: Some(response),
                    error: None,
                });
            }
            Err(e) => {
                results.push(BatchResult {
                    index,
                    success: false,
                    result: None,
                    error: Some(e.to_string()),
                });
            }
        }
    }

    let total_time = start.elapsed().as_millis() as u64;

    Ok(HttpResponse::Ok().json(BatchSolveResponse {
        results,
        total_time_ms: total_time,
    }))
}

/// Process a single image in batch
async fn process_single_image(
    state: &web::Data<AppState>,
    request: &SolveRequest,
) -> CaptchaResult<SolveResponse> {
    let start = Instant::now();

    // Decode and load image
    let image_data = decode_base64_image(&request.image_base64)?;
    let image = load_image(&image_data)?;

    // Solve
    let result = state.solver_manager.solve(
        &image,
        request.model.as_deref(),
        request.preprocess.as_ref(),
    ).await?;

    let processing_time = start.elapsed().as_millis() as u64;

    Ok(SolveResponse {
        text: result.text,
        confidence: result.confidence,
        model: result.solver_name,
        processing_time_ms: processing_time,
    })
}

/// Decode base64 image data
fn decode_base64_image(base64_str: &str) -> CaptchaResult<Vec<u8>> {
    // Handle data URL format
    let data = if base64_str.contains(",") {
        base64_str.split(",").last().unwrap_or(base64_str)
    } else {
        base64_str
    };

    BASE64.decode(data)
        .map_err(|e| CaptchaError::InvalidImage(format!("Invalid base64: {}", e)))
}

/// Load image from bytes
fn load_image(data: &[u8]) -> CaptchaResult<image::DynamicImage> {
    ImageReader::new(Cursor::new(data))
        .with_guessed_format()
        .map_err(|e| CaptchaError::InvalidImage(format!("Cannot detect image format: {}", e)))?
        .decode()
        .map_err(|e| CaptchaError::InvalidImage(format!("Cannot decode image: {}", e)))
}

/// Calculate SHA256 hash of data
fn calculate_hash(data: &[u8]) -> String {
    let mut hasher = Sha256::new();
    hasher.update(data);
    hex::encode(hasher.finalize())
}