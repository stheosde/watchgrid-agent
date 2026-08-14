//go:build darwin

package sysinfo

import (
	"syscall"
)

func kernelVersion() string {
	if release, err := syscall.Sysctl("kern.osrelease"); err == nil {
		return release
	}
	return "Unknown"
}
