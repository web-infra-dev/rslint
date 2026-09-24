package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSelectPackages(t *testing.T) {
	root := filepath.Join(t.TempDir(), "module with spaces")
	files := map[string]string{
		"internal/core.test/other.go":            "package independent\n",
		"internal/root.go":                       "package internal\n",
		"cmd/root.go":                            "package cmd\n",
		"go.mod":                                 "module example.com/selection\n\ngo 1.27.0\n",
		"cmd/app/main.go":                        "package main\nimport _ \"example.com/selection/internal/core\"\nfunc main() {}\n",
		"internal/core/core.go":                  "package core\nimport \"embed\"\n//go:embed schema.json docs/guide.md child.test/data.json\nvar Data embed.FS\n",
		"internal/core/schema.json":              "{}",
		"internal/core/docs/guide.md":            "embedded documentation",
		"internal/core/README.md":                "ordinary documentation",
		"internal/core/testdata/example.txtar":   "-- index.ts --\nexport {};\n",
		"internal/core/testdata/project/main.go": "package fixture\n",
		"internal/core/testdata/case/README.md":  "fixture documentation",
		"internal/core/core_test.go":             "package core\nimport (_ \"embed\"; \"testing\")\n//go:embed test.txt\nvar data string\nfunc TestCore(t *testing.T) {}\n",
		"internal/core/external_test.go":         "package core_test\nimport (_ \"embed\"; \"testing\"; _ \"example.com/selection/internal/core\")\n//go:embed external.txt\nvar data string\nfunc TestExternal(t *testing.T) {}\n",
		"internal/core/test.txt":                 "internal test embed",
		"internal/core/external.txt":             "external test embed",
		"internal/core/child.test/go.go":         "package child\nimport _ \"embed\"\n//go:embed data.json\nvar Data string\n",
		"internal/core/child.test/data.json":     "{}",
		"internal/child_user/child.go":           "package child_user\nimport _ \"example.com/selection/internal/core/child.test\"\n",
		"internal/consumer/consumer.go":          "package consumer\nimport _ \"example.com/selection/internal/core\"\n",
		"internal/test_user/user.go":             "package test_user\n",
		"internal/test_user/user_test.go":        "package test_user_test\nimport (\"testing\"; _ \"example.com/selection/internal/consumer\")\nfunc TestUser(t *testing.T) {}\n",
		"internal/unrelated/other.go":            "package unrelated\n",
	}
	for name, content := range files {
		name = filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(root)
	t.Setenv("GOWORK", "off")
	t.Setenv("GOTOOLCHAIN", "auto")
	cmd := exec.Command("go", "list", "-test", "-json", "./cmd/...", "./internal/...")
	metadata, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	core := []string{"cmd/app", "internal/consumer", "internal/core", "internal/test_user"}
	child := []string{"internal/child_user", "internal/core/child.test"}
	cases := []struct {
		name    string
		files   []string
		want    []string
		wantErr bool
	}{
		{name: "internal root package", files: []string{"internal/root.go"}, want: []string{"internal"}},
		{name: "cmd root package", files: []string{"cmd/root.go"}, want: []string{"cmd"}},
		{name: "source", files: []string{"internal/core/core.go"}, want: core},
		{name: "deleted source", files: []string{"internal/core/deleted.go"}, want: core},
		{name: "schema", files: []string{"internal/core/schema.json"}, want: core},
		{name: "internal test embed", files: []string{"internal/core/test.txt"}, want: core},
		{name: "external test embed", files: []string{"internal/core/external.txt"}, want: core},
		{name: "embedded documentation", files: []string{"internal/core/docs/guide.md"}, want: core},
		{name: "txtar", files: []string{"internal/core/testdata/example.txtar"}, want: core},
		{name: "Go fixture", files: []string{"internal/core/testdata/project/main.go"}, want: core},
		{name: "fixture documentation", files: []string{"internal/core/testdata/case/README.md"}, want: core},
		{name: "deleted fixture documentation", files: []string{"internal/core/testdata/README.md"}, want: core},
		{name: "multiple embed owners", files: []string{"internal/core/child.test/data.json"}, want: slices.Concat(core, child)},
		{name: "real package ending in test", files: []string{"internal/core/child.test/go.go"}, want: child},
		{name: "real package shares test binary name", files: []string{"internal/core.test/other.go"}, want: []string{"internal/core.test"}},
		{name: "ordinary documentation", files: []string{"internal/core/README.md"}},
		{name: "documentation with code", files: []string{"internal/core/README.md", "internal/unrelated/other.go"}, want: []string{"internal/unrelated"}},
		{name: "duplicate changes", files: []string{"internal/core/core.go", "internal/core/core.go", ""}, want: core},
		{name: "unknown input", files: []string{"internal/core/config.json"}, wantErr: true},
		{name: "deleted embed", files: []string{"internal/core/deleted.json"}, wantErr: true},
		{name: "deleted unowned documentation", files: []string{"internal/core/deleted.md"}, wantErr: true},
		{name: "deleted package", files: []string{"internal/deleted/gone.go"}, wantErr: true},
		{name: "fixture owner has no tests", files: []string{"internal/unrelated/testdata/file.txtar"}, wantErr: true},
		{name: "fixture prefix collision", files: []string{"internal/core/testdata-other/file.txtar"}, wantErr: true},
		{name: "mixed mapped and unknown", files: []string{"internal/core/schema.json", "internal/core/unknown.txt"}, wantErr: true},
		{name: "outside suite", files: []string{"tools/other.go"}, wantErr: true},
		{name: "path traversal", files: []string{"internal/../other.go"}, wantErr: true},
		{name: "empty changes", files: []string{""}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectPackages(tc.files, bytes.NewReader(metadata))
			if tc.wantErr {
				if err == nil || len(got) != 0 {
					t.Fatalf("got %v, %v; want error without partial selection", got, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := make([]string, len(tc.want))
			for i, pkg := range tc.want {
				want[i] = "example.com/selection/" + pkg
			}
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
		})
	}
	for _, invalid := range []string{"", "{}", "[]", string(metadata) + "{", strings.ReplaceAll(string(metadata), `"Module":`, `"MissingModule":`)} {
		if got, err := selectPackages([]string{"internal/core/README.md"}, strings.NewReader(invalid)); err == nil || len(got) != 0 {
			t.Fatalf("invalid metadata returned %v, %v; want error without selection", got, err)
		}
	}
	// The command must report input/output failures, rather than turn them into
	// an empty successful selection that would skip the tests.
	changedPath, metadataPath := filepath.Join(root, "changed.txt"), filepath.Join(root, "metadata.json")
	if err := os.WriteFile(changedPath, []byte("internal/core/core.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metadataPath, metadata, 0o644); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run([]string{changedPath, metadataPath}, &output); err != nil || len(strings.Fields(output.String())) != len(core) {
		t.Fatalf("command output: %q, error: %v", output.String(), err)
	}
	if err := run([]string{changedPath, metadataPath}, failingWriter{}); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("output failure was not propagated: %v", err)
	}
	for _, args := range [][]string{nil, {"missing", metadataPath}, {changedPath, "missing"}} {
		output.Reset()
		if err := run(args, &output); err == nil || output.Len() != 0 {
			t.Fatalf("invalid arguments %v returned %q, %v", args, output.String(), err)
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, os.ErrPermission
}
