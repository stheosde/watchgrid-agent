package ports

type PortStatus struct {
	Error    bool  `json:"errors"`
	Affected []int `json:"affected"`
}
