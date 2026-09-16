package no_process_env_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_process_env"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// AST edge cases checked against eslint-plugin-n v18.3.0.
func TestNoProcessEnvExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_process_env.NoProcessEnvRule,
		[]rule_tester.ValidTestCase{
			// Only the exact process identifier and env property match.
			{
				Code: `process.Env; foo.process.env; this.process.env; globalThis.process.env; getProcess().env; require("node:process").env;`,
			},
			// Computed keys are not evaluated or resolved.
			{
				Code: "process[`env`]; process[\"en\" + \"v\"]; process[null]; process[42]; process[true];",
			},
			// Authored TypeScript wrappers stay visible in ESTree.
			{
				Code: `(process as any).env; process!.env; (process satisfies object).env; (<any>process).env;`,
			},
			// Ordinary type names and type queries are not member expressions.
			{
				Code: `type Env = process.env; type Query = typeof process.env; interface I extends Base<process.env> {}`,
			},
			// JSX member names are not ordinary members.
			{
				Code: `const el = <process.env><process.env.NODE_ENV /></process.env>;`,
				Tsx:  true,
			},
			// Parentheses around the object, member and literal key are transparent.
			{
				Code:    `(process.env).NODE_ENV; process.env[("NODE_ENV")]; (process)["env"]["NODE_ENV"];`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
			},
			// An unbroken optional chain preserves the parent member.
			{
				Code:    `process?.env.NODE_ENV; process.env?.NODE_ENV; process?.["env"]?.["NODE_ENV"];`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
			},
			// Use cooked strings, including empty, Unicode and duplicate option values.
			{
				Code:    `process["e\u006ev"]["N\u004fDE_ENV"]; process.env["😀"]; process.env[""]; process.env["A.B"];`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV", "😀", "", "A.B", "NODE_ENV"}}},
			},
			// Heritage names expose parent members to the allowlist.
			{
				Code:    `interface I extends process.env.NODE_ENV {} class C implements process.env.NODE_ENV {}`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
			},
			// JSDoc casts synthesize wrappers that ESTree omits.
			{
				Code:     `(/** @type {any} */ (process)).env.NODE_ENV; (/** @type {any} */ (process.env)).NODE_ENV;`,
				Options:  []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     `(/** @satisfies {any} */ (process)).env.NODE_ENV; (/** @satisfies {any} */ (process.env)).NODE_ENV;`,
				Options:  []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			// Assertions on computed keys must not become static string matches.
			{
				Code: `process["env" satisfies "env"]; process[("env" as const)];`,
			},
			// Inherited JavaScript object names are ordinary allowlist entries.
			{
				Code:    `process.env.toString; process.env.constructor; process.env["__proto__"];`,
				Options: []any{map[string]any{"allowedVariables": []any{"toString", "constructor", "__proto__"}}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// An empty options object uses the default empty allowlist.
			{
				Code:    `process.env.NODE_ENV`,
				Options: []any{map[string]any{}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 12)},
			},
			// An explicit empty allowlist has the default behavior.
			{
				Code:    `process.env.NODE_ENV`,
				Options: []any{map[string]any{"allowedVariables": []any{}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 12)},
			},
			// Whole-object reads, writes, calls and construction remain forbidden.
			{
				Code: `process.env = {};
delete process.env.NODE_ENV;
const { NODE_ENV } = process.env;
f(process.env);
process.env();
new process.env();`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 12), noProcessEnvAt(2, 8, 2, 19), noProcessEnvAt(3, 22, 3, 33), noProcessEnvAt(4, 3, 4, 14), noProcessEnvAt(5, 1, 5, 12), noProcessEnvAt(6, 5, 6, 16)},
			},
			// Parentheses preserve matches and the reported member range.
			{
				Code:   `((process)).env; process[("env")]; (process.env);`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 16), noProcessEnvAt(1, 18, 1, 34), noProcessEnvAt(1, 37, 1, 48)},
			},
			// Parentheses ending an optional chain insert a ChainExpression parent.
			{
				Code:    `(process?.env).NODE_ENV; (process?.env)?.NODE_ENV; (process?.["env"])["NODE_ENV"];`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 2, 1, 14), noProcessEnvAt(1, 27, 1, 39), noProcessEnvAt(1, 53, 1, 69)},
			},
			// Only literal string keys and dotted identifiers can be allowed.
			{
				Code:    "process.env[NODE_ENV]; process.env[`NODE_ENV`]; process.env[\"NODE_\" + \"ENV\"]; process.env[0]; process.env[true];",
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV", "0", "true"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 12), noProcessEnvAt(1, 24, 1, 35), noProcessEnvAt(1, 49, 1, 60), noProcessEnvAt(1, 79, 1, 90), noProcessEnvAt(1, 95, 1, 106)},
			},
			// TypeScript assertions and non-null expressions block the parent exemption.
			{
				Code:    `(process.env as any).NODE_ENV; process.env!.NODE_ENV; (process.env satisfies object).NODE_ENV; process.env["NODE_ENV" as const];`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 2, 1, 13), noProcessEnvAt(1, 32, 1, 43), noProcessEnvAt(1, 56, 1, 67), noProcessEnvAt(1, 96, 1, 107)},
			},
			// Upstream matches private env names, but never allows private variable names.
			{
				Code:    `class C { #env; #NODE_ENV; m(process) { process.#env; process.#env.NODE_ENV; process.env.#NODE_ENV; } }`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 41, 1, 53), noProcessEnvAt(1, 78, 1, 89)},
			},
			// Shadowing does not affect this syntactic rule.
			{
				Code:   `function f(process) { return process.env; }`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 30, 1, 41)},
			},
			// A member used as a computed key does not gain an exemption.
			{
				Code:    `obj[process.env]; obj[process?.env].NODE_ENV;`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 5, 1, 16), noProcessEnvAt(1, 23, 1, 35)},
			},
			// Class extends and TypeScript heritage names are ESTree members.
			{
				Code: `class C extends process.env {}
interface I extends process.env {}
class D implements process.env {}`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 17, 1, 28), noProcessEnvAt(2, 21, 2, 32), noProcessEnvAt(3, 20, 3, 31)},
			},
			// JSX expression containers are checked while tag names are ignored.
			{
				Code:   `const el = <process.env value={process.env.NODE_ENV}>{process["env"]}</process.env>;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 32, 1, 43), noProcessEnvAt(1, 55, 1, 69)},
			},
			// Ranges use UTF-16 columns and cover multiline members.
			{
				Code: `"😀"; process
  ["env"];`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 7, 2, 10)},
			},
			// JSDoc object casts are transparent to matching.
			{
				Code:     `/** @type {any} */ (process).env`,
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 20, 1, 33)},
			},
			// JSDoc casts do not erase optional-chain boundaries.
			{
				Code:     `(/** @type {any} */ (process?.env)).NODE_ENV;`,
				Options:  []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 22, 1, 34)},
			},
			// Allowed variables are case-sensitive; allowing one does not allow the object.
			{
				Code:    `process.env.node_env; process.env;`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors:  []rule_tester.InvalidTestCaseError{noProcessEnvAt(1, 1, 1, 12), noProcessEnvAt(1, 23, 1, 34)},
			},
			// Instantiation expressions, optional non-null assertions and typed keys
			// remain visible to the parent-member check.
			{
				Code:    `(process.env<string>).NODE_ENV; process?.env!.NODE_ENV; process.env[<any>"NODE_ENV"];`,
				Options: []any{map[string]any{"allowedVariables": []any{"NODE_ENV"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					noProcessEnvAt(1, 2, 1, 13),
					noProcessEnvAt(1, 33, 1, 45),
					noProcessEnvAt(1, 57, 1, 68),
				},
			},
			{
				Code:   "\"😀\";\r\nprocess /* comment */\r\n  [\"env\" /* comment */];",
				Errors: []rule_tester.InvalidTestCaseError{noProcessEnvAt(2, 1, 3, 24)},
			},
		},
	)
}
