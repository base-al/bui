package utils

import (
	"strings"
)

// NuxtField extends Field with Nuxt/TypeScript specific information
type NuxtField struct {
	Field
	TypeScriptType       string
	FormType             string
	FormRows             int
	ShowInTable          bool
	ShowInForm           bool
	ShowInDetail         bool
	IsFilterable         bool
	IsSortable           bool
	IsNullable           bool
	IsRequired           bool
	DefaultValue         string
	Label                string
	LabelLower           string
	RelationLabel        string // For belongs_to: Clean label without "_id" suffix (e.g., "Client" instead of "Client Id")
	RelationDisplayField string // For belongs_to: Display field of the related model (e.g., "name" for Client)
	RelationModelPlural  string // Plural form of related model for API calls
	RelationModelKebab   string // Kebab case for API endpoints
	RelationObjectName   string // For belongs_to: JSONName with _id suffix removed (e.g., "client" from "client_id")
	RelationModelSingular string // Singular form of related model (e.g., "comment" for comments hasMany)
	RelationModelSnake   string // Snake case singular (e.g., "comment" for Comment)
}

// ConvertToNuxtField converts a Go Field to a NuxtField with TypeScript types
func ConvertToNuxtField(field Field) NuxtField {
	// Clean JSONName for label (remove ,omitempty suffix)
	cleanJSONName := strings.TrimSuffix(field.JSONName, ",omitempty")

	nf := NuxtField{
		Field:          field,
		TypeScriptType: GetTypeScriptType(field.Type),
		FormType:       GetFormType(field),
		FormRows:       GetFormRows(field),
		ShowInTable:    ShouldShowInTable(field),
		ShowInForm:     ShouldShowInForm(field),
		ShowInDetail:   true,
		IsFilterable:   IsFilterable(field),
		IsSortable:     IsSortable(field),
		IsNullable:     IsNullableField(field),
		IsRequired:     IsRequiredField(field),
		DefaultValue:   GetDefaultValue(field),
		Label:          ToCapitalCase(cleanJSONName),
		LabelLower:     strings.ToLower(ToCapitalCase(cleanJSONName)),
	}

	// Handle relation-specific fields
	if field.IsRelation && field.RelatedModel != "" {
		// Extract model name from package.Model format (e.g., "users.User" -> "User")
		relatedModelName := field.RelatedModel
		if strings.Contains(relatedModelName, ".") {
			parts := strings.Split(relatedModelName, ".")
			relatedModelName = parts[len(parts)-1]
		}

		switch field.Relationship {
		case "belongs_to":
			nf.FormType = "select"
			nf.RelationModelPlural = ToPlural(relatedModelName)
			nf.RelationModelKebab = ToKebabCase(ToPlural(relatedModelName))
			nf.RelationObjectName = strings.TrimSuffix(field.JSONName, "_id")
			nf.RelationLabel = ToCapitalCase(nf.RelationObjectName)
			nf.RelationModelSingular = strings.ToLower(relatedModelName)
			nf.RelationModelSnake = ToSnakeCase(relatedModelName)
			nf.ShowInForm = false   // Don't show in regular form section, will be handled by relation section
			nf.ShowInTable = false  // Don't show FK in table, will show relation object instead
			nf.ShowInDetail = false // Don't show FK in detail, will show relation object instead
			nf.IsFilterable = true

		case "has_many":
			// hasMany: show count in table with link
			nf.RelationModelPlural = ToPlural(relatedModelName)
			nf.RelationModelKebab = ToKebabCase(ToPlural(relatedModelName))
			nf.RelationModelSingular = strings.ToLower(relatedModelName)
			nf.RelationModelSnake = ToSnakeCase(relatedModelName)
			nf.RelationLabel = ToCapitalCase(field.JSONName)
			nf.ShowInForm = false   // hasMany not shown in form (managed separately)
			nf.ShowInTable = true   // Show count in table
			nf.ShowInDetail = true  // Show list in detail view
			nf.IsFilterable = false

		case "many_to_many":
			// manyToMany: show chips in table
			nf.RelationModelPlural = ToPlural(relatedModelName)
			nf.RelationModelKebab = ToKebabCase(ToPlural(relatedModelName))
			nf.RelationModelSingular = strings.ToLower(relatedModelName)
			nf.RelationModelSnake = ToSnakeCase(relatedModelName)
			nf.RelationLabel = ToCapitalCase(field.JSONName)
			nf.ShowInForm = false   // manyToMany shown in special multi-select
			nf.ShowInTable = true   // Show chips in table
			nf.ShowInDetail = true  // Show chips in detail view
			nf.IsFilterable = false
		}
	}

	return nf
}

// GetTypeScriptType converts Go type to TypeScript type
func GetTypeScriptType(goType string) string {
	switch {
	case strings.HasPrefix(goType, "[]"):
		// Array type
		innerType := strings.TrimPrefix(goType, "[]")
		innerType = strings.TrimPrefix(innerType, "*")
		return GetTypeScriptType(innerType) + "[]"
	case strings.HasPrefix(goType, "*"):
		// Pointer type - remove pointer
		return GetTypeScriptType(strings.TrimPrefix(goType, "*"))
	case goType == "string":
		return "string"
	case goType == "int", goType == "int8", goType == "int16", goType == "int32", goType == "int64":
		return "number"
	case goType == "uint", goType == "uint8", goType == "uint16", goType == "uint32", goType == "uint64":
		return "number"
	case goType == "float32", goType == "float64":
		return "number"
	case goType == "bool":
		return "boolean"
	case goType == "time.Time", goType == "types.DateTime":
		return "string"
	case goType == "datatypes.JSON", goType == "json.RawMessage":
		return "Record<string, any>"
	case strings.Contains(goType, "storage.Attachment"):
		return "string" // URL to the file
	default:
		// Custom types or enums - assume string
		return "any"
	}
}

// GetFormType determines the form input type
func GetFormType(field Field) string {
	fieldName := strings.ToLower(field.JSONName)

	// Check for explicit select/radio/checkbox fields (takes priority)
	if field.IsSelect && len(field.Options) > 0 {
		// Return the specific select type: "select", "radio", or "checkbox"
		return field.SelectType
	}

	// Check for specific field types
	if field.IsFile || field.IsImage || field.IsAttachment {
		return "file"
	}

	switch field.Type {
	case "bool":
		return "checkbox"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return "number"
	case "types.DateTime":
		// types.DateTime is used for date/timestamp fields
		if strings.Contains(fieldName, "time") {
			return "datetime"
		}
		return "date"
	case "time.Time":
		if strings.Contains(fieldName, "date") && !strings.Contains(fieldName, "time") {
			return "date"
		}
		return "datetime"
	case "text", "string":
		// Check field name for hints
		if strings.Contains(fieldName, "content") || strings.Contains(fieldName, "description") || strings.Contains(fieldName, "bio") {
			return "textarea"
		}
		if strings.Contains(fieldName, "email") {
			return "email"
		}
		if strings.Contains(fieldName, "url") || strings.Contains(fieldName, "link") {
			return "url"
		}
		if strings.Contains(fieldName, "password") {
			return "password"
		}
		// Check for enum-like fields
		if strings.Contains(fieldName, "status") || strings.Contains(fieldName, "category") || strings.Contains(fieldName, "type") {
			return "select"
		}
		return "text"
	default:
		return "text"
	}
}

// GetFormRows determines number of rows for textarea
func GetFormRows(field Field) int {
	fieldName := strings.ToLower(field.JSONName)
	if strings.Contains(fieldName, "content") {
		return 6
	}
	if strings.Contains(fieldName, "description") || strings.Contains(fieldName, "excerpt") {
		return 3
	}
	return 4
}

// ShouldShowInTable determines if field should appear in table
func ShouldShowInTable(field Field) bool {
	fieldName := strings.ToLower(field.JSONName)

	// Never show these in table
	if fieldName == "id" || fieldName == "created_at" || fieldName == "updated_at" || fieldName == "deleted_at" {
		return false
	}

	// Don't show belongs_to FK fields in table (show the relation object instead)
	if field.IsRelation && field.Relationship == "belongs_to" {
		return false
	}

	// Never show large text fields or JSON in table
	if field.Type == "text" || field.Type == "datatypes.JSON" || field.Type == "json.RawMessage" {
		return false
	}

	if strings.Contains(fieldName, "content") || strings.Contains(fieldName, "description") {
		return false
	}

	if field.IsFile || field.IsImage || field.IsAttachment {
		return false
	}

	return true
}

// ShouldShowInForm determines if field should appear in form
func ShouldShowInForm(field Field) bool {
	fieldName := strings.ToLower(field.JSONName)

	// Never show these in form
	if fieldName == "id" || fieldName == "created_at" || fieldName == "updated_at" || fieldName == "deleted_at" {
		return false
	}

	// Skip relation fields (they're handled separately)
	if field.IsRelation {
		return false
	}

	return true
}

// IsFilterable determines if field can be used as a filter
func IsFilterable(field Field) bool {
	// Can filter by: strings, enums, booleans, numbers, foreign keys
	switch field.Type {
	case "string", "bool", "int", "uint", "float32", "float64":
		return true
	case "time.Time", "types.DateTime":
		return true
	default:
		return false
	}
}

// IsSortable determines if field can be used for sorting
func IsSortable(field Field) bool {
	// Can sort by: strings, numbers, dates
	switch field.Type {
	case "string", "int", "uint", "float32", "float64", "time.Time", "types.DateTime":
		return true
	default:
		return false
	}
}

// GetDefaultValue returns the TypeScript default value for a field
func GetDefaultValue(field Field) string {
	if field.Type == "bool" {
		return "false"
	}
	if strings.Contains(field.Type, "int") || strings.Contains(field.Type, "float") {
		return "0"
	}
	if field.Type == "string" {
		return "''"
	}
	if strings.HasPrefix(field.Type, "[]") {
		return "[]"
	}
	if field.Type == "datatypes.JSON" || field.Type == "json.RawMessage" {
		return "{}"
	}
	if field.Type == "time.Time" || field.Type == "types.DateTime" {
		// Use empty string for date/datetime fields to avoid TypeScript errors
		return "''"
	}
	return "undefined"
}

// IsNullableField determines if a field is nullable
func IsNullableField(field Field) bool {
	// Pointer types are nullable
	if strings.HasPrefix(field.Type, "*") {
		return true
	}
	// time.Time is nullable
	if field.Type == "time.Time" {
		return true
	}
	// JSON fields are nullable
	if field.Type == "datatypes.JSON" || field.Type == "json.RawMessage" {
		return true
	}
	return false
}

// IsRequiredField determines if a field is required in forms
func IsRequiredField(field Field) bool {
	fieldName := strings.ToLower(field.JSONName)

	// These fields are typically not required
	if fieldName == "id" || fieldName == "created_at" || fieldName == "updated_at" || fieldName == "deleted_at" {
		return false
	}

	// Nullable fields are not required
	if IsNullableField(field) {
		return false
	}

	// Fields with "optional" in name are not required
	if strings.Contains(fieldName, "optional") {
		return false
	}

	// By default, non-nullable fields are required
	return true
}
