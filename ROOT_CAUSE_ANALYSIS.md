# Root Cause Analysis: DevPorts Pro Executable Failure

**Investigation Date**: 2025-10-01
**System**: Windows 10/11 x64
**Go Version**: go1.25.0 windows/amd64
**Compiler**: TDM-GCC 10.3.0 (x86_64-w64-mingw32)

---

## Executive Summary

The `devports-pro.exe` fails to run with error "Det går inte att köra appen på den här datorn" (Cannot run the app on this computer) due to **build tag conflicts** introduced in the source code. The build tags create mutually exclusive compilation conditions that prevent the executable from building correctly.

---

## Evidence Chain

### 1. File Modifications Detected

**Critical Discovery**: Build tags were added to source files:

**File**: `C:\Users\david\Documents\FSU23D\Egna Projekt\DevPorts Pro\port_scanner.go`
```go
//go:build netgo
// +build netgo
```

**File**: `C:\Users\david\Documents\FSU23D\Egna Projekt\DevPorts Pro\main.go`
```go
//go:build windows && cgo
// +build windows,cgo
```

### 2. Build Tag Conflict Analysis

**Root Cause**: The build tags create **mutually exclusive conditions**:

- `port_scanner.go` requires: `netgo` tag → **CGO disabled**
- `main.go` requires: `windows AND cgo` → **CGO enabled**

**Result**: When building `devports-pro.exe`:
- Build system includes `main.go` (CGO enabled, satisfies `windows && cgo`)
- Build system **EXCLUDES** `port_scanner.go` (requires `netgo`, incompatible with CGO)
- Executable compiles but **missing core port scanning functionality**
- Corrupted binary structure causes Windows to reject the executable

### 3. Technical Evidence

#### DLL Dependencies Analysis
```bash
# Both executables have identical DLL dependencies:
devports-pro.exe:        GDI32.dll, KERNEL32.dll, msvcrt.dll, OPENGL32.dll, SHELL32.dll, USER32.dll
devports-pro-static.exe: GDI32.dll, KERNEL32.dll, msvcrt.dll, OPENGL32.dll, SHELL32.dll, USER32.dll
```

**Conclusion**: DLL dependencies are NOT the issue. The problem is source code exclusion.

#### File Size Comparison
```bash
devports-pro.exe:        42MB  (24 sections)
devports-pro-static.exe: 23MB  (12 sections, stripped)
```

**Analysis**:
- `devports-pro.exe` is **larger** but **non-functional** → corrupted build with incomplete code
- `devports-pro-static.exe` is smaller, works, but missing port 5174

#### PE Header Analysis
```bash
# Both have identical PE headers:
Magic:           020b (PE32+)
Subsystem:       00000002 (Windows GUI)
ImageBase:       0000000140000000
```

**Conclusion**: PE structure is correct. The issue is at the Go build level, not Windows compatibility.

### 4. Port 5174 Missing in Static Build

**Evidence**:
```bash
netstat -ano | findstr ":5174"
  TCP    [::1]:5174             [::]:0                 LISTENING       7060
```

**Root Cause**: Port 5174 is listening on **IPv6 only** (`[::1]`)

**Code Analysis** (`port_scanner.go` line 86):
```go
conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), timeout)
```

**Problem**: Scanner only checks **IPv4 (127.0.0.1)**, never checks **IPv6 (::1)**

---

## Root Cause Statement

The `devports-pro.exe` fails due to **conflicting build tags** that exclude critical source files during compilation:

1. **Build Tag Conflict**: `main.go` requires `windows && cgo`, `port_scanner.go` requires `netgo`
2. **Exclusion Effect**: When CGO is enabled (default), `port_scanner.go` is excluded from build
3. **Corrupted Binary**: Executable compiles with missing core functionality, causing Windows to reject it
4. **Secondary Issue**: Port scanner only checks IPv4 (`127.0.0.1`), missing IPv6 (`::1`) ports

---

## Technical Explanation

### Why the Regular Build Fails

**Go Build System Behavior**:
1. Default Go build has `CGO_ENABLED=1` on Windows with TDM-GCC
2. `main.go` satisfies build constraints: `windows && cgo` ✓
3. `port_scanner.go` fails build constraints: `netgo` requires CGO disabled ✗
4. **Result**: Binary builds without `port_scanner.go` → incomplete/corrupted

**Windows Error Explanation**:
- Windows attempts to load the executable
- PE structure is valid, but runtime initialization fails
- Missing critical function definitions from excluded `port_scanner.go`
- Windows reports: "Den angivna filen är inte ett giltigt program" (Not a valid program)

### Why the Static Build Works (Partially)

**Build Command**: `go build -ldflags="-H windowsgui -s -w" -tags static -o devports-pro-static.exe .`

**Behavior**:
1. `-tags static` overrides build constraints (neither `netgo` nor `cgo` required)
2. Both `main.go` and `port_scanner.go` compile successfully
3. `-s -w` strips debug symbols (23MB vs 42MB)
4. Executable works but misses IPv6 ports due to scanner limitation

---

## Remediation Steps

### Solution 1: Remove Build Tags (Recommended)

**Action**: Delete conflicting build tags from source files

**File**: `C:\Users\david\Documents\FSU23D\Egna Projekt\DevPorts Pro\port_scanner.go`
```go
// DELETE these lines:
//go:build netgo
// +build netgo
```

**File**: `C:\Users\david\Documents\FSU23D\Egna Projekt\DevPorts Pro\main.go`
```go
// DELETE these lines:
//go:build windows && cgo
// +build windows,cgo
```

**Build Command**:
```bash
go build -ldflags="-H windowsgui -s -w" -o devports-pro.exe .
```

**Expected Outcome**:
- ✓ Full functionality with CGO enabled
- ✓ Proper Fyne GUI rendering
- ✓ Complete port scanning (1-9999)
- ✓ Process kill functionality
- ⚠ Still missing IPv6 ports (requires code fix)

### Solution 2: Fix IPv6 Port Detection

**Problem**: Port scanner only checks IPv4 (`127.0.0.1`)

**Fix**: Update `isPortOpen()` function in `port_scanner.go`:

```go
func isPortOpen(port int) bool {
    timeout := time.Millisecond * 100

    // Check IPv4
    conn4, err4 := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), timeout)
    if err4 == nil {
        conn4.Close()
        return true
    }

    // Check IPv6
    conn6, err6 := net.DialTimeout("tcp", fmt.Sprintf("[::1]:%d", port), timeout)
    if err6 == nil {
        conn6.Close()
        return true
    }

    return false
}
```

**Expected Outcome**:
- ✓ Detects both IPv4 and IPv6 ports
- ✓ Will find port 5174 (listening on `[::1]`)
- ✓ Complete port coverage

### Solution 3: Optimized Build Configuration

**Recommended Build Flags**:
```bash
# Production build (recommended)
go build -ldflags="-H windowsgui -s -w" -trimpath -o devports-pro.exe .

# Development build (with debugging)
go build -ldflags="-H windowsgui" -gcflags="all=-N -l" -o devports-pro-dev.exe .
```

**Flag Explanation**:
- `-H windowsgui`: Hide console window
- `-s -w`: Strip debug symbols (reduce size from 42MB → ~23MB)
- `-trimpath`: Remove absolute file paths (security best practice)
- `-gcflags="all=-N -l"`: Disable optimization for debugging (dev only)

---

## Verification Steps

### Step 1: Remove Build Tags
```bash
cd "C:\Users\david\Documents\FSU23D\Egna Projekt\DevPorts Pro"

# Remove build tags from both files
# (Manual edit or automated script)
```

### Step 2: Clean Build
```bash
go clean -cache
go clean -modcache -i -r
del devports-pro.exe
```

### Step 3: Rebuild
```bash
go build -ldflags="-H windowsgui -s -w" -trimpath -o devports-pro.exe .
```

### Step 4: Verify Executable
```bash
# Check file size (should be ~23-25MB with stripped symbols)
ls -lh devports-pro.exe

# Check PE format
file devports-pro.exe
# Expected: PE32+ executable (GUI) x86-64, for MS Windows

# Check DLL dependencies
objdump -p devports-pro.exe | grep "DLL Name"
# Expected: GDI32.dll, KERNEL32.dll, msvcrt.dll, OPENGL32.dll, SHELL32.dll, USER32.dll
```

### Step 5: Test Execution
```bash
# Run executable
./devports-pro.exe

# Should launch GUI without errors
# Should detect ports 1-9999
# Should detect both IPv4 and IPv6 ports (after code fix)
```

### Step 6: Verify Port Detection
```bash
# Check if port 5174 is detected
# After applying IPv6 fix, port should appear in GUI table
netstat -ano | findstr ":5174"
```

---

## Prevention Strategy

### Build System Best Practices

1. **Avoid Build Tags Unless Necessary**
   - Build tags should only be used for platform-specific code
   - Don't use conflicting tags in the same project
   - Document tag requirements in `README.md`

2. **Consistent Build Configuration**
   - Use Makefile or build script for reproducible builds
   - Document all required build flags
   - Use `-trimpath` for production builds

3. **Automated Testing**
   - Add CI/CD pipeline to catch build tag conflicts
   - Test executables on clean Windows environment
   - Verify port detection with both IPv4 and IPv6

### Code Quality Improvements

1. **Network Layer Abstraction**
   - Create helper function for dual-stack (IPv4+IPv6) port scanning
   - Test against both protocol versions
   - Handle edge cases (IPv6-only, IPv4-only systems)

2. **Error Handling**
   - Add comprehensive error messages for build failures
   - Log network connection errors for debugging
   - Provide user feedback when ports can't be scanned

---

## Conclusion

### Primary Issue: Build Tag Conflict
- **Root Cause**: Mutually exclusive build tags (`netgo` vs `windows && cgo`)
- **Impact**: Core functionality excluded from build, corrupted executable
- **Solution**: Remove conflicting build tags
- **Effort**: 2 minutes (delete 4 lines)

### Secondary Issue: IPv6 Port Detection
- **Root Cause**: Scanner only checks IPv4 (`127.0.0.1`)
- **Impact**: Missing ports listening on IPv6 (`[::1]`)
- **Solution**: Add IPv6 connection check in `isPortOpen()`
- **Effort**: 15 minutes (code + testing)

### Recommendation Priority
1. **Critical**: Remove build tags → Restore functionality
2. **High**: Add IPv6 support → Complete port coverage
3. **Medium**: Optimize build flags → Reduce executable size
4. **Low**: Add automated testing → Prevent regressions

---

## Appendix: Build Environment Details

### Go Environment
```
GOARCH=amd64
GOOS=windows
CGO_ENABLED=1
CC=gcc
CGO_CFLAGS=-O2 -g
CGO_LDFLAGS=-O2 -g
```

### GCC Configuration
```
Target: x86_64-w64-mingw32
Thread model: posix
gcc version: 10.3.0 (tdm64-1)
Location: C:\TDM-GCC-64\bin\gcc.exe
```

### Runtime DLLs Available
```
libgcc_s_seh_64-1.dll    (123KB)
libwinpthread-1.dll      (64KB)
libstdc++-6.dll          (2.1MB)
```

**Note**: These DLLs are NOT required by the Go executable (Go uses static linking). The failure is NOT a missing DLL issue.
