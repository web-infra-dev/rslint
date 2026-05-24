// TestComputedPropertySpacingUpstream migrates the full valid/invalid suite
// from upstream packages/eslint-plugin/rules/computed-property-spacing/
// computed-property-spacing._js_.test.ts and computed-property-spacing._ts_.test.ts
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in
// computed_property_spacing_extras_test.go.
package computed_property_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/computed_property_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func optsAlways() []interface{} { return []interface{}{"always"} }
func optsNever() []interface{}  { return []interface{}{"never"} }
func optsAlwaysObj(o map[string]interface{}) []any {
	return []interface{}{"always", o}
}
func optsNeverObj(o map[string]interface{}) []any {
	return []interface{}{"never", o}
}

func TestComputedPropertySpacingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&computed_property_spacing.ComputedPropertySpacingRule,
		[]rule_tester.ValidTestCase{
			// ---- default - never ----
			{Code: `obj[foo]`},
			{Code: `obj['foo']`},
			{Code: `var x = {[b]: a}`},

			// ---- always ----
			{Code: `obj[ foo ]`, Options: optsAlways()},
			{Code: "obj[\nfoo\n]", Options: optsAlways()},
			{Code: `obj[ 'foo' ]`, Options: optsAlways()},
			{Code: `obj[ 'foo' + 'bar' ]`, Options: optsAlways()},
			{Code: `obj[ obj2[ foo ] ]`, Options: optsAlways()},
			{Code: "obj.map(function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsAlways()},
			{Code: "obj[ 'map' ](function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsAlways()},
			{Code: "obj[ 'for' + 'Each' ](function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsAlways()},
			{Code: `var foo = obj[ 1 ]`, Options: optsAlways()},
			{Code: `var foo = obj[ 'foo' ];`, Options: optsAlways()},
			{Code: `var foo = obj[ [1, 1] ];`, Options: optsAlways()},

			// ---- always - objectLiteralComputedProperties ----
			{Code: `var x = {[ "a" ]: a}`, Options: optsAlways()},
			{Code: `var y = {[ x ]: a}`, Options: optsAlways()},
			{Code: `var x = {[ "a" ]() {}}`, Options: optsAlways()},
			{Code: `var y = {[ x ]() {}}`, Options: optsAlways()},

			// ---- always - unrelated cases ----
			{Code: `var foo = {};`, Options: optsAlways()},
			{Code: `var foo = [];`, Options: optsAlways()},

			// ---- never ----
			{Code: `obj[foo]`, Options: optsNever()},
			{Code: `obj['foo']`, Options: optsNever()},
			{Code: `obj['foo' + 'bar']`, Options: optsNever()},
			{Code: `obj['foo'+'bar']`, Options: optsNever()},
			{Code: `obj[obj2[foo]]`, Options: optsNever()},
			{Code: "obj.map(function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsNever()},
			{Code: "obj['map'](function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsNever()},
			{Code: "obj['for' + 'Each'](function(item) { return [\n1,\n2,\n3,\n4\n]; })", Options: optsNever()},
			{Code: "obj[\nfoo]", Options: optsNever()},
			{Code: "obj[foo\n]", Options: optsNever()},
			{Code: `var foo = obj[1]`, Options: optsNever()},
			{Code: `var foo = obj['foo'];`, Options: optsNever()},
			{Code: `var foo = obj[[ 1, 1 ]];`, Options: optsNever()},

			// ---- never - objectLiteralComputedProperties ----
			{Code: `var x = {["a"]: a}`, Options: optsNever()},
			{Code: `var y = {[x]: a}`, Options: optsNever()},
			{Code: `var x = {["a"]() {}}`, Options: optsNever()},
			{Code: `var y = {[x]() {}}`, Options: optsNever()},

			// ---- never - unrelated cases ----
			{Code: `var foo = {};`, Options: optsNever()},
			{Code: `var foo = [];`, Options: optsNever()},

			// ---- Classes — enforceForClassMembers: false ----
			{Code: `class A { [ a ](){} }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": false})},
			{Code: `A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": false})},
			{Code: `A = class { [a](){} }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": false})},
			{Code: `class A { [a](){} get [b](){} set [b](foo){} static [c](){} static get [d](){} static set [d](bar){} }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": false})},
			{Code: `class A { [ a ]; }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": false})},
			{Code: `class A { [a]; }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": false})},

			// ---- Classes — valid spacing ----
			{Code: `A = class { [a](){} }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `class A { [a] ( ) { } }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: "A = class { [ \n a \n ](){} }", Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `class A { [a](){} get [b](){} set [b](foo){} static [c](){} static get [d](){} static set [d](bar){} }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `class A { [ a ](){} }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `class A { [ a ](){}[ b ](){} }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: "A = class { [\na\n](){} }", Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `A = class { [a]; static [a]; [a] = 0; static [a] = 0; }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `A = class { [ a ]; static [ a ]; [ a ] = 0; static [ a ] = 0; }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},

			// ---- Classes — non-computed ----
			{Code: `class A { a ( ) { } get b(){} set b ( foo ){} static c (){} static get d() {} static set d( bar ) {} }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `A = class {a(){}get b(){}set b(foo){}static c(){}static get d(){}static set d(bar){} }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `A = class { foo; #a; static #b; #c = 0; static #d = 0; }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `A = class { foo; #a; static #b; #c = 0; static #d = 0; }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true})},

			// ---- handling of parens and comments ----
			{Code: "const foo = {\n  [ (a) ]: 1\n}", Options: optsAlways()},
			{Code: "const foo = {\n  [ ( a ) ]: 1\n}", Options: optsAlways()},
			{Code: "const foo = {\n  [( a )]: 1\n}", Options: optsNever()},
			{Code: "const foo = {\n  [ /**/ a /**/ ]: 1\n}", Options: optsAlways()},
			{Code: "const foo = {\n  [/**/ a /**/]: 1\n}", Options: optsNever()},
			{Code: "const foo = {\n  [ a[ b ] ]: 1\n}", Options: optsAlways()},
			{Code: "const foo = {\n  [a[b]]: 1\n}", Options: optsNever()},
			{Code: "const foo = {\n  [ a[ /**/ b ]/**/ ]: 1\n}", Options: optsAlways()},
			{Code: "const foo = {\n  [/**/a[b /**/] /**/]: 1\n}", Options: optsNever()},

			// ---- Destructuring assignment ----
			{Code: `const { [a]: someProp } = obj;`, Options: optsNever()},
			{Code: `({ [a]: someProp } = obj);`, Options: optsNever()},
			{Code: `const { [ a ]: someProp } = obj;`, Options: optsAlways()},
			{Code: `({ [ a ]: someProp } = obj);`, Options: optsAlways()},

			// ---- TS only: accessor properties and indexed-access types ----
			{Code: `class A { accessor [ b ]; }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": false})},
			{Code: `class A { accessor [b]; }`, Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": false})},
			{Code: `A = class { accessor [b] = 1 }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `class A { accessor [b] = 1 }`, Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: "A = class {\n  accessor [\n    b\n  ] = 1\n}", Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true})},
			{Code: `type Foo = A[B]`},

			// ---- Regression: non-computed property is a no-op (eslint-stylistic#1053) ----
			{Code: `obj = { foo: bar }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `var foo = obj[ 1];`,
				Output:  []string{`var foo = obj[ 1 ];`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `var foo = obj[1 ];`,
				Output:  []string{`var foo = obj[ 1 ];`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `var foo = obj[ 1];`,
				Output:  []string{`var foo = obj[1];`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    `var foo = obj[1 ];`,
				Output:  []string{`var foo = obj[1];`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `obj[ foo ]`,
				Output:  []string{`obj[foo]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    `obj[foo ]`,
				Output:  []string{`obj[foo]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    `obj[ foo]`,
				Output:  []string{`obj[foo]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
				},
			},
			{
				Code:    `var foo = obj[1]`,
				Output:  []string{`var foo = obj[ 1 ]`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},

			// ---- multiple spaces ----
			{
				Code:    `obj[    foo]`,
				Output:  []string{`obj[foo]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    `obj[  foo  ]`,
				Output:  []string{`obj[foo]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 7},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 10, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    `obj[   foo ]`,
				Output:  []string{`obj[foo]`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    "obj[ foo + \n  bar   ]",
				Output:  []string{"obj[foo + \n  bar]"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "unexpectedSpaceBefore", Line: 2, Column: 6, EndLine: 2, EndColumn: 9},
				},
			},
			{
				Code:    "obj[\n foo  ]",
				Output:  []string{"obj[\n foo]"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 2, Column: 5, EndLine: 2, EndColumn: 7},
				},
			},

			// ---- always - objectLiteralComputedProperties ----
			{
				Code:    `var x = {[a]: b}`,
				Output:  []string{`var x = {[ a ]: b}`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `var x = {[a ]: b}`,
				Output:  []string{`var x = {[ a ]: b}`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    `var x = {[ a]: b}`,
				Output:  []string{`var x = {[ a ]: b}`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceBefore", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},

			// ---- never - objectLiteralComputedProperties ----
			{
				Code:    `var x = {[ a ]: b}`,
				Output:  []string{`var x = {[a]: b}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    `var x = {[a ]: b}`,
				Output:  []string{`var x = {[a]: b}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `var x = {[ a]: b}`,
				Output:  []string{`var x = {[a]: b}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    "var x = {[ a\n]: b}",
				Output:  []string{"var x = {[a\n]: b}"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},

			// ---- default settings for classes ----
			{
				Code:   `class A { [ a ](){} }`,
				Output: []string{`class A { [a](){} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `class A { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }`,
				Output:  []string{`class A { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 28, EndLine: 1, EndColumn: 29},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 40, EndLine: 1, EndColumn: 41},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 42, EndLine: 1, EndColumn: 43},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 60, EndLine: 1, EndColumn: 61},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 62, EndLine: 1, EndColumn: 63},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 81, EndLine: 1, EndColumn: 82},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 83, EndLine: 1, EndColumn: 84},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 102, EndLine: 1, EndColumn: 103},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 104, EndLine: 1, EndColumn: 105},
				},
			},
			{
				Code:    `A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }`,
				Output:  []string{`A = class { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }`},
				Options: []interface{}{"never", map[string]interface{}{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 28, EndLine: 1, EndColumn: 29},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 30, EndLine: 1, EndColumn: 31},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 42, EndLine: 1, EndColumn: 43},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 44, EndLine: 1, EndColumn: 45},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 62, EndLine: 1, EndColumn: 63},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 64, EndLine: 1, EndColumn: 65},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 83, EndLine: 1, EndColumn: 84},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 85, EndLine: 1, EndColumn: 86},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 104, EndLine: 1, EndColumn: 105},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 106, EndLine: 1, EndColumn: 107},
				},
			},
			{
				Code:    `A = class { [a](){} }`,
				Output:  []string{`A = class { [ a ](){} }`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    `A = class { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }`,
				Output:  []string{`A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 27, EndLine: 1, EndColumn: 28},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 37, EndLine: 1, EndColumn: 38},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 39, EndLine: 1, EndColumn: 40},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 55, EndLine: 1, EndColumn: 56},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 57, EndLine: 1, EndColumn: 58},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 74, EndLine: 1, EndColumn: 75},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 76, EndLine: 1, EndColumn: 77},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 93, EndLine: 1, EndColumn: 94},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 95, EndLine: 1, EndColumn: 96},
				},
			},
			{
				Code:    `class A { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }`,
				Output:  []string{`class A { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }`},
				Options: []interface{}{"always", map[string]interface{}{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 35, EndLine: 1, EndColumn: 36},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 37, EndLine: 1, EndColumn: 38},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 53, EndLine: 1, EndColumn: 54},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 55, EndLine: 1, EndColumn: 56},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 72, EndLine: 1, EndColumn: 73},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 74, EndLine: 1, EndColumn: 75},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 91, EndLine: 1, EndColumn: 92},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 93, EndLine: 1, EndColumn: 94},
				},
			},

			// ---- never - classes ----
			{
				Code:    `class A { [ a](){} }`,
				Output:  []string{`class A { [a](){} }`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `A = class { [a](){} b(){} static [c ](){} static [d](){}}`,
				Output:  []string{`A = class { [a](){} b(){} static [c](){} static [d](){}}`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    `class A { get [a ](){} set [ a](foo){} get b(){} static set b(bar){} static get [ a](){} static set [a ](baz){} }`,
				Output:  []string{`class A { get [a](){} set [a](foo){} get b(){} static set b(bar){} static get [a](){} static set [a](baz){} }`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 82, EndLine: 1, EndColumn: 83},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 103, EndLine: 1, EndColumn: 104},
				},
			},
			{
				Code:    `A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }`,
				Output:  []string{`A = class { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 28, EndLine: 1, EndColumn: 29},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 30, EndLine: 1, EndColumn: 31},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 42, EndLine: 1, EndColumn: 43},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 44, EndLine: 1, EndColumn: 45},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 62, EndLine: 1, EndColumn: 63},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 64, EndLine: 1, EndColumn: 65},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 83, EndLine: 1, EndColumn: 84},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 85, EndLine: 1, EndColumn: 86},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 104, EndLine: 1, EndColumn: 105},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 106, EndLine: 1, EndColumn: 107},
				},
			},
			{
				Code:    `class A { [ a]; [b ]; [ c ]; [ a] = 0; [b ] = 0; [ c ] = 0; }`,
				Output:  []string{`class A { [a]; [b]; [c]; [a] = 0; [b] = 0; [c] = 0; }`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 42, EndLine: 1, EndColumn: 43},
					{MessageId: "unexpectedSpaceAfter", Line: 1, Column: 51, EndLine: 1, EndColumn: 52},
					{MessageId: "unexpectedSpaceBefore", Line: 1, Column: 53, EndLine: 1, EndColumn: 54},
				},
			},

			// ---- always - classes ----
			{
				Code:    `class A { [ a](){} }`,
				Output:  []string{`class A { [ a ](){} }`},
				Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceBefore", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `A = class { [ a ](){} b(){} static [c ](){} static [ d ](){}}`,
				Output:  []string{`A = class { [ a ](){} b(){} static [ c ](){} static [ d ](){}}`},
				Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    `class A { get [a ](){} set [ a](foo){} get b(){} static set b(bar){} static get [ a](){} static set [a ](baz){} }`,
				Output:  []string{`class A { get [ a ](){} set [ a ](foo){} get b(){} static set b(bar){} static get [ a ](){} static set [ a ](baz){} }`},
				Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 84, EndLine: 1, EndColumn: 85},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 101, EndLine: 1, EndColumn: 102},
				},
			},
			{
				Code:    `A = class { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }`,
				Output:  []string{`A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }`},
				Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 27, EndLine: 1, EndColumn: 28},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 37, EndLine: 1, EndColumn: 38},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 39, EndLine: 1, EndColumn: 40},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 55, EndLine: 1, EndColumn: 56},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 57, EndLine: 1, EndColumn: 58},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 74, EndLine: 1, EndColumn: 75},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 76, EndLine: 1, EndColumn: 77},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 93, EndLine: 1, EndColumn: 94},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 95, EndLine: 1, EndColumn: 96},
				},
			},
			{
				Code:    `class A { [ a]; [b ]; [c]; [ a] = 0; [b ] = 0; [c] = 0; }`,
				Output:  []string{`class A { [ a ]; [ b ]; [ c ]; [ a ] = 0; [ b ] = 0; [ c ] = 0; }`},
				Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceBefore", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 38, EndLine: 1, EndColumn: 39},
					{MessageId: "missingSpaceAfter", Line: 1, Column: 48, EndLine: 1, EndColumn: 49},
					{MessageId: "missingSpaceBefore", Line: 1, Column: 50, EndLine: 1, EndColumn: 51},
				},
			},

			// ---- handling of parens and comments ----
			{
				Code:    "const foo = {\n  [(a)]: 1\n}",
				Output:  []string{"const foo = {\n  [ (a) ]: 1\n}"},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 2, Column: 3, EndLine: 2, EndColumn: 4},
					{MessageId: "missingSpaceBefore", Line: 2, Column: 7, EndLine: 2, EndColumn: 8},
				},
			},
			{
				Code:    "const foo = {\n  [( a )]: 1\n}",
				Output:  []string{"const foo = {\n  [ ( a ) ]: 1\n}"},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 2, Column: 3, EndLine: 2, EndColumn: 4},
					{MessageId: "missingSpaceBefore", Line: 2, Column: 9, EndLine: 2, EndColumn: 10},
				},
			},
			{
				Code:    "const foo = {\n  [ ( a ) ]: 1\n}",
				Output:  []string{"const foo = {\n  [( a )]: 1\n}"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 2, Column: 4, EndLine: 2, EndColumn: 5},
					{MessageId: "unexpectedSpaceBefore", Line: 2, Column: 10, EndLine: 2, EndColumn: 11},
				},
			},
			{
				Code:    "const foo = {\n  [/**/ a /**/]: 1\n}",
				Output:  []string{"const foo = {\n  [ /**/ a /**/ ]: 1\n}"},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 2, Column: 3, EndLine: 2, EndColumn: 4},
					{MessageId: "missingSpaceBefore", Line: 2, Column: 15, EndLine: 2, EndColumn: 16},
				},
			},
			{
				Code:    "const foo = {\n  [ /**/ a /**/ ]: 1\n}",
				Output:  []string{"const foo = {\n  [/**/ a /**/]: 1\n}"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 2, Column: 4, EndLine: 2, EndColumn: 5},
					{MessageId: "unexpectedSpaceBefore", Line: 2, Column: 16, EndLine: 2, EndColumn: 17},
				},
			},
			{
				Code:    "const foo = {\n  [a[b]]: 1\n}",
				Output:  []string{"const foo = {\n  [ a[ b ] ]: 1\n}"},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 2, Column: 3, EndLine: 2, EndColumn: 4},
					{MessageId: "missingSpaceAfter", Line: 2, Column: 5, EndLine: 2, EndColumn: 6},
					{MessageId: "missingSpaceBefore", Line: 2, Column: 7, EndLine: 2, EndColumn: 8},
					{MessageId: "missingSpaceBefore", Line: 2, Column: 8, EndLine: 2, EndColumn: 9},
				},
			},
			{
				Code:    "const foo = {\n  [ a[ b ] ]: 1\n}",
				Output:  []string{"const foo = {\n  [a[b]]: 1\n}"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 2, Column: 4, EndLine: 2, EndColumn: 5},
					{MessageId: "unexpectedSpaceAfter", Line: 2, Column: 7, EndLine: 2, EndColumn: 8},
					{MessageId: "unexpectedSpaceBefore", Line: 2, Column: 9, EndLine: 2, EndColumn: 10},
					{MessageId: "unexpectedSpaceBefore", Line: 2, Column: 11, EndLine: 2, EndColumn: 12},
				},
			},
			{
				Code:    "const foo = {\n  [a[/**/ b ]/**/]: 1\n}",
				Output:  []string{"const foo = {\n  [ a[ /**/ b ]/**/ ]: 1\n}"},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Line: 2, Column: 3, EndLine: 2, EndColumn: 4},
					{MessageId: "missingSpaceAfter", Line: 2, Column: 5, EndLine: 2, EndColumn: 6},
					{MessageId: "missingSpaceBefore", Line: 2, Column: 18, EndLine: 2, EndColumn: 19},
				},
			},
			{
				Code:    "const foo = {\n  [ /**/a[ b /**/ ] /**/]: 1\n}",
				Output:  []string{"const foo = {\n  [/**/a[b /**/] /**/]: 1\n}"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Line: 2, Column: 4, EndLine: 2, EndColumn: 5},
					{MessageId: "unexpectedSpaceAfter", Line: 2, Column: 11, EndLine: 2, EndColumn: 12},
					{MessageId: "unexpectedSpaceBefore", Line: 2, Column: 18, EndLine: 2, EndColumn: 19},
				},
			},

			// ---- Optional chaining ----
			{
				Code:    `obj?.[1];`,
				Output:  []string{`obj?.[ 1 ];`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter"},
					{MessageId: "missingSpaceBefore"},
				},
			},
			{
				Code:    `obj?.[ 1 ];`,
				Output:  []string{`obj?.[1];`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter"},
					{MessageId: "unexpectedSpaceBefore"},
				},
			},

			// ---- Destructuring Assignment ----
			{
				Code:    `const { [ a]: someProp } = obj;`,
				Output:  []string{`const { [a]: someProp } = obj;`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter"},
				},
			},
			{
				Code:    `const { [a ]: someProp } = obj;`,
				Output:  []string{`const { [a]: someProp } = obj;`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceBefore"},
				},
			},
			{
				Code:    `const { [ a ]: someProp } = obj;`,
				Output:  []string{`const { [a]: someProp } = obj;`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter"},
					{MessageId: "unexpectedSpaceBefore"},
				},
			},
			{
				Code:    `({ [ a ]: someProp } = obj);`,
				Output:  []string{`({ [a]: someProp } = obj);`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter"},
					{MessageId: "unexpectedSpaceBefore"},
				},
			},
			{
				Code:    `const { [a]: someProp } = obj;`,
				Output:  []string{`const { [ a ]: someProp } = obj;`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter"},
					{MessageId: "missingSpaceBefore"},
				},
			},
			{
				Code:    `({ [a]: someProp } = obj);`,
				Output:  []string{`({ [ a ]: someProp } = obj);`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter"},
					{MessageId: "missingSpaceBefore"},
				},
			},

			// ---- TS — AccessorProperty and IndexedAccessType ----
			{
				Code:    `class A { accessor [ a ] = 0 }`,
				Output:  []string{`class A { accessor [a] = 0 }`},
				Options: optsNeverObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter", Column: 21, EndColumn: 22},
					{MessageId: "unexpectedSpaceBefore", Column: 23, EndColumn: 24},
				},
			},
			{
				Code:    `class A { accessor [a] = 0 }`,
				Output:  []string{`class A { accessor [ a ] = 0 }`},
				Options: optsAlwaysObj(map[string]interface{}{"enforceForClassMembers": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter", Column: 20, EndColumn: 21},
					{MessageId: "missingSpaceBefore", Column: 22, EndColumn: 23},
				},
			},
			{
				Code:   `type Foo = A[ B ]`,
				Output: []string{`type Foo = A[B]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSpaceAfter"},
					{MessageId: "unexpectedSpaceBefore"},
				},
			},
			{
				Code:    `type Foo = A[B]`,
				Output:  []string{`type Foo = A[ B ]`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSpaceAfter"},
					{MessageId: "missingSpaceBefore"},
				},
			},
		},
	)
}
