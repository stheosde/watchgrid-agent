//go:build linux

package sysinfo

import (
	"os"
	"strings"
)

func kernelVersion() string {
	if release, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		return strings.TrimSpace(string(release))
	}
	return "Unknown"
}
