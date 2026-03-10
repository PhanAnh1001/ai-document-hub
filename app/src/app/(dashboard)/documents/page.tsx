"use client";

import { useEffect, useRef, useState } from "react";
import { Upload } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { listDocuments, uploadDocument, retryDocument, type Document } from "@/lib/api";
import { useAuthStore } from "@/lib/auth-store";
import Pagination from "@/components/pagination";
import DocumentDetailModal from "./document-detail-modal";

const DOC_TYPE_LABELS: Record<string, string> = {
  invoice: "Hóa đơn",
  receipt: "Phiếu thu/chi",
  bank_statement: "Sao kê ngân hàng",
  other: "Khác",
};

const STATUS_LABELS: Record<string, string> = {
  uploaded: "Đã upload",
  processing: "Đang xử lý",
  extracted: "Đã trích xuất",
  booked: "Đã hạch toán",
  error: "Lỗi",
};

export default function DocumentsPage() {
  const token = useAuthStore((s) => s.token);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [detailDoc, setDetailDoc] = useState<Document | null>(null);
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState("");
  const [search, setSearch] = useState("");
  const PAGE_SIZE = 20;

  const [retryingId, setRetryingId] = useState<string | null>(null);

  // Upload form state
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [docType, setDocType] = useState<Document["document_type"]>("other");
  const [isUploading, setIsUploading] = useState(false);
  const [uploadError, setUploadError] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchDocuments = async () => {
    if (!token) return;
    setFetchError(null);
    try {
      const res = await listDocuments(token, {
        status: statusFilter || undefined,
        q: search || undefined,
        limit: PAGE_SIZE,
        offset: (page - 1) * PAGE_SIZE,
      });
      setDocuments(res.documents ?? []);
      setTotal(res.total ?? 0);
    } catch {
      setFetchError("Không thể tải danh sách chứng từ. Vui lòng thử lại.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchDocuments();
  }, [token, page, statusFilter, search]);

  const handleRetry = async (id: string) => {
    if (!token) return;
    setRetryingId(id);
    try {
      await retryDocument(id, token);
      fetchDocuments();
    } finally {
      setRetryingId(null);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile || !token) return;
    setIsUploading(true);
    setUploadError("");
    try {
      const doc = await uploadDocument(selectedFile, docType, token);
      setDocuments((prev) => [doc, ...prev]);
      setDialogOpen(false);
      setSelectedFile(null);
      if (fileInputRef.current) fileInputRef.current.value = "";
    } catch (err) {
      setUploadError("Upload thất bại: " + (err instanceof Error ? err.message : "Lỗi không xác định"));
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-bold">Chứng từ</h1>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => fetchDocuments()} disabled={isLoading}>
            Làm mới
          </Button>
          <Button onClick={() => setDialogOpen(true)}>
            <Upload className="mr-2 h-4 w-4" />
            Upload chứng từ
          </Button>
        </div>
      </div>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Upload chứng từ</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <label htmlFor="doc_type" className="text-sm font-medium">
                Loại chứng từ
              </label>
              <select
                id="doc_type"
                className="w-full rounded-md border px-3 py-2 text-sm"
                value={docType}
                onChange={(e) =>
                  setDocType(e.target.value as Document["document_type"])
                }
              >
                {Object.entries(DOC_TYPE_LABELS).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <label htmlFor="file_input" className="text-sm font-medium">
                Chọn file
              </label>
              <input
                id="file_input"
                ref={fileInputRef}
                type="file"
                accept=".pdf,.jpg,.jpeg,.png"
                className="w-full text-sm"
                onChange={(e) => setSelectedFile(e.target.files?.[0] ?? null)}
              />
              {selectedFile && (
                <div className="flex items-center gap-2 rounded-md bg-gray-50 px-3 py-2 text-sm border border-gray-200">
                  <span className="font-medium truncate flex-1">{selectedFile.name}</span>
                  <span className="text-gray-500 shrink-0">
                    {selectedFile.size >= 1024 * 1024
                      ? `${(selectedFile.size / 1024 / 1024).toFixed(1)} MB`
                      : `${(selectedFile.size / 1024).toFixed(1)} KB`}
                  </span>
                </div>
              )}
            </div>
            {uploadError && (
              <p className="text-sm text-destructive">{uploadError}</p>
            )}
            <Button
              className="w-full"
              disabled={isUploading || !selectedFile}
              onClick={handleUpload}
            >
              {isUploading ? "Đang tải lên..." : "Tải lên"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {fetchError && (
        <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
          {fetchError}
        </div>
      )}

      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-center gap-3 justify-between">
            <CardTitle>Danh sách chứng từ</CardTitle>
            <div className="flex items-center gap-2">
              <input
                type="search"
                placeholder="Tìm theo tên file..."
                value={search}
                onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                className="rounded-md border border-input bg-background px-3 py-1.5 text-sm w-48 focus:outline-none focus:ring-2 focus:ring-gray-400"
              />
              <select
                aria-label="Lọc trạng thái"
                value={statusFilter}
                onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
                className="rounded-md border border-input bg-background px-3 py-1.5 text-sm"
              >
                <option value="">Tất cả trạng thái</option>
                <option value="uploaded">Đã upload</option>
                <option value="processing">Đang xử lý</option>
                <option value="extracted">Đã trích xuất</option>
                <option value="booked">Đã hạch toán</option>
                <option value="error">Lỗi</option>
              </select>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-muted-foreground">Đang tải...</p>
          ) : documents.length === 0 ? (
            <p className="text-muted-foreground">
              Chưa có chứng từ nào. Bấm &quot;Upload chứng từ&quot; để bắt đầu.
            </p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-2 font-medium">Tên file</th>
                  <th className="pb-2 font-medium">Loại</th>
                  <th className="pb-2 font-medium">Trạng thái</th>
                  <th className="pb-2 font-medium">Ngày upload</th>
                  <th className="pb-2 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {documents.map((doc) => (
                  <tr key={doc.id} className="border-b">
                    <td className="py-2">
                      <div className="flex items-center gap-2">
                        <span className={`inline-block rounded px-1 py-0.5 text-xs font-bold ${
                          /\.(pdf)$/i.test(doc.file_name)
                            ? "bg-red-100 text-red-700"
                            : /\.(jpe?g|png|gif|webp)$/i.test(doc.file_name)
                            ? "bg-blue-100 text-blue-700"
                            : "bg-gray-100 text-gray-600"
                        }`}>
                          {/\.(pdf)$/i.test(doc.file_name) ? "PDF"
                            : /\.(jpe?g|png|gif|webp)$/i.test(doc.file_name) ? "IMG"
                            : "FILE"}
                        </span>
                        {doc.file_name}
                      </div>
                    </td>
                    <td className="py-2">{DOC_TYPE_LABELS[doc.document_type] ?? doc.document_type}</td>
                    <td className="py-2">{STATUS_LABELS[doc.status] ?? doc.status}</td>
                    <td className="py-2">{new Date(doc.created_at).toLocaleDateString("vi-VN")}</td>
                    <td className="py-2 flex gap-1">
                      <button
                        onClick={() => setDetailDoc(doc)}
                        className="px-2 py-1 text-xs bg-gray-100 rounded hover:bg-gray-200"
                      >
                        Xem
                      </button>
                      {doc.status === "error" && (
                        <button
                          onClick={() => handleRetry(doc.id)}
                          disabled={retryingId === doc.id}
                          className="px-2 py-1 text-xs bg-amber-100 text-amber-800 rounded hover:bg-amber-200 disabled:opacity-50"
                        >
                          {retryingId === doc.id ? "..." : "Thử lại"}
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <Pagination page={page} total={total} pageSize={PAGE_SIZE} onPage={setPage} />
        </CardContent>
      </Card>

      {detailDoc && (
        <DocumentDetailModal
          doc={detailDoc}
          open={!!detailDoc}
          onClose={() => setDetailDoc(null)}
        />
      )}
    </div>
  );
}
