package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/base-go/mamba"
	"github.com/base-go/mamba/pkg/spinner"
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
	if err := cloneWithSpinner(cmd, "backend", "git@github.com:base-al/admin-api-template.git", "admin-api-template"); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to clone backend template: %v", err))
		cleanup(projectName)
		os.Exit(1)
	}

	// Clone frontend template with spinner
	if err := cloneWithSpinner(cmd, "frontend", "git@github.com:base-al/admin-template.git", "admin-template"); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to clone frontend template: %v", err))
		cleanup(projectName)
		os.Exit(1)
	}

	// Cleanup and initialize
	if err := cleanupAndInit(cmd, projectName); err != nil {
		cmd.PrintWarning(fmt.Sprintf("Setup incomplete: %v", err))
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

	return spinner.WithSpinner(fmt.Sprintf("Cloning %s template...", name), func() error {
		return cloneTemplate(repoURL, targetDir)
	})
}

func cleanupAndInit(cmd *mamba.Command, projectName string) error {
	// Remove .git directories from templates
	if Verbose {
		cmd.PrintInfo("Cleaning up template git histories...")
	}
	os.RemoveAll(filepath.Join("admin-api-template", ".git"))
	os.RemoveAll(filepath.Join("admin-template", ".git"))

	// Initialize new git repository
	if Verbose {
		cmd.PrintInfo("Initializing git repository...")
	}

	var err error
	if !Verbose {
		err = spinner.WithSpinner("Initializing project...", func() error {
			if err := initGitRepo(); err != nil {
				return err
			}
			createProjectReadme(projectName)
			return nil
		})
	} else {
		if err := initGitRepo(); err != nil {
			cmd.PrintWarning(fmt.Sprintf("Failed to initialize git: %v", err))
		} else {
			cmd.PrintSuccess("Git repository initialized")
		}
		cmd.PrintInfo("Creating project README...")
		createProjectReadme(projectName)
	}

	return err
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

func createProjectReadme(projectName string) {
	readme := fmt.Sprintf(`# %s

Base Stack project created with [Bui CLI](https://github.com/base-al/bui).

## Project Structure

- **admin-api-template/** - Backend API (Go + Base Framework)
- **admin-template/** - Frontend Admin Dashboard (Nuxt 4 + TypeScript)

## Getting Started

### Prerequisites

- Go 1.24+
- Bun (for frontend)
- PostgreSQL
- Redis (optional)

### Backend Setup

` + "```bash" + `
cd admin-api-template
cp .env.sample .env
# Edit .env with your database credentials
go mod tidy
bui start
` + "```" + `

Backend will run on http://localhost:8000

### Frontend Setup

` + "```bash" + `
cd admin-template
bun install
bun dev
` + "```" + `

Frontend will run on http://localhost:3030

### Development (Both Servers)

From project root:

` + "```bash" + `
bui dev
` + "```" + `

This starts both backend and frontend servers concurrently.

## Generating Modules

Generate a complete CRUD module for both backend and frontend:

` + "```bash" + `
# Generate both backend and frontend
bui g product name:string price:float description:text

# Backend only
bui g backend product name:string price:float

# Frontend only
bui g frontend product name:string price:float
` + "```" + `

## Documentation

- [Bui CLI Documentation](https://github.com/base-al/bui)
- [Backend Template](https://github.com/base-al/admin-api-template)
- [Frontend Template](https://github.com/base-al/admin-template)

## License

MIT
`, projectName)

	os.WriteFile("README.md", []byte(readme), 0644)
}

func printSuccessMessage(cmd *mamba.Command, projectName string) {
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
	cmd.PrintInfo("  cd admin-api-template")
	cmd.PrintInfo("  cp .env.sample .env")
	cmd.PrintInfo("  # Edit .env with your database credentials")
	cmd.PrintInfo("  go mod tidy")
	cmd.PrintInfo("  bui start")
	cmd.PrintInfo("")
	cmd.PrintInfo("Frontend setup (in another terminal):")
	cmd.PrintInfo("  cd admin-template")
	cmd.PrintInfo("  bun install")
	cmd.PrintInfo("  bun dev")
	cmd.PrintInfo("")
	cmd.PrintInfo("Or start both servers at once:")
	cmd.PrintInfo("  bui dev")
	cmd.PrintInfo("")
	cmd.PrintInfo("Generate your first module:")
	cmd.PrintInfo("  bui g product name:string price:float")
	cmd.PrintInfo("")
	cmd.PrintInfo("Happy coding! 🚀")
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
