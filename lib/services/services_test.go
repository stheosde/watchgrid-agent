package services

import (
	"testing"
)

func TestCheckStoppedServicesAndServiceIsRunning(t *testing.T) {
	oldCheck := CheckServiceRunning
	CheckServiceRunning = func(service string) bool {
		if service == "docker" {
			return false
		}
		return true
	}
	defer func() { CheckServiceRunning = oldCheck }()

	if ServiceIsRunning("docker") {
		t.Error("expected docker ServiceIsRunning to return false")
	}
	if !ServiceIsRunning("nginx") {
		t.Error("expected nginx ServiceIsRunning to return true")
	}

	stopped := CheckStoppedServices([]string{"nginx", "systemd"})
	if len(stopped) != 0 {
		t.Errorf("expected empty, got %v", stopped)
	}

	stopped = CheckStoppedServices([]string{"nginx", "docker", "systemd"})
	if len(stopped) != 1 || stopped[0] != "docker" {
		t.Errorf("expected [docker], got %v", stopped)
	}
}
