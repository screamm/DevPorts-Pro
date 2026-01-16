package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	netstatCache     string
	netstatCacheMu   sync.RWMutex
	netstatCacheTime time.Time
	netstatCacheTTL  = 2 * time.Second
)

type PortInfo struct {
	Port    int
	PID     string
	Process string
	Status  string
}

func ScanPorts() []PortInfo {
	var activePorts []PortInfo
	var mutex sync.Mutex

	// Use worker pool with channel for concurrent scanning
	portChan := make(chan int, AppConfig.PortRangeEnd)
	resultChan := make(chan PortInfo, 100)
	var wg sync.WaitGroup

	// Start concurrent workers for fast scanning
	numWorkers := AppConfig.NumWorkers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					// Worker recovered from panic - continue without crashing
					// The port will simply not be reported
				}
			}()
			for port := range portChan {
				if isPortOpen(port) {
					pid, process := getProcessInfo(port)
					resultChan <- PortInfo{
						Port:    port,
						PID:     pid,
						Process: process,
						Status:  "Active",
					}
				}
			}
		}()
	}

	// Result collector goroutine
	done := make(chan bool)
	go func() {
		for portInfo := range resultChan {
			mutex.Lock()
			activePorts = append(activePorts, portInfo)
			mutex.Unlock()
		}
		done <- true
	}()

	// Send ports to workers
	for port := AppConfig.PortRangeStart; port <= AppConfig.PortRangeEnd; port++ {
		portChan <- port
	}
	close(portChan)

	// Wait for workers to finish
	wg.Wait()
	close(resultChan)
	<-done

	// Sort by port number for consistent ordering
	sort.Slice(activePorts, func(i, j int) bool {
		return activePorts[i].Port < activePorts[j].Port
	})

	return activePorts
}

func isPortOpen(port int) bool {
	timeout := AppConfig.PortTimeout

	// Check IPv4 first (most common)
	conn4, err4 := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), timeout)
	if err4 == nil {
		conn4.Close()
		return true
	}

	// Check IPv6 (::1) for ports that only listen on IPv6
	conn6, err6 := net.DialTimeout("tcp", fmt.Sprintf("[::1]:%d", port), timeout)
	if err6 == nil {
		conn6.Close()
		return true
	}

	return false
}

func getCachedNetstatOutput(ctx context.Context) (string, error) {
	// Check cache first
	netstatCacheMu.RLock()
	if time.Since(netstatCacheTime) < netstatCacheTTL && netstatCache != "" {
		cached := netstatCache
		netstatCacheMu.RUnlock()
		return cached, nil
	}
	netstatCacheMu.RUnlock()

	// Cache miss or expired - run netstat
	cmd := exec.CommandContext(ctx, "netstat", "-ano")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Update cache
	netstatCacheMu.Lock()
	netstatCache = string(output)
	netstatCacheTime = time.Now()
	netstatCacheMu.Unlock()

	return string(output), nil
}

func clearNetstatCache() {
	netstatCacheMu.Lock()
	netstatCache = ""
	netstatCacheTime = time.Time{}
	netstatCacheMu.Unlock()
}

func getProcessInfo(port int) (string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), AppConfig.CommandTimeout)
	defer cancel()

	var output string
	var err error

	if runtime.GOOS == "windows" {
		output, err = getCachedNetstatOutput(ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return "Timeout", "Timeout"
			}
			return "Unknown", "Unknown"
		}
	} else {
		cmd := exec.CommandContext(ctx, "lsof", "-i", fmt.Sprintf(":%d", port))
		outputBytes, cmdErr := cmd.Output()
		if cmdErr != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return "Timeout", "Timeout"
			}
			return "Unknown", "Unknown"
		}
		output = string(outputBytes)
	}

	// Use regex for exact port match - handles LISTENING, ESTABLISHED, end of line, and CRLF
	// Match patterns like ":80 ", ":80\t", ":80\r", or ":80" at end of line
	portPattern := regexp.MustCompile(fmt.Sprintf(`[:\s]%d(?:[\s\t\r]|$)`, port))

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if portPattern.MatchString(line) {
			if runtime.GOOS == "windows" {
				fields := strings.Fields(line)
				if len(fields) >= 5 {
					pid := fields[len(fields)-1]
					process := getProcessName(pid)
					return pid, process
				}
			} else {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					process := fields[0]
					pid := fields[1]
					return pid, process
				}
			}
		}
	}

	return "Unknown", "Unknown"
}

func getProcessName(pid string) string {
	ctx, cancel := context.WithTimeout(context.Background(), AppConfig.CommandTimeout)
	defer cancel()

	if runtime.GOOS == "windows" {
		cmd := exec.CommandContext(ctx, "tasklist", "/fi", fmt.Sprintf("PID eq %s", pid), "/fo", "csv")
		// Hide console window on Windows
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		output, err := cmd.Output()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return "Timeout"
			}
			return "Unknown"
		}

		lines := strings.Split(string(output), "\n")
		if len(lines) >= 2 {
			fields := strings.Split(lines[1], ",")
			if len(fields) >= 1 {
				return strings.Trim(fields[0], "\"")
			}
		}
	} else {
		cmd := exec.CommandContext(ctx, "ps", "-p", pid, "-o", "comm=")
		output, err := cmd.Output()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return "Timeout"
			}
			return "Unknown"
		}
		return strings.TrimSpace(string(output))
	}

	return "Unknown"
}

func KillProcess(pid string) error {
	// Validate PID is a valid positive integer
	pidNum, err := strconv.Atoi(pid)
	if err != nil || pidNum <= 0 {
		return fmt.Errorf("invalid PID: %q", pid)
	}

	// Protect against killing critical system processes
	if runtime.GOOS == "windows" && pidNum <= 4 {
		return fmt.Errorf("cannot kill system process (PID %d)", pidNum)
	}

	ctx, cancel := context.WithTimeout(context.Background(), AppConfig.CommandTimeout)
	defer cancel()

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "taskkill", "/PID", pid, "/F")
		// Hide console window on Windows
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	} else {
		cmd = exec.CommandContext(ctx, "kill", "-9", pid)
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to kill process %s: %w", pid, err)
	}

	// Clear netstat cache since process state changed
	clearNetstatCache()

	// Verify process is actually killed by checking if PID still exists
	return verifyProcessKilled(pid)
}

func verifyProcessKilled(pid string) error {
	// Try multiple times with increasing delays
	maxAttempts := AppConfig.KillVerifyAttempts
	for attempt := 0; attempt < maxAttempts; attempt++ {
		time.Sleep(time.Duration(int(AppConfig.KillVerifyBaseDelay.Milliseconds())*(attempt+1)) * time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), AppConfig.CommandTimeout)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "tasklist", "/fi", fmt.Sprintf("PID eq %s", pid), "/fo", "csv", "/nh")
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		} else {
			cmd = exec.CommandContext(ctx, "ps", "-p", pid, "-o", "pid=")
		}

		output, err := cmd.Output()
		cancel() // Release context resources

		if runtime.GOOS == "windows" {
			// On Windows, check if tasklist returns actual process info
			outputStr := strings.TrimSpace(string(output))
			if err != nil || outputStr == "" || !strings.Contains(outputStr, pid) {
				return nil // Process killed successfully
			}
		} else {
			// On Unix, if command fails or output empty, process doesn't exist
			if err != nil || strings.TrimSpace(string(output)) == "" {
				return nil // Process killed successfully
			}
		}
	}

	return fmt.Errorf("process %s still running after %d kill attempts", pid, maxAttempts)
}
