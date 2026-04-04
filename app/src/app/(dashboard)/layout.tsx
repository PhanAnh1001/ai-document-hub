"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  LayoutDashboard,
  FileText,
  BookOpen,
  Settings,
  User,
  LogOut,
  MessageSquare,
  BarChart2,
  FolderOpen,
} from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import { getEntryStats } from "@/lib/api";
import { Toaster } from "@/components/ui/toaster";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/documents", label: "Chứng từ", icon: FileText },
  { href: "/entries", label: "Định khoản", icon: BookOpen },
  { href: "/settings", label: "Cài đặt", icon: Settings },
  { href: "/profile", label: "Hồ sơ", icon: User },
];

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const { clearAuth, token } = useAuthStore((s) => s as { clearAuth: () => void; token: string });
  const [pendingCount, setPendingCount] = useState(0);

  useEffect(() => {
    if (!token) return;
    getEntryStats(token)
      .then((res) => setPendingCount(res.by_status["pending"] ?? 0))
      .catch(() => {});
  }, [token]);

  const handleLogout = () => {
    clearAuth();
    router.push("/login");
  };

  // Add pending badge data to the Định khoản nav item
  const navItemsWithBadge = navItems.map((item) =>
    item.href === "/entries"
      ? { ...item, badge: pendingCount > 0 ? pendingCount : undefined }
      : { ...item, badge: undefined }
  );

  return (
    <div className="flex min-h-screen">
      <aside className="w-64 border-r bg-muted/40 p-4">
        <div className="mb-8">
          <h2 className="text-lg font-bold">AI Kế Toán</h2>
        </div>
        <nav className="space-y-1">
          {navItemsWithBadge.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium hover:bg-accent"
            >
              <item.icon className="h-4 w-4" />
              <span className="flex-1">{item.label}</span>
              {item.badge != null && (
                <span className="ml-auto inline-flex items-center justify-center rounded-full bg-red-500 px-1.5 py-0.5 text-xs font-bold text-white min-w-[1.25rem]">
                  {item.badge}
                </span>
              )}
            </Link>
          ))}
        </nav>
        <div className="mt-auto pt-8">
          <button
            onClick={handleLogout}
            className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-accent"
          >
            <LogOut className="h-4 w-4" />
            Đăng xuất
          </button>
        </div>
      </aside>
      <main className="flex-1 p-8">{children}</main>
      <Toaster />
    </div>
  );
}
