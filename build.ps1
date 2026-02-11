# Stratavore Windows Build Script (PowerShell)
# Builds all binaries for Windows with protobuf generation

$ErrorActionPreference = "Stop"

Write-Host "========================================"  -ForegroundColor Cyan
Write-Host "Stratavore v1.3 Windows Build" -ForegroundColor Cyan
Write-Host "=======================================" -ForegroundColor Cyan
Write-Host ""

# Set version info
$VERSION = "1.3.0"
$BUILD_TIME = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$COMMIT = "windows-build"

# Try to get git commit
try {
    $COMMIT = git rev-parse --short HEAD 2>$null
    if (-not $COMMIT) { $COMMIT = "windows-build" }
} catch {
    $COMMIT = "windows-build"
}

# Check for protoc
Write-Host "Checking protobuf compiler..." -ForegroundColor Yellow
$hasProtoc = $false
try {
    $protocVersion = protoc --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        $hasProtoc = $true
        Write-Host "[OK] Found: $protocVersion" -ForegroundColor Green
    }
} catch {
    Write-Host "[SKIP] protoc not found - using fallback types" -ForegroundColor Yellow
}

# Check for protoc-gen-go
if ($hasProtoc) {
    Write-Host "Checking protoc-gen-go plugin..." -ForegroundColor Yellow
    $hasPlugin = $false
    try {
        $pluginCheck = protoc-gen-go --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            $hasPlugin = $true
            Write-Host "[OK] protoc-gen-go installed" -ForegroundColor Green
        }
    } catch {
        Write-Host "[SKIP] protoc-gen-go not found" -ForegroundColor Yellow
    }
    
    # Check for protoc-gen-go-grpc
    if ($hasPlugin) {
        Write-Host "Checking protoc-gen-go-grpc plugin..." -ForegroundColor Yellow
        try {
            $grpcCheck = protoc-gen-go-grpc --version 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "[OK] protoc-gen-go-grpc installed" -ForegroundColor Green
            } else {
                $hasPlugin = $false
                Write-Host "[SKIP] protoc-gen-go-grpc not found" -ForegroundColor Yellow
            }
        } catch {
            $hasPlugin = $false
            Write-Host "[SKIP] protoc-gen-go-grpc not found" -ForegroundColor Yellow
        }
    }
}

# Generate protobuf code if tools are available
if ($hasProtoc -and $hasPlugin) {
    Write-Host ""
    Write-Host "Generating protobuf code..." -ForegroundColor Yellow
    
    if (-not (Test-Path "pkg\api\generated")) {
        New-Item -ItemType Directory -Path "pkg\api\generated" | Out-Null
    }
    
    protoc --go_out=pkg/api/generated --go_opt=paths=source_relative `
           --go-grpc_out=pkg/api/generated --go-grpc_opt=paths=source_relative `
           --proto_path=pkg/api `
           pkg/api/stratavore.proto
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Protobuf code generated in pkg\api\generated\" -ForegroundColor Green
    } else {
        Write-Host "[WARN] Protobuf generation failed - using fallback" -ForegroundColor Yellow
    }
} else {
    Write-Host ""
    Write-Host "[INFO] Using fallback API types (no protobuf)" -ForegroundColor Cyan
    Write-Host "       To enable gRPC: install protoc and Go plugins" -ForegroundColor Cyan
    Write-Host "       See: https://grpc.io/docs/languages/go/quickstart/" -ForegroundColor Cyan
}

Write-Host ""

# Create bin directory
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build CLI
Write-Host "Building stratavore CLI..." -ForegroundColor Yellow
$ldflags = "-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME' -X 'main.Commit=$COMMIT'"
go build -ldflags=$ldflags -o bin\stratavore.exe .\cmd\stratavore
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

# Show protobuf status
if ($hasProtoc -and $hasPlugin) {
    Write-Host "gRPC: ENABLED (protobuf generated)" -ForegroundColor Green
} else {
    Write-Host "gRPC: Using HTTP API (protobuf tools not installed)" -ForegroundColor Yellow
}
Write-Host ""
