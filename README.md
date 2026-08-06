# WatchGrid Agent

[![Build Status](https://github.com/stheosde/watchgrid-agent/actions/workflows/release.yml/badge.svg)](https://github.com/stheosde/watchgrid-agent/actions)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25.0-blue.svg)](https://go.dev/)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

A lightweight background daemon written in Go that monitors system health and reports back to [WatchGrid](https://watchgrid.de), a centralized server monitoring platform.

---

## Features

- **System Metrics**: Collects CPU, memory, uptime, and disk usage natively (reading procfs and kernel statistics directly instead of spawning external shell tools).
- **Port Checking**: Inspects active network sockets to detect if non-whitelisted ports are open.
- **Service Monitoring**: Verifies the status of configured system services using systemd.
- **Auto-Updates**: Safely downloads newer versions of the agent, verifies the download checksum, and updates the executable.
- **Single Instance Lock**: Ensures only one instance of the agent daemon runs on a host.

---

## WatchGrid Monitoring Service

This agent collects metrics to be visualized on the **[WatchGrid](https://watchgrid.de)** monitoring platform. 

To use this agent:
1. Register an account on **[WatchGrid](https://watchgrid.de)**.
2. Add a new server in your dashboard to obtain an `Agent ID` and `API Token`.
3. Use those credentials when running the installer script below.

---

## How to Install

The easiest way to install and configure the WatchGrid Agent on a remote machine is using the official installer script:

```bash
curl -sfL https://wgri.de/install | WG_AGENT="<agent_id>" WG_TOKEN="<token_id>" sh -
```

This registers the agent as a system service, creates the configuration directory at `/etc/watchgrid`, and starts the background daemon.

---

## Configuration

The agent uses two configuration files, typically stored in `/etc/watchgrid/` (or `./config/` during local debugging):

### 1. `agent.json` (Static Identity)
Defines the agent's identity, token credentials, and communication endpoint:

```json
{
  "agent_id": "your-agent-uuid",
  "token": "your-secure-api-token",
  "base_url": "https://wgri.de"
}
```

### 2. `config.json` (Dynamic Monitor Settings)
Automatically pulled and synchronized from the WatchGrid server. It controls what the agent monitors and how often:

```json
{
  "ports": [22, 80, 443, 9100],
  "services": ["nginx", "docker"],
  "allowed_uptime_days": 120,
  "reporting_interval_seconds": 60
}
```

---

## Command Line Flags

You can run the agent manually to verify system stats or perform diagnostics:

```text
Usage of watchgrid:
  -check
    	Check system status without transmitting data
  -help
    	Show this help message
  -info
    	Print system information locally and exit
  -isend
    	Print system information and transmit to server immediately
  -metrics
    	Print system metrics locally and exit
  -msend
    	Print system metrics and transmit to server immediately
  -update
    	Perform an update of the agent if available
  -version
    	Print version information and exit
```

---

## Development & Building

### Prerequisites
- Go 1.25.0 or newer installed locally.

### Building from Source
To build the binary for your local platform:
```bash
go build -o watchgrid .
```

To cross-compile for a specific platform (e.g., Linux ARM64):
```bash
env GOOS=linux GOARCH=arm64 go build -ldflags="-X main.Version=1.0.0 -s -w" -o watchgrid .
```

### Running Tests
To run unit tests (requires bypass sandbox / local bind permission for httptest network binding):
```bash
go test ./...
```

---

## CI/CD Release Pipeline
The project uses GitHub Actions for releases. When a new release tag is pushed to the repository:
1. The compiler builds binaries for target architectures (`amd64`, `arm64`).
2. Binaries and SHA-256 checksums are uploaded to the S3 packages bucket.
3. A new **GitHub Release** is drafted with the compiled binaries attached as release assets.

---

## License

This project is licensed under the GPL-3.0 License - see the [LICENSE](LICENSE) file for details.
