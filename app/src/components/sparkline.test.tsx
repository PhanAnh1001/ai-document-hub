import { render } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { Sparkline } from "./sparkline";

describe("Sparkline", () => {
  it("renders an SVG element", () => {
    const { container } = render(<Sparkline data={[1, 2, 3, 4, 5]} />);
    const svg = container.querySelector("svg");
    expect(svg).not.toBeNull();
  });

  it("renders a polyline when data has values", () => {
    const { container } = render(<Sparkline data={[3, 1, 4, 1, 5, 9, 2]} />);
    const polyline = container.querySelector("polyline");
    expect(polyline).not.toBeNull();
  });

  it("renders nothing meaningful for empty data", () => {
    const { container } = render(<Sparkline data={[]} />);
    const svg = container.querySelector("svg");
    expect(svg).not.toBeNull();
    // No polyline for empty data
    const polyline = container.querySelector("polyline");
    expect(polyline).toBeNull();
  });

  it("renders nothing meaningful for all-zero data", () => {
    const { container } = render(<Sparkline data={[0, 0, 0]} />);
    const polyline = container.querySelector("polyline");
    // All zeros → flat line still renders
    expect(polyline).not.toBeNull();
  });

  it("accepts color prop", () => {
    const { container } = render(<Sparkline data={[1, 2, 3]} color="#ff0000" />);
    const polyline = container.querySelector("polyline");
    expect(polyline?.getAttribute("stroke")).toBe("#ff0000");
  });
});
