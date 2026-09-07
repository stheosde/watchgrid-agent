package docker

import (
	"errors"
	"reflect"
	"testing"
)

func TestContainerIsRunningAndCheckStoppedDockerContainers(t *testing.T) {
	oldCheck := CheckContainerRunning
	CheckContainerRunning = func(container string) bool {
		if container == "stopped-app" {
			return false
		}
		return true
	}
	defer func() { CheckContainerRunning = oldCheck }()

	if ContainerIsRunning("stopped-app") {
		t.Error("expected stopped-app ContainerIsRunning to return false")
	}
	if !ContainerIsRunning("running-app") {
		t.Error("expected running-app ContainerIsRunning to return true")
	}

	stopped := CheckStoppedDockerContainers([]string{"running-app", "redis"})
	if len(stopped) != 0 {
		t.Errorf("expected empty stopped list, got %v", stopped)
	}

	stopped = CheckStoppedDockerContainers([]string{"running-app", "stopped-app", "redis"})
	if len(stopped) != 1 || stopped[0] != "stopped-app" {
		t.Errorf("expected ['stopped-app'], got %v", stopped)
	}
}

func TestRunningDockerContainers(t *testing.T) {
	oldList := ListRunningContainers
	defer func() { ListRunningContainers = oldList }()

	// Success case
	ListRunningContainers = func() ([]string, error) {
		return []string{"web", "db", "redis"}, nil
	}

	containers := RunningDockerContainers()
	expected := []string{"web", "db", "redis"}
	if !reflect.DeepEqual(containers, expected) {
		t.Errorf("expected %v, got %v", expected, containers)
	}

	// Error case
	ListRunningContainers = func() ([]string, error) {
		return nil, errors.New("docker daemon not running")
	}

	containers = RunningDockerContainers()
	if len(containers) != 0 {
		t.Errorf("expected empty list on error, got %v", containers)
	}
}
