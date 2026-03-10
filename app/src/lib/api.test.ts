import { describe, it, expect, vi, beforeEach } from "vitest";
import { uploadDocument, listDocuments, listEntries, approveEntry, rejectEntry, syncEntry, getEntry, getDocument, exportEntriesCSV, updateSettings } from "./api";

const mockClearAuth = vi.fn();
vi.mock("./auth-store", () => ({
  useAuthStore: {
    getState: () => ({ clearAuth: mockClearAuth }),
  },
}));

const mockFetch = vi.fn();
global.fetch = mockFetch;

describe("uploadDocument", () => {
  beforeEach(() => mockFetch.mockReset());

  it("sends FormData with file and document_type", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ id: "doc-1", file_name: "test.pdf" }),
    });

    const file = new File(["content"], "test.pdf", { type: "application/pdf" });
    await uploadDocument(file, "invoice", "my-token");

    expect(mockFetch).toHaveBeenCalledOnce();
    const [_url, opts] = mockFetch.mock.calls[0];
    expect(opts.method).toBe("POST");
    expect(opts.body).toBeInstanceOf(FormData);

    const fd = opts.body as FormData;
    expect(fd.get("file")).toBe(file);
    expect(fd.get("document_type")).toBe("invoice");
  });

  it("passes Bearer token in Authorization header", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({}),
    });

    const file = new File([""], "f.pdf");
    await uploadDocument(file, "other", "tok-abc");

    const [_url, opts] = mockFetch.mock.calls[0];
    expect(opts.headers?.Authorization).toBe("Bearer tok-abc");
  });

  it("throws on non-ok response", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 400,
      json: async () => ({ error: "file is required" }),
    });

    const file = new File([""], "f.pdf");
    await expect(uploadDocument(file, "other", "tok")).rejects.toThrow("file is required");
  });
});

describe("listDocuments", () => {
  beforeEach(() => mockFetch.mockReset());

  it("fetches from /api/documents with Authorization header", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ documents: [], total: 0 }),
    });

    await listDocuments("my-token");

    expect(mockFetch).toHaveBeenCalledOnce();
    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/documents");
    expect(opts.headers?.Authorization).toBe("Bearer my-token");
  });

  it("throws on 401", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 401,
      json: async () => ({ error: "unauthorized" }),
    });

    await expect(listDocuments("bad-token")).rejects.toThrow("unauthorized");
  });
});

describe("listEntries", () => {
  beforeEach(() => mockFetch.mockReset());

  it("fetches /api/entries with Authorization header", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ entries: [], total: 0 }),
    });

    await listEntries("my-token");

    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/entries");
    expect(opts.headers?.Authorization).toBe("Bearer my-token");
  });

  it("appends status query param", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ entries: [], total: 0 }),
    });

    await listEntries("tok", { status: "pending" });

    const [url] = mockFetch.mock.calls[0];
    expect(url).toContain("status=pending");
  });

  it("returns entries and total", async () => {
    const entries = [
      { id: "e-1", debit_account: "156", credit_account: "331", amount: 1000000, status: "pending" },
    ];
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ entries, total: 1 }),
    });

    const result = await listEntries("tok");
    expect(result.entries).toHaveLength(1);
    expect(result.total).toBe(1);
  });
});

describe("approveEntry", () => {
  beforeEach(() => mockFetch.mockReset());

  it("posts to /api/entries/{id}/approve", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ entry: { id: "e-1", status: "approved" } }),
    });

    const result = await approveEntry("e-1", "tok");

    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/entries/e-1/approve");
    expect(opts.method).toBe("POST");
    expect(result.entry.status).toBe("approved");
  });

  it("throws on error response", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 404,
      json: async () => ({ error: "entry not found" }),
    });

    await expect(approveEntry("bad-id", "tok")).rejects.toThrow("entry not found");
  });
});

describe("rejectEntry", () => {
  beforeEach(() => mockFetch.mockReset());

  it("posts to /api/entries/{id}/reject", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ entry: { id: "e-1", status: "rejected" } }),
    });

    await rejectEntry("e-1", "tok");

    const [url] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/entries/e-1/reject");
  });
});

describe("syncEntry", () => {
  beforeEach(() => mockFetch.mockReset());

  it("posts to /api/entries/{id}/sync", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ entry: { id: "e-1", status: "synced" } }),
    });

    const result = await syncEntry("e-1", "tok");

    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/entries/e-1/sync");
    expect(opts.method).toBe("POST");
    expect(result.entry.status).toBe("synced");
  });
});

describe("getEntry", () => {
  beforeEach(() => mockFetch.mockReset());

  it("fetches GET /api/entries/{id}", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ entry: { id: "e-1", debit_account: "156" } }),
    });

    const result = await getEntry("e-1", "tok");

    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/entries/e-1");
    expect(opts.headers?.Authorization).toBe("Bearer tok");
    expect(result.entry.id).toBe("e-1");
  });

  it("throws on 404", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 404,
      json: async () => ({ error: "entry not found" }),
    });

    await expect(getEntry("bad-id", "tok")).rejects.toThrow("entry not found");
  });
});

describe("getDocument", () => {
  beforeEach(() => mockFetch.mockReset());

  it("fetches GET /api/documents/{id}", async () => {
    const doc = { id: "doc-1", file_name: "invoice.pdf", status: "booked" };
    mockFetch.mockResolvedValue({ ok: true, json: async () => doc });

    const result = await getDocument("doc-1", "tok");

    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/documents/doc-1");
    expect(opts.headers?.Authorization).toBe("Bearer tok");
    expect(result.id).toBe("doc-1");
  });
});

describe("exportEntriesCSV", () => {
  beforeEach(() => mockFetch.mockReset());

  it("calls /api/entries/export with Authorization header", async () => {
    const mockBlob = new Blob(["id,amount\n"], { type: "text/csv" });
    mockFetch.mockResolvedValue({ ok: true, blob: async () => mockBlob, headers: new Headers() });

    // Mock URL.createObjectURL and anchor click
    const mockCreateObjectURL = vi.fn(() => "blob:fake-url");
    const mockRevokeObjectURL = vi.fn();
    vi.stubGlobal("URL", { createObjectURL: mockCreateObjectURL, revokeObjectURL: mockRevokeObjectURL });

    const mockAnchor = { href: "", download: "", click: vi.fn() } as unknown as HTMLAnchorElement;
    vi.spyOn(document, "createElement").mockReturnValueOnce(mockAnchor);

    await exportEntriesCSV("tok", "approved");

    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/entries/export");
    expect(url).toContain("status=approved");
    expect(opts.headers?.Authorization).toBe("Bearer tok");
    expect(mockAnchor.download).toMatch(/^entries-\d{4}-\d{2}-\d{2}\.csv$/);
    expect(mockAnchor.click).toHaveBeenCalled();
  });
});

describe("exportEntriesCSV returns count", () => {
  beforeEach(() => mockFetch.mockReset());

  it("returns the export count from X-Export-Count header", async () => {
    const mockBlob = new Blob(["id,amount\n"], { type: "text/csv" });
    const headers = new Headers({ "X-Export-Count": "7" });
    mockFetch.mockResolvedValue({ ok: true, blob: async () => mockBlob, headers });

    vi.stubGlobal("URL", { createObjectURL: vi.fn(() => "blob:x"), revokeObjectURL: vi.fn() });
    vi.spyOn(document, "createElement").mockReturnValueOnce({
      href: "", download: "", click: vi.fn(),
    } as unknown as HTMLAnchorElement);

    const count = await exportEntriesCSV("tok");
    expect(count).toBe(7);
  });

  it("returns 0 when header missing", async () => {
    const mockBlob = new Blob(["id,amount\n"], { type: "text/csv" });
    const headers = new Headers();
    mockFetch.mockResolvedValue({ ok: true, blob: async () => mockBlob, headers });

    vi.stubGlobal("URL", { createObjectURL: vi.fn(() => "blob:x"), revokeObjectURL: vi.fn() });
    vi.spyOn(document, "createElement").mockReturnValueOnce({
      href: "", download: "", click: vi.fn(),
    } as unknown as HTMLAnchorElement);

    const count = await exportEntriesCSV("tok");
    expect(count).toBe(0);
  });
});

describe("exportEntriesCSV with date range", () => {
  beforeEach(() => mockFetch.mockReset());

  it("appends from and to params when provided", async () => {
    const mockBlob = new Blob(["id,amount\n"], { type: "text/csv" });
    mockFetch.mockResolvedValue({ ok: true, blob: async () => mockBlob, headers: new Headers() });

    vi.stubGlobal("URL", { createObjectURL: vi.fn(() => "blob:x"), revokeObjectURL: vi.fn() });
    vi.spyOn(document, "createElement").mockReturnValueOnce({
      href: "", download: "", click: vi.fn(),
    } as unknown as HTMLAnchorElement);

    await exportEntriesCSV("tok", "approved", { from: "2026-03-01", to: "2026-03-31" });

    const [url] = mockFetch.mock.calls[0];
    expect(url).toContain("from=2026-03-01");
    expect(url).toContain("to=2026-03-31");
  });

  it("does not append from/to when not provided", async () => {
    const mockBlob = new Blob(["id,amount\n"], { type: "text/csv" });
    mockFetch.mockResolvedValue({ ok: true, blob: async () => mockBlob, headers: new Headers() });

    vi.stubGlobal("URL", { createObjectURL: vi.fn(() => "blob:x"), revokeObjectURL: vi.fn() });
    vi.spyOn(document, "createElement").mockReturnValueOnce({
      href: "", download: "", click: vi.fn(),
    } as unknown as HTMLAnchorElement);

    await exportEntriesCSV("tok", "approved");

    const [url] = mockFetch.mock.calls[0];
    expect(url).not.toContain("from=");
    expect(url).not.toContain("to=");
  });
});

describe("exportEntriesCSV with search query", () => {
  beforeEach(() => mockFetch.mockReset());

  it("appends q param when provided", async () => {
    const mockBlob = new Blob(["id,amount\n"], { type: "text/csv" });
    mockFetch.mockResolvedValue({ ok: true, blob: async () => mockBlob, headers: new Headers() });

    vi.stubGlobal("URL", { createObjectURL: vi.fn(() => "blob:x"), revokeObjectURL: vi.fn() });
    vi.spyOn(document, "createElement").mockReturnValueOnce({
      href: "", download: "", click: vi.fn(),
    } as unknown as HTMLAnchorElement);

    await exportEntriesCSV("tok", "approved", undefined, "mua hàng");

    const [url] = mockFetch.mock.calls[0];
    expect(url).toContain("q=mua+h%C3%A0ng");
  });
});

describe("updateSettings", () => {
  beforeEach(() => mockFetch.mockReset());

  it("sends PUT /api/settings with token", async () => {
    const org = { id: "o-1", name: "Công ty", ocr_provider: "fpt" };
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ organization: org }),
    });

    const result = await updateSettings("tok", { name: "Công ty", ocr_provider: "fpt" });

    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toContain("/api/settings");
    expect(opts.method).toBe("PUT");
    expect(opts.headers?.Authorization).toBe("Bearer tok");
    expect(result.organization.ocr_provider).toBe("fpt");
  });
});

describe("apiFetch 401 auto-redirect", () => {
  beforeEach(() => {
    mockFetch.mockReset();
    mockClearAuth.mockReset();
    // jsdom allows setting window.location.href
    delete (window as { location?: unknown }).location;
    (window as { location: { href: string } }).location = { href: "" };
  });

  it("calls clearAuth and redirects to /login on 401 response", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 401,
      json: async () => ({ error: "unauthorized" }),
    });

    await expect(listEntries("expired-token")).rejects.toThrow();

    expect(mockClearAuth).toHaveBeenCalledOnce();
    expect(window.location.href).toBe("/login");
  });

  it("does NOT redirect on non-401 errors", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      json: async () => ({ error: "server error" }),
    });

    await expect(listEntries("tok")).rejects.toThrow("server error");

    expect(mockClearAuth).not.toHaveBeenCalled();
    expect(window.location.href).toBe("");
  });
});
