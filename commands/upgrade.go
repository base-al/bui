package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/base-al/bui/version"
	"github.com/base-go/mamba"
	"github.com/base-go/mamba/pkg/spinner"
)

var upgradeCmd = &mamba.Command{
	Use:   "upgrade",
	Short: "Upgrade Bui CLI to the latest version",
	Long:  `Download and install the latest version of Bui CLI.`,
	Run:   runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *mamba.Command, args []string) {
	// Check current and latest versions
	currentVersion := version.Version
	if currentVersion == "" {
		currentVersion = "unknown"
	}

	var latestVersion string
	err := spinner.WithSpinner("Checking for updates...", func() error {
		version, err := getLatestVersion()
		if err != nil {
			return err
		}
		latestVersion = version
		return nil
	})

	if err != nil {
		cmd.PrintWarning(fmt.Sprintf("Failed to check latest version: %v", err))
		latestVersion = "unknown"
	}

	cmd.PrintInfo("")
	cmd.PrintInfo(fmt.Sprintf("Current version: %s", currentVersion))
	cmd.PrintInfo(fmt.Sprintf("Latest version:  %s", latestVersion))
	cmd.PrintInfo("")

	// Check if already up to date
	if currentVersion != "unknown" && latestVersion != "unknown" {
		if strings.TrimPrefix(currentVersion, "v") == strings.TrimPrefix(latestVersion, "v") {
			cmd.PrintSuccess("You are already running the latest version!")
			return
		}
	}

	// Detect if installed via go install or install script
	exePath, err := os.Executable()
	if err != nil {
		cmd.PrintError("Failed to detect installation path")
		os.Exit(1)
	}

	cmd.PrintHeader("Upgrading Bui CLI")
	cmd.PrintInfo(fmt.Sprintf("Installation path: %s", exePath))
	cmd.PrintInfo("")

	// Determine the install script command based on OS
	var installCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: download and run with PowerShell
		installCmd = exec.Command("powershell", "-Command",
			"Invoke-WebRequest -Uri https://raw.githubusercontent.com/base-al/bui/main/install.sh -OutFile install.sh; bash install.sh; Remove-Item install.sh")
	} else {
		// Unix: use curl and bash
		installCmd = exec.Command("bash", "-c",
			"curl -sSL https://raw.githubusercontent.com/base-al/bui/main/install.sh | bash")
	}

	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Stdin = os.Stdin

	cmd.PrintInfo("Running installation script...")
	if err := installCmd.Run(); err != nil {
		cmd.PrintError("Failed to run install script")
		cmd.PrintInfo("")
		cmd.PrintHeader("Manual Installation")
		cmd.PrintBullet("curl -sSL https://raw.githubusercontent.com/base-al/bui/main/install.sh | bash")
		os.Exit(1)
	}

	cmd.PrintInfo("")
	cmd.PrintSuccess("Successfully upgraded Bui CLI!")
	cmd.PrintInfo("Run 'bui version' to check the new version")
}

// getLatestVersion fetches the latest release version from GitHub
func getLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/base-al/bui/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}

	return release.TagName, nil
}
