// cspell:ignore DBFF DFFF
package prefer_unicode_code_point_escapes_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_unicode_code_point_escapes"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferUnicodeCodePointEscapesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_unicode_code_point_escapes.PreferUnicodeCodePointEscapesRule,
		[]rule_tester.ValidTestCase{
			// ordinary literals
			{Code: "const values = [true, 1, null, 'plain'];", Tsx: true},
			// escaped regex backslashes
			{Code: "const p = /\\\\u0061\\\\x61\\\\u{61}/;", Tsx: true},
			// malformed fixed regex escapes
			{Code: "const p = /\\xG0\\u12\\u{}\\c1/;", Tsx: true},
			// out-of-range code point does not wrap
			{Code: "const p = /\\u{FFFFFFFF}\\u{100000000}\\u{0000110000}/;", Tsx: true},
			// tagged head middle and tail
			{Code: "const t = tag`\\x61${x}\\u0062${y}\\x63`;", Tsx: true},
			// string control escapes stay identity escapes
			{Code: "const s = '\\cA';", Tsx: true},
			// null and regex backreferences
			{Code: "const p = /(a)\\1\\0/;", Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Espree normalizes raw template line endings in JavaScript.
			{Code: "const t = `a\r\n\\x61\r\\u0062`;", FileName: "test.js", Output: []string{"const t = `a\n\\u{61}\n\\u{62}`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 3, EndColumn: 8}}},
			// surrogates and following escapes
			{Code: "const s = '\\uD800\\u0041\\uDC00\\uDBFF\\uDFFF\\x00\\xff';", Tsx: true, Output: []string{"const s = '\\u{D800}\\u{41}\\u{DC00}\\u{10FFFF}\\u{0}\\u{FF}';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 51}}},
			// odd backslash run
			{Code: "const s = '\\\\\\u0061';", Tsx: true, Output: []string{"const s = '\\\\\\u{61}';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21}}},
			// template middle
			{Code: "const t = `${a}\\x61${b}`;", Tsx: true, Output: []string{"const t = `${a}\\u{61}${b}`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 15, EndLine: 1, EndColumn: 22}}},
			// nested templates in tagged interpolations
			{Code: "const t = tag`\\x61${`\\x62${x}\\x63`}\\x64`;", Tsx: true, Output: []string{"const t = tag`\\x61${`\\u{62}${x}\\u{63}`}\\x64`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 21, EndLine: 1, EndColumn: 28}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 29, EndLine: 1, EndColumn: 35}}},
			// non-ASCII and multiline template locations
			{Code: "const 名 = \"😀\";\nconst t = `😀\\x61\n\\u0062${x}\\u0063`;", Tsx: true, Output: []string{"const 名 = \"😀\";\nconst t = `😀\\u{61}\n\\u{62}${x}\\u{63}`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 2, Column: 11, EndLine: 3, EndColumn: 9}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 3, Column: 10, EndLine: 3, EndColumn: 18}}},
			// TypeScript template line endings are preserved
			{Code: "const t = `a\r\n\\x61\r\\u0062`;", Tsx: true, Output: []string{"const t = `a\r\n\\u{61}\r\\u{62}`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 3, EndColumn: 8}}},
			// line continuation
			{Code: "const s = 'a\\\n\\x61';", Tsx: true, Output: []string{"const s = 'a\\\n\\u{61}';"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 2, EndColumn: 6}}},
			// computed optional property and parentheses
			{Code: "const obj = ({['\\x61']: 1}); obj?.[('\\u0061')];", Tsx: true, Output: []string{"const obj = ({['\\u{61}']: 1}); obj?.[('\\u{61}')];"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 16, EndLine: 1, EndColumn: 22}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 37, EndLine: 1, EndColumn: 45}}},
			// JSX expression
			{Code: "const el = <div title={'\\x61'}>{'\\u0062'}</div>;", Tsx: true, Output: []string{"const el = <div title={'\\u{61}'}>{'\\u{62}'}</div>;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 24, EndLine: 1, EndColumn: 30}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 33, EndLine: 1, EndColumn: 41}}},
			// JSX quoted attribute follows upstream literal behavior
			{Code: "const el = <div title=\"\\u0061\"/>;", Tsx: true, Output: []string{"const el = <div title=\"\\u{61}\"/>;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 23, EndLine: 1, EndColumn: 31}}},
			// TypeScript literal types and assertions
			{Code: "type T = '\\u0061'; const s = ('\\x62' as const) satisfies string;", Tsx: true, Output: []string{"type T = '\\u{61}'; const s = ('\\u{62}' as const) satisfies string;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 31, EndLine: 1, EndColumn: 37}}},
			// TypeScript template literal type
			{Code: "type T = `\\u0061${string}\\x62`;", Tsx: true, Output: []string{"type T = `\\u{61}${string}\\u{62}`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 25, EndLine: 1, EndColumn: 31}}},
			// private field initializer
			{Code: "class A { #value = '\\x61'; }", Tsx: true, Output: []string{"class A { #value = '\\u{61}'; }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26}}},
			// imports and exports
			{Code: "import { '\\x61' as a } from '\\u0062'; export { a as '\\x63' };", Tsx: true, Output: []string{"import { '\\u{61}' as a } from '\\u{62}'; export { a as '\\u{63}' };"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 10, EndLine: 1, EndColumn: 16}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 29, EndLine: 1, EndColumn: 37}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 53, EndLine: 1, EndColumn: 59}}},
			// regex single surrogates and lowercase controls
			{Code: "const p = /\\uD800\\u0041\\uDC00\\cz[\\cZ]/u;", Tsx: true, Output: []string{"const p = /\\u{D800}\\u{41}\\u{DC00}\\u{1A}[\\u{1A}]/u;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 40}}},
			// legacy nested bracket is not a nested class
			{Code: "const p = /[[\\x61]\\uD83D\\uDCA9/;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const p = /[[\\u{61}]\\u{1F4A9}/u;"}}}}},
			// nested unicode sets and surrogate boundary
			{Code: "const p = /[[\\x61]--[\\u0062]]\\uD83D\\uDCA9/v;", Tsx: true, Output: []string{"const p = /[[\\u{61}]--[\\u{62}]]\\u{1F4A9}/v;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 44}}},
			// escaped class closing bracket
			{Code: "const p = /[\\]\\uD800\\x61]\\\\\\u0062/u;", Tsx: true, Output: []string{"const p = /[\\]\\uD800\\u{61}]\\\\\\u{62}/u;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 36}}},
			// code point with leading zeros needs a flag
			{Code: "const p = /\\u{0000000000000000000000061}/gim;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const p = /\\u{0000000000000000000000061}/gimu;"}}}}},
			// surrogate pair suggestion
			{Code: "const p = /\\uD83D\\uDCA9/;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const p = /\\u{1F4A9}/u;"}}}}},
			// named capture identifiers
			{Code: "const p = /(?<\\u{61}>\\x61)\\k<a>/;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const p = /(?<\\u{61}>\\u{61})\\k<a>/u;"}}}}},
			// exclusive duplicate captures
			{Code: "const p = /(?<name>\\x61)|(?<name>b)/;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const p = /(?<name>\\u{61})|(?<name>b)/u;"}}}}},
			// Unicode property and lookbehind
			{Code: "const p = /(?<=\\x61)\\p{Script=Greek}/i;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const p = /(?<=\\u{61})\\p{Script=Greek}/iu;"}}}}},
			// invalid Unicode identity escape suppresses suggestion
			{Code: "const p = /\\x61\\a/;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19}}},
			// legacy octal regex escape suppresses suggestion
			{Code: "const p = /\\x61\\123/;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21}}},
			// bare bracket suppresses suggestion
			{Code: "const p = /\\x61]/;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 18}}},
			// Annex B quantified lookahead suppresses suggestion
			{Code: "const p = /(?=\\x61)+/;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 22}}},
			// unknown property suppresses suggestion
			{Code: "const p = /\\x61\\p{Unknown}/;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 11, EndLine: 1, EndColumn: 28}}},
		},
	)
}

func lintEscapesSource(code string, demand rule.EditDemand) []rule.RuleDiagnostic {
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.js", Path: "/test.js"}, code, core.ScriptKindJS)
	comments := rule.NewCommentStore(source)
	var diagnostics []rule.RuleDiagnostic
	ctx := (rule.RuleContext{SourceFile: source, Comments: comments, DisableManager: rule.NewDisableManager(source, comments)}).WithDiagnosticConsumer(
		prefer_unicode_code_point_escapes.PreferUnicodeCodePointEscapesRule.Name,
		rule.SeverityError,
		rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
	)
	listeners := prefer_unicode_code_point_escapes.PreferUnicodeCodePointEscapesRule.Run(ctx, nil)
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	return diagnostics
}

func TestPreferUnicodeCodePointEscapesEditDemand(t *testing.T) {
	const source = "const a = '\\x61'; const b = /\\u0061/u; const c = /\\x62/; const d = /\\x61\\_/; const e = `\\u0063`;"
	all := lintEscapesSource(source, rule.EditDemandAll)
	if len(all) != 5 {
		t.Fatalf("got %d diagnostics, want 5", len(all))
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		if disabled := lintEscapesSource("/* eslint-disable unicorn/prefer-unicode-code-point-escapes */\n"+source, demand); len(disabled) != 0 {
			t.Errorf("demand %d: disabled rule produced diagnostics or edits", demand)
		}
		diagnostics := lintEscapesSource(source, demand)
		if len(diagnostics) != len(all) {
			t.Fatalf("demand %d: got %d diagnostics", demand, len(diagnostics))
		}
		for i, d := range diagnostics {
			if d.Range != all[i].Range || !reflect.DeepEqual(d.Message, all[i].Message) {
				t.Errorf("demand %d: diagnostic %d identity changed", demand, i)
			}
			wantFix := (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll) && (i == 0 || i == 1 || i == 4)
			wantSuggestion := (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll) && i == 2
			if (d.FixesPtr != nil) != wantFix || wantFix && !reflect.DeepEqual(d.FixesPtr, all[i].FixesPtr) {
				t.Errorf("demand %d: diagnostic %d has unexpected fixes", demand, i)
			}
			if (d.Suggestions != nil) != wantSuggestion || wantSuggestion && !reflect.DeepEqual(d.Suggestions, all[i].Suggestions) {
				t.Errorf("demand %d: diagnostic %d has unexpected suggestions", demand, i)
			}
		}
	}
	suggestion := (*all[2].Suggestions)[0]
	if suggestion.Message.Id != "prefer-unicode-code-point-escapes/add-unicode-flag" || suggestion.Message.Description != "Use Unicode code point escapes and add the `u` flag." {
		t.Fatalf("unexpected suggestion message: %#v", suggestion.Message)
	}
	fixed, _, applied := linter.ApplyRuleFixes(source, all)
	if !applied || fixed != "const a = '\\u{61}'; const b = /\\u{61}/u; const c = /\\x62/; const d = /\\x61\\_/; const e = `\\u{63}`;" {
		t.Fatalf("unexpected autofix: %q", fixed)
	}
	suggested, _, applied := linter.ApplyRuleFixes(source, []rule.RuleSuggestion{suggestion})
	if !applied || suggested != "const a = '\\x61'; const b = /\\u0061/u; const c = /\\u{62}/u; const d = /\\x61\\_/; const e = `\\u0063`;" {
		t.Fatalf("unexpected suggestion: %q", suggested)
	}
}
