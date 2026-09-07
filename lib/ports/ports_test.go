package ports

import (
	"os"
	"slices"
	"testing"
)

func TestOpenPorts(t *testing.T) {
	tmpTCP, err := os.CreateTemp("", "mock-tcp")
	if err != nil {
		t.Fatalf("failed to create temp tcp file: %v", err)
	}
	defer os.Remove(tmpTCP.Name())

	tmpTCP6, err := os.CreateTemp("", "mock-tcp6")
	if err != nil {
		t.Fatalf("failed to create temp tcp6 file: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(tmpTCP6.Name())
	// 0016 hex = 22 dec, 0050 hex = 80 dec, 01BB hex = 443 dec, 238c hex = 9100 dec
	tcpContent := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 00000000
   1: 00000000:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 00000000
   2: 00000000:01BB 00000000:0000 03 00000000:00000000 00:00000000 00000000  1000        0 12345 1 00000000
`
	tcp6Content := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000000000000000000001000000:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 00000000
   1: 00000000000000000000000000000000:238C 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 00000000
`
	if _, err := tmpTCP.WriteString(tcpContent); err != nil {
		t.Fatalf("failed to write mock tcp: %v", err)
	}
	if _, err := tmpTCP6.WriteString(tcp6Content); err != nil {
		t.Fatalf("failed to write mock tcp6: %v", err)
	}
	err = tmpTCP.Close()
	if err != nil {
		return
	}
	err = tmpTCP6.Close()
	if err != nil {
		return
	}

	oldTCPPath, oldTCP6Path := tcpPath, tcp6Path
	tcpPath = tmpTCP.Name()
	tcp6Path = tmpTCP6.Name()
	defer func() {
		tcpPath = oldTCPPath
		tcp6Path = oldTCP6Path
	}()

	portsList, err := OpenPorts()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(portsList) != 2 {
		t.Errorf("expected 2 open ports, got %d: %v", len(portsList), portsList)
	}
	if !slices.Contains(portsList, 22) {
		t.Errorf("expected ports to contain 22, got %v", portsList)
	}
	if !slices.Contains(portsList, 9100) {
		t.Errorf("expected ports to contain 9100, got %v", portsList)
	}
}


func TestCheckNonWhitelistedPorts(t *testing.T) {
	tmpTCP, err := os.CreateTemp("", "mock-tcp-nonwhite")
	if err != nil {
		t.Fatalf("failed to create temp tcp file: %v", err)
	}
	defer os.Remove(tmpTCP.Name())

	tcpContent := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 00000000
   1: 00000000:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 00000000
   2: 00000000:238C 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 00000000
`
	if _, err := tmpTCP.WriteString(tcpContent); err != nil {
		t.Fatalf("failed to write mock tcp: %v", err)
	}
	tmpTCP.Close()

	oldTCPPath, oldTCP6Path := tcpPath, tcp6Path
	tcpPath = tmpTCP.Name()
	tcp6Path = ""
	defer func() {
		tcpPath = oldTCPPath
		tcp6Path = oldTCP6Path
	}()

	// 22 and 9100 are open, if allowed is only [22], then 9100 should be non-whitelisted
	nonWhite, err := CheckNonWhitelistedPorts([]int{22})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(nonWhite) != 1 || nonWhite[0] != 9100 {
		t.Errorf("expected [9100], got %v", nonWhite)
	}

	nonWhite, err = CheckNonWhitelistedPorts([]int{22, 9100})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(nonWhite) != 0 {
		t.Errorf("expected empty, got %v", nonWhite)
	}
}

func TestOpenPorts_ReadError(t *testing.T) {
	oldTCPPath := tcpPath
	tcpPath = "non-existent-tcp-file.json"
	defer func() {
		t.Logf("restoring tcpPath to %s", oldTCPPath)
		tcpPath = oldTCPPath
	}()

	portsList, err := OpenPorts()
	if err == nil {
		t.Error("expected error when reading non-existent TCP file, got nil")
	}
	if len(portsList) != 0 {
		t.Errorf("expected 0 open ports on error, got %v", portsList)
	}

	nonWhite, err := CheckNonWhitelistedPorts([]int{22})
	if err == nil {
		t.Error("expected error from CheckNonWhitelistedPorts, got nil")
	}
	if len(nonWhite) != 0 {
		t.Errorf("expected 0 non-whitelisted ports on error, got %v", nonWhite)
	}
}
