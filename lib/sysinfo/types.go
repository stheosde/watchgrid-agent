package sysinfo

import (
	"watchgrid.de/agent/lib/agent"
)

type SystemInformation struct {
	Agent           agent.AgentConfig `json:"agent"`
	Hostname        string            `json:"hostname"`
	OperatingSystem string            `json:"os"`
	Architecture    string            `json:"arch"`
	KernelVersion   string            `json:"version"`
	UptimeSeconds   int64             `json:"uptime"`
	CPUCores                 int               `json:"cpu_cores"`
	RAMGB                    float64           `json:"ram_gb"`
	DiskGB                   float64           `json:"disk_gb"`
	DetectedDockerContainers []string          `json:"detected_docker_containers,omitempty"`
}
