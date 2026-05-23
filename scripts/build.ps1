$ErrorActionPreference = "Stop"

Write-Host "Building ade.exe..." -ForegroundColor Green

go build -ldflags="-s -w" -o ade.exe ./cmd/ade

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "Build succeeded: $((Get-Item ade.exe).Length) bytes" -ForegroundColor Green
