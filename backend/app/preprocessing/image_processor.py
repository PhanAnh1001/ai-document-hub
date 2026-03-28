"""
Image preprocessing for OCR quality improvement.
Uses PIL (always available); OpenCV is optional.
"""
import logging
from pathlib import Path

logger = logging.getLogger(__name__)


class ImageProcessor:
    """Preprocess images before OCR to improve recognition accuracy."""

    def preprocess(self, image_path: str) -> str:
        """
        Apply preprocessing pipeline to image.
        Returns path to processed image (may be the same file or a temp file).
        """
        try:
            return self._preprocess_with_pil(image_path)
        except Exception as e:
            logger.warning(f"Image preprocessing failed: {e}, using original")
            return image_path

    def _preprocess_with_pil(self, image_path: str) -> str:
        """Basic preprocessing with Pillow."""
        from PIL import Image, ImageEnhance, ImageFilter

        img = Image.open(image_path)

        # Convert to RGB if needed
        if img.mode not in ("RGB", "L"):
            img = img.convert("RGB")

        # Convert to grayscale for OCR
        if img.mode == "RGB":
            img = img.convert("L")

        # Enhance contrast
        enhancer = ImageEnhance.Contrast(img)
        img = enhancer.enhance(1.5)

        # Sharpen
        img = img.filter(ImageFilter.SHARPEN)

        # Save preprocessed image
        out_path = str(Path(image_path).with_suffix(".preprocessed.png"))
        img.save(out_path)
        return out_path

    def _preprocess_with_opencv(self, image_path: str) -> str:
        """Advanced preprocessing with OpenCV (optional)."""
        import cv2
        import numpy as np

        img = cv2.imread(image_path)
        if img is None:
            return image_path

        # Convert to grayscale
        gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

        # Denoise
        denoised = cv2.fastNlMeansDenoising(gray, h=10)

        # Adaptive thresholding for better binarization
        binary = cv2.adaptiveThreshold(
            denoised, 255,
            cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
            cv2.THRESH_BINARY, 11, 2
        )

        out_path = str(Path(image_path).with_suffix(".preprocessed.png"))
        cv2.imwrite(out_path, binary)
        return out_path
