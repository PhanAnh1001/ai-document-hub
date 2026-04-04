const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

import { useAuthStore } from "./auth-store";

export type AuthUser = {
  id: string;
  organization_id: string;
  email: string;
  full_name: string | null;
  role: string;
  created_at: string;
};

export type Document = {
  id: string;
  organization_id: string;
  file_url: string;
  file_name: string;
  document_type: "invoice" | "receipt" | "bank_statement" | "other";
  status: "uploaded" | "processing" | "extracted" | "booked" | "error";
  ocr_data?: unknown;
  created_by: string;
  created_at: string;
  updated_at: string;
};

type FetchOptions = RequestInit & {
  token?: string;
};

async function apiFetch<T>(path: string, opts: FetchOptions = {}): Promise<T> {
  const { token, headers, ...rest } = opts;

  const res = await fetch(`${API_URL}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...headers,
    },
    ...rest,
  });

  if (!res.ok) {
    if (res.status === 401) {
      useAuthStore.getState().clearAuth();
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
    }
    const error = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(error.error || `API error: ${res.status}`);
  }

  return res.json();
}

// Auth
// Documents
export async function uploadDocument(
  file: File,
  documentType: Document["document_type"],
  token: string
): Promise<Document> {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("document_type", documentType);

  const res = await fetch(`${API_URL}/api/documents`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  });
  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(error.error || `API error: ${res.status}`);
  }
  return res.json();
}

type ListDocumentsParams = {
  status?: string;
  q?: string;
  limit?: number;
  offset?: number;
};

export async function listDocuments(
  token: string,
  params?: ListDocumentsParams
): Promise<{ documents: Document[]; total: number }> {
  const qs = new URLSearchParams();
  if (params?.status) qs.set("status", params.status);
  if (params?.q) qs.set("q", params.q);
  if (params?.limit != null) qs.set("limit", String(params.limit));
  if (params?.offset != null) qs.set("offset", String(params.offset));
  const query = qs.toString() ? `?${qs}` : "";
  return apiFetch<{ documents: Document[]; total: number }>(`/api/documents${query}`, {
    token,
  });
}

// Auth
export async function login(email: string, password: string) {
  return apiFetch<{ token: string; user: AuthUser }>(
    "/api/auth/login",
    {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }
  );
}

export async function register(data: {
  email: string;
  password: string;
  full_name: string;
  company_name: string;
}) {
  return apiFetch<{ token: string; user: AuthUser }>(
    "/api/auth/register",
    {
      method: "POST",
      body: JSON.stringify(data),
    }
  );
}

// Accounting entries
export type AccountingEntry = {
  id: string;
  organization_id: string;
  document_id: string | null;
  document_name?: string | null;
  entry_date: string;
  description: string;
  debit_account: string;
  credit_account: string;
  amount: number;
  status: "draft" | "pending" | "approved" | "rejected" | "synced";
  reject_reason?: string | null;
  ai_confidence: number | null;
  created_at: string;
  updated_at: string;
};

type ListEntriesParams = {
  status?: string;
  document_id?: string;
  q?: string;
  limit?: number;
  offset?: number;
  date_from?: string;
  date_to?: string;
  sort_by?: string;
  sort_dir?: string;
};

export async function listEntries(
  token: string,
  params?: ListEntriesParams
): Promise<{ entries: AccountingEntry[]; total: number }> {
  const qs = new URLSearchParams();
  if (params?.status) qs.set("status", params.status);
  if (params?.document_id) qs.set("document_id", params.document_id);
  if (params?.q) qs.set("q", params.q);
  if (params?.limit != null) qs.set("limit", String(params.limit));
  if (params?.offset != null) qs.set("offset", String(params.offset));
  if (params?.date_from) qs.set("date_from", params.date_from);
  if (params?.date_to) qs.set("date_to", params.date_to);
  if (params?.sort_by) qs.set("sort_by", params.sort_by);
  if (params?.sort_dir) qs.set("sort_dir", params.sort_dir);
  const query = qs.toString() ? `?${qs}` : "";
  return apiFetch<{ entries: AccountingEntry[]; total: number }>(
    `/api/entries${query}`,
    { token }
  );
}

export async function approveEntry(
  id: string,
  token: string
): Promise<{ entry: AccountingEntry }> {
  return apiFetch<{ entry: AccountingEntry }>(`/api/entries/${id}/approve`, {
    method: "POST",
    token,
  });
}

export async function rejectEntry(
  id: string,
  token: string,
  reason?: string
): Promise<{ entry: AccountingEntry }> {
  return apiFetch<{ entry: AccountingEntry }>(`/api/entries/${id}/reject`, {
    method: "POST",
    token,
    ...(reason ? { body: JSON.stringify({ reason }) } : {}),
  });
}

export async function syncEntry(
  id: string,
  token: string
): Promise<{ entry: AccountingEntry }> {
  return apiFetch<{ entry: AccountingEntry }>(`/api/entries/${id}/sync`, {
    method: "POST",
    token,
  });
}

// Settings
export type OrgSettings = {
  id: string;
  name: string;
  tax_code: string | null;
  accounting_standard: "TT133" | "TT200";
  ocr_provider: "fpt" | "mock";
  misa_api_url?: string | null;
  misa_api_key?: string | null;
  created_at: string;
  updated_at: string;
};

export async function getSettings(
  token: string
): Promise<{ organization: OrgSettings }> {
  return apiFetch<{ organization: OrgSettings }>("/api/settings", { token });
}

export async function updateSettings(
  token: string,
  data: {
    name?: string;
    tax_code?: string | null;
    accounting_standard?: string;
    ocr_provider?: "fpt" | "mock";
    misa_api_url?: string | null;
    misa_api_key?: string | null;
  }
): Promise<{ organization: OrgSettings }> {
  return apiFetch<{ organization: OrgSettings }>("/api/settings", {
    method: "PUT",
    token,
    body: JSON.stringify(data),
  });
}

export type EntryStats = {
  total: number;
  by_status: Record<string, number>;
  total_approved_amount: number;
  daily_trend?: DayCount[];
};

export async function getEntryStats(
  token: string
): Promise<EntryStats> {
  return apiFetch<EntryStats>(
    "/api/entries/stats",
    { token }
  );
}

export async function getEntry(
  id: string,
  token: string
): Promise<{ entry: AccountingEntry }> {
  return apiFetch<{ entry: AccountingEntry }>(`/api/entries/${id}`, { token });
}

export async function getDocument(
  id: string,
  token: string
): Promise<Document> {
  return apiFetch<Document>(`/api/documents/${id}`, { token });
}

export async function retryDocument(
  id: string,
  token: string
): Promise<Document> {
  return apiFetch<Document>(`/api/documents/${id}/retry`, {
    method: "POST",
    token,
  });
}

// Export entries as CSV — fetches with auth header, triggers download.
// Returns the number of exported records (from X-Export-Count response header).
export async function exportEntriesCSV(
  token: string,
  status?: string,
  dateRange?: { from?: string; to?: string },
  q?: string
): Promise<number> {
  const params = new URLSearchParams();
  if (status) params.set("status", status);
  if (dateRange?.from) params.set("from", dateRange.from);
  if (dateRange?.to) params.set("to", dateRange.to);
  if (q) params.set("q", q);
  const res = await fetch(`${API_URL}/api/entries/export?${params.toString()}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error("export failed");
  const count = parseInt(res.headers.get("X-Export-Count") ?? "0", 10) || 0;
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  const today = new Date().toISOString().slice(0, 10);
  a.download = `entries-${today}.csv`;
  a.click();
  URL.revokeObjectURL(url);
  return count;
}

export async function bulkApproveEntries(ids: string[], token: string) {
  return apiFetch<{ updated: number }>("/api/entries/bulk-approve", {
    method: "POST",
    token,
    body: JSON.stringify({ ids }),
  });
}

export async function updateEntry(
  id: string,
  token: string,
  fields: { description: string; debit_account: string; credit_account: string; amount: number }
) {
  return apiFetch<{ entry: AccountingEntry }>(`/api/entries/${id}`, {
    method: "PATCH",
    token,
    body: JSON.stringify(fields),
  });
}

export type DayCount = {
  date: string;
  count: number;
};

export async function getDocumentStats(
  token: string,
  days = 7,
  docType?: string
): Promise<{ days: DayCount[] }> {
  const params = new URLSearchParams({ days: String(days) });
  if (docType) params.set("type", docType);
  return apiFetch<{ days: DayCount[] }>(`/api/documents/stats?${params}`, { token });
}

export async function changePassword(
  token: string,
  currentPassword: string,
  newPassword: string
): Promise<{ message: string }> {
  return apiFetch<{ message: string }>("/api/auth/change-password", {
    method: "POST",
    token,
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
  });
}

// Health
export async function healthCheck() {
  return apiFetch<{ status: string; timestamp: string; version: string }>(
    "/api/health"
  );
}

export { apiFetch };

// ── AI Document Hub API (FastAPI backend at NEXT_PUBLIC_API_URL, port 8000) ──

const HUB_API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

async function hubFetch<T>(path: string, opts: RequestInit = {}): Promise<T> {
  const res = await fetch(`${HUB_API_URL}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...opts.headers,
    },
    ...opts,
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ detail: res.statusText }));
    throw new Error(error.detail || `API error: ${res.status}`);
  }

  return res.json();
}

export interface HubDocument {
  id: string;
  filename: string;
  original_filename: string;
  doc_type: "invoice" | "contract" | "cv" | "report" | "other";
  status:
    | "uploaded"
    | "ocr_processing"
    | "ocr_done"
    | "extracting"
    | "extracted"
    | "indexing"
    | "indexed"
    | "failed";
  ocr_text?: string;
  ocr_confidence?: number;
  extracted_data?: Record<string, unknown>;
  created_at: string;
}

export interface QueryRequest {
  question: string;
  doc_ids?: string[];
}

export interface QueryResponse {
  answer: string;
  sources: Array<{ doc_id: string; chunk_text: string; score: number }>;
}

export interface EvalStats {
  faithfulness: number;
  answer_relevancy: number;
  context_precision: number;
  total_documents: number;
}

export async function uploadHubDocument(
  file: File,
  docType = "other"
): Promise<HubDocument> {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("doc_type", docType);

  const res = await fetch(`${HUB_API_URL}/api/v1/documents/upload`, {
    method: "POST",
    body: formData,
  });
  if (!res.ok) {
    const error = await res.json().catch(() => ({ detail: res.statusText }));
    throw new Error(error.detail || `API error: ${res.status}`);
  }
  return res.json();
}

export async function listHubDocuments(): Promise<HubDocument[]> {
  return hubFetch<HubDocument[]>("/api/v1/documents/");
}

export async function getHubDocument(id: string): Promise<HubDocument> {
  return hubFetch<HubDocument>(`/api/v1/documents/${id}`);
}

export async function deleteHubDocument(id: string): Promise<void> {
  await hubFetch<void>(`/api/v1/documents/${id}`, { method: "DELETE" });
}

export async function queryRAG(req: QueryRequest): Promise<QueryResponse> {
  return hubFetch<QueryResponse>("/api/v1/query/", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function getEvalStats(): Promise<EvalStats> {
  return hubFetch<EvalStats>("/api/v1/eval/stats");
}

export async function runEvaluation(): Promise<{ message: string }> {
  return hubFetch<{ message: string }>("/api/v1/eval/run", { method: "POST" });
}
