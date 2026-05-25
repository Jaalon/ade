$ErrorActionPreference = "Stop"

Write-Host "=== Generating protobuf code ===" -ForegroundColor Cyan

$bufPath = "$env:GOPATH/bin/buf.exe"
if (-not (Test-Path $bufPath)) {
    Write-Host "Installing buf..." -ForegroundColor Yellow
    go install github.com/bufbuild/buf/cmd/buf@latest
}

Write-Host "Running buf generate..." -ForegroundColor Green
buf generate

if ($LASTEXITCODE -ne 0) {
    Write-Host "buf generate failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "Protobuf generation complete." -ForegroundColor Green
