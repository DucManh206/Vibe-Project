//! Database module for Captcha Service

use sqlx::{mysql::MySqlPoolOptions, MySql, Pool};
use crate::config::DatabaseSettings;
use crate::error::{CaptchaError, CaptchaResult};
use crate::models::{CaptchaLog, CaptchaModel, TrainingJob, TrainingStatus, ModelType};
use chrono::{DateTime, Utc};

/// Database wrapper
pub struct Database {
    pool: Pool<MySql>,
}

impl Database {
    /// Create a new database connection
    pub async fn new(config: &DatabaseSettings) -> CaptchaResult<Self> {
        let pool = MySqlPoolOptions::new()
            .max_connections(config.max_connections)
            .connect(&config.connection_url())
            .await
            .map_err(|e| CaptchaError::DatabaseError(e.to_string()))?;

        Ok(Self { pool })
    }

    /// Get the connection pool
    pub fn pool(&self) -> &Pool<MySql> {
        &self.pool
    }

    // ==================== Model Operations ====================

    /// Get all active models
    pub async fn get_active_models(&self) -> CaptchaResult<Vec<CaptchaModel>> {
        let models = sqlx::query_as!(
            CaptchaModel,
            r#"
            SELECT 
                id, name, 
                type as "model_type: ModelType",
                version, file_path, file_size_bytes,
                accuracy, is_active, is_default,
                metadata as "metadata: serde_json::Value",
                description, created_by, created_at, updated_at
            FROM captcha_models 
            WHERE is_active = true
            ORDER BY is_default DESC, accuracy DESC
            "#
        )
        .fetch_all(&self.pool)
        .await?;

        Ok(models)
    }

    /// Get default model
    pub async fn get_default_model(&self) -> CaptchaResult<Option<CaptchaModel>> {
        let model = sqlx::query_as!(
            CaptchaModel,
            r#"
            SELECT 
                id, name, 
                type as "model_type: ModelType",
                version, file_path, file_size_bytes,
                accuracy, is_active, is_default,
                metadata as "metadata: serde_json::Value",
                description, created_by, created_at, updated_at
            FROM captcha_models 
            WHERE is_default = true AND is_active = true
            LIMIT 1
            "#
        )
        .fetch_optional(&self.pool)
        .await?;

        Ok(model)
    }

    /// Get model by name
    pub async fn get_model_by_name(&self, name: &str) -> CaptchaResult<Option<CaptchaModel>> {
        let model = sqlx::query_as!(
            CaptchaModel,
            r#"
            SELECT 
                id, name, 
                type as "model_type: ModelType",
                version, file_path, file_size_bytes,
                accuracy, is_active, is_default,
                metadata as "metadata: serde_json::Value",
                description, created_by, created_at, updated_at
            FROM captcha_models 
            WHERE name = ? AND is_active = true
            "#,
            name
        )
        .fetch_optional(&self.pool)
        .await?;

        Ok(model)
    }

    /// Create a new model
    pub async fn create_model(
        &self,
        name: &str,
        model_type: &ModelType,
        version: &str,
        file_path: &str,
        file_size: u64,
        description: Option<&str>,
        created_by: Option<u64>,
    ) -> CaptchaResult<u64> {
        let result = sqlx::query!(
            r#"
            INSERT INTO captcha_models 
                (name, type, version, file_path, file_size_bytes, description, created_by, is_active, is_default)
            VALUES (?, ?, ?, ?, ?, ?, ?, true, false)
            "#,
            name,
            model_type.to_string(),
            version,
            file_path,
            file_size,
            description,
            created_by
        )
        .execute(&self.pool)
        .await?;

        Ok(result.last_insert_id())
    }

    // ==================== Log Operations ====================

    /// Create a log entry
    pub async fn create_log(
        &self,
        user_id: Option<u64>,
        model_id: Option<u64>,
        image_hash: &str,
        predicted_text: Option<&str>,
        confidence: Option<f64>,
        processing_time_ms: u32,
        request_ip: Option<&str>,
    ) -> CaptchaResult<u64> {
        let result = sqlx::query!(
            r#"
            INSERT INTO captcha_logs 
                (user_id, model_id, image_hash, predicted_text, confidence, processing_time_ms, request_ip)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            "#,
            user_id,
            model_id,
            image_hash,
            predicted_text,
            confidence,
            processing_time_ms,
            request_ip
        )
        .execute(&self.pool)
        .await?;

        Ok(result.last_insert_id())
    }

    /// Get logs with pagination
    pub async fn get_logs(
        &self,
        page: u32,
        limit: u32,
        model_id: Option<u64>,
        is_correct: Option<bool>,
    ) -> CaptchaResult<(Vec<CaptchaLog>, u64)> {
        let offset = (page - 1) * limit;

        // Build dynamic query
        let mut conditions = vec!["1=1".to_string()];
        
        if let Some(mid) = model_id {
            conditions.push(format!("model_id = {}", mid));
        }
        
        if let Some(correct) = is_correct {
            conditions.push(format!("is_correct = {}", correct));
        }

        let where_clause = conditions.join(" AND ");

        // Get total count
        let count_query = format!(
            "SELECT COUNT(*) as count FROM captcha_logs WHERE {}",
            where_clause
        );
        let count_row: (i64,) = sqlx::query_as(&count_query)
            .fetch_one(&self.pool)
            .await?;
        let total = count_row.0 as u64;

        // Get logs
        let logs_query = format!(
            r#"
            SELECT id, user_id, model_id, image_hash, predicted_text, 
                   actual_text, confidence, is_correct, processing_time_ms, 
                   request_ip, created_at
            FROM captcha_logs 
            WHERE {}
            ORDER BY created_at DESC
            LIMIT {} OFFSET {}
            "#,
            where_clause, limit, offset
        );

        let logs = sqlx::query_as::<_, CaptchaLog>(&logs_query)
            .fetch_all(&self.pool)
            .await?;

        Ok((logs, total))
    }

    // ==================== Training Operations ====================

    /// Create a training job
    pub async fn create_training_job(
        &self,
        user_id: Option<u64>,
        name: &str,
        model_type: &ModelType,
        config: &serde_json::Value,
        dataset_path: Option<&str>,
    ) -> CaptchaResult<u64> {
        let result = sqlx::query!(
            r#"
            INSERT INTO training_jobs 
                (user_id, name, status, model_type, config, dataset_path, progress)
            VALUES (?, ?, 'pending', ?, ?, ?, 0)
            "#,
            user_id,
            name,
            model_type.to_string(),
            config,
            dataset_path
        )
        .execute(&self.pool)
        .await?;

        Ok(result.last_insert_id())
    }

    /// Get training job by ID
    pub async fn get_training_job(&self, job_id: u64) -> CaptchaResult<Option<TrainingJob>> {
        let job = sqlx::query_as!(
            TrainingJob,
            r#"
            SELECT 
                id, user_id, name,
                status as "status: TrainingStatus",
                model_type as "model_type: ModelType",
                config as "config: serde_json::Value",
                dataset_path, dataset_size, progress,
                current_epoch, total_epochs,
                results as "results: serde_json::Value",
                output_model_id, error_message,
                started_at, completed_at, created_at, updated_at
            FROM training_jobs 
            WHERE id = ?
            "#,
            job_id
        )
        .fetch_optional(&self.pool)
        .await?;

        Ok(job)
    }

    /// Update training job status
    pub async fn update_training_status(
        &self,
        job_id: u64,
        status: &TrainingStatus,
        progress: f64,
        current_epoch: Option<u32>,
        error_message: Option<&str>,
    ) -> CaptchaResult<()> {
        sqlx::query!(
            r#"
            UPDATE training_jobs 
            SET status = ?, progress = ?, current_epoch = ?, error_message = ?, updated_at = NOW()
            WHERE id = ?
            "#,
            status.to_string(),
            progress,
            current_epoch,
            error_message,
            job_id
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    // ==================== Statistics ====================

    /// Get statistics
    pub async fn get_stats(&self) -> CaptchaResult<(u64, u64, u64, f64, f64, u64, u64)> {
        // Total requests
        let total: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM captcha_logs")
            .fetch_one(&self.pool)
            .await?;

        // Successful (non-null predictions)
        let successful: (i64,) = sqlx::query_as(
            "SELECT COUNT(*) FROM captcha_logs WHERE predicted_text IS NOT NULL"
        )
        .fetch_one(&self.pool)
        .await?;

        // Failed
        let failed = total.0 - successful.0;

        // Average processing time
        let avg_time: (Option<f64>,) = sqlx::query_as(
            "SELECT AVG(processing_time_ms) FROM captcha_logs"
        )
        .fetch_one(&self.pool)
        .await?;

        // Accuracy rate
        let accuracy: (Option<f64>,) = sqlx::query_as(
            "SELECT AVG(CASE WHEN is_correct = true THEN 1.0 ELSE 0.0 END) FROM captcha_logs WHERE is_correct IS NOT NULL"
        )
        .fetch_one(&self.pool)
        .await?;

        // Models count
        let models: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM captcha_models")
            .fetch_one(&self.pool)
            .await?;

        // Active models
        let active: (i64,) = sqlx::query_as(
            "SELECT COUNT(*) FROM captcha_models WHERE is_active = true"
        )
        .fetch_one(&self.pool)
        .await?;

        Ok((
            total.0 as u64,
            successful.0 as u64,
            failed as u64,
            avg_time.0.unwrap_or(0.0),
            accuracy.0.unwrap_or(0.0),
            models.0 as u64,
            active.0 as u64,
        ))
    }
}

// Implement FromRow for CaptchaLog
impl<'r> sqlx::FromRow<'r, sqlx::mysql::MySqlRow> for CaptchaLog {
    fn from_row(row: &'r sqlx::mysql::MySqlRow) -> Result<Self, sqlx::Error> {
        use sqlx::Row;
        
        Ok(CaptchaLog {
            id: row.try_get("id")?,
            user_id: row.try_get("user_id")?,
            model_id: row.try_get("model_id")?,
            image_hash: row.try_get("image_hash")?,
            predicted_text: row.try_get("predicted_text")?,
            actual_text: row.try_get("actual_text")?,
            confidence: row.try_get("confidence")?,
            is_correct: row.try_get("is_correct")?,
            processing_time_ms: row.try_get("processing_time_ms")?,
            request_ip: row.try_get("request_ip")?,
            created_at: row.try_get("created_at")?,
        })
    }
}

impl std::fmt::Display for TrainingStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            TrainingStatus::Pending => write!(f, "pending"),
            TrainingStatus::Running => write!(f, "running"),
            TrainingStatus::Completed => write!(f, "completed"),
            TrainingStatus::Failed => write!(f, "failed"),
            TrainingStatus::Cancelled => write!(f, "cancelled"),
        }
    }
}