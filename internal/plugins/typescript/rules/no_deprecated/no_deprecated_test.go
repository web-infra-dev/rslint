package no_deprecated

import (
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func sourceFileFromCode(t *testing.T, code string) *ast.SourceFile {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	fileName := "file.ts"
	fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("failed to resolve source file for %s", fileName)
	}
	return sourceFile
}

func TestNameImportedFromPackage(t *testing.T) {
	sourceFile := sourceFileFromCode(t, `
import { oldValue } from 'pkg';
const another = 1;
`)
	if !nameImportedFromPackage(sourceFile, "oldValue", "pkg") {
		t.Fatalf("expected static import to match package allow entry")
	}
	if nameImportedFromPackage(sourceFile, "oldValue", "other-pkg") {
		t.Fatalf("did not expect import to match other package")
	}
}

func TestNameImportedFromPackageDynamicImport(t *testing.T) {
	sourceFile := sourceFileFromCode(t, `
async function run() {
  const { oldValue } = await import('pkg');
  oldValue();
}
`)
	if !nameImportedFromPackage(sourceFile, "oldValue", "pkg") {
		t.Fatalf("expected dynamic import binding to match package allow entry")
	}
}

func TestNoDeprecatedRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDeprecatedRule, []rule_tester.ValidTestCase{
		{Code: `const value = 1; value;`},
		{
			Code: `
/** @deprecated */
const oldValue = 1;
oldValue;
      `,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{"oldValue"},
				},
			},
		},
		{
			Code: `
/** @deprecated */
const oldValue = 1;
oldValue;
      `,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{
							"from": "file",
							"name": "oldValue",
						},
					},
				},
			},
		},
		{
			Code: `
/** @deprecated */
const oldValue = 1;
oldValue;
      `,
			Options: map[string]interface{}{
				"allow": []interface{}{
					map[string]interface{}{
						"from": "file",
						"name": "oldValue",
					},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `/** @deprecated */ const oldValue = 1; oldValue;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 1, Column: 40},
			},
		},
		{
			Code: `
/** @deprecated Use newValue instead. */
const oldValue = 1;
oldValue;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecatedWithReason"},
			},
		},
		{
			Code: `
/** @deprecated */
const oldValue = 1;
oldValue;
      `,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{
							"from":    "package",
							"name":    "oldValue",
							"package": "other-pkg",
						},
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated"},
			},
		},
	})
}

func runNoDeprecatedDiagnosticsForFiles(t *testing.T, files map[string]string, entryFile string, options any) []rule.RuleDiagnostic {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	virtualFiles := make(map[string]string, len(files))
	for fileName, content := range files {
		virtualFiles[tspath.ResolvePath(rootDir, fileName)] = content
	}
	fs := utils.NewOverlayVFS(bundled.WrapFS(cachedvfs.From(osvfs.FS())), virtualFiles)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(entryFile)
	if sourceFile == nil {
		t.Fatalf("failed to resolve entry file: %s", entryFile)
	}
	diagnostics := []rule.RuleDiagnostic{}
	var diagnosticsMu sync.Mutex
	_, err = linter.RunLinter(
		[]*compiler.Program{program},
		true,
		[]string{sourceFile.FileName()},
		[]string{},
		func(_ *ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{
				{
					Name:     "test",
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoDeprecatedRule.Run(ctx, options)
					},
				},
			}
		},
		func(diagnostic rule.RuleDiagnostic) {
			diagnosticsMu.Lock()
			diagnostics = append(diagnostics, diagnostic)
			diagnosticsMu.Unlock()
		},
	)
	if err != nil {
		t.Fatalf("error running linter: %v", err)
	}
	return diagnostics
}

func TestAllowFromFileImportedValue(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"deprecated.ts": `
/** @deprecated */
export const oldValue = 1;
		`,
		"main.ts": `
import { oldValue } from './deprecated';
oldValue;
		`,
	}

	t.Run("allow-from-file-does-not-suppress-imported-value", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(
			t,
			files,
			"main.ts",
			[]interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{
							"from": "file",
							"name": "oldValue",
						},
					},
				},
			},
		)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
		}
		if diagnostics[0].Message.Id != "deprecated" {
			t.Fatalf("expected message id deprecated, got %s", diagnostics[0].Message.Id)
		}
	})

	t.Run("without-allow-reports-imported-value", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, files, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
		}
		if diagnostics[0].Message.Id != "deprecated" {
			t.Fatalf("expected message id deprecated, got %s", diagnostics[0].Message.Id)
		}
	})
}
