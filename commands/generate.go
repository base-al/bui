package commands

import (
	"os"

	"github.com/base-al/bui/commands/backend"
	"github.com/base-al/bui/commands/frontend"
	"github.com/base-go/mamba"
	"github.com/base-go/mamba/pkg/spinner"
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

	moduleName := args[0]

	// Set verbose pointers for subcommands
	backend.Verbose = &Verbose
	frontend.Verbose = &Verbose

	// Generate backend
	if !Verbose {
		err := spinner.WithSpinner("Generating backend module...", func() error {
			backend.GenerateBackendCmd.Run(cmd, args)
			return nil
		})
		if err != nil {
			cmd.PrintError("Failed to generate backend module")
			os.Exit(1)
		}
	} else {
		cmd.PrintInfo("Generating backend module...")
		backend.GenerateBackendCmd.Run(cmd, args)
	}

	// Generate frontend
	if !Verbose {
		err := spinner.WithSpinner("Generating frontend module...", func() error {
			frontend.GenerateFrontendCmd.Run(cmd, args)
			return nil
		})
		if err != nil {
			cmd.PrintError("Failed to generate frontend module")
			os.Exit(1)
		}
	} else {
		cmd.PrintInfo("Generating frontend module...")
		frontend.GenerateFrontendCmd.Run(cmd, args)
	}

	cmd.PrintSuccess("Successfully generated " + moduleName + " module (backend + frontend)")
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Add backend and frontend subcommands
	generateCmd.AddCommand(backend.GenerateBackendCmd)
	generateCmd.AddCommand(frontend.GenerateFrontendCmd)
}
