@echo off
REM Stratavore Windows Build Script
REM Builds all binaries for Windows

setlocal enabledelayedexpansion

echo ========================================
echo Stratavore Windows Build
echo ========================================
echo.

REM Set version info
set VERSION=1.2.0
set BUILD_TIME=%date% %time%
set COMMIT=windows-build

REM Get commit hash if git is available
git rev-parse --short HEAD > nul 2>&1
if !errorlevel! == 0 (
    for /f %%i in ('git rev-parse --short HEAD') do set COMMIT=%%i
)

REM Create bin directory
if not exist bin mkdir bin

echo Building stratavore CLI...
go mod tidy
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
echo   bin\stratavored.exe    (start daemon)
echo   bin\stratavore.exe     (CLI)
echo.

endlocal
