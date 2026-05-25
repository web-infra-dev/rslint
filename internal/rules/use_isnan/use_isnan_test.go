package use_isnan

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestUseIsNaNRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&UseIsNaNRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// ── Non-comparison usage ──
			{Code: `var x = NaN;`},
			{Code: `isNaN(NaN) === true;`},
			{Code: `Number.isNaN(NaN) === true;`},
			{Code: `isNaN(123);`},
			{Code: `Number.isNaN(123);`},

			// ── Arithmetic operators (not comparison) ──
			{Code: `NaN + 1;`},
			{Code: `1 + NaN;`},
			{Code: `NaN - 1;`},
			{Code: `NaN * 2;`},
			{Code: `2 / NaN;`},
			{Code: `Number.NaN + 1;`},

			// ── Assignment (not comparison) ──
			{Code: `var q; if (q = NaN) {}`},

			// ── Lookalike identifiers that are NOT NaN ──
			{Code: `x === Nan;`},
			{Code: `x === nan;`},
			{Code: `x === NAN;`},
			{Code: `x === Number.Nan;`},
			{Code: `x === window.NaN;`},
			{Code: `x === globalThis.NaN;`},
			{Code: `x === Math.NaN;`},
			{Code: `x === Number[NaN];`},          // identifier key, not string
			{Code: `x === Number.NaN.toString();`}, // method on NaN value

			// ── Sequence expression: NaN NOT last ──
			{Code: `x === (NaN, 1);`},
			{Code: `x === (NaN, a);`},
			{Code: `x === (a, NaN, 1);`},
			{Code: `x === (Number.NaN, 1);`},

			// ── Nested comma: only ONE level resolved (matches ESLint) ──
			{Code: `x === (a, (b, NaN));`},         // nested comma → not resolved
			{Code: `x === (a, (b, Number.NaN));`},

			// ── switch: enforceForSwitchCase: false ──
			{
				Code:    `switch(NaN) { case foo: break; }`,
				Options: map[string]interface{}{"enforceForSwitchCase": false},
			},
			{
				Code:    `switch(foo) { case NaN: break; }`,
				Options: map[string]interface{}{"enforceForSwitchCase": false},
			},

			// ── switch: valid discriminant ──
			{Code: `switch(foo) { case bar: break; }`},
			{Code: `switch(true) { case true: break; }`},
			{Code: `switch(Nan) {}`},
			{Code: `switch('NaN') {}`},
			{Code: `switch(foo(NaN)) {}`},
			{Code: `switch(foo.NaN) {}`},
			{Code: `switch((NaN, 1)) {}`},
			{Code: `switch((Number.NaN, 1)) {}`},

			// ── switch: valid case clause ──
			{Code: `switch(foo) { case Nan: break; }`},
			{Code: `switch(foo) { case 'NaN': break; }`},
			{Code: `switch(foo) { case foo(NaN): break; }`},
			{Code: `switch(foo) { case foo.NaN: break; }`},
			{Code: `switch(foo) { case bar: NaN; }`},
			{Code: `switch(foo) { default: NaN; }`},
			{Code: `switch(foo) { case (NaN, 1): break; }`},

			// ── indexOf: default enforceForIndexOf=false ──
			{Code: `foo.indexOf(NaN)`},
			{Code: `foo.lastIndexOf(NaN)`},

			// ── indexOf: enforceForIndexOf=true, valid ──
			{
				Code:    `foo.indexOf(bar)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `foo.lastIndexOf(bar)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `foo.indexOf(NaN, 0, extra)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `foo.lastIndexOf(NaN, 0, extra)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `foo.indexOf(a, NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `foo.indexOf()`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `foo.indexOf(Nan)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `foo.indexOf((NaN, 1))`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `foo.indexOf((Number.NaN, 1))`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},

			// ── indexOf: nested comma is NOT resolved (matches ESLint) ──
			{
				Code:    `foo.indexOf((a, (b, NaN)))`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},

			// ── indexOf: not a method call ──
			{
				Code:    `indexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `lastIndexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},

			// ── indexOf: case-sensitive / wrong method ──
			{
				Code:    `foo.IndexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `foo.bar(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},

			// ── indexOf: computed with identifier (not string literal) ──
			{
				Code:    `foo[indexOf](NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},

			// ── indexOf: new expression ──
			{
				Code:    `new foo.indexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},

			// ── indexOf: indirect call ──
			{
				Code:    `foo.indexOf.call(arr, NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// ════════════════════════════════════════
			// DIMENSION 1: All comparison operators × NaN
			// ════════════════════════════════════════
			{
				Code:   `123 == NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `123 === NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `NaN === "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `NaN == "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `123 != NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `123 !== NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `NaN < "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `"abc" < NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `NaN > "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `"abc" >= NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `NaN <= "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},

			// ════════════════════════════════════════
			// DIMENSION 2: All NaN representations
			// ════════════════════════════════════════
			{
				Code:   `123 === Number.NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `Number.NaN === "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `x === Number['NaN'];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `x !== Number['NaN'];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `x === Number?.NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `x === Number?.['NaN'];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			// template literal in bracket notation
			{
				Code:   "x === Number[`NaN`];",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},

			// ════════════════════════════════════════
			// DIMENSION 3: Sequence expressions (comma)
			// ════════════════════════════════════════
			{
				Code:   `x = (foo, NaN) === 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 5}},
			},
			{
				Code:   `x = 1 === (foo, NaN);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 5}},
			},
			{
				Code:   `x = (a, b, NaN) === 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 5}},
			},
			{
				Code:   `x = (a, Number.NaN) === 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 5}},
			},
			{
				Code:   `x = (a, Number['NaN']) === 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 5}},
			},
			{
				Code:   `x = (1, 2) === NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 5}},
			},
			// ── Parens inside comma: NaN wrapped in parens as last comma element ──
			{
				Code:   `x === (a, (NaN));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `x === (a, ((NaN)));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `x === (a, (Number.NaN));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},

			{
				Code:   `x === (doStuff(), NaN);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},

			// ════════════════════════════════════════
			// DIMENSION 4: Parenthesized depth
			// ════════════════════════════════════════
			{
				Code:   `x === (NaN);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `x === ((NaN));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `((x)) === ((NaN));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `x === ((a, NaN));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},

			// ════════════════════════════════════════
			// DIMENSION 5: NaN both sides
			// ════════════════════════════════════════
			{
				Code:   `NaN === NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `NaN == Number.NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 1}},
			},

			// ════════════════════════════════════════
			// DIMENSION 6: Switch discriminant variations
			// ════════════════════════════════════════
			{
				Code:   `switch(NaN) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "switchNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `switch(Number.NaN) { case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "switchNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `switch(Number['NaN']) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "switchNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `switch(Number?.NaN) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "switchNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `switch((a, NaN)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "switchNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `switch((a, b, NaN)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "switchNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `switch((NaN)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "switchNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `switch(((NaN))) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "switchNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `switch((doStuff(), NaN)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "switchNaN", Line: 1, Column: 1}},
			},
			{
				Code:   `switch((doStuff(), Number.NaN)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "switchNaN", Line: 1, Column: 1}},
			},

			// ════════════════════════════════════════
			// DIMENSION 7: Case clause variations
			// ════════════════════════════════════════
			{
				Code:   `switch(foo) { case NaN: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caseNaN", Line: 1, Column: 15}},
			},
			{
				Code:   `switch(foo) { case Number.NaN: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caseNaN", Line: 1, Column: 15}},
			},
			{
				Code:   `switch(foo) { case Number['NaN']: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caseNaN", Line: 1, Column: 15}},
			},
			{
				Code:   `switch(foo) { case Number?.NaN: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caseNaN", Line: 1, Column: 15}},
			},
			{
				Code:   `switch(foo) { case (NaN): break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caseNaN", Line: 1, Column: 15}},
			},
			{
				Code:   `switch(foo) { case ((NaN)): break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caseNaN", Line: 1, Column: 15}},
			},
			{
				Code:   `switch(foo) { case (a, NaN): break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caseNaN", Line: 1, Column: 15}},
			},
			{
				Code:   `switch(foo) { case (a, Number.NaN): break; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caseNaN", Line: 1, Column: 15}},
			},

			// ── Multiple NaN cases ──
			{
				Code: `switch(foo) { case NaN: case Number.NaN: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "caseNaN", Line: 1, Column: 15},
					{MessageId: "caseNaN", Line: 1, Column: 25},
				},
			},
			{
				Code: `switch(foo) { case NaN: break; case Number['NaN']: break; case bar: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "caseNaN", Line: 1, Column: 15},
					{MessageId: "caseNaN", Line: 1, Column: 32},
				},
			},

			// ── switch(NaN) + case NaN = 2 errors ──
			{
				Code: `switch(NaN) { case NaN: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchNaN", Line: 1, Column: 1},
					{MessageId: "caseNaN", Line: 1, Column: 15},
				},
			},
			{
				Code: `switch(Number.NaN) { case Number.NaN: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchNaN", Line: 1, Column: 1},
					{MessageId: "caseNaN", Line: 1, Column: 22},
				},
			},

			// ── Nested switch ──
			{
				Code: `switch(foo) { case bar: switch(NaN) { case NaN: break; } break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchNaN", Line: 1, Column: 25},
					{MessageId: "caseNaN", Line: 1, Column: 39},
				},
			},

			// ════════════════════════════════════════
			// DIMENSION 8: indexOf callee variations
			// ════════════════════════════════════════
			{
				Code:    `foo.indexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.lastIndexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo['indexOf'](NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo['lastIndexOf'](NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    "foo[`indexOf`](NaN)",
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    "foo[`lastIndexOf`](NaN)",
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo?.indexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo?.lastIndexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.indexOf?.(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo?.indexOf?.(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `(foo?.indexOf)(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `((foo.indexOf))(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo().indexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.bar.indexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.bar.baz.lastIndexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `(a || b).indexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},

			// ════════════════════════════════════════
			// DIMENSION 9: indexOf arg variations
			// ════════════════════════════════════════
			{
				Code:    `foo.indexOf(Number.NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.indexOf(Number['NaN'])`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.indexOf(Number?.NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.indexOf((a, NaN))`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.indexOf((a, Number.NaN))`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.indexOf((NaN))`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.indexOf(NaN, 1)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.indexOf(NaN, b)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.lastIndexOf(NaN, NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.indexOf(Number.NaN, 1)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},
			{
				Code:    `foo.lastIndexOf(Number.NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "indexOfNaN", Line: 1, Column: 1}},
			},

			// ════════════════════════════════════════
			// DIMENSION 10: NaN in nested expression contexts
			// ════════════════════════════════════════
			{
				Code:   `if (NaN === x) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 5}},
			},
			{
				Code:   `while (x !== NaN) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 8}},
			},
			{
				Code:   `for (; NaN < x;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 8}},
			},
			{
				Code:   `var t = x === NaN ? 'yes' : 'no';`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 9}},
			},
			{
				Code:   `var u = [NaN === 1];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 10}},
			},
			{
				Code:   `var v = { key: NaN === 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 16}},
			},
			{
				Code:   `function f() { return NaN === x; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 23}},
			},
			{
				Code:   `var g = () => NaN === x;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 15}},
			},
			{
				Code:   `var h = function() { return x === NaN; };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 29}},
			},
			{
				Code:   `console.log(NaN === 1);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 13}},
			},
			{
				Code:   `void (NaN === x);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 7}},
			},
			{
				Code:   `!(NaN === x);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "comparisonWithNaN", Line: 1, Column: 3}},
			},
		},
	)
}
