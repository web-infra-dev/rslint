package number_literal_case_test

import (
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/number_literal_case"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/test/number-literal-case.js
func TestNumberLiteralCaseUpstream(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	for _, value := range []string{
		// Number and BigInt.
		"1234", "0b10", "0o1234567", "0xABCDEF",
		"1234n", "0b10n", "0o1234567n", "0xABCDEFn",
		// Symbolic values and exponential notation.
		"NaN", "+Infinity", "-Infinity", "1.2e3", "1.2e-3", "1.2e+3",
		// Not numbers.
		"'0Xff'", "'0Xffn'",
		// Numeric separators.
		"123_456", "0b10_10", "0o1_234_567", "0xDEED_BEEF",
		"123_456n", "0b10_10n", "0o1_234_567n", "0xDEED_BEEFn",
		// Negative numbers.
		"-1234", "-0b10", "-0o1234567", "-0xABCDEF",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: "const foo = " + value, FileName: "file.js"})
	}

	var invalid []rule_tester.InvalidTestCase
	for _, pair := range [][2]string{
		// Number and BigInt.
		{"0B10", "0b10"}, {"0O1234567", "0o1234567"}, {"0XaBcDeF", "0xABCDEF"},
		{"0B10n", "0b10n"}, {"0O1234567n", "0o1234567n"}, {"0XaBcDeFn", "0xABCDEFn"},
		// BigInt zero.
		{"0B0n", "0b0n"}, {"0O0n", "0o0n"}, {"0X0n", "0x0n"},
		// Exponential notation, including integer mantissas.
		{"1.2E3", "1.2e3"}, {"5E3", "5e3"}, {"5E+3", "5e+3"},
		{"1.2E-3", "1.2e-3"}, {"1.2E+3", "1.2e+3"},
		// Numeric separators and negative numbers.
		{"0XdeEd_Beefn", "0xDEED_BEEFn"},
		{"-0B10", "-0b10"}, {"-0O1234567", "-0o1234567"},
		{"-0XaBcDeF", "-0xABCDEF"}, {"-0XaBcn", "-0xABCn"},
	} {
		invalid = append(invalid, invalidLiteral(t, "const foo = "+pair[0], "const foo = "+pair[1], strings.TrimPrefix(pair[0], "-"), nil))
	}
	const multiline = "const foo = 255;\n\nif (foo === 0xff) {\n\tconsole.log('invalid');\n}"
	invalid = append(invalid, invalidLiteral(t, multiline, strings.Replace(multiline, "0xff", "0xFF", 1), "0xff", nil))
	for _, pair := range [][2]string{
		{"0XaBcDeF", "0xabcdef"}, {"0xaBcDeF", "0xabcdef"},
		{"0XaBcDeFn", "0xabcdefn"}, {"0XdeEd_Beefn", "0xdeed_beefn"},
	} {
		invalid = append(invalid, invalidLiteral(t, "const foo = "+pair[0], "const foo = "+pair[1], pair[0], map[string]any{"hexadecimalValue": "lowercase"}))
	}

	// Upstream's Vue parser service is unavailable in rslint. Keep every Vue
	// case visible rather than treating template text as JavaScript.
	for _, code := range []string{
		`<template><input value="0XdeEd_Beef"></div></template>`,
		`<template><div v-if="0xDEED_BEEF > 0"></div></template>`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code, Skip: true})
	}
	for _, code := range []string{
		`<template><div v-if="0XdeEd_Beef > 0"></div></template>`,
		`<template><div v-if="0XdeEd_Beefn > 0n"></div></template>`,
		`<template><div>{{1.2E3}}</div></template>`,
		`<template><div>{{0B1n}}</div></template>`,
		`<script>export default {data() {return {n: 0XdeEd_Beefn}}}</script>`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Skip: true})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &number_literal_case.NumberLiteralCaseRule, valid, invalid)
}

func TestNumberLiteralCaseUpstreamLegacyOctal(t *testing.T) {
	t.Parallel()
	// tsgo emits parse diagnostics for these sloppy-mode literals, which the
	// Go RuleTester rejects. Check the rule on the parsed literals directly;
	// neither legacy spelling needs a casing diagnostic or an edit.
	for _, code := range []string{"var foo = 0777", "var foo = 0888"} {
		sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: "/legacy.js", Path: "/legacy.js",
		}, code, core.ScriptKindJS)
		if diagnostics := lintLiteralSource(t, sourceFile, rule.EditDemandAll); len(diagnostics) != 0 {
			t.Errorf("legacy literal %q: unexpected diagnostics: %#v", code, diagnostics)
		}
	}
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/number-literal-case.md
func TestNumberLiteralCaseUpstreamDocumentation(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, pair := range [][2]string{
		{"0XFF", "0xFF"}, {"0xff", "0xFF"}, {"0Xff", "0xFF"},
		{"0Xffn", "0xFFn"}, {"0B10", "0b10"}, {"0B10n", "0b10n"},
		{"0O76", "0o76"}, {"0O76n", "0o76n"},
		{"2E-5", "2e-5"}, {"2E+5", "2e+5"}, {"2E5", "2e5"},
	} {
		invalid = append(invalid, invalidLiteral(t, "const foo = "+pair[0]+";", "const foo = "+pair[1]+";", pair[0], nil))
		valid = append(valid, rule_tester.ValidTestCase{Code: "const foo = " + pair[1] + ";", FileName: "file.js"})
	}
	lowercase := map[string]any{"hexadecimalValue": "lowercase"}
	for _, pair := range [][2]string{
		{"0XFF", "0xff"}, {"0xFF", "0xff"}, {"0XFFn", "0xffn"}, {"0xFFn", "0xffn"},
	} {
		invalid = append(invalid, invalidLiteral(t, "const foo = "+pair[0]+";", "const foo = "+pair[1]+";", pair[0], lowercase))
		valid = append(valid, rule_tester.ValidTestCase{Code: "const foo = " + pair[1] + ";", FileName: "file.js", Options: lowercase})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &number_literal_case.NumberLiteralCaseRule, valid, invalid)
}

func invalidLiteral(t testing.TB, code, output, literal string, options any) rule_tester.InvalidTestCase {
	t.Helper()
	start := strings.Index(code, literal)
	if literal == "" || start < 0 || strings.Count(code, literal) != 1 {
		t.Fatalf("expected exactly one occurrence of %q in %q", literal, code)
	}
	line := strings.Count(code[:start], "\n") + 1
	lineStart := strings.LastIndex(code[:start], "\n") + 1
	column := len(utf16.Encode([]rune(code[lineStart:start]))) + 1
	return rule_tester.InvalidTestCase{
		Code: code, FileName: "file.js", Options: options, Output: []string{output},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "number-literal-case", Message: "Invalid number literal casing.",
			Line: line, Column: column, EndLine: line, EndColumn: column + len(literal),
		}},
	}
}

func lintLiteralSource(t testing.TB, sourceFile *ast.SourceFile, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()
	var diagnostics []rule.RuleDiagnostic
	comments := rule.NewCommentStore(sourceFile)
	ctx := rule.RuleContext{
		SourceFile: sourceFile, Comments: comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithDiagnosticConsumer(number_literal_case.NumberLiteralCaseRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
		Demand: demand,
		Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
	})
	listeners := number_literal_case.NumberLiteralCaseRule.Run(ctx, nil)
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
		t.Fatal("expected at least one numeric or BigInt literal")
	}
	return diagnostics
}
