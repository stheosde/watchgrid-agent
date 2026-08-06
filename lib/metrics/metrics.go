package metrics

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"watchgrid.de/agent/lib/agent"
	"watchgrid.de/agent/lib/config"
	"watchgrid.de/agent/lib/ports"
)

const (
	retryInterval = 10 // seconds
	totalRetries  = 5
)

var (
	Metric    MetricSnapshot
	AgentFile agent.AgentFileConfig
)

func Collect() (*MetricSnapshot, error) {
	cfg := config.Content()

	var portStatus PortStatus
	if len(cfg.Ports) > 0 {
		nonWhitelistedPorts, err := ports.CheckNonWhitelistedPorts(cfg.Ports)
		portStatus = PortStatus{
			Error:    err != nil || len(nonWhitelistedPorts) > 0,
			Affected: nonWhitelistedPorts,
		}
	} else {
		portStatus = PortStatus{
			Error: false,
		}
	}

	var serviceStatus ServiceStatus
	if len(cfg.Services) > 0 {
		stoppedServices := ports.CheckStoppedServices(cfg.Services)
		serviceStatus = ServiceStatus{
			Error:    len(stoppedServices) > 0,
			Affected: stoppedServices,
		}
	} else {
		serviceStatus = ServiceStatus{
			Error: false,
		}
	}

	snapshot := &MetricSnapshot{
		Agent: *agent.Info(),
		Usage: Usages{
			CPUPercent:    CPUPercent(),
			MemoryPercent: MemoryPercent(),
			DiskPercent:   DiskPercent(),
		},
		Ports:     portStatus,
		Services:  serviceStatus,
		Timestamp: time.Now().Format("2006-01-02 15:04:05 -0700"),
	}
	return snapshot, nil
}

func Print(metric *MetricSnapshot) {
	cfg := config.Content()
	var parts []string

	parts = append(parts, fmt.Sprintf("agent{id:%s version:%s}", metric.Agent.ID, metric.Agent.Version))
	parts = append(parts, fmt.Sprintf("usage{cpu:%.2f%% memory:%.2f%% disk:%.2f%%}", metric.Usage.CPUPercent, metric.Usage.MemoryPercent, metric.Usage.DiskPercent))

	if len(cfg.Ports) > 0 {
		if metric.Ports.Error {
			parts = append(parts, fmt.Sprintf("ports:error[affected:%v]", metric.Ports.Affected))
		} else {
			parts = append(parts, "ports:ok")
		}
	}

	if len(cfg.Services) > 0 {
		if metric.Services.Error {
			parts = append(parts, fmt.Sprintf("services:error[affected:%v]", metric.Services.Affected))
		} else {
			parts = append(parts, "services:ok")
		}
	}

	parts = append(parts, fmt.Sprintf("timestamp:%s", metric.Timestamp))

	fmt.Println(strings.Join(parts, " "))
}

func Send(metric *MetricSnapshot) error {
	agentConfig := agent.Config()

	jsonData, err := json.Marshal(metric)
	if err != nil {
		errorMessage := fmt.Sprintf("Error serializing metrics to JSON: %s\n", err)
		return errors.New(errorMessage)
	}
	log.Printf("Data: %s\n", string(jsonData))
	bodyReader := bytes.NewReader(jsonData)
	req, err := http.NewRequest("POST", agentConfig.BaseURL+"/agent/metrics", bodyReader)
	if err != nil {
		return errors.New("Error creating HTTP request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+agentConfig.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	for i := range [totalRetries]int{} {
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending metrics to server: %v\n", err)
			log.Printf("Trying again in %d seconds...\n", retryInterval)
			time.Sleep(retryInterval * time.Second)
			if i == totalRetries-1 {
				errorMessage := fmt.Sprintf("Failed to send sysinfo after %d attempts. Exiting.\n", totalRetries)
				return errors.New(errorMessage)
			}
			continue
		}
		defer resp.Body.Close()

		serverResponseBody := new(bytes.Buffer)
		serverResponseBody.ReadFrom(resp.Body)

		if resp.StatusCode != 200 {
			errorMessage := fmt.Sprintf("Error response from server: [%d] %s\n", resp.StatusCode, serverResponseBody.String())
			return errors.New(errorMessage)
		}

		log.Printf("Metrics sent successfully: [%d] %s\n", resp.StatusCode, serverResponseBody.String())
		break
	}
	return nil
}

func CollectAndPrint() {
	metric, err := Collect()
	if err != nil {
		log.Fatalf("Error during collection of system metrics: %s\n", err)
	}
	Print(metric)
}

func CollectAndSend() error {
	metric, err := Collect()
	if err != nil {
		errorMessage := fmt.Sprintf("System metrics could not be collected: %s\n", err)
		return errors.New(errorMessage)
	}
	if err = Send(metric); err != nil {
		errorMessage := fmt.Sprintf("System metrics could not be send: %s\n", err)
		return errors.New(errorMessage)
	}
	return nil
}

