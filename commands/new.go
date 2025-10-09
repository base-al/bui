package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/base-go/mamba"
)

var newCmd = &mamba.Command{
	Use:   "new [project-name]",
	Short: "Create a new Base Stack project",
	Long: `Create a new Base Stack project by cloning backend and frontend templates.

This command will:
  1. Create a new project directory
  2. Clone admin-api-template (backend)
  3. Clone admin-template (frontend)
  4. Initialize git repository
  5. Set up configuration files

Example:
  bui new my-awesome-project`,
	Args: mamba.ExactArgs(1),
	Run:  createNewProject,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func createNewProject(cmd *mamba.Command, args []string) {
	projectName := args[0]

	// Validate project name
	if !isValidProjectName(projectName) {
		cmd.PrintError("Invalid project name")
		cmd.PrintInfo("Project name must contain only letters, numbers, hyphens, and underscores")
		os.Exit(1)
	}

	// Check if directory already exists
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		cmd.PrintError(fmt.Sprintf("Directory '%s' already exists", projectName))
		os.Exit(1)
	}

	cmd.PrintInfo(fmt.Sprintf("Creating new Base Stack project: %s", projectName))

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to create directory: %v", err))
		os.Exit(1)
	}

	// Change to project directory
	if err := os.Chdir(projectName); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to change directory: %v", err))
		os.Exit(1)
	}

	// Clone backend template with spinner
	backendDir := projectName + "-api"
	if err := cloneWithSpinner(cmd, "backend", "git@github.com:base-al/admin-api-template.git", backendDir); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to clone backend template: %v", err))
		cleanup(projectName)
		os.Exit(1)
	}

	// Clone frontend template with spinner
	frontendDir := projectName + "-app"
	if err := cloneWithSpinner(cmd, "frontend", "git@github.com:base-al/admin-template.git", frontendDir); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to clone frontend template: %v", err))
		cleanup(projectName)
		os.Exit(1)
	}

	// Cleanup and initialize
	if err := cleanupAndInit(cmd, projectName, backendDir, frontendDir); err != nil {
		cmd.PrintWarning(fmt.Sprintf("Setup incomplete: %v", err))
	}

	// Update configuration files
	if err := updateProjectFiles(cmd, projectName, backendDir, frontendDir); err != nil {
		cmd.PrintWarning(fmt.Sprintf("Failed to update project files: %v", err))
	}

	// Copy .env.example to .env
	if err := copyEnvFile(cmd, backendDir, frontendDir); err != nil {
		cmd.PrintWarning(fmt.Sprintf("Failed to copy .env.example to .env: %v", err))
	}

	// Print success message and next steps
	printSuccessMessage(cmd, projectName)
}

func cloneTemplate(repoURL, targetDir string) error {
	gitCmd := exec.Command("git", "clone", "--depth", "1", repoURL, targetDir)
	if Verbose {
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr
	}
	return gitCmd.Run()
}

func cloneWithSpinner(cmd *mamba.Command, name, repoURL, targetDir string) error {
	if Verbose {
		cmd.PrintInfo(fmt.Sprintf("Cloning %s template...", name))
		if err := cloneTemplate(repoURL, targetDir); err != nil {
			return err
		}
		cmd.PrintSuccess(fmt.Sprintf("%s template cloned", name))
		return nil
	}

	// Show spinner (non-blocking)
	cmd.PrintInfo(fmt.Sprintf("Cloning %s template...", name))

	// Clone without spinner wrapper to avoid deadlocks
	if err := cloneTemplate(repoURL, targetDir); err != nil {
		return fmt.Errorf("failed to clone %s: %w", name, err)
	}

	return nil
}

func updateGoImports(dir, projectName string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-.go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		contentStr := string(content)
		// Replace all "base/" imports with "projectName/"
		newContent := strings.ReplaceAll(contentStr, "\"base/", fmt.Sprintf("\"%s/", projectName))

		// Also update Swagger documentation comments in main.go
		if strings.HasSuffix(path, "/main.go") || strings.HasSuffix(path, "/Main.go") {
			titleCase := strings.ToUpper(projectName[:1]) + projectName[1:]
			lowerName := strings.ToLower(projectName)

			// Use regex to replace any API name (more flexible than hardcoded names)
			// Update @title - match any word before " API"
			titleRegex := regexp.MustCompile(`// @title (\w+) API`)
			newContent = titleRegex.ReplaceAllString(newContent, fmt.Sprintf("// @title %s API", titleCase))

			// Update @description - match "for [Name]"
			descRegex := regexp.MustCompile(`// @description This is the API documentation for (\w+)`)
			newContent = descRegex.ReplaceAllString(newContent, fmt.Sprintf("// @description This is the API documentation for %s", titleCase))

			// Update @contact.name - match "[Name] Team"
			contactNameRegex := regexp.MustCompile(`// @contact\.name (\w+) Team`)
			newContent = contactNameRegex.ReplaceAllString(newContent, fmt.Sprintf("// @contact.name %s Team", titleCase))

			// Update @contact.email - match "info@[domain].com"
			contactEmailRegex := regexp.MustCompile(`// @contact\.email info@(\w+)\.com`)
			newContent = contactEmailRegex.ReplaceAllString(newContent, fmt.Sprintf("// @contact.email info@%s.com", lowerName))

			// Update @contact.url - match "https://[domain].com"
			contactUrlRegex := regexp.MustCompile(`// @contact\.url https://(\w+)\.com`)
			newContent = contactUrlRegex.ReplaceAllString(newContent, fmt.Sprintf("// @contact.url https://%s.com", lowerName))

			// Update @termsOfService - match "https://[domain].com/terms"
			termsRegex := regexp.MustCompile(`// @termsOfService https://(\w+)\.com/terms`)
			newContent = termsRegex.ReplaceAllString(newContent, fmt.Sprintf("// @termsOfService https://%s.com/terms", lowerName))
		}

		// Only write if content changed
		if newContent != contentStr {
			if err := os.WriteFile(path, []byte(newContent), info.Mode()); err != nil {
				return err
			}
		}

		return nil
	})
}

func updateFrontendProjectStrings(dir, projectName string) error {
	// Create title case version of project name (capitalize first letter)
	titleCase := strings.ToUpper(projectName[:1]) + projectName[1:]
	projectAdmin := titleCase + " Admin"

	// Update login page (index.vue)
	indexPath := filepath.Join(dir, "app", "pages", "index.vue")
	if _, err := os.Stat(indexPath); err == nil {
		content, err := os.ReadFile(indexPath)
		if err != nil {
			return fmt.Errorf("failed to read index.vue: %w", err)
		}

		contentStr := string(content)
		// Replace BaseAdmin with ProjectName Admin
		contentStr = strings.ReplaceAll(contentStr, "BaseAdmin", projectAdmin)
		// Replace Admin Management System with custom description
		contentStr = strings.ReplaceAll(contentStr, "Admin Management System", projectAdmin+" Management System")
		// Replace example email placeholder
		contentStr = strings.ReplaceAll(contentStr, "admin@example.com", fmt.Sprintf("admin@%s.com", strings.ToLower(projectName)))

		if err := os.WriteFile(indexPath, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to write index.vue: %w", err)
		}
	}

	// Update auth store (app/stores/auth.ts)
	authStorePath := filepath.Join(dir, "app", "stores", "auth.ts")
	if _, err := os.Stat(authStorePath); err == nil {
		content, err := os.ReadFile(authStorePath)
		if err != nil {
			return fmt.Errorf("failed to read auth.ts: %w", err)
		}

		contentStr := string(content)
		// Replace localStorage key
		contentStr = strings.ReplaceAll(contentStr, "base_auth", fmt.Sprintf("%s_auth", strings.ToLower(projectName)))

		if err := os.WriteFile(authStorePath, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to write auth.ts: %w", err)
		}
	}

	// Update settings store (app/stores/settings.ts)
	settingsStorePath := filepath.Join(dir, "app", "stores", "settings.ts")
	if _, err := os.Stat(settingsStorePath); err == nil {
		content, err := os.ReadFile(settingsStorePath)
		if err != nil {
			return fmt.Errorf("failed to read settings.ts: %w", err)
		}

		contentStr := string(content)
		// Replace default company name
		contentStr = strings.ReplaceAll(contentStr, `|| 'Base'`, fmt.Sprintf(`|| '%s'`, titleCase))

		if err := os.WriteFile(settingsStorePath, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to write settings.ts: %w", err)
		}
	}

	return nil
}

func updateProjectFiles(cmd *mamba.Command, projectName, backendDir, frontendDir string) error {
	// Update backend go.mod
	goModPath := filepath.Join(backendDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		if Verbose {
			cmd.PrintInfo("Updating go.mod...")
		}

		content, err := os.ReadFile(goModPath)
		if err != nil {
			return fmt.Errorf("failed to read go.mod: %w", err)
		}

		contentStr := string(content)
		contentStr = strings.ReplaceAll(contentStr, "module base", fmt.Sprintf("module %s", projectName))

		if err := os.WriteFile(goModPath, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to write go.mod: %w", err)
		}

		if Verbose {
			cmd.PrintSuccess("Updated go.mod")
		}
	}

	// Update all .go files in backend to replace "base/" imports with projectName
	if Verbose {
		cmd.PrintInfo("Updating Go import statements...")
	}
	if err := updateGoImports(backendDir, projectName); err != nil {
		return fmt.Errorf("failed to update Go imports: %w", err)
	}
	if Verbose {
		cmd.PrintSuccess("Updated Go import statements")
	}

	// Update frontend package.json
	packageJsonPath := filepath.Join(frontendDir, "package.json")
	if _, err := os.Stat(packageJsonPath); err == nil {
		if Verbose {
			cmd.PrintInfo("Updating package.json...")
		}

		content, err := os.ReadFile(packageJsonPath)
		if err != nil {
			return fmt.Errorf("failed to read package.json: %w", err)
		}

		contentStr := string(content)
		// Replace the name field - handle both "nuxt-app" and "admin-template"
		contentStr = strings.ReplaceAll(contentStr, `"name": "nuxt-app"`, fmt.Sprintf(`"name": "%s"`, projectName))
		contentStr = strings.ReplaceAll(contentStr, `"name": "admin-template"`, fmt.Sprintf(`"name": "%s"`, projectName))

		if err := os.WriteFile(packageJsonPath, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to write package.json: %w", err)
		}

		if Verbose {
			cmd.PrintSuccess("Updated package.json")
		}
	}

	// Update frontend project-specific strings
	if Verbose {
		cmd.PrintInfo("Updating frontend project strings...")
	}
	if err := updateFrontendProjectStrings(frontendDir, projectName); err != nil {
		return fmt.Errorf("failed to update frontend strings: %w", err)
	}
	if Verbose {
		cmd.PrintSuccess("Updated frontend project strings")
	}

	return nil
}

func copyEnvFile(cmd *mamba.Command, backendDir, frontendDir string) error {
	// Copy .env.sample to .env for backend (backend uses .env.sample)
	if Verbose {
		cmd.PrintInfo("Setting up environment files...")
	}

	backendEnvSample := filepath.Join(backendDir, ".env.sample")
	backendEnv := filepath.Join(backendDir, ".env")

	// Check if .env.sample exists (backend)
	if _, err := os.Stat(backendEnvSample); err == nil {
		if err := copyFileNew(backendEnvSample, backendEnv); err != nil {
			cmd.PrintWarning(fmt.Sprintf("Failed to copy backend .env: %v", err))
		} else if Verbose {
			cmd.PrintSuccess("Created backend .env from .env.sample")
		}
	}

	// Copy .env.example to .env for frontend (if it exists)
	frontendEnvExample := filepath.Join(frontendDir, ".env.example")
	frontendEnv := filepath.Join(frontendDir, ".env")

	if _, err := os.Stat(frontendEnvExample); err == nil {
		if err := copyFileNew(frontendEnvExample, frontendEnv); err != nil {
			cmd.PrintWarning(fmt.Sprintf("Failed to copy frontend .env: %v", err))
		} else if Verbose {
			cmd.PrintSuccess("Created frontend .env from .env.example")
		}
	}

	if Verbose {
		cmd.PrintSuccess("Environment setup complete")
	}

	// Check if bun is installed
	if _, err := exec.LookPath("bun"); err != nil {
		cmd.PrintWarning("Bun is not installed. Skipping frontend dependency installation.")
		cmd.PrintInfo("Please install Bun from https://bun.sh and run 'bun install' in the frontend directory.")
		return nil
	}

	// Run bun install
	if Verbose {
		cmd.PrintInfo("Installing frontend dependencies...")
	}
	bunInstallCmd := exec.Command("bun", "install")
	bunInstallCmd.Dir = frontendDir
	bunInstallCmd.Stdout = os.Stdout
	bunInstallCmd.Stderr = os.Stderr

	if err := bunInstallCmd.Run(); err != nil {
		cmd.PrintWarning(fmt.Sprintf("Failed to run bun install: %v", err))
		cmd.PrintInfo(fmt.Sprintf("Please run 'bun install' manually in %s", frontendDir))
		return nil
	}

	if Verbose {
		cmd.PrintSuccess("Frontend dependencies installed")
	}

	return nil
}

func cleanupAndInit(cmd *mamba.Command, projectName, backendDir, frontendDir string) error {
	// Remove .git directories from templates
	if Verbose {
		cmd.PrintInfo("Cleaning up template git histories...")
	}
	os.RemoveAll(filepath.Join(backendDir, ".git"))
	os.RemoveAll(filepath.Join(frontendDir, ".git"))

	// Initialize new git repository
	if !Verbose {
		cmd.PrintInfo("Initializing project...")
	} else {
		cmd.PrintInfo("Initializing git repository...")
	}

	if err := initGitRepo(); err != nil {
		cmd.PrintWarning(fmt.Sprintf("Failed to initialize git: %v", err))
	} else if Verbose {
		cmd.PrintSuccess("Git repository initialized")
	}

	if Verbose {
		cmd.PrintInfo("Creating project README...")
	}
	createProjectReadme(projectName, backendDir, frontendDir)

	return nil
}

func initGitRepo() error {
	// Initialize git
	if err := exec.Command("git", "init").Run(); err != nil {
		return err
	}

	// Create .gitignore
	gitignoreContent := `.DS_Store
Thumbs.db
*.swp
*.swo
*~
.vscode/
.idea/

# Production build
dist/
*-dist/
*.tar.gz
deploy.tar.gz
`
	if err := os.WriteFile(".gitignore", []byte(gitignoreContent), 0644); err != nil {
		return err
	}

	// Add all files
	if err := exec.Command("git", "add", ".").Run(); err != nil {
		return err
	}

	// Create initial commit
	commitMsg := "Initial commit from Base Stack templates"
	if err := exec.Command("git", "commit", "-m", commitMsg).Run(); err != nil {
		return err
	}

	return nil
}

func createProjectReadme(projectName, backendDir, frontendDir string) {
	readme := fmt.Sprintf(`# %s

Base Stack project created with [Bui CLI](https://github.com/base-al/bui).

## Project Structure

- **%s/** - Backend API (Go + Base Framework)
- **%s/** - Frontend Admin Dashboard (Nuxt 4 + TypeScript)

## Getting Started

### Prerequisites

- Go 1.24+
- Bun (for frontend)
- PostgreSQL
- Redis (optional)

### Backend Setup

`+"```bash"+`
cd %s
cp .env.sample .env
# Edit .env with your database credentials
go mod tidy
bui start
`+"```"+`

Backend will run on http://localhost:8000

### Frontend Setup

`+"```bash"+`
cd %s
bun install
bun dev
`+"```"+`

Frontend will run on http://localhost:3030

### Development (Both Servers)

From project root:

`+"```bash"+`
bui dev
`+"```"+`

This starts both backend and frontend servers concurrently.

## Generating Modules

Generate a complete CRUD module for both backend and frontend:

`+"```bash"+`
# Generate both backend and frontend
bui g product name:string price:float description:text

# Backend only
bui g backend product name:string price:float

# Frontend only
bui g frontend product name:string price:float
`+"```"+`

## Documentation

- [Bui CLI Documentation](https://github.com/base-al/bui)
- [Backend Template](https://github.com/base-al/admin-api-template)
- [Frontend Template](https://github.com/base-al/admin-template)

## License

MIT
`, projectName, backendDir, frontendDir, backendDir, frontendDir)

	os.WriteFile("README.md", []byte(readme), 0644)
}

func printSuccessMessage(cmd *mamba.Command, projectName string) {
	backendDir := projectName + "-api"
	frontendDir := projectName + "-app"

	cmd.PrintInfo("")
	cmd.PrintInfo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	cmd.PrintSuccess(fmt.Sprintf("Project '%s' created successfully!", projectName))
	cmd.PrintInfo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	cmd.PrintInfo("")
	cmd.PrintInfo("Next steps:")
	cmd.PrintInfo("")
	cmd.PrintInfo(fmt.Sprintf("  cd %s", projectName))
	cmd.PrintInfo("")
	cmd.PrintInfo("Backend setup:")
	cmd.PrintInfo(fmt.Sprintf("  cd %s", backendDir))
	cmd.PrintInfo("  cp .env.sample .env")
	cmd.PrintInfo("  # Edit .env with your database credentials")
	cmd.PrintInfo("  go mod tidy")
	cmd.PrintInfo("  bui start")
	cmd.PrintInfo("")
	cmd.PrintInfo("Frontend setup (in another terminal):")
	cmd.PrintInfo(fmt.Sprintf("  cd %s", frontendDir))
	cmd.PrintInfo("  bun install")
	cmd.PrintInfo("  bun dev")
	cmd.PrintInfo("")
	cmd.PrintInfo("Or start both servers at once:")
	cmd.PrintInfo("  bui dev")
	cmd.PrintInfo("")
	cmd.PrintInfo("Generate your first module:")
	cmd.PrintInfo("  bui g product name:string price:float")
	cmd.PrintInfo("")
	cmd.PrintInfo("Happy coding!")
	cmd.PrintInfo("")
}

func isValidProjectName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_') {
			return false
		}
	}
	return true
}

func cleanup(projectName string) {
	os.Chdir("..")
	os.RemoveAll(projectName)
}

// copyFileNew copies a file from src to dst
func copyFileNew(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
