import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface AuthUser {
  id: string;
  email: string;
  full_name: string | null;
  role: string;
  organization_id: string;
}

interface AuthState {
  token: string | null;
  user: AuthUser | null;
  isAuthenticated: boolean;
  setAuth: (token: string, user: AuthUser) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      isAuthenticated: false,

      setAuth: (token, user) => {
        // Sync cookie for Next.js server-side middleware
        if (typeof document !== "undefined") {
          document.cookie = `token=${token}; path=/; max-age=${24 * 60 * 60}`;
        }
        set({ token, user, isAuthenticated: true });
      },

      clearAuth: () => {
        // Clear cookie
        if (typeof document !== "undefined") {
          document.cookie = "token=; path=/; max-age=0";
        }
        set({ token: null, user: null, isAuthenticated: false });
      },
    }),
    { name: "auth-storage" }
  )
);
