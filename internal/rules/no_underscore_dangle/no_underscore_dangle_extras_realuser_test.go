package no_underscore_dangle

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnderscoreDangleExtrasRealuser locks in code shapes taken from the
// upstream rule's issue tracker and from the ecosystems that hit this rule most
// often, so the port stays aligned on the inputs real projects produce rather
// than only on upstream's contrived cases. Upstream-migrated cases live in
// no_underscore_dangle_upstream_test.go.
func TestNoUnderscoreDangleExtrasRealuser(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnderscoreDangleRule,
		[]rule_tester.ValidTestCase{
			// ---- Real-user: eslint#10755 — prototype assignment is a member access, an object-literal method is not ----
			{Code: `Foo.prototype = { _bar() {} };`, Options: map[string]any{"enforceInMethodNames": false}},

			// ---- Real-user: eslint#11488 — `this.constructor` static lookups ----
			{Code: `class A { m() { return this.constructor._registry; } }`, Options: map[string]any{"allowAfterThisConstructor": true}},

			// ---- Real-user: MongoDB documents — `_id` reads and rest-siblings ----
			{Code: `const { _id, ...rest } = doc;`},
			{Code: `const { _id, ...rest } = doc;`, Options: map[string]any{"allowInObjectDestructuring": false, "allow": []any{"_id"}}},
			{Code: `console.log(doc._id);`, Options: map[string]any{"allow": []any{"_id"}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user: eslint#15810 — enforceInMethodNames alone never reaches class fields ----
			{
				Code:    `class A { _field = 1; #_priv = 1; _method = () => {}; _m() {} }`,
				Options: map[string]any{"enforceInMethodNames": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 55, EndLine: 1, EndColumn: 62},
				},
			},
			{
				Code:    `class A { _field = 1; #_priv = 1; _method = () => {}; _m() {} }`,
				Options: map[string]any{"enforceInMethodNames": true, "enforceInClassFields": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_field'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 22},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '#_priv'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 34},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_method'.", Line: 1, Column: 35, EndLine: 1, EndColumn: 54},
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_m'.", Line: 1, Column: 55, EndLine: 1, EndColumn: 62},
				},
			},

			// ---- Real-user: eslint#10755 — prototype assignment is a member access, an object-literal method is not ----
			{
				Code:    `Foo.prototype._bar = function () {};`,
				Options: map[string]any{"enforceInMethodNames": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},

			// ---- Real-user: eslint#7064 — a doubled underscore still dangles ----
			{
				Code: `const __foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '__foo'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `foo.__bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '__bar'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},

			// ---- Real-user: eslint#11488 — `this.constructor` static lookups ----
			{
				Code:    `class A { m() { return this.constructor._registry; } }`,
				Options: map[string]any{"allowAfterThis": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_registry'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 50},
				},
			},

			// ---- Real-user: MongoDB documents — `_id` reads and rest-siblings ----
			{
				Code:    `const { _id, ...rest } = doc;`,
				Options: map[string]any{"allowInObjectDestructuring": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_id'.", Line: 1, Column: 7, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: `console.log(doc._id);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_id'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},

			// ---- Real-user: dependency-injection constructors keep their parameter properties ----
			{
				Code:    `class Service { constructor(private readonly _http: Http, _plain: Plain) {} }`,
				Options: map[string]any{"allowFunctionParams": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnderscore", Message: "Unexpected dangling '_' in '_plain'.", Line: 1, Column: 59, EndLine: 1, EndColumn: 72},
				},
			},
		},
	)
}
