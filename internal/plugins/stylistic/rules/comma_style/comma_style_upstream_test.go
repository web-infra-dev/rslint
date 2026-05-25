// TestCommaStyleUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/rules/comma-style/comma-style.test.ts 1:1.
// Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases (tsgo AST edge shapes, branch lock-ins,
// real-user shapes) live in comma_style_extras_test.go.
//
// cspell:ignore nsnd fifi
// `nsnd` is a tokenization artifact from cspell merging the `n` of `\n` with
// the following `snd` token in the `{fst:1,\nsnd: [1,\n2]};` fixtures —
// `snd` is upstream's "second" abbreviation, kept verbatim. `fifi` is a
// person-name string literal upstream uses in `[\n ,'fifi' \n]`.
package comma_style_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/comma_style"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// optStr wraps a single string option in the array form rule_tester expects.
func optStr(s string) []any { return []any{s} }

// optWithExc wraps a (style, exceptions) pair into the multi-element array
// form the rule_tester / CLI emits. Using a bare map alongside the style
// string exercises the JSON path the CLI takes, not a typed struct path.
func optWithExc(style string, exc map[string]any) []any {
	return []any{style, map[string]any{"exceptions": exc}}
}

func errLast(line, col int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{
		{MessageId: "expectedCommaLast", Line: line, Column: col, EndLine: line, EndColumn: col + 1},
	}
}

func errFirst(line, col int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{
		{MessageId: "expectedCommaFirst", Line: line, Column: col, EndLine: line, EndColumn: col + 1},
	}
}

func errLone(line, col int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{
		{MessageId: "unexpectedLineBeforeAndAfterComma", Line: line, Column: col, EndLine: line, EndColumn: col + 1},
	}
}

type errSpec struct {
	id        string
	line, col int
}

func errs(specs ...errSpec) []rule_tester.InvalidTestCaseError {
	out := make([]rule_tester.InvalidTestCaseError, len(specs))
	for i, s := range specs {
		out[i] = rule_tester.InvalidTestCaseError{
			MessageId: s.id,
			Line:      s.line,
			Column:    s.col,
			EndLine:   s.line,
			EndColumn: s.col + 1,
		}
	}
	return out
}

func TestCommaStyleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&comma_style.CommaStyleRule,
		[]rule_tester.ValidTestCase{
			// ---- baseline single-line / multi-line shapes (default "last") ----
			{Code: `var foo = 1, bar = 3;`},
			{Code: `var foo = {'a': 1, 'b': 2};`},
			{Code: `var foo = [1, 2];`},
			{Code: `var foo = [, 2];`},
			{Code: `var foo = [1, ];`},
			{Code: "var foo = ['apples', \n 'oranges'];"},
			{Code: "var foo = {'a': 1, \n 'b': 2, \n'c': 3};"},
			{Code: "var foo = {'a': 1, \n 'b': 2, 'c':\n 3};"},
			{Code: "var foo = {'a': 1, \n 'b': 2, 'c': [{'d': 1}, \n {'e': 2}, \n {'f': 3}]};"},
			{Code: "var foo = [1, \n2, \n3];"},
			{Code: "function foo(){var a=[1,\n 2]}"},
			{Code: "function foo(){return {'a': 1,\n'b': 2}}"},
			{Code: "var foo = \n1, \nbar = \n2;"},
			{Code: "var foo = [\n(bar),\nbaz\n];"},
			{Code: "var foo = [\n(bar\n),\nbaz\n];"},
			{Code: "var foo = [\n(\nbar\n),\nbaz\n];"},
			{Code: "new Foo(a\n,b);", Options: optWithExc("last", map[string]any{"NewExpression": true})},
			{Code: "var foo = [\n(bar\n)\n,baz\n];", Options: optStr("first")},
			{Code: "var foo = \n1, \nbar = [1,\n2,\n3]"},
			{Code: "var foo = ['apples'\n,'oranges'];", Options: optStr("first")},
			{Code: `var foo = 1, bar = 2;`, Options: optStr("first")},
			{Code: "var foo = 1 \n ,bar = 2;", Options: optStr("first")},
			{Code: "var foo = {'a': 1 \n ,'b': 2 \n,'c': 3};", Options: optStr("first")},
			{Code: "var foo = [1 \n ,2 \n, 3];", Options: optStr("first")},
			{Code: "function foo(){return {'a': 1\n,'b': 2}}", Options: optStr("first")},
			{Code: "function foo(){var a=[1\n, 2]}", Options: optStr("first")},
			{Code: "new Foo(a,\nb);", Options: optWithExc("first", map[string]any{"NewExpression": true})},
			{Code: "f(1\n, 2);", Options: optWithExc("last", map[string]any{"CallExpression": true})},
			{Code: "function foo(a\n, b) { return a + b; }", Options: optWithExc("last", map[string]any{"FunctionDeclaration": true})},
			{
				Code:    "var a = 'a',\no = 'o';",
				Options: optWithExc("first", map[string]any{"VariableDeclaration": true}),
			},
			{
				Code:    "var arr = ['a',\n'o'];",
				Options: optWithExc("first", map[string]any{"ArrayExpression": true}),
			},
			{
				Code:    "var obj = {a: 'a',\nb: 'b'};",
				Options: optWithExc("first", map[string]any{"ObjectExpression": true}),
			},
			{
				Code:    "var a = 'a',\no = 'o',\narr = [1,\n2];",
				Options: optWithExc("first", map[string]any{"VariableDeclaration": true, "ArrayExpression": true}),
			},
			{
				Code:    "var ar ={fst:1,\nsnd: [1,\n2]};",
				Options: optWithExc("first", map[string]any{"ArrayExpression": true, "ObjectExpression": true}),
			},
			{
				Code: "var a = 'a',\nar ={fst:1,\nsnd: [1,\n2]};",
				Options: optWithExc("first", map[string]any{
					"ArrayExpression":     true,
					"ObjectExpression":    true,
					"VariableDeclaration": true,
				}),
			},
			{
				Code:    "const foo = (a\n, b) => { return a + b; }",
				Options: optWithExc("last", map[string]any{"ArrowFunctionExpression": true}),
			},
			{
				Code:    "function foo([a\n, b]) { return a + b; }",
				Options: optWithExc("last", map[string]any{"ArrayPattern": true}),
			},
			{
				Code:    "const foo = ([a\n, b]) => { return a + b; }",
				Options: optWithExc("last", map[string]any{"ArrayPattern": true}),
			},
			{
				Code:    "import { a\n, b } from './source';",
				Options: optWithExc("last", map[string]any{"ImportDeclaration": true}),
			},
			{
				Code:    "const foo = function (a\n, b) { return a + b; }",
				Options: optWithExc("last", map[string]any{"FunctionExpression": true}),
			},
			{
				Code:    "var {foo\n, bar} = {foo:'apples', bar:'oranges'};",
				Options: optWithExc("last", map[string]any{"ObjectPattern": true}),
			},
			{
				Code:    "var {foo\n, bar} = {foo:'apples', bar:'oranges'};",
				Options: optWithExc("first", map[string]any{"ObjectPattern": true}),
			},
			{
				Code:    "new Foo(a,\nb);",
				Options: optWithExc("first", map[string]any{"NewExpression": true}),
			},
			{
				Code:    "f(1\n, 2);",
				Options: optWithExc("last", map[string]any{"CallExpression": true}),
			},
			{
				Code:    "function foo(a\n, b) { return a + b; }",
				Options: optWithExc("last", map[string]any{"FunctionDeclaration": true}),
			},
			{
				Code:    "const foo = function (a\n, b) { return a + b; }",
				Options: optWithExc("last", map[string]any{"FunctionExpression": true}),
			},
			{
				Code:    "function foo([a\n, b]) { return a + b; }",
				Options: optWithExc("last", map[string]any{"ArrayPattern": true}),
			},
			{
				Code:    "const foo = (a\n, b) => { return a + b; }",
				Options: optWithExc("last", map[string]any{"ArrowFunctionExpression": true}),
			},
			{
				Code:    "const foo = ([a\n, b]) => { return a + b; }",
				Options: optWithExc("last", map[string]any{"ArrayPattern": true}),
			},
			{
				Code:    "import { a\n, b } from './source';",
				Options: optWithExc("last", map[string]any{"ImportDeclaration": true}),
			},
			{
				Code:    "var {foo\n, bar} = {foo:'apples', bar:'oranges'};",
				Options: optWithExc("last", map[string]any{"ObjectPattern": true}),
			},
			{
				Code:    "new Foo(a,\nb);",
				Options: optWithExc("last", map[string]any{"NewExpression": false}),
			},
			{
				Code:    "new Foo(a\n,b);",
				Options: optWithExc("last", map[string]any{"NewExpression": true}),
			},
			{Code: "var foo = [\n , \n 1, \n 2 \n];"},
			{
				Code:    "const [\n , \n , \n a, \n b, \n] = arr;",
				Options: optWithExc("last", map[string]any{"ArrayPattern": false}),
			},
			{
				Code:    "const [\n ,, \n a, \n b, \n] = arr;",
				Options: optWithExc("last", map[string]any{"ArrayPattern": false}),
			},
			{
				Code:    "const arr = [\n 1 \n , \n ,2 \n]",
				Options: optStr("first"),
			},
			{
				Code:    "const arr = [\n ,'fifi' \n]",
				Options: optStr("first"),
			},

			// ---- exception: ImportDeclaration ----
			{
				Code: "import {\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"} from 'module3' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};\n" +
					"import 'module4' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};",
				Options: optWithExc("last", map[string]any{"ImportDeclaration": true}),
			},
			// ---- exception: ExportAllDeclaration + ExportNamedDeclaration ----
			{
				Code: "let a, b, c;\n" +
					"export {\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					"};\n" +
					"export {\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"} from 'module1' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};\n" +
					"export * from 'module2' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};",
				Options: optWithExc("last", map[string]any{
					"ExportAllDeclaration":   true,
					"ExportNamedDeclaration": true,
				}),
			},
			// ---- exception: ImportExpression ----
			{
				Code: "import(\n" +
					"  a,\n" +
					"  b\n" +
					");\n" +
					"import(\n" +
					"  c\n" +
					"  , d\n" +
					");",
				Options: optWithExc("first", map[string]any{"ImportExpression": true}),
			},
			// ---- ImportExpression: false / unset still validates calls ----
			{
				Code: "import(\n" +
					"  a,\n" +
					");\n" +
					"import(\n" +
					"  b, c,\n" +
					");",
				Options: optWithExc("last", map[string]any{"ImportExpression": false}),
			},
			{
				Code: "import(\n" +
					"  a\n" +
					",);\n" +
					"import(\n" +
					"  b, c\n" +
					",);",
				Options: optWithExc("first", map[string]any{"ImportExpression": false}),
			},
			// ---- exception: SequenceExpression ----
			{
				Code: "const x = (\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					");",
				Options: optWithExc("first", map[string]any{"SequenceExpression": true}),
			},
			// ---- exception: ClassDeclaration + ClassExpression (implements) ----
			{
				Code: "class MyClass implements\n" +
					"  A,\n" +
					"  B\n" +
					", C {\n" +
					"}\n" +
					"const a = class implements\n" +
					"  A,\n" +
					"  B\n" +
					", C {\n" +
					"}",
				Options: optWithExc("first", map[string]any{
					"ClassDeclaration": true,
					"ClassExpression":  true,
				}),
			},
			// ---- exception: TSDeclareFunction / TSFunctionType / TSConstructorType / TSEmptyBodyFunctionExpression ----
			{
				Code: "function f(\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					");\n" +
					"type a = (\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					") => r\n" +
					"type a = new (\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					") => r\n" +
					"abstract class Base {\n" +
					"  f(\n" +
					"    a,\n" +
					"    b\n" +
					"    , c\n" +
					"  );\n" +
					"}",
				Options: optWithExc("first", map[string]any{
					"TSDeclareFunction":             true,
					"TSFunctionType":                true,
					"TSConstructorType":             true,
					"TSEmptyBodyFunctionExpression": true,
				}),
			},
			// ---- exception: TSEnumBody ----
			{
				Code: "enum MyEnum {\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"}",
				Options: optWithExc("first", map[string]any{"TSEnumBody": true}),
			},
			// ---- exception: TSTypeLiteral ----
			{
				Code: "type foo = {\n" +
					"  a: string,\n" +
					"  b: string\n" +
					"  , c: string\n" +
					"}",
				Options: optWithExc("first", map[string]any{"TSTypeLiteral": true}),
			},
			// ---- exception: TSTypeLiteral + signatures + TSIndexSignature ----
			{
				Code: "type foo = {\n" +
					"  new (\n" +
					"    a,\n" +
					"    b\n" +
					"    , c\n" +
					"  ): any,\n" +
					"  (\n" +
					"    a,\n" +
					"    b\n" +
					"    , c\n" +
					"  ): any,\n" +
					"  [\n" +
					"    a: string,\n" +
					"    b: string\n" +
					"    , c: string\n" +
					"  ]: string,\n" +
					"\n" +
					"  f(\n" +
					"    a: string,\n" +
					"    b: string\n" +
					"    , c: string\n" +
					"  ): number,\n" +
					"}",
				Options: optWithExc("first", map[string]any{
					"TSTypeLiteral":                  true,
					"TSCallSignatureDeclaration":     true,
					"TSConstructSignatureDeclaration": true,
					"TSIndexSignature":               true,
					"TSMethodSignature":              true,
				}),
			},
			// ---- exception: TSInterfaceBody + TSInterfaceDeclaration ----
			{
				Code: "interface Foo extends\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"{\n" +
					"  a: string,\n" +
					"  b: string\n" +
					"  , c: string\n" +
					"}",
				Options: optWithExc("first", map[string]any{
					"TSInterfaceBody":        true,
					"TSInterfaceDeclaration": true,
				}),
			},
			// ---- exception: TSTupleType ----
			{
				Code: "type Foo = [\n" +
					"  \"A\",\n" +
					"  \"B\"\n" +
					"  , \"C\"\n" +
					"];",
				Options: optWithExc("first", map[string]any{"TSTupleType": true}),
			},
			// ---- exception: TSTypeParameterDeclaration + TSTypeParameterInstantiation ----
			{
				Code: "type Foo<\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"> = Bar<\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					">;",
				Options: optWithExc("first", map[string]any{
					"TSTypeParameterDeclaration":    true,
					"TSTypeParameterInstantiation":  true,
				}),
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "var foo = { a: 1. //comment \n, b: 2\n}",
				Output: []string{"var foo = { a: 1., //comment \n b: 2\n}"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "var foo = { a: 1. //comment \n //comment1 \n //comment2 \n, b: 2\n}",
				Output: []string{"var foo = { a: 1., //comment \n //comment1 \n //comment2 \n b: 2\n}"},
				Errors: errLast(4, 1),
			},
			{
				Code:   "var foo = 1\n,\nbar = 2;",
				Output: []string{"var foo = 1,\nbar = 2;"},
				Errors: errLone(2, 1),
			},
			{
				Code:   "var foo = 1 //comment\n,\nbar = 2;",
				Output: []string{"var foo = 1, //comment\nbar = 2;"},
				Errors: errLone(2, 1),
			},
			{
				Code:   "var foo = 1 //comment\n, // comment 2\nbar = 2;",
				Output: []string{"var foo = 1, //comment // comment 2\nbar = 2;"},
				Errors: errLone(2, 1),
			},
			{
				Code:   "new Foo(a\n,\nb);",
				Output: []string{"new Foo(a,\nb);"},
				Errors: errLone(2, 1),
			},
			{
				Code:   "var foo = 1\n,bar = 2;",
				Output: []string{"var foo = 1,\nbar = 2;"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "f([1,2\n,3]);",
				Output: []string{"f([1,2,\n3]);"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "f([1,2\n,]);",
				Output: []string{"f([1,2,\n]);"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "f([,2\n,3]);",
				Output: []string{"f([,2,\n3]);"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "var foo = ['apples'\n, 'oranges'];",
				Output: []string{"var foo = ['apples',\n 'oranges'];"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "var [foo\n, bar] = ['apples', 'oranges'];",
				Output: []string{"var [foo,\n bar] = ['apples', 'oranges'];"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "f(1\n, 2);",
				Output: []string{"f(1,\n 2);"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "function foo(a\n, b) { return a + b; }",
				Output: []string{"function foo(a,\n b) { return a + b; }"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "const foo = function (a\n, b) { return a + b; }",
				Output: []string{"const foo = function (a,\n b) { return a + b; }"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "function foo([a\n, b]) { return a + b; }",
				Output: []string{"function foo([a,\n b]) { return a + b; }"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "const foo = (a\n, b) => { return a + b; }",
				Output: []string{"const foo = (a,\n b) => { return a + b; }"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "const foo = ([a\n, b]) => { return a + b; }",
				Output: []string{"const foo = ([a,\n b]) => { return a + b; }"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "import { a\n, b } from './source';",
				Output: []string{"import { a,\n b } from './source';"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "var {foo\n, bar} = {foo:'apples', bar:'oranges'};",
				Output: []string{"var {foo,\n bar} = {foo:'apples', bar:'oranges'};"},
				Errors: errLast(2, 1),
			},
			{
				Code:    "var foo = 1,\nbar = 2;",
				Output:  []string{"var foo = 1\n,bar = 2;"},
				Options: optStr("first"),
				Errors:  errFirst(1, 12),
			},
			{
				Code:    "f([1,\n2,3]);",
				Output:  []string{"f([1\n,2,3]);"},
				Options: optStr("first"),
				Errors:  errFirst(1, 5),
			},
			{
				Code:    "var foo = ['apples', \n 'oranges'];",
				Output:  []string{"var foo = ['apples' \n ,'oranges'];"},
				Options: optStr("first"),
				Errors:  errFirst(1, 20),
			},
			{
				Code:    "var foo = {'a': 1, \n 'b': 2\n ,'c': 3};",
				Output:  []string{"var foo = {'a': 1 \n ,'b': 2\n ,'c': 3};"},
				Options: optStr("first"),
				Errors:  errFirst(1, 18),
			},
			{
				Code:    "var a = 'a',\no = 'o',\narr = [1,\n2];",
				Output:  []string{"var a = 'a',\no = 'o',\narr = [1\n,2];"},
				Options: optWithExc("first", map[string]any{"VariableDeclaration": true}),
				Errors:  errFirst(3, 9),
			},
			{
				Code:    "var a = 'a',\nobj = {a: 'a',\nb: 'b'};",
				Output:  []string{"var a = 'a',\nobj = {a: 'a'\n,b: 'b'};"},
				Options: optWithExc("first", map[string]any{"VariableDeclaration": true}),
				Errors:  errFirst(2, 14),
			},
			{
				Code:    "var a = 'a',\nobj = {a: 'a',\nb: 'b'};",
				Output:  []string{"var a = 'a'\n,obj = {a: 'a',\nb: 'b'};"},
				Options: optWithExc("first", map[string]any{"ObjectExpression": true}),
				Errors:  errFirst(1, 12),
			},
			{
				Code:    "var a = 'a',\narr = [1,\n2];",
				Output:  []string{"var a = 'a'\n,arr = [1,\n2];"},
				Options: optWithExc("first", map[string]any{"ArrayExpression": true}),
				Errors:  errFirst(1, 12),
			},
			{
				Code:    "var ar =[1,\n{a: 'a',\nb: 'b'}];",
				Output:  []string{"var ar =[1,\n{a: 'a'\n,b: 'b'}];"},
				Options: optWithExc("first", map[string]any{"ArrayExpression": true}),
				Errors:  errFirst(2, 8),
			},
			{
				Code:    "var ar =[1,\n{a: 'a',\nb: 'b'}];",
				Output:  []string{"var ar =[1\n,{a: 'a',\nb: 'b'}];"},
				Options: optWithExc("first", map[string]any{"ObjectExpression": true}),
				Errors:  errFirst(1, 11),
			},
			{
				Code:    "var ar ={fst:1,\nsnd: [1,\n2]};",
				Output:  []string{"var ar ={fst:1,\nsnd: [1\n,2]};"},
				Options: optWithExc("first", map[string]any{"ObjectExpression": true}),
				Errors:  errFirst(2, 8),
			},
			{
				Code:    "var ar ={fst:1,\nsnd: [1,\n2]};",
				Output:  []string{"var ar ={fst:1\n,snd: [1,\n2]};"},
				Options: optWithExc("first", map[string]any{"ArrayExpression": true}),
				Errors:  errFirst(1, 15),
			},
			{
				Code:    "new Foo(a,\nb);",
				Output:  []string{"new Foo(a\n,b);"},
				Options: optStr("first"),
				Errors:  errFirst(1, 10),
			},
			{
				Code:   "var foo = [\n(bar\n)\n,\nbaz\n];",
				Output: []string{"var foo = [\n(bar\n),\nbaz\n];"},
				Errors: errLone(4, 1),
			},
			{
				Code:   "[(foo),\n,\nbar]",
				Output: []string{"[(foo),,\nbar]"},
				Errors: errLone(2, 1),
			},
			{
				Code:   "new Foo(a\n,b);",
				Output: []string{"new Foo(a,\nb);"},
				Errors: errLast(2, 1),
			},
			{
				Code:   "[\n[foo(3)],\n,\nbar\n];",
				Output: []string{"[\n[foo(3)],,\nbar\n];"},
				Errors: errLone(3, 1),
			},
			{
				// https://github.com/eslint/eslint/issues/10632
				Code:   "[foo//\n,/*block\ncomment*/];",
				Output: []string{"[foo,//\n/*block\ncomment*/];"},
				Errors: errLone(2, 1),
			},
			// ---- import / with attributes — default "last" ----
			{
				Code: "import {\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"} from 'module3' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};\n" +
					"import 'module4' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};",
				Output: []string{
					"import {\n" +
						"  A,\n" +
						"  B,\n" +
						"   C\n" +
						"} from 'module3' with {\n" +
						"  a: 'v',\n" +
						"  b: 'v',\n" +
						"   c: 'v'\n" +
						"};\n" +
						"import 'module4' with {\n" +
						"  a: 'v',\n" +
						"  b: 'v',\n" +
						"   c: 'v'\n" +
						"};",
				},
				Errors: errs(
					errSpec{"expectedCommaLast", 4, 3},
					errSpec{"expectedCommaLast", 8, 3},
					errSpec{"expectedCommaLast", 13, 3},
				),
			},
			// ---- import / with attributes — "first" ----
			{
				Code: "import {\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"} from 'module3' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};\n" +
					"import 'module4' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};",
				Output: []string{
					"import {\n" +
						"  A\n" +
						"  ,B\n" +
						"  , C\n" +
						"} from 'module3' with {\n" +
						"  a: 'v'\n" +
						"  ,b: 'v'\n" +
						"  , c: 'v'\n" +
						"};\n" +
						"import 'module4' with {\n" +
						"  a: 'v'\n" +
						"  ,b: 'v'\n" +
						"  , c: 'v'\n" +
						"};",
				},
				Options: optStr("first"),
				Errors: errs(
					errSpec{"expectedCommaFirst", 2, 4},
					errSpec{"expectedCommaFirst", 6, 9},
					errSpec{"expectedCommaFirst", 11, 9},
				),
			},
			// ---- export {} / with attributes — default "last" ----
			{
				Code: "let a, b, c;\n" +
					"export {\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					"};\n" +
					"export {\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"} from 'module1' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};\n" +
					"export * from 'module2' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};",
				Output: []string{
					"let a, b, c;\n" +
						"export {\n" +
						"  a,\n" +
						"  b,\n" +
						"   c\n" +
						"};\n" +
						"export {\n" +
						"  A,\n" +
						"  B,\n" +
						"   C\n" +
						"} from 'module1' with {\n" +
						"  a: 'v',\n" +
						"  b: 'v',\n" +
						"   c: 'v'\n" +
						"};\n" +
						"export * from 'module2' with {\n" +
						"  a: 'v',\n" +
						"  b: 'v',\n" +
						"   c: 'v'\n" +
						"};",
				},
				Errors: errs(
					errSpec{"expectedCommaLast", 5, 3},
					errSpec{"expectedCommaLast", 10, 3},
					errSpec{"expectedCommaLast", 14, 3},
					errSpec{"expectedCommaLast", 19, 3},
				),
			},
			// ---- export {} / with attributes — "first" ----
			{
				Code: "let a, b, c;\n" +
					"export {\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					"};\n" +
					"export {\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"} from 'module1' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};\n" +
					"export * from 'module2' with {\n" +
					"  a: 'v',\n" +
					"  b: 'v'\n" +
					"  , c: 'v'\n" +
					"};",
				Output: []string{
					"let a, b, c;\n" +
						"export {\n" +
						"  a\n" +
						"  ,b\n" +
						"  , c\n" +
						"};\n" +
						"export {\n" +
						"  A\n" +
						"  ,B\n" +
						"  , C\n" +
						"} from 'module1' with {\n" +
						"  a: 'v'\n" +
						"  ,b: 'v'\n" +
						"  , c: 'v'\n" +
						"};\n" +
						"export * from 'module2' with {\n" +
						"  a: 'v'\n" +
						"  ,b: 'v'\n" +
						"  , c: 'v'\n" +
						"};",
				},
				Options: optStr("first"),
				Errors: errs(
					errSpec{"expectedCommaFirst", 3, 4},
					errSpec{"expectedCommaFirst", 8, 4},
					errSpec{"expectedCommaFirst", 12, 9},
					errSpec{"expectedCommaFirst", 17, 9},
				),
			},
			// ---- SequenceExpression — default "last" ----
			{
				Code: "const x = (\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					");",
				Output: []string{
					"const x = (\n" +
						"  a,\n" +
						"  b,\n" +
						"   c\n" +
						");",
				},
				Errors: errLast(4, 3),
			},
			// ---- SequenceExpression — "first" ----
			{
				Code: "const x = (\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					");",
				Output: []string{
					"const x = (\n" +
						"  a\n" +
						"  ,b\n" +
						"  , c\n" +
						");",
				},
				Options: optStr("first"),
				Errors:  errFirst(2, 4),
			},
			// ---- ImportExpression — default "last" ----
			{
				Code: "import(\n" +
					"  a,\n" +
					"  b\n" +
					");\n" +
					"import(\n" +
					"  c\n" +
					"  , d\n" +
					");",
				Output: []string{
					"import(\n" +
						"  a,\n" +
						"  b\n" +
						");\n" +
						"import(\n" +
						"  c,\n" +
						"   d\n" +
						");",
				},
				Errors: errLast(7, 3),
			},
			// ---- ImportExpression — "first" ----
			{
				Code: "import(\n" +
					"  a,\n" +
					"  b\n" +
					");\n" +
					"import(\n" +
					"  c\n" +
					"  , d\n" +
					");",
				Output: []string{
					"import(\n" +
						"  a\n" +
						"  ,b\n" +
						");\n" +
						"import(\n" +
						"  c\n" +
						"  , d\n" +
						");",
				},
				Options: optStr("first"),
				Errors:  errFirst(2, 4),
			},
			// ---- Class implements — default "last" ----
			{
				Code: "class MyClass implements\n" +
					"  A,\n" +
					"  B\n" +
					", C {\n" +
					"}\n" +
					"const a = class implements\n" +
					"  A,\n" +
					"  B\n" +
					", C {\n" +
					"}",
				Output: []string{
					"class MyClass implements\n" +
						"  A,\n" +
						"  B,\n" +
						" C {\n" +
						"}\n" +
						"const a = class implements\n" +
						"  A,\n" +
						"  B,\n" +
						" C {\n" +
						"}",
				},
				Errors: errs(
					errSpec{"expectedCommaLast", 4, 1},
					errSpec{"expectedCommaLast", 9, 1},
				),
			},
			// ---- Class implements — "first" ----
			{
				Code: "class MyClass implements\n" +
					"  A,\n" +
					"  B\n" +
					", C {\n" +
					"}\n" +
					"const a = class implements\n" +
					"  A,\n" +
					"  B\n" +
					", C {\n" +
					"}",
				Output: []string{
					"class MyClass implements\n" +
						"  A\n" +
						"  ,B\n" +
						", C {\n" +
						"}\n" +
						"const a = class implements\n" +
						"  A\n" +
						"  ,B\n" +
						", C {\n" +
						"}",
				},
				Options: optStr("first"),
				Errors: errs(
					errSpec{"expectedCommaFirst", 2, 4},
					errSpec{"expectedCommaFirst", 7, 4},
				),
			},
			// ---- TSDeclareFunction / TSFunctionType / TSConstructorType / TSEmptyBodyFunctionExpression — default "last" ----
			{
				Code: "function f(\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					")\n" +
					"type a = (\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					") => r\n" +
					"type a = new (\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					") => r\n" +
					"abstract class Base {\n" +
					"  f(\n" +
					"    a,\n" +
					"    b\n" +
					"    , c\n" +
					"  );\n" +
					"}",
				Output: []string{
					"function f(\n" +
						"  a,\n" +
						"  b,\n" +
						"   c\n" +
						")\n" +
						"type a = (\n" +
						"  a,\n" +
						"  b,\n" +
						"   c\n" +
						") => r\n" +
						"type a = new (\n" +
						"  a,\n" +
						"  b,\n" +
						"   c\n" +
						") => r\n" +
						"abstract class Base {\n" +
						"  f(\n" +
						"    a,\n" +
						"    b,\n" +
						"     c\n" +
						"  );\n" +
						"}",
				},
				Errors: errs(
					errSpec{"expectedCommaLast", 4, 3},
					errSpec{"expectedCommaLast", 9, 3},
					errSpec{"expectedCommaLast", 14, 3},
					errSpec{"expectedCommaLast", 20, 5},
				),
			},
			// ---- Same code — "first" ----
			{
				Code: "function f(\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					")\n" +
					"type a = (\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					") => r\n" +
					"type a = new (\n" +
					"  a,\n" +
					"  b\n" +
					"  , c\n" +
					") => r\n" +
					"abstract class Base {\n" +
					"  f(\n" +
					"    a,\n" +
					"    b\n" +
					"    , c\n" +
					"  );\n" +
					"}",
				Output: []string{
					"function f(\n" +
						"  a\n" +
						"  ,b\n" +
						"  , c\n" +
						")\n" +
						"type a = (\n" +
						"  a\n" +
						"  ,b\n" +
						"  , c\n" +
						") => r\n" +
						"type a = new (\n" +
						"  a\n" +
						"  ,b\n" +
						"  , c\n" +
						") => r\n" +
						"abstract class Base {\n" +
						"  f(\n" +
						"    a\n" +
						"    ,b\n" +
						"    , c\n" +
						"  );\n" +
						"}",
				},
				Options: optStr("first"),
				Errors: errs(
					errSpec{"expectedCommaFirst", 2, 4},
					errSpec{"expectedCommaFirst", 7, 4},
					errSpec{"expectedCommaFirst", 12, 4},
					errSpec{"expectedCommaFirst", 18, 6},
				),
			},
			// ---- enum — default "last" ----
			{
				Code: "enum MyEnum {\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"}",
				Output: []string{
					"enum MyEnum {\n" +
						"  A,\n" +
						"  B,\n" +
						"   C\n" +
						"}",
				},
				Errors: errLast(4, 3),
			},
			// ---- enum — "first" ----
			{
				Code: "enum MyEnum {\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"}",
				Output: []string{
					"enum MyEnum {\n" +
						"  A\n" +
						"  ,B\n" +
						"  , C\n" +
						"}",
				},
				Options: optStr("first"),
				Errors:  errFirst(2, 4),
			},
			// ---- TSTypeLiteral — default "last" ----
			{
				Code: "type foo = {\n" +
					"  a: string,\n" +
					"  b: string\n" +
					"  , c: string\n" +
					"}",
				Output: []string{
					"type foo = {\n" +
						"  a: string,\n" +
						"  b: string,\n" +
						"   c: string\n" +
						"}",
				},
				Errors: errLast(4, 3),
			},
			// ---- TSTypeLiteral — "first" ----
			{
				Code: "type foo = {\n" +
					"  a: string,\n" +
					"  b: string\n" +
					"  , c: string\n" +
					"}",
				Output: []string{
					"type foo = {\n" +
						"  a: string\n" +
						"  ,b: string\n" +
						"  , c: string\n" +
						"}",
				},
				Options: optStr("first"),
				Errors:  errFirst(2, 12),
			},
			// ---- TSTypeLiteral with signatures — default "last" ----
			{
				Code: "type foo = {\n" +
					"  new (\n" +
					"    a,\n" +
					"    b\n" +
					"    , c\n" +
					"  ): any,\n" +
					"  (\n" +
					"    a,\n" +
					"    b\n" +
					"    , c\n" +
					"  ): any,\n" +
					"  [\n" +
					"    a: string,\n" +
					"    b: string\n" +
					"    , c: string\n" +
					"  ]: string,\n" +
					"\n" +
					"  f(\n" +
					"    a: string,\n" +
					"    b: string\n" +
					"    , c: string\n" +
					"  ): number,\n" +
					"}",
				Output: []string{
					"type foo = {\n" +
						"  new (\n" +
						"    a,\n" +
						"    b,\n" +
						"     c\n" +
						"  ): any,\n" +
						"  (\n" +
						"    a,\n" +
						"    b,\n" +
						"     c\n" +
						"  ): any,\n" +
						"  [\n" +
						"    a: string,\n" +
						"    b: string,\n" +
						"     c: string\n" +
						"  ]: string,\n" +
						"\n" +
						"  f(\n" +
						"    a: string,\n" +
						"    b: string,\n" +
						"     c: string\n" +
						"  ): number,\n" +
						"}",
				},
				Errors: errs(
					errSpec{"expectedCommaLast", 5, 5},
					errSpec{"expectedCommaLast", 10, 5},
					errSpec{"expectedCommaLast", 15, 5},
					errSpec{"expectedCommaLast", 21, 5},
				),
			},
			// ---- TSTypeLiteral with signatures — "first" ----
			// Listener firing order: outer KindTypeLiteral first (the four
			// member-separator commas), then each nested signature in
			// source order (their `a,`-shaped param-list commas).
			{
				Code: "type foo = {\n" +
					"  new (\n" +
					"    a,\n" +
					"    b\n" +
					"    , c\n" +
					"  ): any,\n" +
					"  (\n" +
					"    a,\n" +
					"    b\n" +
					"    , c\n" +
					"  ): any,\n" +
					"  [\n" +
					"    a: string,\n" +
					"    b: string\n" +
					"    , c: string\n" +
					"  ]: string,\n" +
					"\n" +
					"  f(\n" +
					"    a: string,\n" +
					"    b: string\n" +
					"    , c: string\n" +
					"  ): number,\n" +
					"}",
				Output: []string{
					"type foo = {\n" +
						"  new (\n" +
						"    a\n" +
						"    ,b\n" +
						"    , c\n" +
						"  ): any\n" +
						"  ,(\n" +
						"    a\n" +
						"    ,b\n" +
						"    , c\n" +
						"  ): any\n" +
						"  ,[\n" +
						"    a: string\n" +
						"    ,b: string\n" +
						"    , c: string\n" +
						"  ]: string\n" +
						"\n" +
						"  ,f(\n" +
						"    a: string\n" +
						"    ,b: string\n" +
						"    , c: string\n" +
						"  ): number\n" +
						",}",
				},
				Options: optStr("first"),
				Errors: errs(
					errSpec{"expectedCommaFirst", 6, 9},
					errSpec{"expectedCommaFirst", 11, 9},
					errSpec{"expectedCommaFirst", 16, 12},
					errSpec{"expectedCommaFirst", 22, 12},
					errSpec{"expectedCommaFirst", 3, 6},
					errSpec{"expectedCommaFirst", 8, 6},
					errSpec{"expectedCommaFirst", 13, 14},
					errSpec{"expectedCommaFirst", 19, 14},
				),
			},
			// ---- TSInterfaceBody + TSInterfaceDeclaration — default "last" ----
			{
				Code: "interface Foo extends\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"{\n" +
					"  a: string,\n" +
					"  b: string\n" +
					"  , c: string\n" +
					"}",
				Output: []string{
					"interface Foo extends\n" +
						"  A,\n" +
						"  B,\n" +
						"   C\n" +
						"{\n" +
						"  a: string,\n" +
						"  b: string,\n" +
						"   c: string\n" +
						"}",
				},
				Errors: errs(
					errSpec{"expectedCommaLast", 4, 3},
					errSpec{"expectedCommaLast", 8, 3},
				),
			},
			// ---- TSInterfaceBody + TSInterfaceDeclaration — "first" ----
			{
				Code: "interface Foo extends\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"{\n" +
					"  a: string,\n" +
					"  b: string\n" +
					"  , c: string\n" +
					"}",
				Output: []string{
					"interface Foo extends\n" +
						"  A\n" +
						"  ,B\n" +
						"  , C\n" +
						"{\n" +
						"  a: string\n" +
						"  ,b: string\n" +
						"  , c: string\n" +
						"}",
				},
				Options: optStr("first"),
				Errors: errs(
					errSpec{"expectedCommaFirst", 2, 4},
					errSpec{"expectedCommaFirst", 6, 12},
				),
			},
			// ---- TSTupleType — default "last" ----
			{
				Code: "type Foo = [\n" +
					"  \"A\",\n" +
					"  \"B\"\n" +
					"  , \"C\"\n" +
					"];",
				Output: []string{
					"type Foo = [\n" +
						"  \"A\",\n" +
						"  \"B\",\n" +
						"   \"C\"\n" +
						"];",
				},
				Errors: errLast(4, 3),
			},
			// ---- TSTupleType — "first" ----
			{
				Code: "type Foo = [\n" +
					"  \"A\",\n" +
					"  \"B\"\n" +
					"  , \"C\"\n" +
					"];",
				Output: []string{
					"type Foo = [\n" +
						"  \"A\"\n" +
						"  ,\"B\"\n" +
						"  , \"C\"\n" +
						"];",
				},
				Options: optStr("first"),
				Errors:  errFirst(2, 6),
			},
			// ---- TSTypeParameterDeclaration + TSTypeParameterInstantiation — default "last" ----
			{
				Code: "type Foo<\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"> = Bar<\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					">;",
				Output: []string{
					"type Foo<\n" +
						"  A,\n" +
						"  B,\n" +
						"   C\n" +
						"> = Bar<\n" +
						"  A,\n" +
						"  B,\n" +
						"   C\n" +
						">;",
				},
				Errors: errs(
					errSpec{"expectedCommaLast", 4, 3},
					errSpec{"expectedCommaLast", 8, 3},
				),
			},
			// ---- TSTypeParameterDeclaration + TSTypeParameterInstantiation — "first" ----
			{
				Code: "type Foo<\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					"> = Bar<\n" +
					"  A,\n" +
					"  B\n" +
					"  , C\n" +
					">;",
				Output: []string{
					"type Foo<\n" +
						"  A\n" +
						"  ,B\n" +
						"  , C\n" +
						"> = Bar<\n" +
						"  A\n" +
						"  ,B\n" +
						"  , C\n" +
						">;",
				},
				Options: optStr("first"),
				Errors: errs(
					errSpec{"expectedCommaFirst", 2, 4},
					errSpec{"expectedCommaFirst", 6, 4},
				),
			},
			// ---- ImportDeclaration: default specifier + named bindings comma — "last" ----
			{
				Code: "import a\n" +
					"  , {Foo} from 'module'\n" +
					"import b\n" +
					"  , {} from 'module'\n" +
					"import c,\n" +
					"  {Bar} from 'module'\n" +
					"import d,\n" +
					"  {} from 'module'",
				Output: []string{
					"import a,\n" +
						"   {Foo} from 'module'\n" +
						"import b,\n" +
						"   {} from 'module'\n" +
						"import c,\n" +
						"  {Bar} from 'module'\n" +
						"import d,\n" +
						"  {} from 'module'",
				},
				Options: optWithExc("last", map[string]any{"ImportDeclaration": false}),
				Errors: errs(
					errSpec{"expectedCommaLast", 2, 3},
					errSpec{"expectedCommaLast", 4, 3},
				),
			},
			// ---- ImportDeclaration: default + named bindings comma — "first" ----
			{
				Code: "import a\n" +
					"  , {Foo} from 'module'\n" +
					"import b\n" +
					"  , {} from 'module'\n" +
					"import c,\n" +
					"  {Bar} from 'module'\n" +
					"import d,\n" +
					"  {} from 'module'",
				Output: []string{
					"import a\n" +
						"  , {Foo} from 'module'\n" +
						"import b\n" +
						"  , {} from 'module'\n" +
						"import c\n" +
						"  ,{Bar} from 'module'\n" +
						"import d\n" +
						"  ,{} from 'module'",
				},
				Options: optWithExc("first", map[string]any{"ImportDeclaration": false}),
				Errors: errs(
					errSpec{"expectedCommaFirst", 5, 9},
					errSpec{"expectedCommaFirst", 7, 9},
				),
			},
			// ---- trailing comma at end of multi-line object — "last" ----
			{
				Code: "const x = {a,b\n" +
					",}",
				Output: []string{
					"const x = {a,b,\n" +
						"}",
				},
				Options: optStr("last"),
				Errors:  errLast(2, 1),
			},
			// ---- trailing comma at end of multi-line object — "first" ----
			{
				Code: "const x = {a,b,\n" +
					"}",
				Output: []string{
					"const x = {a,b\n" +
						",}",
				},
				Options: optStr("first"),
				Errors:  errFirst(1, 15),
			},
			// ---- complex array with parenthesized elements — "last" ----
			{
				Code: "const x = [,\n" +
					"  (a),\n" +
					"  (b),\n" +
					"  (c),\n" +
					"]\n" +
					"const y = [\n" +
					"  ,(a)\n" +
					"  ,(b)\n" +
					"  ,(c)\n" +
					"  ,]",
				Output: []string{
					"const x = [,\n" +
						"  (a),\n" +
						"  (b),\n" +
						"  (c),\n" +
						"]\n" +
						"const y = [\n" +
						"  ,(a),\n" +
						"  (b),\n" +
						"  (c),\n" +
						"  ]",
				},
				Options: optStr("last"),
				Errors: errs(
					errSpec{"expectedCommaLast", 8, 3},
					errSpec{"expectedCommaLast", 9, 3},
					errSpec{"expectedCommaLast", 10, 3},
				),
			},
			// ---- complex array with parenthesized elements — "first" ----
			{
				Code: "const x = [,\n" +
					"  (a),\n" +
					"  (b),\n" +
					"  (c),\n" +
					"]\n" +
					"const y = [\n" +
					"  ,(a)\n" +
					"  ,(b)\n" +
					"  ,(c)\n" +
					"  ,]",
				Output: []string{
					"const x = [,\n" +
						"  (a)\n" +
						"  ,(b)\n" +
						"  ,(c)\n" +
						",]\n" +
						"const y = [\n" +
						"  ,(a)\n" +
						"  ,(b)\n" +
						"  ,(c)\n" +
						"  ,]",
				},
				Options: optStr("first"),
				Errors: errs(
					errSpec{"expectedCommaFirst", 2, 6},
					errSpec{"expectedCommaFirst", 3, 6},
					errSpec{"expectedCommaFirst", 4, 6},
				),
			},
		},
	)
}
