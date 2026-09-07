package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"watchgrid.de/agent/lib/agent"
)

var (
	Config     ConfigFile
	Debug      = os.Getenv("DEBUG") == "true"
	configPath string
)

func init() {
	basePath := ""
	if Debug {
		basePath = "./config"
	} else {
		basePath = "/etc/watchgrid"
	}
	configPath = basePath + "/config.json"
	Config = readConfigJSON(configPath)
}

func PullFromServer() (bool, error) {
	agentFile := agent.Config()
	if agentFile == nil || agentFile.BaseURL == "" || agentFile.Token == "" || agentFile.AgentID == "" {
		return false, fmt.Errorf("agent configuration not initialized or incomplete")
	}

	url := fmt.Sprintf("%s/agent/config?agent_id=%s", agentFile.BaseURL, agentFile.AgentID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+agentFile.Token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiConfig ServerConfig

	if err := json.NewDecoder(resp.Body).Decode(&apiConfig); err != nil {
		return false, fmt.Errorf("error decoding response: %w", err)
	}

	// Parse ports (comma-separated string of ints)
	var ports []int
	if apiConfig.Ports != "" {
		parts := strings.Split(apiConfig.Ports, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			p, err := strconv.Atoi(part)
			if err != nil {
				log.Printf("Warning: invalid port %q in configuration: %v\n", part, err)
				continue
			}
			ports = append(ports, p)
		}
	}

	// Parse services (comma-separated string)
	var services []string
	if apiConfig.Services != "" {
		parts := strings.Split(apiConfig.Services, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			services = append(services, part)
		}
	}

	// Parse docker containers (comma-separated string)
	var dockerContainers []string
	if apiConfig.DockerContainers != "" {
		parts := strings.Split(apiConfig.DockerContainers, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			dockerContainers = append(dockerContainers, part)
		}
	}

	Config.Ports = ports
	Config.Services = services
	Config.DockerContainers = dockerContainers
	Config.AllowedUptimeDays = apiConfig.AllowedUptimeDays
	Config.ReportingIntervalSeconds = apiConfig.ReportingIntervalSeconds

	updatedJSON, err := json.MarshalIndent(Config, "", "    ")
	if err != nil {
		return false, fmt.Errorf("error marshaling config to JSON: %w", err)
	}

	if err := os.WriteFile(configPath, updatedJSON, 0644); err != nil {
		return false, fmt.Errorf("error writing config file: %w", err)
	}

	log.Printf("Successfully pulled and updated configuration from server\n")
	return apiConfig.UpdateRequested, nil
}

// TODO: config file not needed in default case
func configFileExists(configPath string) bool {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Monitoring configuration file not found at %s\n", configPath)
		return false
	}
	return true
}

func readConfigJSON(configPath string) ConfigFile {
	if !configFileExists(configPath) {
		return ConfigFile{}
	}

	plan, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Warning: error reading config file %s: %v. Using empty configuration.\n", configPath, err)
		return ConfigFile{}
	}

	trimmed := strings.TrimSpace(string(plan))
	if trimmed == "" || trimmed == "{}" {
		return ConfigFile{}
	}

	var data ConfigFile
	var raw map[string]any

	if err := json.Unmarshal(plan, &raw); err != nil {
		log.Printf("Warning: error parsing raw config file %s: %v. Using empty configuration.\n", configPath, err)
		return ConfigFile{}
	}

	if err := json.Unmarshal(plan, &data); err != nil {
		log.Printf("Warning: error parsing data config file %s: %v. Using empty configuration.\n", configPath, err)
		return ConfigFile{}
	}
	return data
}

func Content() ConfigFile {
	return Config
}
