package prefer_boolean_return_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_boolean_return"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferBooleanReturnExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "function f(x) { if (x) return true; }", FileName: "case.js"},
		{Code: "function f(a, b) { if (a) if (b) return true; return false; }", FileName: "case.js"},
		{Code: "function f(a) { if (a) return true; ; return false; }", FileName: "case.js"},
		{Code: "function f(a) { if (a) {; ;} else return false; }", FileName: "case.js"},
		{Code: "function f(a) { if (a) return true; else {return false; doSomething();} }", FileName: "case.js"},
		{Code: "function f(a) { switch (a) { case 0: if (a) return true; case 1: return false; } }", FileName: "case.js"},
		{Code: "function f(a: unknown, b: unknown) { if (a) {; {return false as const;} } if (b) return false; return true; }", FileName: "case.ts"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Statement and source ranges.
		invalidCase("case.js",
			"function f(x) { if (x) {; {{return (true);}} ;} else {; {return (false);} ;} }",
			"function f(x) { return Boolean(x); }",
			1, 17, 1, 77),
		invalidCase("case.js",
			"function f(a, b) { if (a) return false; ; if (b) return false; return true; }",
			"function f(a, b) { if (a) return false; ; return !b; }",
			1, 43, 1, 63),
		invalidCase("case.js",
			"function f(a, b) { if (a) return true; else work(); if (b) return true; return false; }",
			"function f(a, b) { if (a) return true; else work(); return Boolean(b); }",
			1, 53, 1, 72),
		invalidCase("case.js",
			"function f(x) { switch (x) { case 0: if (x > 0) return true; return false; default: return null; } }",
			"function f(x) { switch (x) { case 0: return x > 0; default: return null; } }",
			1, 38, 1, 61),
		invalidCase("case.js",
			"function f(x) { \"🐱\"; if (x > 0) return true\nreturn false\n}",
			"function f(x) { \"🐱\"; return x > 0;\n}",
			1, 23, 1, 45),
		invalidCase("case.js",
			"function f(x) { if (x) return false; return true;\u2028/* next line */ }",
			"function f(x) { return !x;\u2028/* next line */ }",
			1, 17, 1, 37),
		// Boolean classification and expression grouping.
		invalidCase("case.js",
			"function f(x) { if (((x > 0))) return true; return false; }",
			"function f(x) { return ((x > 0)); }",
			1, 17, 1, 44),
		invalidCase("case.js",
			"function f(a, b) { if ((a, b)) return false; return true; }",
			"function f(a, b) { return !(a, b); }",
			1, 20, 1, 45),
		invalidCase("case.js",
			"function f(a) { if (a++) return false; return true; }",
			"function f(a) { return !(a++); }",
			1, 17, 1, 39),
		invalidCase("case.js",
			"function f(a) { if (++a) return false; return true; }",
			"function f(a) { return !(++a); }",
			1, 17, 1, 39),
		invalidCase("case.js",
			"function f(a, b) { if (a ** b) return false; return true; }",
			"function f(a, b) { return !(a ** b); }",
			1, 20, 1, 45),
		invalidCase("case.js",
			"function* f() { if (yield) return false; return true; }",
			"function* f() { return !(yield); }",
			1, 17, 1, 41),
		invalidCase("case.js",
			"function f(a, b) { if (a in b ?? a instanceof b) return true; return false; }",
			"function f(a, b) { return a in b ?? a instanceof b; }",
			1, 20, 1, 62),
		invalidCase("case.js",
			"function f(a, b) { if (a == b && a != b) return true; return false; }",
			"function f(a, b) { return a == b && a != b; }",
			1, 20, 1, 54),
		invalidCase("case.js",
			"async function f(a) { if (await a()) return false; return true; }",
			"async function f(a) { return !await a(); }",
			1, 23, 1, 51),
		invalidCase("case.tsx",
			"function f() { if (<Item />) return false; return true; }",
			"function f() { return !<Item />; }",
			1, 16, 1, 43),
		invalidCase("case.ts",
			"function f() { if (foo<string>) return false; return true; }",
			"function f() { return !foo<string>; }",
			1, 16, 1, 46),
		invalidCase("case.ts",
			"function f() { if (a?.b!) return false; return true; }",
			"function f() { return !a?.b!; }",
			1, 16, 1, 40),
		invalidCase("case.ts",
			"function f() { if ((a?.b)!) return false; return true; }",
			"function f() { return !((a?.b)!); }",
			1, 16, 1, 42),
		invalidCase("case.ts",
			"function f(x) { if (<boolean>x) return false; return true; }",
			"function f(x) { return !(<boolean>x); }",
			1, 17, 1, 46),
		invalidCase("case.ts",
			"function f(a, b, c) { if ((a ? b : c) as boolean) return true; return false; }",
			"function f(a, b, c) { return Boolean((a ? b : c) as boolean); }",
			1, 23, 1, 63),
		invalidCase("case.ts",
			"function f() { if (true) return (<boolean>(true as const))!; return false; }",
			"function f() { return true; }",
			1, 16, 1, 61),
		// Global calls and shadowing.
		invalidCase("case.js",
			"function f(x) { if ((Boolean)(x) && (Array).isArray(x)) return true; return false; }",
			"function f(x) { return (Boolean)(x) && (Array).isArray(x); }",
			1, 17, 1, 69),
		invalidCase("case.js",
			"function f(x) { if (Boolean?.(x) && Array.isArray(x)) return true; return false; }",
			"function f(x) { return Boolean(Boolean?.(x) && Array.isArray(x)); }",
			1, 17, 1, 67),
		invalidCase("case.js",
			"function f(x) { if (Array?.isArray(x)) return true; return false; }",
			"function f(x) { return Boolean(Array?.isArray(x)); }",
			1, 17, 1, 52),
		invalidCase("case.js",
			"function f(x) { if ((Array?.isArray)(x)) return true; return false; }",
			"function f(x) { return Boolean((Array?.isArray)(x)); }",
			1, 17, 1, 54),
		invalidCase("case.js",
			"function f(x) { if (Array[\"isArray\"](x)) return true; return false; }",
			"function f(x) { return Boolean(Array[\"isArray\"](x)); }",
			1, 17, 1, 54),
		invalidCase("case.js",
			"class C { #isArray(x) { return x; } f(x) { if (Array.#isArray(x)) return true; return false; } }",
			"class C { #isArray(x) { return x; } f(x) { return Boolean(Array.#isArray(x)); } }",
			1, 44, 1, 79),
		invalidCase("case.js",
			"function f(args) { if (Boolean(...args)) return true; return false; }",
			"function f(args) { return Boolean(Boolean(...args)); }",
			1, 20, 1, 54),
		invalidCase("case.js",
			"function f(x) { if (function Boolean() {}) return true; return false; }",
			"function f(x) { return Boolean(function Boolean() {}); }",
			1, 17, 1, 56),
		invalidCase("case.js",
			"/* global Array:off */ function f(x) { if (Array.isArray(x)) return true; return false; }",
			"/* global Array:off */ function f(x) { return Boolean(Array.isArray(x)); }",
			1, 40, 1, 74),
		invalidCase("case.js",
			"/* global Boolean:off */ function f(x) { if (Boolean(x)) return true; return false; }",
			"",
			1, 42, 1, 70),
		invalidCase("case.js",
			"/* global Boolean:off */ function f(x) { if (x) return true; return false; }",
			"",
			1, 42, 1, 61),
		invalidCase("case.js",
			"/* global Boolean:off */ function f(x) { if (x) return false; return true; }",
			"/* global Boolean:off */ function f(x) { return !x; }",
			1, 42, 1, 62),
		invalidCase("case.ts",
			"function f(x: unknown) { if (Array!.isArray(x)) return true; return false; }",
			"function f(x: unknown) { return Boolean(Array!.isArray(x)); }",
			1, 26, 1, 61),
		invalidCase("case.ts",
			"import type { Boolean } from \"./types\"; function f(x: unknown) { if (x) return true; return false; }",
			"",
			1, 66, 1, 85),
		invalidCase("case.ts",
			"import { type Boolean } from \"./types\"; function f(x) { if (x) return true; return false; }",
			"",
			1, 57, 1, 76),
		invalidCase("case.ts",
			"import Boolean = require(\"./types\"); function f(x: unknown) { if (x) return true; return false; }",
			"",
			1, 63, 1, 82),
		invalidCase("case.ts",
			"function check(value: unknown, Boolean: (value: unknown) => string) {\n  {\n    type Boolean = string;\n    if (value) return true;\n    return false;\n  }\n}",
			"",
			4, 5, 4, 28),
		// Comments across statement and cast boundaries.
		invalidCase("case.js",
			"function f(x) { if ((/** @type {boolean} */ (x))) return true; return false; }",
			"function f(x) { return Boolean((/** @type {boolean} */ (x))); }",
			1, 17, 1, 63),
		invalidCase("case.js",
			"function f(x) { if (/** @type {boolean} */ (x)) return true; return false; }",
			"",
			1, 17, 1, 61),
		invalidCase("case.js",
			"function f(x) { if (x) return false; else {return true;} /* next */ }",
			"function f(x) { return !x; /* next */ }",
			1, 17, 1, 57),
		invalidCase("case.js",
			"function f(x) { if (x) return false; else return true; /* next */ }",
			"",
			1, 17, 1, 55),
		invalidCase("case.js",
			"function f(x) { if (x) return true; return false; unreachable(); /* keep */ }",
			"function f(x) { return Boolean(x); unreachable(); /* keep */ }",
			1, 17, 1, 36),
		invalidCase("case.js",
			"function f() { if (/\\/\\//.test(x)) return true; return false; }",
			"function f() { return Boolean(/\\/\\//.test(x)); }",
			1, 16, 1, 48),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_boolean_return.PreferBooleanReturnRule, valid, invalid)
}

func TestPreferBooleanReturnWith(t *testing.T) {
	invalid := []rule_tester.InvalidTestCase{
		invalidCase("case.js",
			"function f(scope, x) { with (scope) { if (x) return true; return false; } }",
			"",
			1, 39, 1, 58),
		invalidCase("case.js",
			"function f(scope, x) { with (scope) { if (Boolean(x)) return true; return false; } }",
			"",
			1, 39, 1, 67),
		invalidCase("case.js",
			"function f(scope, x) { with (scope) { if (Array.isArray(x)) return true; return false; } }",
			"",
			1, 39, 1, 73),
		invalidCase("case.js",
			"function f(scope, x) { with (scope) { if (x > 0) return true; return false; } }",
			"function f(scope, x) { with (scope) { return x > 0; } }",
			1, 39, 1, 62),
		invalidCase("case.js",
			"function f(scope, x) { with (scope) { if (x) return false; return true; } }",
			"function f(scope, x) { with (scope) { return !x; } }",
			1, 39, 1, 59),
		// Evaluating the object expression precedes entering the with body.
		invalidCase("case.js",
			"function f(x) { with ((function () { if (x) return true; return false; })()) {} }",
			"function f(x) { with ((function () { return Boolean(x); })()) {} }",
			1, 38, 1, 57),
	}
	for index := range invalid {
		invalid[index].LanguageOptions.SourceType = "script"
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_boolean_return.PreferBooleanReturnRule, nil, invalid)
}

func TestPreferBooleanReturnEditDemand(t *testing.T) {
	const code = "function f(x) { if (x) return true; return false; }\n" +
		"function g(x, Boolean) { if (x) return true; return false; }"
	const fixed = "function f(x) { return Boolean(x); }\n" +
		"function g(x, Boolean) { if (x) return true; return false; }"

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(code, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	lintProgram := lintprogram.NewFromCompiler(program)
	target := prefer_boolean_return.PreferBooleanReturnRule
	var all []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintProgram, File: sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: target.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners { return target.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			}},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: got %d diagnostics, want 2", demand, len(diagnostics))
		}
		if demand == rule.EditDemandAll {
			all = diagnostics
		}

		for i, diagnostic := range diagnostics {
			if diagnostic.Range != all[i].Range || !reflect.DeepEqual(diagnostic.Message, all[i].Message) || diagnostic.Severity != all[i].Severity {
				t.Errorf("demand %d changed diagnostic %d", demand, i)
			}
			wantFix := i == 0 && demand&rule.EditDemandAutofix != 0
			if (diagnostic.FixesPtr != nil) != wantFix || (wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all[i].FixesPtr)) {
				t.Errorf("demand %d: unexpected fixes for diagnostic %d", demand, i)
			}
			if diagnostic.Suggestions != nil {
				t.Errorf("demand %d: unexpected suggestions for diagnostic %d", demand, i)
			}
		}

		wantOutput := code
		wantFixed := demand&rule.EditDemandAutofix != 0
		if wantFixed {
			wantOutput = fixed
		}
		output, _, didFix := linter.ApplyRuleFixes(code, diagnostics)
		if output != wantOutput || didFix != wantFixed {
			t.Errorf("demand %d: got output %q, want %q", demand, output, wantOutput)
		}
	}
}
