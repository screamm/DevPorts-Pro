package main

import "fmt"

func main() {
	fmt.Println("DevPorts Pro - Port Scanner Test")
	fmt.Println("=================================")
	
	ports := ScanPorts()
	
	if len(ports) == 0 {
		fmt.Println("No active ports found.")
		return
	}
	
	fmt.Println("\nActive Ports Found:")
	fmt.Println("Port\tPID\tProcess")
	fmt.Println("----\t---\t-------")
	
	for _, port := range ports {
		fmt.Printf("%d\t%s\t%s\n", port.Port, port.PID, port.Process)
	}
}