package agent

type AgentConfig struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type AgentFileConfig struct {
	AgentID string         `json:"agent_id"`
	Token   string         `json:"token"`
	BaseURL string         `json:"base_url"`
	Extra   map[string]any `json:"-"`
}
