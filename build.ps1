# DevPorts Pro - Production Build Script
# Ensures reproducible, working builds for Windows x64

Write-Host "================================" -ForegroundColor Cyan
Write-Host "  DevPorts Pro - Build Script" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# 1. Verify Build Environment
Write-Host "[1/5] Verifying build environment..." -ForegroundColor Yellow

$goVersion = go version
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go compiler not found" -ForegroundColor Red
    exit 1
}
Write-Host "  ✓ $goVersion" -ForegroundColor Green

$gccVersion = gcc --version 2>&1 | Select-Object -First 1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: GCC (MinGW-w64) not found - CGO required for Fyne" -ForegroundColor Red
    Write-Host "Install: choco install mingw" -ForegroundColor Yellow
    exit 1
}
Write-Host "  ✓ $gccVersion" -ForegroundColor Green

# 2. Set Build Environment
Write-Host ""
Write-Host "[2/5] Configuring build environment..." -ForegroundColor Yellow

$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "  ✓ CGO_ENABLED=1 (Fyne requires CGO)" -ForegroundColor Green
Write-Host "  ✓ GOOS=windows, GOARCH=amd64" -ForegroundColor Green

# 3. Clean Previous Builds
Write-Host ""
Write-Host "[3/5] Cleaning previous builds..." -ForegroundColor Yellow

if (Test-Path "devports-pro.exe") {
    Remove-Item "devports-pro.exe" -Force
    Write-Host "  ✓ Removed old devports-pro.exe" -ForegroundColor Green
}

# 4. Build Application
Write-Host ""
Write-Host "[4/5] Building DevPorts Pro..." -ForegroundColor Yellow

$buildCommand = @(
    "build"
    "-v"
    "-ldflags=-H windowsgui -extldflags '-static'"
    "-tags", "netgo"
    "-trimpath"
    "-o", "devports-pro.exe"
    "."
)

go @buildCommand

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    exit 1
}

# 5. Verify Build Output
Write-Host ""
Write-Host "[5/5] Verifying build output..." -ForegroundColor Yellow

if (-Not (Test-Path "devports-pro.exe")) {
    Write-Host "ERROR: devports-pro.exe not created" -ForegroundColor Red
    exit 1
}

$fileSize = (Get-Item "devports-pro.exe").Length / 1MB
Write-Host "  ✓ devports-pro.exe created ($([math]::Round($fileSize, 2)) MB)" -ForegroundColor Green

# Success Summary
Write-Host ""
Write-Host "================================" -ForegroundColor Green
Write-Host "  BUILD SUCCESSFUL" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green
Write-Host ""
Write-Host "Output: devports-pro.exe" -ForegroundColor Cyan
Write-Host "Size: $([math]::Round($fileSize, 2)) MB" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "  1. Test: .\devports-pro.exe" -ForegroundColor White
Write-Host "  2. Verify all ports detected (including 5174)" -ForegroundColor White
Write-Host "  3. Test process kill functionality" -ForegroundColor White
Write-Host ""
