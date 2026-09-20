package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
)

type lintTargetContractDiagnostic struct {
	RuleName string `json:"ruleName"`
	FilePath string `json:"filePath"`
}

func writeLintTargetContractFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for relative, content := range files {
		fileName := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(fileName), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(fileName), err)
		}
		if err := os.WriteFile(fileName, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", fileName, err)
		}
	}
}

func parseLintTargetContractDiagnostics(t *testing.T, stdout string) []lintTargetContractDiagnostic {
	t.Helper()
	var diagnostics []lintTargetContractDiagnostic
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var diagnostic lintTargetContractDiagnostic
		if err := json.Unmarshal([]byte(line), &diagnostic); err != nil {
			t.Fatalf("decode JSONLine diagnostic %q: %v", line, err)
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan JSONLine diagnostics: %v", err)
	}
	return diagnostics
}

func lintTargetContractPaths(
	t *testing.T,
	cwd string,
	diagnostics []lintTargetContractDiagnostic,
	matchRule func(string) bool,
) []string {
	t.Helper()
	paths := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if !matchRule(diagnostic.RuleName) {
			continue
		}
		fileName := filepath.FromSlash(diagnostic.FilePath)
		if !filepath.IsAbs(fileName) {
			fileName = filepath.Join(cwd, fileName)
		}
		paths = append(paths, tspath.NormalizePath(fileName))
	}
	sort.Strings(paths)
	return paths
}

func TestCLINoArgsUsesDefaultScriptExtensions(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"source.js":   "debugger;\n",
		"source.mjs":  "debugger;\n",
		"source.cjs":  "debugger;\n",
		"source.jsx":  "debugger;\n",
		"source.ts":   "debugger;\n",
		"source.mts":  "debugger;\n",
		"source.cts":  "debugger;\n",
		"source.tsx":  "debugger;\n",
		"source.css":  "debugger;\n",
		"source.json": "debugger;\n",
	}
	writeLintTargetContractFiles(t, dir, files)

	code, stdout, stderr := runLintCommandForTest(t, dir, lintArgs{
		ConfigCatalog: explicitConfigCatalogForTest(dir, rslintconfig.RslintConfig{{
			Rules: rslintconfig.Rules{"no-debugger": "error"},
		}}),
		Format:         "default",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 1 {
		t.Fatalf("default-extension lint exit code = %d, want 1; stdout=%q stderr=%q", code, stdout, stderr)
	}

	for _, extension := range []string{".js", ".mjs", ".cjs", ".jsx", ".ts", ".mts", ".cts", ".tsx"} {
		if !strings.Contains(stdout, "source"+extension+":") {
			t.Errorf("default scan missed source%s: stdout=%q stderr=%q", extension, stdout, stderr)
		}
	}
	for _, extension := range []string{".css", ".json"} {
		if strings.Contains(stdout, "source"+extension+":") {
			t.Errorf("default scan linted unsupported source%s: stdout=%q stderr=%q", extension, stdout, stderr)
		}
	}
	if !strings.Contains(stdout, "(8 files, 1 rule, 1 thread)") {
		t.Fatalf("default scan did not lint exactly 8 script files: stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestCLITypeCheckKeepsLintTargetsAndChecksWholeProject(t *testing.T) {
	dir := t.TempDir()
	writeLintTargetContractFiles(t, dir, map[string]string{
		"tsconfig.json": `{
			"compilerOptions": {"strict": true},
			"files": ["selected.ts", "unselected.ts"]
		}`,
		"selected.ts":   "debugger;\nexport const selected = true;\n",
		"unselected.ts": "debugger;\nexport const broken: number = 'value';\n",
	})
	selected := tspath.NormalizePath(filepath.Join(dir, "selected.ts"))
	unselected := tspath.NormalizePath(filepath.Join(dir, "unselected.ts"))

	run := func(t *testing.T, typeCheck bool) []lintTargetContractDiagnostic {
		t.Helper()
		code, stdout, stderr := runLintCommandForTest(t, dir, lintArgs{
			ConfigCatalog: explicitConfigCatalogForTest(dir, rslintconfig.RslintConfig{{
				Files: []string{"**/*.ts"},
				LanguageOptions: &rslintconfig.LanguageOptions{
					ParserOptions: &rslintconfig.ParserOptions{
						Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
					},
				},
				Rules: rslintconfig.Rules{"no-debugger": "error"},
			}}),
			AllowFiles:     []string{selected},
			Format:         "jsonline",
			NoColor:        true,
			SingleThreaded: true,
			TypeCheck:      typeCheck,
		})
		if code != 1 {
			t.Fatalf("typeCheck=%v exit code = %d, want 1; stdout=%q stderr=%q", typeCheck, code, stdout, stderr)
		}
		return parseLintTargetContractDiagnostics(t, stdout)
	}

	plainDiagnostics := run(t, false)
	typeCheckDiagnostics := run(t, true)
	wantLintTargets := []string{selected}
	for name, diagnostics := range map[string][]lintTargetContractDiagnostic{
		"plain":      plainDiagnostics,
		"type-check": typeCheckDiagnostics,
	} {
		got := lintTargetContractPaths(t, dir, diagnostics, func(ruleName string) bool {
			return ruleName == "no-debugger"
		})
		if !slices.Equal(got, wantLintTargets) {
			t.Fatalf("%s no-debugger targets = %v, want %v", name, got, wantLintTargets)
		}
	}

	plainTypeCheckPaths := lintTargetContractPaths(t, dir, plainDiagnostics, func(ruleName string) bool {
		return strings.HasPrefix(ruleName, "TypeScript(TS")
	})
	if len(plainTypeCheckPaths) != 0 {
		t.Fatalf("plain lint unexpectedly reported project diagnostics: diagnostics=%+v", plainDiagnostics)
	}
	typeCheckPaths := lintTargetContractPaths(t, dir, typeCheckDiagnostics, func(ruleName string) bool {
		return strings.HasPrefix(ruleName, "TypeScript(TS")
	})
	if !slices.Equal(typeCheckPaths, []string{unselected}) {
		t.Fatalf("--type-check project diagnostic targets = %v, want [%s]: diagnostics=%+v", typeCheckPaths, unselected, typeCheckDiagnostics)
	}
	if !slices.ContainsFunc(typeCheckDiagnostics, func(diagnostic lintTargetContractDiagnostic) bool {
		return diagnostic.RuleName == "TypeScript(TS2322)"
	}) {
		t.Fatalf("--type-check did not report TS2322 from the unselected tsconfig file: diagnostics=%+v", typeCheckDiagnostics)
	}
}

func TestCLIFixOnlyWritesSelectedTargets(t *testing.T) {
	dir := t.TempDir()
	writeLintTargetContractFiles(t, dir, map[string]string{
		"selected.js":   "var selected = 1;\nexport { selected };\n",
		"unselected.js": "var unselected = 1;\nexport { unselected };\n",
	})
	selected := tspath.NormalizePath(filepath.Join(dir, "selected.js"))
	unselected := filepath.Join(dir, "unselected.js")

	code, stdout, stderr := runLintCommandForTest(t, dir, lintArgs{
		ConfigCatalog: explicitConfigCatalogForTest(dir, rslintconfig.RslintConfig{{
			Files: []string{"**/*.js"},
			Rules: rslintconfig.Rules{"no-var": "error"},
		}}),
		AllowFiles:     []string{selected},
		Fix:            true,
		Format:         "jsonline",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 0 || stdout != "" {
		t.Fatalf("targeted fix failed: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	selectedContent, err := os.ReadFile(selected)
	if err != nil {
		t.Fatalf("read selected file: %v", err)
	}
	if strings.Contains(string(selectedContent), "var selected") {
		t.Fatalf("selected file was not fixed: %q", selectedContent)
	}
	unselectedContent, err := os.ReadFile(unselected)
	if err != nil {
		t.Fatalf("read unselected file: %v", err)
	}
	if string(unselectedContent) != "var unselected = 1;\nexport { unselected };\n" {
		t.Fatalf("--fix changed an unselected file: %q", unselectedContent)
	}
}

func TestCLIProjectRootsKeepTargetsAcrossInvocationForms(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_roots.txtar").Materialize(t, "roots"))
	src := tspath.ResolvePath(dir, "src")
	file := tspath.ResolvePath(src, "a.js")
	clean := tspath.ResolvePath(src, "clean.js")
	for _, project := range []string{"./tsconfig.json", "./typed.json", "./missing.json"} {
		for _, singleThreaded := range []bool{true, false} {
			for _, invocation := range []struct {
				name               string
				cwd                string
				files, directories []string
				wantFiles          int
			}{
				{name: "implicit", cwd: dir, wantFiles: 2},
				{name: "dot", cwd: dir, directories: []string{dir}, wantFiles: 2},
				{name: "subdirectory", cwd: dir, directories: []string{src}, wantFiles: 2},
				{name: "file", cwd: dir, files: []string{file}, wantFiles: 1},
				{name: "files", cwd: dir, files: []string{file, clean}, wantFiles: 2},
				{name: "overlap", cwd: dir, files: []string{file}, directories: []string{src}, wantFiles: 2},
				{name: "child-implicit", cwd: src, wantFiles: 2},
				{name: "child-file", cwd: src, files: []string{file}, wantFiles: 1},
			} {
				t.Run(fmt.Sprintf("%s/serial=%t/%s", project, singleThreaded, invocation.name), func(t *testing.T) {
					config := rslintconfig.RslintConfig{
						{Ignores: []string{"**/ignored.js"}},
						{Files: []string{"**/*.js"}, Plugins: []string{"@typescript-eslint"},
							LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{project}}},
							Rules:           rslintconfig.Rules{"no-debugger": "error", "@typescript-eslint/no-for-in-array": "error"}},
					}
					code, stdout, stderr := runLintCommandForTest(t, invocation.cwd, lintArgs{
						ConfigCatalog: explicitConfigCatalogForTest(dir, config),
						AllowFiles:    invocation.files, AllowDirs: invocation.directories,
						Format: "default", NoColor: true, SingleThreaded: singleThreaded,
					})
					if project == "./missing.json" {
						if code != 1 || !strings.Contains(stderr, "missing.json") || strings.Contains(stdout, "no-debugger") {
							t.Fatalf("config failure became a gap: exit=%d stdout=%q stderr=%q", code, stdout, stderr)
						}
						return
					}
					wantTyped := 0
					if project == "./typed.json" {
						wantTyped = 1
					}
					countText := fmt.Sprintf("(%d files,", invocation.wantFiles)
					if invocation.wantFiles == 1 {
						countText = "(1 file,"
					}
					if code != 1 || stderr != "" || !strings.Contains(stdout, countText) ||
						strings.Count(stdout, "no-debugger") != 1 || strings.Count(stdout, "no-for-in-array") != wantTyped ||
						strings.Contains(stdout, "ignored.js:") {
						t.Fatalf("changed targets or capabilities: exit=%d stdout=%q stderr=%q", code, stdout, stderr)
					}
				})
			}
		}
	}
}

func TestCLIProjectExtensionGateKeepsMixedTargetsAndEligibleFallback(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_roots.txtar").Materialize(t, "aliases"))
	if err := os.Symlink("src/a.js", tspath.ResolvePath(dir, "alias.ts")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	main := tspath.ResolvePath(dir, "main.ts")
	js := tspath.ResolvePath(dir, "src/a.js")
	extra := tspath.ResolvePath(dir, "src/extra.js")
	for _, laterProject := range []bool{false, true} {
		projects := rslintconfig.ProjectPaths{"./tsconfig.json"}
		if laterProject {
			projects = append(projects, "./typed.json")
		}
		for _, invocation := range []struct {
			name               string
			files, directories []string
			want               []string
		}{
			{name: "single-js", files: []string{js}, want: []string{js}},
			{name: "mixed", files: []string{main, js, extra}, want: []string{main, js, extra}},
			{name: "directory", directories: []string{dir}, want: []string{main, js, extra}},
			{name: "implicit", want: []string{main, js, extra}},
		} {
			t.Run(fmt.Sprintf("later-project=%t/%s", laterProject, invocation.name), func(t *testing.T) {
				config := rslintconfig.RslintConfig{
					{Ignores: []string{"alias.ts", "typed-main.ts"}},
					{Files: []string{"**/*.js", "**/*.ts"}, Plugins: []string{"@typescript-eslint"},
						LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: projects}},
						Rules:           rslintconfig.Rules{"no-debugger": "error", "@typescript-eslint/no-for-in-array": "error"}},
				}
				code, stdout, stderr := runLintCommandForTest(t, dir, lintArgs{
					ConfigCatalog: explicitConfigCatalogForTest(dir, config),
					AllowFiles:    invocation.files, AllowDirs: invocation.directories,
					Format: "jsonline", NoColor: true,
				})
				if code != 1 || stderr != "" {
					t.Fatalf("lint failed: exit=%d stdout=%q stderr=%q", code, stdout, stderr)
				}
				diagnostics := parseLintTargetContractDiagnostics(t, stdout)
				// Every source has a debugger sentinel, including ignored helpers,
				// so these paths expose dropped targets and leaked Program sources.
				gotTargets := lintTargetContractPaths(t, dir, diagnostics, func(ruleName string) bool {
					return ruleName == "no-debugger"
				})
				if !slices.Equal(gotTargets, invocation.want) {
					t.Fatalf("lint targets = %v, want %v", gotTargets, invocation.want)
				}
				var wantTyped []string
				if slices.Contains(invocation.want, main) {
					wantTyped = append(wantTyped, main)
				}
				if laterProject {
					wantTyped = append(wantTyped, js)
				}
				gotTyped := lintTargetContractPaths(t, dir, diagnostics, func(ruleName string) bool {
					return ruleName == "@typescript-eslint/no-for-in-array"
				})
				if !slices.Equal(gotTyped, wantTyped) || len(diagnostics) != len(invocation.want)+len(wantTyped) {
					t.Fatalf("typed targets = %v, want %v; diagnostics=%+v", gotTyped, wantTyped, diagnostics)
				}
			})
		}
	}
}
