package commands

import (
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/base-go/mamba"
)

var devCmd = &mamba.Command{
	Use:   "dev",
	Short: "Start both backend and frontend development servers",
	Long:  `Start both backend (admin-api) and frontend (admin) development servers concurrently.`,
	Run:   runDev,
}

func init() {
	rootCmd.AddCommand(devCmd)
}

func runDev(cmd *mamba.Command, args []string) {
	// Check for backend and frontend directories
	// Support both standalone directories and monorepo structure
	backendDir := ""
	frontendDir := ""

	// Check for standalone structure (running from individual project directory)
	if fileExists("main.go") {
		backendDir = "."
	} else {
		// Look for directories ending with -api (new structure)
		backendDir = findDirWithSuffix("-api")
		// Fallback to old structure
		if backendDir == "" {
			if dirExists("admin-api-template") {
				backendDir = "admin-api-template"
			} else if dirExists("admin-api") {
				backendDir = "admin-api"
			}
		}
	}

	// Generate Swagger docs if backend is found
	if backendDir != "" {
		generateSwaggerDocs(cmd, backendDir)
	}

	if fileExists("nuxt.config.ts") {
		frontendDir = "."
	} else {
		// Look for directories ending with -app (new structure)
		frontendDir = findDirWithSuffix("-app")
		// Fallback to old structure
		if frontendDir == "" {
			if dirExists("admin-template") {
				frontendDir = "admin-template"
			} else if dirExists("admin") {
				frontendDir = "admin"
			}
		}
	}

	if backendDir == "" && frontendDir == "" {
		cmd.PrintError("Neither backend nor frontend directory found")
		cmd.PrintInfo("Run this command from your project root, backend, or frontend directory")
		os.Exit(1)
	}

	// Create channel to handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var processes []*exec.Cmd

	// Start backend
	if backendDir != "" {
		cmd.PrintInfo("Starting backend server...")
		backendCmd := exec.Command("go", "run", "main.go")
		if backendDir != "." {
			backendCmd.Dir = backendDir
		}
		// Suppress backend output during startup
		backendCmd.Stdout = nil
		backendCmd.Stderr = nil

		if err := backendCmd.Start(); err != nil {
			cmd.PrintError("Error starting backend: " + err.Error())
		} else {
			processes = append(processes, backendCmd)
			// Wait a bit for backend to initialize
			waitForBackend(cmd)
			cmd.PrintSuccess("Backend server ready (http://localhost:8000)")
		}
	}

	// Start frontend
	if frontendDir != "" {
		cmd.PrintInfo("Starting frontend server...")
		frontendCmd := exec.Command("bun", "dev")
		if frontendDir != "." {
			frontendCmd.Dir = frontendDir
		}
		// Suppress frontend output during startup
		frontendCmd.Stdout = nil
		frontendCmd.Stderr = nil

		if err := frontendCmd.Start(); err != nil {
			cmd.PrintError("Error starting frontend: " + err.Error())
		} else {
			processes = append(processes, frontendCmd)
			// Wait a bit for frontend to initialize
			waitForFrontend(cmd)
			cmd.PrintSuccess("Frontend server ready (http://localhost:3030)")
		}
	}

	if len(processes) == 0 {
		cmd.PrintError("No servers started")
		os.Exit(1)
	}

	cmd.PrintSuccess("All servers running. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	<-sigChan

	// Stop all processes
	cmd.PrintInfo("Stopping servers...")
	for _, p := range processes {
		if p.Process != nil {
			p.Process.Kill()
		}
	}

	cmd.PrintSuccess("All servers stopped")
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// findDirWithSuffix finds the first directory in current directory with the given suffix (case-sensitive)
func findDirWithSuffix(suffix string) string {
	entries, err := os.ReadDir(".")
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirName := entry.Name()
			if strings.HasSuffix(dirName, suffix) {
				return dirName
			}
		}
	}
	return ""
}

// waitForBackend waits for the backend server to be ready
func waitForBackend(cmd *mamba.Command) {
	client := &http.Client{Timeout: 1 * time.Second}
	for i := 0; i < 50; i++ {
		resp, err := client.Get("http://localhost:8000/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// waitForFrontend waits for the frontend server to be ready
func waitForFrontend(cmd *mamba.Command) {
	client := &http.Client{Timeout: 1 * time.Second}
	for i := 0; i < 50; i++ {
		resp, err := client.Get("http://localhost:3030")
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// generateSwaggerDocs generates Swagger documentation for the backend
func generateSwaggerDocs(cmd *mamba.Command, backendDir string) {
	// Find go executable
	goPath, err := exec.LookPath("go")
	if err != nil {
		return
	}

	// Ensure swag is installed
	if _, err := exec.LookPath("swag"); err != nil {
		installCmd := exec.Command(goPath, "install", "github.com/swaggo/swag/cmd/swag@latest")
		// Suppress output
		if err := installCmd.Run(); err != nil {
			return
		}
	}

	// Generate swagger docs (suppress output)
	swagCmd := exec.Command("swag", "init", "--dir", "./", "--output", "./swagger", "--parseDependency", "--parseInternal", "--parseVendor", "--parseDepth", "1", "--generatedTime", "false", "--quiet")
	swagCmd.Dir = backendDir
	// Don't pipe output to suppress all swagger logs

	if err := swagCmd.Run(); err != nil {
		// Silently fail - not critical for dev
		return
	}
}
