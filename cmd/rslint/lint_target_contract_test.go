package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
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
		"rslint.jsonc": `[{
			"rules": {"no-debugger": "error"}
		}]`,
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
		Config:         "rslint.jsonc",
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
	if !strings.Contains(stdout, "linted 8 files") {
		t.Fatalf("default scan did not lint exactly 8 script files: stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestCLITypeCheckKeepsLintTargetsAndChecksWholeProject(t *testing.T) {
	dir := t.TempDir()
	writeLintTargetContractFiles(t, dir, map[string]string{
		"rslint.jsonc": `[{
			"files": ["**/*.ts"],
			"languageOptions": {"parserOptions": {"project": ["./tsconfig.json"]}},
			"rules": {"no-debugger": "error"}
		}]`,
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
			Config:         "rslint.jsonc",
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
		"rslint.jsonc": `[{
			"files": ["**/*.js"],
			"rules": {"no-var": "error"}
		}]`,
		"selected.js":   "var selected = 1;\nexport { selected };\n",
		"unselected.js": "var unselected = 1;\nexport { unselected };\n",
	})
	selected := tspath.NormalizePath(filepath.Join(dir, "selected.js"))
	unselected := filepath.Join(dir, "unselected.js")

	code, stdout, stderr := runLintCommandForTest(t, dir, lintArgs{
		Config:         "rslint.jsonc",
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
