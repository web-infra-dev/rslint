package no_deprecated

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
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
