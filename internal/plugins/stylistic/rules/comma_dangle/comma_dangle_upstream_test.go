// TestCommaDangleUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/rules/comma-dangle/comma-dangle._js_.test.ts
// and comma-dangle._ts_.test.ts 1:1. Position assertions cover line/column
// for every invalid case. rslint-specific lock-in cases (tsgo AST edge
// shapes, branch lock-ins, real-user shapes) live in
// comma_dangle_extras_test.go.
//
// Cases that depend on ESLint's `parserOptions.ecmaVersion` to flip the
// `functions` / `dynamicImports` slot to `'ignore'` (upstream's behavior
// for ecmaVersion < 2017 / 2025 respectively) are kept here as `Skip: true`
// with a `// SKIP: <reason>` marker — rslint does not model ecmaVersion,
// so we always behave as if `'latest'` were in effect.
package comma_dangle_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/comma_dangle"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func optStr(s string) []any { return []any{s} }

func optMap(m map[string]any) []any { return []any{m} }

func errUnexpected(line, col int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{
		{MessageId: "unexpected", Line: line, Column: col},
	}
}

func errMissing(line, col int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{
		{MessageId: "missing", Line: line, Column: col},
	}
}

func TestCommaDangleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&comma_dangle.CommaDangleRule,
		[]rule_tester.ValidTestCase{
			// ---- default (never) — array / object literals ----
			{Code: `var foo = { bar: 'baz' }`},
			{Code: "var foo = {\nbar: 'baz'\n}"},
			{Code: `var foo = [ 'baz' ]`},
			{Code: "var foo = [\n'baz'\n]"},
			{Code: `[,,]`},
			{Code: "[\n,\n,\n]"},
			{Code: `[,]`},
			{Code: "[\n,\n]"},
			{Code: `[]`},
			{Code: "[\n]"},
			{Code: "var foo = [\n      (bar ? baz : qux),\n    ];", Options: optStr("always-multiline")},

			// ---- never ----
			{Code: `var foo = { bar: 'baz' }`, Options: optStr("never")},
			{Code: "var foo = {\nbar: 'baz'\n}", Options: optStr("never")},
			{Code: `var foo = [ 'baz' ]`, Options: optStr("never")},
			{Code: `var { a, b } = foo;`, Options: optStr("never")},
			{Code: `var [ a, b ] = foo;`, Options: optStr("never")},
			{Code: "var { a,\n b, \n} = foo;", Options: optStr("only-multiline")},
			{Code: "var [ a,\n b, \n] = foo;", Options: optStr("only-multiline")},

			// ---- always ----
			{Code: `[(1),]`, Options: optStr("always")},
			{Code: `var x = { foo: (1),};`, Options: optStr("always")},
			{Code: `var foo = { bar: 'baz', }`, Options: optStr("always")},
			{Code: "var foo = {\nbar: 'baz',\n}", Options: optStr("always")},
			{Code: "var foo = {\nbar: 'baz'\n,}", Options: optStr("always")},
			{Code: `var foo = [ 'baz', ]`, Options: optStr("always")},
			{Code: "var foo = [\n'baz',\n]", Options: optStr("always")},
			{Code: "var foo = [\n'baz'\n,]", Options: optStr("always")},
			{Code: `[,,]`, Options: optStr("always")},
			{Code: "[\n,\n,\n]", Options: optStr("always")},
			{Code: `[,]`, Options: optStr("always")},
			{Code: "[\n,\n]", Options: optStr("always")},
			{Code: `[]`, Options: optStr("always")},
			{Code: "[\n]", Options: optStr("always")},

			// ---- always-multiline / only-multiline ----
			{Code: `var foo = { bar: 'baz' }`, Options: optStr("always-multiline")},
			{Code: `var foo = { bar: 'baz' }`, Options: optStr("only-multiline")},
			{Code: "var foo = {\nbar: 'baz',\n}", Options: optStr("always-multiline")},
			{Code: "var foo = {\nbar: 'baz',\n}", Options: optStr("only-multiline")},
			{Code: `var foo = [ 'baz' ]`, Options: optStr("always-multiline")},
			{Code: `var foo = [ 'baz' ]`, Options: optStr("only-multiline")},
			{Code: "var foo = [\n'baz',\n]", Options: optStr("always-multiline")},
			{Code: "var foo = [\n'baz',\n]", Options: optStr("only-multiline")},
			{Code: "var foo = { bar:\n\n'bar' }", Options: optStr("always-multiline")},
			{Code: "var foo = { bar:\n\n'bar' }", Options: optStr("only-multiline")},
			{Code: `var foo = {a: 1, b: 2, c: 3, d: 4}`, Options: optStr("always-multiline")},
			{Code: `var foo = {a: 1, b: 2, c: 3, d: 4}`, Options: optStr("only-multiline")},
			{Code: "var foo = {a: 1, b: 2,\n c: 3, d: 4}", Options: optStr("always-multiline")},
			{Code: "var foo = {a: 1, b: 2,\n c: 3, d: 4}", Options: optStr("only-multiline")},
			{Code: "var foo = {x: {\nfoo: 'bar',\n}}", Options: optStr("always-multiline")},
			{Code: "var foo = {x: {\nfoo: 'bar',\n}}", Options: optStr("only-multiline")},
			{Code: "var foo = new Map([\n[key, {\na: 1,\nb: 2,\nc: 3,\n}],\n])", Options: optStr("always-multiline")},
			{Code: "var foo = new Map([\n[key, {\na: 1,\nb: 2,\nc: 3,\n}],\n])", Options: optStr("only-multiline")},

			// ---- rest binding / spread (upstream issue #3627, #7297) ----
			{Code: `var [a, ...rest] = [];`, Options: optStr("always")},
			{Code: "var [\n    a,\n    ...rest\n] = [];", Options: optStr("always")},
			{Code: "var [\n    a,\n    ...rest\n] = [];", Options: optStr("always-multiline")},
			{Code: "var [\n    a,\n    ...rest\n] = [];", Options: optStr("only-multiline")},
			{Code: `[a, ...rest] = [];`, Options: optStr("always")},
			{Code: `for ([a, ...rest] of []);`, Options: optStr("always")},
			{Code: `var a = [b, ...spread,];`, Options: optStr("always")},
			{Code: `var {foo, ...bar} = baz`, Options: optStr("always")},

			// ---- import / export (upstream issue #3794) ----
			{Code: `import {foo,} from 'foo';`, Options: optStr("always")},
			{Code: `import foo from 'foo';`, Options: optStr("always")},
			{Code: `import foo, {abc,} from 'foo';`, Options: optStr("always")},
			{Code: `import * as foo from 'foo';`, Options: optStr("always")},
			{Code: `export {foo,} from 'foo';`, Options: optStr("always")},
			{Code: `import {foo} from 'foo';`, Options: optStr("never")},
			{Code: `import foo from 'foo';`, Options: optStr("never")},
			{Code: `import foo, {abc} from 'foo';`, Options: optStr("never")},
			{Code: `import * as foo from 'foo';`, Options: optStr("never")},
			{Code: `export {foo} from 'foo';`, Options: optStr("never")},
			{Code: `import {foo} from 'foo';`, Options: optStr("always-multiline")},
			{Code: `import {foo} from 'foo';`, Options: optStr("only-multiline")},
			{Code: `export {foo} from 'foo';`, Options: optStr("always-multiline")},
			{Code: `export {foo} from 'foo';`, Options: optStr("only-multiline")},
			{Code: "import {\n  foo,\n} from 'foo';", Options: optStr("always-multiline")},
			{Code: "import {\n  foo,\n} from 'foo';", Options: optStr("only-multiline")},
			{Code: "export {\n  foo,\n} from 'foo';", Options: optStr("always-multiline")},
			{Code: "export {\n  foo,\n} from 'foo';", Options: optStr("only-multiline")},
			{Code: "import {foo} from \n'foo';", Options: optStr("always-multiline")},
			{Code: "import {foo} from \n'foo';", Options: optStr("only-multiline")},

			// ---- functions ----
			// SKIP: rslint does not honor parserOptions.ecmaVersion; the upstream
			// rule sets `functions` to 'ignore' for ecmaVersion < 2017. With
			// 'latest' semantics rslint correctly enforces trailing-comma rules,
			// which would flip these cases from valid → invalid.
			{Code: `function foo(a) {}`, Options: optStr("always"), Skip: true},
			{Code: `foo(a)`, Options: optStr("always"), Skip: true},
			{Code: `function foo(a) {}`, Options: optStr("never")},
			{Code: `foo(a)`, Options: optStr("never")},
			{Code: "function foo(a,\nb) {}", Options: optStr("always-multiline")},
			{Code: "foo(a,\nb\n)", Options: optStr("always-multiline"), Skip: true},
			{Code: "function foo(a,\nb\n) {}", Options: optStr("always-multiline"), Skip: true},
			{Code: "foo(a,\nb)", Options: optStr("always-multiline")},
			{Code: "function foo(a,\nb) {}", Options: optStr("only-multiline")},
			{Code: "foo(a,\nb)", Options: optStr("only-multiline")},
			{Code: `function foo(a) {}`, Options: optStr("always"), Skip: true},
			{Code: `foo(a)`, Options: optStr("always"), Skip: true},
			{Code: `function foo(a) {}`, Options: optStr("never")},
			{Code: `foo(a)`, Options: optStr("never")},
			{Code: "function foo(a,\nb) {}", Options: optStr("always-multiline")},
			{Code: "foo(a,\nb)", Options: optStr("always-multiline")},
			// SKIP: ecmaVersion=7 ignores functions; rslint enforces (would report
			// missing trailing comma for the multi-line param list).
			{Code: "function foo(a,\nb\n) {}", Options: optStr("always-multiline"), Skip: true},
			{Code: "foo(a,\nb\n)", Options: optStr("always-multiline"), Skip: true},
			{Code: "function foo(a,\nb) {}", Options: optStr("only-multiline")},
			{Code: "foo(a,\nb)", Options: optStr("only-multiline")},

			// ---- functions (ES8 trailing comma allowed) ----
			{Code: `function foo(a) {}`, Options: optStr("never")},
			{Code: `foo(a)`, Options: optStr("never")},
			{Code: `function foo(a,) {}`, Options: optStr("always")},
			{Code: `foo(a,)`, Options: optStr("always")},
			{Code: "function foo(\na,\nb,\n) {}", Options: optStr("always-multiline")},
			{Code: "foo(\na,b)", Options: optStr("always-multiline")},
			{Code: `function foo(a,b) {}`, Options: optStr("always-multiline")},
			{Code: `foo(a,b)`, Options: optStr("always-multiline")},
			{Code: `function foo(a,b) {}`, Options: optStr("only-multiline")},
			{Code: `foo(a,b)`, Options: optStr("only-multiline")},

			// ---- functions slot in object form ----
			{Code: `function foo(a) {} `, Options: optMap(map[string]any{})},
			{Code: `foo(a)`, Options: optMap(map[string]any{})},
			{Code: `function foo(a) {} `, Options: optMap(map[string]any{"functions": "never"})},
			{Code: `foo(a)`, Options: optMap(map[string]any{"functions": "never"})},
			{Code: `function foo(a,) {}`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `function bar(a, ...b) {}`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `foo(a,)`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `foo(a,)`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `bar(...a,)`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `function foo(a) {} `, Options: optMap(map[string]any{"functions": "always-multiline"})},
			{Code: `foo(a)`, Options: optMap(map[string]any{"functions": "always-multiline"})},
			{Code: "function foo(\na,\nb,\n) {} ", Options: optMap(map[string]any{"functions": "always-multiline"})},
			{Code: "function foo(\na,\n...b\n) {} ", Options: optMap(map[string]any{"functions": "always-multiline"})},
			{Code: "foo(\na,\nb,\n)", Options: optMap(map[string]any{"functions": "always-multiline"})},
			{Code: "foo(\na,\n...b,\n)", Options: optMap(map[string]any{"functions": "always-multiline"})},
			{Code: `function foo(a) {} `, Options: optMap(map[string]any{"functions": "only-multiline"})},
			{Code: `foo(a)`, Options: optMap(map[string]any{"functions": "only-multiline"})},
			{Code: "function foo(\na,\nb,\n) {} ", Options: optMap(map[string]any{"functions": "only-multiline"})},
			{Code: "foo(\na,\nb,\n)", Options: optMap(map[string]any{"functions": "only-multiline"})},
			{Code: "function foo(\na,\nb\n) {} ", Options: optMap(map[string]any{"functions": "only-multiline"})},
			{Code: "foo(\na,\nb\n)", Options: optMap(map[string]any{"functions": "only-multiline"})},

			// ---- arrow function single arg without parens (upstream issue #158) ----
			{Code: `a => 42;`, Options: optStr("always")},

			// ---- dynamic import ----
			{Code: `import(source)`},
			{Code: `import(source, )`, Options: optStr("always")},
			{Code: `import(source, options, )`, Options: optStr("always")},
			// SKIP: ecmaVersion 15 means dynamicImports→'ignore' upstream; rslint enforces.
			{Code: `import(source)`, Options: optStr("always"), Skip: true},
			{Code: `import(source,)`, Options: optStr("always")},
			{Code: `import(source)`, Options: optStr("never")},
			{Code: `import(source, options)`, Options: optStr("never")},
			{Code: `import(source)`, Options: optStr("always-multiline")},
			{Code: `import(source, options)`, Options: optStr("always-multiline")},
			{Code: "import(\n  source,\n)", Options: optStr("always-multiline")},
			{Code: "import(\n  source,\n  options,\n)", Options: optStr("always-multiline")},
			{Code: `import(source)`, Options: optStr("only-multiline")},
			{Code: `import(source, options)`, Options: optStr("only-multiline")},
			{Code: "import(\n  source,\n)", Options: optStr("only-multiline")},
			{Code: "import(\n  source\n)", Options: optStr("only-multiline")},
			{Code: "import(\n  source,\n  options,\n)", Options: optStr("only-multiline")},
			{Code: "import(\n  source,\n  options\n)", Options: optStr("only-multiline")},
			{Code: `import(source)`, Options: optMap(map[string]any{"functions": "always"})},
			{Code: `import(source,)`, Options: optMap(map[string]any{"functions": "never", "dynamicImports": "always"})},

			// ---- import attributes ----
			{Code: `import foo from "foo" with {type: "json"}`},
			{Code: `import foo from "foo" with {type: "json",}`, Options: optStr("always")},
			{Code: `import foo from "foo" with {type: "json"}`, Options: optStr("never")},
			{Code: `import foo from "foo" with {type: "json"}`, Options: optStr("always-multiline")},
			{Code: "import foo from \"foo\" with {\n  type: \"json\",\n}", Options: optStr("always-multiline")},
			{Code: `import foo from "foo" with {type: "json"}`, Options: optStr("only-multiline")},
			{Code: `import foo from "foo" with {type: "json",}`, Options: optMap(map[string]any{"functions": "never", "importAttributes": "always"})},
			{Code: `export {foo} from "foo" with {type: "json"}`},
			{Code: `export {foo,} from "foo" with {type: "json",}`, Options: optStr("always")},
			{Code: `export {foo} from "foo" with {type: "json"}`, Options: optStr("never")},
			{Code: `export {foo} from "foo" with {type: "json"}`, Options: optStr("always-multiline")},
			{Code: "export {foo} from \"foo\" with {\n  type: \"json\",\n}", Options: optStr("always-multiline")},
			{Code: `export {foo} from "foo" with {type: "json"}`, Options: optStr("only-multiline")},
			{Code: `export {foo} from "foo" with {type: "json",}`, Options: optMap(map[string]any{"functions": "never", "importAttributes": "always"})},
			{Code: `export * from "foo" with {type: "json"}`},
			{Code: `export * from "foo" with {type: "json",}`, Options: optStr("always")},
			{Code: `export * from "foo" with {type: "json"}`, Options: optStr("never")},
			{Code: `export * from "foo" with {type: "json"}`, Options: optStr("always-multiline")},
			{Code: "export * from \"foo\" with {\n  type: \"json\",\n}", Options: optStr("always-multiline")},
			{Code: `export * from "foo" with {type: "json"}`, Options: optStr("only-multiline")},
			{Code: `export * from "foo" with {type: "json",}`, Options: optMap(map[string]any{"functions": "never", "importAttributes": "always"})},

			// ---- TS default ----
			{Code: `enum Foo {}`},
			{Code: "enum Foo {\n}"},
			{Code: `enum Foo {Bar}`},
			{Code: `function Foo<T>() {}`},
			{Code: `type Foo = []`},
			{Code: "type Foo = [\n]"},

			// ---- TS never ----
			{Code: `enum Foo {Bar}`, Options: optStr("never")},
			{Code: "enum Foo {Bar\n}", Options: optStr("never")},
			{Code: "enum Foo {Bar\n}", Options: optMap(map[string]any{"enums": "never"})},
			{Code: `function Foo<T>() {}`, Options: optStr("never")},
			{Code: "function Foo<T\n>() {}", Options: optStr("never")},
			{Code: "function Foo<T\n>() {}", Options: optMap(map[string]any{"generics": "never"})},
			{Code: `type Foo = [string]`, Options: optStr("never")},
			{Code: `type Foo = [string]`, Options: optMap(map[string]any{"tuples": "never"})},

			// ---- TS always ----
			{Code: `enum Foo {Bar,}`, Options: optStr("always")},
			{Code: "enum Foo {Bar,\n}", Options: optStr("always")},
			{Code: "enum Foo {Bar,\n}", Options: optMap(map[string]any{"enums": "always"})},
			{Code: `function Foo<T,>() {}`, Options: optStr("always")},
			{Code: "function Foo<T,\n>() {}", Options: optStr("always")},
			{Code: "function Foo<T,\n>() {}", Options: optMap(map[string]any{"generics": "always"})},
			{Code: `type Foo = [string,]`, Options: optStr("always")},
			{Code: "type Foo = [string,\n]", Options: optMap(map[string]any{"tuples": "always"})},

			// ---- TS always-multiline ----
			{Code: `enum Foo {Bar}`, Options: optStr("always-multiline")},
			{Code: "enum Foo {Bar,\n}", Options: optStr("always-multiline")},
			{Code: "enum Foo {Bar,\n}", Options: optMap(map[string]any{"enums": "always-multiline"})},
			{Code: `function Foo<T>() {}`, Options: optStr("always-multiline")},
			{Code: "function Foo<T,\n>() {}", Options: optStr("always-multiline")},
			{Code: "function Foo<T,\n>() {}", Options: optMap(map[string]any{"generics": "always-multiline"})},
			{Code: `type Foo = [string]`, Options: optStr("always-multiline")},
			{Code: "type Foo = [string,\n]", Options: optStr("always-multiline")},
			{Code: "type Foo = [string,\n]", Options: optMap(map[string]any{"tuples": "always-multiline"})},

			// ---- TS only-multiline ----
			{Code: `enum Foo {Bar}`, Options: optStr("only-multiline")},
			{Code: "enum Foo {Bar\n}", Options: optStr("only-multiline")},
			{Code: "enum Foo {Bar,\n}", Options: optStr("only-multiline")},
			{Code: "enum Foo {Bar,\n}", Options: optMap(map[string]any{"enums": "only-multiline"})},
			{Code: `function Foo<T>() {}`, Options: optStr("only-multiline")},
			{Code: "function Foo<T\n>() {}", Options: optStr("only-multiline")},
			{Code: "function Foo<T,\n>() {}", Options: optStr("only-multiline")},
			{Code: "function Foo<T\n>() {}", Options: optMap(map[string]any{"generics": "only-multiline"})},
			{Code: "function Foo<T,\n>() {}", Options: optMap(map[string]any{"generics": "only-multiline"})},
			{Code: "type Foo = [string\n]", Options: optMap(map[string]any{"tuples": "only-multiline"})},
			{Code: "type Foo = [string,\n]", Options: optMap(map[string]any{"tuples": "only-multiline"})},

			// ---- TS ignore ----
			{Code: `const a = <TYPE,>() => {}`, Options: optMap(map[string]any{"generics": "ignore"}), Tsx: true},

			// ---- TS each option independent ----
			{
				Code: "const Obj = { a: 1 };\nenum Foo {Bar}\nfunction Baz<T,>() {}\ntype Qux = [string,\n]",
				Options: optMap(map[string]any{
					"enums":    "never",
					"generics": "always",
					"tuples":   "always-multiline",
				}),
			},

			// ---- TSX <T,> disambiguation (upstream eslint-stylistic#35) ----
			{Code: `const id = <T,>(x: T) => x;`, Tsx: true},
			{Code: `const id = <T,R>(x: T) => x;`, Tsx: true},

			// ---- Babel/Flow type-annotation cases (upstream's `if (!skipBabel)`
			// loop block, /tmp/comma-dangle-js-test.ts L2202-2228). Migrated
			// here because the type annotations (`{a: string,}`, `: {b: boolean}`)
			// parse the same way under tsgo. The rule's listeners don't visit
			// TS TypeLiteral, so the trailing comma inside the type annotation
			// is never inspected — exactly upstream's behavior. ----
			{Code: `function foo({a}: {a: string,}) {}`, Options: optStr("never")},
			// SKIP: ecmaVersion=5 ignores functions slot upstream; rslint enforces.
			{Code: `function foo({a,}: {a: string}) {}`, Options: optStr("always"), Skip: true},
			{Code: `function foo(a): {b: boolean,} {}`, Options: optMap(map[string]any{"functions": "never"})},
			{Code: `function foo(a,): {b: boolean} {}`, Options: optMap(map[string]any{"functions": "always"})},
		},
		[]rule_tester.InvalidTestCase{
			// ---- default (never) — objects ----
			{
				Code:   `var foo = { bar: 'baz', }`,
				Output: []string{`var foo = { bar: 'baz' }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23, EndColumn: 24},
				},
			},
			{
				Code:   "var foo = {\nbar: 'baz',\n}",
				Output: []string{"var foo = {\nbar: 'baz'\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 11, EndColumn: 12},
				},
			},
			{
				Code:   `foo({ bar: 'baz', qux: 'quux', });`,
				Output: []string{`foo({ bar: 'baz', qux: 'quux' });`},
				Errors: errUnexpected(1, 30),
			},
			{
				Code:   "foo({\nbar: 'baz',\nqux: 'quux',\n});",
				Output: []string{"foo({\nbar: 'baz',\nqux: 'quux'\n});"},
				Errors: errUnexpected(3, 12),
			},
			{
				Code:   `var foo = [ 'baz', ]`,
				Output: []string{`var foo = [ 'baz' ]`},
				Errors: errUnexpected(1, 18),
			},
			{
				Code:   "var foo = [ 'baz',\n]",
				Output: []string{"var foo = [ 'baz'\n]"},
				Errors: errUnexpected(1, 18),
			},
			{
				Code:   "var foo = { bar: 'bar'\n\n, }",
				Output: []string{"var foo = { bar: 'bar'\n\n }"},
				Errors: errUnexpected(3, 1),
			},

			// ---- never (explicit) ----
			{Code: `var foo = { bar: 'baz', }`, Output: []string{`var foo = { bar: 'baz' }`}, Options: optStr("never"), Errors: errUnexpected(1, 23)},
			{Code: `var foo = { bar: 'baz', }`, Output: []string{`var foo = { bar: 'baz' }`}, Options: optStr("only-multiline"), Errors: errUnexpected(1, 23)},
			{Code: "var foo = {\nbar: 'baz',\n}", Output: []string{"var foo = {\nbar: 'baz'\n}"}, Options: optStr("never"), Errors: errUnexpected(2, 11)},
			{Code: `foo({ bar: 'baz', qux: 'quux', });`, Output: []string{`foo({ bar: 'baz', qux: 'quux' });`}, Options: optStr("never"), Errors: errUnexpected(1, 30)},
			{Code: `foo({ bar: 'baz', qux: 'quux', });`, Output: []string{`foo({ bar: 'baz', qux: 'quux' });`}, Options: optStr("only-multiline"), Errors: errUnexpected(1, 30)},

			// ---- always: missing ----
			{
				Code:   `var foo = { bar: 'baz' }`,
				Output: []string{`var foo = { bar: 'baz', }`},
				Options: optStr("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:   "var foo = {\nbar: 'baz'\n}",
				Output: []string{"var foo = {\nbar: 'baz',\n}"},
				Options: optStr("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 11, EndLine: 3, EndColumn: 1},
				},
			},
			{
				Code:   "var foo = {\nbar: 'baz'\r\n}",
				Output: []string{"var foo = {\nbar: 'baz',\r\n}"},
				Options: optStr("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 11, EndLine: 3, EndColumn: 1},
				},
			},
			// SKIP: upstream test uses ecmaVersion=5 so the outer `foo(...)` call
			// drops to functions='ignore' and only the inner object is reported.
			// rslint always uses 'latest', so the outer call is checked too and
			// receives a trailing comma — yielding `foo({...},)` instead of
			// `foo({...})`. Lock-in cases for this `'always'`-on-call shape live
			// in comma_dangle_extras_test.go.
			{
				Code:    `foo({ bar: 'baz', qux: 'quux' });`,
				Output:  []string{`foo({ bar: 'baz', qux: 'quux', });`},
				Options: optStr("always"),
				Skip:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 30, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "foo({\nbar: 'baz',\nqux: 'quux'\n});",
				Output:  []string{"foo({\nbar: 'baz',\nqux: 'quux',\n});"},
				Options: optStr("always"),
				Skip:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 3, Column: 12, EndLine: 4, EndColumn: 1},
				},
			},
			{
				Code:   `var foo = [ 'baz' ]`,
				Output: []string{`var foo = [ 'baz', ]`},
				Options: optStr("always"),
				Errors: errMissing(1, 18),
			},
			{
				Code:   `var foo = ['baz']`,
				Output: []string{`var foo = ['baz',]`},
				Options: optStr("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 17, EndColumn: 18},
				},
			},
			{
				Code:   "var foo = [ 'baz'\n]",
				Output: []string{"var foo = [ 'baz',\n]"},
				Options: optStr("always"),
				Errors: errMissing(1, 18),
			},
			{
				Code:   "var foo = { bar:\n\n'bar' }",
				Output: []string{"var foo = { bar:\n\n'bar', }"},
				Options: optStr("always"),
				Errors: errMissing(3, 6),
			},

			// ---- always-multiline ----
			{
				Code:   "var foo = {\nbar: 'baz'\n}",
				Output: []string{"var foo = {\nbar: 'baz',\n}"},
				Options: optStr("always-multiline"),
				Errors: errMissing(2, 11),
			},
			{
				Code:    "var foo = [\n  bar,\n  (\n    baz\n  )\n];",
				Output:  []string{"var foo = [\n  bar,\n  (\n    baz\n  ),\n];"},
				Options: optStr("always"),
				Errors:  errMissing(5, 4),
			},
			{
				Code:    "var foo = {\n  foo: 'bar',\n  baz: (\n    qux\n  )\n};",
				Output:  []string{"var foo = {\n  foo: 'bar',\n  baz: (\n    qux\n  ),\n};"},
				Options: optStr("always"),
				Errors:  errMissing(5, 4),
			},
			{
				// upstream issue #7291 — conditional inside parens
				Code:    "var foo = [\n  (bar\n    ? baz\n    : qux\n  )\n];",
				Output:  []string{"var foo = [\n  (bar\n    ? baz\n    : qux\n  ),\n];"},
				Options: optStr("always"),
				Errors:  errMissing(5, 4),
			},
			{Code: `var foo = { bar: 'baz', }`, Output: []string{`var foo = { bar: 'baz' }`}, Options: optStr("always-multiline"), Errors: errUnexpected(1, 23)},
			{Code: "foo({\nbar: 'baz',\nqux: 'quux'\n});", Output: []string{"foo({\nbar: 'baz',\nqux: 'quux',\n});"}, Options: optStr("always-multiline"), Errors: errMissing(3, 12)},
			{Code: `foo({ bar: 'baz', qux: 'quux', });`, Output: []string{`foo({ bar: 'baz', qux: 'quux' });`}, Options: optStr("always-multiline"), Errors: errUnexpected(1, 30)},
			{Code: "var foo = [\n'baz'\n]", Output: []string{"var foo = [\n'baz',\n]"}, Options: optStr("always-multiline"), Errors: errMissing(2, 6)},
			{Code: `var foo = ['baz',]`, Output: []string{`var foo = ['baz']`}, Options: optStr("always-multiline"), Errors: errUnexpected(1, 17)},
			{Code: `var foo = ['baz',]`, Output: []string{`var foo = ['baz']`}, Options: optStr("only-multiline"), Errors: errUnexpected(1, 17)},
			{Code: "var foo = {x: {\nfoo: 'bar',\n},}", Output: []string{"var foo = {x: {\nfoo: 'bar',\n}}"}, Options: optStr("always-multiline"), Errors: errUnexpected(3, 2)},
			{Code: "var foo = {a: 1, b: 2,\nc: 3, d: 4,}", Output: []string{"var foo = {a: 1, b: 2,\nc: 3, d: 4}"}, Options: optStr("always-multiline"), Errors: errUnexpected(2, 11)},
			{Code: "var foo = {a: 1, b: 2,\nc: 3, d: 4,}", Output: []string{"var foo = {a: 1, b: 2,\nc: 3, d: 4}"}, Options: optStr("only-multiline"), Errors: errUnexpected(2, 11)},
			{Code: "var foo = [{\na: 1,\nb: 2,\nc: 3,\nd: 4,\n},]", Output: []string{"var foo = [{\na: 1,\nb: 2,\nc: 3,\nd: 4,\n}]"}, Options: optStr("always-multiline"), Errors: errUnexpected(6, 2)},

			// ---- destructuring ----
			{Code: `var { a, b, } = foo;`, Output: []string{`var { a, b } = foo;`}, Options: optStr("never"), Errors: errUnexpected(1, 11)},
			{Code: `var { a, b, } = foo;`, Output: []string{`var { a, b } = foo;`}, Options: optStr("only-multiline"), Errors: errUnexpected(1, 11)},
			{Code: `var [ a, b, ] = foo;`, Output: []string{`var [ a, b ] = foo;`}, Options: optStr("never"), Errors: errUnexpected(1, 11)},
			{Code: `var [ a, b, ] = foo;`, Output: []string{`var [ a, b ] = foo;`}, Options: optStr("only-multiline"), Errors: errUnexpected(1, 11)},

			// ---- parenthesized element ----
			{Code: `[(1),]`, Output: []string{`[(1)]`}, Options: optStr("never"), Errors: errUnexpected(1, 5)},
			{Code: `[(1),]`, Output: []string{`[(1)]`}, Options: optStr("only-multiline"), Errors: errUnexpected(1, 5)},
			{Code: `var x = { foo: (1),};`, Output: []string{`var x = { foo: (1)};`}, Options: optStr("never"), Errors: errUnexpected(1, 19)},
			{Code: `var x = { foo: (1),};`, Output: []string{`var x = { foo: (1)};`}, Options: optStr("only-multiline"), Errors: errUnexpected(1, 19)},

			// ---- import / export (upstream issue #3794) ----
			{Code: `import {foo} from 'foo';`, Output: []string{`import {foo,} from 'foo';`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 12}}},
			{Code: `import foo, {abc} from 'foo';`, Output: []string{`import foo, {abc,} from 'foo';`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 17}}},
			{Code: `export {foo} from 'foo';`, Output: []string{`export {foo,} from 'foo';`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 12}}},
			{Code: `import {foo,} from 'foo';`, Output: []string{`import {foo} from 'foo';`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 12}}},
			{Code: `import {foo,} from 'foo';`, Output: []string{`import {foo} from 'foo';`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 12}}},
			{Code: `import foo, {abc,} from 'foo';`, Output: []string{`import foo, {abc} from 'foo';`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 17}}},
			{Code: `import foo, {abc,} from 'foo';`, Output: []string{`import foo, {abc} from 'foo';`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 17}}},
			{Code: `export {foo,} from 'foo';`, Output: []string{`export {foo} from 'foo';`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 12}}},
			{Code: `export {foo,} from 'foo';`, Output: []string{`export {foo} from 'foo';`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 12}}},
			{Code: `import {foo,} from 'foo';`, Output: []string{`import {foo} from 'foo';`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 12}}},
			{Code: `export {foo,} from 'foo';`, Output: []string{`export {foo} from 'foo';`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 12}}},
			{Code: "import {\n  foo\n} from 'foo';", Output: []string{"import {\n  foo,\n} from 'foo';"}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 2, Column: 6}}},
			{Code: "export {\n  foo\n} from 'foo';", Output: []string{"export {\n  foo,\n} from 'foo';"}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 2, Column: 6}}},

			// ---- parenthesized last element (upstream issue #6233) ----
			{Code: `var foo = {a: (1)}`, Output: []string{`var foo = {a: (1),}`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 18}}},
			{Code: `var foo = [(1)]`, Output: []string{`var foo = [(1),]`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}}},
			{Code: "var foo = [\n1,\n(2)\n]", Output: []string{"var foo = [\n1,\n(2),\n]"}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 3, Column: 4}}},

			// ---- functions: never (object form) ----
			{Code: `function foo(a,) {}`, Output: []string{`function foo(a) {}`}, Options: optMap(map[string]any{"functions": "never"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `(function foo(a,) {})`, Output: []string{`(function foo(a) {})`}, Options: optMap(map[string]any{"functions": "never"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}}},
			{Code: `(a,) => a`, Output: []string{`(a) => a`}, Options: optMap(map[string]any{"functions": "never"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 3}}},
			{Code: `(a,) => (a)`, Output: []string{`(a) => (a)`}, Options: optMap(map[string]any{"functions": "never"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 3}}},
			{Code: `({foo(a,) {}})`, Output: []string{`({foo(a) {}})`}, Options: optMap(map[string]any{"functions": "never"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 8}}},
			{Code: `class A {foo(a,) {}}`, Output: []string{`class A {foo(a) {}}`}, Options: optMap(map[string]any{"functions": "never"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `foo(a,)`, Output: []string{`foo(a)`}, Options: optMap(map[string]any{"functions": "never"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 6}}},
			{Code: `foo(...a,)`, Output: []string{`foo(...a)`}, Options: optMap(map[string]any{"functions": "never"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 9}}},

			// ---- functions: always (object form) ----
			{Code: `function foo(a) {}`, Output: []string{`function foo(a,) {}`}, Options: optMap(map[string]any{"functions": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}}},
			{Code: `(function foo(a) {})`, Output: []string{`(function foo(a,) {})`}, Options: optMap(map[string]any{"functions": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 16}}},
			{Code: `(a) => a`, Output: []string{`(a,) => a`}, Options: optMap(map[string]any{"functions": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 3}}},
			{Code: `(a) => (a)`, Output: []string{`(a,) => (a)`}, Options: optMap(map[string]any{"functions": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 3}}},
			{Code: `({foo(a) {}})`, Output: []string{`({foo(a,) {}})`}, Options: optMap(map[string]any{"functions": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 8}}},
			{Code: `class A {foo(a) {}}`, Output: []string{`class A {foo(a,) {}}`}, Options: optMap(map[string]any{"functions": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}}},
			{Code: `foo(a)`, Output: []string{`foo(a,)`}, Options: optMap(map[string]any{"functions": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 6}}},
			{Code: `foo(...a)`, Output: []string{`foo(...a,)`}, Options: optMap(map[string]any{"functions": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 9}}},

			// ---- functions: always-multiline (object form) ----
			{Code: `function foo(a,) {}`, Output: []string{`function foo(a) {}`}, Options: optMap(map[string]any{"functions": "always-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `(function foo(a,) {})`, Output: []string{`(function foo(a) {})`}, Options: optMap(map[string]any{"functions": "always-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}}},
			{Code: `foo(a,)`, Output: []string{`foo(a)`}, Options: optMap(map[string]any{"functions": "always-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 6}}},
			{Code: `foo(...a,)`, Output: []string{`foo(...a)`}, Options: optMap(map[string]any{"functions": "always-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 9}}},
			{Code: "function foo(\na,\nb\n) {}", Output: []string{"function foo(\na,\nb,\n) {}"}, Options: optMap(map[string]any{"functions": "always-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 3, Column: 2}}},
			{Code: "foo(\na,\nb\n)", Output: []string{"foo(\na,\nb,\n)"}, Options: optMap(map[string]any{"functions": "always-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 3, Column: 2}}},
			{Code: "foo(\n...a,\n...b\n)", Output: []string{"foo(\n...a,\n...b,\n)"}, Options: optMap(map[string]any{"functions": "always-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 3, Column: 5}}},

			// ---- functions: only-multiline (object form) ----
			{Code: `function foo(a,) {}`, Output: []string{`function foo(a) {}`}, Options: optMap(map[string]any{"functions": "only-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `(function foo(a,) {})`, Output: []string{`(function foo(a) {})`}, Options: optMap(map[string]any{"functions": "only-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}}},
			{Code: `foo(a,)`, Output: []string{`foo(a)`}, Options: optMap(map[string]any{"functions": "only-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 6}}},
			{Code: `foo(...a,)`, Output: []string{`foo(...a)`}, Options: optMap(map[string]any{"functions": "only-multiline"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 9}}},

			// ---- functions: never (string form) ----
			{Code: `function foo(a,) {}`, Output: []string{`function foo(a) {}`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `(function foo(a,) {})`, Output: []string{`(function foo(a) {})`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}}},
			{Code: `(a,) => a`, Output: []string{`(a) => a`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 3}}},
			{Code: `(a,) => (a)`, Output: []string{`(a) => (a)`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 3}}},
			{Code: `({foo(a,) {}})`, Output: []string{`({foo(a) {}})`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 8}}},
			{Code: `class A {foo(a,) {}}`, Output: []string{`class A {foo(a) {}}`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `foo(a,)`, Output: []string{`foo(a)`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 6}}},
			{Code: `foo(...a,)`, Output: []string{`foo(...a)`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 9}}},

			// ---- functions: always (string form) ----
			{Code: `function foo(a) {}`, Output: []string{`function foo(a,) {}`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}}},
			{Code: `(function foo(a) {})`, Output: []string{`(function foo(a,) {})`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 16}}},
			{Code: `(a) => a`, Output: []string{`(a,) => a`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 3}}},
			{Code: `(a) => (a)`, Output: []string{`(a,) => (a)`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 3}}},
			// Note: `'always'` with `({foo(a) {}})` emits TWO missing diagnostics — one
			// for the outer object's properties and one for the inner method's params.
			// rslint reports them in listener-fire order (outer parent first, then
			// inner child), so the order is [col 12, col 8] rather than source order
			// [col 8, col 12]. Upstream doesn't assert position for this case.
			{
				Code:    `({foo(a) {}})`,
				Output:  []string{`({foo(a,) {},})`},
				Options: optStr("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 12},
					{MessageId: "missing", Line: 1, Column: 8},
				},
			},
			{Code: `class A {foo(a) {}}`, Output: []string{`class A {foo(a,) {}}`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}}},
			{Code: `foo(a)`, Output: []string{`foo(a,)`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 6}}},
			{Code: `foo(...a)`, Output: []string{`foo(...a,)`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 9}}},

			// ---- functions: always-multiline (string form) ----
			{Code: `function foo(a,) {}`, Output: []string{`function foo(a) {}`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `(function foo(a,) {})`, Output: []string{`(function foo(a) {})`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}}},
			{Code: `foo(a,)`, Output: []string{`foo(a)`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 6}}},
			{Code: `foo(...a,)`, Output: []string{`foo(...a)`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 9}}},
			{Code: "function foo(\na,\nb\n) {}", Output: []string{"function foo(\na,\nb,\n) {}"}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 3, Column: 2}}},
			{Code: "foo(\na,\nb\n)", Output: []string{"foo(\na,\nb,\n)"}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 3, Column: 2}}},
			{Code: "foo(\n...a,\n...b\n)", Output: []string{"foo(\n...a,\n...b,\n)"}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 3, Column: 5}}},

			// ---- functions: only-multiline (string form) ----
			{Code: `function foo(a,) {}`, Output: []string{`function foo(a) {}`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `(function foo(a,) {})`, Output: []string{`(function foo(a) {})`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}}},
			{Code: `foo(a,)`, Output: []string{`foo(a)`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 6}}},
			{Code: `foo(...a,)`, Output: []string{`foo(...a)`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 9}}},
			{Code: `function foo(a) {}`, Output: []string{`function foo(a,) {}`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}}},

			// ---- separated options (per-slot interaction) ----
			{
				Code: "let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from \"foo\";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);",
				Output: []string{
					"let {a} = {a: 1};\nlet [b,] = [1,];\nimport {c,} from \"foo\";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);",
				},
				Options: optMap(map[string]any{
					"objects":   "never",
					"arrays":    "ignore",
					"imports":   "ignore",
					"exports":   "ignore",
					"functions": "ignore",
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1},
					{MessageId: "unexpected", Line: 1},
				},
			},
			{
				Code: "let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from \"foo\";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);",
				Output: []string{
					"let {a,} = {a: 1,};\nlet [b] = [1];\nimport {c,} from \"foo\";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);",
				},
				Options: optMap(map[string]any{
					"objects":   "ignore",
					"arrays":    "never",
					"imports":   "ignore",
					"exports":   "ignore",
					"functions": "ignore",
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2},
					{MessageId: "unexpected", Line: 2},
				},
			},
			{
				Code: "let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from \"foo\";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);",
				Output: []string{
					"let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c} from \"foo\";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);",
				},
				Options: optMap(map[string]any{
					"objects":   "ignore",
					"arrays":    "ignore",
					"imports":   "never",
					"exports":   "ignore",
					"functions": "ignore",
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 3},
				},
			},
			{
				Code: "let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from \"foo\";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);",
				Output: []string{
					"let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from \"foo\";\nlet d = 0;export {d};\n(function foo(e,) {})(f,);",
				},
				Options: optMap(map[string]any{
					"objects":   "ignore",
					"arrays":    "ignore",
					"imports":   "ignore",
					"exports":   "never",
					"functions": "ignore",
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 4},
				},
			},
			{
				Code: "let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from \"foo\";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);",
				Output: []string{
					"let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from \"foo\";\nlet d = 0;export {d,};\n(function foo(e) {})(f);",
				},
				Options: optMap(map[string]any{
					"objects":   "ignore",
					"arrays":    "ignore",
					"imports":   "ignore",
					"exports":   "ignore",
					"functions": "never",
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 5},
					{MessageId: "unexpected", Line: 5},
				},
			},

			// ---- default with trailing-comma call (upstream issue #11502) ----
			{
				Code:   `foo(a,)`,
				Output: []string{`foo(a)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 6}},
			},

			// ---- upstream issue #15660 — multi-line import with trailing-comma toggle ----
			{
				Code: "/*eslint add-named-import:1*/\nimport {\n    StyleSheet,\n    View,\n    TextInput,\n    ImageBackground,\n    Image,\n    TouchableOpacity,\n    SafeAreaView\n} from 'react-native';",
				Output: []string{
					"/*eslint add-named-import:1*/\nimport {\n    StyleSheet,\n    View,\n    TextInput,\n    ImageBackground,\n    Image,\n    TouchableOpacity,\n    SafeAreaView,\n} from 'react-native';",
				},
				Options: optMap(map[string]any{"imports": "always-multiline"}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 9, Column: 17},
				},
			},
			{
				Code: "/*eslint add-named-import:1*/\nimport {\n    StyleSheet,\n    View,\n    TextInput,\n    ImageBackground,\n    Image,\n    TouchableOpacity,\n    SafeAreaView,\n} from 'react-native';",
				Output: []string{
					"/*eslint add-named-import:1*/\nimport {\n    StyleSheet,\n    View,\n    TextInput,\n    ImageBackground,\n    Image,\n    TouchableOpacity,\n    SafeAreaView\n} from 'react-native';",
				},
				Options: optMap(map[string]any{"imports": "never"}),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 9, Column: 17},
				},
			},

			// ---- dynamic import ----
			{
				Code:   `import(source,)`,
				Output: []string{`import(source)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}},
			},
			{
				Code:    `import(source)`,
				Output:  []string{`import(source,)`},
				Options: optStr("always"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 14}},
			},
			{
				Code:    `import(source, options)`,
				Output:  []string{`import(source, options,)`},
				Options: optStr("always"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 23}},
			},
			{
				Code:    `import(source,)`,
				Output:  []string{`import(source)`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}},
			},
			{
				Code:    `import(source, options,)`,
				Output:  []string{`import(source, options)`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 23}},
			},
			{
				Code:    `import(source,)`,
				Output:  []string{`import(source)`},
				Options: optStr("always-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}},
			},
			{
				Code:    `import(source, options,)`,
				Output:  []string{`import(source, options)`},
				Options: optStr("always-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 23}},
			},
			{
				Code:    "import(\n  source\n)",
				Output:  []string{"import(\n  source,\n)"},
				Options: optStr("always-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 2, Column: 9}},
			},
			{
				Code:    "import(\n  source,\n  options\n)",
				Output:  []string{"import(\n  source,\n  options,\n)"},
				Options: optStr("always-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 3, Column: 10}},
			},
			{
				Code:    `import(source,)`,
				Output:  []string{`import(source)`},
				Options: optStr("only-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}},
			},
			{
				Code:    `import(source, options,)`,
				Output:  []string{`import(source, options)`},
				Options: optStr("only-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 23}},
			},
			{
				Code:    `import(source)`,
				Output:  []string{`import(source,)`},
				Options: optMap(map[string]any{"functions": "never", "dynamicImports": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 14}},
			},

			// ---- import attributes (invalid) ----
			{Code: `import foo from "foo" with {type: "json",}`, Output: []string{`import foo from "foo" with {type: "json"}`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 41}}},
			{Code: `import foo from "foo" with {type: "json"}`, Output: []string{`import foo from "foo" with {type: "json",}`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 41}}},
			{Code: `import foo from "foo" with {type: "json",}`, Output: []string{`import foo from "foo" with {type: "json"}`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 41}}},
			{Code: `import foo from "foo" with {type: "json",}`, Output: []string{`import foo from "foo" with {type: "json"}`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 41}}},
			{
				Code:    "import foo from \"foo\" with {\n  type: \"json\"\n}",
				Output:  []string{"import foo from \"foo\" with {\n  type: \"json\",\n}"},
				Options: optStr("always-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 2, Column: 15}},
			},
			{Code: `import foo from "foo" with {type: "json",}`, Output: []string{`import foo from "foo" with {type: "json"}`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 41}}},
			{Code: `import foo from "foo" with {type: "json"}`, Output: []string{`import foo from "foo" with {type: "json",}`}, Options: optMap(map[string]any{"functions": "never", "importAttributes": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 41}}},
			{Code: `export {foo} from "foo" with {type: "json",}`, Output: []string{`export {foo} from "foo" with {type: "json"}`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 43}}},
			{
				Code:    `export {foo} from "foo" with {type: "json"}`,
				Output:  []string{`export {foo,} from "foo" with {type: "json",}`},
				Options: optStr("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 12},
					{MessageId: "missing", Line: 1, Column: 43},
				},
			},
			{Code: `export {foo} from "foo" with {type: "json",}`, Output: []string{`export {foo} from "foo" with {type: "json"}`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 43}}},
			{Code: `export {foo} from "foo" with {type: "json",}`, Output: []string{`export {foo} from "foo" with {type: "json"}`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 43}}},
			{
				Code:    "export {foo} from \"foo\" with {\n  type: \"json\"\n}",
				Output:  []string{"export {foo} from \"foo\" with {\n  type: \"json\",\n}"},
				Options: optStr("always-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 2, Column: 15}},
			},
			{Code: `export {foo} from "foo" with {type: "json",}`, Output: []string{`export {foo} from "foo" with {type: "json"}`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 43}}},
			{Code: `export {foo} from "foo" with {type: "json"}`, Output: []string{`export {foo} from "foo" with {type: "json",}`}, Options: optMap(map[string]any{"functions": "never", "importAttributes": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 43}}},
			{Code: `export * from "foo" with {type: "json",}`, Output: []string{`export * from "foo" with {type: "json"}`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 39}}},
			{Code: `export * from "foo" with {type: "json"}`, Output: []string{`export * from "foo" with {type: "json",}`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 39}}},
			{Code: `export * from "foo" with {type: "json",}`, Output: []string{`export * from "foo" with {type: "json"}`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 39}}},
			{Code: `export * from "foo" with {type: "json",}`, Output: []string{`export * from "foo" with {type: "json"}`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 39}}},
			{
				Code:    "export * from \"foo\" with {\n  type: \"json\"\n}",
				Output:  []string{"export * from \"foo\" with {\n  type: \"json\",\n}"},
				Options: optStr("always-multiline"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 2, Column: 15}},
			},
			{Code: `export * from "foo" with {type: "json",}`, Output: []string{`export * from "foo" with {type: "json"}`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 39}}},
			{Code: `export * from "foo" with {type: "json"}`, Output: []string{`export * from "foo" with {type: "json",}`}, Options: optMap(map[string]any{"functions": "never", "importAttributes": "always"}), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 39}}},

			// ---- TS default (invalid) ----
			{Code: `enum Foo {Bar,}`, Output: []string{`enum Foo {Bar}`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}}},
			{Code: `function Foo<T,>() {}`, Output: []string{`function Foo<T>() {}`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `type Foo = [string,]`, Output: []string{`type Foo = [string]`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 19}}},

			// ---- TS never (invalid) ----
			{Code: `enum Foo {Bar,}`, Output: []string{`enum Foo {Bar}`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}}},
			{Code: "enum Foo {Bar,\n}", Output: []string{"enum Foo {Bar\n}"}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}}},
			{Code: `function Foo<T,>() {}`, Output: []string{`function Foo<T>() {}`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: "function Foo<T,\n>() {}", Output: []string{"function Foo<T\n>() {}"}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `type Foo = [string,]`, Output: []string{`type Foo = [string]`}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 19}}},
			{Code: "type Foo = [string,\n]", Output: []string{"type Foo = [string\n]"}, Options: optStr("never"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 19}}},

			// ---- TS always (invalid) ----
			{Code: `enum Foo {Bar}`, Output: []string{`enum Foo {Bar,}`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 14}}},
			{Code: "enum Foo {Bar\n}", Output: []string{"enum Foo {Bar,\n}"}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 14}}},
			{Code: `function Foo<T>() {}`, Output: []string{`function Foo<T,>() {}`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}}},
			{Code: "function Foo<T\n>() {}", Output: []string{"function Foo<T,\n>() {}"}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}}},
			{Code: `type Foo = [string]`, Output: []string{`type Foo = [string,]`}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 19}}},
			{Code: "type Foo = [string\n]", Output: []string{"type Foo = [string,\n]"}, Options: optStr("always"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 19}}},

			// ---- TS always-multiline (invalid) ----
			{Code: `enum Foo {Bar,}`, Output: []string{`enum Foo {Bar}`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}}},
			{Code: "enum Foo {Bar\n}", Output: []string{"enum Foo {Bar,\n}"}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 14}}},
			{Code: `function Foo<T,>() {}`, Output: []string{`function Foo<T>() {}`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: "function Foo<T\n>() {}", Output: []string{"function Foo<T,\n>() {}"}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}}},
			{Code: `type Foo = [string,]`, Output: []string{`type Foo = [string]`}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 19}}},
			{Code: "type Foo = [string\n]", Output: []string{"type Foo = [string,\n]"}, Options: optStr("always-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 19}}},

			// ---- TS only-multiline (invalid) ----
			{Code: `enum Foo {Bar,}`, Output: []string{`enum Foo {Bar}`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 14}}},
			{Code: `function Foo<T,>() {}`, Output: []string{`function Foo<T>() {}`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}}},
			{Code: `type Foo = [string,]`, Output: []string{`type Foo = [string]`}, Options: optStr("only-multiline"), Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 19}}},

			// ---- TSX multi-generic (eslint-stylistic#35) ----
			{
				Code:   `const id = <T,R,>(x: T) => x;`,
				Output: []string{`const id = <T,R>(x: T) => x;`},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}},
			},

			// ---- TS declare function / TSFunctionType ----
			{
				Code:   `declare function foo(a: number, b: number,): void`,
				Output: []string{`declare function foo(a: number, b: number): void`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 42}},
			},
			{
				Code:   `type Foo = (a: number, b: number,) => void`,
				Output: []string{`type Foo = (a: number, b: number) => void`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 33}},
			},

			// ---- TS type-argument instantiation (predicate.never) ----
			{
				Code:   `type Foo<T> = Bar<T,>`,
				Output: []string{`type Foo<T> = Bar<T>`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 20}},
			},

			// ---- Babel/Flow type-annotation cases (upstream invalid section
			// in the `if (!skipBabel)` loop, /tmp/comma-dangle-js-test.ts
			// L2229-2262). ----
			// SKIP: upstream uses ecmaVersion=5 to make `functions` slot 'ignore',
			// so only the inner ObjectBindingPattern `{a}` is reported. rslint
			// uses 'latest' and would additionally report the outer Parameters
			// list, mismatching the expected single-error count.
			{
				Code:    `function foo({a}: {a: string,}) {}`,
				Output:  []string{`function foo({a,}: {a: string,}) {}`},
				Options: optStr("always"),
				Skip:    true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 16}},
			},
			{
				Code:    `function foo({a,}: {a: string}) {}`,
				Output:  []string{`function foo({a}: {a: string}) {}`},
				Options: optStr("never"),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 16}},
			},
			{
				Code:    `function foo(a): {b: boolean,} {}`,
				Output:  []string{`function foo(a,): {b: boolean,} {}`},
				Options: optMap(map[string]any{"functions": "always"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missing", Line: 1, Column: 15}},
			},
			{
				Code:    `function foo(a,): {b: boolean} {}`,
				Output:  []string{`function foo(a): {b: boolean} {}`},
				Options: optMap(map[string]any{"functions": "never"}),
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpected", Line: 1, Column: 15}},
			},
		},
	)
}
