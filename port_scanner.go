package main

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
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
	
	for port := 1; port <= 9999; port++ {
		if isPortOpen(port) {
			pid, process := getProcessInfo(port)
			activePorts = append(activePorts, PortInfo{
				Port:    port,
				PID:     pid,
				Process: process,
				Status:  "Active",
			})
			fmt.Printf("Found active port: %d (PID: %s, Process: %s)\n", port, pid, process)
		}
		
		if port%1000 == 0 {
			fmt.Printf("Scanned %d ports...\n", port)
		}
	}
	
	fmt.Printf("Scan complete. Found %d active ports.\n", len(activePorts))
	return activePorts
}

func isPortOpen(port int) bool {
	timeout := time.Millisecond * 50
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
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
	// Wait a moment for process to terminate
	time.Sleep(500 * time.Millisecond)
	
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist", "/fi", fmt.Sprintf("PID eq %s", pid), "/fo", "csv")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	} else {
		cmd = exec.Command("ps", "-p", pid)
	}
	
	output, err := cmd.Output()
	if err != nil {
		// If command fails, process likely doesn't exist (killed successfully)
		return nil
	}
	
	if runtime.GOOS == "windows" {
		lines := strings.Split(string(output), "\n")
		if len(lines) <= 1 {
			// Only header line, no process found
			return nil
		}
		// Check if we have actual process data beyond header
		for _, line := range lines[1:] {
			if strings.TrimSpace(line) != "" {
				return fmt.Errorf("process %s still running after kill attempt", pid)
			}
		}
		return nil
	} else {
		// On Unix, if ps succeeds, process still exists
		if strings.TrimSpace(string(output)) != "" {
			return fmt.Errorf("process %s still running after kill attempt", pid)
		}
		return nil
	}
}