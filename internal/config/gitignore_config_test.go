package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

func TestConfigWithGitignore_DefaultIgnoresRemainBaseline(t *testing.T) {
	root := t.TempDir()
	tests := []struct {
		name              string
		relativeTarget    string
		gitignoreNegation string
		authoredNegation  string
	}{
		{name: "node_modules", relativeTarget: "node_modules/pkg/index.ts", gitignoreNegation: "!node_modules/", authoredNegation: "!**/node_modules/"},
		{name: ".git", relativeTarget: ".git/internal.ts", gitignoreNegation: "!.git/", authoredNegation: "!.git/"},
	}
	for _, test := range tests {
		target := filepath.Join(root, filepath.FromSlash(test.relativeTarget))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("export {};\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("!node_modules/\n!.git/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root = tspath.NormalizePath(root)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target := tspath.NormalizePath(filepath.Join(root, filepath.FromSlash(test.relativeTarget)))
			effective := ConfigWithGitignore(RslintConfig{{}}, root, osvfs.FS(), []string{target})
			if !effective.IsFileIgnored(target, root) {
				t.Fatalf(".gitignore negation %q must not expand the default discovery baseline", test.gitignoreNegation)
			}

			effective = ConfigWithGitignore(RslintConfig{
				{Ignores: []string{test.authoredNegation}},
			}, root, osvfs.FS(), []string{target})
			if effective.IsFileIgnored(target, root) {
				t.Fatalf("authored config negation %q must reopen the default entry", test.authoredNegation)
			}
		})
	}
}
