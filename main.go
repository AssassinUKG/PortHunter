package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ScanResult stores discovered open/closed/filtered ports and services
type ScanResult struct {
	DateTime string              `json:"datetime"`
	Ports    map[string][]string `json:"ports"`
}

// File paths
const scanFolder = "scan_data"
const scanFile = scanFolder + "/previous_scan.json"
const backupScanFile = scanFolder + "/previous_previous_scan.json"

// EnsureScanFolderExists creates the scan_data folder if it doesn't exist
func EnsureScanFolderExists() error {
	if _, err := os.Stat(scanFolder); os.IsNotExist(err) {
		return os.Mkdir(scanFolder, 0755)
	}
	return nil
}

// RunScan executes the user-supplied Nmap command and returns the results
func RunScan(command string, target string) (ScanResult, error) {
	// Validate input
	command = strings.TrimSpace(command)
	target = strings.TrimSpace(target)
	if command == "" {
		return ScanResult{}, errors.New("scan command cannot be empty")
	}
	if target == "" {
		return ScanResult{}, errors.New("target cannot be empty")
	}

	// Parse command into executable and args
	args := strings.Fields(command)

	// Handle "sudo" in command but still execute the full command
	executable := args[0]
	if executable == "sudo" && len(args) > 1 {
		executable = args[1] // Extract the real executable (Nmap)
	}

	args = append(args, target) // Append target at the end

	// Create command execution (keep original command structure)
	cmd := exec.Command(executable, args[1:]...)

	// Capture output
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Spinner for activity indication
	done := make(chan bool)
	go Spinner(done)

	// Run command
	err := cmd.Run()
	done <- true // Stop the spinner

	if err != nil {
		return ScanResult{}, fmt.Errorf("scan failed: %v\nOutput: %s", err, out.String())
	}

	// Parse only Nmap output
	results := ParseNmapOutput(out.String())

	// Return scan results with full timestamp
	return ScanResult{
		DateTime: time.Now().Format(time.RFC3339),
		Ports:    results,
	}, nil
}

// Spinner function to show activity while scan is running
func Spinner(done chan bool) {
	spinnerChars := []rune{'|', '/', '-', '\\'}
	i := 0

	for {
		select {
		case <-done:
			fmt.Print("\r") // Clear spinner when done
			return
		default:
			fmt.Printf("\rScanning... %c", spinnerChars[i%len(spinnerChars)])
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}

// ParseNmapOutput extracts all port states (open, closed, filtered) from Nmap output
func ParseNmapOutput(output string) map[string][]string {
	results := make(map[string][]string)

	lines := strings.Split(output, "\n")
	var currentIP string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect the scanned IP from "Nmap scan report for <IP>"
		if strings.HasPrefix(line, "Nmap scan report for ") {
			parts := strings.Fields(line)
			currentIP = parts[len(parts)-1]
			currentIP = strings.Trim(currentIP, "()") // Remove brackets if present
		} else if strings.Contains(line, "/tcp") && currentIP != "" {
			// Example Nmap port output:
			// 80/tcp  open     http
			// 443/tcp closed   https
			// 22/tcp  filtered ssh

			cols := strings.Fields(line)
			if len(cols) >= 3 {
				port := cols[0]    // Extract "80/tcp"
				state := cols[1]   // Extract "open", "closed", "filtered"
				service := cols[2] // Extract "http"

				// Save all states for proper tracking
				results[currentIP] = append(results[currentIP], fmt.Sprintf("%s [%s] (%s)", port, state, service))
			}
		}
	}
	return results
}

// LoadPreviousScan loads previous scan results from a JSON file
func LoadPreviousScan() (ScanResult, error) {
	data, err := os.ReadFile(scanFile)
	if err != nil {
		return ScanResult{}, err
	}

	var scan ScanResult
	err = json.Unmarshal(data, &scan)
	if err != nil {
		return ScanResult{}, err
	}

	return scan, nil
}

// SaveScan saves scan results to a JSON file, preserving the old scan before overwriting
func SaveScan(scan ScanResult) error {
	// Ensure the scan_data folder exists
	err := EnsureScanFolderExists()
	if err != nil {
		return err
	}

	// If a previous scan exists, move it before overwriting
	if _, err := os.Stat(scanFile); err == nil {
		os.Rename(scanFile, backupScanFile) // Move previous scan to backup before overwriting
	}

	data, err := json.MarshalIndent(scan, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(scanFile, data, 0644)
}

// CompareScans finds differences between scans, updates stored scan if changes are detected
func CompareScans(old, new ScanResult) {
	// ANSI color codes
	green := "\033[32m" // Green for added ports
	red := "\033[31m"   // Red for removed ports
	reset := "\033[0m"  // Reset to default color

	// Parse timestamps
	oldTime, err := time.Parse(time.RFC3339, old.DateTime)
	if err != nil {
		fmt.Println("Error parsing old scan time:", err)
		return
	}

	newTime, err := time.Parse(time.RFC3339, new.DateTime)
	if err != nil {
		fmt.Println("Error parsing new scan time:", err)
		return
	}

	// Calculate elapsed time
	elapsed := newTime.Sub(oldTime)
	fmt.Printf("\n--- Checking Previous Scan Data (Last scan was %s ago) ---\n\n", formatElapsedTime(elapsed))

	noChanges := true
	totalAdded, totalRemoved := 0, 0

	// Track changes for new scan results
	for ip, newPorts := range new.Ports {
		oldPorts := old.Ports[ip]
		added, removed := DiffPorts(oldPorts, newPorts)

		if len(added) > 0 || len(removed) > 0 {
			noChanges = false
			fmt.Printf("Changes for %s:\n", ip)

			if len(added) > 0 {
				totalAdded += len(added)
				fmt.Println("  [+] Added Ports:")
				for _, port := range added {
					fmt.Printf("    - %s%s%s\n", green, port, reset) // Green for added
				}
			}

			if len(removed) > 0 {
				totalRemoved += len(removed)
				fmt.Println("  [-] Removed Ports:")
				for _, port := range removed {
					fmt.Printf("    - %s%s%s\n", red, port, reset) // Red for removed
				}
			}
			fmt.Println()
		}
	}

	// Detect IPs and ports that were present in old scan but missing in the new scan
	for ip, oldPorts := range old.Ports {
		if _, exists := new.Ports[ip]; !exists {
			noChanges = false
			fmt.Printf("All ports for %s removed:\n", ip)
			for _, port := range oldPorts {
				totalRemoved++
				fmt.Printf("  [-] %s%s%s\n", red, port, reset) // Red for removed
			}
			fmt.Println()
		}
	}

	if noChanges {
		fmt.Println("No changes detected.")
	} else {
		fmt.Printf("Summary: %d new ports added, %d removed.\n", totalAdded, totalRemoved)
		SaveScan(new) // Save updated scan data
	}
}

// formatElapsedTime converts duration to human-readable format
func formatElapsedTime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	days := hours / 24

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours", days, hours%24)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else {
		return fmt.Sprintf("%d minutes", minutes)
	}
}

// DiffPorts finds added and removed ports
func DiffPorts(old, new []string) (added, removed []string) {
	oldSet := make(map[string]bool)
	for _, p := range old {
		oldSet[p] = true
	}

	newSet := make(map[string]bool)
	for _, p := range new {
		newSet[p] = true
	}

	for p := range newSet {
		if !oldSet[p] {
			added = append(added, p)
		}
	}

	for p := range oldSet {
		if !newSet[p] {
			removed = append(removed, p)
		}
	}

	return added, removed
}

// Main Execution
func main() {
	banner := `                                                            
 ____   ___  ____ _____ _   _ _   _ _   _ _____ _____ ____  
|  _ \ / _ \|  _ |_   _| | | | | | | \ | |_   _| ____|  _ \ 
| |_) | | | | |_) || | | |_| | | | |  \| | | | |  _| | |_) |
|  __/| |_| |  _ < | | |  _  | |_| | |\  | | | | |___|  _ < 
|_|    \___/|_| \_\|_| |_| |_|\___/|_| \_| |_| |_____|_| \_\
                                                            
		🔎 PortHunter - The Ultimate Port Checker
                    ⚡ Created by Richard Jones ⚡
`

	fmt.Println(banner)

	scanCmd := flag.String("c", "", "Full scan command (e.g., 'nmap -p- -T4')")
	target := flag.String("t", "", "Target IP/hostname")
	flag.Parse()

	scan, err := RunScan(*scanCmd, *target)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prevScan, err := LoadPreviousScan()
	if err == nil {
		CompareScans(prevScan, scan)
	} else {
		fmt.Println("No previous scan data found.")
	}

	SaveScan(scan)
	fmt.Println("Scan completed and saved.")
}
