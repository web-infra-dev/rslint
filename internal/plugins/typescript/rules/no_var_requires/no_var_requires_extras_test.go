package no_var_requires

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Boundary cases were verified against typescript-eslint v8.67.0.
func TestNoVarRequiresParityRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoVarRequiresRule, []rule_tester.ValidTestCase{
		{Code: `const require = () => ({}); const x = require("foo");`},
		{Code: `const x = require("foo"); const require = () => ({});`},
		{Code: `function f(require: Function) { const x = require("foo"); }`},
		{Code: `import require from "custom"; const x = require("foo");`},
		{Code: `import type { require } from "custom"; const x = require("foo");`},
		{Code: `interface require {} const x = require("foo");`},
		{Code: `type require = string; const x = require("foo");`},
		{Code: `declare function require(x: string): any; const x = require("foo");`},
		{Code: `namespace require {} const x = require("foo");`},
		{Code: `const x = require("");`, Options: map[string]any{"allow": []any{"^$"}}},
		{Code: "const x = require(``);", Options: map[string]any{"allow": []any{"^$"}}},
		{Code: "const x = require(`f\\u006fo`);", Options: map[string]any{"allow": []any{"^foo$"}}},
		{Code: `const x = require("f\u006fo");`, Options: map[string]any{"allow": []any{"^foo$"}}},
		{Code: `const x = require(("foo"));`, Options: map[string]any{"allow": []any{"^foo$"}}},
		{Code: "const x = require((`foo`));", Options: map[string]any{"allow": []any{"^foo$"}}},
		{Code: `const x = true ? require("foo") : 0;`},
		{Code: `const x = { value: require("foo") };`},
		{Code: `const x = () => { require("foo"); };`},
		{Code: `const x = require("foo")!;`},
		{Code: `let x; x = require("foo");`},
		{Code: `const x = require("bar");`, Options: map[string]any{"allow": []any{"^foo$", "^bar$"}}},
	}, []rule_tester.InvalidTestCase{
		{Code: `{ const require = () => ({}); const x = require("foo"); } const x = require("foo");`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 69, EndLine: 1, EndColumn: 83},
		}},
		{Code: `function f() { const require = () => ({}); } const x = require("foo");`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 56, EndLine: 1, EndColumn: 70},
		}},
		{Code: `const x = require("\"foo\"");`, Options: map[string]any{"allow": []any{"^foo$"}}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 11, EndLine: 1, EndColumn: 29},
		}},
		{Code: "const x = require(`foo${\"bar\"}`);", Options: map[string]any{"allow": []any{"^foobar$"}}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 11, EndLine: 1, EndColumn: 33},
		}},
		{Code: "const x = require(`foo${unknown}`);", Options: map[string]any{"allow": []any{"^$"}}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 11, EndLine: 1, EndColumn: 35},
		}},
		{Code: `const x = require(name);`, Options: map[string]any{"allow": []any{".*"}}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 11, EndLine: 1, EndColumn: 24},
		}},
		{Code: `const x = require(42);`, Options: map[string]any{"allow": []any{".*"}}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 11, EndLine: 1, EndColumn: 22},
		}},
		{Code: `const x = (require)("foo");`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 11, EndLine: 1, EndColumn: 27},
		}},
		{Code: `const x = (require("foo"));`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 12, EndLine: 1, EndColumn: 26},
		}},
		{Code: `require("foo").default;`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: `require("foo")["default"];`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: `require?.("foo").default;`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
		}},
		{Code: `require("foo")?.default;`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
		}},
		{Code: `const x = require();`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 11, EndLine: 1, EndColumn: 20},
		}},
		{Code: `const x = require("foo");`, Options: map[string]any{"allow": []any{}}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "noVarReqs", Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
		}},
	})
}
