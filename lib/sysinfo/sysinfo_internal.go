package sysinfo

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func OperatingSystem() string {
	switch runtime.GOOS {
	case "linux":
		return "Linux"
	case "darwin":
		return "Darwin"
	case "windows":
		return "Windows"
	default:
		return strings.Title(runtime.GOOS)
	}
}

func Architecture() string {
	return runtime.GOARCH
}

func KernelVersion() string {
	version := kernelVersion()
	if version != "Unknown" {
		return version
	}
	if kernelInfo, err := exec.Command("uname", "-r").Output(); err == nil {
		return strings.TrimSpace(string(kernelInfo))
	}
	return "Unknown"
}

func Hostname() string {
	host, err := os.Hostname()
	if err != nil {
		log.Printf("Hostname collection failed: %s\n", err)
		return "Unknown"
	}
	return host
}

func UptimeSeconds() int64 {
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		log.Printf("Uptime collection failed: %s\n", err)
		return 0
	}
	fields := strings.Fields(string(content))
	if len(fields) < 1 {
		return 0
	}
	uptimeSeconds, _ := strconv.ParseFloat(fields[0], 64)
	return int64(uptimeSeconds)
}


func CPUCores() int {
	return runtime.NumCPU()
}

func TotalRAMGB() float64 {
	// For macOS/Darwin: sysctl -n hw.memsize
	if OperatingSystem() == "Darwin" {
		out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err == nil {
			bytes, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
			if err == nil {
				return bytes / (1024.0 * 1024.0 * 1024.0)
			}
		}
	}

	// For Linux: read /proc/meminfo
	content, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					kb, err := strconv.ParseFloat(fields[1], 64)
					if err == nil {
						return kb / (1024.0 * 1024.0)
					}
				}
			}
		}
	}

	// Fallback using free -b command
	out, err := exec.Command("sh", "-c", "free -b | grep Mem | awk '{print $2}'").Output()
	if err == nil {
		bytes, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		if err == nil {
			return bytes / (1024.0 * 1024.0 * 1024.0)
		}
	}

	return 0.0
}

func TotalDiskGB() float64 {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return 0.0
	}
	totalBytes := float64(stat.Blocks) * float64(stat.Bsize)
	return totalBytes / (1024.0 * 1024.0 * 1024.0)
}
