package dbGen

import (
	"github.com/stoewer/go-strcase"
	"text/template"
)

func getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"inc": func(n int) int {
			return n + 1
		},
		"pascalCased": func(s string) string {
			return strcase.UpperCamelCase(s)
		},
		"camelCased": func(s string) string {
			return strcase.LowerCamelCase(s)
		},
		"snakeCased": func(s string) string {
			return strcase.SnakeCase(s)
		},
	}
}
