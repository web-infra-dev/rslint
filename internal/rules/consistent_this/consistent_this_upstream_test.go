// TestConsistentThisUpstream migrates the full valid/invalid suite from
// ESLint v10.8.1 tests/lib/rules/consistent-this.js 1:1. rslint's parser
// does not gate syntax on languageOptions.ecmaVersion the way espree does, so
// that option is dropped from ported cases (this rule's behavior does not
// depend on it, apart from the sourceType variant noted below). rslint infers
// module-ness from actual import/export syntax rather than an explicit
// languageOptions.sourceType flag; the one upstream case that sets
// sourceType: "module" has no such syntax, so it is migrated without the
// flag — it is expected to (and does) behave identically to the immediately
// preceding non-module case. rslint-specific lock-ins live in
// consistent_this_extras_test.go.
package consistent_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentThisUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentThisRule,
		[]rule_tester.ValidTestCase{
			{Code: "var foo = 42, that = this"},
			{Code: "var foo = 42, self = this", Options: []any{"self"}},
			{Code: "var self = 42", Options: []any{"that"}},
			{Code: "var self", Options: []any{"that"}},
			{Code: "var self; self = this", Options: []any{"self"}},
			{Code: "var foo, self; self = this", Options: []any{"self"}},
			{Code: "var foo, self; foo = 42; self = this", Options: []any{"self"}},
			{Code: "self = 42", Options: []any{"that"}},
			{Code: "var foo = {}; foo.bar = this", Options: []any{"self"}},
			{Code: "var self = this; var vm = this;", Options: []any{"self", "vm"}},

			// destructuringTest cases: options ["self"], ecmaVersion 6.
			{Code: "var {foo, bar} = this", Options: []any{"self"}},
			{Code: "({foo, bar} = this)", Options: []any{"self"}},
			{Code: "var [foo, bar] = this", Options: []any{"self"}},
			{Code: "[foo, bar] = this", Options: []any{"self"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "var context = this",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAlias", Message: "Unexpected alias 'context' for 'this'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var that = this",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAlias", Line: 1, Column: 5, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "var foo = 42, self = this",
				Options: []any{"that"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAlias", Line: 1, Column: 15, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "var self = 42",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Message: "Designated alias 'self' is not assigned to 'this'.", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    "var self",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    "var self; self = 42",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 11, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "context = this",
				Options: []any{"that"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAlias", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "that = this",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAlias", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    "self = this",
				Options: []any{"that"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAlias", Line: 1, Column: 1, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    "self += this",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    "var self; (function() { self = this; }())",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},
			// Upstream's sourceType: "module" variant of the case above — see
			// the file header for why the flag itself is dropped.
			{
				Code:    "var self; (function() { self = this; }())",
				Options: []any{"self"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "aliasNotAssignedToThis", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},
		},
	)
}
