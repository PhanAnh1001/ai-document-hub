import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import EntryDetailModal from "./entry-detail-modal";
import type { AccountingEntry } from "@/lib/api";

const sampleEntry: AccountingEntry = {
  id: "entry-1",
  organization_id: "org-1",
  document_id: "doc-1",
  entry_date: "2026-03-28T00:00:00Z",
  description: "Mua hàng hóa nhập kho - Công ty ABC",
  debit_account: "156",
  credit_account: "331",
  amount: 1000000,
  status: "pending",
  ai_confidence: 0.92,
  created_at: "2026-03-28T10:00:00Z",
};

describe("EntryDetailModal", () => {
  it("renders entry debit and credit accounts", () => {
    render(
      <EntryDetailModal entry={sampleEntry} open={true} onClose={vi.fn()} />
    );
    expect(screen.getByText("156")).toBeInTheDocument();
    expect(screen.getByText("331")).toBeInTheDocument();
  });

  it("renders entry amount formatted", () => {
    render(
      <EntryDetailModal entry={sampleEntry} open={true} onClose={vi.fn()} />
    );
    // 1,000,000 formatted
    expect(screen.getByText(/1[,.]000[,.]000/)).toBeInTheDocument();
  });

  it("renders entry description", () => {
    render(
      <EntryDetailModal entry={sampleEntry} open={true} onClose={vi.fn()} />
    );
    expect(screen.getByText(/mua hàng hóa/i)).toBeInTheDocument();
  });

  it("renders entry status badge", () => {
    render(
      <EntryDetailModal entry={sampleEntry} open={true} onClose={vi.fn()} />
    );
    expect(screen.getByText(/pending|chờ duyệt/i)).toBeInTheDocument();
  });

  it("renders AI confidence", () => {
    render(
      <EntryDetailModal entry={sampleEntry} open={true} onClose={vi.fn()} />
    );
    expect(screen.getByText(/92%/)).toBeInTheDocument();
  });

  it("calls onClose when close button clicked", () => {
    const onClose = vi.fn();
    render(
      <EntryDetailModal entry={sampleEntry} open={true} onClose={onClose} />
    );
    fireEvent.click(screen.getByRole("button", { name: /đóng/i }));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("does not render when open=false", () => {
    render(
      <EntryDetailModal entry={sampleEntry} open={false} onClose={vi.fn()} />
    );
    expect(screen.queryByText("156")).not.toBeInTheDocument();
  });

  it("shows Sửa button for pending entries", () => {
    render(
      <EntryDetailModal entry={sampleEntry} open={true} onClose={vi.fn()} />
    );
    expect(screen.getByRole("button", { name: /sửa/i })).toBeInTheDocument();
  });

  it("does not show Sửa button for approved entries", () => {
    render(
      <EntryDetailModal entry={{ ...sampleEntry, status: "approved" }} open={true} onClose={vi.fn()} />
    );
    expect(screen.queryByRole("button", { name: /sửa/i })).not.toBeInTheDocument();
  });

  it("shows edit form when Sửa clicked", () => {
    render(
      <EntryDetailModal entry={sampleEntry} open={true} onClose={vi.fn()} />
    );
    fireEvent.click(screen.getByRole("button", { name: /sửa/i }));
    expect(screen.getByDisplayValue("156")).toBeInTheDocument();
    expect(screen.getByDisplayValue("331")).toBeInTheDocument();
  });

  it("calls onSave with updated fields when form submitted", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(
      <EntryDetailModal entry={sampleEntry} open={true} onClose={vi.fn()} onSave={onSave} />
    );
    fireEvent.click(screen.getByRole("button", { name: /sửa/i }));
    const debitInput = screen.getByDisplayValue("156");
    fireEvent.change(debitInput, { target: { value: "152" } });
    fireEvent.click(screen.getByRole("button", { name: /lưu/i }));
    await waitFor(() => expect(onSave).toHaveBeenCalledOnce());
  });

  it("shows document_name in modal title when available", () => {
    const entryWithName: AccountingEntry = { ...sampleEntry, document_name: "invoice-2026-03.pdf" };
    render(
      <EntryDetailModal entry={entryWithName} open={true} onClose={vi.fn()} />
    );
    expect(screen.getByText(/invoice-2026-03\.pdf/i)).toBeInTheDocument();
  });

  it("shows reject_reason when entry is rejected", () => {
    const rejectedEntry = {
      ...sampleEntry,
      status: "rejected" as const,
      reject_reason: "Sai tài khoản kế toán",
    };
    render(
      <EntryDetailModal entry={rejectedEntry} open={true} onClose={vi.fn()} />
    );
    expect(screen.getByText("Sai tài khoản kế toán")).toBeInTheDocument();
  });

  it("falls back to 'Chi tiết định khoản' when no document_name", () => {
    render(
      <EntryDetailModal entry={sampleEntry} open={true} onClose={vi.fn()} />
    );
    expect(screen.getByText(/chi tiết định khoản/i)).toBeInTheDocument();
  });
});
