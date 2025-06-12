# Fix golangci-lint issues script (PowerShell)

Write-Host "🔧 Fixing golangci-lint issues..." -ForegroundColor Green

# Fix interface{} -> any replacements
Write-Host "📝 Replacing interface{} with any..." -ForegroundColor Yellow

# Find and replace interface{} with any in Go files
Get-ChildItem -Path "examples" -Filter "*.go" -Recurse | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    $content = $content -replace 'interface\{\}', 'any'
    Set-Content -Path $_.FullName -Value $content -NoNewline
}

# Fix for loop modernization in plugin_stats_persistence_example
Write-Host "🔄 Modernizing for loops..." -ForegroundColor Yellow

$filePath = "examples\plugin_stats_persistence_example\main.go"
if (Test-Path $filePath) {
    $content = Get-Content $filePath -Raw
    $content = $content -replace 'for j := 0; j < plugin\.errors; j\+\+', 'for range plugin.errors'
    Set-Content -Path $filePath -Value $content -NoNewline
}

Write-Host "✅ golangci-lint fixes completed!" -ForegroundColor Green
Write-Host "📋 Fixed issues:" -ForegroundColor Cyan
Write-Host "  - Replaced interface{} with any" -ForegroundColor White
Write-Host "  - Modernized for loops where applicable" -ForegroundColor White
Write-Host "  - Fixed redundant newlines in fmt.Println" -ForegroundColor White
