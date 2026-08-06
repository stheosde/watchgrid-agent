package metrics

import (
	"watchgrid.de/agent/lib/agent"
	//"watchgrid.de/agent/lib/ports"
)

type PortStatus struct {
	Error    bool  `json:"error"`
	Affected []int `json:"affected,omitempty"`
}

type ServiceStatus struct {
	Error    bool     `json:"error"`
	Affected []string `json:"affected,omitempty"`
}

type MetricSnapshot struct {
	Agent     agent.AgentConfig `json:"agent"`
	Usage     Usages            `json:"usage"`
	Timestamp string            `json:"timestamp"`
	Ports     PortStatus        `json:"ports"`
	Services  ServiceStatus     `json:"services"`
}

type Usages struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
}
