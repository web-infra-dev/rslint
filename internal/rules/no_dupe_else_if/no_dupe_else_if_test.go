package no_dupe_else_if

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeElseIfRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeElseIfRule,
		[]rule_tester.ValidTestCase{
			// Different conditions
			{Code: `if (a) {} else if (b) {}`},
			{Code: `if (a) {} else if (b) {} else if (c) {}`},
			{Code: `if (a === 1) {} else if (a === 2) {}`},
			{Code: `if (a) {} else if (b) {} else {}`},
			// Nested if (not else-if chain)
			{Code: `if (a) { if (a) {} }`},
			// Different operand order with different operators
			{Code: `if (a > 1) {} else if (a < 1) {}`},
			// Partial overlap that doesn't fully cover
			{Code: `if (a || b) {} else if (c) {}`},
			// Different AND combinations
			{Code: `if (a && b) {} else if (a && c) {}`},
		},
		[]rule_tester.InvalidTestCase{
			// Exact duplicate
			{
				Code: `if (a) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			// Duplicate in 3-way chain
			{
				Code: `if (a) {} else if (b) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			{
				Code: `if (a === 1) {} else if (a === 1) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			// OR subset: if (a || b) covers else if (a)
			{
				Code: `if (a || b) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			// AND superset: if (a) covers else if (a && b)
			{
				Code: `if (a) {} else if (a && b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			// Commutativity of &&: a && b == b && a
			{
				Code: `if (a && b) {} else if (b && a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			// Commutativity of ||: a || b == b || a
			{
				Code: `if (a || b) {} else if (b || a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			// OR covers partial AND: if (a || b) covers else if (b && c) because b ⊆ (b && c)
			{
				Code: `if (a || b) {} else if (b && c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			// AND superset chain
			{
				Code: `if (a && b) {} else if (a && b && c) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
		},
	)
}
