import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import SettingsPage from "./page";

const mockToken = "test-token";
vi.mock("@/lib/auth-store", () => ({
  useAuthStore: (selector: (s: unknown) => unknown) =>
    selector({ token: mockToken }),
}));

vi.mock("@/lib/api", () => ({
  getSettings: vi.fn(),
  updateSettings: vi.fn(),
}));

import { getSettings, updateSettings } from "@/lib/api";

const sampleOrg = {
  id: "org-1",
  name: "Công ty ABC",
  tax_code: "0123456789",
  accounting_standard: "TT200",
  ocr_provider: "mock",
  misa_api_url: null,
  misa_api_key: null,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

describe("SettingsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getSettings).mockResolvedValue({ organization: sampleOrg });
  });

  it("renders settings heading", async () => {
    render(<SettingsPage />);
    expect(screen.getByText(/cài đặt/i)).toBeInTheDocument();
  });

  it("loads and displays organization name", async () => {
    render(<SettingsPage />);
    await waitFor(() => {
      expect(screen.getByDisplayValue("Công ty ABC")).toBeInTheDocument();
    });
  });

  it("loads and displays tax code", async () => {
    render(<SettingsPage />);
    await waitFor(() => {
      expect(screen.getByDisplayValue("0123456789")).toBeInTheDocument();
    });
  });

  it("calls updateSettings with new values on submit", async () => {
    vi.mocked(updateSettings).mockResolvedValue({ organization: { ...sampleOrg, name: "Công ty XYZ" } });

    render(<SettingsPage />);
    await waitFor(() => screen.getByDisplayValue("Công ty ABC"));

    const nameInput = screen.getByDisplayValue("Công ty ABC");
    fireEvent.change(nameInput, { target: { value: "Công ty XYZ" } });

    fireEvent.click(screen.getByRole("button", { name: /^lưu$/i }));

    await waitFor(() => {
      expect(updateSettings).toHaveBeenCalledWith(
        mockToken,
        expect.objectContaining({ name: "Công ty XYZ" })
      );
    });
  });

  it("shows success message after save", async () => {
    vi.mocked(updateSettings).mockResolvedValue({ organization: sampleOrg });

    render(<SettingsPage />);
    await waitFor(() => screen.getByDisplayValue("Công ty ABC"));

    fireEvent.click(screen.getByRole("button", { name: /^lưu$/i }));

    await waitFor(() => {
      expect(screen.getByText(/đã lưu/i)).toBeInTheDocument();
    });
  });

  it("shows error message when save fails", async () => {
    vi.mocked(updateSettings).mockRejectedValue(new Error("server error"));

    render(<SettingsPage />);
    await waitFor(() => screen.getByDisplayValue("Công ty ABC"));

    fireEvent.click(screen.getByRole("button", { name: /^lưu$/i }));

    await waitFor(() => {
      expect(screen.getByText(/lỗi/i)).toBeInTheDocument();
    });
  });

  it("renders OCR provider select with fpt and mock options", async () => {
    render(<SettingsPage />);
    await waitFor(() => screen.getByDisplayValue("Công ty ABC"));

    const ocrSelect = screen.getByLabelText(/ocr provider/i);
    expect(ocrSelect).toBeInTheDocument();
    expect(screen.getByRole("option", { name: /fpt\.ai/i })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: /mock/i })).toBeInTheDocument();
  });

  it("calls updateSettings with ocr_provider when connection form is saved", async () => {
    vi.mocked(updateSettings).mockResolvedValue({
      organization: { ...sampleOrg, ocr_provider: "fpt" },
    });

    render(<SettingsPage />);
    await waitFor(() => screen.getByDisplayValue("Công ty ABC"));

    const ocrSelect = screen.getByLabelText(/ocr provider/i);
    fireEvent.change(ocrSelect, { target: { value: "fpt" } });

    fireEvent.click(screen.getByRole("button", { name: /lưu kết nối/i }));

    await waitFor(() => {
      expect(updateSettings).toHaveBeenCalledWith(
        mockToken,
        expect.objectContaining({ ocr_provider: "fpt" })
      );
    });
  });

  it("MISA API Key field is type password by default (hidden)", async () => {
    render(<SettingsPage />);
    await waitFor(() => screen.getByLabelText(/misa api key/i));

    const keyInput = screen.getByLabelText(/misa api key/i);
    expect(keyInput).toHaveAttribute("type", "password");
  });

  it("toggles MISA API Key visibility when show/hide button clicked", async () => {
    render(<SettingsPage />);
    await waitFor(() => screen.getByLabelText(/misa api key/i));

    const keyInput = screen.getByLabelText(/misa api key/i);
    expect(keyInput).toHaveAttribute("type", "password");

    const toggleBtn = screen.getByRole("button", { name: /hiện|ẩn/i });
    fireEvent.click(toggleBtn);

    expect(keyInput).toHaveAttribute("type", "text");

    fireEvent.click(toggleBtn);
    expect(keyInput).toHaveAttribute("type", "password");
  });

  it("shows validation error when MISA API URL is not a valid URL", async () => {
    render(<SettingsPage />);
    await waitFor(() => screen.getByDisplayValue("Công ty ABC"));

    const urlInput = screen.getByLabelText(/misa api url/i);
    fireEvent.change(urlInput, { target: { value: "not-a-url" } });

    fireEvent.click(screen.getByRole("button", { name: /lưu kết nối/i }));

    await waitFor(() => {
      expect(screen.getByText(/url không hợp lệ/i)).toBeInTheDocument();
    });
    expect(updateSettings).not.toHaveBeenCalled();
  });

  it("does not show URL error when MISA API URL is empty", async () => {
    render(<SettingsPage />);
    await waitFor(() => screen.getByDisplayValue("Công ty ABC"));

    fireEvent.click(screen.getByRole("button", { name: /lưu kết nối/i }));

    // No error since empty URL is allowed (optional field)
    expect(screen.queryByText(/url không hợp lệ/i)).not.toBeInTheDocument();
  });

  it("calls updateSettings with misa_api_url and misa_api_key", async () => {
    vi.mocked(updateSettings).mockResolvedValue({
      organization: { ...sampleOrg, misa_api_url: "https://misa.test", misa_api_key: "key123" },
    });

    render(<SettingsPage />);
    await waitFor(() => screen.getByDisplayValue("Công ty ABC"));

    const urlInput = screen.getByLabelText(/misa api url/i);
    const keyInput = screen.getByLabelText(/misa api key/i);
    fireEvent.change(urlInput, { target: { value: "https://misa.test" } });
    fireEvent.change(keyInput, { target: { value: "key123" } });

    fireEvent.click(screen.getByRole("button", { name: /lưu kết nối/i }));

    await waitFor(() => {
      expect(updateSettings).toHaveBeenCalledWith(
        mockToken,
        expect.objectContaining({
          misa_api_url: "https://misa.test",
          misa_api_key: "key123",
        })
      );
    });
  });
});
