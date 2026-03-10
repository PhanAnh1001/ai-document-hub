import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import DocumentDetailModal from "./document-detail-modal";
import type { Document } from "@/lib/api";

const sampleDoc: Document = {
  id: "doc-1",
  organization_id: "org-1",
  file_url: "/uploads/org-1/doc-1-invoice.pdf",
  file_name: "invoice-jan.pdf",
  document_type: "invoice",
  status: "booked",
  created_by: "user-1",
  created_at: "2026-03-28T10:00:00Z",
  updated_at: "2026-03-28T10:05:00Z",
};

describe("DocumentDetailModal", () => {
  it("renders document file name", () => {
    render(<DocumentDetailModal doc={sampleDoc} open onClose={() => {}} />);
    expect(screen.getByText("invoice-jan.pdf")).toBeInTheDocument();
  });

  it("renders document type", () => {
    render(<DocumentDetailModal doc={sampleDoc} open onClose={() => {}} />);
    expect(screen.getByText(/hóa đơn/i)).toBeInTheDocument();
  });

  it("renders document status badge", () => {
    render(<DocumentDetailModal doc={sampleDoc} open onClose={() => {}} />);
    expect(screen.getByText(/đã hạch toán/i)).toBeInTheDocument();
  });

  it("renders upload date", () => {
    render(<DocumentDetailModal doc={sampleDoc} open onClose={() => {}} />);
    expect(screen.getByText(/28\/3\/2026/i)).toBeInTheDocument();
  });

  it("calls onClose when close button clicked", () => {
    const onClose = vi.fn();
    render(<DocumentDetailModal doc={sampleDoc} open onClose={onClose} />);
    fireEvent.click(screen.getByRole("button", { name: /đóng/i }));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("does not render content when open=false", () => {
    render(<DocumentDetailModal doc={sampleDoc} open={false} onClose={() => {}} />);
    expect(screen.queryByText("invoice-jan.pdf")).not.toBeInTheDocument();
  });
});
