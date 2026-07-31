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
			{Code: `a.b = a.c`, Options: []interface{}{map[string]interface{}{"props": true}}},
			{Code: `a.b = c.b`, Options: []interface{}{map[string]interface{}{"props": true}}},
			{Code: `a.b = a[b]`, Options: []interface{}{map[string]interface{}{"props": true}}},
			{Code: `a[b] = a.b`, Options: []interface{}{map[string]interface{}{"props": true}}},
			{Code: `a.b().c = a.b().c`, Options: []interface{}{map[string]interface{}{"props": true}}},
			{Code: `b().c = b().c`, Options: []interface{}{map[string]interface{}{"props": true}}},
			{Code: `a[b + 1] = a[b + 1]`, Options: []interface{}{map[string]interface{}{"props": true}}},
			{Code: "a.null = a[/(?<zero>0)/]", Options: []interface{}{map[string]interface{}{"props": true}}},
			{Code: `this.x = this.y`, Options: []interface{}{map[string]interface{}{"props": true}}},
			{Code: `a[0] = a[1]`},

			// Member expressions with props:false
			{Code: `a.b = a.b`, Options: []interface{}{map[string]interface{}{"props": false}}},
			{Code: `a.b.c = a.b.c`, Options: []interface{}{map[string]interface{}{"props": false}}},
			{Code: `a[b] = a[b]`, Options: []interface{}{map[string]interface{}{"props": false}}},
			{Code: `a['b'] = a['b']`, Options: []interface{}{map[string]interface{}{"props": false}}},
			{Code: `this.x = this.x`, Options: []interface{}{map[string]interface{}{"props": false}}},
			{Code: `a[0] = a[0]`, Options: []interface{}{map[string]interface{}{"props": false}}},

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
			{
				Code: `\u0061 = \u0061`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Message: "'\\u0061' is assigned to itself.", Line: 1, Column: 10},
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
			// Continue past an earlier same-name property whose value differs.
			// ESLint compares every matching right-hand property, including duplicates.
			{
				Code: `({a} = {a: other, a, a})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Message: "'a' is assigned to itself.", Line: 1, Column: 19},
					{MessageId: "selfAssignment", Message: "'a' is assigned to itself.", Line: 1, Column: 22},
				},
			},
			// Only properties after the final spread have a statically known value.
			{
				Code: `({a} = {a, ...source, a: other, a})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 33},
				},
			},
			// An empty string is a valid static property name, not the sentinel for
			// an unknown computed property.
			{
				Code: `({"": value} = {"": value})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Message: "'value' is assigned to itself.", Line: 1, Column: 21},
				},
			},
			// Exercise the allocation-free cached path with reversed properties.
			// Message text verifies that report order still follows left-hand
			// property order.
			{
				Code: `({p00: v00, p01: v01, p02: v02, p03: v03, p04: v04, p05: v05, p06: v06, p07: v07, p08: v08, p09: v09, p10: v10, p11: v11, p12: v12, p13: v13, p14: v14, p15: v15} = {p15: v15, p14: v14, p13: v13, p12: v12, p11: v11, p10: v10, p09: v09, p08: v08, p07: v07, p06: v06, p05: v05, p04: v04, p03: v03, p02: v02, p01: v01, p00: v00})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Message: "'v00' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v01' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v02' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v03' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v04' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v05' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v06' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v07' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v08' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v09' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v10' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v11' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v12' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v13' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v14' is assigned to itself."},
					{MessageId: "selfAssignment", Message: "'v15' is assigned to itself."},
				},
			},
			// Thirty-three right-hand properties cross the indexing threshold.
			{
				Code: `({target} = {p00: v00, p01: v01, p02: v02, p03: v03, p04: v04, p05: v05, p06: v06, p07: v07, p08: v08, p09: v09, p10: v10, p11: v11, p12: v12, p13: v13, p14: v14, p15: v15, p16: v16, p17: v17, p18: v18, p19: v19, p20: v20, p21: v21, p22: v22, p23: v23, p24: v24, p25: v25, p26: v26, p27: v27, p28: v28, p29: v29, p30: v30, p31: v31, target})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Message: "'target' is assigned to itself."},
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
				Options: []interface{}{map[string]interface{}{"props": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			{
				Code:    `a.b.c = a.b.c`,
				Options: []interface{}{map[string]interface{}{"props": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 9},
				},
			},
			{
				Code:    `a[b] = a[b]`,
				Options: []interface{}{map[string]interface{}{"props": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 8},
				},
			},
			{
				Code:    `a['b'] = a['b']`,
				Options: []interface{}{map[string]interface{}{"props": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 10},
				},
			},
			{
				Code:    `this.x = this.x`,
				Options: []interface{}{map[string]interface{}{"props": true}},
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
				Options: []interface{}{map[string]interface{}{"props": true}},
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
