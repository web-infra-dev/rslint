package no_unnecessary_string_trim_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unnecessary_string_trim"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const closureFactoryReceiver = `/** @type {function(): {trim: function(): string}} */
const get = () => ({trim() { return "x"; }});
get().trim().startsWith("x");`

func TestNoUnnecessaryStringTrimExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// JavaScript trims the BOM, non-breaking space, and line separators.
		{Code: `value.trim().startsWith("x\uFEFF"); value.trim().endsWith("\u00A0x"); value.trim().startsWith("x\u2028");`, FileName: "file.js"},
		{Code: `value.trim().endsWith({}); value.trim().startsWith([]); value.trim().startsWith(Symbol.iterator);`, FileName: "file.js"},
		{Code: `let prefix = "x"; prefix = "x "; value.trim().startsWith(prefix);`, FileName: "file.js"},
		{Code: `function f(undefined) { value.trim().startsWith(undefined); }`, FileName: "file.js"},
		{Code: `const value = [load()]; value.trim().startsWith("-"); (function(){}).trim().endsWith("-"); (class {}).trim().startsWith("-"); (()=>{}).trim().endsWith("-");`, FileName: "file.js"},
		{Code: "/** @param {{ trim(): string }} value */\nfunction f(value) { value.trim().startsWith(\"-\"); }", FileName: "file.js"},
		{Code: `/** @param {?{trim: function(): string}} value */
function f(value) { return value.trim().startsWith("x"); }`, FileName: "file.js"},
		{Code: `/** @param {{trim: function(): string}=} value */
function f(value) { return value.trim().startsWith("x"); }`, FileName: "file.js"},
		{Code: closureFactoryReceiver, FileName: "file.js"},
		{Code: "/** @type {number} */\nconst value = get(); value.trim().startsWith(\"-\");", FileName: "file.js"},
		{Code: "/** @returns {number} */\nfunction get() { return 1; } get().trim().startsWith(\"-\");", FileName: "file.js"},
		{Code: `function get(): number { return 1; } get().trim().startsWith("-");`, FileName: "file.ts"},
		{Code: `function f(value: {trim(): string}, fallback: {trim(): string}) {
			let assigned;
			((assigned = value) as unknown).trim().startsWith("-");
			(value ?? fallback).trim().endsWith("-");
		}`, FileName: "file.ts"},
		{Code: `((value?.handler)<string>).prop.trim().startsWith("-");`, FileName: "file.ts"},
		{Code: `import {z} from "zod"; (z.string()<string>).trim().startsWith("-");`, FileName: "file.ts"},
		{Code: `import {z} from "zod"; (z!).string().trim().startsWith("-");`, FileName: "file.ts"},
		{Code: `import {default as z, string as schema} from "zod";
z.string({error: "bad"}).trim().startsWith("x");
schema().trim().endsWith("x");`, FileName: "file.js"},
		{Code: `import {\u007a as schema} from "z\u006fd"; schema.string().trim().startsWith("-");`, FileName: "file.ts"},
		{Code: `import { z } from 'zod';
function createSchema() {
  type z = string;
  return z.string().trim().startsWith('-');
}`, FileName: "file.ts"},
		{Code: `value.trim()!.startsWith("-"); (value.trim as any)().endsWith("-");`, FileName: "file.ts"},
		{Code: `class Example { #trim() {} #startsWith(s) {} f(value) { this.#trim().startsWith("-"); value.trim().#startsWith("-"); } }`, FileName: "file.ts"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// U+0085 is whitespace in Go but is outside JavaScript's trim set.
		invalidTrimCase(`value.trim().startsWith("x\u0085");`, "trimStart", 1, 7, "file.js"),
		invalidTrimCase(`value.trim().startsWith("\uFEFFx");`, "trimStart", 1, 7, "file.js"),
		invalidTrimCase(`value.trim().startsWith("");`, "trimStart", 1, 7, "file.js"),
		invalidTrimCase("value.trim().startsWith(`x${2}`);", "trimStart", 1, 7, "file.js"),
		invalidTrimCase(`const { value } = { value: "ok" }; value.trim().startsWith("o");`, "trimStart", 1, 42, "file.js"),
		invalidTrimCase(`let value = [load()]; value.trim().startsWith("x");`, "trimStart", 1, 29, "file.js"),
		invalidTrimCase(`const value = [load()]; function f(value) { return value.trim().startsWith("x"); }`, "trimStart", 1, 58, "file.js"),
		invalidTrimCase(`(/** @type {number} */ (value)).trim().endsWith("-");`, "trimEnd", 1, 33, "file.js"),
		invalidTrimCase(`/** @param {?string} value */
function f(value) { return value.trim().startsWith("x"); }`, "trimStart", 2, 34, "file.js"),
		invalidTrimCase(`/** @type {function(): string} */
const get = () => "x"; get().trim().startsWith("x");`, "trimStart", 2, 30, "file.js"),
		invalidTrimCase(`/** @type {function(): {trim: function(): string}} */
let get = () => ({trim() { return "x"; }});
get = () => "x";
get().trim().startsWith("x");`, "trimStart", 4, 7, "file.js"),
		// Optional chains in arguments and keys are outside the receiver path.
		invalidTrimCase(`make(value?.name).trim().startsWith("x");`, "trimStart", 1, 19, "file.js"),
		invalidTrimCase(`items[value?.key].trim().endsWith("x");`, "trimEnd", 1, 19, "file.js"),
		invalidTrimCase(`import {type z as schema} from "zod"; schema.string().trim().startsWith("-");`, "trimStart", 1, 55, "file.ts"),
		invalidTrimCase(`import {string as schema} from "zod"; function f(schema) { return schema().trim().endsWith("x"); }`, "trimEnd", 1, 76, "file.js"),
		invalidTrimCase(`value.trim<string>().startsWith<string>("-");`, "trimStart", 1, 7, "file.ts"),
		invalidTrimCase(`const view = <div>{value.trim().endsWith("-")}</div>;`, "trimEnd", 1, 26, "file.tsx"),
		{
			Code:     `"😀". /* note */ tr\u0069m().endsWith("x");`,
			FileName: "file.js",
			Output:   []string{`"😀". /* note */ trimEnd().endsWith("x");`},
			Errors:   []rule_tester.InvalidTestCaseError{trimError("trimEnd", 1, 18, 27)},
		},
	}
	sourceOnlyRule := no_unnecessary_string_trim.NoUnnecessaryStringTrimRule
	sourceOnlyRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return no_unnecessary_string_trim.NoUnnecessaryStringTrimRule.Run(ctx, options)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &sourceOnlyRule, valid, invalid)
}

func TestNoUnnecessaryStringTrimCallableJSDocWithTypes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unnecessary_string_trim.NoUnnecessaryStringTrimRule,
		[]rule_tester.ValidTestCase{{Code: closureFactoryReceiver, FileName: "file.js"}}, nil)
}

func TestNoUnnecessaryStringTrimEditDemand(t *testing.T) {
	const source = `value.trim().endsWith("-");`
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.js",
		Path:     "/edit-demand.js",
	}, source, core.ScriptKindJS)
	comments := rule.NewCommentStore(sourceFile)
	lintRule := no_unnecessary_string_trim.NoUnnecessaryStringTrimRule

	var all rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(lintRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		})
		listener := lintRule.Run(ctx, nil)[ast.KindCallExpression]
		listener(sourceFile.Statements.Nodes[0].Expression())
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
		}
		got := diagnostics[0]
		if demand == rule.EditDemandAll {
			all = got
			if len(all.Fixes()) != 1 || all.Suggestions != nil {
				t.Fatalf("expected one autofix: %#v", all)
			}
		}
		want := all
		if demand&rule.EditDemandAutofix == 0 {
			want.FixesPtr = nil
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d: got %#v, want %#v", demand, got, want)
		}
	}
}
