"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { useAuthStore } from "@/lib/auth-store";
import { listEntries, approveEntry, rejectEntry, syncEntry, exportEntriesCSV, updateEntry, bulkApproveEntries, AccountingEntry } from "@/lib/api";
import EntryDetailModal from "./entry-detail-modal";
import Pagination from "@/components/pagination";
import { toast } from "@/hooks/use-toast";

type TabFilter = "pending" | "approved" | "all";

export default function EntriesPage() {
  const token = useAuthStore((s) => (s as { token: string }).token);
  const [entries, setEntries] = useState<AccountingEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [tab, setTab] = useState<TabFilter>("pending");
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [detailEntry, setDetailEntry] = useState<AccountingEntry | null>(null);
  const [page, setPage] = useState(1);
  const [exporting, setExporting] = useState(false);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [bulkLoading, setBulkLoading] = useState(false);
  const [exportFrom, setExportFrom] = useState("");
  const [exportTo, setExportTo] = useState("");
  const [rejectDialogId, setRejectDialogId] = useState<string | null>(null);
  const [rejectReason, setRejectReason] = useState("");
  const [sortBy, setSortBy] = useState<string>("");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");
  const PAGE_SIZE = 20;

  const fetchEntries = useCallback(() => {
    if (!token) return;
    setLoading(true);
    setFetchError(null);
    const status = tab === "all" ? undefined : tab;
    listEntries(token, {
      status,
      q: search || undefined,
      limit: PAGE_SIZE,
      offset: (page - 1) * PAGE_SIZE,
      date_from: exportFrom || undefined,
      date_to: exportTo || undefined,
      sort_by: sortBy || undefined,
      sort_dir: sortBy ? sortDir : undefined,
    })
      .then((res) => {
        setEntries(res.entries);
        setTotal(res.total);
      })
      .catch(() => setFetchError("Không thể tải dữ liệu. Vui lòng thử lại."))
      .finally(() => setLoading(false));
  }, [token, tab, page, search, exportFrom, exportTo, sortBy, sortDir]);

  useEffect(() => { fetchEntries(); }, [fetchEntries]);

  async function handleApprove(id: string) {
    if (!token) return;
    setActionLoading(id);
    try {
      const res = await approveEntry(id, token);
      setEntries((prev) => prev.map((e) => (e.id === id ? res.entry : e)));
      toast({ title: "Đã duyệt định khoản", variant: "success" });
    } catch {
      toast({ title: "Không thể duyệt", variant: "error" });
    } finally {
      setActionLoading(null);
    }
  }

  async function handleReject(id: string, reason?: string) {
    if (!token) return;
    setActionLoading(id);
    try {
      const res = await rejectEntry(id, token, reason || undefined);
      setEntries((prev) => prev.map((e) => (e.id === id ? res.entry : e)));
      toast({ title: "Đã từ chối định khoản", variant: "info" });
    } catch {
      toast({ title: "Không thể từ chối", variant: "error" });
    } finally {
      setActionLoading(null);
      setRejectDialogId(null);
      setRejectReason("");
    }
  }

  async function handleSync(id: string) {
    if (!token) return;
    setActionLoading(id);
    try {
      const res = await syncEntry(id, token);
      setEntries((prev) => prev.map((e) => (e.id === id ? res.entry : e)));
      toast({ title: "Đã đồng bộ lên MISA", variant: "success" });
    } catch {
      toast({ title: "Không thể sync MISA", variant: "error" });
    } finally {
      setActionLoading(null);
    }
  }

  async function handleUpdate(id: string, fields: { description: string; debit_account: string; credit_account: string; amount: number }) {
    if (!token) return;
    const res = await updateEntry(id, token, fields);
    setEntries((prev) => prev.map((e) => (e.id === id ? res.entry : e)));
    setDetailEntry(res.entry);
    toast({ title: "Đã lưu thay đổi", variant: "success" });
  }

  async function handleBulkApprove() {
    if (!token || selectedIds.size === 0) return;
    setBulkLoading(true);
    try {
      await bulkApproveEntries(Array.from(selectedIds), token);
      toast({ title: `Đã duyệt ${selectedIds.size} định khoản`, variant: "success" });
      setSelectedIds(new Set());
      fetchEntries();
    } catch {
      toast({ title: "Không thể duyệt nhiều", variant: "error" });
    } finally {
      setBulkLoading(false);
    }
  }

  const pendingEntries = entries.filter((e) => e.status === "pending");
  const allPendingSelected = pendingEntries.length > 0 && pendingEntries.every((e) => selectedIds.has(e.id));

  function toggleSelectAll() {
    if (allPendingSelected) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(pendingEntries.map((e) => e.id)));
    }
  }

  function toggleSelect(id: string) {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function handleSort(col: string) {
    if (sortBy === col) {
      setSortDir((d) => (d === "desc" ? "asc" : "desc"));
    } else {
      setSortBy(col);
      setSortDir("desc");
    }
    setPage(1);
  }

  const sortIndicator = (col: string) => {
    if (sortBy !== col) return null;
    return sortDir === "desc" ? " ↓" : " ↑";
  };

  const formatAmount = (amount: number) =>
    new Intl.NumberFormat("vi-VN", { style: "currency", currency: "VND" }).format(amount);

  const statusBadge = (status: AccountingEntry["status"]) => {
    const styles: Record<string, string> = {
      pending: "bg-yellow-100 text-yellow-800",
      approved: "bg-green-100 text-green-800",
      rejected: "bg-red-100 text-red-800",
      draft: "bg-gray-100 text-gray-700",
      synced: "bg-blue-100 text-blue-800",
    };
    const labels: Record<string, string> = {
      pending: "Chờ duyệt",
      approved: "Đã duyệt",
      rejected: "Từ chối",
      draft: "Nháp",
      synced: "Đã đồng bộ",
    };
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium ${styles[status] ?? ""}`}>
        {labels[status] ?? status}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Định khoản</h1>
        <div className="flex items-center gap-3">
          <span className="text-sm text-gray-500">{total} bút toán</span>
          {selectedIds.size > 0 && (
            <button
              onClick={handleBulkApprove}
              disabled={bulkLoading}
              className="px-3 py-1.5 bg-green-600 text-white text-sm rounded hover:bg-green-700 disabled:opacity-50"
            >
              {bulkLoading ? "Đang duyệt..." : `Duyệt nhiều (${selectedIds.size})`}
            </button>
          )}
          <button
            onClick={fetchEntries}
            disabled={loading}
            className="px-3 py-1.5 bg-white border border-gray-300 text-gray-700 text-sm rounded hover:bg-gray-50 disabled:opacity-50"
          >
            Làm mới
          </button>
          <label htmlFor="export-from" className="text-xs text-gray-500">Từ ngày</label>
          <input
            id="export-from"
            type="date"
            value={exportFrom}
            onChange={(e) => setExportFrom(e.target.value)}
            className="border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-gray-400"
            aria-label="Từ ngày"
          />
          <label htmlFor="export-to" className="text-xs text-gray-500">Đến ngày</label>
          <input
            id="export-to"
            type="date"
            value={exportTo}
            onChange={(e) => setExportTo(e.target.value)}
            className="border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-gray-400"
            aria-label="Đến ngày"
          />
          <button
            onClick={async () => {
              if (!token || exporting) return;
              setExporting(true);
              try {
                const status = tab === "all" ? undefined : tab;
                const dateRange = (exportFrom || exportTo)
                  ? { from: exportFrom || undefined, to: exportTo || undefined }
                  : undefined;
                const count = await exportEntriesCSV(token, status, dateRange, search || undefined);
                toast({ title: `Đã xuất ${count} bút toán`, variant: "success" });
              } catch (err) {
                console.error(err);
                toast({ title: "Không thể xuất CSV", variant: "error" });
              } finally {
                setExporting(false);
              }
            }}
            disabled={exporting}
            className="px-3 py-1.5 bg-gray-800 text-white text-sm rounded hover:bg-gray-900 disabled:opacity-50"
          >
            {exporting ? "Đang xuất..." : "Xuất CSV"}
          </button>
        </div>
      </div>

      {/* Tab filters + search */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex gap-2">
          {(["pending", "approved", "all"] as TabFilter[]).map((t) => {
            const labels: Record<TabFilter, string> = {
              pending: "Chờ duyệt",
              approved: "Đã duyệt",
              all: "Tất cả",
            };
            return (
              <button
                key={t}
                onClick={() => { setTab(t); setPage(1); }}
                className={`px-4 py-1.5 rounded-full text-sm font-medium transition-colors
                  ${tab === t
                    ? "bg-gray-900 text-white"
                    : "bg-gray-100 text-gray-600 hover:bg-gray-200"
                  }`}
              >
                {labels[t]}
              </button>
            );
          })}
        </div>
        <input
          type="search"
          placeholder="Tìm kiếm mô tả..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1); }}
          className="ml-auto rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-400 w-60"
        />
      </div>

      {/* Error banner */}
      {fetchError && (
        <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
          {fetchError}
        </div>
      )}

      {/* Table */}
      {loading ? (
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-14 bg-gray-100 animate-pulse rounded" />
          ))}
        </div>
      ) : entries.length === 0 ? (
        <div className="text-center py-16 text-gray-500">
          Không có định khoản nào
        </div>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 w-8">
                  <input
                    type="checkbox"
                    checked={allPendingSelected}
                    onChange={toggleSelectAll}
                    disabled={pendingEntries.length === 0}
                    className="rounded border-gray-300"
                    aria-label="Chọn tất cả"
                  />
                </th>
                <th className="px-4 py-3 text-left font-semibold text-gray-600">
                  <button onClick={() => handleSort("entry_date")} className="hover:text-gray-900">
                    Ngày{sortIndicator("entry_date")}
                  </button>
                </th>
                <th className="px-4 py-3 text-left font-semibold text-gray-600">Mô tả</th>
                <th className="px-4 py-3 text-left font-semibold text-gray-600">Chứng từ</th>
                <th className="px-4 py-3 text-center font-semibold text-gray-600">TK Nợ</th>
                <th className="px-4 py-3 text-center font-semibold text-gray-600">TK Có</th>
                <th className="px-4 py-3 text-right font-semibold text-gray-600">
                  <button onClick={() => handleSort("amount")} className="hover:text-gray-900">
                    Số tiền{sortIndicator("amount")}
                  </button>
                </th>
                <th className="px-4 py-3 text-center font-semibold text-gray-600">AI</th>
                <th className="px-4 py-3 text-center font-semibold text-gray-600">Trạng thái</th>
                <th className="px-4 py-3 text-right font-semibold text-gray-600">Thao tác</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-100">
              {entries.map((entry) => (
                <tr key={entry.id} className={`hover:bg-gray-50 ${selectedIds.has(entry.id) ? "bg-green-50" : ""}`}>
                  <td className="px-4 py-3 w-8">
                    {entry.status === "pending" && (
                      <input
                        type="checkbox"
                        checked={selectedIds.has(entry.id)}
                        onChange={() => toggleSelect(entry.id)}
                        className="rounded border-gray-300"
                        aria-label={`Chọn ${entry.id}`}
                      />
                    )}
                  </td>
                  <td className="px-4 py-3 whitespace-nowrap text-gray-500">
                    {entry.entry_date?.slice(0, 10) ?? "—"}
                  </td>
                  <td className="px-4 py-3 max-w-xs truncate" title={entry.description}>
                    {entry.description}
                  </td>
                  <td className="px-4 py-3 max-w-[140px]">
                    {entry.document_id ? (
                      <Link
                        href={`/documents?id=${entry.document_id}`}
                        className="text-blue-600 hover:underline text-xs truncate block"
                        title={entry.document_name ?? entry.document_id}
                      >
                        {entry.document_name ?? entry.document_id.slice(0, 8) + "…"}
                      </Link>
                    ) : (
                      <span className="text-gray-400 text-xs">—</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-center font-mono font-semibold">
                    {entry.debit_account}
                  </td>
                  <td className="px-4 py-3 text-center font-mono font-semibold">
                    {entry.credit_account}
                  </td>
                  <td className="px-4 py-3 text-right font-medium">
                    {formatAmount(entry.amount)}
                  </td>
                  <td className="px-4 py-3 text-center">
                    {entry.ai_confidence != null ? (
                      <span
                        className={`text-xs font-medium ${
                          entry.ai_confidence >= 0.85
                            ? "text-green-600"
                            : entry.ai_confidence >= 0.70
                            ? "text-yellow-600"
                            : "text-red-500"
                        }`}
                      >
                        {Math.round(entry.ai_confidence * 100)}%
                      </span>
                    ) : (
                      "—"
                    )}
                  </td>
                  <td className="px-4 py-3 text-center">
                    <span title={entry.reject_reason ?? undefined}>
                      {statusBadge(entry.status)}
                      {entry.reject_reason && (
                        <span className="ml-1 text-gray-400 text-xs cursor-help" title={entry.reject_reason}>ⓘ</span>
                      )}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex gap-2 justify-end">
                      <button
                        onClick={() => setDetailEntry(entry)}
                        className="px-3 py-1 bg-gray-100 text-gray-700 text-xs rounded hover:bg-gray-200"
                      >
                        Xem
                      </button>
                      {entry.status === "pending" && (
                        <>
                          <button
                            onClick={() => handleApprove(entry.id)}
                            disabled={actionLoading === entry.id}
                            className="px-3 py-1 bg-green-600 text-white text-xs rounded hover:bg-green-700 disabled:opacity-50"
                          >
                            Duyệt
                          </button>
                          <button
                            onClick={() => { setRejectDialogId(entry.id); setRejectReason(""); }}
                            disabled={actionLoading === entry.id}
                            className="px-3 py-1 bg-red-500 text-white text-xs rounded hover:bg-red-600 disabled:opacity-50"
                          >
                            Từ chối
                          </button>
                        </>
                      )}
                      {entry.status === "approved" && (
                        <button
                          onClick={() => handleSync(entry.id)}
                          disabled={actionLoading === entry.id}
                          className="px-3 py-1 bg-blue-600 text-white text-xs rounded hover:bg-blue-700 disabled:opacity-50"
                        >
                          Sync MISA
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Pagination
        page={page}
        total={total}
        pageSize={PAGE_SIZE}
        onPage={setPage}
      />

      {/* Reject reason dialog */}
      {rejectDialogId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="bg-white rounded-lg shadow-xl p-6 w-full max-w-sm space-y-4">
            <h3 className="font-semibold text-base">Từ chối định khoản</h3>
            <textarea
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm resize-none focus:outline-none focus:ring-2 focus:ring-red-300"
              rows={3}
              placeholder="Lý do từ chối (tùy chọn)"
              value={rejectReason}
              onChange={(e) => setRejectReason(e.target.value)}
            />
            <div className="flex justify-end gap-2">
              <button
                onClick={() => { setRejectDialogId(null); setRejectReason(""); }}
                className="px-3 py-1.5 text-sm border border-gray-300 rounded hover:bg-gray-50"
              >
                Hủy
              </button>
              <button
                onClick={() => handleReject(rejectDialogId, rejectReason)}
                disabled={actionLoading === rejectDialogId}
                className="px-3 py-1.5 text-sm bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50"
              >
                {actionLoading === rejectDialogId ? "Đang xử lý..." : "Xác nhận từ chối"}
              </button>
            </div>
          </div>
        </div>
      )}

      {detailEntry && (
        <EntryDetailModal
          entry={detailEntry}
          open={!!detailEntry}
          onClose={() => setDetailEntry(null)}
          onSave={(fields) => handleUpdate(detailEntry.id, fields)}
        />
      )}
    </div>
  );
}
