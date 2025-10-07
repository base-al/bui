package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/base-go/mamba"
)

var previewCmd = &mamba.Command{
	Use:   "preview",
	Short: "Preview the production build",
	Long:  `Preview the production build by running the server binary from the dist directory.`,
	Run:   runPreview,
}

func init() {
	rootCmd.AddCommand(previewCmd)
}

func runPreview(cmd *mamba.Command, args []string) {
	// Find dist directory
	distDir := findDistDir()
	if distDir == "" {
		cmd.PrintError("No dist directory found. Run 'bui build' first.")
		os.Exit(1)
	}

	// Check if server binary exists
	serverPath := filepath.Join(distDir, "server")
	if !fileExistsPreview(serverPath) {
		cmd.PrintError(fmt.Sprintf("Server binary not found at %s. Run 'bui build' first.", serverPath))
		os.Exit(1)
	}

	// Check if .env exists
	envPath := filepath.Join(distDir, ".env")
	if !fileExistsPreview(envPath) {
		cmd.PrintWarning("No .env file found in " + distDir)
		cmd.PrintInfo("Copy .env.example to .env and configure it for preview")
		os.Exit(1)
	}

	cmd.PrintSuccess("Starting production preview server...")
	cmd.PrintInfo(fmt.Sprintf("Running from: %s", distDir))
	cmd.PrintInfo("Press Ctrl+C to stop\n")

	// Run the server
	serverCmd := exec.Command("./server")
	serverCmd.Dir = distDir
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	serverCmd.Env = os.Environ()

	if err := serverCmd.Run(); err != nil {
		cmd.PrintError("Failed to run server: " + err.Error())
		os.Exit(1)
	}
}

// findDistDir finds the dist directory (dist/ or *-dist/)
func findDistDir() string {
	// Check for "dist" first
	if dirExistsPreview("dist") {
		return "dist"
	}

	// Check for *-dist directories
	entries, err := os.ReadDir(".")
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), "-dist") {
			return entry.Name()
		}
	}

	return ""
}

func dirExistsPreview(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func fileExistsPreview(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
