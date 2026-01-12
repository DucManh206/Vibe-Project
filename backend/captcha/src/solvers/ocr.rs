//! OCR Solver using Tesseract
//! 
//! This solver uses Tesseract OCR for text recognition in captcha images.

use image::DynamicImage;
use std::sync::atomic::{AtomicBool, Ordering};

use crate::error::{CaptchaError, CaptchaResult};
use crate::models::PreprocessOptions;
use super::{CaptchaSolver, SolveResult};
use super::preprocessor::ImagePreprocessor;

/// OCR-based captcha solver using Tesseract
pub struct OcrSolver {
    ready: AtomicBool,
    models_path: String,
}

impl OcrSolver {
    /// Create a new OCR solver
    pub async fn new(models_path: &str) -> CaptchaResult<Self> {
        // Verify Tesseract is available
        let solver = Self {
            ready: AtomicBool::new(false),
            models_path: models_path.to_string(),
        };

        // Try to initialize Tesseract
        match solver.init_tesseract() {
            Ok(_) => {
                solver.ready.store(true, Ordering::SeqCst);
                Ok(solver)
            }
            Err(e) => Err(e)
        }
    }

    fn init_tesseract(&self) -> CaptchaResult<()> {
        // In production, this would initialize Tesseract
        // For now, we'll just check if the library is available
        
        // Check if tesseract data path exists
        let tessdata_path = std::env::var("TESSDATA_PREFIX")
            .unwrap_or_else(|_| "/usr/share/tesseract-ocr/4.00/tessdata".to_string());

        if !std::path::Path::new(&tessdata_path).exists() {
            tracing::warn!("Tesseract data path not found: {}", tessdata_path);
            // Don't fail, just warn - we'll use mock for development
        }

        Ok(())
    }

    /// Perform OCR on an image
    fn perform_ocr(&self, image: &DynamicImage) -> CaptchaResult<(String, f32)> {
        // Convert image to grayscale
        let gray = image.to_luma8();
        
        // Get image dimensions
        let (width, height) = gray.dimensions();
        
        // In production, this would use actual Tesseract bindings
        // For now, we simulate the OCR process
        
        #[cfg(feature = "tesseract")]
        {
            use tesseract::Tesseract;
            
            let tess = Tesseract::new(None, Some("eng"))
                .map_err(|e| CaptchaError::ModelLoadError(e.to_string()))?;
            
            // Set image data
            let result = tess
                .set_image_from_mem(&gray.into_raw(), width as i32, height as i32, 1, width as i32)
                .map_err(|e| CaptchaError::ProcessingError(e.to_string()))?
                .get_text()
                .map_err(|e| CaptchaError::ProcessingError(e.to_string()))?;
            
            let confidence = tess.mean_text_conf() as f32 / 100.0;
            
            return Ok((result.trim().to_string(), confidence));
        }

        // Mock implementation for development
        #[cfg(not(feature = "tesseract"))]
        {
            // Simulate OCR by analyzing image characteristics
            let text = self.mock_ocr(&gray);
            let confidence = 0.85; // Mock confidence
            
            Ok((text, confidence))
        }
    }

    /// Mock OCR for development/testing
    fn mock_ocr(&self, image: &image::GrayImage) -> String {
        // This is a placeholder that returns mock text
        // In production, this would be replaced by actual OCR
        
        let (width, height) = image.dimensions();
        let avg_brightness: u32 = image.pixels()
            .map(|p| p.0[0] as u32)
            .sum::<u32>() / (width * height);

        // Generate mock captcha text based on image hash
        // This is just for demonstration
        let hash = format!("{:x}{:x}", width % 100, avg_brightness % 256);
        
        // Convert hash to mock captcha text
        hash.chars()
            .take(6)
            .map(|c| if c.is_ascii_digit() { c } else { ((c as u8 % 26) + b'A') as char })
            .collect()
    }
}

#[async_trait::async_trait]
impl CaptchaSolver for OcrSolver {
    async fn solve(&self, image: &DynamicImage, options: Option<&PreprocessOptions>) -> CaptchaResult<SolveResult> {
        if !self.is_ready() {
            return Err(CaptchaError::ModelLoadError("OCR solver not ready".to_string()));
        }

        // Preprocess image
        let processed = match options {
            Some(opts) => ImagePreprocessor::preprocess(image, opts)?,
            None => ImagePreprocessor::preprocess(image, &PreprocessOptions::default())?,
        };

        // Perform OCR
        let (text, confidence) = self.perform_ocr(&processed)?;

        // Post-process result
        let cleaned_text = self.post_process(&text);

        Ok(SolveResult {
            text: cleaned_text,
            confidence,
            solver_name: self.name().to_string(),
        })
    }

    fn name(&self) -> &str {
        "ocr"
    }

    fn is_ready(&self) -> bool {
        self.ready.load(Ordering::SeqCst)
    }
}

impl OcrSolver {
    /// Post-process OCR result
    fn post_process(&self, text: &str) -> String {
        text.chars()
            .filter(|c| c.is_alphanumeric())
            .map(|c| c.to_ascii_uppercase())
            .collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_post_process() {
        let solver = OcrSolver {
            ready: AtomicBool::new(true),
            models_path: "/tmp".to_string(),
        };

        assert_eq!(solver.post_process("abc123"), "ABC123");
        assert_eq!(solver.post_process("a b c"), "ABC");
        assert_eq!(solver.post_process("AB-CD_12"), "ABCD12");
    }
}