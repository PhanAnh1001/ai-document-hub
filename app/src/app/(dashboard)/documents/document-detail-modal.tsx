"use client";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import type { Document } from "@/lib/api";

const DOC_TYPE_LABELS: Record<string, string> = {
  invoice: "Hóa đơn",
  receipt: "Phiếu thu/chi",
  bank_statement: "Sao kê ngân hàng",
  other: "Khác",
};

const STATUS_LABELS: Record<string, string> = {
  uploaded:   "Mới tải lên",
  processing: "Đang xử lý",
  extracted:  "Đã trích xuất",
  booked:     "Đã hạch toán",
  error:      "Lỗi",
};

const STATUS_STYLES: Record<string, string> = {
  uploaded:   "bg-gray-100 text-gray-700",
  processing: "bg-yellow-100 text-yellow-800",
  extracted:  "bg-blue-100 text-blue-800",
  booked:     "bg-green-100 text-green-800",
  error:      "bg-red-100 text-red-800",
};

interface Props {
  doc: Document;
  open: boolean;
  onClose: () => void;
}

export default function DocumentDetailModal({ doc, open, onClose }: Props) {
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Chi tiết chứng từ</DialogTitle>
        </DialogHeader>

        <div className="space-y-3 text-sm">
          <Row label="Tên file" value={doc.file_name} mono />
          <Row label="Loại" value={DOC_TYPE_LABELS[doc.document_type] ?? doc.document_type} />
          <Row
            label="Trạng thái"
            value={
              <span className={`px-2 py-0.5 rounded text-xs font-medium ${STATUS_STYLES[doc.status] ?? ""}`}>
                {STATUS_LABELS[doc.status] ?? doc.status}
              </span>
            }
          />
          <Row
            label="Ngày tải lên"
            value={new Date(doc.created_at).toLocaleDateString("vi-VN")}
          />
        </div>

        <div className="mt-4 flex justify-end">
          <Button variant="outline" onClick={onClose}>
            Đóng
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function Row({ label, value, mono }: { label: string; value: React.ReactNode; mono?: boolean }) {
  return (
    <div className="flex items-start gap-3">
      <span className="w-32 shrink-0 text-gray-500">{label}</span>
      <span className={mono ? "font-mono break-all" : ""}>{value}</span>
    </div>
  );
}
