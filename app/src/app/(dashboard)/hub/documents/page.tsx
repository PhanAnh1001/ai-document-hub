"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Upload, RefreshCw, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  listHubDocuments,
  uploadHubDocument,
  deleteHubDocument,
  type HubDocument,
} from "@/lib/api";

const DOC_TYPE_LABELS: Record<HubDocument["doc_type"], string> = {
  invoice: "Hóa đơn",
  contract: "Hợp đồng",
  cv: "CV",
  report: "Báo cáo",
  other: "Khác",
};

// Badge color per status
const STATUS_BADGE: Record<
  HubDocument["status"],
  { label: string; cls: string }
> = {
  uploaded:       { label: "Đã tải lên",     cls: "bg-gray-100 text-gray-700" },
  ocr_processing: { label: "Đang OCR",        cls: "bg-blue-100 text-blue-700 animate-pulse" },
  ocr_done:       { label: "OCR xong",        cls: "bg-cyan-100 text-cyan-700" },
  extracting:     { label: "Đang trích xuất", cls: "bg-yellow-100 text-yellow-700" },
  extracted:      { label: "Đã trích xuất",   cls: "bg-green-100 text-green-700" },
  indexing:       { label: "Đang index",      cls: "bg-indigo-100 text-indigo-700" },
  indexed:        { label: "Đã index",        cls: "bg-purple-100 text-purple-700" },
  failed:         { label: "Lỗi",             cls: "bg-red-100 text-red-700" },
};

function StatusBadge({ status }: { status: HubDocument["status"] }) {
  const s = STATUS_BADGE[status] ?? { label: status, cls: "bg-gray-100 text-gray-700" };
  return (
    <span
      data-status={status}
      className={`inline-block rounded px-2 py-0.5 text-xs font-medium ${s.cls}`}
    >
      {s.label}
    </span>
  );
}

export default function HubDocumentsPage() {
  const [documents, setDocuments] = useState<HubDocument[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState<string | null>(null);

  // Upload dialog state
  const [uploadOpen, setUploadOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [docType, setDocType] = useState<HubDocument["doc_type"]>("other");
  const [isUploading, setIsUploading] = useState(false);
  const [uploadError, setUploadError] = useState("");
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Detail modal state
  const [detailDoc, setDetailDoc] = useState<HubDocument | null>(null);

  // Delete confirmation
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchDocuments = useCallback(async () => {
    setFetchError(null);
    setIsLoading(true);
    try {
      const docs = await listHubDocuments();
      setDocuments(docs);
    } catch {
      setFetchError("Không thể tải danh sách tài liệu. Vui lòng thử lại.");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDocuments();
  }, [fetchDocuments]);

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(false);
    const file = e.dataTransfer.files?.[0];
    if (file) {
      setSelectedFile(file);
      setUploadOpen(true);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) return;
    setIsUploading(true);
    setUploadError("");
    try {
      const doc = await uploadHubDocument(selectedFile, docType);
      setDocuments((prev) => [doc, ...prev]);
      setUploadOpen(false);
      setSelectedFile(null);
      if (fileInputRef.current) fileInputRef.current.value = "";
    } catch (err) {
      setUploadError(
        "Upload thất bại: " +
          (err instanceof Error ? err.message : "Lỗi không xác định")
      );
    } finally {
      setIsUploading(false);
    }
  };

  const handleDelete = async (id: string) => {
    setDeletingId(id);
    try {
      await deleteHubDocument(id);
      setDocuments((prev) => prev.filter((d) => d.id !== id));
    } catch {
      // Silently fail — keep document in list
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Tài liệu</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Upload tài liệu để OCR, trích xuất và lập chỉ mục RAG
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={fetchDocuments}
            disabled={isLoading}
            aria-label="Làm mới"
          >
            <RefreshCw className={`h-4 w-4 mr-1 ${isLoading ? "animate-spin" : ""}`} />
            Làm mới
          </Button>
          <Button onClick={() => setUploadOpen(true)}>
            <Upload className="mr-2 h-4 w-4" />
            Upload tài liệu
          </Button>
        </div>
      </div>

      {/* Drag-and-drop zone */}
      <div
        onDragOver={(e) => { e.preventDefault(); setIsDragging(true); }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={handleDrop}
        className={`rounded-lg border-2 border-dashed p-6 text-center transition-colors ${
          isDragging
            ? "border-indigo-400 bg-indigo-50"
            : "border-gray-200 hover:border-gray-300"
        }`}
      >
        <Upload className="mx-auto h-8 w-8 text-gray-400 mb-2" />
        <p className="text-sm text-muted-foreground">
          Kéo thả file vào đây, hoặc{" "}
          <button
            className="text-indigo-600 underline"
            onClick={() => setUploadOpen(true)}
          >
            chọn file
          </button>
        </p>
        <p className="text-xs text-gray-400 mt-1">PDF, ảnh (JPG/PNG), TXT</p>
      </div>

      {/* Upload dialog */}
      <Dialog open={uploadOpen} onOpenChange={setUploadOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Upload tài liệu</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <label htmlFor="hub_doc_type" className="text-sm font-medium">
                Loại tài liệu
              </label>
              <select
                id="hub_doc_type"
                className="w-full rounded-md border px-3 py-2 text-sm"
                value={docType}
                onChange={(e) =>
                  setDocType(e.target.value as HubDocument["doc_type"])
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
              <label htmlFor="hub_file_input" className="text-sm font-medium">
                Chọn file
              </label>
              <input
                id="hub_file_input"
                ref={fileInputRef}
                type="file"
                accept=".pdf,.jpg,.jpeg,.png,.txt"
                className="w-full text-sm"
                onChange={(e) => setSelectedFile(e.target.files?.[0] ?? null)}
              />
              {selectedFile && (
                <div className="flex items-center gap-2 rounded-md bg-gray-50 px-3 py-2 text-sm border border-gray-200">
                  <span className="font-medium truncate flex-1">
                    {selectedFile.name}
                  </span>
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

      {/* Error state */}
      {fetchError && (
        <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
          {fetchError}
        </div>
      )}

      {/* Document list */}
      <Card>
        <CardHeader>
          <CardTitle>Danh sách tài liệu</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-muted-foreground">Đang tải...</p>
          ) : documents.length === 0 ? (
            <p className="text-muted-foreground">
              Chưa có tài liệu nào. Upload tài liệu để bắt đầu.
            </p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-2 font-medium">Tên file</th>
                  <th className="pb-2 font-medium">Loại</th>
                  <th className="pb-2 font-medium">Trạng thái</th>
                  <th className="pb-2 font-medium">Ngày tải</th>
                  <th className="pb-2 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {documents.map((doc) => (
                  <tr
                    key={doc.id}
                    className="border-b hover:bg-muted/30 cursor-pointer"
                    onClick={() => setDetailDoc(doc)}
                  >
                    <td className="py-2 pr-4 max-w-xs truncate" title={doc.original_filename}>
                      {doc.original_filename}
                    </td>
                    <td className="py-2 pr-4">
                      {DOC_TYPE_LABELS[doc.doc_type] ?? doc.doc_type}
                    </td>
                    <td className="py-2 pr-4">
                      <StatusBadge status={doc.status} />
                    </td>
                    <td className="py-2 pr-4 whitespace-nowrap text-xs text-gray-500">
                      {new Date(doc.created_at).toLocaleDateString("vi-VN")}
                    </td>
                    <td className="py-2">
                      <button
                        aria-label="Xóa tài liệu"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDelete(doc.id);
                        }}
                        disabled={deletingId === doc.id}
                        className="p-1 rounded hover:bg-red-50 text-gray-400 hover:text-red-500 disabled:opacity-40"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </CardContent>
      </Card>

      {/* Detail modal */}
      <Dialog open={!!detailDoc} onOpenChange={(open) => !open && setDetailDoc(null)}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <div className="flex items-start justify-between gap-4">
              <DialogTitle className="truncate">
                {detailDoc?.original_filename}
              </DialogTitle>
              <button
                onClick={() => setDetailDoc(null)}
                className="shrink-0 rounded p-1 hover:bg-gray-100"
                aria-label="Đóng"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </DialogHeader>
          {detailDoc && (
            <div className="space-y-4 py-2 text-sm">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <p className="text-xs text-muted-foreground">Loại</p>
                  <p className="font-medium">{DOC_TYPE_LABELS[detailDoc.doc_type]}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Trạng thái</p>
                  <StatusBadge status={detailDoc.status} />
                </div>
                {detailDoc.ocr_confidence != null && (
                  <div>
                    <p className="text-xs text-muted-foreground">OCR Confidence</p>
                    <p className="font-medium">
                      {(detailDoc.ocr_confidence * 100).toFixed(1)}%
                    </p>
                  </div>
                )}
                <div>
                  <p className="text-xs text-muted-foreground">Ngày tải</p>
                  <p className="font-medium">
                    {new Date(detailDoc.created_at).toLocaleString("vi-VN")}
                  </p>
                </div>
              </div>

              {detailDoc.ocr_text && (
                <div>
                  <p className="text-xs text-muted-foreground mb-1">Văn bản OCR</p>
                  <pre className="rounded-md bg-gray-50 p-3 text-xs whitespace-pre-wrap max-h-40 overflow-y-auto border">
                    {detailDoc.ocr_text}
                  </pre>
                </div>
              )}

              {detailDoc.extracted_data && (
                <div>
                  <p className="text-xs text-muted-foreground mb-1">Dữ liệu trích xuất</p>
                  <pre className="rounded-md bg-gray-50 p-3 text-xs whitespace-pre-wrap max-h-48 overflow-y-auto border">
                    {JSON.stringify(detailDoc.extracted_data, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
