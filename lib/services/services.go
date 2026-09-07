package services

import (
	"os/exec"
)

var CheckServiceRunning = func(service string) bool {
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "is-active", "--quiet", service)
		return cmd.Run() == nil
	}
	return true
}

func ServiceIsRunning(service string) bool {
	return CheckServiceRunning(service)
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
