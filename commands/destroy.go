package commands

import (
	"os"
	"path/filepath"

	"github.com/base-al/bui/utils"
	"github.com/base-go/mamba"
)

var destroyCmd = &mamba.Command{
	Use:     "destroy [module]",
	Aliases: []string{"d"},
	Short:   "Destroy backend and/or frontend modules",
	Long: `Destroy (delete) backend and/or frontend modules.

Examples:
  bui d product               # Destroy product module (frontend and backend)
  bui d backend product       # Destroy backend module only
  bui d frontend product      # Destroy frontend module only`,
	Run: destroyBothModules,
}

var destroyBackendCmd = &mamba.Command{
	Use:   "backend [module]",
	Short: "Destroy a backend module",
	Args:  mamba.ExactArgs(1),
	Run:   destroyBackend,
}

var destroyFrontendCmd = &mamba.Command{
	Use:   "frontend [module]",
	Short: "Destroy a frontend module",
	Args:  mamba.ExactArgs(1),
	Run:   destroyFrontend,
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.AddCommand(destroyBackendCmd)
	destroyCmd.AddCommand(destroyFrontendCmd)
}

func destroyBothModules(cmd *mamba.Command, args []string) {
	if len(args) < 1 {
		cmd.PrintError("module name required")
		cmd.PrintInfo("Usage: bui d [module] or bui d [backend|frontend] [module]")
		os.Exit(1)
	}

	moduleName := args[0]
	naming := utils.NewNamingConvention(moduleName)

	cmd.PrintWarning("Destroying module: " + naming.Model + " (backend + frontend)")

	// Detect project structure
	backendDir, frontendDir := detectProjectDirs()

	// Destroy backend
	backendPaths := []string{
		filepath.Join(backendDir, "app", "models", naming.ModelSnake+".go"),
		filepath.Join(backendDir, "app", naming.DirName),
	}

	backendDeleted := 0
	for _, path := range backendPaths {
		if _, err := os.Stat(path); err == nil {
			if err := os.RemoveAll(path); err != nil {
				cmd.PrintError("Failed to delete: " + path)
			} else {
				cmd.PrintInfo("Deleted: " + path)
				backendDeleted++
			}
		}
	}

	// Destroy frontend
	frontendPaths := []string{
		filepath.Join(frontendDir, "app", "modules", naming.PluralSnake),
		filepath.Join(frontendDir, "app", "pages", "app", naming.PluralKebab),
	}

	frontendDeleted := 0
	for _, path := range frontendPaths {
		if _, err := os.Stat(path); err == nil {
			if err := os.RemoveAll(path); err != nil {
				cmd.PrintError("Failed to delete: " + path)
			} else {
				cmd.PrintInfo("Deleted: " + path)
				frontendDeleted++
			}
		}
	}

	if backendDeleted == 0 && frontendDeleted == 0 {
		cmd.PrintWarning("No module found: " + naming.Model)
		return
	}

	if backendDeleted > 0 {
		cmd.PrintSuccess("Backend module destroyed: " + naming.Model)
		cmd.PrintInfo("Remember to remove from app/init.go if needed")
	}

	if frontendDeleted > 0 {
		cmd.PrintSuccess("Frontend module destroyed: " + naming.Model)
	}
}

// detectProjectDirs detects backend and frontend directories
func detectProjectDirs() (backend, frontend string) {
	// Check if we're in project root with separate backend/frontend dirs
	entries, err := os.ReadDir(".")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			// Look for *-api or *-backend directories
			if filepath.Ext(name) == "" && (contains(name, "-api") || contains(name, "-backend") || name == "backend" || name == "api") {
				// Check if it has app/ directory
				if _, err := os.Stat(filepath.Join(name, "app")); err == nil {
					backend = name
				}
			}
			// Look for *-app or *-frontend directories
			if filepath.Ext(name) == "" && (contains(name, "-app") || contains(name, "-frontend") || name == "frontend" || name == "app") {
				// Check if it has app/ directory
				if _, err := os.Stat(filepath.Join(name, "app")); err == nil {
					frontend = name
				}
			}
		}
	}

	// If not found, assume current directory
	if backend == "" {
		backend = "."
	}
	if frontend == "" {
		frontend = "."
	}

	return backend, frontend
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		(len(s) > len(substr) && (s[:len(substr)+1] == substr+"-" || s[len(s)-len(substr)-1:] == "-"+substr)))
}

func destroyBackend(cmd *mamba.Command, args []string) {
	moduleName := args[0]
	naming := utils.NewNamingConvention(moduleName)

	cmd.PrintWarning("Destroying backend module: " + naming.Model)

	// Paths to delete
	paths := []string{
		filepath.Join("app", "models", naming.ModelSnake+".go"),
		filepath.Join("app", naming.DirName),
	}

	deleted := 0
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if err := os.RemoveAll(path); err != nil {
				cmd.PrintError("Failed to delete: " + path)
			} else {
				cmd.PrintInfo("Deleted: " + path)
				deleted++
			}
		}
	}

	if deleted == 0 {
		cmd.PrintWarning("No backend module found: " + naming.Model)
		return
	}

	cmd.PrintSuccess("Backend module destroyed: " + naming.Model)
	cmd.PrintInfo("Remember to remove from app/init.go if needed")
}

func destroyFrontend(cmd *mamba.Command, args []string) {
	moduleName := args[0]
	naming := utils.NewNamingConvention(moduleName)

	cmd.PrintWarning("Destroying frontend module: " + naming.Model)

	// Paths to delete
	paths := []string{
		filepath.Join("app", "modules", naming.PluralSnake),
		filepath.Join("app", "pages", "app", naming.PluralKebab),
	}

	deleted := 0
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if err := os.RemoveAll(path); err != nil {
				cmd.PrintError("Failed to delete: " + path)
			} else {
				cmd.PrintInfo("Deleted: " + path)
				deleted++
			}
		}
	}

	if deleted == 0 {
		cmd.PrintWarning("No frontend module found: " + naming.Model)
		return
	}

	cmd.PrintSuccess("Frontend module destroyed: " + naming.Model)
}
