package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	cmd.PrintInfo("Building production deployment...")

	// Detect project structure
	backendDir := detectBackendDir()
	frontendDir := detectFrontendDir()

	if backendDir == "" && frontendDir == "" {
		cmd.PrintError("No backend or frontend directories found")
		os.Exit(1)
	}

	// Determine dist directory name based on project structure
	distDir := determineDistDir(backendDir, frontendDir)

	// Clean and create dist directory
	cmd.PrintInfo("Preparing " + distDir + " directory...")
	os.RemoveAll(distDir)
	os.MkdirAll(distDir, 0755)

	// Build backend
	if backendDir != "" {
		buildBackendToDist(cmd, backendDir, distDir)
	}

	// Build frontend
	if frontendDir != "" {
		buildFrontendToDist(cmd, frontendDir, distDir)
	}

	// Create deployment files
	if backendDir != "" && frontendDir != "" {
		createDeploymentFiles(cmd, backendDir, distDir)
		cmd.PrintSuccess("Production build complete!")
		cmd.PrintInfo("Deployment files created in ./" + distDir + "/")
		cmd.PrintInfo("  • Backend binary: " + distDir + "/server")
		cmd.PrintInfo("  • Frontend files: " + distDir + "/public/")
		cmd.PrintInfo("  • Dockerfile: " + distDir + "/Dockerfile")
		cmd.PrintInfo("  • CapRover: " + distDir + "/captain-definition.json")
	} else {
		cmd.PrintSuccess("Build complete in " + distDir + "/")
	}
}

func buildBackend(cmd *mamba.Command, args []string) {
	backendDir := "admin-api"

	if !dirExists(backendDir) {
		cmd.PrintError("admin-api directory not found")
		os.Exit(1)
	}

	// Generate Swagger docs before building
	generateSwaggerDocsForBuild(cmd, backendDir)

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

// generateSwaggerDocsForBuild generates Swagger documentation for the backend during build
func generateSwaggerDocsForBuild(cmd *mamba.Command, backendDir string) {
	cmd.PrintInfo("Generating Swagger documentation...")

	// Find go executable
	goPath, err := exec.LookPath("go")
	if err != nil {
		cmd.PrintWarning("Go executable not found, skipping swagger generation")
		return
	}

	// Ensure swag is installed
	if _, err := exec.LookPath("swag"); err != nil {
		cmd.PrintInfo("Installing swag...")
		installCmd := exec.Command(goPath, "install", "github.com/swaggo/swag/cmd/swag@latest")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			cmd.PrintWarning(fmt.Sprintf("Failed to install swag: %v", err))
			return
		}
	}

	// Generate swagger docs
	swagCmd := exec.Command("swag", "init", "--dir", "./", "--output", "./docs", "--parseDependency", "--parseInternal", "--parseVendor", "--parseDepth", "1", "--generatedTime", "false")
	swagCmd.Dir = backendDir
	swagCmd.Stdout = os.Stdout
	swagCmd.Stderr = os.Stderr

	if err := swagCmd.Run(); err != nil {
		cmd.PrintWarning(fmt.Sprintf("Failed to generate docs: %v", err))
	} else {
		cmd.PrintSuccess("Swagger documentation generated")
	}
}

// determineDistDir determines the dist directory name based on project structure
func determineDistDir(backendDir, frontendDir string) string {
	// If backend or frontend has -api or -app suffix, extract project name
	if backendDir != "" && backendDir != "." && strings.HasSuffix(backendDir, "-api") {
		projectName := strings.TrimSuffix(backendDir, "-api")
		return projectName + "-dist"
	}

	if frontendDir != "" && frontendDir != "." && strings.HasSuffix(frontendDir, "-app") {
		projectName := strings.TrimSuffix(frontendDir, "-app")
		return projectName + "-dist"
	}

	// Default to "dist"
	return "dist"
}

// detectBackendDir finds the backend directory
func detectBackendDir() string {
	candidates := []string{
		"admin-api-template",
		"admin-api",
	}

	// Check for -api suffix directories
	if dir := findDirWithSuffixBuild("-api"); dir != "" {
		return dir
	}

	for _, dir := range candidates {
		if dirExists(dir) {
			return dir
		}
	}

	// Check if current directory is backend
	if fileExistsBuild("main.go") {
		return "."
	}

	return ""
}

// detectFrontendDir finds the frontend directory
func detectFrontendDir() string {
	candidates := []string{
		"admin-template",
		"admin",
	}

	// Check for -app suffix directories
	if dir := findDirWithSuffixBuild("-app"); dir != "" {
		return dir
	}

	for _, dir := range candidates {
		if dirExists(dir) {
			return dir
		}
	}

	// Check if current directory is frontend
	if fileExistsBuild("nuxt.config.ts") {
		return "."
	}

	return ""
}

// buildBackendToDist builds the backend to distDir/server
func buildBackendToDist(cmd *mamba.Command, backendDir, distDir string) {
	cmd.PrintInfo("Building backend...")

	// Generate Swagger docs
	generateSwaggerDocsForBuild(cmd, backendDir)

	// Build binary
	err := spinner.WithSpinner("Compiling backend binary...", func() error {
		outputPath := filepath.Join("..", distDir, "server")
		buildCmd := exec.Command("go", "build", "-o", outputPath, "main.go")
		buildCmd.Dir = backendDir
		return buildCmd.Run()
	})

	if err != nil {
		cmd.PrintError("Failed to build backend: " + err.Error())
		os.Exit(1)
	}

	// Copy necessary directories
	cmd.PrintInfo("Copying backend assets...")
	copyDir(filepath.Join(backendDir, "docs"), filepath.Join(distDir, "docs"))
	copyDir(filepath.Join(backendDir, "templates"), filepath.Join(distDir, "templates"))
	copyDir(filepath.Join(backendDir, "static"), filepath.Join(distDir, "static"))

	// Create storage directory
	os.MkdirAll(filepath.Join(distDir, "storage"), 0755)

	// Copy .env.example if exists
	if fileExistsBuild(filepath.Join(backendDir, ".env.example")) {
		copyFile(filepath.Join(backendDir, ".env.example"), filepath.Join(distDir, ".env.example"))
	}

	// Copy .env if exists (for local testing)
	if fileExistsBuild(filepath.Join(backendDir, ".env")) {
		copyFile(filepath.Join(backendDir, ".env"), filepath.Join(distDir, ".env"))
		cmd.PrintInfo("Copied .env for local preview")
	}

	cmd.PrintSuccess("Backend built successfully")
}

// buildFrontendToDist builds the frontend to distDir/public
func buildFrontendToDist(cmd *mamba.Command, frontendDir, distDir string) {
	cmd.PrintInfo("Building frontend...")

	// Run nuxt generate
	err := spinner.WithSpinner("Generating static frontend...", func() error {
		generateCmd := exec.Command("bun", "run", "generate")
		generateCmd.Dir = frontendDir
		generateCmd.Stdout = os.Stdout
		generateCmd.Stderr = os.Stderr
		return generateCmd.Run()
	})

	if err != nil {
		cmd.PrintError("Failed to build frontend: " + err.Error())
		os.Exit(1)
	}

	// Copy .output/public to distDir/public
	cmd.PrintInfo("Copying frontend files...")
	outputDir := filepath.Join(frontendDir, ".output", "public")
	if dirExists(outputDir) {
		copyDir(outputDir, filepath.Join(distDir, "public"))
		cmd.PrintSuccess("Frontend built successfully")
	} else {
		cmd.PrintWarning("Frontend output not found at " + outputDir)
	}
}

// createDeploymentFiles creates Dockerfile and captain-definition.json
func createDeploymentFiles(cmd *mamba.Command, _ string, distDir string) {
	cmd.PrintInfo("Creating deployment files...")

	// Create Dockerfile
	dockerfile := `FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy everything from dist
COPY . .

# Make binary executable
RUN chmod +x ./server

# Expose port
EXPOSE 8000

# Run the binary
CMD ["./server"]
`
	os.WriteFile(filepath.Join(distDir, "Dockerfile"), []byte(dockerfile), 0644)

	// Create captain-definition.json
	captainDef := `{
  "schemaVersion": 2,
  "dockerfilePath": "./Dockerfile"
}
`
	os.WriteFile(filepath.Join(distDir, "captain-definition.json"), []byte(captainDef), 0644)

	// Create .dockerignore
	dockerignore := `*.db
*.log
.env
storage/upload/*
!storage/upload/.gitkeep
`
	os.WriteFile(filepath.Join(distDir, ".dockerignore"), []byte(dockerignore), 0644)

	// Create README for deployment
	readme := `# Production Deployment

This directory contains a complete production build ready for deployment.

## Structure
- server - Backend binary
- public/ - Frontend static files
- docs/ - Swagger documentation
- templates/ - Email templates
- storage/ - File storage directory
- Dockerfile - Docker image definition
- captain-definition.json - CapRover deployment config

## Deployment Options

### CapRover
1. Create a new app in CapRover
2. Deploy this directory as a tarball:
   tar -czf deploy.tar.gz -C ` + distDir + ` .
3. Upload deploy.tar.gz to CapRover
4. Set environment variables in CapRover dashboard

### Docker
cd ` + distDir + `
docker build -t myapp .
docker run -p 8000:8000 --env-file .env myapp

### Direct Deployment
1. Copy this directory to your server
2. Create .env file with production settings
3. Run: ./server

## Environment Variables
Copy .env.example to .env and configure:
- Database settings
- JWT secret
- Storage settings
- Email configuration
`

	os.WriteFile(filepath.Join(distDir, "README.md"), []byte(readme), 0644)

	cmd.PrintSuccess("Deployment files created")
}

// Helper functions
func findDirWithSuffixBuild(suffix string) string {
	entries, err := os.ReadDir(".")
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
			return entry.Name()
		}
	}
	return ""
}

func fileExistsBuild(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func copyDir(src, dst string) error {
	if !dirExists(src) {
		return nil
	}

	os.MkdirAll(dst, 0755)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
