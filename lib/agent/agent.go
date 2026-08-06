package agent

import (
	"encoding/json"
	"log"
	"os"
)

var (
	AgentFile *AgentFileConfig
	Agent     *AgentConfig
	Debug     = os.Getenv("DEBUG") == "true"
)

func Initialize(version string) {
	basePath := ""
	if Debug {
		basePath = "./config"
	} else {
		basePath = "/etc/watchgrid"
	}
	agentConfigPath := basePath + "/agent.json"
	AgentFile = readConfigJSON(agentConfigPath)
	Agent = &AgentConfig{
		ID:      AgentFile.AgentID,
		Version: version,
	}
}

func agentFileExists(configPath string) bool {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Agent configuration file not found at %s\n", configPath)
		return false
	}
	return true
}

func readConfigJSON(configPath string) *AgentFileConfig {
	if !agentFileExists(configPath) {
		return &AgentFileConfig{}
	}

	plan, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %s\n", err)
	}

	var data *AgentFileConfig
	var raw map[string]any

	if err := json.Unmarshal(plan, &raw); err != nil {
		log.Fatalf("Error parsing raw agent file: %s\n", err)
	}

	if err := json.Unmarshal(plan, &data); err != nil {
		log.Fatalf("Error parsing data agent file: %s\n", err)
	}
	return data
}

func Config() *AgentFileConfig {
	return AgentFile
}

func Info() *AgentConfig {
	return Agent
}
