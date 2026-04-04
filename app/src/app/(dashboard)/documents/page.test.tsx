import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import DocumentsPage from "./page";

const mockToken = "test-token";
vi.mock("@/lib/auth-store", () => ({
  useAuthStore: (selector: (s: unknown) => unknown) =>
    selector({ token: mockToken }),
}));

vi.mock("@/lib/api", () => ({
  listDocuments: vi.fn(),
  uploadDocument: vi.fn(),
  retryDocument: vi.fn(),
}));

import { listDocuments, uploadDocument, retryDocument } from "@/lib/api";

describe("DocumentsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(listDocuments).mockResolvedValue({ documents: [], total: 0 });
  });

  it("renders Upload chứng từ button", async () => {
    render(<DocumentsPage />);
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /upload chứng từ/i })
      ).toBeInTheDocument();
    });
  });

  it("shows upload dialog when button clicked", async () => {
    render(<DocumentsPage />);
    await waitFor(() =>
      expect(screen.getByRole("button", { name: /upload chứng từ/i })).toBeInTheDocument()
    );
    fireEvent.click(screen.getByRole("button", { name: /upload chứng từ/i }));
    await waitFor(() => {
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });
  });

  it("shows file input and document_type select in the dialog", async () => {
    render(<DocumentsPage />);
    await waitFor(() =>
      screen.getByRole("button", { name: /upload chứng từ/i })
    );
    fireEvent.click(screen.getByRole("button", { name: /upload chứng từ/i }));
    await waitFor(() => {
      expect(screen.getByLabelText(/loại chứng từ/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/chọn file/i)).toBeInTheDocument();
    });
  });

  it("shows empty state when no documents exist", async () => {
    render(<DocumentsPage />);
    await waitFor(() => {
      expect(screen.getByText(/chưa có chứng từ/i)).toBeInTheDocument();
    });
  });

  it("shows document list after loading", async () => {
    vi.mocked(listDocuments).mockResolvedValue({
      documents: [
        {
          id: "doc-1",
          organization_id: "org-1",
          file_name: "hoadon.pdf",
          file_url: "/uploads/hoadon.pdf",
          document_type: "invoice",
          status: "uploaded",
          created_by: "user-1",
          created_at: "2026-03-28T10:00:00Z",
          updated_at: "2026-03-28T10:00:00Z",
        },
      ],
      total: 1,
    });

    render(<DocumentsPage />);
    await waitFor(() => {
      expect(screen.getByText("hoadon.pdf")).toBeInTheDocument();
    });
  });

  it("shows PDF icon for pdf files", async () => {
    vi.mocked(listDocuments).mockResolvedValue({
      documents: [{
        id: "doc-1",
        organization_id: "org-1",
        file_name: "invoice.pdf",
        file_url: "/uploads/invoice.pdf",
        document_type: "invoice",
        status: "uploaded",
        created_by: "user-1",
        created_at: "2026-03-28T10:00:00Z",
        updated_at: "2026-03-28T10:00:00Z",
      }],
      total: 1,
    });

    render(<DocumentsPage />);
    await waitFor(() => {
      expect(screen.getByText("PDF")).toBeInTheDocument();
    });
  });

  it("shows IMG icon for image files", async () => {
    vi.mocked(listDocuments).mockResolvedValue({
      documents: [{
        id: "doc-1",
        organization_id: "org-1",
        file_name: "receipt.jpg",
        file_url: "/uploads/receipt.jpg",
        document_type: "receipt",
        status: "uploaded",
        created_by: "user-1",
        created_at: "2026-03-28T10:00:00Z",
        updated_at: "2026-03-28T10:00:00Z",
      }],
      total: 1,
    });

    render(<DocumentsPage />);
    await waitFor(() => {
      expect(screen.getByText("IMG")).toBeInTheDocument();
    });
  });

  it("calls uploadDocument and refreshes list on upload", async () => {
    const newDoc = {
      id: "doc-new",
      organization_id: "org-1",
      file_name: "receipt.pdf",
      file_url: "/uploads/receipt.pdf",
      document_type: "receipt" as const,
      status: "uploaded" as const,
      created_by: "user-1",
      created_at: "2026-03-28T10:00:00Z",
      updated_at: "2026-03-28T10:00:00Z",
    };
    vi.mocked(uploadDocument).mockResolvedValue(newDoc);

    render(<DocumentsPage />);
    await waitFor(() =>
      screen.getByRole("button", { name: /upload chứng từ/i })
    );
    fireEvent.click(screen.getByRole("button", { name: /upload chứng từ/i }));

    await waitFor(() => screen.getByRole("dialog"));

    const fileInput = screen.getByLabelText(/chọn file/i);
    const file = new File(["content"], "receipt.pdf", { type: "application/pdf" });
    Object.defineProperty(fileInput, "files", { value: [file], configurable: true });
    fireEvent.change(fileInput);

    fireEvent.click(screen.getByRole("button", { name: /tải lên/i }));

    await waitFor(() => {
      expect(uploadDocument).toHaveBeenCalledWith(file, expect.any(String), mockToken);
    });
  });

  it("renders status filter select with options", async () => {
    render(<DocumentsPage />);
    await waitFor(() => {
      const select = screen.getByLabelText(/lọc trạng thái/i);
      expect(select).toBeInTheDocument();
    });
    expect(screen.getByRole("option", { name: /tất cả trạng thái/i })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: /lỗi/i })).toBeInTheDocument();
  });

  it("calls listDocuments with status filter when changed", async () => {
    render(<DocumentsPage />);
    await waitFor(() => screen.getByLabelText(/lọc trạng thái/i));

    const select = screen.getByLabelText(/lọc trạng thái/i);
    fireEvent.change(select, { target: { value: "error" } });

    await waitFor(() => {
      expect(listDocuments).toHaveBeenCalledWith(
        mockToken,
        expect.objectContaining({ status: "error" })
      );
    });
  });

  it("opens document detail modal when Xem button is clicked", async () => {
    const doc = {
      id: "doc-1",
      organization_id: "org-1",
      file_url: "/uploads/doc.pdf",
      file_name: "invoice.pdf",
      document_type: "invoice" as const,
      status: "booked" as const,
      created_by: "user-1",
      created_at: "2026-03-28T10:00:00Z",
      updated_at: "2026-03-28T10:00:00Z",
    };
    vi.mocked(listDocuments).mockResolvedValue({ documents: [doc], total: 1 });

    render(<DocumentsPage />);
    await waitFor(() => screen.getByRole("button", { name: /xem/i }));
    fireEvent.click(screen.getByRole("button", { name: /xem/i }));

    await waitFor(() => {
      expect(screen.getByText(/chi tiết chứng từ/i)).toBeInTheDocument();
    });
  });

  it("shows error message when listDocuments fails", async () => {
    vi.mocked(listDocuments).mockRejectedValue(new Error("Network error"));

    render(<DocumentsPage />);
    await waitFor(() => {
      expect(screen.getByText(/không thể tải danh sách/i)).toBeInTheDocument();
    });
  });

  it("shows file name preview after file selected", async () => {
    render(<DocumentsPage />);
    await waitFor(() => screen.getByRole("button", { name: /upload chứng từ/i }));
    fireEvent.click(screen.getByRole("button", { name: /upload chứng từ/i }));
    await waitFor(() => screen.getByRole("dialog"));

    const fileInput = screen.getByLabelText(/chọn file/i);
    const file = new File(["content"], "hoadon-2026-03.pdf", { type: "application/pdf" });
    Object.defineProperty(fileInput, "files", { value: [file], configurable: true });
    fireEvent.change(fileInput);

    await waitFor(() => {
      expect(screen.getByText("hoadon-2026-03.pdf")).toBeInTheDocument();
    });
  });

  it("shows file size in preview after file selected", async () => {
    render(<DocumentsPage />);
    await waitFor(() => screen.getByRole("button", { name: /upload chứng từ/i }));
    fireEvent.click(screen.getByRole("button", { name: /upload chứng từ/i }));
    await waitFor(() => screen.getByRole("dialog"));

    const fileInput = screen.getByLabelText(/chọn file/i);
    // Create a file with known content to check size display
    const content = "x".repeat(1024); // 1KB
    const file = new File([content], "invoice.pdf", { type: "application/pdf" });
    Object.defineProperty(fileInput, "files", { value: [file], configurable: true });
    fireEvent.change(fileInput);

    await waitFor(() => {
      // Should show KB
      expect(screen.getByText(/1(\.\d+)?\s*KB/i)).toBeInTheDocument();
    });
  });

  it("shows error message when upload fails", async () => {
    vi.mocked(uploadDocument).mockRejectedValue(new Error("file is required"));

    render(<DocumentsPage />);
    await waitFor(() =>
      screen.getByRole("button", { name: /upload chứng từ/i })
    );
    fireEvent.click(screen.getByRole("button", { name: /upload chứng từ/i }));
    await waitFor(() => screen.getByRole("dialog"));

    const fileInput = screen.getByLabelText(/chọn file/i);
    const file = new File([""], "test.pdf");
    Object.defineProperty(fileInput, "files", { value: [file], configurable: true });
    fireEvent.change(fileInput);

    fireEvent.click(screen.getByRole("button", { name: /tải lên/i }));

    await waitFor(() => {
      expect(screen.getByText(/upload thất bại/i)).toBeInTheDocument();
    });
  });

  it("shows Retry button for error-status documents and calls retryDocument", async () => {
    vi.mocked(listDocuments).mockResolvedValue({
      documents: [{
        id: "doc-err",
        organization_id: "org-1",
        file_url: "/uploads/invoice.pdf",
        file_name: "invoice.pdf",
        document_type: "invoice",
        status: "error",
        created_by: "user-1",
        created_at: "2026-03-28T10:00:00Z",
        updated_at: "2026-03-28T10:00:00Z",
      }],
      total: 1,
    });
    vi.mocked(retryDocument).mockResolvedValue({ id: "doc-err", status: "processing" } as never);

    render(<DocumentsPage />);
    await waitFor(() => screen.getByText("invoice.pdf"));

    const retryBtn = screen.getByRole("button", { name: /thử lại/i });
    expect(retryBtn).toBeInTheDocument();

    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(retryDocument).toHaveBeenCalledWith("doc-err", mockToken);
    });
  });
});
