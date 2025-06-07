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
        try {
            fs.rmSync(newNextDir, { recursive: true, force: true });
        } catch (err) {
            console.log(`⚠️  Warning: Could not remove existing next directory: ${err.message}`);
            // Try to continue anyway
        }
    }

    // Copy _next to next (more reliable on Windows than rename)
    try {
        console.log('📁 Copying _next directory to next...');
        fs.cpSync(nextDir, newNextDir, { recursive: true });
        console.log('✅ Successfully copied _next to next');

        // Remove the original _next directory
        console.log('🗑️  Removing original _next directory...');
        fs.rmSync(nextDir, { recursive: true, force: true });
        console.log('✅ Successfully removed original _next directory');
    } catch (err) {
        console.log(`❌ Failed to copy/rename _next directory: ${err.message}`);
        console.log('💡 This might be due to file locks on Windows. Try closing any file explorers or editors that might be accessing the dist directory.');
        process.exit(1);
    }
    
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
