package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/base-al/bui/version"
	"github.com/base-go/mamba"
)

// CheckForUpdate checks if a new version is available and prints a message
func CheckForUpdate(cmd *mamba.Command) {
	currentVersion := version.Version
	if currentVersion == "" || currentVersion == "unknown" {
		return
	}

	// Check for updates in background (with timeout)
	done := make(chan string, 1)
	go func() {
		latestVersion, err := getLatestVersion()
		if err == nil {
			done <- latestVersion
		} else {
			done <- ""
		}
	}()

	// Wait for result with 2 second timeout
	var latestVersion string
	select {
	case latestVersion = <-done:
	case <-time.After(2 * time.Second):
		return // Timeout, don't show update message
	}

	if latestVersion == "" {
		return
	}

	// Compare versions (strip 'v' prefix if present)
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestVersion, "v")

	if current != latest {
		cmd.PrintInfo("")
		cmd.PrintWarning(fmt.Sprintf("New version available: %s → %s", currentVersion, latestVersion))
		cmd.PrintInfo("Run 'bui upgrade' to update")
		cmd.PrintInfo("")
	}
}
