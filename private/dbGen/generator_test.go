package dbGen

import (
	"os"
	"path/filepath"
	"text/template"

	"testing"
)

func TestChangeCase(t *testing.T) {
	cases := []struct {
		in, kase, want string
	}{
		{"hello_world", "pascalcase", "HelloWorld"},
		{"hello_world", "camelcase", "helloWorld"},
		{"HelloWorld", "snakecase", "hello_world"},
	}
	for _, c := range cases {
		if got := changeCase(c.in, c.kase); got != c.want {
			t.Errorf("changeCase(%q,%q) = %q, want %q", c.in, c.kase, got, c.want)
		}
	}
}

func TestFileMd5SumAndHashes(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("same"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("same"), 0o600); err != nil {
		t.Fatal(err)
	}

	ha, err := fileMd5Sum(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, _ := fileMd5Sum(b)
	if ha == "" || ha != hb {
		t.Errorf("identical files should hash equal: %q vs %q", ha, hb)
	}

	hashes, err := generateFileHashes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(*hashes) != 2 {
		t.Errorf("expected 2 hashes, got %d", len(*hashes))
	}
	if (*hashes)[filepath.Clean(a)] != ha {
		t.Errorf("hash map missing/incorrect entry for a.txt")
	}
}

func TestGenerateFile_ChangeDetection(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "out.txt")

	tmpl, err := template.New("t").Funcs(getTemplateFunctions()).Parse("Hello {{.Name}}")
	if err != nil {
		t.Fatal(err)
	}
	data := struct{ Name string }{Name: "World"}

	hashes := map[string]string{}
	generated := map[string]bool{}

	// first generation -> file is new -> changed
	changed, err := generateFile(data, tmpl, fp, &hashes, &generated)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("first generation should report changed")
	}
	content, _ := os.ReadFile(fp)
	if string(content) != "Hello World" {
		t.Errorf("rendered content = %q", content)
	}
	if !generated[filepath.Clean(fp)] {
		t.Error("file should be tracked as generated")
	}

	// second generation with the prior hash recorded -> unchanged
	hashes[filepath.Clean(fp)], _ = fileMd5Sum(fp)
	changed, err = generateFile(data, tmpl, fp, &hashes, &generated)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("identical regeneration should report unchanged")
	}
}

func TestRemoveOrphanedFiles(t *testing.T) {
	dir := t.TempDir()
	orphan := filepath.Join(dir, "orphan.txt")
	keep := filepath.Join(dir, "keep.txt")
	if err := os.WriteFile(orphan, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keep, []byte("y"), 0o600); err != nil {
		t.Fatal(err)
	}

	hashes := map[string]string{
		filepath.Clean(orphan): "x",
		filepath.Clean(keep):   "y",
	}
	generated := map[string]bool{filepath.Clean(keep): true}

	if err := removeOrphanedFiles(&hashes, &generated, &Config{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Error("orphan file should have been removed")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Error("kept file should still exist")
	}
}

func TestEnsureOutputFolder_Clear(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(out, "stale.txt")
	if err := os.WriteFile(stale, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	config := &Config{OutputFolder: out, ClearOutputFolder: true}
	if err := ensureOutputFolder(config); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("ClearOutputFolder should have removed stale file")
	}
	if _, err := os.Stat(out); err != nil {
		t.Error("output folder should still exist")
	}
}
