#!/usr/bin/env node

/**
 * Post-build script for Next.js application
 * Fixes common issues with static export and optimizes for Go embedding
 */

const fs = require('fs');
const path = require('path');

const DIST_DIR = path.join(__dirname, 'dist');
const OUT_DIR = path.join(__dirname, 'out');

console.log('🔧 Starting post-build fixes...');

// Ensure dist directory exists
if (!fs.existsSync(DIST_DIR)) {
  if (fs.existsSync(OUT_DIR)) {
    console.log('📁 Moving out/ to dist/');
    fs.renameSync(OUT_DIR, DIST_DIR);
  } else {
    console.error('❌ Neither dist/ nor out/ directory found');
    process.exit(1);
  }
}

// Fix index.html if it exists
const indexPath = path.join(DIST_DIR, 'index.html');
if (fs.existsSync(indexPath)) {
  console.log('🔧 Fixing index.html...');
  let content = fs.readFileSync(indexPath, 'utf8');
  
  // Fix asset paths for Go embedding
  content = content.replace(/\/_next\//g, '/next/');
  content = content.replace(/href="\/favicon\.ico"/g, 'href="/favicon.ico"');
  
  fs.writeFileSync(indexPath, content);
  console.log('✅ Fixed index.html asset paths');
}

// Fix CSS files
const cssDir = path.join(DIST_DIR, '_next', 'static', 'css');
const newCssDir = path.join(DIST_DIR, 'next', 'static', 'css');

if (fs.existsSync(cssDir)) {
  console.log('🔧 Moving CSS files...');
  
  // Create new directory structure
  fs.mkdirSync(path.dirname(newCssDir), { recursive: true });
  
  // Move CSS directory
  fs.renameSync(cssDir, newCssDir);
  console.log('✅ Moved CSS files to next/static/css/');
}

// Fix JS files
const jsDir = path.join(DIST_DIR, '_next', 'static', 'chunks');
const newJsDir = path.join(DIST_DIR, 'next', 'static', 'chunks');

if (fs.existsSync(jsDir)) {
  console.log('🔧 Moving JS files...');
  
  // Create new directory structure
  fs.mkdirSync(path.dirname(newJsDir), { recursive: true });
  
  // Move JS directory
  fs.renameSync(jsDir, newJsDir);
  console.log('✅ Moved JS files to next/static/chunks/');
}

// Remove old _next directory if it's empty
const oldNextDir = path.join(DIST_DIR, '_next');
if (fs.existsSync(oldNextDir)) {
  try {
    const files = fs.readdirSync(oldNextDir, { recursive: true });
    if (files.length === 0) {
      fs.rmSync(oldNextDir, { recursive: true });
      console.log('✅ Removed empty _next directory');
    }
  } catch (err) {
    console.log('⚠️  Could not remove _next directory:', err.message);
  }
}

// Create index.txt for Go embedding verification
const indexTxtPath = path.join(DIST_DIR, 'index.txt');
const indexTxtContent = `Next.js build completed at ${new Date().toISOString()}
Build directory: ${DIST_DIR}
Files processed: ${getFileCount(DIST_DIR)}
`;

fs.writeFileSync(indexTxtPath, indexTxtContent);
console.log('✅ Created index.txt');

// Helper function to count files
function getFileCount(dir) {
  let count = 0;
  try {
    const files = fs.readdirSync(dir, { recursive: true });
    count = files.filter(file => {
      const fullPath = path.join(dir, file);
      return fs.statSync(fullPath).isFile();
    }).length;
  } catch (err) {
    console.log('⚠️  Could not count files:', err.message);
  }
  return count;
}

console.log('🎉 Post-build fixes completed successfully!');
console.log(`📊 Total files in dist: ${getFileCount(DIST_DIR)}`);
