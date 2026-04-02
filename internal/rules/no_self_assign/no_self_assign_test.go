package no_self_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSelfAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoSelfAssignRule,
		[]rule_tester.ValidTestCase{
			// Basic
			{Code: `var a = a`},
			{Code: `a = b`},
			{Code: `a += a`},
			{Code: `a = +a`},
			{Code: `a = [a]`},
			{Code: `a &= a`},
			{Code: `a |= a`},
			{Code: `let a = a`},
			{Code: `const a = a`},
			{Code: `a = a.b`},
			{Code: `a = -a`},

			// Array destructuring
			{Code: `[a] = a`},
			{Code: `[a = 1] = [a]`},
			{Code: `[a, b] = [b, a]`},
			{Code: `[a,, b] = [, b, a]`},
			{Code: `[x, a] = [...x, a]`},
			{Code: `[...a] = [...a, 1]`},
			{Code: `[a, ...b] = [0, ...b, 1]`},
			{Code: `[a, b] = {a, b}`},

			// Object destructuring
			{Code: `({a} = a)`},
			{Code: `({a = 1} = {a})`},
			{Code: `({a: b} = {a})`},
			{Code: `({a} = {a: b})`},
			{Code: `({a} = {a() {}})`},
			{Code: `({a} = {[a]: a})`},
			{Code: `({[a]: b} = {[a]: b})`},
			{Code: "({'foo': a, 1: a} = {'bar': a, 2: a})"},
			{Code: `({a, ...b} = {a, ...b})`},
			{Code: `({a, ...b} = {c, ...b})`},
			{Code: `({a: b} = {a: c})`},

			// Member expressions with props:true (default)
			{Code: `a.b = a.c`, Options: map[string]interface{}{"props": true}},
			{Code: `a.b = c.b`, Options: map[string]interface{}{"props": true}},
			{Code: `a.b = a[b]`, Options: map[string]interface{}{"props": true}},
			{Code: `a[b] = a.b`, Options: map[string]interface{}{"props": true}},
			{Code: `a.b().c = a.b().c`, Options: map[string]interface{}{"props": true}},
			{Code: `b().c = b().c`, Options: map[string]interface{}{"props": true}},
			{Code: `a[b + 1] = a[b + 1]`, Options: map[string]interface{}{"props": true}},
			{Code: "a.null = a[/(?<zero>0)/]", Options: map[string]interface{}{"props": true}},
			{Code: `this.x = this.y`, Options: map[string]interface{}{"props": true}},
			{Code: `a[0] = a[1]`},

			// Member expressions with props:false
			{Code: `a.b = a.b`, Options: map[string]interface{}{"props": false}},
			{Code: `a.b.c = a.b.c`, Options: map[string]interface{}{"props": false}},
			{Code: `a[b] = a[b]`, Options: map[string]interface{}{"props": false}},
			{Code: `a['b'] = a['b']`, Options: map[string]interface{}{"props": false}},
			{Code: `this.x = this.x`, Options: map[string]interface{}{"props": false}},
			{Code: `a[0] = a[0]`, Options: map[string]interface{}{"props": false}},

			// Spread copy
			{Code: `a = {...a}`},
		},
		[]rule_tester.InvalidTestCase{
			// Basic identifiers
			{
				Code: `a = a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 5},
				},
			},

			// Array destructuring
			{
				Code: `[a] = [a]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 8},
				},
			},
			{
				Code: `[a, b] = [a, b]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 11},
					{MessageId: "selfAssignment", Line: 1, Column: 14},
				},
			},
			{
				Code: `[a, b] = [a, c]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 11},
				},
			},
			{
				Code: `[a, b] = [, b]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 13},
				},
			},
			{
				Code: `[a, ...b] = [a, ...b]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 14},
					{MessageId: "selfAssignment", Line: 1, Column: 20},
				},
			},
			{
				Code: `[[a], {b}] = [[a], {b}]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 16},
					{MessageId: "selfAssignment", Line: 1, Column: 21},
				},
			},

			// Object destructuring
			{
				Code: `({a} = {a})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 9},
				},
			},
			{
				Code: `({a: b} = {a: b})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 15},
				},
			},
			{
				Code: "({'a': b} = {'a': b})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 19},
				},
			},
			{
				Code: "({a: b} = {'a': b})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 17},
				},
			},
			{
				Code: "({'a': b} = {a: b})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 17},
				},
			},
			{
				Code: `({1: b} = {1: b})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 15},
				},
			},
			{
				Code: "({1: b} = {'1': b})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 17},
				},
			},
			{
				Code: "({'1': b} = {1: b})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 17},
				},
			},
			{
				Code: "({['a']: b} = {a: b})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 19},
				},
			},
			{
				Code: "({'a': b} = {[`a`]: b})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 21},
				},
			},
			{
				Code: "({1: b} = {[1]: b})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 17},
				},
			},
			{
				Code: `({a, b} = {a, b})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 12},
					{MessageId: "selfAssignment", Line: 1, Column: 15},
				},
			},
			{
				Code: `({a, b} = {b, a})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 15},
					{MessageId: "selfAssignment", Line: 1, Column: 12},
				},
			},
			{
				Code: `({a, b} = {c, a})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 15},
				},
			},
			{
				Code: `({a: {b}, c: [d]} = {a: {b}, c: [d]})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 26},
					{MessageId: "selfAssignment", Line: 1, Column: 34},
				},
			},
			{
				Code: `({a, b} = {a, ...x, b})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 21},
				},
			},

			// Member expressions (props:true default)
			{
				Code: `a.b = a.b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			{
				Code: `a.b.c = a.b.c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 9},
				},
			},
			{
				Code: `a[b] = a[b]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 8},
				},
			},
			{
				Code: `a['b'] = a['b']`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 10},
				},
			},
			{
				Code:    `a.b = a.b`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			{
				Code:    `a.b.c = a.b.c`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 9},
				},
			},
			{
				Code:    `a[b] = a[b]`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 8},
				},
			},
			{
				Code:    `a['b'] = a['b']`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 10},
				},
			},
			{
				Code:    `this.x = this.x`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 10},
				},
			},
			{
				Code: `a[1] = a[1]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 8},
				},
			},
			{
				Code: `a["b"] = a["b"]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 10},
				},
			},
			{
				Code: `a[0] = a[0]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 8},
				},
			},
			// Cross-type member expression: a.b = a['b'] (ESLint uses getStaticPropertyName)
			{
				Code: `a.b = a['b']`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			{
				Code: `a['b'] = a.b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 10},
				},
			},
			// Numeric coercion: a[0] = a['0']
			{
				Code: `a[0] = a['0']`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 8},
				},
			},
			// Multiline element access
			{
				Code: "a[\n    'b'\n] = a[\n    'b'\n]",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 3, Column: 5},
				},
			},
			// Regex literal vs string literal
			{
				Code:    "a['/(?<zero>0)/'] = a[/(?<zero>0)/]",
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 21},
				},
			},

			// Optional chaining - still self-assignment
			{
				Code: `(a?.b).c = (a?.b).c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 12},
				},
			},
			{
				Code: `a.b = a?.b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			{
				Code: `a[0] = a?.[0]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 8},
				},
			},

			// Logical assignment operators
			{
				Code: `a &&= a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			{
				Code: `a ||= a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			{
				Code: `a ??= a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
		},
	)
}
