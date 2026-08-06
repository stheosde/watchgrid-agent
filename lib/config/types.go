package config

type ConfigFile struct {
	Ports                    []int          `json:"ports"`
	Services                 []string       `json:"services"`
	AllowedUptimeDays        int            `json:"allowed_uptime_days"`
	ReportingIntervalSeconds int            `json:"reporting_interval_seconds"`
	Extra                    map[string]any `json:"-"`
}

type ServerConfig struct {
	Ports                    string `json:"ports"`
	Services                 string `json:"services"`
	AllowedUptimeDays        int    `json:"allowed_uptime_days"`
	UpdateRequested          bool   `json:"update_requested"`
	ReportingIntervalSeconds int    `json:"reporting_interval_seconds"`
}
