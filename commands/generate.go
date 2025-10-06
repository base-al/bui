package commands

import (
	"fmt"
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
		fmt.Println("Error: module name required")
		fmt.Println("Usage: bui g [module] [field:type...]")
		os.Exit(1)
	}

	// Generate backend
	fmt.Println("\nGenerating backend module...")
	backend.GenerateBackendCmd.Run(cmd, args)

	// Generate frontend
	fmt.Println("\nGenerating frontend module...")
	frontend.GenerateFrontendCmd.Run(cmd, args)

	fmt.Printf("\n✓ Successfully generated %s module (backend + frontend)\n", args[0])
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Add backend and frontend subcommands
	generateCmd.AddCommand(backend.GenerateBackendCmd)
	generateCmd.AddCommand(frontend.GenerateFrontendCmd)
}
