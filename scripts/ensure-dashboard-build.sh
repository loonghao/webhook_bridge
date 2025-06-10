#!/bin/bash

# Ensure Dashboard Build Script
# This script ensures the dashboard is built before Go embed operations

set -e

echo "🏗️ Ensuring dashboard build for Go embed..."

cd web-nextjs

# Check if dist directory exists
if [ ! -d "dist" ]; then
    echo "📁 Creating dist directory..."
    mkdir -p dist
fi

# Check if we have a valid build
if [ ! -f "dist/index.html" ] || [ ! -d "dist/next" ]; then
    echo "🔨 Building dashboard..."
    
    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        echo "📦 Installing dependencies..."
        if [ -f "package-lock.json" ]; then
            npm ci
        else
            npm install
        fi
    fi
    
    # Try to build
    if npm run build; then
        echo "✅ Dashboard build successful"
    else
        echo "⚠️ Dashboard build failed, creating minimal structure..."
        
        # Create minimal structure for embed
        mkdir -p dist/next/static/css
        mkdir -p dist/next/static/chunks
        mkdir -p public
        
        # Create minimal index.html
        cat > dist/index.html << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Webhook Bridge Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .status { padding: 20px; background: #f0f0f0; border-radius: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Webhook Bridge Dashboard</h1>
        <div class="status">
            <h2>Status: Development Mode</h2>
            <p>The dashboard is running in development mode.</p>
            <p>For full functionality, please build the dashboard:</p>
            <pre>cd web-nextjs && npm run build</pre>
        </div>
    </div>
</body>
</html>
EOF

        # Create minimal CSS file
        echo "/* Minimal CSS for development */" > dist/next/static/css/app.css
        
        # Create minimal JS file
        echo "// Minimal JS for development" > dist/next/static/chunks/app.js
        
        # Create favicon
        if [ ! -f "public/favicon.ico" ]; then
            # Create a minimal favicon (empty file)
            touch public/favicon.ico
        fi
        
        # Copy favicon to dist
        cp public/favicon.ico dist/ 2>/dev/null || touch dist/favicon.ico
        
        echo "✅ Minimal dashboard structure created"
    fi
else
    echo "✅ Dashboard build already exists"
fi

cd ..

echo "🎯 Dashboard build check completed"
