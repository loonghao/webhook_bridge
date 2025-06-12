#!/usr/bin/env node

/**
 * Production server starter for Next.js application
 * Optimized for production deployment with proper error handling
 */

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

const PORT = process.env.PORT || 3002;
const HOST = process.env.HOST || '0.0.0.0';

console.log('🚀 Starting webhook-bridge Next.js production server...');

// Check if dist directory exists
const distDir = path.join(__dirname, 'dist');
if (!fs.existsSync(distDir)) {
  console.error('❌ dist/ directory not found. Please run "npm run build:production" first.');
  process.exit(1);
}

// Check if index.html exists
const indexPath = path.join(distDir, 'index.html');
if (!fs.existsSync(indexPath)) {
  console.error('❌ index.html not found in dist/. Build may have failed.');
  process.exit(1);
}

console.log(`📁 Serving from: ${distDir}`);
console.log(`🌐 Server will start on: http://${HOST}:${PORT}`);

// Start Next.js production server
const nextStart = spawn('npx', ['next', 'start', '-p', PORT, '-H', HOST], {
  stdio: 'inherit',
  cwd: __dirname,
  env: {
    ...process.env,
    NODE_ENV: 'production',
    NEXT_PUBLIC_ENABLE_STAGEWISE: 'false'
  }
});

// Handle process termination
process.on('SIGINT', () => {
  console.log('\n🛑 Shutting down production server...');
  nextStart.kill('SIGINT');
  process.exit(0);
});

process.on('SIGTERM', () => {
  console.log('\n🛑 Shutting down production server...');
  nextStart.kill('SIGTERM');
  process.exit(0);
});

nextStart.on('error', (err) => {
  console.error('❌ Failed to start production server:', err);
  process.exit(1);
});

nextStart.on('exit', (code) => {
  if (code !== 0) {
    console.error(`❌ Production server exited with code ${code}`);
    process.exit(code);
  }
  console.log('✅ Production server stopped gracefully');
});
