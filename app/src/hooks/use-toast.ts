import { useState, useEffect } from "react";

export type ToastVariant = "success" | "error" | "info";

export type Toast = {
  id: string;
  title: string;
  description?: string;
  variant: ToastVariant;
};

type ToastInput = {
  title: string;
  description?: string;
  variant?: ToastVariant;
};

// Module-level state so toasts are shared across all hook instances
let toasts: Toast[] = [];
const listeners = new Set<() => void>();

function notify() {
  listeners.forEach((l) => l());
}

function addToast(input: ToastInput): string {
  const id = Math.random().toString(36).slice(2);
  toasts = [...toasts, { id, title: input.title, description: input.description, variant: input.variant ?? "info" }];
  notify();
  // Auto-dismiss after 4s
  setTimeout(() => dismissToast(id), 4000);
  return id;
}

function dismissToast(id?: string) {
  toasts = id ? toasts.filter((t) => t.id !== id) : [];
  notify();
}

// Callable function + dismiss method
export function toast(input: ToastInput): string {
  return addToast(input);
}
toast.dismiss = dismissToast;

export function useToast() {
  const [state, setState] = useState<Toast[]>(toasts);

  useEffect(() => {
    const listener = () => setState([...toasts]);
    listeners.add(listener);
    return () => { listeners.delete(listener); };
  }, []);

  return { toasts: state };
}
