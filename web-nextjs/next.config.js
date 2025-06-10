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
  // 支持 MCP 调试和生产优化
  webpack: (config, { dev, isServer }) => {
    if (dev && !isServer) {
      // 添加 React DevTools 支持
      config.resolve.alias = {
        ...config.resolve.alias,
        'react-dom$': 'react-dom/profiling',
        'scheduler/tracing': 'scheduler/tracing-profiling',
      };
    }

    // Production optimization: remove stagewise debug code if not explicitly enabled
    if (!dev && !process.env.NEXT_PUBLIC_ENABLE_STAGEWISE) {
      const path = require('path');
      config.resolve.alias = {
        ...config.resolve.alias,
        // Replace stagewise components with no-op versions in production
        '@/components/StagewiseDebugger': path.resolve(__dirname, './lib/no-op-stagewise.js'),
        '@/hooks/useStagewise': path.resolve(__dirname, './lib/no-op-stagewise.js'),
      };
    }

    // Add stagewise analyzer plugin
    if (!isServer && process.env.ANALYZE) {
      const StagewiseAnalyzerPlugin = require('./lib/stagewise-analyzer.js');
      config.plugins.push(new StagewiseAnalyzerPlugin({
        enabled: true,
        outputFile: 'stagewise-analysis.json'
      }));
    }

    return config;
  },
};

module.exports = nextConfig;
