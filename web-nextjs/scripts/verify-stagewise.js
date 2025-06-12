#!/usr/bin/env node

/**
 * Quick verification script for stagewise optimization
 * Usage: node scripts/verify-stagewise.js [production|debug|dev]
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const mode = process.argv[2] || 'production';

console.log(`🔍 Verifying stagewise optimization in ${mode} mode...\n`);

// Configuration for different modes
const configs = {
  production: {
    command: 'npm run build:production',
    expectStagewise: false,
    description: 'Production build should have minimal stagewise references'
  },
  debug: {
    command: 'npm run build:debug', 
    expectStagewise: true,
    description: 'Debug build should include stagewise functionality'
  },
  dev: {
    command: 'npm run build',
    expectStagewise: true,
    description: 'Development build should include full stagewise functionality'
  }
};

const config = configs[mode];
if (!config) {
  console.error(`❌ Invalid mode: ${mode}`);
  console.error(`Available modes: ${Object.keys(configs).join(', ')}`);
  process.exit(1);
}

console.log(`📋 ${config.description}`);
console.log(`⚙️  Running: ${config.command}\n`);

try {
  // Clean previous build
  if (fs.existsSync('dist')) {
    fs.rmSync('dist', { recursive: true, force: true });
  }
  
  // Run build
  const startTime = Date.now();
  execSync(config.command, { stdio: 'inherit' });
  const buildTime = Date.now() - startTime;
  
  console.log(`\n✅ Build completed in ${(buildTime / 1000).toFixed(2)}s`);
  
  // Quick verification
  console.log('\n🔍 Quick verification:');
  
  // Check if dist exists
  if (!fs.existsSync('dist')) {
    throw new Error('Build output directory not found');
  }
  
  // Get build size
  const buildSize = getBuildSize('dist');
  console.log(`📦 Build size: ${(buildSize / 1024 / 1024).toFixed(2)} MB`);
  
  // Check for stagewise references in main files
  const mainFiles = findMainFiles('dist');
  let stagewiseCount = 0;
  
  for (const file of mainFiles) {
    const content = fs.readFileSync(file, 'utf8');
    const matches = (content.match(/stagewise/gi) || []).length;
    stagewiseCount += matches;
  }
  
  console.log(`🔍 Stagewise references in main files: ${stagewiseCount}`);
  
  // Verify expectations
  if (config.expectStagewise) {
    if (stagewiseCount > 0) {
      console.log('✅ Stagewise functionality is present (as expected)');
    } else {
      console.log('⚠️  Warning: No stagewise references found (unexpected for this mode)');
    }
  } else {
    if (stagewiseCount === 0) {
      console.log('✅ No stagewise references in main files (optimal for production)');
    } else {
      console.log(`⚠️  Found ${stagewiseCount} stagewise references (may include route references)`);
    }
  }
  
  // Environment check
  console.log('\n🌍 Environment variables:');
  console.log(`NODE_ENV: ${process.env.NODE_ENV || 'not set'}`);
  console.log(`NEXT_PUBLIC_ENABLE_STAGEWISE: ${process.env.NEXT_PUBLIC_ENABLE_STAGEWISE || 'not set'}`);
  console.log(`NEXT_PUBLIC_DEBUG_MODE: ${process.env.NEXT_PUBLIC_DEBUG_MODE || 'not set'}`);
  
  console.log('\n🎉 Verification completed successfully!');
  
} catch (error) {
  console.error(`\n❌ Verification failed: ${error.message}`);
  process.exit(1);
}

function getBuildSize(dir) {
  let size = 0;
  
  function walk(currentDir) {
    const items = fs.readdirSync(currentDir);
    
    for (const item of items) {
      const fullPath = path.join(currentDir, item);
      const stat = fs.statSync(fullPath);
      
      if (stat.isDirectory()) {
        walk(fullPath);
      } else {
        size += stat.size;
      }
    }
  }
  
  if (fs.existsSync(dir)) {
    walk(dir);
  }
  
  return size;
}

function findMainFiles(dir) {
  const files = [];
  
  // Look for main HTML and JS files (not source maps or analysis files)
  function walk(currentDir) {
    const items = fs.readdirSync(currentDir);
    
    for (const item of items) {
      const fullPath = path.join(currentDir, item);
      const stat = fs.statSync(fullPath);
      
      if (stat.isDirectory()) {
        walk(fullPath);
      } else if (
        (item.endsWith('.html') || item.endsWith('.js')) &&
        !item.endsWith('.map') &&
        !item.includes('analysis') &&
        !item.includes('chunk')
      ) {
        files.push(fullPath);
      }
    }
  }
  
  if (fs.existsSync(dir)) {
    walk(dir);
  }
  
  return files.slice(0, 10); // Limit to first 10 main files for quick check
}
