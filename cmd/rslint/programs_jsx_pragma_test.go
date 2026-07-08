package main

import (
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestCreateProgramsForConfig_NoTsconfig_JsxPragma is a regression test for
// https://github.com/web-infra-dev/rslint/issues/1230: a project with NO
// tsconfig.json (so createProgramsForConfig takes the directory-scan
// fallback branch) that sets languageOptions.parserOptions.jsxPragma to a
// non-default value (e.g. Preact's "h") must have that value reach the
// synthesized Program's CompilerOptions.JsxFactory, so the Go-native
// `@typescript-eslint/no-unused-vars` rule's implicit-JSX-usage check
// (markJsxFactoryUsed) marks the configured pragma's import as used
// instead of hard-coding "React".
func TestCreateProgramsForConfig_NoTsconfig_JsxPragma(t *testing.T) {
	rslintconfig.RegisterAllRules()

	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"src/index.jsx": `import { h } from 'preact';

export function App() {
  return <div>Hello Preact</div>;
}
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	jsxFile := tspath.NormalizePath(filepath.Join(dir, "src/index.jsx"))

	runNoUnusedVars := func(t *testing.T, jsxPragma *string) []rule.RuleDiagnostic {
		t.Helper()
		cfg := rslintconfig.RslintConfig{
			{
				Files:   []string{"**/*.jsx"},
				Plugins: []string{"@typescript-eslint"},
				Rules:   rslintconfig.Rules{"@typescript-eslint/no-unused-vars": "error"},
				LanguageOptions: &rslintconfig.LanguageOptions{
					ParserOptions: &rslintconfig.ParserOptions{JsxPragma: jsxPragma},
				},
			},
		}

		programs, exitCode := createProgramsForConfig(dir, cfg, true, fs, nil, utils.NewParseCache())
		if exitCode != 0 {
			t.Fatalf("createProgramsForConfig exit code = %d", exitCode)
		}
		if len(programs) != 1 {
			t.Fatalf("expected one synthesized no-tsconfig program, got %d", len(programs))
		}

		wantFactory := ""
		if jsxPragma != nil {
			wantFactory = *jsxPragma
		}
		if got := programs[0].Options().JsxFactory; got != wantFactory {
			t.Fatalf("expected Program CompilerOptions.JsxFactory = %q, got %q", wantFactory, got)
		}

		rules, _ := rslintconfig.GlobalRuleRegistry.GetEnabledRules(cfg, jsxFile, dir, false)

		var diags []rule.RuleDiagnostic
		_, err := linter.RunLinter(linter.RunLinterOptions{
			Programs: programs,
			Scope:    linter.FileScope{Files: []string{jsxFile}},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return rules
			},
			OnDiagnostic: func(d rule.RuleDiagnostic) {
				diags = append(diags, d)
			},
		})
		if err != nil {
			t.Fatalf("RunLinter: %v", err)
		}
		return diags
	}

	pragma := "h"
	if diags := runNoUnusedVars(t, &pragma); len(diags) != 0 {
		t.Fatalf("jsxPragma:'h' — expected 'h' to be marked used via JSX, got diagnostics: %+v", diags)
	}

	// Negative control: without jsxPragma configured, "React" is the
	// implicit-usage default, so the `h` import genuinely IS unused — pins
	// that the above assertion is exercising the pragma, not just always
	// passing.
	if diags := runNoUnusedVars(t, nil); len(diags) == 0 {
		t.Fatal("expected 'h' to be reported as unused when jsxPragma is not configured (default pragma is 'React')")
	}
}
