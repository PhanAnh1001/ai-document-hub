import { describe, it, expect, beforeEach, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useToast, toast } from "./use-toast";

describe("useToast", () => {
  beforeEach(() => {
    // Reset state between tests by calling dismiss all
    act(() => {
      toast.dismiss();
    });
  });

  it("adds a toast when toast() is called", () => {
    const { result } = renderHook(() => useToast());

    act(() => {
      toast({ title: "Thành công", variant: "success" });
    });

    expect(result.current.toasts).toHaveLength(1);
    expect(result.current.toasts[0].title).toBe("Thành công");
    expect(result.current.toasts[0].variant).toBe("success");
  });

  it("supports error variant", () => {
    const { result } = renderHook(() => useToast());

    act(() => {
      toast({ title: "Lỗi xảy ra", variant: "error" });
    });

    expect(result.current.toasts[0].variant).toBe("error");
  });

  it("assigns unique ids to toasts", () => {
    const { result } = renderHook(() => useToast());

    act(() => {
      toast({ title: "First" });
      toast({ title: "Second" });
    });

    const ids = result.current.toasts.map((t) => t.id);
    expect(new Set(ids).size).toBe(2);
  });

  it("dismiss removes a specific toast by id", () => {
    const { result } = renderHook(() => useToast());
    let id = "";

    act(() => {
      id = toast({ title: "Remove me" });
    });

    act(() => {
      toast.dismiss(id);
    });

    expect(result.current.toasts.find((t) => t.id === id)).toBeUndefined();
  });

  it("dismiss with no args clears all toasts", () => {
    const { result } = renderHook(() => useToast());

    act(() => {
      toast({ title: "A" });
      toast({ title: "B" });
    });

    act(() => {
      toast.dismiss();
    });

    expect(result.current.toasts).toHaveLength(0);
  });
});
