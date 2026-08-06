package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"watchgrid.de/agent/lib"
	"watchgrid.de/agent/lib/agent"
	"watchgrid.de/agent/lib/config"
	"watchgrid.de/agent/lib/metrics"
	"watchgrid.de/agent/lib/sysinfo"
)

var (
	Version = "unknown"
)

func main() {
	checkFlag := flag.Bool("check", false, "Check System status without transmission")
	helpFlag := flag.Bool("help", false, "Show help message")
	infoFlag := flag.Bool("info", false, "Check System information without transmission")
	infoSendFlag := flag.Bool("isend", false, "Check System information and transmit to server")
	metricsFlag := flag.Bool("metrics", false, "Check System metrics without transmission")
	metricsSendFlag := flag.Bool("msend", false, "Check System metrics and transmit to server")
	updateFlag := flag.Bool("update", false, "Perform an update of the agent if available")
	versionFlag := flag.Bool("version", false, "Print version information and exit")

	agent.Initialize(Version)

	flag.Parse()
	if *versionFlag {
		printVersion()
		os.Exit(0)
	}

	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	if *checkFlag {
		os.Exit(0)
	}

	if *infoFlag {
		sysinfo.CollectAndPrint()
		os.Exit(0)
	}

	if *infoSendFlag {
		if err := sysinfo.CollectAndSend(); err != nil {
			log.Printf("Error in System Information:  %v\n", err)
		}
		os.Exit(0)
	}

	if *metricsFlag {
		metrics.CollectAndPrint()
		os.Exit(0)
	}

	if *metricsSendFlag {
		if err := metrics.CollectAndSend(); err != nil {
			log.Printf("Error in System Metrics:  %v\n", err)
		}
		os.Exit(0)
	}

	if *updateFlag {
		updated, err := lib.Update(Version)
		if err != nil {
			log.Fatalf("Update failed: %v\n", err)
		}
		if updated {
			log.Println("Agent updated successfully.")
		} else {
			log.Println("No update performed.")
		}
		os.Exit(0)
	}

	// Try to acquire lock to ensure only one daemon instance runs
	lockFile, err := os.OpenFile("/tmp/watchgrid.lock", os.O_CREATE|os.O_RDWR, 0666)
	if err == nil {
		defer lockFile.Close()
		err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			log.Fatalf("Another instance of WatchGrid agent daemon is already running.")
		}
	} else {
		log.Printf("Warning: Could not open lock file: %v. Running without instance checking.\n", err)
	}

	// Send static host information once at startup
	if err := sysinfo.CollectAndSend(); err != nil {
		log.Printf("Error in System Information:  %v\n", err)
	}

	for {
		// Pull latest configuration from the server
		updateRequested, err := config.PullFromServer()
		if err != nil {
			log.Printf("Warning: Could not pull latest configuration from server: %v. Using cached/local configuration.\n", err)
		}

		// Collect and send the current metrics snapshot
		if err := metrics.CollectAndSend(); err != nil {
			log.Printf("Error in System Metrics:  %v\n", err)
		}

		// Perform update immediately if requested by the control server
		if updateRequested {
			log.Println("Server requested agent update. Triggering manual update process...")
			updated, err := lib.Update(Version)
			if err != nil {
				log.Printf("Error: Manual update failed: %v\n", err)
			} else if updated {
				log.Println("Agent updated successfully. Exiting to apply update.")
				os.Exit(0)
			}
		}

		interval := 60
		if config.Content().ReportingIntervalSeconds > 0 {
			interval = config.Content().ReportingIntervalSeconds
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func printVersion() {
	fmt.Printf("Version: %s\n", Version)
}
