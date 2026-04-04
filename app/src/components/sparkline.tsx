type SparklineProps = {
  data: number[];
  color?: string;
  width?: number;
  height?: number;
};

// Sparkline renders a tiny inline SVG polyline chart for use in stat cards.
export function Sparkline({ data, color = "#6366f1", width = 64, height = 24 }: SparklineProps) {
  if (data.length < 2) {
    return <svg width={width} height={height} />;
  }

  const max = Math.max(...data, 1);
  const stepX = width / (data.length - 1);
  const points = data
    .map((v, i) => {
      const x = i * stepX;
      const y = height - (v / max) * (height - 2) - 1;
      return `${x},${y}`;
    })
    .join(" ");

  return (
    <svg width={width} height={height} aria-hidden="true">
      <polyline
        points={points}
        fill="none"
        stroke={color}
        strokeWidth={1.5}
        strokeLinecap="round"
        strokeLinejoin="round"
        opacity={0.7}
      />
    </svg>
  );
}
