package metrics

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func readCPUStats() (idle, total float64, err error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)[1:] // skip "cpu"
			var totalTime float64
			var idleTime float64
			for i, valStr := range fields {
				val, err := strconv.ParseFloat(valStr, 64)
				if err != nil {
					continue
				}
				totalTime += val
				// idle is the 4th field (index 3) and iowait is the 5th (index 4)
				if i == 3 || i == 4 {
					idleTime += val
				}
			}
			return idleTime, totalTime, nil
		}
	}
	return 0, 0, fmt.Errorf("failed to parse cpu line")
}

func cpuPercentFallback() float64 {
	usage, err := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | awk '{print $2 + $4}'").Output()
	if err != nil {
		return 0.0
	}
	usageStr := strings.TrimSpace(string(usage))
	usageFloat, _ := strconv.ParseFloat(usageStr, 64)
	return usageFloat
}

func CPUPercent() float64 {
	idle1, total1, err := readCPUStats()
	if err != nil {
		return cpuPercentFallback()
	}

	time.Sleep(200 * time.Millisecond)

	idle2, total2, err := readCPUStats()
	if err != nil {
		return 0.0
	}

	totalDiff := total2 - total1
	if totalDiff <= 0 {
		return 0.0
	}
	idleDiff := idle2 - idle1
	return (1.0 - (idleDiff / totalDiff)) * 100.0
}

func readMeminfoStats() (total, avail float64, err error) {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	lines := strings.Split(string(content), "\n")
	var memTotal, memAvail float64
	var foundTotal, foundAvail bool
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				memTotal, _ = strconv.ParseFloat(fields[1], 64)
				foundTotal = true
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				memAvail, _ = strconv.ParseFloat(fields[1], 64)
				foundAvail = true
			}
		}
	}
	// Fallback for older Linux kernels where MemAvailable is not populated
	if foundTotal && !foundAvail {
		var memFree, buffers, cached float64
		for _, line := range lines {
			if strings.HasPrefix(line, "MemFree:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					memFree, _ = strconv.ParseFloat(fields[1], 64)
				}
			} else if strings.HasPrefix(line, "Buffers:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					buffers, _ = strconv.ParseFloat(fields[1], 64)
				}
			} else if strings.HasPrefix(line, "Cached:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					cached, _ = strconv.ParseFloat(fields[1], 64)
				}
			}
		}
		memAvail = memFree + buffers + cached
		foundAvail = true
	}
	if foundTotal && foundAvail {
		return memTotal, memAvail, nil
	}
	return 0, 0, fmt.Errorf("failed to parse memory info")
}

func memoryPercentFallback() float64 {
	usage, _ := exec.Command("sh", "-c", "free | grep Mem | awk '{print $3/$2 * 100.0}'").Output()
	usageStr := strings.TrimSpace(string(usage))
	usageFloat, _ := strconv.ParseFloat(usageStr, 64)
	return usageFloat
}

func MemoryPercent() float64 {
	memTotal, memAvail, err := readMeminfoStats()
	if err == nil && memTotal > 0 {
		return ((memTotal - memAvail) / memTotal) * 100.0
	}
	return memoryPercentFallback()
}

func DiskPercent() float64 {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return 0.0
	}
	total := float64(stat.Blocks)
	free := float64(stat.Bfree)
	if total == 0 {
		return 0.0
	}
	return ((total - free) / total) * 100.0
}
