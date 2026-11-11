package numeric_separators_style_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/numeric_separators_style"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNumericSeparatorsStyleExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "const examples = ['123456', /123456/, `123456`];", FileName: "input.js"},
		// Valid JSON options can exceed the size of a Go integer.
		{
			Code:     `const values = [1234567, 0xABCDEF, 1.234567, 1e10000];`,
			FileName: "input.js",
			Options: map[string]any{
				"hexadecimal": map[string]any{"minimumDigits": 1e100},
				"number":      map[string]any{"groupLength": 1e100, "fractionGroupLength": 1e100},
			},
		},
	}
	invalid := []rule_tester.InvalidTestCase{
		// An empty option object uses the defaults. Signs, comments, and CRLF are outside the edit.
		invalidNumericCase("const 文 = '😀';\r\n/*😀*/ const value = - ( /* before */ 120000 /* after */ );",
			"const 文 = '😀';\r\n/*😀*/ const value = - ( /* before */ 120_000 /* after */ );",
			map[string]any{}, numericError(2, 39, 45)),
		invalidNumericCase(`const map = {123456: 0XDEEDBEEF}; map?.[123456];`,
			`const map = {123_456: 0XDE_ED_BE_EF}; map?.[123_456];`, nil,
			numericError(1, 14, 20), numericError(1, 22, 32), numericError(1, 41, 47)),
		invalidNumericCase(`const number = .1234567; const exponent = 1.e+12345;`,
			`const number = .123_456_7; const exponent = 1.e+12_345;`,
			map[string]any{"number": map[string]any{"fractionGroupLength": 3}}, numericError(1, 16, 24), numericError(1, 43, 52)),
		invalidNumericCase(`const value = 123.456e+789;`, `const value = 1_2_3.4_5_6e+7_8_9;`,
			map[string]any{"number": map[string]any{"minimumDigits": 0, "groupLength": 1, "fractionGroupLength": 1}}, numericError(1, 15, 27)),
		// A separator in the exponent enables grouping of the whole literal.
		invalidNumericCase(`const value = 123456.12e-1_2;`, `const value = 123_456.12e-12;`,
			map[string]any{"onlyIfContainsSeparator": true}, numericError(1, 15, 29)),
		invalidNumericCase(`const value = 0o7_77777;`, `const value = 0o777777;`,
			map[string]any{"octal": map[string]any{"minimumDigits": 7, "groupLength": 3}}, numericError(1, 15, 24)),
		invalidNumericCase(`const value = 0O7654_321012345670n;`, `const value = 0O7654_3210_1234_5670n;`,
			map[string]any{"octal": map[string]any{"onlyIfContainsSeparator": true}}, numericError(1, 15, 35)),
		// The runtime value in a JSDoc cast is the only literal that gets edited.
		invalidNumericCase("/** @type {123456} */\nconst value = /** @type {123456} */ (123456);",
			"/** @type {123456} */\nconst value = /** @type {123456} */ (123_456);", nil, numericError(2, 38, 44)),
	}
	typeLiteral := invalidNumericCase(`type Value = -1234567 | 0XdeEdBeeFn;`,
		`type Value = -1_234_567 | 0Xde_Ed_Be_eFn;`, nil, numericError(1, 15, 22), numericError(1, 25, 36))
	typeLiteral.FileName = "input.ts"
	jsx := invalidNumericCase(`const view = <Meter value={123456} title="123456">123456</Meter>;`,
		`const view = <Meter value={123_456} title="123456">123456</Meter>;`, nil, numericError(1, 28, 34))
	jsx.FileName = "input.tsx"
	invalid = append(invalid, typeLiteral, jsx)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &numeric_separators_style.NumericSeparatorsStyleRule, valid, invalid)
}

func TestNumericSeparatorsStyleECMAVersion(t *testing.T) {
	const source = `const size = 12345;`
	oldVersion := rule.LanguageOptions{ECMAVersion: 2020}
	modern := invalidNumericCase(source, `const size = 12_345;`, nil, numericError(1, 14, 19))
	modern.LanguageOptions = rule.LanguageOptions{ECMAVersion: 2021}
	removeSeparator := invalidNumericCase(`const size = 1_234;`, `const size = 1234;`, nil, numericError(1, 14, 19))
	removeSeparator.LanguageOptions = oldVersion
	invalid := []rule_tester.InvalidTestCase{
		{
			Code: source, FileName: "input.js", LanguageOptions: oldVersion,
			Errors: []rule_tester.InvalidTestCaseError{numericError(1, 14, 19)},
		},
		modern,
		removeSeparator,
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &numeric_separators_style.NumericSeparatorsStyleRule, nil, invalid)
}

func TestNumericSeparatorsStyleRecoveredLiterals(t *testing.T) {
	t.Parallel()
	// Unicorn can fix a sloppy-mode leading-zero decimal to invalid syntax.
	// tsgo rejects that spelling. The rule also leaves invalid separators alone.
	const source = `const values = [099999.12, 0_12345, 1__2345n];`
	if diagnostics := lintNumericLiterals(t, source, rule.EditDemandAll); len(diagnostics) != 0 {
		t.Fatalf("unexpected numeric separator diagnostics: %#v", diagnostics)
	}
}

func TestNumericSeparatorsStyleSchema(t *testing.T) {
	t.Parallel()
	for _, options := range [][]any{
		{true},
		{map[string]any{}, map[string]any{}},
		{map[string]any{"onlyIfContainsSeparator": 1}},
		{map[string]any{"decimal": map[string]any{}}},
		{map[string]any{"binary": map[string]any{"fractionGroupLength": 3}}},
		{map[string]any{"octal": map[string]any{"onlyIfContainsSeparator": "true"}}},
		{map[string]any{"octal": map[string]any{"minimumDigits": -1}}},
		{map[string]any{"hexadecimal": map[string]any{"groupLength": 1.5}}},
		{map[string]any{"number": map[string]any{"groupLength": 0}}},
		{map[string]any{"number": map[string]any{"fractionGroupLength": 0}}},
	} {
		if err := numeric_separators_style.NumericSeparatorsStyleRule.Schema.Validate(options); err == nil {
			t.Errorf("invalid options accepted: %#v", options)
		}
	}
}

func TestNumericSeparatorsStyleEditDemand(t *testing.T) {
	t.Parallel()
	const source = `const values = [12345678.9, 0XdeEdBeeFn];`
	const expected = `const values = [12_345_678.9, 0Xde_Ed_Be_eFn];`
	all := lintNumericLiterals(t, source, rule.EditDemandAll)
	if len(all) != 2 {
		t.Fatalf("got %d diagnostics, want 2", len(all))
	}
	for _, diagnostic := range all {
		if diagnostic.Suggestions != nil {
			t.Fatalf("unexpected suggestions: %#v", diagnostic.Suggestions)
		}
	}
	output, unapplied, fixed := linter.ApplyRuleFixes(source, all)
	if !fixed || len(unapplied) != 0 || output != expected {
		t.Fatalf("fix result = %q, unapplied = %d, fixed = %v; want %q", output, len(unapplied), fixed, expected)
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		diagnostics := lintNumericLiterals(t, source, demand)
		if len(diagnostics) != len(all) {
			t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(diagnostics), len(all))
		}
		for i, got := range diagnostics {
			want := all[i]
			if demand&rule.EditDemandAutofix == 0 {
				want.FixesPtr = nil
			}
			// Each run parses a new source file.
			got.SourceFile, want.SourceFile = nil, nil
			if !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d: diagnostic %d differs from the all-edits result", demand, i)
			}
		}
	}
}

func lintNumericLiterals(t *testing.T, code string, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()
	const filename = "/input.js"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: filename, Path: filename}, code, core.ScriptKindJS)
	comments := rule.NewCommentStore(sourceFile)
	var diagnostics []rule.RuleDiagnostic
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithDiagnosticConsumer(numeric_separators_style.NumericSeparatorsStyleRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
		Demand: demand,
		Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
	})
	listeners := numeric_separators_style.NumericSeparatorsStyleRule.Run(ctx, nil)
	literalCount := 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if listener := listeners[node.Kind]; listener != nil {
			literalCount++
			listener(node)
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)
	if literalCount == 0 {
		t.Fatal("expected a numeric or BigInt literal")
	}
	return diagnostics
}
