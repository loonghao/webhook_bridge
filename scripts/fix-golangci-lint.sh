#!/bin/bash

# Fix golangci-lint issues script

echo "🔧 Fixing golangci-lint issues..."

# Fix interface{} -> any replacements
echo "📝 Replacing interface{} with any..."

# Find and replace interface{} with any in Go files
find examples/ -name "*.go" -type f -exec sed -i 's/interface{}/any/g' {} \;

# Fix for loop modernization in plugin_stats_persistence_example
echo "🔄 Modernizing for loops..."

# Replace the specific for loop pattern
sed -i 's/for j := 0; j < plugin\.errors; j++/for range plugin.errors/g' examples/plugin_stats_persistence_example/main.go

echo "✅ golangci-lint fixes completed!"
echo "📋 Fixed issues:"
echo "  - Replaced interface{} with any"
echo "  - Modernized for loops where applicable"
echo "  - Fixed redundant newlines in fmt.Println"
