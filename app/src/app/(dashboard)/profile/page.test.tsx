import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import ProfilePage from "./page";

const mockToken = "test-token";
vi.mock("@/lib/auth-store", () => ({
  useAuthStore: (selector: (s: unknown) => unknown) =>
    selector({ token: mockToken, user: { email: "user@test.com", full_name: "Test User" } }),
}));

vi.mock("@/lib/api", () => ({
  changePassword: vi.fn(),
}));

import { changePassword } from "@/lib/api";

describe("ProfilePage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders profile heading", async () => {
    render(<ProfilePage />);
    expect(screen.getByRole("heading", { name: /hồ sơ/i })).toBeInTheDocument();
  });

  it("shows current user email", async () => {
    render(<ProfilePage />);
    expect(screen.getByText(/user@test\.com/i)).toBeInTheDocument();
  });

  it("renders change password form fields", async () => {
    render(<ProfilePage />);
    expect(screen.getByLabelText(/mật khẩu hiện tại/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/mật khẩu mới/i)).toBeInTheDocument();
  });

  it("calls changePassword with correct args on submit", async () => {
    vi.mocked(changePassword).mockResolvedValue({ message: "password updated successfully" });

    render(<ProfilePage />);

    fireEvent.change(screen.getByLabelText(/mật khẩu hiện tại/i), {
      target: { value: "OldPass1!" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu mới/i), {
      target: { value: "NewPass1!" },
    });

    fireEvent.click(screen.getByRole("button", { name: /đổi mật khẩu/i }));

    await waitFor(() => {
      expect(changePassword).toHaveBeenCalledWith(mockToken, "OldPass1!", "NewPass1!");
    });
  });

  it("shows success message after password changed", async () => {
    vi.mocked(changePassword).mockResolvedValue({ message: "password updated successfully" });

    render(<ProfilePage />);

    fireEvent.change(screen.getByLabelText(/mật khẩu hiện tại/i), {
      target: { value: "OldPass1!" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu mới/i), {
      target: { value: "NewPass1!" },
    });

    fireEvent.click(screen.getByRole("button", { name: /đổi mật khẩu/i }));

    await waitFor(() => {
      expect(screen.getByText(/đổi mật khẩu thành công/i)).toBeInTheDocument();
    });
  });

  it("shows error message when changePassword fails", async () => {
    vi.mocked(changePassword).mockRejectedValue(new Error("current password is incorrect"));

    render(<ProfilePage />);

    fireEvent.change(screen.getByLabelText(/mật khẩu hiện tại/i), {
      target: { value: "WrongPass!" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu mới/i), {
      target: { value: "NewPass1!" },
    });

    fireEvent.click(screen.getByRole("button", { name: /đổi mật khẩu/i }));

    await waitFor(() => {
      expect(screen.getByText(/current password is incorrect/i)).toBeInTheDocument();
    });
  });
});
