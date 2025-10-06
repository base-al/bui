package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/base-go/mamba"
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
	cmd.PrintInfo("Upgrading Bui CLI...")
	cmd.PrintInfo("")

	// Detect if installed via go install or install script
	exePath, err := os.Executable()
	if err != nil {
		cmd.PrintError("Failed to detect installation path")
		os.Exit(1)
	}

	cmd.PrintInfo(fmt.Sprintf("Current installation: %s", exePath))
	cmd.PrintInfo("")

	// Check if Go is available (for go install method)
	if _, err := exec.LookPath("go"); err == nil {
		cmd.PrintInfo("Upgrading via go install...")
		upgradeCmd := exec.Command("go", "install", "github.com/base-al/bui@latest")
		upgradeCmd.Stdout = os.Stdout
		upgradeCmd.Stderr = os.Stderr
		if err := upgradeCmd.Run(); err != nil {
			cmd.PrintError("Failed to upgrade via go install")
			cmd.PrintInfo("")
			cmd.PrintInfo("Try running the install script instead:")
			cmd.PrintInfo("  curl -sSL https://raw.githubusercontent.com/base-al/bui/main/install.sh | bash")
			os.Exit(1)
		}
		cmd.PrintInfo("")
		cmd.PrintSuccess("Successfully upgraded Bui CLI!")
		cmd.PrintInfo("Run 'bui version' to check the new version")
	} else {
		// Use install script
		cmd.PrintInfo("Downloading and running install script...")
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

		if err := installCmd.Run(); err != nil {
			cmd.PrintError("Failed to run install script")
			cmd.PrintInfo("")
			cmd.PrintInfo("Try manually running:")
			cmd.PrintInfo("  curl -sSL https://raw.githubusercontent.com/base-al/bui/main/install.sh | bash")
			os.Exit(1)
		}

		cmd.PrintInfo("")
		cmd.PrintSuccess("Successfully upgraded Bui CLI!")
		cmd.PrintInfo("Run 'bui version' to check the new version")
	}
}
