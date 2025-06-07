#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

console.log('🔧 Post-processing Next.js build for Go embed compatibility...');

const distDir = path.join(__dirname, 'dist');
const nextDir = path.join(distDir, '_next');
const newNextDir = path.join(distDir, 'next');

// Check if _next directory exists
if (fs.existsSync(nextDir)) {
    console.log('📁 Found _next directory, renaming to next...');
    
    // Remove existing next directory if it exists
    if (fs.existsSync(newNextDir)) {
        console.log('🗑️  Removing existing next directory...');
        fs.rmSync(newNextDir, { recursive: true, force: true });
    }
    
    // Rename _next to next
    fs.renameSync(nextDir, newNextDir);
    console.log('✅ Successfully renamed _next to next');
    
    // Update HTML files to use new path
    console.log('🔄 Updating HTML files to use new paths...');
    
    function updateHtmlFiles(dir) {
        const files = fs.readdirSync(dir);
        
        for (const file of files) {
            const filePath = path.join(dir, file);
            const stat = fs.statSync(filePath);
            
            if (stat.isDirectory()) {
                updateHtmlFiles(filePath);
            } else if (file.endsWith('.html')) {
                console.log(`   Updating ${filePath}...`);
                let content = fs.readFileSync(filePath, 'utf8');
                
                // Replace _next with next in HTML content
                content = content.replace(/_next\//g, 'next/');
                
                fs.writeFileSync(filePath, content, 'utf8');
            }
        }
    }
    
    updateHtmlFiles(distDir);
    
    console.log('✅ Build post-processing completed successfully!');
    console.log('📦 The dist directory is now ready for Go embed');
} else {
    console.log('❌ _next directory not found in dist');
    process.exit(1);
}
