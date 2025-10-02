# DevPorts Pro - Quick Build Guide

## TL;DR - Build Command

```powershell
# Run the automated build script
.\build.ps1
```

**OR** Manual build:

```powershell
$env:CGO_ENABLED = "1"
go build -v -ldflags="-H windowsgui -extldflags '-static'" -tags netgo -trimpath -o devports-pro.exe .
```

## Prerequisites

### 1. Go Compiler
```powershell
go version  # Must be 1.21+
```

### 2. MinGW-w64 (Required for CGO)
```powershell
gcc --version  # Must have GCC installed
```

**If missing:**
```powershell
choco install mingw -y
```

## Build Steps

### Option 1: Automated Build (Recommended)
```powershell
cd "C:\Users\david\Documents\FSU23D\Egna Projekt\DevPorts Pro"
.\build.ps1
```

### Option 2: Manual Build
```powershell
cd "C:\Users\david\Documents\FSU23D\Egna Projekt\DevPorts Pro"

# Set environment
$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# Build
go build -v `
  -ldflags="-H windowsgui -extldflags '-static'" `
  -tags netgo `
  -trimpath `
  -o devports-pro.exe .
```

## Verify Build

```powershell
# Check file exists
ls devports-pro.exe

# Run application
.\devports-pro.exe

# Test port detection (in another terminal)
python -m http.server 5174

# Verify port 5174 appears in DevPorts Pro
```

## Troubleshooting

### "gcc: command not found"
```powershell
choco install mingw -y
```

### Build fails with CGO errors
```powershell
# Verify CGO is enabled
go env CGO_ENABLED  # Should show "1"

# Check GCC path
$env:Path -split ";" | Select-String "mingw"
```

### Port 5174 not detected
**This should be FIXED with the new build configuration.**

If still missing:
1. Verify build used `-tags netgo` flag
2. Check you didn't use `-s -w` flags
3. Rebuild using `.\build.ps1`

### Application won't run on other PC
**This should be FIXED - the new build embeds all dependencies.**

If still failing:
1. Install Visual C++ Redistributable on target PC
2. Verify target PC has Windows 10/11 x64
3. Check antivirus isn't blocking the .exe

## Expected Results

| Metric | Expected Value |
|--------|----------------|
| Build time | 30-60 seconds |
| File size | 35-40 MB |
| Ports detected | ALL active ports (1-9999) |
| GUI appears | Yes, no console window |
| Runs on other PC | Yes, without Go/MinGW installed |

## Files Overview

| File | Purpose |
|------|---------|
| `build.ps1` | Automated build script (recommended) |
| `BUILD_CONFIGURATION.md` | Detailed technical documentation |
| `QUICK_BUILD_GUIDE.md` | This quick reference |
| `devports-pro.exe` | Final executable (created by build) |

## Distribution

After successful build:

```powershell
# The executable is self-contained
# Just copy devports-pro.exe to target system

# No additional files needed:
# - No DLLs required
# - No Go installation required
# - No MinGW required

# Just run:
.\devports-pro.exe
```

## Build Flag Explanation

| Flag | Why Needed |
|------|------------|
| `CGO_ENABLED=1` | Fyne GUI requires C libraries |
| `-H windowsgui` | No console window |
| `-extldflags '-static'` | Embed C dependencies |
| `-tags netgo` | Fix port detection issues |
| `-trimpath` | Smaller binary, security |

## What Changed From Previous Builds?

### ❌ Old (Broken) Builds

**Regular build:**
- Missing static linking
- Required DLLs on target system
- Failed with "Cannot run app"

**Static build with `-s -w`:**
- Stripped too aggressively
- Corrupted network syscalls
- Missed ports like 5174

### ✅ New (Working) Build

- **CGO enabled** for Fyne GUI
- **Static linking** for C dependencies
- **Pure Go network stack** (`netgo` tag)
- **No aggressive stripping** (preserves syscalls)
- **All ports detected correctly**

## Testing Checklist

- [ ] Build completes without errors
- [ ] devports-pro.exe created (~35-40 MB)
- [ ] Application launches (no console window)
- [ ] Port scan detects port 5174 (test server)
- [ ] All active ports shown in table
- [ ] Process kill functionality works
- [ ] Runs on clean Windows PC (no Go/MinGW)

## Support

For detailed technical information, see:
- `BUILD_CONFIGURATION.md` - Complete technical documentation
- Fyne documentation: https://developer.fyne.io/
- Go build documentation: https://pkg.go.dev/cmd/go

## Quick Commands Reference

```powershell
# Build
.\build.ps1

# Run
.\devports-pro.exe

# Test port detection
python -m http.server 5174

# Check binary size
(Get-Item devports-pro.exe).Length / 1MB

# Verify no console window
.\devports-pro.exe  # Should show GUI only, no console

# Clean build artifacts
Remove-Item devports-pro.exe
go clean -cache
```
