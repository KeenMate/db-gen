package helpers

import "testing"

func TestNormalizeStr(t *testing.T) {
	cases := map[string]string{
		"__number": "number",
		"_user_id": "user_id",
		"name":     "name",
		"":         "",
	}
	for in, want := range cases {
		if got := NormalizeStr(in); got != want {
			t.Errorf("NormalizeStr(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToPascalCase(t *testing.T) {
	cases := map[string]string{
		"hello_world":    "HelloWorld",
		"_user_id":       "UserId",
		"name":           "Name",
		"get_user_by_id": "GetUserById",
	}
	for in, want := range cases {
		if got := ToPascalCase(in); got != want {
			t.Errorf("ToPascalCase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToCamelCase(t *testing.T) {
	cases := map[string]string{
		"hello_world": "helloWorld",
		"_user_id":    "userId",
		"Name":        "name",
	}
	for in, want := range cases {
		if got := ToCamelCase(in); got != want {
			t.Errorf("ToCamelCase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	cases := map[string]string{
		"HelloWorld":  "hello_world",
		"_user_id":    "user_id",
		"GetUserById": "get_user_by_id",
	}
	for in, want := range cases {
		if got := ToSnakeCase(in); got != want {
			t.Errorf("ToSnakeCase(%q) = %q, want %q", in, got, want)
		}
	}
}
