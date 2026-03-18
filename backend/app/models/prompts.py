INVOICE_EXTRACTION_PROMPT = """You are an expert document parser specializing in Vietnamese invoices and receipts.
Extract the following fields from the invoice text and return valid JSON only.

Fields to extract:
- vendor: seller/supplier name (string)
- vendor_address: seller address (string or null)
- tax_id: seller tax identification number (string or null)
- invoice_number: invoice/bill number (string or null)
- date: invoice date in ISO format YYYY-MM-DD (string or null)
- due_date: payment due date (string or null)
- buyer: buyer/customer name (string or null)
- buyer_tax_id: buyer tax ID (string or null)
- currency: currency code, e.g. VND, USD (string, default "VND")
- subtotal: amount before tax (number or null)
- tax_rate: tax percentage (number or null)
- tax_amount: tax amount (number or null)
- total: total amount including tax (number or null)
- line_items: array of {description, quantity, unit_price, amount}
- payment_method: e.g. cash, bank transfer (string or null)
- notes: any additional notes (string or null)

Invoice text:
{text}

Return JSON only, no explanation."""


CONTRACT_EXTRACTION_PROMPT = """You are an expert legal document parser specializing in Vietnamese contracts.
Extract the following fields from the contract text and return valid JSON only.

Fields to extract:
- contract_number: contract reference number (string or null)
- contract_type: type of contract, e.g. service, employment, sale (string or null)
- date: contract signing date in ISO format YYYY-MM-DD (string or null)
- effective_date: date contract takes effect (string or null)
- expiry_date: contract expiry date (string or null)
- party_a: first party name (string or null)
- party_a_address: first party address (string or null)
- party_a_tax_id: first party tax ID (string or null)
- party_b: second party name (string or null)
- party_b_address: second party address (string or null)
- party_b_tax_id: second party tax ID (string or null)
- contract_value: total contract value (number or null)
- currency: currency code (string, default "VND")
- payment_terms: payment conditions (string or null)
- subject: main subject/purpose of contract (string or null)
- jurisdiction: governing law/jurisdiction (string or null)
- key_obligations: list of main obligations (array of strings)
- termination_conditions: conditions for termination (string or null)

Contract text:
{text}

Return JSON only, no explanation."""


CV_EXTRACTION_PROMPT = """You are an expert HR document parser specializing in Vietnamese CVs and resumes.
Extract the following fields from the CV text and return valid JSON only.

Fields to extract:
- full_name: candidate's full name (string or null)
- email: email address (string or null)
- phone: phone number (string or null)
- address: home address (string or null)
- date_of_birth: date of birth in ISO format YYYY-MM-DD (string or null)
- gender: gender (string or null)
- nationality: nationality (string or null)
- objective: career objective/summary (string or null)
- education: array of {institution, degree, major, start_year, end_year, gpa}
- work_experience: array of {company, position, start_date, end_date, responsibilities}
- skills: array of skill strings
- languages: array of {language, proficiency}
- certifications: array of {name, issuer, year}
- references: array of {name, position, company, contact}

CV text:
{text}

Return JSON only, no explanation."""


RAG_GENERATION_PROMPT = """You are a helpful assistant that answers questions based strictly on the provided document context.
Only use information from the context below. If the answer is not in the context, say "Tôi không tìm thấy thông tin này trong tài liệu."

Context from documents:
{context}

Question: {question}

Answer in the same language as the question. Be concise and accurate."""


GENERAL_EXTRACTION_PROMPT = """You are a document parser. Extract key information from the following document text and return valid JSON.

Extract any relevant fields you can identify such as:
- dates, names, amounts, addresses, IDs, reference numbers
- Main subject or purpose of the document
- Key parties involved

Document text:
{text}

Return JSON only with the fields you found."""
