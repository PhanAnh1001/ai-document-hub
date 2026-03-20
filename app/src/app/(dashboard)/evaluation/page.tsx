"use client";

import { useCallback, useEffect, useState } from "react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from "recharts";
import { RefreshCw, PlayCircle, TrendingUp, FileText, Target, CheckCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getEvalStats, runEvaluation, type EvalStats } from "@/lib/api";

const METRIC_CONFIG: Array<{
  key: keyof Omit<EvalStats, "total_documents">;
  label: string;
  color: string;
  icon: React.ElementType;
  description: string;
}> = [
  {
    key: "faithfulness",
    label: "Faithfulness",
    color: "#6366f1",
    icon: CheckCircle,
    description: "Câu trả lời trung thực với nguồn",
  },
  {
    key: "answer_relevancy",
    label: "Answer Relevancy",
    color: "#22c55e",
    icon: Target,
    description: "Độ liên quan của câu trả lời",
  },
  {
    key: "context_precision",
    label: "Context Precision",
    color: "#f59e0b",
    icon: TrendingUp,
    description: "Độ chính xác của ngữ cảnh",
  },
];

function ScoreBar({ score }: { score: number }) {
  const pct = Math.round(score * 100);
  const color =
    pct >= 80 ? "bg-green-500" : pct >= 60 ? "bg-yellow-500" : "bg-red-500";
  return (
    <div className="flex items-center gap-2">
      <div className="flex-1 h-2 rounded-full bg-gray-100 overflow-hidden">
        <div
          className={`h-full rounded-full transition-all ${color}`}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-sm font-medium tabular-nums w-10 text-right">
        {pct}%
      </span>
    </div>
  );
}

export default function EvaluationPage() {
  const [stats, setStats] = useState<EvalStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [isRunning, setIsRunning] = useState(false);
  const [runMessage, setRunMessage] = useState<string | null>(null);

  const fetchStats = useCallback(async () => {
    setFetchError(null);
    setIsLoading(true);
    try {
      const data = await getEvalStats();
      setStats(data);
    } catch {
      setFetchError("Không thể tải số liệu đánh giá. Vui lòng thử lại.");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  const handleRunEvaluation = async () => {
    setIsRunning(true);
    setRunMessage(null);
    try {
      const res = await runEvaluation();
      setRunMessage(res.message || "Đang chạy đánh giá...");
      // Refresh stats after a short delay
      setTimeout(() => fetchStats(), 2000);
    } catch (err) {
      setRunMessage(
        "Lỗi: " + (err instanceof Error ? err.message : "Không thể chạy đánh giá")
      );
    } finally {
      setIsRunning(false);
    }
  };

  const chartData =
    stats != null
      ? METRIC_CONFIG.map((m) => ({
          name: m.label,
          value: Math.round(stats[m.key] * 100),
          color: m.color,
        }))
      : [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Đánh giá</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Số liệu đánh giá chất lượng RAG
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={fetchStats}
            disabled={isLoading}
            aria-label="Làm mới"
          >
            <RefreshCw className={`h-4 w-4 mr-1 ${isLoading ? "animate-spin" : ""}`} />
            Làm mới
          </Button>
          <Button
            onClick={handleRunEvaluation}
            disabled={isRunning}
            aria-label="Chạy đánh giá"
          >
            <PlayCircle className="h-4 w-4 mr-2" />
            {isRunning ? "Đang chạy..." : "Chạy đánh giá"}
          </Button>
        </div>
      </div>

      {runMessage && (
        <div
          className={`rounded-md px-4 py-3 text-sm ${
            runMessage.startsWith("Lỗi")
              ? "bg-red-50 text-red-700"
              : "bg-green-50 text-green-700"
          }`}
        >
          {runMessage}
        </div>
      )}

      {fetchError && (
        <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
          {fetchError}
        </div>
      )}

      {/* Stat cards */}
      <div
        className="grid gap-4 md:grid-cols-2 lg:grid-cols-4"
        data-testid="stat-cards"
      >
        {/* Total documents */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Tổng tài liệu</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {isLoading ? "—" : (stats?.total_documents ?? 0)}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Đã được xử lý
            </p>
          </CardContent>
        </Card>

        {METRIC_CONFIG.map((m) => (
          <Card key={m.key}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{m.label}</CardTitle>
              <m.icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {isLoading
                  ? "—"
                  : stats != null
                  ? `${Math.round(stats[m.key] * 100)}%`
                  : "—"}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                {m.description}
              </p>
              {!isLoading && stats != null && (
                <div className="mt-2">
                  <ScoreBar score={stats[m.key]} />
                </div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Bar chart */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base font-semibold">
            So sánh chỉ số đánh giá
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-sm text-muted-foreground">Đang tải...</p>
          ) : stats == null ? (
            <p className="text-sm text-muted-foreground">
              Chưa có dữ liệu đánh giá. Chạy đánh giá để bắt đầu.
            </p>
          ) : (
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={chartData} margin={{ top: 16, right: 16, left: 0, bottom: 0 }}>
                <XAxis dataKey="name" tick={{ fontSize: 12 }} />
                <YAxis domain={[0, 100]} unit="%" tick={{ fontSize: 12 }} />
                <Tooltip formatter={(val) => [`${val}%`, "Điểm"]} />
                <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                  {chartData.map((entry, i) => (
                    <Cell key={i} fill={entry.color} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          )}
        </CardContent>
      </Card>

      {/* Metrics detail table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base font-semibold">
            Chi tiết chỉ số
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-sm text-muted-foreground">Đang tải...</p>
          ) : stats == null ? (
            <p className="text-sm text-muted-foreground">Chưa có dữ liệu</p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-2 font-medium">Chỉ số</th>
                  <th className="pb-2 font-medium">Mô tả</th>
                  <th className="pb-2 font-medium">Điểm</th>
                  <th className="pb-2 font-medium w-40">Biểu đồ</th>
                </tr>
              </thead>
              <tbody>
                {METRIC_CONFIG.map((m) => (
                  <tr key={m.key} className="border-b last:border-0">
                    <td className="py-3 font-medium">{m.label}</td>
                    <td className="py-3 text-muted-foreground">{m.description}</td>
                    <td className="py-3 font-bold">
                      {Math.round(stats[m.key] * 100)}%
                    </td>
                    <td className="py-3 w-40">
                      <ScoreBar score={stats[m.key]} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
