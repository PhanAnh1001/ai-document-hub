import Link from "next/link";
import { Button } from "@/components/ui/button";
import { FileText, Zap, CheckCircle, RefreshCw } from "lucide-react";

const features = [
  {
    icon: FileText,
    title: "Upload & OCR tự động",
    description:
      "Upload hóa đơn, phiếu thu/chi, sao kê ngân hàng. AI trích xuất dữ liệu chính xác bằng FPT.AI.",
  },
  {
    icon: Zap,
    title: "Định khoản bằng AI",
    description:
      "AI engine tự động đề xuất bút toán theo chuẩn VAS (TT133/TT200). Tiết kiệm 80% thời gian.",
  },
  {
    icon: CheckCircle,
    title: "Review & Phê duyệt",
    description:
      "Kế toán review, duyệt hoặc từ chối từng bút toán. Hiển thị mức độ tin cậy của AI.",
  },
  {
    icon: RefreshCw,
    title: "Đồng bộ MISA AMIS",
    description:
      "Sau khi duyệt, bút toán được đẩy thẳng vào MISA AMIS. Không nhập liệu thủ công.",
  },
];

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col">
      {/* Hero */}
      <main className="flex flex-1 flex-col items-center justify-center px-4 py-24 text-center">
        <span className="mb-4 inline-block rounded-full bg-blue-50 px-3 py-1 text-xs font-semibold text-blue-700 ring-1 ring-blue-200">
          Dành cho SME Việt Nam
        </span>
        <h1 className="mb-4 text-4xl font-bold tracking-tight sm:text-6xl">
          AI Kế Toán
        </h1>
        <p className="mb-8 max-w-xl text-lg text-muted-foreground">
          Tự động hóa kế toán cho doanh nghiệp nhỏ. Upload chứng từ, AI xử lý
          định khoản, đồng bộ MISA AMIS — không cần nhập liệu thủ công.
        </p>
        <div className="flex gap-4 justify-center">
          <Button asChild size="lg">
            <Link href="/register">Dùng thử miễn phí</Link>
          </Button>
          <Button asChild variant="outline" size="lg">
            <Link href="/login">Đăng nhập</Link>
          </Button>
        </div>
      </main>

      {/* Features */}
      <section className="border-t bg-gray-50 px-4 py-16">
        <div className="mx-auto max-w-4xl">
          <h2 className="mb-10 text-center text-2xl font-bold">
            Tính năng chính
          </h2>
          <div className="grid gap-8 sm:grid-cols-2">
            {features.map((f) => (
              <div
                key={f.title}
                className="flex gap-4 rounded-xl bg-white p-6 shadow-sm ring-1 ring-gray-100"
              >
                <div className="shrink-0 rounded-lg bg-blue-50 p-2.5">
                  <f.icon className="h-5 w-5 text-blue-600" />
                </div>
                <div>
                  <h3 className="mb-1 font-semibold">{f.title}</h3>
                  <p className="text-sm text-muted-foreground">{f.description}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t px-4 py-6 text-center text-xs text-muted-foreground">
        © 2026 AI Kế Toán · Được xây dựng cho kế toán Việt Nam
      </footer>
    </div>
  );
}
