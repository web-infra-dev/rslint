package default_rule_test

import (
	"errors"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	default_rule "github.com/web-infra-dev/rslint/internal/plugins/import/rules/default"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

// TestDefaultExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row it covers, so future refactors can't
// silently regress them without breaking a named lock-in.
func TestDefaultExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&default_rule.DefaultRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: declaration forms, side-effect import has no default specifier ----
			{Code: `import "./named-exports";`},
			// ---- Dimension 4: declaration forms, namespace import has no default specifier ----
			{Code: `import * as ns from "./named-exports";`},
			// ---- Dimension 4: declaration forms, named import has no default specifier ----
			{Code: `import { foo } from "./named-exports";`},
			// ---- Dimension 4: access/key forms N/A, import declarations have no receiver/key expression ----
			// ---- Dimension 4: receiver/expression wrappers N/A, default import specifiers cannot be parenthesized ----
			// ---- Dimension 4: nesting/traversal boundaries N/A, ES imports are module-level declarations ----
			// ---- Dimension 4: graceful degradation, unresolved module returns no export map ----
			{Code: `import missing from "./does-not-exist";`},
			// Locks in upstream checkDefault branch 1: defaultSpecifier is absent.
			{Code: `import { default as namedDefault } from "./named-exports";`},
			// Locks in upstream checkDefault branch 2: ExportMapBuilder.get returns null for non-ES modules.
			{Code: `import common from "./common";`},
			// Locks in upstream checkDefault branch 4: exports.get("default") is present.
			{Code: `import value from "./namespace-default";`},
			// ---- Real-user: #54 named export re-exported as default ----
			{Code: `import foo from "./named-default-export";`},
			// ---- Real-user: #545 default re-export from another module ----
			{Code: `import foo from "./default-export-from";`},
			// ---- Dimension 4: explicit unresolved default re-export is not reported ----
			{Code: `import foo from "./reexport-unresolved-as-default";`},
			// ---- Dimension 4: re-export cycle with a local default terminates as valid ----
			{Code: `import foo from "./cycle-with-local-default-a";`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: declaration forms, default import plus named imports still checks the default specifier ----
			{
				Code:     `import missing, { foo } from "./named-exports";`,
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: noDefaultFromNamedExports, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- Dimension 4: declaration forms, type-only default import still has a default specifier ----
			{
				Code:     `import type Missing from "./named-exports";`,
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: noDefaultFromNamedExports, Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			// ---- Dimension 4: diagnostic position on multiline import declaration ----
			{
				Code:     "import\n  baz\nfrom \"./named-exports\";",
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: noDefaultFromNamedExports, Line: 2, Column: 3, EndLine: 2, EndColumn: 6},
				},
			},
			// Locks in upstream checkDefault branch 3: exports.get("default") is undefined.
			{
				Code:     `import missing from "./named-exports";`,
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: noDefaultFromNamedExports, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- Real-user: #328 star exports do not include default ----
			{
				Code: `import barDefault from "./re-export";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: `No default export found in imported module "./re-export".`, Line: 1, Column: 8, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- Dimension 4: explicit default re-export checks the remote local name ----
			{
				Code: `import missing from "./reexport-missing-as-default";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: `No default export found in imported module "./reexport-missing-as-default".`, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- Dimension 4: re-export cycle without a local default terminates as missing ----
			{
				Code: `import cycle from "./cycle-default-a";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: `No default export found in imported module "./cycle-default-a".`, Line: 1, Column: 8, EndLine: 1, EndColumn: 13},
				},
			},
		},
	)
}

func TestDefaultSkippedBabelReExportSyntaxIsNotParsedByTsgo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
	}{
		{name: "default re-export", code: `export bar from "./bar"`},
		{name: "default plus named re-export", code: `export bar, { foo } from "./bar"`},
		{name: "default plus namespace re-export", code: `export bar, * as names from "./bar"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := fixtures.GetRootDir()
			fileName := "file.ts"
			fs := rslint_utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), tc.code)
			host := rslint_utils.CreateCompilerHost(rootDir, fs)
			_, err := rslint_utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
			if err == nil {
				t.Fatalf("expected tsgo parse error for %q", tc.code)
			}
			var syntacticError *rslint_utils.SyntacticError
			if !errors.As(err, &syntacticError) {
				t.Fatalf("expected *utils.SyntacticError for %q, got %T: %v", tc.code, err, err)
			}
		})
	}
}

func TestDefaultSkippedFooES7ParseErrorIsParsedByTsgo(t *testing.T) {
	t.Parallel()

	rootDir := fixtures.GetRootDir()
	fileName := "foo-es7-consumer.ts"
	code := `import Foo from "./jsx/FooES7.js";`
	fs := rslint_utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := rslint_utils.CreateCompilerHost(rootDir, fs)
	if _, err := rslint_utils.CreateProgram(true, fs, rootDir, "tsconfig.allow-js.json", host); err != nil {
		t.Fatalf("expected tsgo to parse upstream FooES7.js fixture, got %T: %v", err, err)
	}
}
