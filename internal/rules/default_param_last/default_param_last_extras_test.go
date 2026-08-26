package default_param_last

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestDefaultParamLastExtras locks in branches and edge shapes that the upstream test suite doesn't exercise.
// Each case carries an inline comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future refactors can't silently regress them without breaking a named lock-in.
func TestDefaultParamLastExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&DefaultParamLastRule,
		[]rule_tester.ValidTestCase{
			// N/A: receiver/expression wrappers and optional chains cannot wrap a parameter declaration.
			// N/A: member-access and property-key forms are not inspected by this rule.

			// ---- Dimension 4: declaration/container forms ----
			{Code: "async function f(required, optional = 1) {}"},
			{Code: "function* f(required, optional = 1) {}"},
			{Code: "async function* f(required, optional = 1) {}"},
			{Code: "const f = async (required, optional = 1) => {};"},
			{Code: "const object = { method(required, optional = 1) {} };"},
			{Code: "class C { static method(required, optional = 1) {} }"},
			{Code: "const C = class { #method(required, optional = 1) {} };"},
			{Code: "class C { field = (required, optional = 1) => {}; }"},
			{Code: "class C { constructor(required, optional = 1) {} }"},

			// ---- Dimension 4: nesting/traversal boundaries ----
			{Code: "function outer(required, optional = 1) { function inner(required, optional = 1) {} }"},
			{Code: "const outer = (required, optional = 1) => (innerRequired, innerOptional = 1) => 0;"},

			// ---- Dimension 4: graceful degradation ----
			{Code: "function empty() {}"},
			{Code: "function destructured({value}, [fallback] = []) {}"},
			{Code: "function rest(defaulted = 1, ...values) {}"},
			{Code: "abstract class C { abstract method(required: number, optional?: number): void; }"},
			{Code: "interface I { method(optional?: number, required?: number): void; }"},
			// N/A: spread assignments and standalone rest binding elements cannot occur as function-list members.

			// Locks in upstream isRequiredParameter() RestElement arm: rest does not make a preceding default invalid.
			{Code: "function f(defaulted = 1, ...rest) {}"},
			// Locks in upstream handleFunction() report guard false arm: no required parameter exists to the right.
			{Code: "function f(first = 1, second = 2) {}"},
			// Locks in the ESTree-parameter filter: a JSDoc optional tag must not become a TypeScript optional parameter.
			{Code: "/** @param {number} [a] */\nfunction f(a, b) {}", FileName: "file.mjs", TSConfig: "tsconfig.allow-js.json"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: declaration/container forms ----
			{
				Code:   "async function f(defaulted = 1, required) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Message: "Default parameters should be last.", Line: 1, Column: 18, EndLine: 1, EndColumn: 31}},
			},
			{
				Code:   "function* f(defaulted = 1, required) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 13, EndLine: 1, EndColumn: 26}},
			},
			{
				Code:   "const f = function(defaulted = 1, required) {};",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 20, EndLine: 1, EndColumn: 33}},
			},
			{
				Code:   "const f = (defaulted = 1, required) => {};",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 12, EndLine: 1, EndColumn: 25}},
			},
			{
				Code:   "const object = { method(defaulted = 1, required) {} };",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 25, EndLine: 1, EndColumn: 38}},
			},
			{
				Code:   "class C {\n  method(\n    defaulted = 1,\n    required,\n  ) {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 3, Column: 5, EndLine: 3, EndColumn: 18}},
			},
			{
				Code:   "class C { constructor(defaulted = 1, required) {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 23, EndLine: 1, EndColumn: 36}},
			},
			{
				Code:   "class C {\n  constructor(\n    public endpoint = \"localhost\",\n    private retries: number,\n  ) {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 3, Column: 5, EndLine: 3, EndColumn: 34}},
			},
			{
				Code:   "class C { field = (defaulted = 1, required) => {}; }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 20, EndLine: 1, EndColumn: 33}},
			},

			// ---- Dimension 4: nesting/traversal boundaries ----
			{
				Code: "function outer(defaulted = 1, required) {\n  function inner(innerDefault = 2, innerRequired) {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "shouldBeLast", Line: 1, Column: 16, EndLine: 1, EndColumn: 29},
					{MessageId: "shouldBeLast", Line: 2, Column: 18, EndLine: 2, EndColumn: 34},
				},
			},

			// ---- Dimension 4: graceful degradation ----
			{
				Code:   "declare function overloaded(optional?: number, required: number): void;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 29, EndLine: 1, EndColumn: 46}},
			},

			// Locks in upstream isRequiredParameter() AssignmentPattern arm and handleFunction() report-true arm.
			{
				Code:   "function f(defaulted = 1, required) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 12, EndLine: 1, EndColumn: 25}},
			},
			// Locks in upstream isRequiredParameter() optional arm.
			{
				Code:   "function f(optional?: number, required: number) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 12, EndLine: 1, EndColumn: 29}},
			},
			// Locks in upstream handleFunction() required-parameter arm before a later optional parameter.
			{
				Code:   "function f(first: number, optional?: number, required: number) {}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 27, EndLine: 1, EndColumn: 44}},
			},
			// Locks in upstream TSParameterProperty unwrap arm.
			{
				Code:   "class C { constructor(public optional?: number, private required: number) {} }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 23, EndLine: 1, EndColumn: 47}},
			},

			// ---- Real-user: eslint/eslint#11361 optional argument before a required identifier ----
			{
				Code:   `function connect(host = "localhost", port) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 18, EndLine: 1, EndColumn: 36}},
			},
			// ---- Real-user: eslint/eslint#19431 TypeScript parameter-property support ----
			{
				Code:   `class Client { constructor(readonly endpoint = "localhost", retries: number) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldBeLast", Line: 1, Column: 28, EndLine: 1, EndColumn: 59}},
			},
		},
	)
}

// N/A: this rule has no options, fixes, suggestions, key comparisons, or configurable equivalence classes.
