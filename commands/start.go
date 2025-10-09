package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/base-go/mamba"
	"github.com/base-go/mamba/pkg/spinner"
)

var (
	docs bool
)

var startCmd = &mamba.Command{
	Use:     "start",
	Aliases: []string{"s"},
	Short:   "Start the application",
	Long:    `Start the application by running the Base application server.`,
	Run:     startApplication,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&docs, "docs", "d", false, "Generate Swagger documentation")
}

func startApplication(c *mamba.Command, args []string) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		c.PrintError(fmt.Sprintf("Failed to get working directory: %v", err))
		return
	}

	// Check if we're in a Base project by looking for main.go
	mainPath := filepath.Join(cwd, "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		c.PrintError("Base project structure not found")
		c.PrintInfo("Make sure you are in the root directory of your Base project")
		c.PrintInfo(fmt.Sprintf("Expected to find main.go at: %s", mainPath))
		return
	}

	// Find go executable using which
	whichCmd := exec.Command("which", "go")
	goPathBytes, err := whichCmd.Output()
	if err != nil {
		c.PrintError("Go executable not found")
		c.PrintInfo("Please ensure Go is properly installed and in your PATH")
		return
	}
	goPath := strings.TrimSpace(string(goPathBytes))

	// Run go mod tidy to ensure dependencies are up to date
	if Verbose {
		c.PrintInfo("Running go mod tidy...")
		tidyCmd := exec.Command(goPath, "mod", "tidy")
		tidyCmd.Dir = cwd
		if err := tidyCmd.Run(); err != nil {
			c.PrintWarning(fmt.Sprintf("Failed to run go mod tidy: %v", err))
		} else {
			c.PrintSuccess("Dependencies updated")
		}
	} else {
		err := spinner.WithSpinner("Updating dependencies...", func() error {
			tidyCmd := exec.Command(goPath, "mod", "tidy")
			tidyCmd.Dir = cwd
			return tidyCmd.Run()
		})
		if err != nil {
			c.PrintWarning(fmt.Sprintf("Failed to run go mod tidy: %v", err))
		}
	}

	if docs {
		c.PrintHeader("Documentation Generation")

		// Ensure swag is installed
		if _, err := exec.LookPath("swag"); err != nil {
			if Verbose {
				c.PrintInfo("Installing swag...")
				installCmd := exec.Command(goPath, "install", "github.com/swaggo/swag/cmd/swag@latest")
				installCmd.Stdout = os.Stdout
				installCmd.Stderr = os.Stderr
				if err := installCmd.Run(); err != nil {
					c.PrintWarning(fmt.Sprintf("Failed to install swag: %v", err))
				} else {
					c.PrintSuccess("Swag installed successfully")
				}
			} else {
				err := spinner.WithSpinner("Installing swag...", func() error {
					installCmd := exec.Command(goPath, "install", "github.com/swaggo/swag/cmd/swag@latest")
					return installCmd.Run()
				})
				if err != nil {
					c.PrintWarning(fmt.Sprintf("Failed to install swag: %v", err))
				}
			}
		}

		// Generate swagger docs using swag
		if Verbose {
			c.PrintInfo("Generating Swagger documentation...")
			swagCmd := exec.Command("swag", "init", "--dir", "./", "--output", "./swagger", "--parseDependency", "--parseInternal", "--parseVendor", "--parseDepth", "1", "--generatedTime", "false")
			swagCmd.Dir = cwd
			swagCmd.Stdout = os.Stdout
			swagCmd.Stderr = os.Stderr

			if err := swagCmd.Run(); err != nil {
				c.PrintWarning(fmt.Sprintf("Failed to generate docs: %v", err))
			} else {
				c.PrintSuccess("Swagger documentation will be available at /swagger/")
			}
		} else {
			err := spinner.WithSpinner("Generating Swagger docs...", func() error {
				swagCmd := exec.Command("swag", "init", "--dir", "./", "--output", "./swagger", "--parseDependency", "--parseInternal", "--parseVendor", "--parseDepth", "1", "--generatedTime", "false")
				swagCmd.Dir = cwd
				return swagCmd.Run()
			})
			if err != nil {
				c.PrintWarning(fmt.Sprintf("Failed to generate docs: %v", err))
			} else {
				c.PrintSuccess("Swagger documentation will be available at /swagger/")
			}
		}
	}

	// Run normally
	c.PrintInfo("Starting the Base application server...")

	mainCmd := exec.Command(goPath, "run", "main.go")
	mainCmd.Stdout = os.Stdout
	mainCmd.Stderr = os.Stderr
	mainCmd.Dir = cwd

	// Set environment variables
	env := os.Environ()
	if docs {
		env = append(env, "SWAGGER_ENABLED=true")
	}
	mainCmd.Env = env

	if err := mainCmd.Run(); err != nil {
		c.PrintError(fmt.Sprintf("Failed to run application: %v", err))
		return
	}
}
