"""
OCR service: PaddleOCR primary, Google Vision fallback, mock for testing.
"""
import logging
import os

logger = logging.getLogger(__name__)


class OCRService:
    def __init__(self, settings):
        self.settings = settings
        self._paddle_ocr = None

    async def extract_text(self, file_path: str, mime_type: str) -> tuple[str, float, str]:
        """
        Extract text from a file.
        Returns: (text, confidence, provider)
        provider = "text" | "paddle" | "google_vision" | "mock"
        """
        # Plain text files: just read them
        if mime_type in ("text/plain",) or file_path.endswith(".txt"):
            return await self._read_text_file(file_path)

        # PDF: try text extraction
        if mime_type == "application/pdf" or file_path.endswith(".pdf"):
            return await self._extract_pdf(file_path)

        # Images: try PaddleOCR, then Google Vision, then mock
        if self.settings.PADDLE_OCR_ENABLED:
            result = await self._paddle_extract(file_path)
            if result:
                return result

        if self.settings.GOOGLE_VISION_API_KEY:
            result = await self._google_vision_extract(file_path)
            if result:
                return result

        return await self._mock_extract(file_path)

    async def _read_text_file(self, file_path: str) -> tuple[str, float, str]:
        """Read plain text file directly."""
        try:
            import aiofiles
            async with aiofiles.open(file_path, "r", encoding="utf-8", errors="replace") as f:
                text = await f.read()
            return text.strip(), 1.0, "text"
        except Exception as e:
            logger.error(f"Failed to read text file {file_path}: {e}")
            return "", 0.0, "text"

    async def _extract_pdf(self, file_path: str) -> tuple[str, float, str]:
        """Extract text from PDF using pdfminer or fallback to mock."""
        try:
            import asyncio
            from io import StringIO

            def _extract_sync():
                try:
                    from pdfminer.high_level import extract_text as pdf_extract
                    return pdf_extract(file_path)
                except ImportError:
                    return None

            loop = asyncio.get_event_loop()
            text = await loop.run_in_executor(None, _extract_sync)
            if text:
                return text.strip(), 0.95, "pdf_text"
        except Exception as e:
            logger.warning(f"PDF extraction failed: {e}")

        return await self._mock_extract(file_path)

    async def _paddle_extract(self, file_path: str) -> tuple[str, float, str] | None:
        """Extract text using PaddleOCR."""
        try:
            import asyncio

            def _run_paddle():
                from paddleocr import PaddleOCR
                ocr = PaddleOCR(use_angle_cls=True, lang="vi", show_log=False)
                result = ocr.ocr(file_path, cls=True)
                if not result or not result[0]:
                    return None, 0.0
                texts = []
                confidences = []
                for line in result[0]:
                    text = line[1][0]
                    conf = line[1][1]
                    texts.append(text)
                    confidences.append(conf)
                avg_conf = sum(confidences) / len(confidences) if confidences else 0.0
                return "\n".join(texts), avg_conf

            loop = asyncio.get_event_loop()
            text, conf = await loop.run_in_executor(None, _run_paddle)
            if text:
                return text, conf, "paddle"
        except ImportError:
            logger.debug("PaddleOCR not installed")
        except Exception as e:
            logger.warning(f"PaddleOCR failed: {e}")
        return None

    async def _google_vision_extract(self, file_path: str) -> tuple[str, float, str] | None:
        """Extract text using Google Cloud Vision."""
        try:
            import asyncio

            def _run_vision():
                from google.cloud import vision
                client = vision.ImageAnnotatorClient()
                with open(file_path, "rb") as f:
                    content = f.read()
                image = vision.Image(content=content)
                response = client.text_detection(image=image)
                texts = response.text_annotations
                if texts:
                    return texts[0].description, 0.9
                return None, 0.0

            loop = asyncio.get_event_loop()
            text, conf = await loop.run_in_executor(None, _run_vision)
            if text:
                return text, conf, "google_vision"
        except ImportError:
            logger.debug("google-cloud-vision not installed")
        except Exception as e:
            logger.warning(f"Google Vision failed: {e}")
        return None

    async def _mock_extract(self, file_path: str) -> tuple[str, float, str]:
        """Mock OCR for testing when no real OCR is available."""
        filename = os.path.basename(file_path)
        return f"[Mock OCR result for {filename}]", 0.75, "mock"
