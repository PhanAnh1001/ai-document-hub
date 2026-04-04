"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { getSettings, updateSettings, type OrgSettings } from "@/lib/api";
import { useAuthStore } from "@/lib/auth-store";
import { toast } from "@/hooks/use-toast";

export default function SettingsPage() {
  const token = useAuthStore((s) => s.token);
  const [org, setOrg] = useState<OrgSettings | null>(null);
  const [name, setName] = useState("");
  const [taxCode, setTaxCode] = useState("");
  const [accountingStandard, setAccountingStandard] = useState<"TT133" | "TT200">("TT200");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [ocrProvider, setOcrProvider] = useState<"fpt" | "mock">("mock");
  const [misaApiUrl, setMisaApiUrl] = useState("");
  const [misaApiKey, setMisaApiKey] = useState("");
  const [savingConn, setSavingConn] = useState(false);
  const [connMessage, setConnMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [showApiKey, setShowApiKey] = useState(false);
  const [urlError, setUrlError] = useState("");

  useEffect(() => {
    if (!token) return;
    getSettings(token)
      .then((res) => {
        setOrg(res.organization);
        setName(res.organization.name);
        setTaxCode(res.organization.tax_code ?? "");
        setAccountingStandard(res.organization.accounting_standard);
        setOcrProvider(res.organization.ocr_provider ?? "mock");
        setMisaApiUrl(res.organization.misa_api_url ?? "");
        setMisaApiKey(res.organization.misa_api_key ?? "");
      })
      .catch(() => {});
  }, [token]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) return;
    setSaving(true);
    setMessage(null);
    try {
      const updated = await updateSettings(token, {
        name,
        tax_code: taxCode || null,
        accounting_standard: accountingStandard,
      });
      setOrg(updated.organization);
      setMessage({ type: "success", text: "Đã lưu thành công" });
      toast({ title: "Đã lưu thành công", variant: "success" });
    } catch {
      setMessage({ type: "error", text: "Lỗi khi lưu, thử lại sau" });
      toast({ title: "Lỗi khi lưu, thử lại sau", variant: "error" });
    } finally {
      setSaving(false);
    }
  };

  const handleConnSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) return;
    // Validate MISA API URL if provided
    if (misaApiUrl) {
      try {
        const parsed = new URL(misaApiUrl);
        if (parsed.protocol !== "http:" && parsed.protocol !== "https:") throw new Error();
        setUrlError("");
      } catch {
        setUrlError("URL không hợp lệ. Vui lòng nhập URL bắt đầu bằng http:// hoặc https://");
        return;
      }
    } else {
      setUrlError("");
    }
    setSavingConn(true);
    setConnMessage(null);
    try {
      const updated = await updateSettings(token, {
        ocr_provider: ocrProvider,
        misa_api_url: misaApiUrl || null,
        misa_api_key: misaApiKey || null,
      });
      setOrg(updated.organization);
      setConnMessage({ type: "success", text: "Đã lưu thành công" });
      toast({ title: "Đã lưu kết nối thành công", variant: "success" });
    } catch {
      setConnMessage({ type: "error", text: "Lỗi khi lưu, thử lại sau" });
      toast({ title: "Lỗi khi lưu kết nối", variant: "error" });
    } finally {
      setSavingConn(false);
    }
  };

  return (
    <div>
      <h1 className="mb-6 text-3xl font-bold">Cài đặt</h1>
      <div className="space-y-4 max-w-xl">
        <Card>
          <CardHeader>
            <CardTitle>Thông tin doanh nghiệp</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium" htmlFor="org-name">
                  Tên doanh nghiệp
                </label>
                <Input
                  id="org-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Nhập tên doanh nghiệp"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium" htmlFor="tax-code">
                  Mã số thuế
                </label>
                <Input
                  id="tax-code"
                  value={taxCode}
                  onChange={(e) => setTaxCode(e.target.value)}
                  placeholder="0123456789"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium" htmlFor="std">
                  Chuẩn kế toán
                </label>
                <select
                  id="std"
                  value={accountingStandard}
                  onChange={(e) => setAccountingStandard(e.target.value as "TT133" | "TT200")}
                  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                >
                  <option value="TT133">TT133 (Doanh nghiệp nhỏ)</option>
                  <option value="TT200">TT200 (Doanh nghiệp lớn)</option>
                </select>
              </div>
              {message && (
                <p className={`text-sm ${message.type === "success" ? "text-green-600" : "text-red-600"}`}>
                  {message.text}
                </p>
              )}
              <Button type="submit" disabled={saving || !org}>
                {saving ? "Đang lưu..." : "Lưu"}
              </Button>
            </form>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Kết nối</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleConnSubmit} className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium" htmlFor="ocr-provider">
                  OCR Provider
                </label>
                <select
                  id="ocr-provider"
                  value={ocrProvider}
                  onChange={(e) => setOcrProvider(e.target.value as "fpt" | "mock")}
                  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                >
                  <option value="fpt">FPT.AI</option>
                  <option value="mock">Mock (dev)</option>
                </select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium" htmlFor="misa-api-url">
                  MISA API URL
                </label>
                <Input
                  id="misa-api-url"
                  value={misaApiUrl}
                  onChange={(e) => { setMisaApiUrl(e.target.value); setUrlError(""); }}
                  placeholder="https://api.misa.vn/..."
                />
                {urlError && (
                  <p className="mt-1 text-xs text-red-600">{urlError}</p>
                )}
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium" htmlFor="misa-api-key">
                  MISA API Key
                </label>
                <div className="flex gap-2">
                  <Input
                    id="misa-api-key"
                    type={showApiKey ? "text" : "password"}
                    value={misaApiKey}
                    onChange={(e) => setMisaApiKey(e.target.value)}
                    placeholder="••••••••"
                  />
                  <button
                    type="button"
                    onClick={() => setShowApiKey((v) => !v)}
                    className="px-3 py-1.5 text-sm border border-gray-300 rounded hover:bg-gray-50 whitespace-nowrap"
                  >
                    {showApiKey ? "Ẩn" : "Hiện"}
                  </button>
                </div>
              </div>
              {connMessage && (
                <p className={`text-sm ${connMessage.type === "success" ? "text-green-600" : "text-red-600"}`}>
                  {connMessage.text}
                </p>
              )}
              <Button type="submit" disabled={savingConn || !org}>
                {savingConn ? "Đang lưu..." : "Lưu kết nối"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
