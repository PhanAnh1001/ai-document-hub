"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { AccountingEntry } from "@/lib/api";

type Props = {
  entry: AccountingEntry;
  open: boolean;
  onClose: () => void;
  onSave?: (fields: { description: string; debit_account: string; credit_account: string; amount: number }) => Promise<void>;
};

function formatAmount(amount: number): string {
  return new Intl.NumberFormat("vi-VN").format(amount);
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("vi-VN");
}

const STATUS_LABELS: Record<string, string> = {
  draft: "Nháp",
  pending: "Chờ duyệt",
  approved: "Đã duyệt",
  rejected: "Từ chối",
  synced: "Đã sync",
};

export default function EntryDetailModal({ entry, open, onClose, onSave }: Props) {
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({
    description: entry.description,
    debit_account: entry.debit_account,
    credit_account: entry.credit_account,
    amount: entry.amount,
  });

  function handleEdit() {
    setForm({
      description: entry.description,
      debit_account: entry.debit_account,
      credit_account: entry.credit_account,
      amount: entry.amount,
    });
    setEditing(true);
  }

  async function handleSave() {
    if (!onSave) return;
    setSaving(true);
    try {
      await onSave(form);
      setEditing(false);
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) { setEditing(false); onClose(); } }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {entry.document_name ?? "Chi tiết định khoản"}
          </DialogTitle>
        </DialogHeader>

        {editing ? (
          <div className="space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-2 items-center">
              <label className="text-muted-foreground">Mô tả</label>
              <Input
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              />
              <label className="text-muted-foreground">TK Nợ</label>
              <Input
                value={form.debit_account}
                onChange={(e) => setForm((f) => ({ ...f, debit_account: e.target.value }))}
              />
              <label className="text-muted-foreground">TK Có</label>
              <Input
                value={form.credit_account}
                onChange={(e) => setForm((f) => ({ ...f, credit_account: e.target.value }))}
              />
              <label className="text-muted-foreground">Số tiền</label>
              <Input
                type="number"
                value={form.amount}
                onChange={(e) => setForm((f) => ({ ...f, amount: Number(e.target.value) }))}
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <Button variant="outline" onClick={() => setEditing(false)} disabled={saving}>
                Hủy
              </Button>
              <Button onClick={handleSave} disabled={saving}>
                {saving ? "Đang lưu..." : "Lưu"}
              </Button>
            </div>
          </div>
        ) : (
          <>
            <div className="space-y-3 text-sm">
              <div className="grid grid-cols-2 gap-2">
                <span className="text-muted-foreground">Ngày</span>
                <span>{formatDate(entry.entry_date)}</span>

                <span className="text-muted-foreground">Mô tả</span>
                <span>{entry.description}</span>

                <span className="text-muted-foreground">Tài khoản nợ</span>
                <span className="font-mono font-semibold">{entry.debit_account}</span>

                <span className="text-muted-foreground">Tài khoản có</span>
                <span className="font-mono font-semibold">{entry.credit_account}</span>

                <span className="text-muted-foreground">Số tiền</span>
                <span className="font-semibold">{formatAmount(entry.amount)} VND</span>

                <span className="text-muted-foreground">Trạng thái</span>
                <span>{STATUS_LABELS[entry.status] ?? entry.status}</span>

                {entry.reject_reason && (
                  <>
                    <span className="text-muted-foreground">Lý do từ chối</span>
                    <span className="text-red-600">{entry.reject_reason}</span>
                  </>
                )}

                {entry.ai_confidence != null && (
                  <>
                    <span className="text-muted-foreground">AI Confidence</span>
                    <span>{Math.round(entry.ai_confidence * 100)}%</span>
                  </>
                )}
              </div>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              {entry.status === "pending" && (
                <Button variant="outline" onClick={handleEdit}>
                  Sửa
                </Button>
              )}
              <Button variant="outline" onClick={onClose}>
                Đóng
              </Button>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
