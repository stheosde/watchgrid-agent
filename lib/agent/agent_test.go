package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInitialize(t *testing.T) {
	oldDebug := Debug
	Debug = true
	defer func() { Debug = oldDebug }()

	configDir := "./config"
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "agent.json")

	var backupContent []byte
	backupExists := false
	if _, err := os.Stat(configPath); err == nil {
		backupContent, err = os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read existing agent.json for backup: %v", err)
		}
		backupExists = true
	}

	defer func() {
		if backupExists {
			err = os.WriteFile(configPath, backupContent, 0644)
			if err != nil {
				t.Errorf("failed to restore agent.json backup: %v", err)
			}
		} else {
			os.Remove(configPath)
		}
	}()

	testConfig := AgentFileConfig{
		AgentID: "test-agent-123",
		Token:   "test-token",
		BaseURL: "http://localhost:8080",
	}
	configBytes, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}

	err = os.WriteFile(configPath, configBytes, 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	Initialize("1.0.0-test")

	cfg := Config()
	if cfg.AgentID != testConfig.AgentID {
		t.Errorf("expected AgentID %q, got %q", testConfig.AgentID, cfg.AgentID)
	}
	if cfg.Token != testConfig.Token {
		t.Errorf("expected Token %q, got %q", testConfig.Token, cfg.Token)
	}
	if cfg.BaseURL != testConfig.BaseURL {
		t.Errorf("expected BaseURL %q, got %q", testConfig.BaseURL, cfg.BaseURL)
	}

	info := Info()
	if info.ID != testConfig.AgentID {
		t.Errorf("expected Info ID %q, got %q", testConfig.AgentID, info.ID)
	}
	if info.Version != "1.0.0-test" {
		t.Errorf("expected Info Version %q, got %q", "1.0.0-test", info.Version)
	}
}

func TestInitialize_MissingConfig(t *testing.T) {
	oldDebug := Debug
	Debug = true
	defer func() { Debug = oldDebug }()

	configDir := "./config"
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "agent.json")

	var backupContent []byte
	backupExists := false
	if _, err := os.Stat(configPath); err == nil {
		backupContent, err = os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read existing agent.json for backup: %v", err)
		}
		backupExists = true
		err = os.Remove(configPath)
		if err != nil {
			t.Fatalf("failed to temporarily remove agent.json: %v", err)
		}
	}

	defer func() {
		if backupExists {
			err = os.WriteFile(configPath, backupContent, 0644)
			if err != nil {
				t.Errorf("failed to restore agent.json backup: %v", err)
			}
		}
	}()

	Initialize("2.0.0-test")

	cfg := Config()
	if cfg.AgentID != "" {
		t.Errorf("expected empty AgentID, got %q", cfg.AgentID)
	}

	info := Info()
	if info.Version != "2.0.0-test" {
		t.Errorf("expected Info Version %q, got %q", "2.0.0-test", info.Version)
	}
}
