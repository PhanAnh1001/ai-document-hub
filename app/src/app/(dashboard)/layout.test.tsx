import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import DashboardLayout from "./layout";

const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush }),
  usePathname: () => "/dashboard",
}));

const mockClearAuth = vi.fn();
vi.mock("@/lib/auth-store", () => ({
  useAuthStore: (selector: (s: unknown) => unknown) =>
    selector({ clearAuth: mockClearAuth, token: "test-token", user: { full_name: "Nguyễn Văn A" } }),
}));

vi.mock("@/lib/api", () => ({
  getEntryStats: vi.fn().mockResolvedValue({ total: 0, by_status: { pending: 3 }, total_approved_amount: 0 }),
}));

describe("DashboardLayout", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders navigation links", () => {
    render(<DashboardLayout><div>content</div></DashboardLayout>);
    expect(screen.getByText("Dashboard")).toBeInTheDocument();
    expect(screen.getByText("Chứng từ")).toBeInTheDocument();
    expect(screen.getByText("Định khoản")).toBeInTheDocument();
    expect(screen.getByText("Cài đặt")).toBeInTheDocument();
  });

  it("renders logout button", () => {
    render(<DashboardLayout><div>content</div></DashboardLayout>);
    expect(screen.getByRole("button", { name: /đăng xuất/i })).toBeInTheDocument();
  });

  it("calls clearAuth and redirects to /login on logout click", async () => {
    render(<DashboardLayout><div>content</div></DashboardLayout>);
    fireEvent.click(screen.getByRole("button", { name: /đăng xuất/i }));
    await waitFor(() => {
      expect(mockClearAuth).toHaveBeenCalledOnce();
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });

  it("renders children", () => {
    render(<DashboardLayout><div data-testid="child">hello</div></DashboardLayout>);
    expect(screen.getByTestId("child")).toBeInTheDocument();
  });

  it("shows pending count badge on Định khoản nav link", async () => {
    render(<DashboardLayout><div>content</div></DashboardLayout>);
    await waitFor(() => {
      expect(screen.getByText("3")).toBeInTheDocument();
    });
  });
});
