import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import Pagination from "./pagination";

describe("Pagination", () => {
  it("shows current page and total pages", () => {
    render(<Pagination page={1} total={50} pageSize={20} onPage={vi.fn()} />);
    expect(screen.getByText(/1\s*\/\s*3/)).toBeInTheDocument();
  });

  it("disables Trước button on first page", () => {
    render(<Pagination page={1} total={50} pageSize={20} onPage={vi.fn()} />);
    expect(screen.getByRole("button", { name: /trước/i })).toBeDisabled();
  });

  it("disables Sau button on last page", () => {
    render(<Pagination page={3} total={50} pageSize={20} onPage={vi.fn()} />);
    expect(screen.getByRole("button", { name: /sau/i })).toBeDisabled();
  });

  it("calls onPage with page+1 when Sau clicked", () => {
    const onPage = vi.fn();
    render(<Pagination page={1} total={50} pageSize={20} onPage={onPage} />);
    fireEvent.click(screen.getByRole("button", { name: /sau/i }));
    expect(onPage).toHaveBeenCalledWith(2);
  });

  it("calls onPage with page-1 when Trước clicked", () => {
    const onPage = vi.fn();
    render(<Pagination page={2} total={50} pageSize={20} onPage={onPage} />);
    fireEvent.click(screen.getByRole("button", { name: /trước/i }));
    expect(onPage).toHaveBeenCalledWith(1);
  });

  it("shows nothing when total=0", () => {
    const { container } = render(
      <Pagination page={1} total={0} pageSize={20} onPage={vi.fn()} />
    );
    expect(container.firstChild).toBeNull();
  });

  it("shows page 1/1 when total <= pageSize", () => {
    render(<Pagination page={1} total={5} pageSize={20} onPage={vi.fn()} />);
    expect(screen.getByText(/1\s*\/\s*1/)).toBeInTheDocument();
  });
});
