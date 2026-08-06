package sysinfo

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"watchgrid.de/agent/lib/agent"
)

func TestCollect(t *testing.T) {
	agent.Agent = &agent.AgentConfig{
		ID:      "test-agent-sysinfo",
		Version: "1.0.0-sysinfo",
	}

	info, err := Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if info.Agent.ID != "test-agent-sysinfo" {
		t.Errorf("expected Agent ID 'test-agent-sysinfo', got %q", info.Agent.ID)
	}
	if info.Hostname == "" {
		t.Error("expected non-empty hostname")
	}
	if info.OperatingSystem == "" {
		t.Error("expected non-empty operating system")
	}
	if info.CPUCores <= 0 {
		t.Errorf("expected positive CPUCores, got %d", info.CPUCores)
	}
}

func TestSend_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %q", r.Method)
		}
		if r.URL.Path != "/agent/info" {
			t.Errorf("expected path '/agent/info', got %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer mock-token" {
			t.Errorf("expected Auth token 'Bearer mock-token', got %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", r.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var received SystemInformation
		if err := json.Unmarshal(body, &received); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if received.Hostname != "test-host" {
			t.Errorf("expected hostname 'test-host', got %q", received.Hostname)
		}
		if received.CPUCores != 4 {
			t.Errorf("expected cpu_cores 4, got %d", received.CPUCores)
		}
		if received.RAMGB != 8.0 {
			t.Errorf("expected ram_gb 8.0, got %f", received.RAMGB)
		}
		if received.DiskGB != 120.0 {
			t.Errorf("expected disk_gb 120.0, got %f", received.DiskGB)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	agent.AgentFile = &agent.AgentFileConfig{
		BaseURL: server.URL,
		Token:   "mock-token",
	}

	sysInfo := &SystemInformation{
		Agent: agent.AgentConfig{
			ID:      "test-agent",
			Version: "1.0.0",
		},
		Hostname:        "test-host",
		OperatingSystem: "Linux",
		Architecture:    "amd64",
		KernelVersion:   "5.4.0",
		UptimeSeconds:   3600,
		CPUCores:        4,
		RAMGB:           8.0,
		DiskGB:          120.0,
	}

	err := Send(sysInfo)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
}

func TestSend_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	agent.AgentFile = &agent.AgentFileConfig{
		BaseURL: server.URL,
		Token:   "mock-token",
	}

	sysInfo := &SystemInformation{
		Hostname: "test-host",
	}

	err := Send(sysInfo)
	if err == nil {
		t.Error("expected Send to return error when server returns 500 status code")
	}
}

func TestInternalOSCollection(t *testing.T) {
	osName := OperatingSystem()
	if osName == "" {
		t.Error("expected non-empty operating system from OperatingSystem()")
	}

	arch := Architecture()
	if arch == "" {
		t.Error("expected non-empty architecture from Architecture()")
	}

	kernel := KernelVersion()
	if kernel == "" {
		t.Error("expected non-empty kernel version from KernelVersion()")
	}

	host := Hostname()
	if host == "" {
		t.Error("expected non-empty hostname from Hostname()")
	}

	// UptimeSeconds might return 0 on systems without /proc/uptime (like macOS),
	// which is the expected fallback behavior. We just verify it doesn't panic.
	_ = UptimeSeconds()
}
