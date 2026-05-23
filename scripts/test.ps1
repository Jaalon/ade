$ErrorActionPreference = "Stop"

Write-Host "=== Unit & Integration Tests ===" -ForegroundColor Green
go test ./internal/... -v -count=1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Unit/integration tests failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "`n=== E2E Tests ===" -ForegroundColor Green
go test -tags=e2e -v -count=1 ./test/e2e/
if ($LASTEXITCODE -ne 0) {
    Write-Host "E2E tests failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "`nAll tests passed!" -ForegroundColor Green
