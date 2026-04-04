import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import RegisterPage from "./page";

const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush }),
}));

const mockSetAuth = vi.fn();
vi.mock("@/lib/auth-store", () => ({
  useAuthStore: (selector: (s: unknown) => unknown) =>
    selector({ setAuth: mockSetAuth }),
}));

vi.mock("@/lib/api", () => ({
  register: vi.fn(),
}));

import { register as mockRegister } from "@/lib/api";

describe("RegisterPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders all four input fields", () => {
    render(<RegisterPage />);
    expect(screen.getByLabelText(/họ tên/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/tên công ty/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/mật khẩu/i)).toBeInTheDocument();
  });

  it("renders submit button with label 'Đăng ký'", () => {
    render(<RegisterPage />);
    expect(
      screen.getByRole("button", { name: /đăng ký/i })
    ).toBeInTheDocument();
  });

  it("shows validation error for missing required fields on submit", async () => {
    render(<RegisterPage />);
    fireEvent.click(screen.getByRole("button", { name: /đăng ký/i }));
    await waitFor(() => {
      expect(screen.getByText(/họ tên không được để trống/i)).toBeInTheDocument();
    });
  });

  it("shows validation error for invalid email format", async () => {
    render(<RegisterPage />);
    fireEvent.change(screen.getByLabelText(/họ tên/i), {
      target: { value: "Nguyễn Văn A" },
    });
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "not-an-email" },
    });
    fireEvent.change(screen.getByLabelText(/tên công ty/i), {
      target: { value: "Công ty ABC" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), {
      target: { value: "secret123" },
    });
    fireEvent.click(screen.getByRole("button", { name: /đăng ký/i }));
    await waitFor(() => {
      expect(screen.getByText(/email không hợp lệ/i)).toBeInTheDocument();
    });
  });

  it("calls register API with full_name, email, company_name, password", async () => {
    vi.mocked(mockRegister).mockResolvedValue({
      token: "tok",
      user: {
        id: "1",
        email: "test@test.vn",
        full_name: "Nguyễn Văn A",
        role: "owner",
        organization_id: "org1",
      },
    });

    render(<RegisterPage />);
    fireEvent.change(screen.getByLabelText(/họ tên/i), {
      target: { value: "Nguyễn Văn A" },
    });
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "test@test.vn" },
    });
    fireEvent.change(screen.getByLabelText(/tên công ty/i), {
      target: { value: "Công ty ABC" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), {
      target: { value: "secret123" },
    });
    fireEvent.click(screen.getByRole("button", { name: /đăng ký/i }));

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        full_name: "Nguyễn Văn A",
        email: "test@test.vn",
        company_name: "Công ty ABC",
        password: "secret123",
      });
    });
  });

  it("calls setAuth and redirects to /dashboard on successful registration", async () => {
    const mockUser = {
      id: "1",
      email: "test@test.vn",
      full_name: "Nguyễn Văn A",
      role: "owner",
      organization_id: "org1",
    };
    vi.mocked(mockRegister).mockResolvedValue({ token: "tok", user: mockUser });

    render(<RegisterPage />);
    fireEvent.change(screen.getByLabelText(/họ tên/i), {
      target: { value: "Nguyễn Văn A" },
    });
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "test@test.vn" },
    });
    fireEvent.change(screen.getByLabelText(/tên công ty/i), {
      target: { value: "Công ty ABC" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), {
      target: { value: "secret123" },
    });
    fireEvent.click(screen.getByRole("button", { name: /đăng ký/i }));

    await waitFor(() => {
      expect(mockSetAuth).toHaveBeenCalledWith("tok", mockUser);
      expect(mockPush).toHaveBeenCalledWith("/dashboard");
    });
  });

  it("shows 'Email đã được đăng ký' when API returns 409 error", async () => {
    vi.mocked(mockRegister).mockRejectedValue(
      new Error("email already registered")
    );

    render(<RegisterPage />);
    fireEvent.change(screen.getByLabelText(/họ tên/i), {
      target: { value: "Nguyễn Văn A" },
    });
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "existing@test.vn" },
    });
    fireEvent.change(screen.getByLabelText(/tên công ty/i), {
      target: { value: "Công ty ABC" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), {
      target: { value: "secret123" },
    });
    fireEvent.click(screen.getByRole("button", { name: /đăng ký/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/email này đã được đăng ký/i)
      ).toBeInTheDocument();
    });
  });

  it("disables button while request is in flight", async () => {
    vi.mocked(mockRegister).mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(<RegisterPage />);
    fireEvent.change(screen.getByLabelText(/họ tên/i), {
      target: { value: "Nguyễn Văn A" },
    });
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "test@test.vn" },
    });
    fireEvent.change(screen.getByLabelText(/tên công ty/i), {
      target: { value: "Công ty ABC" },
    });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), {
      target: { value: "secret123" },
    });
    fireEvent.click(screen.getByRole("button", { name: /đăng ký/i }));

    expect(
      screen.getByRole("button", { name: /đang đăng ký/i })
    ).toBeDisabled();
  });
});
