//! Captcha Solver Module
//! 
//! This module contains different captcha solving strategies:
//! - OCR: Tesseract-based text recognition
//! - CNN: Deep learning based recognition
//! - Ensemble: Combines multiple models for better accuracy

pub mod ocr;
pub mod cnn;
pub mod preprocessor;

use std::sync::Arc;
use tokio::sync::RwLock;
use std::collections::HashMap;
use image::DynamicImage;

use crate::config::ModelsSettings;
use crate::error::{CaptchaError, CaptchaResult};
use crate::models::{SolveResponse, PreprocessOptions, CaptchaModel};

/// Trait for captcha solvers
#[async_trait::async_trait]
pub trait CaptchaSolver: Send + Sync {
    /// Solve a captcha image
    async fn solve(&self, image: &DynamicImage, options: Option<&PreprocessOptions>) -> CaptchaResult<SolveResult>;
    
    /// Get solver name
    fn name(&self) -> &str;
    
    /// Check if solver is ready
    fn is_ready(&self) -> bool;
}

/// Result from a solver
#[derive(Debug, Clone)]
pub struct SolveResult {
    pub text: String,
    pub confidence: f32,
    pub solver_name: String,
}

/// Manages multiple captcha solvers
pub struct SolverManager {
    solvers: HashMap<String, Arc<dyn CaptchaSolver>>,
    default_solver: String,
    models_path: String,
}

impl SolverManager {
    /// Create a new solver manager
    pub async fn new(config: &ModelsSettings) -> CaptchaResult<Self> {
        let mut solvers: HashMap<String, Arc<dyn CaptchaSolver>> = HashMap::new();

        // Initialize OCR solver if enabled
        if config.ocr_enabled {
            match ocr::OcrSolver::new(&config.path).await {
                Ok(solver) => {
                    solvers.insert("ocr".to_string(), Arc::new(solver));
                    tracing::info!("OCR solver initialized");
                }
                Err(e) => {
                    tracing::warn!("Failed to initialize OCR solver: {}", e);
                }
            }
        }

        // Initialize CNN solver if enabled
        if config.cnn_enabled {
            match cnn::CnnSolver::new(&config.path).await {
                Ok(solver) => {
                    solvers.insert("cnn".to_string(), Arc::new(solver));
                    tracing::info!("CNN solver initialized");
                }
                Err(e) => {
                    tracing::warn!("Failed to initialize CNN solver: {}", e);
                }
            }
        }

        // Determine default solver
        let default_solver = if solvers.contains_key("cnn") {
            "cnn".to_string()
        } else if solvers.contains_key("ocr") {
            "ocr".to_string()
        } else {
            return Err(CaptchaError::ModelLoadError(
                "No solvers available".to_string()
            ));
        };

        Ok(Self {
            solvers,
            default_solver,
            models_path: config.path.clone(),
        })
    }

    /// Get the number of loaded models
    pub fn model_count(&self) -> usize {
        self.solvers.len()
    }

    /// Solve a captcha using the specified or default solver
    pub async fn solve(
        &self,
        image: &DynamicImage,
        model_name: Option<&str>,
        options: Option<&PreprocessOptions>,
    ) -> CaptchaResult<SolveResult> {
        let solver_name = model_name.unwrap_or(&self.default_solver);

        let solver = self.solvers.get(solver_name)
            .ok_or_else(|| CaptchaError::ModelNotFound(solver_name.to_string()))?;

        if !solver.is_ready() {
            return Err(CaptchaError::ModelLoadError(
                format!("Solver {} is not ready", solver_name)
            ));
        }

        solver.solve(image, options).await
    }

    /// Solve using all available solvers and return the best result
    pub async fn solve_ensemble(
        &self,
        image: &DynamicImage,
        options: Option<&PreprocessOptions>,
    ) -> CaptchaResult<SolveResult> {
        let mut results: Vec<SolveResult> = Vec::new();

        for (name, solver) in &self.solvers {
            if solver.is_ready() {
                match solver.solve(image, options).await {
                    Ok(result) => results.push(result),
                    Err(e) => {
                        tracing::warn!("Solver {} failed: {}", name, e);
                    }
                }
            }
        }

        if results.is_empty() {
            return Err(CaptchaError::ProcessingError(
                "All solvers failed".to_string()
            ));
        }

        // Return result with highest confidence
        results.sort_by(|a, b| b.confidence.partial_cmp(&a.confidence).unwrap());
        
        Ok(results.remove(0))
    }

    /// Get list of available solvers
    pub fn available_solvers(&self) -> Vec<String> {
        self.solvers.keys().cloned().collect()
    }

    /// Load a custom model
    pub async fn load_model(&mut self, model: &CaptchaModel) -> CaptchaResult<()> {
        // Implementation depends on model type
        tracing::info!("Loading model: {} ({})", model.name, model.model_type);
        
        // TODO: Implement custom model loading
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_solver_manager_creation() {
        let config = ModelsSettings {
            path: "/tmp/models".to_string(),
            default_model: "ocr".to_string(),
            ocr_enabled: true,
            cnn_enabled: false,
        };

        // This will likely fail without actual tesseract installed
        // Just testing the structure
        let _ = SolverManager::new(&config).await;
    }
}