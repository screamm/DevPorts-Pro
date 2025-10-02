# DevPorts Pro v1.0.1 - Quick Start Guide

## 🚀 Launch Application

### Windows
```bash
devports-pro-improved.exe
```

### First Launch
1. Application opens at 1100x800 resolution
2. Initial port scan starts automatically (5-15 seconds)
3. Active ports appear in table sorted by port number

---

## ⚡ Key Features

### Lightning-Fast Scanning
- **Speed**: 9999 ports in 5-15 seconds
- **Technology**: 500 concurrent worker goroutines
- **Improvement**: 100-500x faster than v1.0.0

### Modern UI
- **Theme**: Cyan/green terminal aesthetic
- **Icons**: ⚡✓✗⟳⏳⨯ for clear status indication
- **Timing**: Shows scan duration in status bar

### Robust Process Killing
- **Verification**: Up to 5 retry attempts
- **Backoff**: Exponential delays (200ms-1000ms)
- **Feedback**: Error dialogs for failed attempts

---

## 📋 Interface Guide

### Main Window Components

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚡ DevPorts Pro v1.0 - Port Scanner
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

▸ Scanning ports 1-9999 | Auto-refresh: 5min

[⟳ Refresh Scan]    ⚡ Ready to scan...

┌─────┬──────┬───────────────┬────────┐
│ Port│ PID  │ Process       │ Action │
├─────┼──────┼───────────────┼────────┤
│ 3000│ 1234 │ node.exe      │⨯ Kill │
│ 5432│ 5678 │ postgres.exe  │⨯ Kill │
│ 8080│ 9012 │ java.exe      │⨯ Kill │
└─────┴──────┴───────────────┴────────┘

━━━ DevPorts Pro © 2024 | Press [Refresh Scan] to update ━━━
```

### Status Messages

| Icon | Message | Meaning |
|------|---------|---------|
| ⚡ | Ready to scan... | Idle state, ready for scan |
| ⟳ | Scanning ports... | Scan in progress |
| ✓ | Scan complete: X ports (Y.YYs) | Scan finished successfully |
| ⏳ | Terminating process PID X... | Kill in progress |
| ✓ | Process PID X terminated | Kill successful |
| ✗ | Failed to kill PID X | Kill failed |

---

## 🎯 Common Tasks

### Scan for Active Ports
1. Click **⟳ Refresh Scan**
2. Wait 5-15 seconds for scan to complete
3. View results in table

**Status Example**:
```
✓ Scan complete: 8 active ports found (7.23s)
```

### Kill a Process
1. Locate process in table
2. Click **⨯ Kill** button
3. Confirm in dialog
4. Wait for verification (up to 5 attempts)
5. Auto-refresh after 1.5 seconds

**Confirmation Dialog**:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  Terminate Process

Process: node.exe
PID: 1234
Port: 3000

Are you sure you want to terminate this process?
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[No]  [Yes]
```

### Auto-Refresh
- **Interval**: 5 minutes
- **Automatic**: Runs in background
- **Override**: Click refresh button anytime

---

## ⚙️ Performance Tips

### For Best Performance
1. **Close unnecessary applications**: Reduces ports to scan
2. **Run as Administrator**: For process termination privileges
3. **Stable network**: Ensures accurate port detection

### Expected Performance
- **Scan time**: 5-15 seconds for all 9999 ports
- **Memory usage**: ~50MB during scan
- **CPU usage**: Moderate during scan, idle otherwise

---

## 🛡️ Safety Notes

### Process Termination
⚠️ **Warning**: Killing processes can cause data loss or system instability

**Safe to Kill**:
- Development servers (node, python, etc.)
- Test applications
- Your own processes

**Avoid Killing**:
- System processes (svchost, csrss, etc.)
- Antivirus software
- Critical system services

### Permissions
- **Windows**: Run as Administrator for full functionality
- **Process Access**: Some processes require elevated privileges
- **Verification**: Application confirms termination before reporting success

---

## 🔧 Troubleshooting

### Scan is Slow
- **Expected**: 5-15 seconds is normal
- **Longer**: May indicate high port usage or system load
- **Solution**: Close unnecessary applications

### Can't Kill Process
**Error**: "Failed to kill PID X: access denied"

**Solutions**:
1. Run application as Administrator
2. Check if process is system-protected
3. Verify you have permission to terminate

### Application Won't Start
**Check**:
1. Windows 10 or later required
2. No other instance running
3. Antivirus not blocking

---

## 📊 Performance Comparison

### v1.0.0 vs v1.0.1

| Feature | v1.0.0 | v1.0.1 | Improvement |
|---------|--------|--------|-------------|
| Scan Time | 5-10 min | 5-15 sec | **100-500x faster** |
| Workers | 1 | 500 | **500x parallel** |
| Kill Verify | 1 attempt | 5 attempts | **5x reliable** |
| UI Response | Good | Excellent | **Better UX** |

---

## 🎓 Advanced Usage

### Understanding the Scan
```
9999 ports ÷ 500 workers × 100ms timeout ≈ 2 seconds minimum
+ Network latency + Process lookup ≈ 5-15 seconds total
```

### Process Verification
```
Attempt 1: 200ms delay → Check if process exists
Attempt 2: 400ms delay → Check if process exists
Attempt 3: 600ms delay → Check if process exists
Attempt 4: 800ms delay → Check if process exists
Attempt 5: 1000ms delay → Check if process exists
Total: Up to 3 seconds maximum for verification
```

---

## 💡 Tips & Tricks

1. **Quick Refresh**: Press refresh immediately after kill for instant update
2. **Sort Order**: Ports are always sorted numerically (1-9999)
3. **Process Names**: Full executable name shown (e.g., "node.exe")
4. **Timing**: Watch status bar for scan duration
5. **Errors**: Error dialogs show detailed failure reasons

---

## 📞 Support

### Common Issues
- **Build**: `devports-pro-improved.exe` (41MB)
- **Version**: v1.0.1 (2025-09-30)
- **Go Version**: go1.25.0

### Need Help?
- Check IMPROVEMENTS_SUMMARY.md for technical details
- See README.md for full documentation
- Review test_improvements.md for testing checklist

---

**Ready to go! Launch the app and enjoy lightning-fast port scanning! ⚡**