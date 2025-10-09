package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/base-al/bui/utils"
	"github.com/base-go/mamba"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Verbose is set by root command
var Verbose *bool

var GenerateBackendCmd = &mamba.Command{
	Use:     "backend [name] [field:type...]",
	Aliases: []string{"be", "api"},
	Short:   "Generate a backend Go module",
	Long:    `Generate a backend module with model, service, controller, and validator.`,
	Args:    mamba.MinimumNArgs(1),
	Run:     generateBackendModule,
}

// generateBackendModule generates a new backend module with the specified name and fields.
func generateBackendModule(cmd *mamba.Command, args []string) {
	singularName := args[0]
	fields := args[1:]

	// Detect backend directory
	backendDir := detectBackendDir()
	if backendDir != "" && backendDir != "." {
		// Change to backend directory
		if err := os.Chdir(backendDir); err != nil {
			cmd.PrintError(fmt.Sprintf("Failed to change to backend directory: %v", err))
			return
		}
		if Verbose != nil && *Verbose {
			cmd.PrintInfo(fmt.Sprintf("Working in: %s", backendDir))
		}
	}

	// Create naming convention from the input name
	naming := utils.NewNamingConvention(singularName)

	// Create directories (plural names in snake_case)
	dirs := []string{
		filepath.Join("app", "models"),
		filepath.Join("app", naming.DirName),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			cmd.PrintError(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
			return
		}
		if Verbose != nil && *Verbose {
			cmd.PrintInfo(fmt.Sprintf("Created directory: %s", dir))
		}
	}

	// Generate field structs and set module name
	fieldStructs := utils.NewTemplateData(naming.Model, fields)
	fieldStructs.ModuleName = getGoModuleName()

	// Generate model
	utils.GenerateFileFromTemplate(
		filepath.Join("app", "models"),
		naming.ModelSnake+".go",
		"model.tmpl",
		naming,
		fieldStructs.Fields,
	)
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated app/models/%s.go", naming.ModelSnake))
	}

	// Generate service
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"service.go",
		"service.tmpl",
		naming,
		fieldStructs.Fields,
	)
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated app/%s/service.go", naming.DirName))
	}

	// Generate controller
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"controller.go",
		"controller.tmpl",
		naming,
		fieldStructs.Fields,
	)
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated app/%s/controller.go", naming.DirName))
	}

	// Generate module
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"module.go",
		"module.tmpl",
		naming,
		fieldStructs.Fields,
	)
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated app/%s/module.go", naming.DirName))
	}

	// Generate validator
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"validator.go",
		"validator.tmpl",
		naming,
		fieldStructs.Fields,
	)
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated app/%s/validator.go", naming.DirName))
	}

	// Generate tests - disabled for now, will be added in future
	// if err := utils.GenerateTests(naming, fieldStructs); err != nil {
	// 	fmt.Printf("Error generating tests: %v\n", err)
	// 	return
	// }

	// Check if goimports is installed
	if _, err := exec.LookPath("goimports"); err != nil {
		if Verbose != nil && *Verbose {
			cmd.PrintInfo("Installing goimports...")
		}
		if err := exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest").Run(); err != nil {
			cmd.PrintWarning("Failed to install goimports")
			if Verbose != nil && *Verbose {
				cmd.PrintInfo("Install manually: go install golang.org/x/tools/cmd/goimports@latest")
			}
			return
		}
		if Verbose != nil && *Verbose {
			cmd.PrintSuccess("goimports installed")
		}
	}

	// Run goimports on generated files
	generatedPath := filepath.Join("app", naming.DirName)

	if Verbose != nil && *Verbose {
		cmd.PrintInfo("Formatting generated files...")
	}

	// Run goimports on the generated directory
	if err := exec.Command("find", generatedPath, "-name", "*.go", "-exec", "goimports", "-w", "{}", ";").Run(); err != nil {
		if Verbose != nil && *Verbose {
			cmd.PrintWarning(fmt.Sprintf("Failed to run goimports on %s", generatedPath))
		}
	}

	// Run goimports on the model file
	modelPath := filepath.Join("app", "models", naming.ModelSnake+".go")
	if err := exec.Command("goimports", "-w", modelPath).Run(); err != nil {
		if Verbose != nil && *Verbose {
			cmd.PrintWarning(fmt.Sprintf("Failed to run goimports on %s", modelPath))
		}
	}

	// Format all generated files with gofmt
	if err := exec.Command("gofmt", "-w", generatedPath).Run(); err != nil {
		if Verbose != nil && *Verbose {
			cmd.PrintWarning(fmt.Sprintf("Failed to format %s", generatedPath))
		}
	}
	if err := exec.Command("gofmt", "-w", modelPath).Run(); err != nil {
		if Verbose != nil && *Verbose {
			cmd.PrintWarning(fmt.Sprintf("Failed to format %s", modelPath))
		}
	}

	// Add module to app/init.go
	if err := addModuleToAppInit(naming.DirName); err != nil {
		cmd.PrintWarning("Could not add module to app/init.go")
		cmd.PrintInfo(fmt.Sprintf("Manually add to app/init.go: modules[\"%s\"] = %s.Init(deps)", naming.DirName, naming.DirName))
	} else {
		if Verbose != nil && *Verbose {
			cmd.PrintSuccess("Added module to app/init.go")
		}

		// Format init.go after modification
		initGoPath := filepath.Join("app", "init.go")
		if err := exec.Command("gofmt", "-w", initGoPath).Run(); err != nil {
			if Verbose != nil && *Verbose {
				cmd.PrintWarning("Failed to format app/init.go")
			}
		}
	}

	// Run go mod tidy to ensure dependencies are up to date
	if Verbose != nil && *Verbose {
		cmd.PrintInfo("Running go mod tidy...")
	}
	if err := exec.Command("go", "mod", "tidy").Run(); err != nil {
		if Verbose != nil && *Verbose {
			cmd.PrintWarning("Failed to run go mod tidy")
		}
	}

	if Verbose == nil || !*Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated backend module: %s", naming.Model))
	}
}

// addModuleToAppInit adds the module to app/init.go
func addModuleToAppInit(moduleName string) error {
	initGoPath := filepath.Join("app", "init.go")
	goModuleName := getGoModuleName()

	// Check if app/init.go exists
	if _, err := os.Stat(initGoPath); os.IsNotExist(err) {
		// Create app/init.go if it doesn't exist
		if err := os.MkdirAll("app", os.ModePerm); err != nil {
			return fmt.Errorf("failed to create app directory: %w", err)
		}

		content := fmt.Sprintf(`package app

import (
	"%s/app/%s"
	"%s/core/module"
)

// AppModules implements module.AppModuleProvider interface
type AppModules struct{}

// GetAppModules returns the list of app modules to initialize
// This is the only function that needs to be updated when adding new app modules
func (am *AppModules) GetAppModules(deps module.Dependencies) map[string]module.Module {
	modules := make(map[string]module.Module)

	// App modules - custom system functionality
	modules["%s"] = %s.Init(deps)

	return modules
}

// NewAppModules creates a new AppModules provider
func NewAppModules() *AppModules {
	return &AppModules{}
}
`, goModuleName, moduleName, goModuleName, moduleName, moduleName)

		if err := os.WriteFile(initGoPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create app/init.go: %w", err)
		}
		return nil
	}

	// Read existing app/init.go content
	content, err := os.ReadFile(initGoPath)
	if err != nil {
		return fmt.Errorf("failed to read app/init.go: %w", err)
	}

	contentStr := string(content)

	// Check if module already exists
	moduleInit := fmt.Sprintf("modules[\"%s\"] = %s.Init(deps)", moduleName, moduleName)
	if strings.Contains(contentStr, moduleInit) {
		return nil // Already added
	}

	// Add import if not exists using the proper AddImport function
	importLine := fmt.Sprintf("\"%s/app/%s\"", goModuleName, moduleName)
	contentBytes, importAdded := utils.AddImport([]byte(contentStr), importLine)
	if importAdded {
		contentStr = string(contentBytes)
	}

	// Add module initialization
	// Find the return modules line
	returnIndex := strings.Index(contentStr, "return modules")
	if returnIndex == -1 {
		return fmt.Errorf("could not find 'return modules' in app/init.go")
	}

	// Insert the module initialization before return
	insertPoint := returnIndex - 1
	caser := cases.Title(language.English)
	moduleInitLine := fmt.Sprintf("\n\t// %s module\n\t%s\n", caser.String(moduleName), moduleInit)
	contentStr = contentStr[:insertPoint] + moduleInitLine + contentStr[insertPoint:]

	// Write back to file
	if err := os.WriteFile(initGoPath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write app/init.go: %w", err)
	}

	return nil
}

// getGoModuleName reads the module name from go.mod
func getGoModuleName() string {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return "base" // fallback to default
	}

	// Parse first line: "module <name>"
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}

	return "base" // fallback to default
}

// detectBackendDir finds the backend directory in the current working directory
func detectBackendDir() string {
	// Check if we're already in a backend directory
	if _, err := os.Stat("main.go"); err == nil {
		if _, err := os.Stat(filepath.Join("app", "models")); err == nil {
			return "." // Already in backend directory
		}
	}

	// Check for directories with -api suffix
	entries, err := os.ReadDir(".")
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), "-api") {
			// Check if it has main.go
			if _, err := os.Stat(filepath.Join(entry.Name(), "main.go")); err == nil {
				return entry.Name()
			}
		}
	}

	// Check for standard names
	standardNames := []string{"admin-api-template", "admin-api", "backend", "api"}
	for _, name := range standardNames {
		if _, err := os.Stat(filepath.Join(name, "main.go")); err == nil {
			return name
		}
	}

	return "" // No backend directory found
}
