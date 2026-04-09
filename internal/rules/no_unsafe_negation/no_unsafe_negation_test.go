package no_unsafe_negation

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeNegationRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeNegationRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// ===== Basic relational =====
			{Code: `a in b`},
			{Code: `a instanceof b`},

			// Comparing relational result is fine
			{Code: `a in b === false`},
			{Code: `a instanceof b === false`},

			// ===== Parenthesized negation (intentional) =====
			{Code: `(!a) in b`},
			{Code: `(!a) instanceof b`},

			// Multiple layers of parenthesized negation
			{Code: `((!a)) in b`},
			{Code: `(((!a))) instanceof b`},

			// ===== Negation of the whole expression =====
			{Code: `!(a in b)`},
			{Code: `!(a instanceof b)`},

			// Properly negated compound expression
			{Code: `!(a in b && c in d)`},

			// ===== Other unary operators — should NEVER trigger =====
			{Code: `~a in b`},
			{Code: `typeof a in b`},
			{Code: `void a in b`},
			{Code: `-a instanceof b`},
			{Code: `+a in b`},

			// ===== Non-relational binary operators — should NEVER trigger =====
			// Equality
			{Code: `!a == b`},
			{Code: `!a != b`},
			{Code: `!a === b`},
			{Code: `!a !== b`},
			// Arithmetic
			{Code: `!a + b`},
			{Code: `!a - b`},
			{Code: `!a * b`},
			{Code: `!a / b`},
			{Code: `!a % b`},
			// Bitwise
			{Code: `!a & b`},
			{Code: `!a | b`},
			{Code: `!a ^ b`},
			{Code: `!a << b`},
			{Code: `!a >> b`},
			{Code: `!a >>> b`},
			// Logical
			{Code: `!a && b`},
			{Code: `!a || b`},

			// ===== Proper usage in assignment =====
			{Code: `x = (!a) in b`},
			{Code: `x = !(a in b)`},

			// ===== for-in (not a BinaryExpression) =====
			{Code: `for (var x in obj) {}`},

			// ===== Ordering relations NOT checked by default =====
			{Code: `!a < b`},
			{Code: `!a > b`},
			{Code: `!a <= b`},
			{Code: `!a >= b`},
			{Code: `if (! a < b) {}`},
			{Code: `while (! a > b) {}`},
			{Code: `foo = ! a <= b`},
			{Code: `foo = ! a >= b`},
			// Complex operand ordering — not checked by default
			{Code: `!a.b < c`},
			{Code: `!f() > d`},
			{Code: `!a[0] <= b`},

			// Empty options object still defaults to not enforcing ordering
			{
				Code:    `! a <= b`,
				Options: map[string]interface{}{},
			},
			// Explicitly disabled enforceForOrderingRelations
			{
				Code:    `foo = ! a >= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": false},
			},

			// ===== Parenthesized negation with ordering (option enabled) =====
			{
				Code:    `(!a) < b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},
			{
				Code:    `(!a) > b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},
			{
				Code:    `(!a) <= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},
			{
				Code:    `(!a) >= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},
			{
				Code:    `foo = (!a) >= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},

			// Normal comparisons with option enabled (no negation)
			{
				Code:    `a <= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},
			{
				Code:    `foo = a > b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},

			// Properly negated whole expression with option enabled
			{
				Code:    `!(a < b)`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// ===== Basic in / instanceof =====
			{
				Code: `!a in b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a in b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a) in b`},
						},
					},
				},
			},
			{
				Code: `!a instanceof b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a instanceof b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a) instanceof b`},
						},
					},
				},
			},

			// ===== Parenthesized context =====
			// Outer parenthesized: (!a in b)
			{
				Code: `(!a in b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `(!(a in b))`},
							{MessageId: "suggestParenthesisedNegation", Output: `((!a) in b)`},
						},
					},
				},
			},
			// Double outer parenthesized: ((!a in b))
			{
				Code: `((!a in b))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `((!(a in b)))`},
							{MessageId: "suggestParenthesisedNegation", Output: `(((!a) in b))`},
						},
					},
				},
			},
			// Negation of parenthesized identifier: !(a) in b
			{
				Code: `!(a) in b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!((a) in b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!(a)) in b`},
						},
					},
				},
			},
			// Outer parenthesized instanceof
			{
				Code: `(!a instanceof b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `(!(a instanceof b))`},
							{MessageId: "suggestParenthesisedNegation", Output: `((!a) instanceof b)`},
						},
					},
				},
			},
			// !(a) instanceof b
			{
				Code: `!(a) instanceof b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!((a) instanceof b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!(a)) instanceof b`},
						},
					},
				},
			},

			// ===== Complex left operand =====
			// Double negation: !!a in b
			{
				Code: `!!a in b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(!a in b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!!a) in b`},
						},
					},
				},
			},
			// Negation of property access: !a.b in c
			{
				Code: `!a.b in c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a.b in c)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a.b) in c`},
						},
					},
				},
			},
			// Negation of call expression: !f() in b
			{
				Code: `!f() in b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(f() in b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!f()) in b`},
						},
					},
				},
			},
			// Negation of call expression with instanceof: !f() instanceof b
			{
				Code: `!f() instanceof b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(f() instanceof b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!f()) instanceof b`},
						},
					},
				},
			},
			// Negation of computed property: !a[0] in b
			{
				Code: `!a[0] in b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a[0] in b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a[0]) in b`},
						},
					},
				},
			},
			// Negation of new expression: !new Foo() instanceof Bar
			{
				Code: `!new Foo() instanceof Bar`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(new Foo() instanceof Bar)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!new Foo()) instanceof Bar`},
						},
					},
				},
			},
			// Negation with optional chaining: !a?.b in c
			{
				Code: `!a?.b in c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a?.b in c)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a?.b) in c`},
						},
					},
				},
			},
			// Negation of parenthesized complex expression: !(a + b) in c
			{
				Code: `!(a + b) in c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!((a + b) in c)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!(a + b)) in c`},
						},
					},
				},
			},

			// ===== Complex right operand =====
			// Negation with complex right: !a in b.c.d
			{
				Code: `!a in b.c.d`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a in b.c.d)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a) in b.c.d`},
						},
					},
				},
			},

			// ===== Expression context =====
			// In logical expression: !a in b && c
			{
				Code: `!a in b && c`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a in b) && c`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a) in b && c`},
						},
					},
				},
			},
			// In ternary condition: !a in b ? c : d
			{
				Code: `!a in b ? c : d`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a in b) ? c : d`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a) in b ? c : d`},
						},
					},
				},
			},
			// In assignment: x = !a in b
			{
				Code: `x = !a in b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 5,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `x = !(a in b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `x = (!a) in b`},
						},
					},
				},
			},
			// In template literal
			{
				Code: "var s = `${!a in b}`",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: "var s = `${!(a in b)}`"},
							{MessageId: "suggestParenthesisedNegation", Output: "var s = `${(!a) in b}`"},
						},
					},
				},
			},
			// In array literal
			{
				Code: `var arr = [!a in b]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `var arr = [!(a in b)]`},
							{MessageId: "suggestParenthesisedNegation", Output: `var arr = [(!a) in b]`},
						},
					},
				},
			},
			// In function argument
			{
				Code: `foo(!a in b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 5,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `foo(!(a in b))`},
							{MessageId: "suggestParenthesisedNegation", Output: `foo((!a) in b)`},
						},
					},
				},
			},
			// In arrow function body
			{
				Code: `var fn = () => !a in b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `var fn = () => !(a in b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `var fn = () => (!a) in b`},
						},
					},
				},
			},

			// ===== Multiple violations =====
			// Two violations in comma expression
			{
				Code: `!a in b, !c instanceof d`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a in b), !c instanceof d`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a) in b, !c instanceof d`},
						},
					},
					{
						MessageId: "unexpected", Line: 1, Column: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!a in b, !(c instanceof d)`},
							{MessageId: "suggestParenthesisedNegation", Output: `!a in b, (!c) instanceof d`},
						},
					},
				},
			},

			// ===== Whitespace preservation =====
			{
				Code:    `if (! a < b) {}`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 5,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `if (!( a < b)) {}`},
							{MessageId: "suggestParenthesisedNegation", Output: `if ((! a) < b) {}`},
						},
					},
				},
			},
			{
				Code:    `while (! a > b) {}`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `while (!( a > b)) {}`},
							{MessageId: "suggestParenthesisedNegation", Output: `while ((! a) > b) {}`},
						},
					},
				},
			},
			{
				Code:    `foo = ! a <= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `foo = !( a <= b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `foo = (! a) <= b`},
						},
					},
				},
			},
			{
				Code:    `foo = ! a >= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `foo = !( a >= b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `foo = (! a) >= b`},
						},
					},
				},
			},
			{
				Code:    `! a <= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!( a <= b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(! a) <= b`},
						},
					},
				},
			},

			// ===== Ordering without space =====
			{
				Code:    `!a < b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a < b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a) < b`},
						},
					},
				},
			},
			{
				Code:    `!a > b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a > b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a) > b`},
						},
					},
				},
			},
			{
				Code:    `!a <= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a <= b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a) <= b`},
						},
					},
				},
			},
			{
				Code:    `!a >= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a >= b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a) >= b`},
						},
					},
				},
			},

			// ===== Complex ordering cases =====
			// Ordering in ternary: x = !a < b ? c : d
			{
				Code:    `x = !a < b ? c : d`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 5,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `x = !(a < b) ? c : d`},
							{MessageId: "suggestParenthesisedNegation", Output: `x = (!a) < b ? c : d`},
						},
					},
				},
			},
			// Computed property with ordering: !a[0] >= b
			{
				Code:    `!a[0] >= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected", Line: 1, Column: 1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestNegatedExpression", Output: `!(a[0] >= b)`},
							{MessageId: "suggestParenthesisedNegation", Output: `(!a[0]) >= b`},
						},
					},
				},
			},
		},
	)
}
