//go:build !linux && !darwin

package sysinfo

func kernelVersion() string {
	return "Unknown"
}
