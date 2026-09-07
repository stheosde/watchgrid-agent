package sysinfo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"watchgrid.de/agent/lib/agent"
	"watchgrid.de/agent/lib/ports"
)

const (
	retryInterval = 10 // seconds
	totalRetries  = 5
)

var (
	SysInfo SystemInformation
)

func Collect() (*SystemInformation, error) {
	agentInfo := agent.Info()
	sysInfo := &SystemInformation{
		Agent:                    *agentInfo,
		Hostname:                 Hostname(),
		OperatingSystem:          OperatingSystem(),
		Architecture:             Architecture(),
		KernelVersion:            KernelVersion(),
		UptimeSeconds:            UptimeSeconds(),
		CPUCores:                 CPUCores(),
		RAMGB:                    TotalRAMGB(),
		DiskGB:                   TotalDiskGB(),
		DetectedDockerContainers: ports.RunningDockerContainers(),
	}
	return sysInfo, nil
}

func Print(sysinfo *SystemInformation) {
	fmt.Printf("agent{id:%s version:%s} host:%s os:%s version:%s arch:%s uptime:%ds cpu_cores:%d ram_gb:%.1f disk_gb:%.1f\n",
		sysinfo.Agent.ID,
		sysinfo.Agent.Version,
		sysinfo.Hostname,
		sysinfo.OperatingSystem,
		sysinfo.KernelVersion,
		sysinfo.Architecture,
		sysinfo.UptimeSeconds,
		sysinfo.CPUCores,
		sysinfo.RAMGB,
		sysinfo.DiskGB,
	)
}

func Send(sysinfo *SystemInformation) error {
	agentConfig := agent.Config()

	jsonData, err := json.Marshal(sysinfo)
	if err != nil {
		errorMessage := fmt.Sprintf("Error serializing sysinfo to JSON: %s\n", err)
		return errors.New(errorMessage)
	}
	log.Printf("Data: %s\n", string(jsonData))
	bodyReader := bytes.NewReader(jsonData)
	req, err := http.NewRequest("POST", agentConfig.BaseURL+"/agent/info", bodyReader)
	if err != nil {
		return errors.New("Error creating HTTP request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+agentConfig.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	for i := range [totalRetries]int{} {
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending sysinfo to server: %v\n", err)
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

		log.Printf("System Information sent successfully: [%d] %s\n", resp.StatusCode, serverResponseBody.String())
		break
	}
	return nil
}

func CollectAndPrint() {
	sysInfo, err := Collect()
	if err != nil {
		log.Fatalf("Error during collection of system information: %s\n", err)
	}
	Print(sysInfo)
}

func CollectAndSend() error {
	sysInfo, err := Collect()
	if err != nil {
		errorMessage := fmt.Sprintf("System information could not be collected: %s\n", err)
		return errors.New(errorMessage)
	}
	if err = Send(sysInfo); err != nil {
		errorMessage := fmt.Sprintf("System information could not be send: %s\n", err)
		return errors.New(errorMessage)
	}
	return nil
}
