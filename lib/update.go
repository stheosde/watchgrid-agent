package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"watchgrid.de/agent/lib/sysinfo"
)

var (
	baseURL     = "https://wgri.de/packages"
	InstallPath = "/usr/local/bin/"
	httpClient  = &http.Client{Timeout: 15 * time.Second}
)

var (
	LatestVersion string
)

// Update checks for updates and performs the update if the latest version is different from the current version.
// It returns true if an update was successfully performed.
func Update(version string) (bool, error) {
	LatestVersion = latestAgentVersion()
	if LatestVersion == "unknown" {
		return false, fmt.Errorf("could not determine latest version from server")
	}

	if LatestVersion == version {
		println("The latest version is already installed:", version)
		return false, nil
	}

	println("Updating agent from version", version, "to", LatestVersion)
	if err := downloadBinary(binaryPath()); err != nil {
		return false, fmt.Errorf("failed to download binary: %w", err)
	}

	if err := createLink(InstallPath+"watchgrid-"+LatestVersion, InstallPath+"watchgrid"); err != nil {
		return false, fmt.Errorf("failed to create symlink: %w", err)
	}

	println("Updated successfully to version", LatestVersion)
	return true, nil
}



// parseVersion parses a semver string like "1.2.3" or "v1.2.3" into major, minor, patch ints.
// It also returns if the version is a pre-release (contains '-') and true on success.
func parseVersion(v string) (int, int, int, bool, bool) {
	v = strings.TrimPrefix(v, "v")
	isPre := strings.Contains(v, "-")
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		return 0, 0, 0, false, false
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	patch, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false, false
	}
	return major, minor, patch, isPre, true
}

// isNewerVersion returns true if latest is strictly newer than current.
func isNewerVersion(current, latest string) bool {
	currMajor, currMinor, currPatch, currPre, ok1 := parseVersion(current)
	latMajor, latMinor, latPatch, latPre, ok2 := parseVersion(latest)
	if !ok1 || !ok2 {
		return false
	}
	if latMajor > currMajor {
		return true
	}
	if latMajor < currMajor {
		return false
	}
	if latMinor > currMinor {
		return true
	}
	if latMinor < currMinor {
		return false
	}
	if latPatch > currPatch {
		return true
	}
	if latPatch < currPatch {
		return false
	}
	// If base versions are equal, a release is newer than a pre-release.
	if currPre && !latPre {
		return true
	}
	return false
}

// isCompatibleUpdate returns true if latest is newer than current and has the same major version.
func isCompatibleUpdate(current, latest string) bool {
	currMajor, _, _, _, ok1 := parseVersion(current)
	latMajor, _, _, _, ok2 := parseVersion(latest)
	if !ok1 || !ok2 {
		return false
	}
	if currMajor != latMajor {
		return false
	}
	return isNewerVersion(current, latest)
}

func latestAgentVersion() string {
	resp, err := httpClient.Get(baseURL + "/changelog/VERSION.txt")
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "unknown"
	}

	buf := make([]byte, 64)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		return "unknown"
	}

	return strings.Trim(strings.TrimSpace(string(buf[:n])), "'$")
}

func binaryPath() string {
	architecture := sysinfo.Architecture()
	platform := strings.ToLower(sysinfo.OperatingSystem())
	return baseURL + "/" + platform + "/" + architecture + "/watchgrid"
}

func downloadBinary(url string) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status code %d", resp.StatusCode)
	}

	targetFile := InstallPath + "watchgrid-" + LatestVersion
	// Remove the target file first if it exists to avoid "text file busy" (ETXTBSY)
	// when updating a binary that is currently running.
	_ = os.Remove(targetFile)

	outFile, err := os.Create(targetFile)
	if err != nil {
		return err
	}

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		outFile.Close()
		_ = os.Remove(targetFile)
		return err
	}

	if err := outFile.Close(); err != nil {
		_ = os.Remove(targetFile)
		return err
	}

	checksumURL := url + ".sha256"
	if err := verifyChecksum(targetFile, checksumURL); err != nil {
		_ = os.Remove(targetFile)
		return fmt.Errorf("checksum validation failed: %w", err)
	}

	err = os.Chmod(targetFile, 0755)
	if err != nil {
		_ = os.Remove(targetFile)
		return err
	}

	return nil
}

func verifyChecksum(filePath string, sha256URL string) error {
	resp, err := httpClient.Get(sha256URL)
	if err != nil {
		return fmt.Errorf("failed to fetch checksum: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status code %d when fetching checksum", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read checksum response body: %w", err)
	}

	fields := strings.Fields(string(bodyBytes))
	if len(fields) < 1 {
		return fmt.Errorf("invalid checksum file format")
	}
	expectedHash := strings.ToLower(fields[0])

	if len(expectedHash) != 64 {
		return fmt.Errorf("invalid checksum length: expected 64 hex characters, got %d", len(expectedHash))
	}
	if _, err := hex.DecodeString(expectedHash); err != nil {
		return fmt.Errorf("invalid checksum format: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum verification: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))

	if expectedHash != actualHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}


func createLink(target string, linkName string) error {
	if _, err := os.Lstat(linkName); err == nil {
		if err := os.Remove(linkName); err != nil {
			return err
		}
	}
	return os.Symlink(target, linkName)
}
