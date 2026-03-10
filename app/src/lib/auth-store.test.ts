import { beforeEach, describe, expect, it, vi } from "vitest";
import { useAuthStore } from "./auth-store";

// Mock document.cookie in jsdom
const cookieStore: Record<string, string> = {};
Object.defineProperty(document, "cookie", {
  get: () => Object.entries(cookieStore).map(([k, v]) => `${k}=${v}`).join("; "),
  set: (val: string) => {
    const [pair] = val.split(";");
    const [key, value] = pair.split("=");
    if (value === "" || val.includes("max-age=0")) {
      delete cookieStore[key.trim()];
    } else {
      cookieStore[key.trim()] = value?.trim() ?? "";
    }
  },
  configurable: true,
});

const mockUser = {
  id: "user-1",
  email: "test@company.vn",
  full_name: "Nguyễn Văn A",
  role: "owner",
  organization_id: "org-1",
};

describe("useAuthStore", () => {
  beforeEach(() => {
    // Reset store and storage before each test
    useAuthStore.setState({ token: null, user: null, isAuthenticated: false });
    localStorage.clear();
    Object.keys(cookieStore).forEach((k) => delete cookieStore[k]);
  });

  it("initial state: token null, user null, isAuthenticated false", () => {
    const { token, user, isAuthenticated } = useAuthStore.getState();
    expect(token).toBeNull();
    expect(user).toBeNull();
    expect(isAuthenticated).toBe(false);
  });

  it("setAuth stores token and user, sets isAuthenticated true", () => {
    useAuthStore.getState().setAuth("my-jwt-token", mockUser);

    const { token, user, isAuthenticated } = useAuthStore.getState();
    expect(token).toBe("my-jwt-token");
    expect(user).toEqual(mockUser);
    expect(isAuthenticated).toBe(true);
  });

  it("setAuth sets token cookie for Next.js middleware", () => {
    useAuthStore.getState().setAuth("my-jwt-token", mockUser);
    expect(document.cookie).toContain("token=my-jwt-token");
  });

  it("clearAuth resets all state to initial values", () => {
    useAuthStore.getState().setAuth("my-jwt-token", mockUser);
    useAuthStore.getState().clearAuth();

    const { token, user, isAuthenticated } = useAuthStore.getState();
    expect(token).toBeNull();
    expect(user).toBeNull();
    expect(isAuthenticated).toBe(false);
  });

  it("clearAuth removes token cookie", () => {
    useAuthStore.getState().setAuth("my-jwt-token", mockUser);
    useAuthStore.getState().clearAuth();
    expect(document.cookie).not.toContain("token=my-jwt-token");
  });
});
