//! Configuration module for Captcha Service

use serde::Deserialize;
use config::{Config, ConfigError, Environment, File};

/// Main settings structure
#[derive(Debug, Clone, Deserialize)]
pub struct Settings {
    pub server: ServerSettings,
    pub database: DatabaseSettings,
    pub models: ModelsSettings,
    pub processing: ProcessingSettings,
}

/// Server configuration
#[derive(Debug, Clone, Deserialize)]
pub struct ServerSettings {
    pub port: u16,
    pub host: String,
}

/// Database configuration
#[derive(Debug, Clone, Deserialize)]
pub struct DatabaseSettings {
    pub host: String,
    pub port: u16,
    pub name: String,
    pub user: String,
    pub password: String,
    pub max_connections: u32,
}

/// Models configuration
#[derive(Debug, Clone, Deserialize)]
pub struct ModelsSettings {
    pub path: String,
    pub default_model: String,
    pub ocr_enabled: bool,
    pub cnn_enabled: bool,
}

/// Processing configuration
#[derive(Debug, Clone, Deserialize)]
pub struct ProcessingSettings {
    pub max_image_size_mb: usize,
    pub timeout_seconds: u64,
    pub batch_size: usize,
}

impl Settings {
    /// Load settings from environment variables and config files
    pub fn new() -> Result<Self, ConfigError> {
        let run_mode = std::env::var("RUN_MODE").unwrap_or_else(|_| "development".into());

        let s = Config::builder()
            // Start with default values
            .set_default("server.port", 8082)?
            .set_default("server.host", "0.0.0.0")?
            .set_default("database.host", "localhost")?
            .set_default("database.port", 3306)?
            .set_default("database.name", "captcha_platform")?
            .set_default("database.user", "captcha_user")?
            .set_default("database.password", "")?
            .set_default("database.max_connections", 10)?
            .set_default("models.path", "/app/models")?
            .set_default("models.default_model", "tesseract-default")?
            .set_default("models.ocr_enabled", true)?
            .set_default("models.cnn_enabled", true)?
            .set_default("processing.max_image_size_mb", 10)?
            .set_default("processing.timeout_seconds", 30)?
            .set_default("processing.batch_size", 10)?
            // Load config file if exists
            .add_source(File::with_name("config/default").required(false))
            .add_source(File::with_name(&format!("config/{}", run_mode)).required(false))
            // Add in settings from environment variables
            // Format: CAPTCHA_SERVER__PORT=8082
            .add_source(
                Environment::with_prefix("CAPTCHA")
                    .separator("__")
                    .try_parsing(true),
            )
            // Also support simpler env vars
            .add_source(
                Environment::default()
                    .prefix("")
                    .separator("_")
                    .try_parsing(true)
                    .with_list_parse_key("CAPTCHA_SERVICE_PORT")
                    .with_list_parse_key("DB_HOST")
                    .with_list_parse_key("DB_PORT")
                    .with_list_parse_key("DB_NAME")
                    .with_list_parse_key("DB_USER")
                    .with_list_parse_key("DB_PASSWORD"),
            )
            .build()?;

        // Handle direct environment variable overrides
        let mut settings: Settings = s.try_deserialize()?;

        // Override with direct env vars if present
        if let Ok(port) = std::env::var("CAPTCHA_SERVICE_PORT") {
            if let Ok(p) = port.parse() {
                settings.server.port = p;
            }
        }
        if let Ok(host) = std::env::var("DB_HOST") {
            settings.database.host = host;
        }
        if let Ok(port) = std::env::var("DB_PORT") {
            if let Ok(p) = port.parse() {
                settings.database.port = p;
            }
        }
        if let Ok(name) = std::env::var("DB_NAME") {
            settings.database.name = name;
        }
        if let Ok(user) = std::env::var("DB_USER") {
            settings.database.user = user;
        }
        if let Ok(password) = std::env::var("DB_PASSWORD") {
            settings.database.password = password;
        }
        if let Ok(path) = std::env::var("MODELS_PATH") {
            settings.models.path = path;
        }

        Ok(settings)
    }
}

impl DatabaseSettings {
    /// Get database connection URL
    pub fn connection_url(&self) -> String {
        format!(
            "mysql://{}:{}@{}:{}/{}",
            self.user, self.password, self.host, self.port, self.name
        )
    }
}