import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";

const geistSans = localFont({
  src: "./fonts/GeistVF.woff",
  variable: "--font-geist-sans",
  weight: "100 900",
});
const geistMono = localFont({
  src: "./fonts/GeistMonoVF.woff",
  variable: "--font-geist-mono",
  weight: "100 900",
});

export const metadata: Metadata = {
  title: "AI Kế Toán - Tự động hóa kế toán cho doanh nghiệp nhỏ",
  description:
    "AI Agent kế toán tự động: OCR hóa đơn, định khoản thông minh, tích hợp MISA AMIS. Giảm 70% thời gian kế toán cho SME Việt Nam.",
  keywords: [
    "phần mềm kế toán AI",
    "tự động hóa kế toán",
    "AI kế toán cho doanh nghiệp nhỏ",
    "OCR hóa đơn",
    "MISA AMIS",
  ],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="vi" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
