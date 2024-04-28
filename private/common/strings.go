package common

import (
	"github.com/stoewer/go-strcase"
	"strings"
)

func NormalizeStr(s string) string {
	return strings.TrimLeft(s, "_")
}

func ToPascalCase(s string) string {
	return strcase.UpperCamelCase(NormalizeStr(s))
}

func ToCamelCase(s string) string {
	return strcase.LowerCamelCase(NormalizeStr(s))
}

func ToSnakeCase(s string) string {
	return strcase.SnakeCase(NormalizeStr(s))

}
