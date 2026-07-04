package no_unused_labels

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnusedLabelsUpstream migrates the full valid/invalid suite from upstream
// tests/lib/rules/no-unused-labels.js 1:1. Position assertions cover line/column
// for every invalid case. rslint-specific lock-in cases live in the
// no_unused_labels_extras_test.go file.
func TestNoUnusedLabelsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedLabelsRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid ----
			{Code: `A: break A;`},
			{Code: `A: { foo(); break A; bar(); }`},
			{Code: `A: if (a) { foo(); if (b) break A; bar(); }`},
			{Code: `A: for (var i = 0; i < 10; ++i) { foo(); if (a) break A; bar(); }`},
			{Code: `A: for (var i = 0; i < 10; ++i) { foo(); if (a) continue A; bar(); }`},
			{Code: `A: { B: break B; C: for (var i = 0; i < 10; ++i) { foo(); if (a) break A; if (c) continue C; bar(); } }`},
			{Code: `A: { var A = 0; console.log(A); break A; console.log(A); }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid ----
			{
				Code:   `A: var foo = 0;`,
				Output: []string{`var foo = 0;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unused",
						Message:   "'A:' is defined but never used.",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 2,
					},
				},
			},
			{
				Code:   `A: { foo(); bar(); }`,
				Output: []string{`{ foo(); bar(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `A: if (a) { foo(); bar(); }`,
				Output: []string{`if (a) { foo(); bar(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `A: for (var i = 0; i < 10; ++i) { foo(); if (a) break; bar(); }`,
				Output: []string{`for (var i = 0; i < 10; ++i) { foo(); if (a) break; bar(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `A: for (var i = 0; i < 10; ++i) { foo(); if (a) continue; bar(); }`,
				Output: []string{`for (var i = 0; i < 10; ++i) { foo(); if (a) continue; bar(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `A: for (var i = 0; i < 10; ++i) { B: break A; }`,
				Output: []string{`A: for (var i = 0; i < 10; ++i) { break A; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 35, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code:   `A: { var A = 0; console.log(A); }`,
				Output: []string{`{ var A = 0; console.log(A); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code: `A: /* comment */ foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code: `A /* comment */: foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- upstream invalid: https://github.com/eslint/eslint/issues/16988 ----
			{
				Code: `A: "use strict"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code: `"use strict"; foo: "bar"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 15, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: `A: ("use strict")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code: "A: `use strict`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `if (foo) { bar: 'baz' }`,
				Output: []string{`if (foo) { 'baz' }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 12, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `A: B: 'foo'`,
				Output: []string{`B: 'foo'`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 4, EndLine: 1, EndColumn: 5},
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `A: B: C: 'foo'`,
				Output: []string{`C: 'foo'`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
					{MessageId: "unused", Line: 1, Column: 4, EndLine: 1, EndColumn: 5},
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `A: B: C: D: 'foo'`,
				Output: []string{`D: 'foo'`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "unused", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
					{MessageId: "unused", Line: 1, Column: 4, EndLine: 1, EndColumn: 5},
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `A: B: C: D: E: 'foo'`,
				Output: []string{`E: 'foo'`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "unused", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "unused", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
					{MessageId: "unused", Line: 1, Column: 4, EndLine: 1, EndColumn: 5},
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			{
				Code:   `A: 42`,
				Output: []string{`42`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// SKIP: upstream lists the remaining cases as parser-fatal syntax errors.
		},
	)
}
