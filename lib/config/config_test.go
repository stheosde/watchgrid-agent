package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"watchgrid.de/agent/lib/agent"
)

func TestConfigFileExists(t *testing.T) {
	if configFileExists("non-existent-file.json") {
		t.Error("expected configFileExists to return false for non-existent file")
	}

	tmpFile, err := os.CreateTemp("", "config-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if !configFileExists(tmpFile.Name()) {
		t.Error("expected configFileExists to return true for existing file")
	}
}

func TestReadConfigJSON(t *testing.T) {
	cfg := readConfigJSON("missing-config.json")
	if len(cfg.Ports) != 0 || len(cfg.Services) != 0 {
		t.Errorf("expected empty config for missing file, got %+v", cfg)
	}

	tmpFile, err := os.CreateTemp("", "config-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	expectedConfig := ConfigFile{
		Ports:            []int{80, 443, 8080},
		Services:         []string{"nginx", "postgresql"},
		DockerContainers: []string{"redis", "postgres"},
	}
	bytes, err := json.Marshal(expectedConfig)
	if err != nil {
		t.Fatalf("failed to marshal expected config: %v", err)
	}

	if _, err := tmpFile.Write(bytes); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg = readConfigJSON(tmpFile.Name())
	if len(cfg.Ports) != 3 || cfg.Ports[0] != 80 || cfg.Ports[1] != 443 || cfg.Ports[2] != 8080 {
		t.Errorf("expected Ports [80, 443, 8080], got %v", cfg.Ports)
	}
	if len(cfg.Services) != 2 || cfg.Services[0] != "nginx" || cfg.Services[1] != "postgresql" {
		t.Errorf("expected Services ['nginx', 'postgresql'], got %v", cfg.Services)
	}
	if len(cfg.DockerContainers) != 2 || cfg.DockerContainers[0] != "redis" || cfg.DockerContainers[1] != "postgres" {
		t.Errorf("expected DockerContainers ['redis', 'postgres'], got %v", cfg.DockerContainers)
	}
}

func TestContent(t *testing.T) {
	cfg := Content()
	_ = cfg
}

func TestPullFromServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agent/config" {
			t.Errorf("expected path /agent/config, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("agent_id") != "test-agent-123" {
			t.Errorf("expected agent_id test-agent-123, got %s", r.URL.Query().Get("agent_id"))
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token-xyz" {
			t.Errorf("expected Authorization header Bearer test-token-xyz, got %s", authHeader)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"agent": "test-agent-123",
			"reporting_interval_seconds": 60,
			"cpu_percent_alert": 90.0,
			"ports": "80, 443, 8080",
			"services": "nginx, postgresql",
			"docker_containers": "redis, postgres",
			"allowed_uptime_days": 120
		}`))
	}))
	defer server.Close()

	agent.AgentFile = &agent.AgentFileConfig{
		AgentID: "test-agent-123",
		Token:   "test-token-xyz",
		BaseURL: server.URL,
	}

	tmpFile, err := os.CreateTemp("", "config-pull-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	oldConfigPath := configPath
	configPath = tmpFile.Name()
	defer func() { configPath = oldConfigPath }()

	updateRequested, err := PullFromServer()
	if err != nil {
		t.Fatalf("PullFromServer failed: %v", err)
	}
	if updateRequested {
		t.Error("expected updateRequested to be false")
	}

	if len(Config.Ports) != 3 || Config.Ports[0] != 80 || Config.Ports[1] != 443 || Config.Ports[2] != 8080 {
		t.Errorf("expected Ports [80, 443, 8080], got %v", Config.Ports)
	}
	if len(Config.Services) != 2 || Config.Services[0] != "nginx" || Config.Services[1] != "postgresql" {
		t.Errorf("expected Services ['nginx', 'postgresql'], got %v", Config.Services)
	}
	if len(Config.DockerContainers) != 2 || Config.DockerContainers[0] != "redis" || Config.DockerContainers[1] != "postgres" {
		t.Errorf("expected DockerContainers ['redis', 'postgres'], got %v", Config.DockerContainers)
	}
	if Config.AllowedUptimeDays != 120 {
		t.Errorf("expected AllowedUptimeDays 120, got %d", Config.AllowedUptimeDays)
	}
	if Config.ReportingIntervalSeconds != 60 {
		t.Errorf("expected ReportingIntervalSeconds 60, got %d", Config.ReportingIntervalSeconds)
	}

	savedCfg := readConfigJSON(configPath)
	if len(savedCfg.Ports) != 3 || savedCfg.Ports[0] != 80 || savedCfg.Ports[1] != 443 || savedCfg.Ports[2] != 8080 {
		t.Errorf("expected saved Ports [80, 443, 8080], got %v", savedCfg.Ports)
	}
	if savedCfg.AllowedUptimeDays != 120 {
		t.Errorf("expected saved AllowedUptimeDays 120, got %d", savedCfg.AllowedUptimeDays)
	}
	if savedCfg.ReportingIntervalSeconds != 60 {
		t.Errorf("expected saved ReportingIntervalSeconds 60, got %d", savedCfg.ReportingIntervalSeconds)
	}
}

func TestPullFromServer_UpdateRequested(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"agent": "test-agent-123",
			"reporting_interval_seconds": 60,
			"update_requested": true,
			"ports": "",
			"services": ""
		}`))
	}))
	defer server.Close()

	agent.AgentFile = &agent.AgentFileConfig{
		AgentID: "test-agent-123",
		Token:   "test-token-xyz",
		BaseURL: server.URL,
	}

	tmpFile, err := os.CreateTemp("", "config-pull-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	oldConfigPath := configPath
	configPath = tmpFile.Name()
	defer func() { configPath = oldConfigPath }()

	updateRequested, err := PullFromServer()
	if err != nil {
		t.Fatalf("PullFromServer failed: %v", err)
	}
	if !updateRequested {
		t.Error("expected updateRequested to be true")
	}
}

func TestReadConfigJSON_EmptyAndInvalid(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "empty-config-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := readConfigJSON(tmpFile.Name())
	if len(cfg.Ports) != 0 || len(cfg.Services) != 0 {
		t.Errorf("expected empty config from empty file, got %+v", cfg)
	}

	tmpFileInvalid, err := os.CreateTemp("", "invalid-config-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFileInvalid.Name())
	if _, err := tmpFileInvalid.WriteString("invalid json data here"); err != nil {
		t.Fatalf("failed to write to invalid temp file: %v", err)
	}
	tmpFileInvalid.Close()

	cfgInvalid := readConfigJSON(tmpFileInvalid.Name())
	if len(cfgInvalid.Ports) != 0 || len(cfgInvalid.Services) != 0 {
		t.Errorf("expected empty config from invalid file, got %+v", cfgInvalid)
	}
}
