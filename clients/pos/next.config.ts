import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  ...(process.env.STANDALONE && { output: "standalone" }),
};

export default nextConfig;
