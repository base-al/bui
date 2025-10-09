package commands

import (
	"os"

	"github.com/base-al/bui/commands/backend"
	"github.com/base-al/bui/commands/frontend"
	"github.com/base-go/mamba"
)

var generateCmd = &mamba.Command{
	Use:     "generate [module] [field:type...]",
	Aliases: []string{"g"},
	Short:   "Generate modules",
	Long: `Generate modules for backend, frontend, or both.

Examples:
  bui g product name:string price:float          # Generate both backend and frontend
  bui g backend product name:string              # Backend only
  bui g frontend product name:string             # Frontend only`,
	Run: generateBothModules,
}

// generateBothModules generates both backend and frontend modules
func generateBothModules(cmd *mamba.Command, args []string) {
	if len(args) < 1 {
		cmd.PrintError("Module name required")
		cmd.PrintInfo("Usage: bui g [module] [field:type...]")
		os.Exit(1)
	}

	// Save the original working directory
	originalDir, err := os.Getwd()
	if err != nil {
		cmd.PrintError("Failed to get current directory")
		os.Exit(1)
	}

	// Set verbose pointers for subcommands
	backend.Verbose = &Verbose
	frontend.Verbose = &Verbose

	// Generate backend (subcommand handles its own logging)
	backend.GenerateBackendCmd.Run(cmd, args)

	// Return to original directory before generating frontend
	if err := os.Chdir(originalDir); err != nil {
		cmd.PrintError("Failed to return to original directory")
		os.Exit(1)
	}

	// Generate frontend (subcommand handles its own logging)
	frontend.GenerateFrontendCmd.Run(cmd, args)

	// Return to original directory after both generations
	if err := os.Chdir(originalDir); err != nil {
		cmd.PrintError("Failed to return to original directory")
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Add backend and frontend subcommands
	generateCmd.AddCommand(backend.GenerateBackendCmd)
	generateCmd.AddCommand(frontend.GenerateFrontendCmd)
}
