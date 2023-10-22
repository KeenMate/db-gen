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
		"uCamel": func(s string) string {
			return strcase.UpperCamelCase(s)
		},
		"lCamel": func(s string) string {
			return strcase.LowerCamelCase(s)
		},
		"snake": func(s string) string {
			return strcase.SnakeCase(s)
		},
	}
}
