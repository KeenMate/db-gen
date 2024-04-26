package dbGen

import (
	"github.com/keenmate/db-gen/common"
	"text/template"
)

func getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"inc": func(n int) int {
			return n + 1
		},
		"pascalCased": func(s string) string {
			return common.ToPascalCase(s)
		},
		"camelCased": func(s string) string {
			return common.ToCamelCase(s)
		},
		"snakeCased": func(s string) string {
			return common.ToSnakeCase(s)
		},
	}
}
