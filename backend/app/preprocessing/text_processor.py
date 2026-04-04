"""
Vietnamese text normalization and post-processing for OCR output.
"""
import re
import unicodedata


class TextProcessor:
    """Normalize and clean OCR text output, with Vietnamese support."""

    def normalize(self, text: str) -> str:
        """Full normalization pipeline."""
        if not text:
            return ""
        text = self._normalize_unicode(text)
        text = self._fix_encoding(text)
        text = self._normalize_whitespace(text)
        text = self._fix_common_ocr_errors(text)
        return text.strip()

    def _normalize_unicode(self, text: str) -> str:
        """Normalize Unicode to NFC form (standard for Vietnamese)."""
        return unicodedata.normalize("NFC", text)

    def _fix_encoding(self, text: str) -> str:
        """Fix common encoding issues in Vietnamese text."""
        # Replace Windows-1258 encoded chars that appear as mojibake
        replacements = {
            "\u00e2\u0081\u00ba": "ề",
            "Ã ": "à",
            "Ã¡": "á",
        }
        for bad, good in replacements.items():
            text = text.replace(bad, good)
        return text

    def _normalize_whitespace(self, text: str) -> str:
        """Collapse multiple spaces/newlines."""
        # Normalize line endings
        text = text.replace("\r\n", "\n").replace("\r", "\n")
        # Collapse multiple spaces (but keep newlines)
        text = re.sub(r" {2,}", " ", text)
        # Collapse 3+ consecutive newlines into 2
        text = re.sub(r"\n{3,}", "\n\n", text)
        return text

    def _fix_common_ocr_errors(self, text: str) -> str:
        """Fix common OCR character substitution errors."""
        # Common OCR errors in numbers
        text = re.sub(r"(?<!\w)O(?=\d)", "0", text)   # O → 0 before digit
        text = re.sub(r"(?<=\d)O(?!\w)", "0", text)   # O → 0 after digit
        text = re.sub(r"(?<!\w)l(?=\d)", "1", text)   # l → 1 before digit
        return text

    def extract_amounts(self, text: str) -> list[float]:
        """Extract Vietnamese currency amounts from text."""
        # Match patterns like: 1,000,000 VND / 1.000.000đ / 5,000,000
        patterns = [
            r"(\d{1,3}(?:[.,]\d{3})+)(?:\s*(?:VND|đồng|đ))?",
            r"(\d+)(?:\s*(?:VND|đồng|đ))",
        ]
        amounts = []
        for pattern in patterns:
            for match in re.finditer(pattern, text, re.IGNORECASE):
                raw = match.group(1).replace(",", "").replace(".", "")
                try:
                    amounts.append(float(raw))
                except ValueError:
                    pass
        return amounts

    def extract_dates(self, text: str) -> list[str]:
        """Extract dates in various Vietnamese formats."""
        patterns = [
            r"\d{4}-\d{2}-\d{2}",             # ISO: 2024-01-15
            r"\d{2}/\d{2}/\d{4}",             # DD/MM/YYYY
            r"\d{2}-\d{2}-\d{4}",             # DD-MM-YYYY
            r"ngày\s+\d{1,2}\s+tháng\s+\d{1,2}\s+năm\s+\d{4}",  # Vietnamese
        ]
        dates = []
        for pattern in patterns:
            dates.extend(re.findall(pattern, text, re.IGNORECASE))
        return dates
