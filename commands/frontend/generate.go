package frontend

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/base-al/bui/utils"
	"github.com/base-go/mamba"
)

// Verbose is set by root command
var Verbose *bool

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
			cmd.PrintError(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
			return
		}
		if Verbose != nil && *Verbose {
			cmd.PrintInfo(fmt.Sprintf("Created directory: %s", dir))
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
		nf := utils.ConvertToNuxtField(field)

		// For belongs_to relations, fetch the display field from the related model's type file
		if field.IsRelation && field.Relationship == "belongs_to" && field.RelatedModel != "" {
			relatedDisplayField := getRelatedModelDisplayField(adminPath, field.RelatedModel)
			nf.RelationDisplayField = relatedDisplayField
		}

		nuxtFields = append(nuxtFields, nf)
	}

	// Determine display field (first non-relation string field)
	displayField := "id" // fallback
	for _, field := range parsedFields {
		if !field.IsRelation && !field.IsMediaFK && (field.Type == "string" || field.Type == "translation.Field") {
			displayField = field.JSONName
			break
		}
	}

	// Template data combining naming and fields
	type TemplateData struct {
		*utils.NamingConvention
		Fields       []utils.NuxtField
		DisplayField string
	}

	templateData := &TemplateData{
		NamingConvention: naming,
		Fields:           nuxtFields,
		DisplayField:     displayField,
	}

	// Generate module.config.ts
	if err := utils.GenerateNuxtFile(
		moduleBasePath,
		"module.config.ts",
		"nuxt/module.config.ts.tmpl",
		templateData,
	); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to generate module.config.ts: %v", err))
		return
	}
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess("Generated module.config.ts")
	}

	// Generate types file
	if err := utils.GenerateNuxtFile(
		filepath.Join(moduleBasePath, "types"),
		naming.ModelSnake+".ts",
		"nuxt/types.ts.tmpl",
		templateData,
	); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to generate types: %v", err))
		return
	}
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated types/%s.ts", naming.ModelSnake))
	}

	// Generate store
	if err := utils.GenerateNuxtFile(
		filepath.Join(moduleBasePath, "stores"),
		naming.PluralSnake+".ts",
		"nuxt/store.ts.tmpl",
		templateData,
	); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to generate store: %v", err))
		return
	}
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated stores/%s.ts", naming.PluralSnake))
	}

	// Generate form modal component
	if err := utils.GenerateNuxtFile(
		filepath.Join(moduleBasePath, "components"),
		naming.Model+"FormModal.vue",
		"nuxt/form-modal.vue.tmpl",
		templateData,
	); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to generate form modal: %v", err))
		return
	}
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated components/%sFormModal.vue", naming.Model))
	}

	// Generate formatters utils
	if err := utils.GenerateNuxtFile(
		filepath.Join(moduleBasePath, "utils"),
		"formatters.ts",
		"nuxt/formatters.ts.tmpl",
		templateData,
	); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to generate formatters: %v", err))
		return
	}
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess("Generated utils/formatters.ts")
	}

	// Generate index page
	if err := utils.GenerateNuxtFile(
		filepath.Join(adminPath, "pages", "app", naming.PluralKebab),
		"index.vue",
		"nuxt/index.vue.tmpl",
		templateData,
	); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to generate index page: %v", err))
		return
	}
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated pages/app/%s/index.vue", naming.PluralKebab))
	}

	// Generate detail page
	if err := utils.GenerateNuxtFile(
		filepath.Join(adminPath, "pages", "app", naming.PluralKebab),
		"[id].vue",
		"nuxt/detail.vue.tmpl",
		templateData,
	); err != nil {
		cmd.PrintError(fmt.Sprintf("Failed to generate detail page: %v", err))
		return
	}
	if Verbose != nil && *Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated pages/app/%s/[id].vue", naming.PluralKebab))
	}

	if Verbose == nil || !*Verbose {
		cmd.PrintSuccess(fmt.Sprintf("Generated frontend module: %s", naming.Model))
	}
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

// getRelatedModelDisplayField reads the related model's type file and extracts the first string field
func getRelatedModelDisplayField(adminPath, relatedModelName string) string {
	// Create naming convention for the related model
	relatedNaming := utils.NewNamingConvention(relatedModelName)

	// Path to the related model's type file
	typePath := filepath.Join(adminPath, "modules", relatedNaming.PluralSnake, "types", relatedNaming.ModelSnake+".ts")

	// Check if file exists
	if _, err := os.Stat(typePath); os.IsNotExist(err) {
		// Fall back to common field names
		return "name"
	}

	// Read the file
	file, err := os.Open(typePath)
	if err != nil {
		return "name"
	}
	defer file.Close()

	// Parse the file to find the first string field
	scanner := bufio.NewScanner(file)
	inInterface := false
	fieldRegex := regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z0-9_]*)\??:\s*string`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if we're entering the main interface
		if strings.Contains(line, fmt.Sprintf("export interface %s {", relatedNaming.Model)) {
			inInterface = true
			continue
		}

		// Exit if we've left the interface
		if inInterface && strings.TrimSpace(line) == "}" {
			break
		}

		// Look for string fields
		if inInterface {
			// Skip id, timestamps, and comment lines
			if strings.Contains(line, "id:") ||
				strings.Contains(line, "created_at") ||
				strings.Contains(line, "updated_at") ||
				strings.Contains(line, "deleted_at") ||
				strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}

			// Check if this is a string field
			matches := fieldRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				return matches[1]
			}
		}
	}

	// Default fallback
	return "name"
}
