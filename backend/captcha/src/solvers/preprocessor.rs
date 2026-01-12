//! Image Preprocessor Module
//!
//! This module provides image preprocessing utilities for captcha solving.

use image::{DynamicImage, GrayImage, ImageBuffer, Luma, Rgb};
use imageproc::contrast::{adaptive_threshold, threshold};
use imageproc::filter::{gaussian_blur_f32, median_filter};
use imageproc::morphology::{dilate, erode};
use imageproc::distance_transform::Norm;

use crate::error::{CaptchaError, CaptchaResult};
use crate::models::PreprocessOptions;

/// Image preprocessor for captcha images
pub struct ImagePreprocessor;

impl ImagePreprocessor {
    /// Preprocess an image according to the given options
    pub fn preprocess(image: &DynamicImage, options: &PreprocessOptions) -> CaptchaResult<DynamicImage> {
        let mut result = image.clone();

        // Resize if dimensions specified
        if let (Some(width), Some(height)) = (options.resize_width, options.resize_height) {
            result = result.resize_exact(width, height, image::imageops::FilterType::Lanczos3);
        }

        // Convert to grayscale if requested
        if options.grayscale.unwrap_or(true) {
            result = DynamicImage::ImageLuma8(result.to_luma8());
        }

        // Apply denoising if requested
        if options.denoise.unwrap_or(false) {
            result = Self::denoise(&result)?;
        }

        // Apply threshold if specified
        if let Some(thresh_value) = options.threshold {
            result = Self::apply_threshold(&result, thresh_value)?;
        }

        Ok(result)
    }

    /// Apply Gaussian blur for denoising
    fn denoise(image: &DynamicImage) -> CaptchaResult<DynamicImage> {
        let gray = image.to_luma8();
        let blurred = gaussian_blur_f32(&gray, 1.0);
        Ok(DynamicImage::ImageLuma8(blurred))
    }

    /// Apply binary threshold
    fn apply_threshold(image: &DynamicImage, thresh_value: u8) -> CaptchaResult<DynamicImage> {
        let gray = image.to_luma8();
        let thresholded = threshold(&gray, thresh_value);
        Ok(DynamicImage::ImageLuma8(thresholded))
    }

    /// Apply adaptive threshold for varying lighting conditions
    pub fn adaptive_threshold(image: &DynamicImage, block_radius: u32) -> CaptchaResult<DynamicImage> {
        let gray = image.to_luma8();
        let thresholded = adaptive_threshold(&gray, block_radius);
        Ok(DynamicImage::ImageLuma8(thresholded))
    }

    /// Apply median filter to remove salt-and-pepper noise
    pub fn median_denoise(image: &DynamicImage, radius: u32) -> CaptchaResult<DynamicImage> {
        let gray = image.to_luma8();
        let filtered = median_filter(&gray, radius, radius);
        Ok(DynamicImage::ImageLuma8(filtered))
    }

    /// Apply morphological erosion
    pub fn erode_image(image: &DynamicImage, radius: u8) -> CaptchaResult<DynamicImage> {
        let gray = image.to_luma8();
        let eroded = erode(&gray, Norm::LInf, radius);
        Ok(DynamicImage::ImageLuma8(eroded))
    }

    /// Apply morphological dilation
    pub fn dilate_image(image: &DynamicImage, radius: u8) -> CaptchaResult<DynamicImage> {
        let gray = image.to_luma8();
        let dilated = dilate(&gray, Norm::LInf, radius);
        Ok(DynamicImage::ImageLuma8(dilated))
    }

    /// Remove noise lines by analyzing connected components
    pub fn remove_lines(image: &DynamicImage) -> CaptchaResult<DynamicImage> {
        let gray = image.to_luma8();
        let (width, height) = gray.dimensions();
        
        // Create output image
        let mut output = gray.clone();

        // Simple line removal: remove very thin horizontal or vertical patterns
        for y in 1..height - 1 {
            for x in 1..width - 1 {
                let current = gray.get_pixel(x, y).0[0];
                
                // Check if this is a thin line (dark pixel with light neighbors)
                if current < 128 {
                    let top = gray.get_pixel(x, y - 1).0[0];
                    let bottom = gray.get_pixel(x, y + 1).0[0];
                    let left = gray.get_pixel(x - 1, y).0[0];
                    let right = gray.get_pixel(x + 1, y).0[0];

                    // Horizontal line detection
                    if top > 200 && bottom > 200 && left < 128 && right < 128 {
                        continue; // Keep character pixels
                    }

                    // Vertical line detection
                    if left > 200 && right > 200 && top < 128 && bottom < 128 {
                        output.put_pixel(x, y, Luma([255])); // Remove line
                    }

                    // Isolated noise detection
                    let neighbor_dark_count = [top, bottom, left, right]
                        .iter()
                        .filter(|&&p| p < 128)
                        .count();

                    if neighbor_dark_count <= 1 {
                        output.put_pixel(x, y, Luma([255])); // Remove isolated noise
                    }
                }
            }
        }

        Ok(DynamicImage::ImageLuma8(output))
    }

    /// Enhance contrast using histogram equalization
    pub fn enhance_contrast(image: &DynamicImage) -> CaptchaResult<DynamicImage> {
        let gray = image.to_luma8();
        let (width, height) = gray.dimensions();

        // Calculate histogram
        let mut histogram = [0u32; 256];
        for pixel in gray.pixels() {
            histogram[pixel.0[0] as usize] += 1;
        }

        // Calculate cumulative distribution function
        let total_pixels = (width * height) as f32;
        let mut cdf = [0f32; 256];
        let mut sum = 0u32;
        for (i, &count) in histogram.iter().enumerate() {
            sum += count;
            cdf[i] = sum as f32 / total_pixels;
        }

        // Apply histogram equalization
        let mut output = gray.clone();
        for (x, y, pixel) in gray.enumerate_pixels() {
            let new_value = (cdf[pixel.0[0] as usize] * 255.0) as u8;
            output.put_pixel(x, y, Luma([new_value]));
        }

        Ok(DynamicImage::ImageLuma8(output))
    }

    /// Segment characters from the image
    pub fn segment_characters(image: &DynamicImage) -> CaptchaResult<Vec<DynamicImage>> {
        let gray = image.to_luma8();
        let (width, height) = gray.dimensions();

        // Find vertical projections to locate character boundaries
        let mut projection = vec![0u32; width as usize];
        for x in 0..width {
            for y in 0..height {
                if gray.get_pixel(x, y).0[0] < 128 {
                    projection[x as usize] += 1;
                }
            }
        }

        // Find character boundaries
        let mut segments = Vec::new();
        let mut in_char = false;
        let mut start = 0u32;

        for (x, &count) in projection.iter().enumerate() {
            if count > 0 && !in_char {
                in_char = true;
                start = x as u32;
            } else if count == 0 && in_char {
                in_char = false;
                if (x as u32) - start > 3 {
                    // Minimum character width
                    let char_image = image.crop_imm(start, 0, (x as u32) - start, height);
                    segments.push(char_image);
                }
            }
        }

        // Handle last character
        if in_char {
            let char_image = image.crop_imm(start, 0, width - start, height);
            segments.push(char_image);
        }

        Ok(segments)
    }

    /// Apply full preprocessing pipeline optimized for text captchas
    pub fn full_pipeline(image: &DynamicImage) -> CaptchaResult<DynamicImage> {
        let options = PreprocessOptions {
            grayscale: Some(true),
            denoise: Some(true),
            threshold: Some(128),
            resize_width: None,
            resize_height: None,
        };

        let result = Self::preprocess(image, &options)?;
        let result = Self::enhance_contrast(&result)?;
        let result = Self::remove_lines(&result)?;
        
        Ok(result)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use image::RgbImage;

    fn create_test_image() -> DynamicImage {
        let img = RgbImage::from_fn(100, 50, |x, y| {
            if x > 20 && x < 80 && y > 10 && y < 40 {
                Rgb([0, 0, 0])
            } else {
                Rgb([255, 255, 255])
            }
        });
        DynamicImage::ImageRgb8(img)
    }

    #[test]
    fn test_preprocess_grayscale() {
        let image = create_test_image();
        let options = PreprocessOptions {
            grayscale: Some(true),
            denoise: None,
            threshold: None,
            resize_width: None,
            resize_height: None,
        };

        let result = ImagePreprocessor::preprocess(&image, &options).unwrap();
        assert!(matches!(result, DynamicImage::ImageLuma8(_)));
    }

    #[test]
    fn test_preprocess_resize() {
        let image = create_test_image();
        let options = PreprocessOptions {
            grayscale: Some(false),
            denoise: None,
            threshold: None,
            resize_width: Some(200),
            resize_height: Some(100),
        };

        let result = ImagePreprocessor::preprocess(&image, &options).unwrap();
        assert_eq!(result.width(), 200);
        assert_eq!(result.height(), 100);
    }

    #[test]
    fn test_preprocess_threshold() {
        let image = create_test_image();
        let options = PreprocessOptions {
            grayscale: Some(true),
            denoise: None,
            threshold: Some(128),
            resize_width: None,
            resize_height: None,
        };

        let result = ImagePreprocessor::preprocess(&image, &options).unwrap();
        
        // Check that pixels are either 0 or 255
        if let DynamicImage::ImageLuma8(gray) = result {
            for pixel in gray.pixels() {
                assert!(pixel.0[0] == 0 || pixel.0[0] == 255);
            }
        }
    }
}