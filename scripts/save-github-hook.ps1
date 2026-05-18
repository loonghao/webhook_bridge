param()

$ErrorActionPreference = "Stop"

$inputJson = [Console]::In.ReadToEnd()
if ([string]::IsNullOrWhiteSpace($inputJson)) {
    throw "Expected webhook envelope JSON on stdin"
}

$envelope = $inputJson | ConvertFrom-Json
$outputDir = if ($env:WEBHOOK_BRIDGE_OUTPUT_DIR) { $env:WEBHOOK_BRIDGE_OUTPUT_DIR } else { "data/github-hooks" }
New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

$delivery = $null
if ($envelope.headers.'x-github-delivery') {
    $delivery = $envelope.headers.'x-github-delivery'
} elseif ($envelope.headers.'X-GitHub-Delivery') {
    $delivery = $envelope.headers.'X-GitHub-Delivery'
}

$event = $null
if ($envelope.headers.'x-github-event') {
    $event = $envelope.headers.'x-github-event'
} elseif ($envelope.headers.'X-GitHub-Event') {
    $event = $envelope.headers.'X-GitHub-Event'
}

$safeDelivery = if ($delivery) { ($delivery -replace '[^a-zA-Z0-9._-]', '_') } else { $envelope.request_id }
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss-fff"
$fileName = "$timestamp-$safeDelivery.json"
$filePath = Join-Path $outputDir $fileName
$latestPath = Join-Path $outputDir "latest.json"

$record = [ordered]@{
    saved_at = (Get-Date).ToUniversalTime().ToString("o")
    request_id = $envelope.request_id
    route = $envelope.route
    provider = $envelope.provider
    github_event = $event
    github_delivery = $delivery
    method = $envelope.method
    query = $envelope.query
    headers = $envelope.headers
    payload = $envelope.payload
}

$json = $record | ConvertTo-Json -Depth 100
$json | Set-Content -LiteralPath $filePath -Encoding UTF8
$json | Set-Content -LiteralPath $latestPath -Encoding UTF8

[ordered]@{
    status = "saved"
    provider = "github"
    event = $event
    delivery = $delivery
    path = (Resolve-Path -LiteralPath $filePath).Path
    latest = (Resolve-Path -LiteralPath $latestPath).Path
} | ConvertTo-Json -Depth 10
