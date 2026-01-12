//! Captcha Service - Main Entry Point
//! 
//! This service handles captcha solving using multiple approaches:
//! - OCR (Tesseract-based)
//! - CNN (Deep Learning models)
//! - Ensemble (combining multiple models)

mod api;
mod config;
mod models;
mod solvers;
mod error;
mod db;

use actix_cors::Cors;
use actix_web::{web, App, HttpServer, middleware};
use tracing::{info, Level};
use tracing_subscriber::{EnvFilter, fmt, prelude::*};

use crate::config::Settings;
use crate::db::Database;
use crate::solvers::SolverManager;

/// Application state shared across handlers
pub struct AppState {
    pub db: Database,
    pub solver_manager: SolverManager,
    pub config: Settings,
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Load environment variables
    dotenv::dotenv().ok();

    // Initialize tracing
    let filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new("info"));
    
    tracing_subscriber::registry()
        .with(fmt::layer().json())
        .with(filter)
        .init();

    info!("Starting Captcha Service...");

    // Load configuration
    let config = Settings::new().expect("Failed to load configuration");
    let port = config.server.port;

    // Initialize database connection
    let db = Database::new(&config.database)
        .await
        .expect("Failed to connect to database");

    info!("Connected to database");

    // Initialize solver manager
    let solver_manager = SolverManager::new(&config.models)
        .await
        .expect("Failed to initialize solver manager");

    info!("Solver manager initialized with {} models", solver_manager.model_count());

    // Create shared application state
    let app_state = web::Data::new(AppState {
        db,
        solver_manager,
        config: config.clone(),
    });

    info!("Starting HTTP server on port {}", port);

    // Start HTTP server
    HttpServer::new(move || {
        // Configure CORS
        let cors = Cors::default()
            .allow_any_origin()
            .allow_any_method()
            .allow_any_header()
            .max_age(3600);

        App::new()
            .app_data(app_state.clone())
            .wrap(cors)
            .wrap(middleware::Logger::default())
            .wrap(middleware::Compress::default())
            // Health check
            .route("/health", web::get().to(api::health::health_check))
            // API routes
            .service(
                web::scope("/captcha")
                    .route("/solve", web::post().to(api::captcha::solve))
                    .route("/solve/batch", web::post().to(api::captcha::solve_batch))
                    .route("/models", web::get().to(api::models::list_models))
                    .route("/models/upload", web::post().to(api::models::upload_model))
                    .route("/train", web::post().to(api::training::start_training))
                    .route("/train/{job_id}", web::get().to(api::training::get_training_status))
                    .route("/logs", web::get().to(api::logs::get_logs))
                    .route("/stats", web::get().to(api::stats::get_stats))
            )
    })
    .bind(format!("0.0.0.0:{}", port))?
    .run()
    .await
}