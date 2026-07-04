package no_useless_switch_case_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_switch_case"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUselessSwitchCaseExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / real-user scenario it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
func TestNoUselessSwitchCaseExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_useless_switch_case.NoUselessSwitchCaseRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream create() guard arm 1: fewer than two clauses.
			jsValid(lines(
				"switch (foo) {",
				"\tdefault:",
				"\t\thandleDefaultCase();",
				"}",
			)),

			// Locks in upstream create() guard arm 2 / Real-user: #1779 known
			// limitation. A non-last default is intentionally ignored.
			jsValid(lines(
				"switch (value) {",
				"\tcase 1:",
				"\tdefault:",
				"\t\tconsole.log('one');",
				"\tcase 2:",
				"\t\tconsole.log('two');",
				"}",
			)),

			// Locks in upstream loop arm 1: scanning stops at the first non-empty
			// case, so earlier empty cases are not inspected.
			jsValid(lines(
				"switch (foo) {",
				"\tcase a:",
				"\tcase b:",
				"\t\thandleB();",
				"\tdefault:",
				"\t\thandleDefaultCase();",
				"}",
			)),

			// N/A: access/key forms are not inspected by this rule.
			// N/A: declaration/container forms do not affect switch-case
			// semantics; the invalid nested-switch case below covers traversal
			// boundaries.
			// N/A: spread/rest and body-absent declaration forms cannot appear
			// as switch case bodies in a way this rule treats specially.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized nullish case test in JS file ----
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase (undefined):",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				expectedError(2, 2, 2, 19, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
			),

			// ---- Dimension 4: .mjs is not a TypeScript file ----
			{
				Code: lines(
					"switch (foo) {",
					"\tcase undefined:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				FileName: "file.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					expectedError(2, 2, 2, 17, lines(
						"switch (foo) {",
						"\t",
						"\tdefault:",
						"\t\thandleDefaultCase();",
						"}",
					)),
				},
			},

			// ---- Dimension 4: TypeScript nullish cases are normal empty cases ----
			tsInvalid(
				lines(
					"switch (foo) {",
					"\tcase undefined:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				expectedError(2, 2, 2, 17, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
			),

			tsInvalid(
				lines(
					"switch (foo) {",
					"\tcase null:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				expectedError(2, 2, 2, 12, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
			),

			// ---- Dimension 4: parenthesized nullish case tests in TS files ----
			tsInvalid(
				lines(
					"switch (foo) {",
					"\tcase ((undefined)):",
					"\tcase ((null)):",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				expectedError(2, 2, 2, 21, lines(
					"switch (foo) {",
					"\t",
					"\tcase ((null)):",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
				expectedError(3, 2, 3, 16, lines(
					"switch (foo) {",
					"\tcase ((undefined)):",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
			),

			// ---- Dimension 4: .mts follows the same behavior ----
			{
				Code: lines(
					"switch (foo) {",
					"\tcase undefined:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				FileName: "file.mts",
				Errors: []rule_tester.InvalidTestCaseError{
					expectedError(2, 2, 2, 17, lines(
						"switch (foo) {",
						"\t",
						"\tdefault:",
						"\t\thandleDefaultCase();",
						"}",
					)),
				},
			},

			// ---- Real-user: #2670 still reports under upstream semantics ----
			tsInvalid(
				lines(
					"switch (when) {",
					"\tcase 'today': {",
					"\t\treturn date;",
					"\t}",
					"\tcase 'tomorrow': {",
					"\t\treturn today.add({days: 1});",
					"\t}",
					"\tcase undefined:",
					"\tdefault: {",
					"\t\tthrow new RangeError('Bug: Unhandled date keyword');",
					"\t}",
					"}",
				),
				expectedError(8, 2, 8, 17, lines(
					"switch (when) {",
					"\tcase 'today': {",
					"\t\treturn date;",
					"\t}",
					"\tcase 'tomorrow': {",
					"\t\treturn today.add({days: 1});",
					"\t}",
					"\t",
					"\tdefault: {",
					"\t\tthrow new RangeError('Bug: Unhandled date keyword');",
					"\t}",
					"}",
				)),
			),

			// ---- Real-user: #2670 nullable union with explicit nullish arms ----
			tsInvalid(
				lines(
					"switch (state) {",
					"\tcase 'ready':",
					"\t\treturn state;",
					"\tcase null:",
					"\tcase undefined:",
					"\tdefault:",
					"\t\tthrow new Error('Unexpected state');",
					"}",
				),
				expectedError(4, 2, 4, 12, lines(
					"switch (state) {",
					"\tcase 'ready':",
					"\t\treturn state;",
					"\t",
					"\tcase undefined:",
					"\tdefault:",
					"\t\tthrow new Error('Unexpected state');",
					"}",
				)),
				expectedError(5, 2, 5, 17, lines(
					"switch (state) {",
					"\tcase 'ready':",
					"\t\treturn state;",
					"\tcase null:",
					"\t",
					"\tdefault:",
					"\t\tthrow new Error('Unexpected state');",
					"}",
				)),
			),

			// ---- Dimension 4: TS non-null assertion wrapper is a normal empty case ----
			tsInvalid(
				lines(
					"switch (foo) {",
					"\tcase undefined!:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				expectedError(2, 2, 2, 18, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
			),

			// ---- Dimension 4: TS type assertion wrapper is a normal empty case ----
			tsInvalid(
				lines(
					"switch (foo) {",
					"\tcase undefined as any:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				expectedError(2, 2, 2, 24, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
			),

			// ---- Dimension 4: TS satisfies wrapper is a normal empty case ----
			tsInvalid(
				lines(
					"switch (foo) {",
					"\tcase undefined satisfies any:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				expectedError(2, 2, 2, 31, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
			),

			// ---- Dimension 4: optional-chain case test is a normal empty case ----
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase foo?.bar:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				expectedError(2, 2, 2, 16, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
			),

			// ---- Dimension 4: comments before the case colon stay in the report range ----
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a /* comment */:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				expectedError(2, 2, 2, 23, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
			),

			// ---- Dimension 4: multi-line case head range ----
			jsInvalid(
				lines(
					"switch (foo) {",
					"  case (",
					"    a",
					"  ):",
					"  default:",
					"    handleDefaultCase();",
					"}",
				),
				expectedError(2, 3, 4, 5, lines(
					"switch (foo) {",
					"  ",
					"  default:",
					"    handleDefaultCase();",
					"}",
				)),
			),

			// ---- Dimension 4: empty block comments stay empty for this rule ----
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a: {",
					"\t\t// comment",
					"\t}",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				),
				expectedError(2, 2, 2, 9, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"}",
				)),
			),

			// ---- Dimension 4: nested switches report independently ----
			jsInvalid(
				lines(
					"switch (outer) {",
					"\tcase outerA:",
					"\tdefault:",
					"\t\tswitch (inner) {",
					"\t\t\tcase innerA:",
					"\t\t\tdefault:",
					"\t\t\t\thandleInnerDefault();",
					"\t\t}",
					"}",
				),
				expectedError(2, 2, 2, 14, lines(
					"switch (outer) {",
					"\t",
					"\tdefault:",
					"\t\tswitch (inner) {",
					"\t\t\tcase innerA:",
					"\t\t\tdefault:",
					"\t\t\t\thandleInnerDefault();",
					"\t\t}",
					"}",
				)),
				expectedError(5, 4, 5, 16, lines(
					"switch (outer) {",
					"\tcase outerA:",
					"\tdefault:",
					"\t\tswitch (inner) {",
					"\t\t\t",
					"\t\t\tdefault:",
					"\t\t\t\thandleInnerDefault();",
					"\t\t}",
					"}",
				)),
			),

			// Locks in upstream isEmptySwitchCase() branch: a non-empty inner
			// switch consequent prevents the outer case from being reported, but
			// the nested switch is still checked independently.
			jsInvalid(
				lines(
					"switch (outer) {",
					"\tcase outerA:",
					"\t\thandleOuterA();",
					"\tdefault:",
					"\t\tswitch (inner) {",
					"\t\t\tcase innerA:",
					"\t\t\tdefault:",
					"\t\t\t\thandleInnerDefault();",
					"\t\t}",
					"}",
				),
				expectedError(6, 4, 6, 16, lines(
					"switch (outer) {",
					"\tcase outerA:",
					"\t\thandleOuterA();",
					"\tdefault:",
					"\t\tswitch (inner) {",
					"\t\t\t",
					"\t\t\tdefault:",
					"\t\t\t\thandleInnerDefault();",
					"\t\t}",
					"}",
				)),
			),
		},
	)
}
