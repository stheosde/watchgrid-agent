package ports

import (
	"bufio"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

var (
	tcpPath  = "/proc/net/tcp"
	tcp6Path = "/proc/net/tcp6"
)

func CheckNonWhitelistedPorts(allowedPorts []int) ([]int, error) {
	openPorts, err := OpenPorts()
	if err != nil {
		return nil, err
	}
	var nonWhitelisted []int
	for _, port := range openPorts {
		if !slices.Contains(allowedPorts, port) {
			nonWhitelisted = append(nonWhitelisted, port)
		}
	}
	return nonWhitelisted, nil
}

func CheckStoppedServices(servicesToCheck []string) []string {
	var stopped []string
	for _, service := range servicesToCheck {
		if !ServiceIsRunning(service) {
			stopped = append(stopped, service)
		}
	}
	return stopped
}

var checkServiceRunning = func(service string) bool {
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "is-active", "--quiet", service)
		return cmd.Run() == nil
	}
	return true
}

func ServiceIsRunning(service string) bool {
	return checkServiceRunning(service)
}

func OpenPorts() ([]int, error) {
	portMap := make(map[int]bool)

	if tcpPath != "" {
		if err := parseNetFile(tcpPath, portMap); err != nil {
			return nil, err
		}
	}

	if tcp6Path != "" {
		if err := parseNetFile(tcp6Path, portMap); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	var openPorts []int
	for port := range portMap {
		openPorts = append(openPorts, port)
	}

	return openPorts, nil
}

func parseNetFile(filename string, portMap map[int]bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Scan()

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}

		// Field 1: local_address:port (hex)
		// Field 3: connection state (0A = listening)
		localAddr := fields[1]
		state := fields[3]

		// Only include listening sockets (state 0A)
		if state != "0A" {
			continue
		}

		parts := strings.Split(localAddr, ":")
		if len(parts) != 2 {
			continue
		}

		// Skip localhost (0100007F = 127.0.0.1, 00000000000000000000000001000000 = ::1)
		if parts[0] == "0100007F" || parts[0] == "00000000000000000000000001000000" {
			continue
		}

		port, err := strconv.ParseInt(parts[1], 16, 64)
		if err == nil {
			portMap[int(port)] = true
		}
	}
	return scanner.Err()
}
