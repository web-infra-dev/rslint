package gitignore_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
)

// For representative patterns, GitPolicy should make the same final
// file decision as Git, including excluded-parent and directory-node ordering.
func TestGitPolicy_MatchesGitCheckIgnore(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is unavailable")
	}
	tests := []struct {
		name       string
		gitignores map[string]string
		files      []string
	}{
		{
			name:       "rooted unrooted wildcards and classes",
			gitignores: map[string]string{".gitignore": "/root.ts\nloose.ts\nlogs/**/debug?.[jt]s\n"},
			files:      []string{"root.ts", "nested/root.ts", "nested/loose.ts", "logs/debug1.js", "logs/deep/debug2.ts", "logs/deep/debug22.ts"},
		},
		{
			name:       "case follows repository filesystem mode",
			gitignores: map[string]string{".gitignore": "CASE-SENSITIVE.ts\n"},
			files:      []string{"case-sensitive.ts"},
		},
		{
			name:       "escaping significant spaces and literal braces",
			gitignores: map[string]string{".gitignore": "\\!important.ts\n\\#generated.ts\ntrailing\\ \n leading.ts\nliteral\\*.ts\n{one,two}.ts\n"},
			files:      []string{"!important.ts", "#generated.ts", "trailing ", "trailing", " leading.ts", "leading.ts", "literal*.ts", "literal-one.ts", "{one,two}.ts", "one.ts"},
		},
		{
			name:       "directory parent propagation",
			gitignores: map[string]string{".gitignore": "cache/\nbuild/*\n!build/keep.ts\n"},
			files:      []string{"cache/deep/file.ts", "build/keep.ts", "build/drop.ts", "build/drop/child.ts"},
		},
		{
			name:       "excluded parent cannot be re-included by child negation",
			gitignores: map[string]string{".gitignore": "blocked/\n!blocked/keep.ts\n"},
			files:      []string{"blocked/keep.ts", "blocked/drop.ts"},
		},
		{
			name:       "trailing doublestar remains reversible",
			gitignores: map[string]string{".gitignore": "build/**\n!build/keep.ts\n"},
			files:      []string{"build/keep.ts", "build/drop.ts", "build/deep/drop.ts"},
		},
		{
			name:       "bare doublestar",
			gitignores: map[string]string{".gitignore": "**\n"},
			files:      []string{"source.ts", "nested/deep/source.ts"},
		},
		{
			name:       "ambiguous trailing doublestar",
			gitignores: map[string]string{".gitignore": "**/cache/**\n"},
			files:      []string{"cache.ts", "cache/cache", "cache/deep/cache", "other/cache/cache", "other/source.ts"},
		},
		{
			name: "directory negation does not reinclude nested ignored directory",
			gitignores: map[string]string{
				".gitignore":                "dist/\ndist-*/\n!dist-path/\n",
				"dist-path/dist/.gitignore": "!output.ts\n",
			},
			files: []string{"dist-path/src/source.ts", "dist-path/dist/output.ts"},
		},
		{
			name:       "directory negation preserves independent descendant matches",
			gitignores: map[string]string{".gitignore": "cache-*/\n*.generated.ts\ncache/\n!cache-path/\n"},
			files:      []string{"cache-path/src/source.ts", "cache-path/src/types.generated.ts", "cache-path/cache/output.ts"},
		},
		{
			name:       "later ancestor ignore overrides earlier descendant negation",
			gitignores: map[string]string{".gitignore": "!build/keep/\nbuild/\n"},
			files:      []string{"build/keep/source.ts"},
		},
		{
			name: "reachable nested source overrides ancestor",
			gitignores: map[string]string{
				".gitignore":     "*.generated.ts\n",
				"pkg/.gitignore": "!keep.generated.ts\n",
			},
			files: []string{"pkg/keep.generated.ts", "pkg/drop.generated.ts"},
		},
		{
			name: "nested source anchoring",
			gitignores: map[string]string{
				"pkg/.gitignore": "/only.ts\nsub/dir.ts\nloose.ts\n",
			},
			files: []string{
				"pkg/only.ts",
				"pkg/nested/only.ts",
				"pkg/sub/dir.ts",
				"pkg/nested/sub/dir.ts",
				"pkg/nested/loose.ts",
			},
		},
		{
			name: "nested source directory is literal glob text",
			gitignores: map[string]string{
				"!pkg[1]{x}/.gitignore": "ignored.ts\n",
			},
			files: []string{
				"!pkg[1]{x}/ignored.ts",
				"!pkg[1]{x}/source.ts",
				"!pkg1x/ignored.ts",
			},
		},
		{
			name: "separate source negation cannot restore excluded parent",
			gitignores: map[string]string{
				".gitignore":     "a-dir/\n",
				"foo/.gitignore": "!a-dir/keep.ts\n",
			},
			files: []string{"foo/a-dir/keep.ts", "foo/a-dir/drop.ts"},
		},
		{
			name: "ignored nested source is unreachable",
			gitignores: map[string]string{
				".gitignore":         "blocked/\n",
				"blocked/.gitignore": "!keep.ts\n",
			},
			files: []string{"blocked/keep.ts"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := setupGitignoreFixture(t, test.gitignores)
			command := exec.Command(git, "init", "--quiet", root)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("git init: %v: %s", err, output)
			}
			for _, relative := range test.files {
				// Windows cannot materialize `*` or a trailing space in a file
				// name. Platform-independent collector tests still pin those
				// conversions; this filesystem parity case skips only the
				// impossible host paths.
				if runtime.GOOS == "windows" &&
					(relative == "literal*.ts" || strings.HasSuffix(relative, " ")) {
					continue
				}
				path := filepath.Join(root, filepath.FromSlash(relative))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			base := config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}}
			full := policyMatcher(t, base, root, nil, config.ResolveDirectoryIdentities([]string{root}, osvfs.FS()))
			for _, relative := range test.files {
				if runtime.GOOS == "windows" &&
					(relative == "literal*.ts" || strings.HasSuffix(relative, " ")) {
					continue
				}
				path := filepath.Join(root, filepath.FromSlash(relative))
				command := exec.Command(git, "-C", root, "check-ignore", "--no-index", "--quiet", "--", filepath.FromSlash(relative))
				err := command.Run()
				want := err == nil
				var exitError *exec.ExitError
				if err != nil && (!errors.As(err, &exitError) || exitError.ExitCode() != 1) {
					t.Fatalf("git check-ignore %q: %v", relative, err)
				}
				if got := full.MatchFile(config.FreezeFileIdentity(path, osvfs.FS())).GloballyIgnored(); got != want {
					t.Errorf("%q: GitPolicy ignored=%v, git ignored=%v", relative, got, want)
				}

				explicit := policyMatcher(t, base, root, []config.PathIdentity{config.FreezeFileIdentity(path, osvfs.FS())}, nil)
				if got := explicit.MatchFile(config.FreezeFileIdentity(path, osvfs.FS())).GloballyIgnored(); got != want {
					t.Errorf("%q explicit: GitPolicy ignored=%v, git ignored=%v", relative, got, want)
				}
			}
		})
	}
}

func policyMatcher(t *testing.T, base config.RslintConfig, root string, files []config.PathIdentity, directories []config.DirectoryIdentity) config.TargetMatcher {
	t.Helper()
	policy := config.CollectGitPolicy(base, root, osvfs.FS(), files, directories)
	spaces := config.NewPathSpaceSnapshot(map[string]config.RslintConfig{root: base}, osvfs.FS())
	matcher, err := config.NewTargetMatcherWithPathSpaces(base, root, osvfs.FS(), spaces, policy)
	if err != nil {
		t.Fatal(err)
	}
	return matcher
}

func setupGitignoreFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return tspath.NormalizePath(root)
}
