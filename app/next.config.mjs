/** @type {import('next').NextConfig} */
const nextConfig = {
  // "export" for Cloudflare Pages (static site, no SSR server needed).
  // "standalone" for Docker/self-hosted deployments.
  output: process.env.NEXT_OUTPUT ?? "standalone",

  images: {
    // Static export does not support Next.js Image Optimization — use unoptimized fallback.
    unoptimized: process.env.NEXT_OUTPUT === "export",
    remotePatterns: [
      {
        protocol: "https",
        hostname: "*.supabase.co",
        pathname: "/storage/v1/object/public/**",
      },
    ],
  },
};

export default nextConfig;
