# Stratavore Windows Build Script (PowerShell)
# Builds all binaries for Windows with proper error handling

$ErrorActionPreference = "Stop"

Write-Host "========================================"  -ForegroundColor Cyan
Write-Host "Stratavore Windows Build" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Set version info
$VERSION = "1.2.0"
$BUILD_TIME = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$COMMIT = "windows-build"

# Try to get git commit
try {
    $COMMIT = git rev-parse --short HEAD 2>$null
    if (-not $COMMIT) { $COMMIT = "windows-build" }
} catch {
    $COMMIT = "windows-build"
}

# Create bin directory
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build CLI
Write-Host "Building stratavore CLI..." -ForegroundColor Yellow
go mod tidy
$ldflags = "-X main.Version=$VERSION -X main.BuildTime=`"$BUILD_TIME`" -X main.Commit=$COMMIT"
go build -ldflags "$ldflags" -o bin\stratavore.exe .\cmd\stratavore
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Failed to build stratavore CLI" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] bin\stratavore.exe" -ForegroundColor Green

Write-Host ""

# Build Daemon
Write-Host "Building stratavored daemon..." -ForegroundColor Yellow
go build -ldflags=$ldflags -o bin\stratavored.exe .\cmd\stratavored
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Failed to build stratavored daemon" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] bin\stratavored.exe" -ForegroundColor Green

Write-Host ""

# Build Agent
Write-Host "Building stratavore-agent..." -ForegroundColor Yellow
go build -ldflags=$ldflags -o bin\stratavore-agent.exe .\cmd\stratavore-agent
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Failed to build stratavore-agent" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] bin\stratavore-agent.exe" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Binaries created in bin\ directory:" -ForegroundColor Cyan
Get-ChildItem bin\*.exe | ForEach-Object { Write-Host "  $($_.Name)" -ForegroundColor White }
Write-Host ""
Write-Host "To run:" -ForegroundColor Cyan
Write-Host "  .\bin\stratavored.exe    (start daemon)" -ForegroundColor White
Write-Host "  .\bin\stratavore.exe     (CLI)" -ForegroundColor White
Write-Host ""
