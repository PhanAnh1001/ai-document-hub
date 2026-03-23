import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import EvaluationPage from "./page";

vi.mock("@/lib/api", () => ({
  getEvalStats: vi.fn(),
  runEvaluation: vi.fn(),
}));

// Recharts uses ResizeObserver which is not available in jsdom
vi.mock("recharts", () => ({
  BarChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="bar-chart">{children}</div>
  ),
  Bar: () => <div data-testid="bar" />,
  XAxis: () => null,
  YAxis: () => null,
  Tooltip: () => null,
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
  Cell: () => null,
}));

import { getEvalStats, runEvaluation } from "@/lib/api";

const mockStats = {
  faithfulness: 0.85,
  answer_relevancy: 0.78,
  context_precision: 0.92,
  total_documents: 42,
};

describe("EvaluationPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getEvalStats).mockResolvedValue(mockStats);
  });

  it("renders evaluation stats cards", async () => {
    render(<EvaluationPage />);
    await waitFor(() => {
      // Total documents card
      expect(screen.getByText("42")).toBeInTheDocument();
    });
    // Metric percentages appear in both cards and table — just verify at least one exists
    expect(screen.getAllByText("85%").length).toBeGreaterThan(0);
    expect(screen.getAllByText("78%").length).toBeGreaterThan(0);
    expect(screen.getAllByText("92%").length).toBeGreaterThan(0);
  });

  it("shows all metric card labels", async () => {
    render(<EvaluationPage />);
    await waitFor(() => {
      expect(screen.getAllByText("Faithfulness").length).toBeGreaterThan(0);
      expect(screen.getAllByText("Answer Relevancy").length).toBeGreaterThan(0);
      expect(screen.getAllByText("Context Precision").length).toBeGreaterThan(0);
    });
  });

  it("handles zero / empty stats gracefully", async () => {
    vi.mocked(getEvalStats).mockResolvedValue({
      faithfulness: 0,
      answer_relevancy: 0,
      context_precision: 0,
      total_documents: 0,
    });

    render(<EvaluationPage />);
    await waitFor(() => {
      // Should show 0 for total docs and 0% for metrics
      expect(screen.getByText("0")).toBeInTheDocument();
    });
    // All metrics should be 0%
    const zeros = screen.getAllByText("0%");
    expect(zeros.length).toBeGreaterThanOrEqual(3);
  });

  it("shows loading state while fetching", () => {
    vi.mocked(getEvalStats).mockImplementation(() => new Promise(() => {}));
    render(<EvaluationPage />);

    // Cards show dash while loading
    const dashes = screen.getAllByText("—");
    expect(dashes.length).toBeGreaterThan(0);
  });

  it("shows error message when fetch fails", async () => {
    vi.mocked(getEvalStats).mockRejectedValue(new Error("Network error"));

    render(<EvaluationPage />);
    await waitFor(() => {
      expect(screen.getByText(/không thể tải số liệu/i)).toBeInTheDocument();
    });
  });

  it("run evaluation button exists and is clickable", async () => {
    vi.mocked(runEvaluation).mockResolvedValue({ message: "Đánh giá hoàn tất" });

    render(<EvaluationPage />);
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /chạy đánh giá/i })
      ).toBeInTheDocument();
    });
  });

  it("shows success message after running evaluation", async () => {
    vi.mocked(runEvaluation).mockResolvedValue({ message: "Đánh giá hoàn tất" });

    render(<EvaluationPage />);
    await waitFor(() =>
      screen.getByRole("button", { name: /chạy đánh giá/i })
    );

    fireEvent.click(screen.getByRole("button", { name: /chạy đánh giá/i }));

    await waitFor(() => {
      expect(screen.getByText("Đánh giá hoàn tất")).toBeInTheDocument();
    });
  });

  it("shows error message when run evaluation fails", async () => {
    vi.mocked(runEvaluation).mockRejectedValue(new Error("Eval failed"));

    render(<EvaluationPage />);
    await waitFor(() =>
      screen.getByRole("button", { name: /chạy đánh giá/i })
    );

    fireEvent.click(screen.getByRole("button", { name: /chạy đánh giá/i }));

    await waitFor(() => {
      expect(screen.getByText(/lỗi.*eval failed/i)).toBeInTheDocument();
    });
  });

  it("renders detail table with metric descriptions", async () => {
    render(<EvaluationPage />);
    await waitFor(() => {
      expect(screen.getByText(/câu trả lời trung thực với nguồn/i)).toBeInTheDocument();
      expect(screen.getByText(/độ liên quan của câu trả lời/i)).toBeInTheDocument();
      expect(screen.getByText(/độ chính xác của ngữ cảnh/i)).toBeInTheDocument();
    });
  });

  it("bar chart renders when stats are available", async () => {
    render(<EvaluationPage />);
    await waitFor(() => {
      expect(screen.getByTestId("bar-chart")).toBeInTheDocument();
    });
  });

  it("refresh button triggers data reload", async () => {
    render(<EvaluationPage />);
    await waitFor(() =>
      screen.getByRole("button", { name: /làm mới/i })
    );

    fireEvent.click(screen.getByRole("button", { name: /làm mới/i }));

    await waitFor(() => {
      expect(getEvalStats).toHaveBeenCalledTimes(2);
    });
  });
});
