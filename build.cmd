@echo off
echo ========================================
echo  DevPorts Pro - Optimized Build Script
echo ========================================
echo.

REM Clean previous builds
echo [1/4] Cleaning previous builds...
if exist devports-pro.exe del /F /Q devports-pro.exe
go clean -cache

REM Set build environment
echo [2/4] Setting build environment...
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

REM Build with optimal flags
echo [3/4] Building DevPorts Pro...
echo   - CGO: Enabled (required for Fyne GUI)
echo   - Network: Pure Go (netgo tag)
echo   - Linking: Static C libraries
echo   - GUI: Windows subsystem (no console)
echo.

go build -v -ldflags="-H windowsgui -extldflags '-static'" -tags netgo -trimpath -o devports-pro.exe .

if errorlevel 1 (
    echo.
    echo ========================================
    echo  BUILD FAILED!
    echo ========================================
    echo.
    echo Troubleshooting:
    echo   1. Ensure Go 1.21+ is installed
    echo   2. Install MinGW-w64: choco install mingw
    echo   3. Check that fyne dependencies are installed
    echo.
    pause
    exit /b 1
)

REM Verify build
echo [4/4] Verifying build...
if exist devports-pro.exe (
    for %%A in (devports-pro.exe) do set size=%%~zA
    echo.
    echo ========================================
    echo  BUILD SUCCESSFUL!
    echo ========================================
    echo.
    echo Executable: devports-pro.exe
    echo Size: %size% bytes
    echo.
    echo Features:
    echo   [X] Full GUI with Fyne v2
    echo   [X] IPv4 + IPv6 port detection
    echo   [X] All ports 1-9999 scanned
    echo   [X] No external dependencies
    echo   [X] Portable - runs on any Windows 10/11 x64
    echo.
    echo Test Instructions:
    echo   1. Start a test server: python -m http.server 5174
    echo   2. Run: devports-pro.exe
    echo   3. Verify port 5174 appears in the table
    echo.
) else (
    echo ERROR: devports-pro.exe was not created!
    pause
    exit /b 1
)

pause
