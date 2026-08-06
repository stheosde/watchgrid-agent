package lib

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLatestAgentVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/changelog/VERSION.txt" {
			t.Errorf("expected path '/changelog/VERSION.txt', got %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("1.5.2"))
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	ver := latestAgentVersion()
	if ver != "1.5.2" {
		t.Errorf("expected version '1.5.2', got %q", ver)
	}
}

func TestLatestAgentVersion_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	ver := latestAgentVersion()
	if ver != "unknown" {
		t.Errorf("expected version 'unknown', got %q", ver)
	}
}

func TestBinaryPath(t *testing.T) {
	oldBaseURL := baseURL
	baseURL = "https://example.com"
	defer func() { baseURL = oldBaseURL }()

	path := binaryPath()
	if path == "" {
		t.Error("expected binaryPath to not be empty")
	}
}

func TestCreateLink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "link-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	target := filepath.Join(tmpDir, "target-file")
	linkName := filepath.Join(tmpDir, "symlink-file")

	// Create target file
	if err := os.WriteFile(target, []byte("target content"), 0644); err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}

	// Create link
	err = createLink(target, linkName)
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	// Verify link points to target
	resolved, err := os.Readlink(linkName)
	if err != nil {
		t.Fatalf("failed to read link: %v", err)
	}
	if resolved != target {
		t.Errorf("expected link to point to %q, got %q", target, resolved)
	}

	// Recreate link (should remove existing link name and create new one)
	newTarget := filepath.Join(tmpDir, "new-target-file")
	if err := os.WriteFile(newTarget, []byte("new target content"), 0644); err != nil {
		t.Fatalf("failed to create new target file: %v", err)
	}

	err = createLink(newTarget, linkName)
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	resolved, err = os.Readlink(linkName)
	if err != nil {
		t.Fatalf("failed to read link: %v", err)
	}
	if resolved != newTarget {
		t.Errorf("expected link to point to %q, got %q", newTarget, resolved)
	}
}

func TestUpdate(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/changelog/VERSION.txt" {
			w.Write([]byte("2.0.0"))
			return
		}
		if strings.HasSuffix(r.URL.Path, ".sha256") {
			w.Write([]byte("1ad13fab15efda46d06bdec941b56a105638c3a05bed580f95b92eb46b90fb5b  watchgrid"))
			return
		}
		// Any other path is binary download
		w.Write([]byte("mock-binary-data"))
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "update-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldBaseURL, oldInstallPath := baseURL, InstallPath
	baseURL = server.URL
	InstallPath = tmpDir + string(filepath.Separator)
	defer func() {
		baseURL = oldBaseURL
		InstallPath = oldInstallPath
	}()

	updated, err := Update("1.0.0")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !updated {
		t.Error("expected Update to return true (updated)")
	}

	// Verify file was downloaded
	downloadedFile := filepath.Join(tmpDir, "watchgrid-2.0.0")
	if _, err := os.Stat(downloadedFile); os.IsNotExist(err) {
		t.Error("expected downloaded binary file to exist")
	}

	// Verify symlink was created
	symlink := filepath.Join(tmpDir, "watchgrid")
	resolved, err := os.Readlink(symlink)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if resolved != downloadedFile {
		t.Errorf("expected symlink to point to %q, got %q", downloadedFile, resolved)
	}
}

func TestVersionParsingAndHelpers(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		newer   bool
		compat  bool
	}{
		{"1.0.0", "1.0.0", false, false},
		{"1.0.0", "1.1.0", true, true},
		{"1.1.0", "1.1.1", true, true},
		{"1.1.0", "1.0.9", false, false},
		{"1.1.0", "2.0.0", true, false}, // newer, but not compatible (major change)
		{"unknown", "1.1.0", false, false},
		{"1.1.0", "unknown", false, false},
		{"v1.1.0", "v1.2.0", true, true},
		{"1.2.0-beta", "1.2.0", true, true},
	}

	for _, tt := range tests {
		newer := isNewerVersion(tt.current, tt.latest)
		compat := isCompatibleUpdate(tt.current, tt.latest)
		if newer != tt.newer {
			t.Errorf("isNewerVersion(%q, %q) = %v; want %v", tt.current, tt.latest, newer, tt.newer)
		}
		if compat != tt.compat {
			t.Errorf("isCompatibleUpdate(%q, %q) = %v; want %v", tt.current, tt.latest, compat, tt.compat)
		}
	}
}


func TestUpdate_ChecksumMismatch(t *testing.T) {
	// Create mock server that serves an invalid checksum (mismatching)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/changelog/VERSION.txt" {
			w.Write([]byte("2.0.0"))
			return
		}
		if strings.HasSuffix(r.URL.Path, ".sha256") {
			// Serve an incorrect checksum
			w.Write([]byte("incorrecthash60b4de09be52467d0cf08cc681a7edfb0ff4c5770d10de8356a4216892  watchgrid"))
			return
		}
		w.Write([]byte("mock-binary-data"))
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "update-fail-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldBaseURL, oldInstallPath := baseURL, InstallPath
	baseURL = server.URL
	InstallPath = tmpDir + string(filepath.Separator)
	defer func() {
		baseURL = oldBaseURL
		InstallPath = oldInstallPath
	}()

	updated, err := Update("1.0.0")
	if err == nil {
		t.Error("expected Update to return an error due to checksum mismatch")
	}
	if updated {
		t.Error("expected updated to be false")
	}

	// Verify the invalid downloaded binary file was deleted
	downloadedFile := filepath.Join(tmpDir, "watchgrid-2.0.0")
	if _, err := os.Stat(downloadedFile); !os.IsNotExist(err) {
		t.Error("expected downloaded binary file to be deleted on checksum failure")
	}
}
