import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  eslint: {
    // Enable linting during builds
    ignoreDuringBuilds: false,
  },
  images: {
    domains: [
      'example.com',
      'res.cloudinary.com',
      'cloudinary.com',
      'images.unsplash.com',
      'placehold.co',
      'localhost',
      'dkfm4o59m.cloudinary.com',
      'api.cloudinary.com',
    ],
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**',
      },
      {
        protocol: 'http',
        hostname: '**',
      },
    ],
  },
  webpack(config) {
    config.module.rules.push({
      test: /\.svg$/,
      use: ["@svgr/webpack"],
    });
    return config;
  },
};

export default nextConfig;

