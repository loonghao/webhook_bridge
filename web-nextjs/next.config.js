/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'export',
  trailingSlash: true,
  distDir: 'dist',
  images: {
    unoptimized: true,
  },
  async rewrites() {
    // Only apply rewrites in development mode
    if (process.env.NODE_ENV === 'development') {
      return [
        {
          source: '/api/:path*',
          destination: 'http://localhost:8000/api/:path*',
        },
      ];
    }
    return [];
  },
  // 支持 MCP 调试
  webpack: (config, { dev, isServer }) => {
    if (dev && !isServer) {
      // 添加 React DevTools 支持
      config.resolve.alias = {
        ...config.resolve.alias,
        'react-dom$': 'react-dom/profiling',
        'scheduler/tracing': 'scheduler/tracing-profiling',
      };
    }
    return config;
  },
};

module.exports = nextConfig;
