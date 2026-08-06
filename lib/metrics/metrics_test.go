package metrics

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"watchgrid.de/agent/lib/agent"
	"watchgrid.de/agent/lib/config"
)

func init() {
	config.Config = config.ConfigFile{
		Ports:    []int{22, 80, 443},
		Services: []string{"nginx"},
	}
}

func TestCollect(t *testing.T) {
	agent.Agent = &agent.AgentConfig{
		ID:      "test-agent-metrics",
		Version: "1.0.0-metrics",
	}

	snapshot, err := Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if snapshot.Agent.ID != "test-agent-metrics" {
		t.Errorf("expected Agent ID 'test-agent-metrics', got %q", snapshot.Agent.ID)
	}
	if snapshot.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestPrint(t *testing.T) {
	snapshot := &MetricSnapshot{
		Agent: agent.AgentConfig{
			ID:      "test-agent",
			Version: "1.0.0",
		},
		Usage: Usages{
			CPUPercent:    12.34,
			MemoryPercent: 56.78,
			DiskPercent:   90.12,
		},
		Timestamp: "2026-06-29 19:15:31 +0200",
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Print(snapshot)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expected := "agent{id:test-agent version:1.0.0} usage{cpu:12.34% memory:56.78% disk:90.12%} ports:ok services:ok timestamp:2026-06-29 19:15:31 +0200\n"
	if output != expected {
		t.Errorf("Print output expected %q, got %q", expected, output)
	}
}

func TestSend_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %q", r.Method)
		}
		if r.URL.Path != "/agent/metrics" {
			t.Errorf("expected path '/agent/metrics', got %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer mock-metrics-token" {
			t.Errorf("expected Auth token 'Bearer mock-metrics-token', got %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", r.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var received MetricSnapshot
		if err := json.Unmarshal(body, &received); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if received.Usage.CPUPercent != 45.67 {
			t.Errorf("expected CPUPercent 45.67, got %f", received.Usage.CPUPercent)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	agent.AgentFile = &agent.AgentFileConfig{
		BaseURL: server.URL,
		Token:   "mock-metrics-token",
	}

	snapshot := &MetricSnapshot{
		Agent: agent.AgentConfig{
			ID:      "test-agent",
			Version: "1.0.0",
		},
		Usage: Usages{
			CPUPercent: 45.67,
		},
	}

	err := Send(snapshot)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
}

func TestSend_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer server.Close()

	agent.AgentFile = &agent.AgentFileConfig{
		BaseURL: server.URL,
		Token:   "mock-metrics-token",
	}

	snapshot := &MetricSnapshot{
		Usage: Usages{
			CPUPercent: 45.67,
		},
	}

	err := Send(snapshot)
	if err == nil {
		t.Error("expected Send to return error on server error")
	}
}

func TestCollectAndSend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	agent.AgentFile = &agent.AgentFileConfig{
		BaseURL: server.URL,
		Token:   "mock-metrics-token",
	}
	agent.Agent = &agent.AgentConfig{
		ID:      "test-agent-metrics",
		Version: "1.0.0-metrics",
	}

	err := CollectAndSend()
	if err != nil {
		t.Fatalf("CollectAndSend failed: %v", err)
	}
}

func TestCollectAndPrint(t *testing.T) {
	oldFlags := log.Flags()
	oldOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer func() {
		log.SetFlags(oldFlags)
		log.SetOutput(oldOutput)
	}()

	agent.Agent = &agent.AgentConfig{
		ID:      "test-agent-metrics",
		Version: "1.0.0-metrics",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	CollectAndPrint()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	if output == "" {
		t.Error("expected CollectAndPrint output to not be empty")
	}
}


func TestInternalMetricsGathering(t *testing.T) {
	_ = CPUPercent()
	_ = MemoryPercent()
	_ = DiskPercent()
}

func TestCollect_NoPortsOrServices(t *testing.T) {
	oldConfig := config.Config
	defer func() { config.Config = oldConfig }()

	config.Config = config.ConfigFile{
		Ports:    nil,
		Services: nil,
	}

	snapshot, err := Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if snapshot.Ports.Error {
		t.Error("expected Ports.Error to be false when ports is nil/empty")
	}
	if snapshot.Services.Error {
		t.Error("expected Services.Error to be false when services is nil/empty")
	}
}

func TestPrint_NoPortsOrServices(t *testing.T) {
	oldConfig := config.Config
	defer func() { config.Config = oldConfig }()

	config.Config = config.ConfigFile{
		Ports:    nil,
		Services: nil,
	}

	snapshot := &MetricSnapshot{
		Agent: agent.AgentConfig{
			ID:      "test-agent",
			Version: "1.0.0",
		},
		Usage: Usages{
			CPUPercent:    12.34,
			MemoryPercent: 56.78,
			DiskPercent:   90.12,
		},
		Timestamp: "2026-06-29 19:15:31 +0200",
	}

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	Print(snapshot)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expected := "agent{id:test-agent version:1.0.0} usage{cpu:12.34% memory:56.78% disk:90.12%} timestamp:2026-06-29 19:15:31 +0200\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
