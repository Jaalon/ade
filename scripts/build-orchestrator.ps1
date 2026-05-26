$ErrorActionPreference = "Stop"
$ImageName = "ade/ade-config:latest"

Write-Host "Building orchestrator image ${ImageName}..." -ForegroundColor Green

docker build -t $ImageName -f Dockerfile .

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "Build succeeded: ${ImageName}" -ForegroundColor Green
Write-Host ""
Write-Host "Pour pousser l'image sur le registre :" -ForegroundColor Yellow
Write-Host "  docker push ${ImageName}" -ForegroundColor Yellow
