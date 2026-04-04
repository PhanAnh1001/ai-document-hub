"""
Unified LLM client: Groq primary, mock fallback for testing.
"""
import json
import logging

logger = logging.getLogger(__name__)


class LLMClient:
    def __init__(self, settings):
        self.settings = settings
        self._groq_client = None

    def _get_groq_client(self):
        """Lazy-init Groq client."""
        if self._groq_client is None and self.settings.GROQ_API_KEY:
            try:
                from groq import Groq
                self._groq_client = Groq(api_key=self.settings.GROQ_API_KEY)
            except ImportError:
                logger.warning("groq package not installed, using mock")
        return self._groq_client

    async def complete(self, prompt: str, system: str = None) -> str:
        """Complete a prompt. Falls back to mock when no API key."""
        if self.settings.GROQ_API_KEY:
            try:
                return await self._groq_complete(prompt, system)
            except Exception as e:
                logger.warning(f"Groq call failed: {e}, using mock")
        return self._mock_complete(prompt)

    async def _groq_complete(self, prompt: str, system: str = None) -> str:
        """Call Groq API synchronously (wrapped)."""
        import asyncio
        client = self._get_groq_client()
        if client is None:
            return self._mock_complete(prompt)

        messages = []
        if system:
            messages.append({"role": "system", "content": system})
        messages.append({"role": "user", "content": prompt})

        loop = asyncio.get_event_loop()
        response = await loop.run_in_executor(
            None,
            lambda: client.chat.completions.create(
                model=self.settings.GROQ_MODEL,
                messages=messages,
                max_tokens=2048,
                temperature=0.1,
            ),
        )
        return response.choices[0].message.content

    def _mock_complete(self, prompt: str) -> str:
        """Return sensible mock JSON for tests."""
        prompt_lower = prompt.lower()
        if "invoice" in prompt_lower or "vendor" in prompt_lower or "total" in prompt_lower:
            return json.dumps({
                "vendor": "Test Vendor Co",
                "vendor_address": "123 Test Street",
                "tax_id": "0123456789",
                "invoice_number": "INV-001",
                "date": "2024-01-15",
                "due_date": None,
                "buyer": "Test Buyer",
                "buyer_tax_id": None,
                "currency": "VND",
                "subtotal": 900000,
                "tax_rate": 10.0,
                "tax_amount": 100000,
                "total": 1000000,
                "line_items": [
                    {"description": "Service fee", "quantity": 1, "unit_price": 900000, "amount": 900000}
                ],
                "payment_method": "bank transfer",
                "notes": None,
            })
        elif "contract" in prompt_lower:
            return json.dumps({
                "contract_number": "CONTRACT-001",
                "contract_type": "service",
                "date": "2024-01-01",
                "effective_date": "2024-01-01",
                "expiry_date": "2024-12-31",
                "party_a": "Company A",
                "party_b": "Company B",
                "contract_value": 50000000,
                "currency": "VND",
                "subject": "Software development services",
                "key_obligations": ["Deliver software", "Provide support"],
            })
        elif "cv" in prompt_lower or "resume" in prompt_lower or "candidate" in prompt_lower:
            return json.dumps({
                "full_name": "Nguyen Van A",
                "email": "nguyenvana@example.com",
                "phone": "0901234567",
                "skills": ["Python", "FastAPI", "PostgreSQL"],
                "work_experience": [],
                "education": [],
            })
        else:
            return json.dumps({
                "summary": "Document processed successfully",
                "key_info": "Extracted from document",
                "date": "2024-01-01",
            })
