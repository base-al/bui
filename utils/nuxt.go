package utils

import (
	"strings"
)

// NuxtField extends Field with Nuxt/TypeScript specific information
type NuxtField struct {
	Field
	TypeScriptType string
	FormType       string
	FormRows       int
	ShowInTable    bool
	ShowInForm     bool
	ShowInDetail   bool
	IsFilterable   bool
	IsSortable     bool
	IsNullable     bool
	IsRequired     bool
	DefaultValue   string
	Label          string
	LabelLower     string
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
	case goType == "time.Time":
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

	// Check for specific field types
	if field.IsFile || field.IsImage || field.IsAttachment {
		return "file"
	}

	switch field.Type {
	case "bool":
		return "checkbox"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return "number"
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
	case "time.Time":
		return true
	default:
		return false
	}
}

// IsSortable determines if field can be used for sorting
func IsSortable(field Field) bool {
	// Can sort by: strings, numbers, dates
	switch field.Type {
	case "string", "int", "uint", "float32", "float64", "time.Time":
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
	if field.Type == "time.Time" {
		return "null"
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
