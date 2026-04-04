import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import HubDocumentsPage from "./page";

vi.mock("@/lib/api", () => ({
  listHubDocuments: vi.fn(),
  uploadHubDocument: vi.fn(),
  deleteHubDocument: vi.fn(),
}));

import { listHubDocuments, uploadHubDocument, deleteHubDocument } from "@/lib/api";

const mockDocs = [
  {
    id: "doc-1",
    filename: "invoice-2026.pdf",
    original_filename: "invoice-2026.pdf",
    doc_type: "invoice" as const,
    status: "indexed" as const,
    created_at: "2026-04-01T00:00:00Z",
  },
  {
    id: "doc-2",
    filename: "contract-abc.pdf",
    original_filename: "contract-abc.pdf",
    doc_type: "contract" as const,
    status: "uploaded" as const,
    created_at: "2026-04-02T00:00:00Z",
  },
  {
    id: "doc-3",
    filename: "cv-john.pdf",
    original_filename: "cv-john.pdf",
    doc_type: "cv" as const,
    status: "failed" as const,
    created_at: "2026-04-03T00:00:00Z",
  },
];

describe("HubDocumentsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(listHubDocuments).mockResolvedValue([]);
  });

  it("renders upload button", async () => {
    render(<HubDocumentsPage />);
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /upload tài liệu/i })
      ).toBeInTheDocument();
    });
  });

  it("shows document list with filename, type, status, and date", async () => {
    vi.mocked(listHubDocuments).mockResolvedValue(mockDocs);

    render(<HubDocumentsPage />);
    await waitFor(() => {
      expect(screen.getByText("invoice-2026.pdf")).toBeInTheDocument();
      expect(screen.getByText("contract-abc.pdf")).toBeInTheDocument();
      expect(screen.getByText("cv-john.pdf")).toBeInTheDocument();
    });

    // Type labels
    expect(screen.getByText("Hóa đơn")).toBeInTheDocument();
    expect(screen.getByText("Hợp đồng")).toBeInTheDocument();
    expect(screen.getByText("CV")).toBeInTheDocument();
  });

  it("status badge colors: uploaded=gray, indexed=purple, failed=red", async () => {
    vi.mocked(listHubDocuments).mockResolvedValue(mockDocs);

    render(<HubDocumentsPage />);
    await waitFor(() => {
      // Indexed badge
      const indexedBadge = screen.getByText("Đã index");
      expect(indexedBadge.className).toMatch(/purple/);

      // Uploaded badge
      const uploadedBadge = screen.getByText("Đã tải lên");
      expect(uploadedBadge.className).toMatch(/gray/);

      // Failed badge
      const failedBadge = screen.getByText("Lỗi");
      expect(failedBadge.className).toMatch(/red/);
    });
  });

  it("upload triggers API call with file and doc type", async () => {
    const newDoc = {
      id: "doc-new",
      filename: "report.pdf",
      original_filename: "report.pdf",
      doc_type: "report" as const,
      status: "uploaded" as const,
      created_at: "2026-04-04T00:00:00Z",
    };
    vi.mocked(uploadHubDocument).mockResolvedValue(newDoc);

    render(<HubDocumentsPage />);
    await waitFor(() =>
      screen.getByRole("button", { name: /upload tài liệu/i })
    );

    // Open dialog
    fireEvent.click(screen.getByRole("button", { name: /upload tài liệu/i }));
    await waitFor(() => screen.getByRole("dialog"));

    // Set doc type
    const select = screen.getByLabelText(/loại tài liệu/i);
    fireEvent.change(select, { target: { value: "report" } });

    // Select file
    const fileInput = screen.getByLabelText(/chọn file/i);
    const file = new File(["content"], "report.pdf", { type: "application/pdf" });
    Object.defineProperty(fileInput, "files", { value: [file], configurable: true });
    fireEvent.change(fileInput);

    // Submit
    fireEvent.click(screen.getByRole("button", { name: /tải lên/i }));

    await waitFor(() => {
      expect(uploadHubDocument).toHaveBeenCalledWith(file, "report");
    });
  });

  it("delete button calls deleteHubDocument with correct id", async () => {
    vi.mocked(listHubDocuments).mockResolvedValue([mockDocs[0]]);
    vi.mocked(deleteHubDocument).mockResolvedValue(undefined);

    render(<HubDocumentsPage />);
    await waitFor(() => screen.getByText("invoice-2026.pdf"));

    const deleteBtn = screen.getByRole("button", { name: /xóa tài liệu/i });
    fireEvent.click(deleteBtn);

    await waitFor(() => {
      expect(deleteHubDocument).toHaveBeenCalledWith("doc-1");
    });
  });

  it("clicking row opens detail modal with extracted data", async () => {
    vi.mocked(listHubDocuments).mockResolvedValue([
      {
        ...mockDocs[0],
        status: "extracted" as const,
        extracted_data: { total: 1000000, vendor: "ABC Corp" },
        ocr_text: "Mẫu hóa đơn",
        ocr_confidence: 0.95,
      },
    ]);

    render(<HubDocumentsPage />);
    await waitFor(() => screen.getByText("invoice-2026.pdf"));

    // Click row (not delete button)
    fireEvent.click(screen.getByText("invoice-2026.pdf"));

    await waitFor(() => {
      expect(screen.getByRole("dialog")).toBeInTheDocument();
      expect(screen.getByText(/dữ liệu trích xuất/i)).toBeInTheDocument();
    });
  });

  it("drag-and-drop zone is rendered", async () => {
    render(<HubDocumentsPage />);
    await waitFor(() => {
      expect(screen.getByText(/kéo thả file vào đây/i)).toBeInTheDocument();
    });
  });

  it("shows refresh button", async () => {
    render(<HubDocumentsPage />);
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /làm mới/i })
      ).toBeInTheDocument();
    });
  });

  it("shows empty state when no documents", async () => {
    render(<HubDocumentsPage />);
    await waitFor(() => {
      expect(screen.getByText(/chưa có tài liệu nào/i)).toBeInTheDocument();
    });
  });

  it("shows error when API fails", async () => {
    vi.mocked(listHubDocuments).mockRejectedValue(new Error("Network error"));

    render(<HubDocumentsPage />);
    await waitFor(() => {
      expect(
        screen.getByText(/không thể tải danh sách tài liệu/i)
      ).toBeInTheDocument();
    });
  });

  it("shows ocr_processing badge with correct class", async () => {
    vi.mocked(listHubDocuments).mockResolvedValue([
      {
        ...mockDocs[0],
        status: "ocr_processing" as const,
      },
    ]);

    render(<HubDocumentsPage />);
    await waitFor(() => {
      const badge = screen.getByText("Đang OCR");
      expect(badge.className).toMatch(/blue/);
    });
  });
});
