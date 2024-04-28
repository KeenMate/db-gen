package dbGen

import (
	"github.com/keenmate/db-gen/private/helpers"
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
	}
}
