"use client";

import { useToast } from "@/hooks/use-toast";

const VARIANT_STYLES = {
  success: "bg-green-600 text-white",
  error:   "bg-red-600 text-white",
  info:    "bg-gray-800 text-white",
};

export function Toaster() {
  const { toasts } = useToast();

  if (toasts.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 max-w-sm">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`rounded-lg px-4 py-3 shadow-lg text-sm font-medium animate-in slide-in-from-bottom-2 ${VARIANT_STYLES[t.variant]}`}
        >
          <p>{t.title}</p>
          {t.description && (
            <p className="mt-0.5 text-xs opacity-80">{t.description}</p>
          )}
        </div>
      ))}
    </div>
  );
}
