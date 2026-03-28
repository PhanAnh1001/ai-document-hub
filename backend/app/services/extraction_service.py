"""
LLM-based structured data extraction service.
"""
import json
import logging
import re
from typing import Any

from app.models.prompts import (
    INVOICE_EXTRACTION_PROMPT,
    CONTRACT_EXTRACTION_PROMPT,
    CV_EXTRACTION_PROMPT,
    GENERAL_EXTRACTION_PROMPT,
)
from app.services.llm_client import LLMClient

logger = logging.getLogger(__name__)

PROMPT_MAP = {
    "invoice": INVOICE_EXTRACTION_PROMPT,
    "contract": CONTRACT_EXTRACTION_PROMPT,
    "cv": CV_EXTRACTION_PROMPT,
}


class ExtractionService:
    def __init__(self, llm_client: LLMClient):
        self.llm = llm_client

    async def extract(self, text: str, doc_type: str) -> dict[str, Any]:
        """
        Extract structured data from text using LLM.
        Returns a dict of extracted fields.
        """
        if not text or not text.strip():
            return {"error": "No text to extract from"}

        prompt_template = PROMPT_MAP.get(doc_type, GENERAL_EXTRACTION_PROMPT)
        # Use simple replace to avoid KeyError from {field} placeholders in prompt
        prompt = prompt_template.replace("{text}", text[:4000])

        raw_response = await self.llm.complete(prompt)
        return self._parse_json_response(raw_response, doc_type)

    def _parse_json_response(self, response: str, doc_type: str) -> dict[str, Any]:
        """Parse JSON from LLM response, handling markdown code blocks."""
        if not response:
            return {"raw_response": ""}

        # Strip markdown code blocks
        cleaned = response.strip()
        cleaned = re.sub(r"^```(?:json)?\s*", "", cleaned)
        cleaned = re.sub(r"\s*```$", "", cleaned)
        cleaned = cleaned.strip()

        try:
            data = json.loads(cleaned)
            if isinstance(data, dict):
                return data
            return {"data": data}
        except json.JSONDecodeError:
            # Try to find JSON object in response
            match = re.search(r"\{.*\}", cleaned, re.DOTALL)
            if match:
                try:
                    return json.loads(match.group())
                except json.JSONDecodeError:
                    pass
            logger.warning(f"Could not parse LLM response as JSON for {doc_type}")
            return {"raw_response": response[:500]}
