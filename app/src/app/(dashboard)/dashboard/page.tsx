"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { FileText, CheckCircle, Clock, AlertCircle, CircleDot, FolderOpen, Upload } from "lucide-react";
import { listDocuments, getEntryStats, getDocumentStats, listHubDocuments, type Document, type DayCount, type EntryStats, type HubDocument } from "@/lib/api";
import { useAuthStore } from "@/lib/auth-store";
import { Sparkline } from "@/components/sparkline";

function TrendChart({ data }: { data: DayCount[] }) {
  const max = Math.max(...data.map((d) => d.count), 1);
  const chartH = 80;
  const barW = 28;
  const gap = 8;
  const totalW = data.length * (barW + gap) - gap;

  return (
    <div className="overflow-x-auto">
      <svg width={totalW} height={chartH + 28} aria-label="Biểu đồ xu hướng chứng từ">
        {data.map((d, i) => {
          const barH = Math.max(4, Math.round((d.count / max) * chartH));
          const x = i * (barW + gap);
          const y = chartH - barH;
          return (
            <g key={d.date}>
              <rect x={x} y={y} width={barW} height={barH} rx={3} fill="#6366f1" opacity={0.8} />
              <text x={x + barW / 2} y={chartH + 14} textAnchor="middle" fontSize={9} fill="#9ca3af">
                {d.date.slice(5)}
              </text>
              {d.count > 0 && (
                <text x={x + barW / 2} y={y - 3} textAnchor="middle" fontSize={9} fill="#374151">
                  {d.count}
                </text>
              )}
            </g>
          );
        })}
      </svg>
    </div>
  );
}

export default function DashboardPage() {
  const token = useAuthStore((s) => s.token);
  const [docs, setDocs] = useState<Document[]>([]);
  const [entryStats, setEntryStats] = useState<EntryStats>({ total: 0, by_status: {}, total_approved_amount: 0 });
  const [docStats, setDocStats] = useState<Record<string, number>>({});
  const [trend, setTrend] = useState<DayCount[]>([]);
  const [chartType, setChartType] = useState<string | undefined>(undefined);
  const [hubDocs, setHubDocs] = useState<HubDocument[]>([]);

  const fetchAll = useCallback(() => {
    if (!token) return;
    listDocuments(token, { limit: 5 }).then((res) => {
      setDocs(res.documents);
    }).catch(() => {});
    listDocuments(token, { limit: 100 }).then((res) => {
      const counts: Record<string, number> = {};
      for (const d of res.documents) {
        counts[d.status] = (counts[d.status] ?? 0) + 1;
      }
      setDocStats(counts);
    }).catch(() => {});
    getEntryStats(token).then((res) => setEntryStats(res)).catch(() => {});
    getDocumentStats(token, 7, chartType).then((res) => setTrend(res.days)).catch(() => {});
  }, [token, chartType]);

  useEffect(() => {
    fetchAll();
    // Auto-refresh every 30 seconds
    const interval = setInterval(fetchAll, 30000);
    return () => clearInterval(interval);
  }, [fetchAll]);

  const totalDocs = Object.values(docStats).reduce((a, b) => a + b, 0);
  const bookedDocs = docStats["booked"] ?? 0;
  const pendingEntries = entryStats.by_status["pending"] ?? 0;
  const approvedAmount = entryStats.total_approved_amount;
  const errorDocs = docStats["error"] ?? 0;

  const formatVND = (amount: number) =>
    new Intl.NumberFormat("vi-VN", { style: "currency", currency: "VND", maximumFractionDigits: 0 }).format(amount);

  const docTrendCounts = trend.map((d) => d.count);
  const entryTrendCounts = (entryStats.daily_trend ?? []).map((d) => d.count);

  const stats = [
    {
      title: "Tổng chứng từ",
      value: String(totalDocs),
      icon: FileText,
      description: "Tháng này",
      href: undefined,
      sparkColor: "#6366f1",
      sparkData: docTrendCounts,
    },
    {
      title: "Đã xử lý",
      value: String(bookedDocs),
      icon: CheckCircle,
      description: "Đã định khoản",
      href: undefined,
      sparkColor: "#22c55e",
      sparkData: docTrendCounts,
    },
    {
      title: "Chờ duyệt",
      value: String(pendingEntries),
      icon: Clock,
      description: "Cần review",
      href: "/entries?status=pending",
      sparkColor: "#f59e0b",
      sparkData: entryTrendCounts,
    },
    {
      title: "Lỗi",
      value: String(errorDocs),
      icon: AlertCircle,
      description: "Cần xử lý",
      href: undefined,
      sparkColor: "#ef4444",
      sparkData: docTrendCounts,
    },
    {
      title: "Tổng tiền đã duyệt",
      value: formatVND(approvedAmount),
      icon: CircleDot,
      description: "Tổng amount approved",
      href: "/entries?status=approved",
      sparkColor: "#0ea5e9",
      sparkData: entryTrendCounts,
    },
  ];

  const recentDocs = docs.slice(0, 5);

  const docStatusBadge = (status: Document["status"]) => {
    const map: Record<string, { label: string; cls: string }> = {
      booked:     { label: "Hoàn thành", cls: "bg-green-100 text-green-800" },
      processing: { label: "Đang xử lý", cls: "bg-yellow-100 text-yellow-800" },
      error:      { label: "Lỗi",        cls: "bg-red-100 text-red-800" },
      uploaded:   { label: "Mới tải",    cls: "bg-gray-100 text-gray-700" },
    };
    const s = map[status] ?? { label: status, cls: "bg-gray-100 text-gray-700" };
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium ${s.cls}`}>
        {s.label}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Dashboard</h1>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => {
          const card = (
            <Card key={stat.title} className={stat.href ? "hover:shadow-md transition-shadow cursor-pointer" : undefined}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
                <stat.icon className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="flex items-end justify-between gap-2">
                  <div>
                    <div className="text-2xl font-bold">{stat.value}</div>
                    <CardDescription>{stat.description}</CardDescription>
                  </div>
                  {(stat.sparkData ?? []).length >= 2 && (
                    <Sparkline data={stat.sparkData!} color={stat.sparkColor} />
                  )}
                </div>
              </CardContent>
            </Card>
          );
          return stat.href ? (
            <Link key={stat.title} href={stat.href}>{card}</Link>
          ) : card;
        })}
      </div>

      {/* Trend chart */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base font-semibold">Xu hướng chứng từ (7 ngày)</CardTitle>
            <div className="flex gap-1">
              {([
                { label: "Tất cả", value: undefined },
                { label: "Hóa đơn", value: "invoice" },
                { label: "Phiếu thu/chi", value: "receipt" },
                { label: "Ngân hàng", value: "bank_statement" },
              ] as { label: string; value: string | undefined }[]).map((f) => (
                <button
                  key={f.label}
                  onClick={() => setChartType(f.value)}
                  className={`px-2 py-1 text-xs rounded transition-colors ${
                    chartType === f.value
                      ? "bg-gray-900 text-white"
                      : "bg-gray-100 text-gray-600 hover:bg-gray-200"
                  }`}
                >
                  {f.label}
                </button>
              ))}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {trend.length === 0 ? (
            <p className="text-sm text-gray-400">Chưa có dữ liệu</p>
          ) : (
            <TrendChart data={trend} />
          )}
        </CardContent>
      </Card>

      {/* Recent documents */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base font-semibold">
            <CircleDot className="h-4 w-4 text-muted-foreground" />
            Chứng từ gần đây
          </CardTitle>
        </CardHeader>
        <CardContent>
          {recentDocs.length === 0 ? (
            <p className="text-sm text-gray-500">Chưa có chứng từ nào</p>
          ) : (
            <ul className="divide-y divide-gray-100">
              {recentDocs.map((doc) => (
                <li key={doc.id} className="flex items-center justify-between py-2.5">
                  <div className="flex items-center gap-3 min-w-0">
                    <FileText className="h-4 w-4 shrink-0 text-gray-400" />
                    <span className="text-sm truncate max-w-xs" title={doc.file_name}>
                      {doc.file_name}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 shrink-0 ml-4">
                    {docStatusBadge(doc.status)}
                    <span className="text-xs text-gray-400">
                      {doc.created_at ? new Date(doc.created_at).toLocaleDateString("vi-VN") : ""}
                    </span>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
