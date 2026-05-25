$ErrorActionPreference = "Stop"

$src = "internal/plugins/sdk"
$dst = "plugins/sdk"
$contractPkg = "internal/plugins/contract"

Write-Host "Extracting SDK to $dst..." -ForegroundColor Cyan

# Remove old extracted files
if (Test-Path "$dst/*.go") {
    Remove-Item "$dst/*.go"
}

# Copy SDK source files (excluding test files and the extract marker)
Get-ChildItem "$src/*.go" -Exclude "*_test.go", "extract.go" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    # Rewrite import paths: automated_dev_environment/... -> github.com/ade/plugins-sdk/...
    $content = $content -replace 'automated_dev_environment/internal/plugins/contract', 'github.com/ade/plugins-sdk/contract'
    $content = $content -replace '"automated_dev_environment/internal/plugins/sdk"', '"github.com/ade/plugins-sdk"'
    $outPath = Join-Path $dst $_.Name
    Set-Content -Path $outPath -Value $content
    Write-Host "  Extracted $($_.Name)" -ForegroundColor Gray
}

# Create contract package in extracted SDK
$contractDst = "$dst/contract"
if (-not (Test-Path $contractDst)) {
    New-Item -ItemType Directory -Path $contractDst -Force | Out-Null
}

# Copy contract types (non-test Go files including proto-generated)
Get-ChildItem "$contractPkg/*.go" -Exclude "*_test.go", "doc.go" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    $outPath = Join-Path $contractDst $_.Name
    Set-Content -Path $outPath -Value $content
    Write-Host "  Extracted contract/$($_.Name)" -ForegroundColor Gray
}

Write-Host "SDK extraction complete." -ForegroundColor Green
Write-Host ""
Write-Host "External plugins can now import:"
Write-Host "  import ""github.com/ade/plugins-sdk"""
Write-Host "  import ""github.com/ade/plugins-sdk/contract"""
