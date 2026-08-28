package yoda

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestYodaUpstream migrates the full valid/invalid suite from upstream
// https://github.com/eslint/eslint/blob/v10.8.0/tests/lib/rules/yoda.js 1:1.
// Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in yoda_extras_test.go.
func TestYodaUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&YodaRule,
		[]rule_tester.ValidTestCase{
			// ---- "never" mode ----
			{Code: `if (value === "red") {}`, Options: "never"},
			{Code: `if (value === value) {}`, Options: "never"},
			{Code: `if (value != 5) {}`, Options: "never"},
			{Code: `if (5 & foo) {}`, Options: "never"},
			{Code: `if (5 === 4) {}`, Options: "never"},
			{Code: "if (value === `red`) {}", Options: "never"},
			{Code: "if (`red` === `red`) {}", Options: "never"},
			{Code: "if (`${foo}` === `red`) {}", Options: "never"},
			{Code: "if (`${\"\"}` === `red`) {}", Options: "never"},
			{Code: "if (`${\"red\"}` === foo) {}", Options: "never"},
			{Code: "if (b > `a` && b > `a`) {}", Options: "never"},
			{Code: "if (`b` > `a` && \"b\" > \"a\") {}", Options: "never"},
			{Code: `if ("blue" === value) {}`, Options: "always"},

			// ---- "always" mode ----
			{Code: `if (value === value) {}`, Options: "always"},
			{Code: `if (4 != value) {}`, Options: "always"},
			{Code: `if (foo & 4) {}`, Options: "always"},
			{Code: `if (5 === 4) {}`, Options: "always"},
			{Code: "if (`red` === value) {}", Options: "always"},
			{Code: "if (`red` === `red`) {}", Options: "always"},
			{Code: "if (`red` === `${foo}`) {}", Options: "always"},
			{Code: "if (`red` === `${\"\"}`) {}", Options: "always"},
			{Code: "if (foo === `${\"red\"}`) {}", Options: "always"},
			{Code: "if (`a` > b && `a` > b) {}", Options: "always"},
			{Code: "if (`b` > `a` && \"b\" > \"a\") {}", Options: "always"},
			{Code: `if ("a" < x && x < MAX ) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (1 < x && x < MAX ) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},

			// ---- Range exception ----
			{Code: `if ('a' < x && x < MAX ) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: "if (x < `x` || `x` <= x) {}", Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < x && x <= 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= x && x < 1) {}`, Options: []any{"always", map[string]any{"exceptRange": true}}},
			{Code: `if ('blue' < x.y && x.y < 'green') {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: "if (0 < x[``] && x[``] < 100) {}", Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: "if (0 < x[''] && x[``] < 100) {}", Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (a < 4 || (b[c[0]].d['e'] < 0 || 1 <= b[c[0]].d['e'])) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= x['y'] && x['y'] <= 100) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (a < 0 && (0 < b && b < 1)) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if ((0 < a && a < 1) && b < 0) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (-1 < x && x < 0) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= this.prop && this.prop <= 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= index && index < list.length) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (ZERO <= index && index < 100) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (value <= MIN || 10 < value) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (value <= 0 || MAX < value) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= a.b && a["b"] <= 100) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: "if (0 <= a.b && a[`b`] <= 100) {}", Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (-1n < x && x <= 1n) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (-1n <= x && x < 1n) {}`, Options: []any{"always", map[string]any{"exceptRange": true}}},
			{Code: "if (x < `1` || `1` < x) {}", Options: []any{"always", map[string]any{"exceptRange": true}}},
			{Code: `if (1 <= a['/(?<zero>0)/'] && a[/(?<zero>0)/] <= 100) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: "if (x <= `bar` || `foo` < x) {}", Options: []any{"always", map[string]any{"exceptRange": true}}},
			{Code: `if ('a' < x && x < MAX ) {}`, Options: []any{"always", map[string]any{"exceptRange": true}}},
			{Code: `if ('a' < x && x < MAX ) {}`, Options: "always"},
			{Code: `if (MIN < x && x < 'a' ) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (MIN < x && x < 'a' ) {}`, Options: "never"},
			{Code: "if (`blue` < x.y && x.y < `green`) {}", Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: "if (0 <= x[`y`] && x[`y`] <= 100) {}", Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: "if (0 <= x[`y`] && x[\"y\"] <= 100) {}", Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if ('a' <= x && x < 'b') {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (x < -1n || 1n <= x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (x < -1n || 1n <= x) {}`, Options: []any{"always", map[string]any{"exceptRange": true}}},
			{Code: `if (1 < a && a <= 2) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (x < -1 || 1 < x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (x <= 'bar' || 'foo' < x) {}`, Options: []any{"always", map[string]any{"exceptRange": true}}},
			{Code: `if (x < 0 || 1 <= x) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if('a' <= x && x < MAX) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 <= obj?.a && obj?.a < 1) {}`, Options: []any{"never", map[string]any{"exceptRange": true}}},
			{Code: `if (0 < x && x <= 1) {}`, Options: []any{"never", map[string]any{"onlyEquality": true}}},

			// ---- onlyEquality ----
			{Code: `if (x !== 'foo' && 'foo' !== x) {}`, Options: []any{"never", map[string]any{"onlyEquality": true}}},
			{Code: `if (x < 2 && x !== -3) {}`, Options: []any{"always", map[string]any{"onlyEquality": true}}},
			{Code: "if (x !== `foo` && `foo` !== x) {}", Options: []any{"never", map[string]any{"onlyEquality": true}}},
			{Code: "if (x < `2` && x !== `-3`) {}", Options: []any{"always", map[string]any{"onlyEquality": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- "never" mode: literal must move off the left side ----
			{
				Code:    `if (x <= 'foo' || 'bar' < x) {}`,
				Output:  []string{`if ('foo' >= x || 'bar' < x) {}`},
				Options: []any{"always", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if ("red" == value) {}`,
				Output:  []string{`if (value == "red") {}`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (true === value) {}`,
				Output:  []string{`if (value === true) {}`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (5 != value) {}`,
				Output:  []string{`if (value != 5) {}`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (5n != value) {}`,
				Output:  []string{`if (value != 5n) {}`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (null !== value) {}`,
				Output:  []string{`if (value !== null) {}`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if ("red" <= value) {}`,
				Output:  []string{`if (value >= "red") {}`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    "if (`red` <= value) {}",
				Output:  []string{"if (value >= `red`) {}"},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    "if (`red` <= `${foo}`) {}",
				Output:  []string{"if (`${foo}` >= `red`) {}"},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    "if (`red` <= `${\"red\"}`) {}",
				Output:  []string{"if (`${\"red\"}` >= `red`) {}"},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (true >= value) {}`,
				Output:  []string{`if (value <= true) {}`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `var foo = (5 < value) ? true : false`,
				Output:  []string{`var foo = (value > 5) ? true : false`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 12}},
			},
			{
				Code:    `function foo() { return (null > value); }`,
				Output:  []string{`function foo() { return (value < null); }`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 26}},
			},
			{
				Code:    `if (-1 < str.indexOf(substr)) {}`,
				Output:  []string{`if (str.indexOf(substr) > -1) {}`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- "always" mode: literal must move onto the left side ----
			{
				Code:    `if (value == "red") {}`,
				Output:  []string{`if ("red" == value) {}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    "if (value == `red`) {}",
				Output:  []string{"if (`red` == value) {}"},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (value === true) {}`,
				Output:  []string{`if (true === value) {}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (value === 5n) {}`,
				Output:  []string{`if (5n === value) {}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    "if (`${\"red\"}` <= `red`) {}",
				Output:  []string{"if (`red` >= `${\"red\"}`) {}"},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- exceptRange invalid: out-of-order, non-matching-reference, or descending bounds are not a valid range test ----
			{
				Code:    `if (a < 0 && 0 <= b && b < 1) {}`,
				Output:  []string{`if (a < 0 && b >= 0 && b < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 14}},
			},
			{
				Code:    `if (0 <= a && a < 1 && b < 1) {}`,
				Output:  []string{`if (a >= 0 && a < 1 && b < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (1 < a && a < 0) {}`,
				Output:  []string{`if (a > 1 && a < 0) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `0 < a && a < 1`,
				Output:  []string{`a > 0 && a < 1`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `var a = b < 0 || 1 <= b;`,
				Output:  []string{`var a = b < 0 || b >= 1;`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 18}},
			},
			{
				Code:    `if (0 <= x && x < -1) {}`,
				Output:  []string{`if (x >= 0 && x < -1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `var a = (b < 0 && 0 <= b);`,
				Output:  []string{`var a = (0 > b && 0 <= b);`},
				Options: []any{"always", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 10}},
			},
			{
				Code:    "var a = (b < `0` && `0` <= b);",
				Output:  []string{"var a = (`0` > b && `0` <= b);"},
				Options: []any{"always", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 10}},
			},
			{
				Code:    "if (`green` < x.y && x.y < `blue`) {}",
				Output:  []string{"if (x.y > `green` && x.y < `blue`) {}"},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (0 <= a[b] && a['b'] < 1) {}`,
				Output:  []string{`if (a[b] >= 0 && a['b'] < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    "if (0 <= a[b] && a[`b`] < 1) {}",
				Output:  []string{"if (a[b] >= 0 && a[`b`] < 1) {}"},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    "if (`0` <= a[b] && a[`b`] < `1`) {}",
				Output:  []string{"if (a[b] >= `0` && a[`b`] < `1`) {}"},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (0 <= a[b] && a.b < 1) {}`,
				Output:  []string{`if (a[b] >= 0 && a.b < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (0 <= a[''] && a.b < 1) {}`,
				Output:  []string{`if (a[''] >= 0 && a.b < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (0 <= a[''] && a[' '] < 1) {}`,
				Output:  []string{`if (a[''] >= 0 && a[' '] < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (0 <= a[''] && a[null] < 1) {}`,
				Output:  []string{`if (a[''] >= 0 && a[null] < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    "if (0 <= a[``] && a[null] < 1) {}",
				Output:  []string{"if (a[``] >= 0 && a[null] < 1) {}"},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (0 <= a[''] && a[b] < 1) {}`,
				Output:  []string{`if (a[''] >= 0 && a[b] < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (0 <= a[''] && a[b()] < 1) {}`,
				Output:  []string{`if (a[''] >= 0 && a[b()] < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    "if (0 <= a[``] && a[b()] < 1) {}",
				Output:  []string{"if (a[``] >= 0 && a[b()] < 1) {}"},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (0 <= a[b()] && a[b()] < 1) {}`,
				Output:  []string{`if (a[b()] >= 0 && a[b()] < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (0 <= a.null && a[/(?<zero>0)/] <= 1) {}`,
				Output:  []string{`if (a.null >= 0 && a[/(?<zero>0)/] <= 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- onlyEquality invalid ----
			{
				Code:    `if (3 == a) {}`,
				Output:  []string{`if (a == 3) {}`},
				Options: []any{"never", map[string]any{"onlyEquality": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `foo(3 === a);`,
				Output:  []string{`foo(a === 3);`},
				Options: []any{"never", map[string]any{"onlyEquality": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `foo(a === 3);`,
				Output:  []string{`foo(3 === a);`},
				Options: []any{"always", map[string]any{"onlyEquality": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    "foo(a === `3`);",
				Output:  []string{"foo(`3` === a);"},
				Options: []any{"always", map[string]any{"onlyEquality": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- default options (bare "never") ----
			{
				Code:   `if (0 <= x && x < 1) {}`,
				Output: []string{`if (x >= 0 && x < 1) {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// ---- comments around the operator are preserved by the fix ----
			{
				Code:    `if ( /* a */ 0 /* b */ < /* c */ foo /* d */ ) {}`,
				Output:  []string{`if ( /* a */ foo /* b */ > /* c */ 0 /* d */ ) {}`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 14}},
			},
			{
				Code:    `if ( /* a */ foo /* b */ > /* c */ 0 /* d */ ) {}`,
				Output:  []string{`if ( /* a */ 0 /* b */ < /* c */ foo /* d */ ) {}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 14}},
			},

			// ---- extra whitespace around the operator is preserved by the fix ----
			{
				Code:    `if (foo()===1) {}`,
				Output:  []string{`if (1===foo()) {}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if (foo()     === 1) {}`,
				Output:  []string{`if (1     === foo()) {}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// https://github.com/eslint/eslint/issues/7326
			{
				Code:    `while (0 === (a));`,
				Output:  []string{`while ((a) === 0);`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 8}},
			},
			{
				Code:    `while (0 === (a = b));`,
				Output:  []string{`while ((a = b) === 0);`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 8}},
			},
			{
				Code:    `while ((a) === 0);`,
				Output:  []string{`while (0 === (a));`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 8}},
			},
			{
				Code:    `while ((a = b) === 0);`,
				Output:  []string{`while (0 === (a = b));`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 8}},
			},

			// ---- deeply nested parentheses ----
			{
				Code:    `if (((((((((((foo)))))))))) === ((((((5)))))));`,
				Output:  []string{`if (((((((5)))))) === ((((((((((foo)))))))))));`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},

			// Adjacent tokens tests
			{
				Code:    `function *foo() { yield(1) < a }`,
				Output:  []string{`function *foo() { yield a > (1) }`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 24}},
			},
			{
				Code:    `function *foo() { yield(1) < \u0061 }`,
				Output:  []string{`function *foo() { yield \u0061 > (1) }`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 24}},
			},
			{
				Code:    `function foo() { return(1) < \u0061 }`,
				Output:  []string{`function foo() { return \u0061 > (1) }`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 24}},
			},
			{
				Code:    `function foo() { throw(1) < \u0061 }`,
				Output:  []string{`function foo() { throw \u0061 > (1) }`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 23}},
			},
			{
				Code:    `function *foo() { yield((1)) < a }`,
				Output:  []string{`function *foo() { yield a > ((1)) }`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 24}},
			},
			{
				Code:    `function *foo() { yield 1 < a }`,
				Output:  []string{`function *foo() { yield a > 1 }`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 25}},
			},
			{
				Code:    `function *foo() { yield/**/1 < a }`,
				Output:  []string{`function *foo() { yield/**/a > 1 }`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 28}},
			},
			{
				Code:    `function *foo() { yield(1) < ++a }`,
				Output:  []string{`function *foo() { yield++a > (1) }`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 24}},
			},
			{
				Code:    `function *foo() { yield(1) < (a) }`,
				Output:  []string{`function *foo() { yield(a) > (1) }`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 24}},
			},
			{
				Code:    `x=1 < a`,
				Output:  []string{`x=a > 1`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 3}},
			},
			{
				Code:    `function *foo() { yield++a < 1 }`,
				Output:  []string{`function *foo() { yield 1 > ++a }`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 24}},
			},
			{
				Code:    `function *foo() { yield(a) < 1 }`,
				Output:  []string{`function *foo() { yield 1 > (a) }`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 24}},
			},
			{
				Code:    `function *foo() { yield a < 1 }`,
				Output:  []string{`function *foo() { yield 1 > a }`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 25}},
			},
			{
				Code:    `function *foo() { yield/**/a < 1 }`,
				Output:  []string{`function *foo() { yield/**/1 > a }`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 28}},
			},
			{
				Code:    `function *foo() { yield++a < (1) }`,
				Output:  []string{`function *foo() { yield(1) > ++a }`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 24}},
			},
			{
				Code:    `x=a < 1`,
				Output:  []string{`x=1 > a`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 3}},
			},

			// ---- adjacency against `in` / `instanceof` on the outer expression ----
			{
				Code:   `0 < f()in obj`,
				Output: []string{`f() > 0 in obj`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `1 > x++instanceof foo`,
				Output:  []string{`x++ < 1 instanceof foo`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `x < ('foo')in bar`,
				Output:  []string{`('foo') > x in bar`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `false <= ((x))in foo`,
				Output:  []string{`((x)) >= false in foo`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `x >= (1)instanceof foo`,
				Output:  []string{`(1) <= x instanceof foo`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `false <= ((x)) in foo`,
				Output:  []string{`((x)) >= false in foo`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `x >= 1 instanceof foo`,
				Output:  []string{`1 <= x instanceof foo`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `x >= 1/**/instanceof foo`,
				Output:  []string{`1 <= x/**/instanceof foo`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `(x >= 1)instanceof foo`,
				Output:  []string{`(1 <= x)instanceof foo`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 2}},
			},
			{
				Code:    `(x) >= (1)instanceof foo`,
				Output:  []string{`(1) <= (x)instanceof foo`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `1 > x===foo`,
				Output:  []string{`x < 1===foo`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},
			{
				Code:    `1 > x`,
				Output:  []string{`x < 1`},
				Options: "never",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 1}},
			},

			// ---- exceptRange: descending bounds fall back to a normal (non-exempt) report ----
			{
				Code:    "if (`green` < x.y && x.y < `blue`) {}",
				Output:  []string{"if (`green` < x.y && `blue` > x.y) {}"},
				Options: []any{"always", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 22}},
			},
			{
				Code:    `if('a' <= x && x < 'b') {}`,
				Output:  []string{`if('a' <= x && 'b' > x) {}`},
				Options: "always",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 16}},
			},
			{
				Code:    `if ('b' <= x && x < 'a') {}`,
				Output:  []string{`if (x >= 'b' && x < 'a') {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
			{
				Code:    `if('a' <= x && x < 1) {}`,
				Output:  []string{`if(x >= 'a' && x < 1) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 4}},
			},
			{
				Code:    `if (0 < a && b < max) {}`,
				Output:  []string{`if (a > 0 && b < max) {}`},
				Options: []any{"never", map[string]any{"exceptRange": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "expected", Line: 1, Column: 5}},
			},
		},
	)
}
