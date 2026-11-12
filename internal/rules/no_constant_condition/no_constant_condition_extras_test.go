package no_constant_condition

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expectations were checked against ESLint 10.8.1. Exercise lexical globals,
// coercion boundaries, cooked literals, diagnostic ranges and generator yields
// both with a checker and with binder-only references.
func TestNoConstantConditionExtras(t *testing.T) {
	for _, withoutChecker := range []bool{false, true} {
		name := "typed"
		if withoutChecker {
			name = "without-checker"
		}
		t.Run(name, func(t *testing.T) {
			testedRule := NoConstantConditionRule
			if withoutChecker {
				testedRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
					ctx.TypeChecker = nil
					_, refs, _ := rule.ResolveLanguageDefaults(ctx.SourceFile.FileName(), ctx.LanguageOptions)
					ctx.Refs = rule.NewRefStore(ctx.SourceFile, &core.CompilerOptions{}, nil, refs)
					return NoConstantConditionRule.Run(ctx, options)
				}
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule,
				[]rule_tester.ValidTestCase{
					{Code: "function f(undefined) { if (undefined) {} }"},
					{Code: "{ if (undefined) {} let undefined; }"},
					{Code: "function f() { if (undefined) {} var undefined; }"},
					{Code: "const f = function undefined() { if (undefined) {} };"},
					{Code: "const C = class undefined { method() { if (undefined) {} } };"},
					{Code: "if (Infinity) {}"},
					{Code: "if (String(1)) {}"},
					{Code: "if (Number(1)) {}"},
					{Code: "if (+new Number(1)) {}"},
					{Code: "if (a && undefined) {}"},
					{Code: "if (a &&= undefined) {}"},
					{Code: "if (a || {}) {}"},
					{Code: "if (a ||= []) {}"},
					{Code: "if (a || `foo`) {}"},
					{Code: "if (\"every\" in (\"x\")) {}"},
					{Code: "function f(Boolean) { if (Boolean()) {} }"},
					{Code: "function f() { if (Boolean(1)) {} var Boolean; }"},
					{Code: "if (Boolean?.(a)) {}"},
					{Code: "function* f(){ for (; true; yield) {} }"},
					{Code: "function* f(){ while (Boolean(yield)) {} }"},
					{Code: "function* f(){ do {} while (Boolean(yield)); }"},
					{Code: "function* f(){ for (; Boolean(yield); ) {} }"},
					{Code: "if (Boolean(1) === a) {}"},
					{Code: "if (Boolean?.(1)) {}"},
					{Code: "if (undefined) {}", Globals: map[string]any{"undefined": "off"}},
					{Code: "if (Boolean(1)) {}", Globals: map[string]any{"Boolean": "off"}},
					{Code: "/* global undefined: off */ if (undefined) {}"},
					{Code: "/* global Boolean: off */ if (Boolean(1)) {}"},
					{Code: "function* f(){ while (true) { const o = { [yield](){} }; } }", Options: []any{map[string]any{"checkLoops": "all"}}},
					{Code: "function* f(){ while (true) { class C { *[yield](){yield;} } } }", Options: []any{map[string]any{"checkLoops": "all"}}},
					{Code: "if (a && \"ab\") {}"},
				}, []rule_tester.InvalidTestCase{
					{Code: "if (true ? a : b) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 9}}},
					{Code: "if (Boolean(1, a)) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 18}}},
					{Code: "if (Boolean(1, ...args)) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 24}}},
					{Code: "if (Boolean(...[1])) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20}}},
					{Code: "if (+Boolean([a])) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 18}}},
					{Code: "if (Boolean(1) === true) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 24}}},
					{Code: "if ((Boolean)(1)) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 17}}},
					{Code: "if (a || \"x\") {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
					{Code: "if (a && 0x0) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
					{Code: "if (a && 0.00) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 14}}},
					{Code: "if (a && 0e0) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
					{Code: "if (a && 0x00n) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 15}}},
					{Code: "if (a && 0x0_0n) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}}},
					{Code: "if (a && \"\\\n\") {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 2, EndColumn: 2}}},
					{Code: "if (((undefined))) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 7, EndLine: 1, EndColumn: 16}}},
					{Code: "/* 😀 */\r\nif ( /* test */ (Boolean(1)) ) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 2, Column: 18, EndLine: 2, EndColumn: 28}}},
					{Code: "function* f(){ for (let x = yield; true;) {} }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 36, EndLine: 1, EndColumn: 40}}},
					{Code: "/* global undefined: readonly */ if (undefined) {}", Globals: map[string]any{"undefined": "off"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 38, EndLine: 1, EndColumn: 47}}},
					{Code: "/* global Boolean: readonly */ if (Boolean(1)) {}", Globals: map[string]any{"Boolean": "off"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 36, EndLine: 1, EndColumn: 46}}},
					{Code: "if (Boolean(1)) { function Boolean() {} }", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 15}}},
					{Code: "while (true) { const o = { *m(){ yield; } }; }", Options: []any{map[string]any{"checkLoops": "all"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}}},
					{Code: "while (true) { class C { *m(){ yield; } } }", Options: []any{map[string]any{"checkLoops": "all"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12}}},
					{Code: "if (a || \"ab\") {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Message: "Unexpected constant condition.", Line: 1, Column: 5, EndLine: 1, EndColumn: 14}}},
				})
		})
	}
}
