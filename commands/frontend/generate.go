package frontend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/base-al/bui/utils"
	"github.com/base-go/mamba"
)

var GenerateFrontendCmd = &mamba.Command{
	Use:     "frontend [name] [field:type...]",
	Aliases: []string{"fe", "ui"},
	Short:   "Generate a frontend Nuxt module",
	Long:    `Generate a Nuxt module with TypeScript types, Pinia store, Vue components, and pages.`,
	Args:    mamba.MinimumNArgs(1),
	Run:     generateFrontendModule,
}

// generateFrontendModule generates a new frontend module with the specified name and fields
func generateFrontendModule(cmd *mamba.Command, args []string) {
	singularName := args[0]
	fields := args[1:]

	// Detect frontend directory
	frontendDir := detectFrontendDir()
	if frontendDir != "" && frontendDir != "." {
		// Change to frontend directory
		if err := os.Chdir(frontendDir); err != nil {
			cmd.PrintError(fmt.Sprintf("Failed to change to frontend directory: %v", err))
			return
		}
		if Verbose != nil && *Verbose {
			cmd.PrintInfo(fmt.Sprintf("Working in: %s", frontendDir))
		}
	}

	// Create naming convention from the input name
	naming := utils.NewNamingConvention(singularName)

	// Base path for app directory
	adminPath := "app"

	// Create directories
	moduleBasePath := filepath.Join(adminPath, "modules", naming.PluralSnake)
	dirs := []string{
		filepath.Join(moduleBasePath, "types"),
		filepath.Join(moduleBasePath, "stores"),
		filepath.Join(moduleBasePath, "components"),
		filepath.Join(moduleBasePath, "utils"),
		filepath.Join(adminPath, "pages", "app", naming.PluralKebab),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	// Parse fields
	parsedFields := make([]utils.Field, 0, len(fields))
	for _, fieldDef := range fields {
		parsedFields = append(parsedFields, utils.ParseField(fieldDef))
	}

	// Convert to Nuxt fields with TypeScript types
	nuxtFields := make([]utils.NuxtField, 0, len(parsedFields))
	for _, field := range parsedFields {
		nuxtFields = append(nuxtFields, utils.ConvertToNuxtField(field))
	}

	// Template data combining naming and fields
	type TemplateData struct {
		*utils.NamingConvention
		Fields []utils.NuxtField
	}

	templateData := &TemplateData{
		NamingConvention: naming,
		Fields:           nuxtFields,
	}

	fmt.Printf("Generating module: %s\n", naming.Model)

	// Generate module.config.ts
	if err := utils.GenerateNuxtFile(
		moduleBasePath,
		"module.config.ts",
		"nuxt/module.config.ts.tmpl",
		templateData,
	); err != nil {
		fmt.Printf("Error generating module.config.ts: %v\n", err)
		return
	}
	fmt.Printf("✅ Generated module.config.ts\n")

	// Generate types file
	if err := utils.GenerateNuxtFile(
		filepath.Join(moduleBasePath, "types"),
		naming.ModelSnake+".ts",
		"nuxt/types.ts.tmpl",
		templateData,
	); err != nil {
		fmt.Printf("Error generating types: %v\n", err)
		return
	}
	fmt.Printf("✅ Generated types/%s.ts\n", naming.ModelSnake)

	// Generate store
	if err := utils.GenerateNuxtFile(
		filepath.Join(moduleBasePath, "stores"),
		naming.PluralSnake+".ts",
		"nuxt/store.ts.tmpl",
		templateData,
	); err != nil {
		fmt.Printf("Error generating store: %v\n", err)
		return
	}
	fmt.Printf("✅ Generated stores/%s.ts\n", naming.PluralSnake)

	// Generate table component
	if err := utils.GenerateNuxtFile(
		filepath.Join(moduleBasePath, "components"),
		naming.Model+"Table.vue",
		"nuxt/table.vue.tmpl",
		templateData,
	); err != nil {
		fmt.Printf("Error generating table component: %v\n", err)
		return
	}
	fmt.Printf("✅ Generated components/%sTable.vue\n", naming.Model)

	// Generate form modal component
	if err := utils.GenerateNuxtFile(
		filepath.Join(moduleBasePath, "components"),
		naming.Model+"FormModal.vue",
		"nuxt/form-modal.vue.tmpl",
		templateData,
	); err != nil {
		fmt.Printf("Error generating form modal: %v\n", err)
		return
	}
	fmt.Printf("✅ Generated components/%sFormModal.vue\n", naming.Model)

	// Generate formatters utils
	if err := utils.GenerateNuxtFile(
		filepath.Join(moduleBasePath, "utils"),
		"formatters.ts",
		"nuxt/formatters.ts.tmpl",
		templateData,
	); err != nil {
		fmt.Printf("Error generating formatters: %v\n", err)
		return
	}
	fmt.Printf("✅ Generated utils/formatters.ts\n")

	// Generate index page
	if err := utils.GenerateNuxtFile(
		filepath.Join(adminPath, "pages", "app", naming.PluralKebab),
		"index.vue",
		"nuxt/index.vue.tmpl",
		templateData,
	); err != nil {
		fmt.Printf("Error generating index page: %v\n", err)
		return
	}
	fmt.Printf("✅ Generated pages/app/%s/index.vue\n", naming.PluralKebab)

	// Generate detail page
	if err := utils.GenerateNuxtFile(
		filepath.Join(adminPath, "pages", "app", naming.PluralKebab),
		"[id].vue",
		"nuxt/detail.vue.tmpl",
		templateData,
	); err != nil {
		fmt.Printf("Error generating detail page: %v\n", err)
		return
	}
	fmt.Printf("✅ Generated pages/app/%s/[id].vue\n", naming.PluralKebab)

	fmt.Printf("\nSuccessfully generated module: %s\n", naming.Model)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Module automatically discovered via app/modules/index.ts\n")
	fmt.Printf("  2. Navigate to /app/%s to see your new module\n", naming.PluralKebab)
	fmt.Printf("  3. Ensure backend API endpoints available at /api/%s\n", naming.PluralSnake)
	fmt.Printf("\nTip: Use 'bui g frontend %s' to regenerate or 'bui d frontend %s' to remove\n", naming.ModelSnake, naming.ModelSnake)
}

// detectFrontendDir finds the frontend directory in the current working directory
func detectFrontendDir() string {
	// Check if we're already in a frontend directory
	if _, err := os.Stat("nuxt.config.ts"); err == nil {
		if _, err := os.Stat(filepath.Join("app", "pages")); err == nil {
			return "." // Already in frontend directory
		}
	}

	// Check for directories with -app suffix
	entries, err := os.ReadDir(".")
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), "-app") {
			// Check if it has nuxt.config.ts
			if _, err := os.Stat(filepath.Join(entry.Name(), "nuxt.config.ts")); err == nil {
				return entry.Name()
			}
		}
	}

	// Check for standard names
	standardNames := []string{"admin-template", "admin", "frontend", "app"}
	for _, name := range standardNames {
		if _, err := os.Stat(filepath.Join(name, "nuxt.config.ts")); err == nil {
			return name
		}
	}

	return "" // No frontend directory found
}
