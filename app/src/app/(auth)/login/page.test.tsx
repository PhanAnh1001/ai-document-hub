import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import LoginPage from "./page";

// Mock next/navigation
const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush }),
}));

// Mock auth store
const mockSetAuth = vi.fn();
vi.mock("@/lib/auth-store", () => ({
  useAuthStore: (selector: (s: unknown) => unknown) =>
    selector({ setAuth: mockSetAuth }),
}));

// Mock api
vi.mock("@/lib/api", () => ({
  login: vi.fn(),
}));

import { login as mockLogin } from "@/lib/api";

describe("LoginPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders email and password inputs", () => {
    render(<LoginPage />);
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/mật khẩu/i)).toBeInTheDocument();
  });

  it("renders submit button with label 'Đăng nhập'", () => {
    render(<LoginPage />);
    expect(
      screen.getByRole("button", { name: /đăng nhập/i })
    ).toBeInTheDocument();
  });

  it("shows validation error when submitted with empty email", async () => {
    render(<LoginPage />);
    fireEvent.click(screen.getByRole("button", { name: /đăng nhập/i }));
    await waitFor(() => {
      expect(screen.getByText(/email không hợp lệ/i)).toBeInTheDocument();
    });
  });

  it("shows validation error when submitted with empty password", async () => {
    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "test@test.vn" },
    });
    fireEvent.click(screen.getByRole("button", { name: /đăng nhập/i }));
    await waitFor(() => {
      expect(
        screen.getByText(/mật khẩu phải có ít nhất/i)
      ).toBeInTheDocument();
    });
  });

  it("calls login API with correct email and password on valid submit", async () => {
    vi.mocked(mockLogin).mockResolvedValue({
      token: "tok",
      user: {
        id: "1",
        email: "test@test.vn",
        full_name: "Test",
        role: "owner",
        organization_id: "org1",
      },
    });

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "test@test.vn" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), {
      target: { value: "secret123" },
    });
    fireEvent.click(screen.getByRole("button", { name: /đăng nhập/i }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith("test@test.vn", "secret123");
    });
  });

  it("calls setAuth and redirects to /dashboard on successful login", async () => {
    const mockUser = {
      id: "1",
      email: "test@test.vn",
      full_name: "Test",
      role: "owner",
      organization_id: "org1",
    };
    vi.mocked(mockLogin).mockResolvedValue({ token: "tok", user: mockUser });

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "test@test.vn" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), {
      target: { value: "secret123" },
    });
    fireEvent.click(screen.getByRole("button", { name: /đăng nhập/i }));

    await waitFor(() => {
      expect(mockSetAuth).toHaveBeenCalledWith("tok", mockUser);
      expect(mockPush).toHaveBeenCalledWith("/dashboard");
    });
  });

  it("shows error message when API returns invalid credentials", async () => {
    vi.mocked(mockLogin).mockRejectedValue(new Error("invalid credentials"));

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "test@test.vn" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), {
      target: { value: "wrongpassword" },
    });
    fireEvent.click(screen.getByRole("button", { name: /đăng nhập/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/email hoặc mật khẩu không đúng/i)
      ).toBeInTheDocument();
    });
  });

  it("disables the button while the request is in flight", async () => {
    vi.mocked(mockLogin).mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "test@test.vn" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), {
      target: { value: "secret123" },
    });
    fireEvent.click(screen.getByRole("button", { name: /đăng nhập/i }));

    expect(
      screen.getByRole("button", { name: /đang đăng nhập/i })
    ).toBeDisabled();
  });
});
