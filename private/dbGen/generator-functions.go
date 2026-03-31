package dbGen

import (
	"fmt"
	"github.com/keenmate/db-gen/private/helpers"
	"strings"
	"text/template"
)

func getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"inc": func(n int) int {
			return n + 1
		},
		"pascalCased": func(s string) string {
			return helpers.ToPascalCase(s)
		},
		"camelCased": func(s string) string {
			return helpers.ToCamelCase(s)
		},
		"snakeCased": func(s string) string {
			return helpers.ToSnakeCase(s)
		},
		"normalizeStr": func(s string) string {
			return helpers.NormalizeStr(s)
		},
		"trimPrefix": func(p string, s string) string {
			return strings.TrimPrefix(s, p)
		},
		"renderValidationTemplate": renderValidationTemplateFunc,
		"getValidationTemplate":    getValidationTemplateFunc,
	}
}

// renderValidationTemplateFunc renders a validation rule template with parameter substitution
func renderValidationTemplateFunc(templateStr interface{}, param Property, rule ValidationRule) string {
	var tmplStr string

	// Handle both string and nested map templates
	switch t := templateStr.(type) {
	case string:
		tmplStr = t
	case map[string]interface{}:
		// For nested templates (like Manual with Check/ErrorBuilder)
		// This is handled by getValidationTemplateFunc
		return ""
	default:
		return ""
	}

	// Replace placeholders
	result := tmplStr
	result = strings.ReplaceAll(result, "{{.ParamName}}", param.PropertyName)
	result = strings.ReplaceAll(result, "{{.ParamDisplayName}}", param.PropertyName)

	// Replace parameters from rule
	for key, value := range rule.Parameters {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

// getValidationTemplateFunc gets a specific template part (for nested templates)
func getValidationTemplateFunc(templateMap interface{}, key string) string {
	if m, ok := templateMap.(map[string]interface{}); ok {
		if val, found := m[key]; found {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}
	return ""
}
