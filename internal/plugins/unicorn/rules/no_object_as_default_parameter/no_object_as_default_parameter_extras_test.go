// TestNoObjectAsDefaultParameterExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
package no_object_as_default_parameter_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_object_as_default_parameter "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_object_as_default_parameter"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoObjectAsDefaultParameterExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_object_as_default_parameter.NoObjectAsDefaultParameterRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TypeScript wrappers remain visible in TSESTree ----
			{Code: `function f(options = ({cache: true} as const)) {}`, FileName: "file.ts", Tsx: false},
			{Code: `function f(options = ({cache: true} satisfies object)) {}`, FileName: "file.ts", Tsx: false},
			{Code: `function f(options = ({cache: true}!)) {}`, FileName: "file.ts", Tsx: false},
			{Code: `function f(options = source?.defaults) {}`, FileName: "file.ts", Tsx: false},

			// ---- Dimension 4: TSParameterProperty is not a direct function parameter upstream ----
			{Code: `class C { constructor(public options: object = {cache: true}) {} }`, FileName: "file.ts", Tsx: false},

			// ---- Dimension 4: nested binding defaults are not top-level parameters ----
			{Code: `function f({deep: {options = {cache: true}}}) {}`, FileName: "file.js"},

			// ---- Dimension 4: graceful degradation for empty and body-less forms ----
			{Code: `function f(options = {}) {}`, FileName: "file.js"},
			{Code: `function f() {}`, FileName: "file.js"},
			{Code: `abstract class C { abstract method(options?: object): void }`, FileName: "file.ts", Tsx: false},
			{Code: `declare function f(options?: object): void;`, FileName: "file.ts", Tsx: false},

			// Locks in upstream gate arm 1: a missing or non-object right-hand side is ignored.
			{Code: `function f(options) {}`, FileName: "file.js"},
			{Code: `function f(options = defaults) {}`, FileName: "file.js"},

			// Locks in upstream gate arm 2: an empty ObjectExpression is ignored.
			{Code: `const f = (options = {}) => {};`, FileName: "file.js"},

			// Locks in upstream gate arms 3 and 4: AssignmentPatterns outside the
			// direct function parameter list are ignored.
			{Code: `const {options = {cache: true}} = source;`, FileName: "file.js"},
			{Code: `const f = ({options = {cache: true}}) => {};`, FileName: "file.js"},
			{Code: `const f = () => (options = {cache: true});`, FileName: "file.js"},

			// ---- Real-user: #208 — a named defaults object is intentionally outside this syntax-only rule ----
			{Code: `const defaults = {cache: true}; const f = (options = defaults) => {};`, FileName: "file.js"},

			// N/A: element-access, dotted-member, and private-key receiver shapes do
			// not apply; the rule only asks whether an object literal has any member.
			// N/A: overload/abstract/declare signatures cannot legally contain a
			// parameter initializer, so their legal body-less forms are covered above.
			// N/A: the rule has no autofix or suggestion, so edit boundaries and
			// edit-demand invariance do not apply.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: single- and multi-level parentheses are transparent ----
			{Code: `function f(options = ((({cache: true})))) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 12, 1, 19)}},
			{Code: `function f({cache} = ((({cache: true})))) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{nonIdentifierError(1, 25, 1, 38)}},

			// ---- Dimension 4: every legal object-member key/member form makes the object non-empty ----
			{Code: `function f(options = {"cache": true}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 12, 1, 19)}},
			{Code: `function f(options = {0: true}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 12, 1, 19)}},
			{Code: `function f(options = {[key]: true}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 12, 1, 19)}},
			{Code: `function f(options = {...defaults}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 12, 1, 19)}},
			{Code: `function f(options = {cache() {}}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 12, 1, 19)}},
			{Code: `function f(options = {cache}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 12, 1, 19)}},

			// ---- Dimension 4: declaration/container forms and TSESTree identifier ranges ----
			{Code: `function f(options: object = {cache: true}) {}`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 12, 1, 27)}},
			{Code: `const f = function(options: object = {cache: true}) {};`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 20, 1, 35)}},
			{Code: `const f = (options: object = {cache: true}) => {};`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 12, 1, 27)}},
			{Code: `const object = {method(options: object = {cache: true}) {}};`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 24, 1, 39)}},
			{Code: `class C { method(options: object = {cache: true}) {} }`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 18, 1, 33)}},
			{Code: `class C { field = (options: object = {cache: true}) => {}; }`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 20, 1, 35)}},

			// ---- Dimension 4: async, generator, and async-generator containers ----
			{Code: `const f = async (options = {cache: true}) => {};`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 18, 1, 25)}},
			{Code: `function* f(options = {cache: true}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 13, 1, 20)}},
			{Code: `class C { static async * method(options = {cache: true}) {} }`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 33, 1, 40)}},

			// ---- Dimension 4: same-kind nesting reports each direct parameter independently ----
			{
				Code:     `function outer(a = {x: 1}) { return function inner(b = {y: 2}) {} }`,
				FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{
					identifierError("a", 1, 16, 1, 17),
					identifierError("b", 1, 52, 1, 53),
				},
			},

			// ---- Dimension 4: binding-pattern rest/empty shapes report the object default ----
			{Code: `function f({...rest} = {...defaults}) {}`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{nonIdentifierError(1, 24, 1, 37)}},
			{Code: `function f({} = {cache: true}) {}`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{nonIdentifierError(1, 17, 1, 30)}},

			// Locks in upstream identifier-left branch, including the TSESTree range
			// that folds the direct TypeScript annotation into the Identifier.
			{Code: `function f(value: {cache: boolean} = {cache: true}) {}`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{identifierError("value", 1, 12, 1, 35)}},
			{Code: `function f(/** @type {object} */ options = {cache: true}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{identifierError("options", 1, 34, 1, 41)}},

			// Locks in upstream non-Identifier-left branch.
			{Code: `function f([value] = {cache: true}) {}`, FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{nonIdentifierError(1, 22, 1, 35)}},

			// ---- Real-user: #2199 — required TypeScript properties do not exempt the literal default ----
			{Code: `function f(range: {min: number, max: number} = {min: 0, max: 10}) {}`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{identifierError("range", 1, 12, 1, 45)}},

			// ---- Real-user: #1433 — direct destructured parameters are also forbidden ----
			{Code: `function f({cache}: {cache?: boolean} = {cache: false}) {}`, FileName: "file.ts", Tsx: false, Errors: []rule_tester.InvalidTestCaseError{nonIdentifierError(1, 41, 1, 55)}},
		},
	)
}
