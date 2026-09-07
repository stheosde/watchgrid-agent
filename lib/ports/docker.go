package ports

import (
	"os/exec"
	"strings"
)

var checkContainerRunning = func(container string) bool {
	if _, err := exec.LookPath("docker"); err == nil {
		cmd := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", container)
		out, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.TrimSpace(string(out)) == "true"
	}
	return false
}

func ContainerIsRunning(container string) bool {
	return checkContainerRunning(container)
}

func CheckStoppedDockerContainers(containersToCheck []string) []string {
	var stopped []string
	for _, container := range containersToCheck {
		if !ContainerIsRunning(container) {
			stopped = append(stopped, container)
		}
	}
	return stopped
}

var listRunningContainers = func() ([]string, error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, err
	}
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var containers []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			containers = append(containers, trimmed)
		}
	}
	return containers, nil
}

func RunningDockerContainers() []string {
	containers, err := listRunningContainers()
	if err != nil {
		return []string{}
	}
	return containers
}
