// TestCommaSpacingUpstream migrates the full valid/invalid suite from upstream
// packages/eslint-plugin/rules/comma-spacing/comma-spacing._js_.test.ts and
// comma-spacing._ts_.test.ts 1:1. Position assertions cover line/column for
// every invalid case. rslint-specific lock-in cases live in
// comma_spacing_extras_test.go.
package comma_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/comma_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Option shortcuts — match upstream's `{ before, after }` shape exactly.
// We feed bare maps (the CLI-facing single-option shape, which goes through
// `utils.GetOptionsMap`'s map branch). Mixing in some array-wrapped cases
// in the extras file exercises the other branch.
func optsNeither() map[string]interface{} {
	return map[string]interface{}{"before": false, "after": false}
}
func optsAfter() map[string]interface{} {
	return map[string]interface{}{"before": false, "after": true}
}
func optsBefore() map[string]interface{} {
	return map[string]interface{}{"before": true, "after": false}
}
func optsBoth() map[string]interface{} {
	return map[string]interface{}{"before": true, "after": true}
}

func TestCommaSpacingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&comma_spacing.CommaSpacingRule,
		[]rule_tester.ValidTestCase{
			// ---- _js_ valid: comment-around-comma shapes (default) ----
			{Code: `myfunc(404, true/* bla bla bla */, 'hello');`},
			{Code: `myfunc(404, true /* bla bla bla */, 'hello');`},
			{Code: `myfunc(404, true/* bla bla bla *//* hi */, 'hello');`},
			{Code: `myfunc(404, true/* bla bla bla */ /* hi */, 'hello');`},
			{Code: `myfunc(404, true, /* bla bla bla */ 'hello');`},
			{Code: "myfunc(404, // comment\n true, /* bla bla bla */ 'hello');"},
			{Code: "myfunc(404, // comment\n true,/* bla bla bla */ 'hello');", Options: optsNeither()},

			// ---- _js_ valid: arrays and array holes (default) ----
			{Code: `var a = 1, b = 2;`},
			{Code: `var arr = [,];`},
			{Code: `var arr = [, ];`},
			{Code: `var arr = [ ,];`},
			{Code: `var arr = [ , ];`},
			{Code: `var arr = [1,];`},
			{Code: `var arr = [1, ];`},
			{Code: `var arr = [, 2];`},
			{Code: `var arr = [ , 2];`},
			{Code: `var arr = [1, 2];`},
			{Code: `var arr = [,,];`},
			{Code: `var arr = [ ,,];`},
			{Code: `var arr = [, ,];`},
			{Code: `var arr = [,, ];`},
			{Code: `var arr = [ , ,];`},
			{Code: `var arr = [ ,, ];`},
			{Code: `var arr = [, , ];`},
			{Code: `var arr = [ , , ];`},
			{Code: `var arr = [1, , ];`},
			{Code: `var arr = [, 2, ];`},
			{Code: `var arr = [, , 3];`},
			{Code: `var arr = [,, 3];`},
			{Code: `var arr = [1, 2, ];`},
			{Code: `var arr = [, 2, 3];`},
			{Code: `var arr = [1, , 3];`},
			{Code: `var arr = [1, 2, 3];`},
			{Code: `var arr = [1, 2, 3,];`},
			{Code: `var arr = [1, 2, 3, ];`},

			// ---- _js_ valid: objects (default) ----
			{Code: `var obj = {'foo':'bar', 'baz':'qur'};`},
			{Code: `var obj = {'foo':'bar', 'baz':'qur', };`},
			{Code: `var obj = {'foo':'bar', 'baz':'qur',};`},
			{Code: "var obj = {'foo':'bar', 'baz':\n'qur'};"},
			{Code: "var obj = {'foo':\n'bar', 'baz':\n'qur'};"},

			// ---- _js_ valid: functions / arrows (default) ----
			{Code: `function foo(a, b){}`},
			{Code: `function foo(a, b = 1){}`},
			{Code: `function foo(a = 1, b, c){}`},
			{Code: `var foo = (a, b) => {}`},
			{Code: `var foo = (a=1, b) => {}`},
			{Code: `var foo = a => a + 2`},

			// ---- _js_ valid: sequence expressions and parens (default) ----
			{Code: `a, b`},
			{Code: `var a = (1 + 2, 2);`},
			{Code: `a(b, c)`},
			{Code: `new A(b, c)`},
			{Code: `foo((a), b)`},
			{Code: `var b = ((1 + 2), 2);`},
			{Code: `parseInt((a + b), 10)`},
			{Code: `go.boom((a + b), 10)`},
			{Code: `go.boom((a + b), 10, (4))`},
			{Code: `var x = [ (a + c), (b + b) ]`},
			{Code: `['  ,  ']`},
			{Code: "[`  ,  `]"},
			{Code: "`${[1, 2]}`"},
			{Code: `fn(a, b,)`},                  // #11295
			{Code: `const fn = (a, b,) => {}`},   // #11295
			{Code: `const fn = function (a, b,) {}`}, // #11295
			{Code: `function fn(a, b,) {}`},      // #11295
			{Code: `foo(/,/, 'a')`},
			{Code: `var x = ',,,,,';`},
			{Code: `var code = 'var foo = 1, bar = 3;'`},
			{Code: "['apples', \n 'oranges'];"},
			{Code: `{x: 'var x,y,z'}`},

			// ---- _js_ valid: { before:true, after:false } ----
			{Code: "var obj = {'foo':\n'bar' ,'baz':\n'qur'};", Options: optsBefore()},
			{Code: `var a = 1 ,b = 2;`, Options: optsBefore()},
			{Code: `function foo(a ,b){}`, Options: optsBefore()},
			{Code: `var arr = [,];`, Options: optsBefore()},
			{Code: `var arr = [ ,];`, Options: optsBefore()},
			{Code: `var arr = [, ];`, Options: optsBefore()},
			{Code: `var arr = [ , ];`, Options: optsBefore()},
			{Code: `var arr = [1 ,];`, Options: optsBefore()},
			{Code: `var arr = [1 , ];`, Options: optsBefore()},
			{Code: `var arr = [ ,2];`, Options: optsBefore()},
			{Code: `var arr = [1 ,2];`, Options: optsBefore()},
			{Code: `var arr = [,,];`, Options: optsBefore()},
			{Code: `var arr = [ ,,];`, Options: optsBefore()},
			{Code: `var arr = [, ,];`, Options: optsBefore()},
			{Code: `var arr = [,, ];`, Options: optsBefore()},
			{Code: `var arr = [ , ,];`, Options: optsBefore()},
			{Code: `var arr = [ ,, ];`, Options: optsBefore()},
			{Code: `var arr = [, , ];`, Options: optsBefore()},
			{Code: `var arr = [ , , ];`, Options: optsBefore()},
			{Code: `var arr = [1 , ,];`, Options: optsBefore()},
			{Code: `var arr = [ ,2 ,];`, Options: optsBefore()},
			{Code: `var arr = [,2 , ];`, Options: optsBefore()},
			{Code: `var arr = [ , ,3];`, Options: optsBefore()},
			{Code: `var arr = [1 ,2 ,];`, Options: optsBefore()},
			{Code: `var arr = [ ,2 ,3];`, Options: optsBefore()},
			{Code: `var arr = [1 , ,3];`, Options: optsBefore()},
			{Code: `var arr = [1 ,2 ,3];`, Options: optsBefore()},

			// ---- _js_ valid: { before:true, after:true } ----
			{Code: `var obj = {'foo':'bar' , 'baz':'qur'};`, Options: optsBoth()},
			{Code: `var obj = {'foo':'bar' ,'baz':'qur' , };`, Options: optsBefore()},
			{Code: `var a = 1 , b = 2;`, Options: optsBoth()},
			{Code: `var arr = [, ];`, Options: optsBoth()},
			{Code: `var arr = [,,];`, Options: optsBoth()},
			{Code: `var arr = [1 , ];`, Options: optsBoth()},
			{Code: `var arr = [ , 2];`, Options: optsBoth()},
			{Code: `var arr = [1 , 2];`, Options: optsBoth()},
			{Code: `var arr = [, , ];`, Options: optsBoth()},
			{Code: `var arr = [1 , , ];`, Options: optsBoth()},
			{Code: `var arr = [ , 2 , ];`, Options: optsBoth()},
			{Code: `var arr = [ , , 3];`, Options: optsBoth()},
			{Code: `var arr = [1 , 2 , ];`, Options: optsBoth()},
			{Code: `var arr = [, 2 , 3];`, Options: optsBoth()},
			{Code: `var arr = [1 , , 3];`, Options: optsBoth()},
			{Code: `var arr = [1 , 2 , 3];`, Options: optsBoth()},
			{Code: `a , b`, Options: optsBoth()},

			// ---- _js_ valid: { before:false, after:false } ----
			{Code: `var arr = [,];`, Options: optsNeither()},
			{Code: `var arr = [ ,];`, Options: optsNeither()},
			{Code: `var arr = [1,];`, Options: optsNeither()},
			{Code: `var arr = [,2];`, Options: optsNeither()},
			{Code: `var arr = [ ,2];`, Options: optsNeither()},
			{Code: `var arr = [1,2];`, Options: optsNeither()},
			{Code: `var arr = [,,];`, Options: optsNeither()},
			{Code: `var arr = [ , , ];`, Options: optsNeither()},
			{Code: `var arr = [ ,,];`, Options: optsNeither()},
			{Code: `var arr = [1,,];`, Options: optsNeither()},
			{Code: `var arr = [,2,];`, Options: optsNeither()},
			{Code: `var arr = [ ,2,];`, Options: optsNeither()},
			{Code: `var arr = [,,3];`, Options: optsNeither()},
			{Code: `var arr = [1,2,];`, Options: optsNeither()},
			{Code: `var arr = [,2,3];`, Options: optsNeither()},
			{Code: `var arr = [1,,3];`, Options: optsNeither()},
			{Code: `var arr = [1,2,3];`, Options: optsNeither()},
			{Code: `var a = (1 + 2,2)`, Options: optsNeither()},

			// ---- _js_ valid: ES6 features and templates (default) ----
			{Code: "var a; console.log(`${a}`, \"a\");"},

			// ---- _js_ valid: destructuring (default) ----
			{Code: `var [a, b] = [1, 2];`},
			{Code: `var [a, b, ] = [1, 2];`},
			{Code: `var [a, b,] = [1, 2];`},
			{Code: `var [a, , b] = [1, 2, 3];`},
			{Code: `var [a,, b] = [1, 2, 3];`},
			{Code: `var [ , b] = a;`},
			{Code: `var [, b] = a;`},
			{Code: `var { a,} = a;`},
			{Code: `import { a,} from 'mod';`},

			// ---- _js_ valid: JSX (default) ----
			{Code: `<a>,</a>`, Tsx: true},
			{Code: `<a>  ,  </a>`, Tsx: true},
			{Code: `<a>Hello, world</a>`, Options: optsBefore(), Tsx: true},

			// ---- _js_ valid: backwards-compatibility null-element + comment shapes ----
			// "Ignoring spacing between a comment and comma of a null element was possibly unintentional." (upstream)
			{Code: `[a, /**/ , ]`, Options: optsAfter()},
			{Code: `[a , /**/, ]`, Options: optsBoth()},
			{Code: `[a, /**/ , ] = foo`, Options: optsAfter()},
			{Code: `[a , /**/, ] = foo`, Options: optsBoth()},

			// ---- _ts_ valid ----
			{Code: `const Foo = <T,>(foo: T) => {}`, Tsx: true},
			{Code: `function foo<T,>() {}`},
			{Code: `class Foo<T, T1> {}`},
			{Code: `type Foo<T, P,> = Bar<T, P>`},
			{Code: `interface Foo<T, T1,>{}`},
			{Code: `let foo,`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- _js_ invalid ----
			{
				Code:    `a(b,c)`,
				Output:  []string{`a(b , c)`},
				Options: optsBoth(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 4},
					{MessageId: "missing", Line: 1, Column: 4},
				},
			},
			{
				Code:    `new A(b,c)`,
				Output:  []string{`new A(b , c)`},
				Options: optsBoth(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 8},
					{MessageId: "missing", Line: 1, Column: 8},
				},
			},
			{
				Code:   `var a = 1 ,b = 2;`,
				Output: []string{`var a = 1, b = 2;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 11},
					{MessageId: "missing", Line: 1, Column: 11},
				},
			},
			{
				Code:   `var arr = [1 , 2];`,
				Output: []string{`var arr = [1, 2];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 14},
				},
			},
			{
				Code:   `var arr = [1 , ];`,
				Output: []string{`var arr = [1, ];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 14},
				},
			},
			{
				Code:   `var arr = [1 ,2];`,
				Output: []string{`var arr = [1, 2];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 14},
					{MessageId: "missing", Line: 1, Column: 14},
				},
			},
			{
				Code:   `var arr = [(1) , 2];`,
				Output: []string{`var arr = [(1), 2];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 16},
				},
			},
			{
				Code:    `var arr = [1, 2];`,
				Output:  []string{`var arr = [1 ,2];`},
				Options: optsBefore(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 13},
					{MessageId: "unexpected", Message: "There should be no space after ','.", Line: 1, Column: 13},
				},
			},
			{
				Code:    "var arr = [1\n  , 2];",
				Output:  []string{"var arr = [1\n  ,2];"},
				Options: optsNeither(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space after ','.", Line: 2, Column: 3},
				},
			},
			{
				Code:    "var arr = [1,\n  2];",
				Output:  []string{"var arr = [1 ,\n  2];"},
				Options: optsBefore(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 13},
				},
			},
			{
				Code:    "var obj = {'foo':\n'bar', 'baz':\n'qur'};",
				Output:  []string{"var obj = {'foo':\n'bar' ,'baz':\n'qur'};"},
				Options: optsBefore(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 6},
					{MessageId: "unexpected", Message: "There should be no space after ','.", Line: 2, Column: 6},
				},
			},
			{
				Code:   "var obj = {a: 1\n  ,b: 2};",
				Output: []string{"var obj = {a: 1\n  , b: 2};"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 3},
				},
			},
			{
				Code:    "var obj = {a: 1 ,\n  b: 2};",
				Output:  []string{"var obj = {a: 1,\n  b: 2};"},
				Options: optsNeither(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 17},
				},
			},
			{
				Code:    `var arr = [1 ,2];`,
				Output:  []string{`var arr = [1 , 2];`},
				Options: optsBoth(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 14},
				},
			},
			{
				Code:    `var arr = [1,2];`,
				Output:  []string{`var arr = [1 , 2];`},
				Options: optsBoth(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 13},
					{MessageId: "missing", Line: 1, Column: 13},
				},
			},
			{
				Code:    "var obj = {'foo':\n'bar','baz':\n'qur'};",
				Output:  []string{"var obj = {'foo':\n'bar' , 'baz':\n'qur'};"},
				Options: optsBoth(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 6},
					{MessageId: "missing", Line: 2, Column: 6},
				},
			},
			{
				Code:    `var arr = [1 , 2];`,
				Output:  []string{`var arr = [1,2];`},
				Options: optsNeither(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 14},
					{MessageId: "unexpected", Message: "There should be no space after ','.", Line: 1, Column: 14},
				},
			},
			{
				Code:    `a ,b`,
				Output:  []string{`a, b`},
				Options: optsAfter(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 3},
					{MessageId: "missing", Line: 1, Column: 3},
				},
			},
			{
				Code:    `function foo(a,b){}`,
				Output:  []string{`function foo(a , b){}`},
				Options: optsBoth(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 15},
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},
			{
				Code:    `var foo = (a,b) => {}`,
				Output:  []string{`var foo = (a , b) => {}`},
				Options: optsBoth(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 13},
					{MessageId: "missing", Line: 1, Column: 13},
				},
			},
			{
				Code:    `var foo = (a = 1,b) => {}`,
				Output:  []string{`var foo = (a = 1 , b) => {}`},
				Options: optsBoth(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 17},
					{MessageId: "missing", Line: 1, Column: 17},
				},
			},
			{
				Code:    `function foo(a = 1 ,b = 2) {}`,
				Output:  []string{`function foo(a = 1, b = 2) {}`},
				Options: optsAfter(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 20},
					{MessageId: "missing", Line: 1, Column: 20},
				},
			},
			{
				Code:   `<a>{foo(1 ,2)}</a>`,
				Output: []string{`<a>{foo(1, 2)}</a>`},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 11},
					{MessageId: "missing", Line: 1, Column: 11},
				},
			},
			{
				Code:   `myfunc(404, true/* bla bla bla */ , 'hello');`,
				Output: []string{`myfunc(404, true/* bla bla bla */, 'hello');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "There should be no space before ','.", Line: 1, Column: 35},
				},
			},
			{
				Code:   `myfunc(404, true,/* bla bla bla */ 'hello');`,
				Output: []string{`myfunc(404, true, /* bla bla bla */ 'hello');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 17},
				},
			},
			{
				Code:   "myfunc(404,// comment\n true, 'hello');",
				Output: []string{"myfunc(404, // comment\n true, 'hello');"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 11},
				},
			},

			// ---- _ts_ invalid ----
			{
				Code:   `function Foo<T,T1>() {}`,
				Output: []string{`function Foo<T, T1>() {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},
			{
				Code:   `function Foo<T , T1>() {}`,
				Output: []string{`function Foo<T, T1>() {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			{
				Code:   `function Foo<T ,T1>() {}`,
				Output: []string{`function Foo<T, T1>() {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
					{MessageId: "missing", Line: 1, Column: 16},
				},
			},
			{
				Code:    `function Foo<T, T1>() {}`,
				Output:  []string{`function Foo<T,T1>() {}`},
				Options: optsNeither(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			{
				Code:    `function Foo<T,T1>() {}`,
				Output:  []string{`function Foo<T ,T1>() {}`},
				Options: optsBefore(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},
			{
				Code:   `let foo ,`,
				Output: []string{`let foo,`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code:   `type Foo<T,P,> = Bar<T,P>`,
				Output: []string{`type Foo<T, P,> = Bar<T, P>`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 11},
					{MessageId: "missing", Line: 1, Column: 23},
				},
			},
		},
	)
}
