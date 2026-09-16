package no_new_require_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_new_require"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewRequireExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_require.NoNewRequireRule,
		[]rule_tester.ValidTestCase{
			{Code: `new Require("app-header");`},
			// Only a direct identifier matches, not member names or call results.
			{Code: `new loader.require("x"); new loader["require"]("x"); new require.Widget();`},
			{Code: `class Loader { #require; make() { return new this.#require("x"); } }`},
			{Code: `new (require?.("app-header"))();`},
			{Code: `new (require?.Factory)(); new (require?.["Factory"])();`},
			{Code: `new (0, require)("x"); new (condition ? require : Factory)("x");`},
			// Authored TypeScript wrappers are visible to ESTree.
			{Code: `new (require as any)("app-header");
new (require!)("app-header");
new (require satisfies Function)("app-header");
new (require<string>)("app-header");
new (<Function>require)("app-header");`},
			// Type-only constructor signatures do not contain NewExpressions.
			{Code: `interface Factory { new(): require; } type Constructor = new () => require;`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `new require;`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 1, 1, 12)},
			},
			// Parentheses around the callee are transparent; surrounding ones
			// are excluded from the reported NewExpression range.
			{
				Code:   `new ((require))("app-header");`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 1, 1, 30)},
			},
			{
				Code:   `(new require("app-header"));`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 2, 1, 27)},
			},
			// Upstream matches the name regardless of its binding or spelling.
			{
				Code:   `function load(require) { return new require("app-header"); }`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 33, 1, 58)},
			},
			{
				Code:   `new r\u0065quire("app-header");`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 1, 1, 31)},
			},
			{
				Code:   `new new require("app-header");`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 5, 1, 30)},
			},
			{
				Code: `new require(new require("x"));`,
				Errors: []rule_tester.InvalidTestCaseError{
					noNewRequireAt(1, 1, 1, 30),
					noNewRequireAt(1, 13, 1, 29),
				},
			},
			// Class heritage can contain a runtime constructor expression.
			{
				Code:   `class C extends (new require("x")) {}`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 18, 1, 34)},
			},
			{
				Code:   `new require<string>("app-header");`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 1, 1, 34)},
			},
			{
				Code:   `new require<string>;`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 1, 1, 20)},
			},
			// JavaScript JSDoc casts synthesize wrappers absent from ESTree.
			{
				Code:     `new (/** @type {any} */ (require))("app-header");`,
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 1, 1, 49)},
			},
			{
				Code:     `new (/** @satisfies {Function} */ (require))("x");`,
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 1, 1, 50)},
			},
			{
				Code:     `/** @type {object} */ (new (require)("x"));`,
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 24, 1, 42)},
			},
			// JSX names are not constructors, but expression containers can contain one.
			{
				Code:   `const view = <require>{new require("app-header")}</require>;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 24, 1, 49)},
			},
			{
				Code:   "\"😀\"; new require(\n  \"app-header\"\n);",
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 7, 3, 2)},
			},
			{
				Code:   "new /* constructor */ (\n  require\n)(\"x\");",
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 1, 3, 7)},
			},
			// Leading trivia is excluded, including BOM, hashbang, comments and
			// non-ASCII whitespace; the complete multiline expression is retained.
			{
				Code:     "\uFEFF#!/usr/bin/env node\r\n// comment\r\n\u00a0new require(\r\n  \"x\"\r\n);",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{noNewRequireAt(3, 2, 5, 2)},
			},
		},
	)
}
