package no_deprecated

import (
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

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
		{
			// Regression: referencing a catch-clause binding must not panic.
			// The binding is a VariableDeclaration whose parent is a
			// CatchClause, not a VariableDeclarationList.
			Code: `
declare function log(x: unknown): void;
try {
  log(1);
} catch (error) {
  log(error);
}
`,
		},
		{
			// A computed access is allowed by the type at its receiver.
			Code: `
interface Thing {
  /** @deprecated */
  oldProp: string;
}
declare const thing: Thing;
thing['oldProp'];
`,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{"from": "file", "name": "Thing"},
					},
				},
			},
		},
		{
			// The receiver is written as the type parameter, not its constraint.
			Code: `
interface Thing {
  /** @deprecated */
  oldProp: string;
}
function g<T extends Thing>(t: T) {
  t['oldProp'];
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{"from": "file", "name": "T"},
					},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			// A computed key names no value, so it never matches a specifier.
			Code: `
interface Thing {
  /** @deprecated */
  oldProp: string;
}
declare const thing: Thing;
thing['oldProp'];
`,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{"oldProp"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated"},
			},
		},
		{
			Code: `
interface Thing {
  /** @deprecated */
  oldProp: string;
}
function g<T extends Thing>(t: T) {
  t['oldProp'];
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{"from": "file", "name": "Thing"},
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated"},
			},
		},
		{
			Code: `/** @deprecated */ const oldValue = 1; oldValue;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 1, Column: 40},
			},
		},
		{
			Code: `/** @deprecated */ let oldValue = undefined; oldValue;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 1, Column: 46},
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

// A specifier that decodes into nothing would otherwise drop the whole allow
// list, leaving the rule reporting everything it was configured to permit.
func TestNoDeprecatedSchemaRejectsUnknownFrom(t *testing.T) {
	t.Parallel()
	allow := func(entry map[string]any) []any {
		return []any{map[string]any{"allow": []any{entry}}}
	}

	if err := NoDeprecatedRule.Schema.Validate(allow(map[string]any{
		"from": "package", "name": "oldValue", "package": "demo-pkg",
	})); err != nil {
		t.Fatalf("expected a well-formed specifier to validate, got %v", err)
	}
	if err := NoDeprecatedRule.Schema.Validate(allow(map[string]any{
		"from": "module", "name": "oldValue",
	})); err == nil {
		t.Fatal("expected an unknown `from` to be rejected")
	}
	if err := NoDeprecatedRule.Schema.Validate(allow(map[string]any{
		"from": "package", "name": "oldValue",
	})); err == nil {
		t.Fatal("expected `from: package` without a package to be rejected")
	}
}

func runNoDeprecatedDiagnosticsForFiles(t *testing.T, files map[string]string, entryFile string, options any) []rule.RuleDiagnostic {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	virtualFiles := make(map[string]string, len(files))
	for fileName, content := range files {
		virtualFiles[tspath.ResolvePath(rootDir.Dir, fileName)] = content
	}
	fs := utils.NewOverlayVFS(cachedvfs.From(rootDir.FS), virtualFiles)
	host := utils.CreateCompilerHost(rootDir.Dir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(entryFile)
	if sourceFile == nil {
		t.Fatalf("failed to resolve entry file: %s", entryFile)
		return nil
	}
	diagnostics := []rule.RuleDiagnostic{}
	var diagnosticsMu sync.Mutex
	_, err = linter.RunLinter(linter.RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		Scope: linter.FileScope{
			Files: []string{sourceFile.FileName()},
		},
		GetRulesForFile: func(_ *ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{
				{
					Name:     "test",
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoDeprecatedRule.Run(ctx, rule_tester.ResolveTestCaseOptions(t, &NoDeprecatedRule, options))
					},
				},
			}
		},
		Consumer: rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnosticsMu.Lock()
				diagnostics = append(diagnostics, diagnostic)
				diagnosticsMu.Unlock()
			},
		},
	})
	if err != nil {
		t.Fatalf("error running linter: %v", err)
	}
	return diagnostics
}

// The deprecated value's type is the literal `1`, which carries no name, so
// every one of these specifiers can only match the referenced value. Matching a
// value narrows past its name for `from: "package"` alone.
func TestAllowSpecifierMatchesImportedValue(t *testing.T) {
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

	allow := func(entries ...interface{}) []interface{} {
		return []interface{}{map[string]interface{}{"allow": entries}}
	}

	for _, testCase := range []struct {
		name     string
		options  []interface{}
		reported bool
	}{
		{name: "without-allow", options: nil, reported: true},
		{name: "shorthand-name", options: allow("oldValue")},
		{name: "from-file", options: allow(map[string]interface{}{"from": "file", "name": "oldValue"})},
		{
			name:    "from-file-with-unrelated-path",
			options: allow(map[string]interface{}{"from": "file", "path": "./nope.ts", "name": "oldValue"}),
		},
		{name: "from-lib", options: allow(map[string]interface{}{"from": "lib", "name": "oldValue"})},
		{
			name:     "from-package",
			options:  allow(map[string]interface{}{"from": "package", "package": "nope-pkg", "name": "oldValue"}),
			reported: true,
		},
		{
			name:     "unrelated-name",
			options:  allow(map[string]interface{}{"from": "file", "name": "other"}),
			reported: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			diagnostics := runNoDeprecatedDiagnosticsForFiles(t, files, "main.ts", testCase.options)
			expected := 0
			if testCase.reported {
				expected = 1
			}
			if len(diagnostics) != expected {
				t.Fatalf("expected %d diagnostics, got %d", expected, len(diagnostics))
			}
			if testCase.reported && diagnostics[0].Message.Id != "deprecated" {
				t.Fatalf("expected message id deprecated, got %s", diagnostics[0].Message.Id)
			}
		})
	}
}

func TestNoDeprecatedReportsLetInitializedToUndefinedInTsxFixture(t *testing.T) {
	t.Parallel()
	diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
		"react.tsx": `
        /** @deprecated */ let a = undefined;
        a;
      `,
	}, "react.tsx", nil)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
	}
	if diagnostics[0].Message.Id != "deprecated" {
		t.Fatalf("expected message id deprecated, got %s", diagnostics[0].Message.Id)
	}
}

func TestNoDeprecatedIgnoresNameOnlyFallbackForElementAccessAndJsxAny(t *testing.T) {
	t.Parallel()

	t.Run("element-access-cross-object-collision", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": `
const a = { /** @deprecated */ b: 'string' };
const a2 = { b: 'string' };
a2['b'];
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected 0 diagnostics, got %#v", diagnostics)
		}
	})

	t.Run("element-access-any-typed-object", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": `
interface U1 { /** @deprecated */ shared: number }
const obj: any = {};
obj['shared'];
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected 0 diagnostics, got %#v", diagnostics)
		}
	})

	t.Run("jsx-attribute-any-props", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.tsx": `
declare namespace JSX {
  interface Element {}
  interface IntrinsicAttributes {}
}

interface U2 { /** @deprecated */ foo: number }
declare function Comp(p: any): any;
const x = <Comp foo={1} />;
			`,
		}, "main.tsx", nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected 0 diagnostics, got %#v", diagnostics)
		}
	})
}

func TestNoDeprecatedMultilineImportsAndMultiDeclarators(t *testing.T) {
	t.Parallel()

	t.Run("shorthand-property-assignment-still-reports-deprecated-variable", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": `
/** @deprecated */
declare const test: string;
const bar = { test };
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		if diagnostics[0].Message.Id != "deprecated" {
			t.Fatalf("expected deprecated message id, got %s", diagnostics[0].Message.Id)
		}
		if !strings.Contains(diagnostics[0].Message.Description, "`test`") {
			t.Fatalf("expected test diagnostic, got %#v", diagnostics[0].Message)
		}
	})

	t.Run("multiline-import-only-reports-usage", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"dep.ts": `
/** @deprecated */
export const oldValue = 1;
			`,
			"main.ts": `
import {
  oldValue,
} from './dep';

oldValue;
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		if diagnostics[0].Message.Id != "deprecated" {
			t.Fatalf("expected deprecated message id, got %s", diagnostics[0].Message.Id)
		}
		if !strings.Contains(diagnostics[0].Message.Description, "oldValue") {
			t.Fatalf("expected oldValue diagnostic, got %#v", diagnostics[0].Message)
		}
	})

	t.Run("statement-jsdoc-only-applies-to-first-declarator", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": `
/** @deprecated */ const a = 1, b = 2;
a;
b;
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		if !strings.Contains(diagnostics[0].Message.Description, "`a`") {
			t.Fatalf("expected only a to be reported, got %#v", diagnostics[0].Message)
		}
	})
}
