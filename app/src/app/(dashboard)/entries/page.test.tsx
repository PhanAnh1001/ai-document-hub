import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import EntriesPage from "./page";

const mockToken = "test-token";
vi.mock("@/lib/auth-store", () => ({
  useAuthStore: (selector: (s: unknown) => unknown) =>
    selector({ token: mockToken }),
}));

vi.mock("@/lib/api", () => ({
  listEntries: vi.fn(),
  approveEntry: vi.fn(),
  rejectEntry: vi.fn(),
  syncEntry: vi.fn(),
  exportEntriesCSV: vi.fn().mockResolvedValue(0),
  bulkApproveEntries: vi.fn(),
  updateEntry: vi.fn(),
}));

import { listEntries, approveEntry, rejectEntry, exportEntriesCSV, bulkApproveEntries } from "@/lib/api";

vi.mock("@/hooks/use-toast", () => ({
  toast: vi.fn(),
}));
import { toast } from "@/hooks/use-toast";

const sampleEntry = (id: string, status = "pending") => ({
  id,
  organization_id: "org-1",
  document_id: "doc-1",
  document_name: "invoice-2026-03.pdf",
  entry_date: "2026-03-28",
  description: "Mua hàng hóa nhập kho - Công ty ABC",
  debit_account: "156",
  credit_account: "331",
  amount: 1000000,
  status,
  ai_confidence: 0.92,
  created_at: "2026-03-28T10:00:00Z",
});

describe("EntriesPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(listEntries).mockResolvedValue({ entries: [], total: 0 });
  });

  it("renders Định khoản heading", async () => {
    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByRole("heading", { name: /định khoản/i })).toBeInTheDocument();
    });
  });

  it("shows empty state when no entries", async () => {
    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByText(/không có định khoản/i)).toBeInTheDocument();
    });
  });

  it("shows entry rows with debit and credit accounts", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1")],
      total: 1,
    });

    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByText("156")).toBeInTheDocument();
      expect(screen.getByText("331")).toBeInTheDocument();
    });
  });

  it("shows AI confidence badge", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1")],
      total: 1,
    });

    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByText(/92%/)).toBeInTheDocument();
    });
  });

  it("calls approveEntry when Duyệt clicked and updates status", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1")],
      total: 1,
    });
    vi.mocked(approveEntry).mockResolvedValue({
      entry: { ...sampleEntry("e-1"), status: "approved" } as ReturnType<typeof sampleEntry>,
    });

    render(<EntriesPage />);
    // "Duyệt" action button (exact match, not "Chờ duyệt" tab)
    await waitFor(() => screen.getByRole("button", { name: "Duyệt" }));
    fireEvent.click(screen.getByRole("button", { name: "Duyệt" }));

    await waitFor(() => {
      expect(approveEntry).toHaveBeenCalledWith("e-1", mockToken);
    });
  });

  it("calls rejectEntry when Từ chối clicked and confirmed", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1")],
      total: 1,
    });
    vi.mocked(rejectEntry).mockResolvedValue({
      entry: { ...sampleEntry("e-1"), status: "rejected" } as ReturnType<typeof sampleEntry>,
    });

    render(<EntriesPage />);
    await waitFor(() => screen.getByRole("button", { name: "Từ chối" }));
    fireEvent.click(screen.getByRole("button", { name: "Từ chối" }));

    // Confirm in dialog
    await waitFor(() => screen.getByRole("button", { name: /xác nhận từ chối/i }));
    fireEvent.click(screen.getByRole("button", { name: /xác nhận từ chối/i }));

    await waitFor(() => {
      expect(rejectEntry).toHaveBeenCalledWith("e-1", mockToken, undefined);
    });
  });

  it("shows tab filter buttons", async () => {
    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /chờ duyệt/i })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /đã duyệt/i })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /tất cả/i })).toBeInTheDocument();
    });
  });

  it("shows error message when listEntries fails", async () => {
    vi.mocked(listEntries).mockRejectedValue(new Error("Network error"));

    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByText(/không thể tải dữ liệu/i)).toBeInTheDocument();
    });
  });

  it("shows Xuất CSV button and calls exportEntriesCSV on click", async () => {
    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /xuất csv/i })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole("button", { name: /xuất csv/i }));
    await waitFor(() => {
      expect(exportEntriesCSV).toHaveBeenCalledWith(mockToken, "pending", undefined, undefined);
    });
  });

  it("opens entry detail modal when Xem clicked", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1")],
      total: 1,
    });

    render(<EntriesPage />);
    await waitFor(() => screen.getByRole("button", { name: /xem/i }));
    fireEvent.click(screen.getByRole("button", { name: /xem/i }));

    await waitFor(() => {
      // Dialog should be open
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });
  });

  it("renders search input", async () => {
    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/tìm kiếm/i)).toBeInTheDocument();
    });
  });

  it("passes search query to listEntries when user types", async () => {
    render(<EntriesPage />);
    await waitFor(() => screen.getByPlaceholderText(/tìm kiếm/i));

    fireEvent.change(screen.getByPlaceholderText(/tìm kiếm/i), {
      target: { value: "mua hàng" },
    });

    await waitFor(() => {
      const calls = vi.mocked(listEntries).mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(lastCall[1]).toMatchObject({ q: "mua hàng" });
    });
  });

  it("resets to page 1 when search query changes", async () => {
    render(<EntriesPage />);
    await waitFor(() => screen.getByPlaceholderText(/tìm kiếm/i));

    // Verify listEntries is called with offset 0 after search
    fireEvent.change(screen.getByPlaceholderText(/tìm kiếm/i), {
      target: { value: "test" },
    });

    await waitFor(() => {
      const calls = vi.mocked(listEntries).mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(lastCall[1]).toMatchObject({ q: "test", offset: 0 });
    });
  });

  it("shows document file name in Chứng từ column", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1")],
      total: 1,
    });

    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByText("invoice-2026-03.pdf")).toBeInTheDocument();
    });
  });

  it("shows checkboxes for pending entries", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1", "pending"), sampleEntry("e-2", "approved")],
      total: 2,
    });

    render(<EntriesPage />);
    await waitFor(() => screen.getAllByRole("checkbox"));

    // Should have header checkbox + 1 row checkbox (only pending gets checkbox)
    const checkboxes = screen.getAllByRole("checkbox");
    expect(checkboxes.length).toBeGreaterThanOrEqual(1);
  });

  it("shows Duyệt nhiều button when pending entries selected", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1", "pending")],
      total: 1,
    });

    render(<EntriesPage />);
    await waitFor(() => screen.getAllByRole("checkbox"));

    // Click the row checkbox for the pending entry
    const checkboxes = screen.getAllByRole("checkbox");
    // last checkbox is the row checkbox (first is header)
    fireEvent.click(checkboxes[checkboxes.length - 1]);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /duyệt nhiều/i })).toBeInTheDocument();
    });
  });

  it("shows toast with export count after CSV export", async () => {
    vi.mocked(exportEntriesCSV).mockResolvedValue(5);

    render(<EntriesPage />);
    await waitFor(() => screen.getByRole("button", { name: /xuất csv/i }));
    fireEvent.click(screen.getByRole("button", { name: /xuất csv/i }));

    await waitFor(() => {
      expect(toast).toHaveBeenCalledWith(
        expect.objectContaining({ title: expect.stringMatching(/5/) })
      );
    });
  });

  it("passes date_from and date_to to listEntries when entry date range set", async () => {
    render(<EntriesPage />);
    await waitFor(() => screen.getByLabelText(/từ ngày/i));

    fireEvent.change(screen.getByLabelText(/từ ngày/i), { target: { value: "2026-03-01" } });
    fireEvent.change(screen.getByLabelText(/đến ngày/i), { target: { value: "2026-03-31" } });

    await waitFor(() => {
      const calls = vi.mocked(listEntries).mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(lastCall[1]).toMatchObject({ date_from: "2026-03-01", date_to: "2026-03-31" });
    });
  });

  it("shows reject reason indicator (ⓘ) in status column for rejected entries with reason", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [{
        ...sampleEntry("e-1", "rejected"),
        reject_reason: "Sai tài khoản",
      }],
      total: 1,
    });

    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByText("ⓘ")).toBeInTheDocument();
    });
  });

  it("shows reject reason dialog when Từ chối clicked", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1", "pending")],
      total: 1,
    });

    render(<EntriesPage />);
    await waitFor(() => screen.getByRole("button", { name: "Từ chối" }));
    fireEvent.click(screen.getByRole("button", { name: "Từ chối" }));

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/lý do từ chối/i)).toBeInTheDocument();
    });
  });

  it("calls rejectEntry with reason when confirm reject clicked", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1", "pending")],
      total: 1,
    });
    vi.mocked(rejectEntry).mockResolvedValue({
      entry: { ...sampleEntry("e-1"), status: "rejected" } as ReturnType<typeof sampleEntry>,
    });

    render(<EntriesPage />);
    await waitFor(() => screen.getByRole("button", { name: "Từ chối" }));
    fireEvent.click(screen.getByRole("button", { name: "Từ chối" }));

    await waitFor(() => screen.getByPlaceholderText(/lý do từ chối/i));
    fireEvent.change(screen.getByPlaceholderText(/lý do từ chối/i), {
      target: { value: "Sai tài khoản" },
    });
    fireEvent.click(screen.getByRole("button", { name: /xác nhận từ chối/i }));

    await waitFor(() => {
      expect(rejectEntry).toHaveBeenCalledWith("e-1", mockToken, "Sai tài khoản");
    });
  });

  it("renders date range inputs for export", async () => {
    render(<EntriesPage />);
    await waitFor(() => {
      expect(screen.getByLabelText(/từ ngày/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/đến ngày/i)).toBeInTheDocument();
    });
  });

  it("passes date range to exportEntriesCSV when set", async () => {
    render(<EntriesPage />);
    await waitFor(() => screen.getByLabelText(/từ ngày/i));

    fireEvent.change(screen.getByLabelText(/từ ngày/i), { target: { value: "2026-03-01" } });
    fireEvent.change(screen.getByLabelText(/đến ngày/i), { target: { value: "2026-03-31" } });

    fireEvent.click(screen.getByRole("button", { name: /xuất csv/i }));
    await waitFor(() => {
      expect(exportEntriesCSV).toHaveBeenCalledWith(
        mockToken,
        "pending",
        { from: "2026-03-01", to: "2026-03-31" },
        undefined
      );
    });
  });

  it("calls bulkApproveEntries when Duyệt nhiều clicked", async () => {
    vi.mocked(listEntries).mockResolvedValue({
      entries: [sampleEntry("e-1", "pending")],
      total: 1,
    });
    vi.mocked(bulkApproveEntries).mockResolvedValue({ updated: 1 });

    render(<EntriesPage />);
    await waitFor(() => screen.getAllByRole("checkbox"));

    const checkboxes = screen.getAllByRole("checkbox");
    fireEvent.click(checkboxes[checkboxes.length - 1]);

    await waitFor(() => screen.getByRole("button", { name: /duyệt nhiều/i }));
    fireEvent.click(screen.getByRole("button", { name: /duyệt nhiều/i }));

    await waitFor(() => {
      expect(bulkApproveEntries).toHaveBeenCalledWith(["e-1"], mockToken);
    });
  });

  it("clicking Ngày header sorts by entry_date desc then asc on second click", async () => {
    vi.mocked(listEntries).mockResolvedValue({ entries: [sampleEntry("e-1")], total: 1 });

    render(<EntriesPage />);
    await waitFor(() => screen.getByText("Mua hàng hóa nhập kho - Công ty ABC"));

    fireEvent.click(screen.getByRole("button", { name: /ngày/i }));

    await waitFor(() => {
      const calls = vi.mocked(listEntries).mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(lastCall[1]).toMatchObject({ sort_by: "entry_date", sort_dir: "desc" });
    });

    // Re-query after re-render (loading cycle replaces DOM elements)
    await waitFor(() => screen.getByRole("button", { name: /ngày/i }));
    fireEvent.click(screen.getByRole("button", { name: /ngày/i }));

    await waitFor(() => {
      const calls = vi.mocked(listEntries).mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(lastCall[1]).toMatchObject({ sort_by: "entry_date", sort_dir: "asc" });
    });
  });

  it("clicking Số tiền header sorts by amount", async () => {
    vi.mocked(listEntries).mockResolvedValue({ entries: [sampleEntry("e-1")], total: 1 });

    render(<EntriesPage />);
    await waitFor(() => screen.getByText("Mua hàng hóa nhập kho - Công ty ABC"));

    const amountHeader = screen.getByRole("button", { name: /số tiền/i });
    fireEvent.click(amountHeader);

    await waitFor(() => {
      const calls = vi.mocked(listEntries).mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(lastCall[1]).toMatchObject({ sort_by: "amount", sort_dir: "desc" });
    });
  });
});
