package number_literal_case_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/number_literal_case"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNumberLiteralCaseExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `const mask = 0xAB_CD;`, Options: map[string]any{}},
		{Code: `const mask = 0xAB_CDn;`, Options: map[string]any{"hexadecimalValue": "uppercase"}},
		{Code: `const values = [0xff, 0xab_cdn, 0b10, 0o76, 1e3];`, Options: map[string]any{"hexadecimalValue": "lowercase"}},
		// Non-numeric tokens, including JSX text, must not be rewritten.
		{Code: "const values = ['0Xff', /0Xff/, `0Xff`, true, false, null];"},
		{Code: `const node = <div title="0Xff">0Xff</div>;`, Tsx: true},
	}
	invalid := []rule_tester.InvalidTestCase{
		invalidLiteral(t, `const mask = 0xff;`, `const mask = 0xFF;`, "0xff", map[string]any{}),
		invalidLiteral(t, `const mask = 0xabn;`, `const mask = 0xABn;`, "0xabn", map[string]any{"hexadecimalValue": "uppercase"}),
		// Lowercase digits do not exempt radix prefixes or decimal exponents.
		invalidLiteral(t, `const value = 0B1n;`, `const value = 0b1n;`, "0B1n", map[string]any{"hexadecimalValue": "lowercase"}),
		invalidLiteral(t, `const value = 1E2;`, `const value = 1e2;`, "1E2", map[string]any{"hexadecimalValue": "lowercase"}),
		// Source spans preserve parentheses, comments, signs and UTF-16 columns.
		invalidLiteral(t, `const value = - /* sign */ ( /* before */ 0Xab /* after */ );`, `const value = - /* sign */ ( /* before */ 0xAB /* after */ );`, "0Xab", nil),
		invalidLiteral(t, "const 文 = '😀';\n/*😀*/ const value = 0Xabn;", "const 文 = '😀';\n/*😀*/ const value = 0xABn;", "0Xabn", nil),
		invalidLiteral(t, `const value = .5E+2;`, `const value = .5e+2;`, ".5E+2", nil),
		invalidLiteral(t, `const value = 1_000E-1_0;`, `const value = 1_000e-1_0;`, "1_000E-1_0", nil),
		// Numeric property names and computed/optional access are literals too.
		invalidLiteral(t, `const object = {0Xab: true};`, `const object = {0xAB: true};`, "0Xab", nil),
		invalidLiteral(t, `object?.[0Xab]`, `object?.[0xAB]`, "0Xab", nil),
		invalidLiteral(t, `class C { #value = 0Xab; }`, `class C { #value = 0xAB; }`, "0Xab", nil),
	}
	for _, pair := range [][3]string{
		{`type Mask = -0Xab;`, `type Mask = -0xAB;`, "0Xab"},
		{`type Mask = 0Xabn;`, `type Mask = 0xABn;`, "0Xabn"},
		{`enum Mask { All = 0Xab }`, `enum Mask { All = 0xAB }`, "0Xab"},
		{`const value = (0Xab as number)!;`, `const value = (0xAB as number)!;`, "0Xab"},
		{`const value = 0Xab satisfies number;`, `const value = 0xAB satisfies number;`, "0Xab"},
	} {
		item := invalidLiteral(t, pair[0], pair[1], pair[2], nil)
		item.FileName = "file.ts"
		invalid = append(invalid, item)
	}
	jsx := invalidLiteral(t, `const node = <div value={0Xab} />;`, `const node = <div value={0xAB} />;`, "0Xab", nil)
	jsx.FileName = "file.tsx"
	jsx.Tsx = true
	invalid = append(invalid, jsx)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &number_literal_case.NumberLiteralCaseRule, valid, invalid)
}

func TestNumberLiteralCaseSchema(t *testing.T) {
	t.Parallel()
	for _, options := range [][]any{
		{}, {map[string]any{}},
		{map[string]any{"hexadecimalValue": "uppercase"}},
		{map[string]any{"hexadecimalValue": "lowercase"}},
	} {
		if err := number_literal_case.NumberLiteralCaseRule.Schema.Validate(options); err != nil {
			t.Errorf("valid options %#v: %v", options, err)
		}
	}
	for _, options := range [][]any{
		{"lowercase"}, {nil},
		{map[string]any{"hexadecimalValue": "mixed"}},
		{map[string]any{"hexadecimalValue": false}},
		{map[string]any{"unknown": true}},
		{map[string]any{}, map[string]any{}},
	} {
		if err := number_literal_case.NumberLiteralCaseRule.Schema.Validate(options); err == nil {
			t.Errorf("invalid options accepted: %#v", options)
		}
	}
}

func TestNumberLiteralCaseEditDemand(t *testing.T) {
	t.Parallel()
	const source = `const values = [0Xab, 0Xabn];`
	const fixedSource = `const values = [0xAB, 0xABn];`
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.ts", Path: "/edit-demand.ts",
	}, source, core.ScriptKindTS)

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()
		diagnostics := lintLiteralSource(t, sourceFile, demand)
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: got %d diagnostics, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	all := run(rule.EditDemandAll)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		for index, diagnostic := range run(demand) {
			if diagnostic.Suggestions != nil {
				t.Errorf("demand %d: unexpected suggestions", demand)
			}
			if demand == rule.EditDemandAutofix || demand == rule.EditDemandAll {
				if diagnostic.FixesPtr == nil || !reflect.DeepEqual(diagnostic.FixesPtr, all[index].FixesPtr) {
					t.Errorf("demand %d: fixes differ from all-edits result", demand)
				}
			} else if diagnostic.FixesPtr != nil {
				t.Errorf("demand %d: unexpected fixes", demand)
			}
			want := all[index]
			want.FixesPtr = nil
			diagnostic.FixesPtr = nil
			if !reflect.DeepEqual(diagnostic, want) {
				t.Errorf("demand %d changed diagnostic metadata", demand)
			}
		}
	}
	output, unapplied, fixed := linter.ApplyRuleFixes(source, all)
	if !fixed || len(unapplied) != 0 || output != fixedSource {
		t.Fatalf("fix result = %q, unapplied = %d, fixed = %v; want %q", output, len(unapplied), fixed, fixedSource)
	}
}
