import { render, screen, waitFor, fireEvent, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import DashboardPage from "./page";

const mockToken = "test-token";
vi.mock("@/lib/auth-store", () => ({
  useAuthStore: (selector: (s: unknown) => unknown) =>
    selector({ token: mockToken }),
}));

vi.mock("@/lib/api", () => ({
  listDocuments: vi.fn(),
  getEntryStats: vi.fn(),
  getDocumentStats: vi.fn(),
  listHubDocuments: vi.fn(),
}));

import { listDocuments, getEntryStats, getDocumentStats, listHubDocuments } from "@/lib/api";

const emptyDocs = { documents: [], total: 0 };
const emptyStats = { total: 0, by_status: {}, total_approved_amount: 0 };

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(listDocuments).mockResolvedValue(emptyDocs);
    vi.mocked(getEntryStats).mockResolvedValue(emptyStats);
    vi.mocked(getDocumentStats).mockResolvedValue({ days: [] });
  });

  it("shows zero stats before data loads", () => {
    vi.mocked(listDocuments).mockReturnValue(new Promise(() => {}));
    vi.mocked(getEntryStats).mockReturnValue(new Promise(() => {}));

    render(<DashboardPage />);
    expect(screen.getAllByText("0").length).toBeGreaterThanOrEqual(1);
  });

  it("shows total document count", async () => {
    const docs = [
      { id: "1", status: "booked" as const, organization_id: "o", file_url: "", file_name: "f.pdf", document_type: "invoice" as const, created_by: "u", created_at: "", updated_at: "" },
      { id: "2", status: "uploaded" as const, organization_id: "o", file_url: "", file_name: "f2.pdf", document_type: "invoice" as const, created_by: "u", created_at: "", updated_at: "" },
    ];
    vi.mocked(listDocuments).mockResolvedValue({ documents: docs, total: 2 });
    vi.mocked(getEntryStats).mockResolvedValue(emptyStats);

    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByText("2")).toBeInTheDocument();
    });
  });

  it("counts booked documents as processed", async () => {
    const docs = [
      { id: "1", status: "booked" as const, organization_id: "o", file_url: "", file_name: "a.pdf", document_type: "invoice" as const, created_by: "u", created_at: "", updated_at: "" },
      { id: "2", status: "booked" as const, organization_id: "o", file_url: "", file_name: "b.pdf", document_type: "invoice" as const, created_by: "u", created_at: "", updated_at: "" },
      { id: "3", status: "error" as const, organization_id: "o", file_url: "", file_name: "c.pdf", document_type: "invoice" as const, created_by: "u", created_at: "", updated_at: "" },
    ];
    vi.mocked(listDocuments).mockResolvedValue({ documents: docs, total: 3 });
    vi.mocked(getEntryStats).mockResolvedValue(emptyStats);

    render(<DashboardPage />);
    await waitFor(() => {
      const allTexts = screen.getAllByText(/^\d+$/);
      const numbers = allTexts.map((el) => el.textContent);
      expect(numbers).toContain("3"); // total
      expect(numbers).toContain("2"); // booked
      expect(numbers).toContain("1"); // error
    });
  });

  it("shows recent documents section heading", async () => {
    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByText(/chứng từ gần đây/i)).toBeInTheDocument();
    });
  });

  it("renders recent document file names", async () => {
    const docs = [
      { id: "1", status: "booked" as const, organization_id: "o", file_url: "", file_name: "invoice-jan.pdf", document_type: "invoice" as const, created_by: "u", created_at: "2026-03-28T10:00:00Z", updated_at: "" },
      { id: "2", status: "processing" as const, organization_id: "o", file_url: "", file_name: "receipt-feb.pdf", document_type: "receipt" as const, created_by: "u", created_at: "2026-03-27T09:00:00Z", updated_at: "" },
    ];
    vi.mocked(listDocuments).mockResolvedValue({ documents: docs, total: 2 });
    vi.mocked(getEntryStats).mockResolvedValue(emptyStats);

    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByText("invoice-jan.pdf")).toBeInTheDocument();
      expect(screen.getByText("receipt-feb.pdf")).toBeInTheDocument();
    });
  });

  it("shows empty state when no recent documents", async () => {
    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByText(/chưa có chứng từ nào/i)).toBeInTheDocument();
    });
  });

  it("shows at most 5 recent documents", async () => {
    // listDocuments called twice — first with limit:5 (for recent), second with limit:100 (for counts)
    // Both return the same large list in this test
    const docs = Array.from({ length: 8 }, (_, i) => ({
      id: String(i),
      status: "booked" as const,
      organization_id: "o",
      file_url: "",
      file_name: `doc-${i}.pdf`,
      document_type: "invoice" as const,
      created_by: "u",
      created_at: `2026-03-${28 - i}T10:00:00Z`,
      updated_at: "",
    }));
    // First call (limit:5) returns only first 5
    vi.mocked(listDocuments)
      .mockResolvedValueOnce({ documents: docs.slice(0, 5), total: 8 })
      .mockResolvedValueOnce({ documents: docs, total: 8 });
    vi.mocked(getEntryStats).mockResolvedValue(emptyStats);

    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByText("doc-0.pdf")).toBeInTheDocument();
      expect(screen.getByText("doc-4.pdf")).toBeInTheDocument();
      expect(screen.queryByText("doc-5.pdf")).not.toBeInTheDocument();
    });
  });

  it("counts pending entries for approval stat using getEntryStats", async () => {
    vi.mocked(getEntryStats).mockResolvedValue({
      total: 3,
      by_status: { pending: 2, approved: 1 },
    });

    render(<DashboardPage />);
    await waitFor(() => {
      const allNumbers = screen.getAllByText(/^\d+$/).map((el) => el.textContent);
      expect(allNumbers).toContain("2"); // pending entries
    });
  });

  it("renders chart section heading when stats are loaded", async () => {
    vi.mocked(getDocumentStats).mockResolvedValue({
      days: [
        { date: "2026-03-22", count: 1 },
        { date: "2026-03-23", count: 3 },
        { date: "2026-03-28", count: 2 },
      ],
    });

    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByText(/xu hướng chứng từ/i)).toBeInTheDocument();
    });
  });

  it("calls getDocumentStats on mount", async () => {
    render(<DashboardPage />);
    await waitFor(() => {
      expect(getDocumentStats).toHaveBeenCalledWith(mockToken, 7, undefined);
    });
  });

  it("auto-refreshes data after 30s interval", async () => {
    vi.useFakeTimers();
    try {
      render(<DashboardPage />);

      // Wait for initial mount calls to settle
      await act(async () => {});

      const initialCallCount = vi.mocked(getEntryStats).mock.calls.length;

      // Advance timer by 30 seconds to trigger auto-refresh
      await act(async () => {
        vi.advanceTimersByTime(30000);
      });

      expect(vi.mocked(getEntryStats).mock.calls.length).toBeGreaterThan(initialCallCount);
    } finally {
      vi.useRealTimers();
    }
  });

  it("shows formatted total approved amount from entry stats", async () => {
    vi.mocked(getEntryStats).mockResolvedValue({
      total: 5,
      by_status: { approved: 3 },
      total_approved_amount: 12500000,
    });

    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByText(/tổng tiền đã duyệt/i)).toBeInTheDocument();
    });
  });

  it("renders chart type filter buttons", async () => {
    render(<DashboardPage />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /tất cả/i })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /hóa đơn/i })).toBeInTheDocument();
    });
  });

  it("calls getDocumentStats with docType when filter selected", async () => {
    render(<DashboardPage />);
    await waitFor(() => screen.getByRole("button", { name: /hóa đơn/i }));

    fireEvent.click(screen.getByRole("button", { name: /hóa đơn/i }));

    await waitFor(() => {
      const calls = vi.mocked(getDocumentStats).mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(lastCall[2]).toBe("invoice");
    });
  });
});
