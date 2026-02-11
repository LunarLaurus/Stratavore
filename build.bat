@echo off
REM Stratavore Windows Build Script v1.3
REM Builds all binaries with optional protobuf generation

setlocal enabledelayedexpansion

echo ========================================
echo Stratavore v1.3 Windows Build
echo ========================================
echo.

REM Set version info
set VERSION=1.3.0
set BUILD_TIME=%date% %time%
set COMMIT=windows-build

REM Get commit hash if git is available
git rev-parse --short HEAD > nul 2>&1
if !errorlevel! == 0 (
    for /f %%i in ('git rev-parse --short HEAD') do set COMMIT=%%i
)

REM Check for protoc
echo Checking protobuf compiler...
where protoc > nul 2>&1
if !errorlevel! == 0 (
    echo [OK] protoc found
    set HAS_PROTOC=1
) else (
    echo [SKIP] protoc not found - using fallback types
    set HAS_PROTOC=0
)

REM Check for Go plugins
if !HAS_PROTOC! == 1 (
    where protoc-gen-go > nul 2>&1
    if !errorlevel! == 0 (
        where protoc-gen-go-grpc > nul 2>&1
        if !errorlevel! == 0 (
            echo [OK] protoc Go plugins found
            set HAS_PLUGINS=1
        ) else (
            echo [SKIP] protoc-gen-go-grpc not found
            set HAS_PLUGINS=0
        )
    ) else (
        echo [SKIP] protoc-gen-go not found
        set HAS_PLUGINS=0
    )
)

REM Generate protobuf code if tools available
if !HAS_PROTOC! == 1 if !HAS_PLUGINS! == 1 (
    echo.
    echo Generating protobuf code...
    if not exist pkg\api\generated mkdir pkg\api\generated
    
    protoc --go_out=pkg/api/generated --go_opt=paths=source_relative ^
           --go-grpc_out=pkg/api/generated --go-grpc_opt=paths=source_relative ^
           --proto_path=pkg/api ^
           pkg/api/stratavore.proto
    
    if !errorlevel! == 0 (
        echo [OK] Protobuf code generated
    ) else (
        echo [WARN] Protobuf generation failed - using fallback
    )
) else (
    echo.
    echo [INFO] Using fallback API types (no protobuf^)
    echo        To enable gRPC: install protoc and Go plugins
)

echo.

REM Create bin directory
if not exist bin mkdir bin

echo Building stratavore CLI...
go build -ldflags="-X 'main.Version=%VERSION%' -X 'main.BuildTime=%BUILD_TIME%' -X 'main.Commit=%COMMIT%'" -o bin\stratavore.exe .\cmd\stratavore
if !errorlevel! neq 0 (
    echo ERROR: Failed to build stratavore CLI
    exit /b 1
)
echo [OK] bin\stratavore.exe

echo.
echo Building stratavored daemon...
go build -ldflags="-X 'main.Version=%VERSION%' -X 'main.BuildTime=%BUILD_TIME%' -X 'main.Commit=%COMMIT%'" -o bin\stratavored.exe .\cmd\stratavored
if !errorlevel! neq 0 (
    echo ERROR: Failed to build stratavored daemon
    exit /b 1
)
echo [OK] bin\stratavored.exe

echo.
echo Building stratavore-agent...
go build -ldflags="-X 'main.Version=%VERSION%' -X 'main.BuildTime=%BUILD_TIME%' -X 'main.Commit=%COMMIT%'" -o bin\stratavore-agent.exe .\cmd\stratavore-agent
if !errorlevel! neq 0 (
    echo ERROR: Failed to build stratavore-agent
    exit /b 1
)
echo [OK] bin\stratavore-agent.exe

echo.
echo ========================================
echo Build Complete!
echo ========================================
echo.
echo Binaries created in bin\ directory:
dir /b bin\*.exe
echo.
echo To run:
echo   bin\stratavored.exe    (start daemon^)
echo   bin\stratavore.exe     (CLI^)
echo.

if !HAS_PROTOC! == 1 if !HAS_PLUGINS! == 1 (
    echo gRPC: ENABLED (protobuf generated^)
) else (
    echo gRPC: Using HTTP API (protobuf tools not installed^)
)
echo.

endlocal
