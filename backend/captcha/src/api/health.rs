//! Health Check Handler

use actix_web::{HttpResponse, web};
use serde::Serialize;
use chrono::Utc;

use crate::AppState;

/// Health check response
#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub service: String,
    pub version: String,
    pub time: String,
    pub checks: HealthChecks,
}

#[derive(Debug, Serialize)]
pub struct HealthChecks {
    pub database: HealthStatus,
    pub solvers: HealthStatus,
}

#[derive(Debug, Serialize)]
pub struct HealthStatus {
    pub status: String,
    pub message: Option<String>,
}

/// Health check endpoint
pub async fn health_check(
    state: web::Data<AppState>,
) -> HttpResponse {
    let db_status = match state.db.ping().await {
        Ok(_) => HealthStatus {
            status: "healthy".to_string(),
            message: None,
        },
        Err(e) => HealthStatus {
            status: "unhealthy".to_string(),
            message: Some(e.to_string()),
        },
    };

    let solver_count = state.solver_manager.model_count();
    let solver_status = if solver_count > 0 {
        HealthStatus {
            status: "healthy".to_string(),
            message: Some(format!("{} models loaded", solver_count)),
        }
    } else {
        HealthStatus {
            status: "degraded".to_string(),
            message: Some("No models loaded".to_string()),
        }
    };

    let overall_status = if db_status.status == "healthy" && solver_status.status != "unhealthy" {
        "healthy"
    } else if db_status.status == "unhealthy" {
        "unhealthy"
    } else {
        "degraded"
    };

    let response = HealthResponse {
        status: overall_status.to_string(),
        service: "captcha".to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
        time: Utc::now().to_rfc3339(),
        checks: HealthChecks {
            database: db_status,
            solvers: solver_status,
        },
    };

    if overall_status == "healthy" {
        HttpResponse::Ok().json(response)
    } else if overall_status == "degraded" {
        HttpResponse::Ok().json(response)
    } else {
        HttpResponse::ServiceUnavailable().json(response)
    }
}

/// Readiness check (for Kubernetes)
pub async fn ready_check(
    state: web::Data<AppState>,
) -> HttpResponse {
    // Check if all critical components are ready
    let db_ready = state.db.ping().await.is_ok();
    let solvers_ready = state.solver_manager.model_count() > 0;

    if db_ready && solvers_ready {
        HttpResponse::Ok().json(serde_json::json!({
            "ready": true
        }))
    } else {
        HttpResponse::ServiceUnavailable().json(serde_json::json!({
            "ready": false,
            "database": db_ready,
            "solvers": solvers_ready
        }))
    }
}

/// Liveness check (for Kubernetes)
pub async fn live_check() -> HttpResponse {
    HttpResponse::Ok().json(serde_json::json!({
        "alive": true
    }))
}