import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import ChatPage from "./page";

// jsdom doesn't implement scrollIntoView
window.HTMLElement.prototype.scrollIntoView = vi.fn();

vi.mock("@/lib/api", () => ({
  queryRAG: vi.fn(),
  listHubDocuments: vi.fn(),
}));

import { queryRAG, listHubDocuments } from "@/lib/api";

describe("ChatPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(listHubDocuments).mockResolvedValue([]);
  });

  it("renders chat input textarea", async () => {
    render(<ChatPage />);
    await waitFor(() => {
      expect(
        screen.getByRole("textbox", { name: /nhập câu hỏi/i })
      ).toBeInTheDocument();
    });
  });

  it("send button is disabled when input is empty", async () => {
    render(<ChatPage />);
    await waitFor(() => {
      const btn = screen.getByRole("button", { name: /gửi tin nhắn/i });
      expect(btn).toBeDisabled();
    });
  });

  it("send button is enabled when input has text", async () => {
    render(<ChatPage />);
    await waitFor(() =>
      screen.getByRole("textbox", { name: /nhập câu hỏi/i })
    );
    fireEvent.change(screen.getByRole("textbox", { name: /nhập câu hỏi/i }), {
      target: { value: "Hợp đồng này là gì?" },
    });
    const btn = screen.getByRole("button", { name: /gửi tin nhắn/i });
    expect(btn).not.toBeDisabled();
  });

  it("shows user message after sending", async () => {
    vi.mocked(queryRAG).mockResolvedValue({
      answer: "Đây là câu trả lời.",
      sources: [],
    });

    render(<ChatPage />);
    await waitFor(() =>
      screen.getByRole("textbox", { name: /nhập câu hỏi/i })
    );

    fireEvent.change(screen.getByRole("textbox", { name: /nhập câu hỏi/i }), {
      target: { value: "Hợp đồng này là gì?" },
    });
    fireEvent.click(screen.getByRole("button", { name: /gửi tin nhắn/i }));

    await waitFor(() => {
      expect(screen.getByText("Hợp đồng này là gì?")).toBeInTheDocument();
    });
  });

  it("shows loading state while waiting for answer", async () => {
    let resolveQuery!: (v: { answer: string; sources: never[] }) => void;
    vi.mocked(queryRAG).mockImplementation(
      () =>
        new Promise((res) => {
          resolveQuery = res;
        })
    );

    render(<ChatPage />);
    await waitFor(() =>
      screen.getByRole("textbox", { name: /nhập câu hỏi/i })
    );

    fireEvent.change(screen.getByRole("textbox", { name: /nhập câu hỏi/i }), {
      target: { value: "Test câu hỏi" },
    });
    fireEvent.click(screen.getByRole("button", { name: /gửi tin nhắn/i }));

    await waitFor(() => {
      expect(
        screen.getByLabelText(/đang tải câu trả lời/i)
      ).toBeInTheDocument();
    });

    // Cleanup
    resolveQuery({ answer: "done", sources: [] });
  });

  it("shows AI answer with sources after response", async () => {
    vi.mocked(queryRAG).mockResolvedValue({
      answer: "Tổng tiền là 100 triệu đồng.",
      sources: [
        {
          doc_id: "abc123",
          chunk_text: "Số tiền thanh toán: 100.000.000 VNĐ",
          score: 0.92,
        },
      ],
    });

    render(<ChatPage />);
    await waitFor(() =>
      screen.getByRole("textbox", { name: /nhập câu hỏi/i })
    );

    fireEvent.change(screen.getByRole("textbox", { name: /nhập câu hỏi/i }), {
      target: { value: "Tổng tiền là bao nhiêu?" },
    });
    fireEvent.click(screen.getByRole("button", { name: /gửi tin nhắn/i }));

    await waitFor(() => {
      expect(
        screen.getByText("Tổng tiền là 100 triệu đồng.")
      ).toBeInTheDocument();
    });

    // Sources section header
    expect(screen.getByText(/nguồn tham khảo/i)).toBeInTheDocument();
  });

  it("shows error message when API call fails", async () => {
    vi.mocked(queryRAG).mockRejectedValue(new Error("Server error"));

    render(<ChatPage />);
    await waitFor(() =>
      screen.getByRole("textbox", { name: /nhập câu hỏi/i })
    );

    fireEvent.change(screen.getByRole("textbox", { name: /nhập câu hỏi/i }), {
      target: { value: "Test" },
    });
    fireEvent.click(screen.getByRole("button", { name: /gửi tin nhắn/i }));

    await waitFor(() => {
      expect(screen.getByText(/lỗi.*server error/i)).toBeInTheDocument();
    });
  });

  it("shows example prompts in empty state", async () => {
    render(<ChatPage />);
    await waitFor(() => {
      expect(
        screen.getByText(/bắt đầu đặt câu hỏi về tài liệu/i)
      ).toBeInTheDocument();
    });
  });

  it("clicking example prompt fills input", async () => {
    render(<ChatPage />);
    await waitFor(() =>
      screen.getByText(/hợp đồng này có những điều khoản/i)
    );

    fireEvent.click(
      screen.getByText(/hợp đồng này có những điều khoản/i)
    );

    await waitFor(() => {
      const textarea = screen.getByRole("textbox", { name: /nhập câu hỏi/i }) as HTMLTextAreaElement;
      expect(textarea.value).toContain("Hợp đồng này có những điều khoản");
    });
  });

  it("shows document filter chips when indexed docs exist", async () => {
    vi.mocked(listHubDocuments).mockResolvedValue([
      {
        id: "doc-1",
        filename: "contract.pdf",
        original_filename: "contract.pdf",
        doc_type: "contract",
        status: "indexed",
        created_at: "2026-04-01T00:00:00Z",
      },
    ]);

    render(<ChatPage />);
    await waitFor(() => {
      expect(screen.getByText("contract.pdf")).toBeInTheDocument();
    });
  });

  it("clears input after sending message", async () => {
    vi.mocked(queryRAG).mockResolvedValue({ answer: "OK", sources: [] });

    render(<ChatPage />);
    await waitFor(() =>
      screen.getByRole("textbox", { name: /nhập câu hỏi/i })
    );

    const textarea = screen.getByRole("textbox", { name: /nhập câu hỏi/i }) as HTMLTextAreaElement;
    fireEvent.change(textarea, { target: { value: "câu hỏi test" } });
    fireEvent.click(screen.getByRole("button", { name: /gửi tin nhắn/i }));

    await waitFor(() => {
      expect(textarea.value).toBe("");
    });
  });
});
