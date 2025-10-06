package commands

import (
	"os"
	"os/exec"

	"github.com/base-go/mamba"
	"github.com/base-go/mamba/pkg/spinner"
)

var buildCmd = &mamba.Command{
	Use:   "build [backend|frontend]",
	Short: "Build backend, frontend, or both",
	Long: `Build the project for production.

Examples:
  bui build              # Build both backend and frontend
  bui build backend      # Build backend only
  bui build frontend     # Build frontend only`,
	Run: buildBoth,
}

var buildBackendCmd = &mamba.Command{
	Use:   "backend",
	Short: "Build backend only",
	Run:   buildBackend,
}

var buildFrontendCmd = &mamba.Command{
	Use:   "frontend",
	Short: "Build frontend only",
	Run:   buildFrontend,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.AddCommand(buildBackendCmd)
	buildCmd.AddCommand(buildFrontendCmd)
}

func buildBoth(cmd *mamba.Command, args []string) {
	cmd.PrintInfo("Building project...")

	// Build backend
	if dirExists("admin-api") {
		cmd.PrintInfo("Building backend...")
		buildBackend(cmd, args)
	}

	// Build frontend
	if dirExists("admin") {
		cmd.PrintInfo("Building frontend...")
		buildFrontend(cmd, args)
	}

	cmd.PrintSuccess("Build complete")
}

func buildBackend(cmd *mamba.Command, args []string) {
	backendDir := "admin-api"

	if !dirExists(backendDir) {
		cmd.PrintError("admin-api directory not found")
		os.Exit(1)
	}

	// Build Go binary with spinner
	err := spinner.WithSpinner("Building backend...", func() error {
		buildCmd := exec.Command("go", "build", "-o", "bin/server", "cmd/server/main.go")
		buildCmd.Dir = backendDir
		return buildCmd.Run()
	})

	if err != nil {
		cmd.PrintError("Error building backend: " + err.Error())
		os.Exit(1)
	}

	cmd.PrintSuccess("Backend built: admin-api/bin/server")
}

func buildFrontend(cmd *mamba.Command, args []string) {
	frontendDir := "admin"

	if !dirExists(frontendDir) {
		cmd.PrintError("admin directory not found")
		os.Exit(1)
	}

	// Build Nuxt app with spinner
	err := spinner.WithSpinner("Building frontend...", func() error {
		buildCmd := exec.Command("bun", "run", "build")
		buildCmd.Dir = frontendDir
		return buildCmd.Run()
	})

	if err != nil {
		cmd.PrintError("Error building frontend: " + err.Error())
		os.Exit(1)
	}

	cmd.PrintSuccess("Frontend built: admin/.output")
}
