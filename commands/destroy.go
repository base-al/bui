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

	// Destroy backend
	backendPaths := []string{
		filepath.Join("app", "models", naming.ModelSnake+".go"),
		filepath.Join("app", naming.DirName),
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
		filepath.Join("app", "modules", naming.PluralSnake),
		filepath.Join("app", "pages", "app", naming.PluralKebab),
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
