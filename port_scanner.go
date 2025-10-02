package main

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

type PortInfo struct {
	Port    int
	PID     string
	Process string
	Status  string
}

func ScanPorts() []PortInfo {
	fmt.Println("Scanning ports 1-9999...")
	var activePorts []PortInfo
	var mutex sync.Mutex

	// Use worker pool with channel for concurrent scanning
	portChan := make(chan int, 9999)
	resultChan := make(chan PortInfo, 100)
	var wg sync.WaitGroup

	// Start 500 concurrent workers for fast scanning
	numWorkers := 500
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
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
			fmt.Printf("Found active port: %d (PID: %s, Process: %s)\n", portInfo.Port, portInfo.PID, portInfo.Process)
		}
		done <- true
	}()

	// Send ports to workers
	for port := 1; port <= 9999; port++ {
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

	fmt.Printf("Scan complete. Found %d active ports.\n", len(activePorts))
	return activePorts
}

func isPortOpen(port int) bool {
	timeout := time.Millisecond * 100

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

func getProcessInfo(port int) (string, string) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("netstat", "-ano")
		// Dölj konsol-fönster på Windows
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	} else {
		cmd = exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	}

	output, err := cmd.Output()
	if err != nil {
		return "Unknown", "Unknown"
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(":%d", port)) {
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
	if runtime.GOOS == "windows" {
		cmd := exec.Command("tasklist", "/fi", fmt.Sprintf("PID eq %s", pid), "/fo", "csv")
		// Dölj konsol-fönster på Windows
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		output, err := cmd.Output()
		if err != nil {
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
		cmd := exec.Command("ps", "-p", pid, "-o", "comm=")
		output, err := cmd.Output()
		if err != nil {
			return "Unknown"
		}
		return strings.TrimSpace(string(output))
	}

	return "Unknown"
}

func KillProcess(pid string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("taskkill", "/PID", pid, "/F")
		// Dölj konsol-fönster på Windows
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	} else {
		cmd = exec.Command("kill", "-9", pid)
	}

	err := cmd.Run()
	if err != nil {
		return err
	}

	// Verify process is actually killed by checking if PID still exists
	return verifyProcessKilled(pid)
}

func verifyProcessKilled(pid string) error {
	// Try multiple times with increasing delays
	maxAttempts := 5
	for attempt := 0; attempt < maxAttempts; attempt++ {
		time.Sleep(time.Duration(200*(attempt+1)) * time.Millisecond)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("tasklist", "/fi", fmt.Sprintf("PID eq %s", pid), "/fo", "csv", "/nh")
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		} else {
			cmd = exec.Command("ps", "-p", pid, "-o", "pid=")
		}

		output, err := cmd.Output()

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
