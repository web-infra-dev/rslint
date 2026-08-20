// TestIdDenylistUpstream migrates the full valid/invalid suite from upstream
// ESLint v10.8.1 tests/lib/rules/id-denylist.js 1:1. Position assertions cover
// line/column for every invalid case. Every case keeps the ECMAScript edition
// it was authored with (the suite default is 5), because the edition decides
// which names count as language globals and this rule leaves those alone.
// languageOptions.sourceType is dropped: rslint infers module-ness from actual
// import/export syntax, which every case that sets it already has.
// rslint-specific lock-in cases live in id_denylist_extras_test.go.
// cspell:ignore bingg mydate myarray
package id_denylist

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// The ECMAScript editions the upstream suite selects, as the four-digit years
// the rule tester takes. The edition picks the set
// of language globals, which this rule never reports.
var (
	es5    = rule.LanguageOptions{ECMAVersion: 5}
	es2015 = rule.LanguageOptions{ECMAVersion: 2015}
	es2018 = rule.LanguageOptions{ECMAVersion: 2018}
	es2020 = rule.LanguageOptions{ECMAVersion: 2020}
	es2022 = rule.LanguageOptions{ECMAVersion: 2022}
	es2025 = rule.LanguageOptions{ECMAVersion: 2025}
)

// deny builds the rule's positional options: one denied name per element.
func deny(names ...string) []any {
	options := make([]any, len(names))
	for i, name := range names {
		options[i] = name
	}
	return options
}

func restricted(name string, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "restricted",
		Message:   fmt.Sprintf("Identifier '%s' is restricted.", name),
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len(name),
	}
}

func restrictedPrivate(name string, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "restrictedPrivate",
		Message:   fmt.Sprintf("Identifier '#%s' is restricted.", name),
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len(name) + 1,
	}
}

func TestIdDenylistUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&IdDenylistRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid: identifiers outside the deny list, member reads, destructuring keys ----
			{Code: `foo = "bar"`, Options: deny("bar"), LanguageOptions: es5},
			{Code: `bar = "bar"`, Options: deny("foo"), LanguageOptions: es5},
			{Code: `foo = "bar"`, Options: deny("f", "fo", "fooo", "bar"), LanguageOptions: es5},
			{Code: `function foo(){}`, Options: deny("bar"), LanguageOptions: es5},
			{Code: `foo()`, Options: deny("f", "fo", "fooo", "bar"), LanguageOptions: es5},
			{Code: `import { foo as bar } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015},
			{Code: `export { foo as bar } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015},
			{Code: `foo.bar()`, Options: deny("f", "fo", "fooo", "b", "ba", "baz"), LanguageOptions: es5},
			{Code: `var foo = bar.baz;`, Options: deny("f", "fo", "fooo", "b", "ba", "barr", "bazz"), LanguageOptions: es5},
			{Code: `var foo = bar.baz.bing;`, Options: deny("f", "fo", "fooo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5},
			{Code: `foo.bar.baz = bing.bong.bash;`, Options: deny("f", "fo", "fooo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5},
			{Code: `if (foo.bar) {}`, Options: deny("f", "fo", "fooo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5},
			{Code: `var obj = { key: foo.bar };`, Options: deny("f", "fo", "fooo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5},
			{Code: `const {foo: bar} = baz`, Options: deny("foo"), LanguageOptions: es2015},
			{Code: `const {foo: {bar: baz}} = qux`, Options: deny("foo", "bar"), LanguageOptions: es2015},
			{Code: `function foo({ bar: baz }) {}`, Options: deny("bar"), LanguageOptions: es2015},
			{Code: `function foo({ bar: {baz: qux} }) {}`, Options: deny("bar", "baz"), LanguageOptions: es2015},
			{Code: `function foo({baz} = obj.qux) {}`, Options: deny("qux"), LanguageOptions: es2015},
			{Code: `function foo({ foo: {baz} = obj.qux }) {}`, Options: deny("qux"), LanguageOptions: es2015},
			{Code: `({a: bar = obj.baz});`, Options: deny("baz"), LanguageOptions: es2015},
			{Code: `({foo: {a: bar = obj.baz}} = qux);`, Options: deny("baz"), LanguageOptions: es2015},
			{Code: `var arr = [foo.bar];`, Options: deny("f", "fo", "fooo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5},
			{Code: `[foo.bar]`, Options: deny("f", "fo", "fooo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5},
			{Code: `[foo.bar.nesting]`, Options: deny("f", "fo", "fooo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5},
			{Code: `if (foo.bar === bar.baz) { [foo.bar] }`, Options: deny("f", "fo", "fooo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5},
			{Code: `var myArray = new Array(); var myDate = new Date();`, Options: deny("array", "date", "mydate", "myarray", "new", "var"), LanguageOptions: es5},
			{Code: `foo()`, Options: deny("foo"), LanguageOptions: es5},
			{Code: `foo.bar()`, Options: deny("bar"), LanguageOptions: es5},
			{Code: `foo.bar`, Options: deny("bar"), LanguageOptions: es5},
			{Code: `({foo: obj.bar.bar.bar.baz} = {});`, Options: deny("foo", "bar"), LanguageOptions: es2015},
			{Code: `({[obj.bar]: a = baz} = qux);`, Options: deny("bar"), LanguageOptions: es2015},

			// ---- upstream valid: references to global variables ----
			{Code: `Number.parseInt()`, Options: deny("Number"), LanguageOptions: es5},
			{Code: `x = Number.NaN;`, Options: deny("Number"), LanguageOptions: es5},
			{Code: `var foo = undefined;`, Options: deny("undefined"), LanguageOptions: es5},
			{Code: `if (foo === undefined);`, Options: deny("undefined"), LanguageOptions: es5},
			{Code: `obj[undefined] = 5;`, Options: deny("undefined"), LanguageOptions: es5},
			{Code: `foo = { [myGlobal]: 1 };`, Options: deny("myGlobal"), LanguageOptions: es2015, Globals: map[string]any{"myGlobal": "readonly"}},
			{Code: `({ myGlobal } = foo);`, Options: deny("myGlobal"), LanguageOptions: es2015, Globals: map[string]any{"myGlobal": "writable"}},
			{Code: `/* global myGlobal: readonly */ myGlobal = 5;`, Options: deny("myGlobal"), LanguageOptions: es5},
			{Code: `var foo = [Map];`, Options: deny("Map"), LanguageOptions: es2015},
			{Code: `var foo = { bar: window.baz };`, Options: deny("window"), LanguageOptions: es5, Globals: map[string]any{"window": "readonly"}},

			// ---- upstream valid: class fields ----
			{Code: `class C { camelCase; #camelCase; #camelCase2() {} }`, Options: deny("foo"), LanguageOptions: es2022},
			{Code: `class C { snake_case; #snake_case; #snake_case2() {} }`, Options: deny("foo"), LanguageOptions: es2022},

			// ---- upstream valid: meta-properties ----
			{Code: `import.meta`, Options: deny("import", "meta"), LanguageOptions: es2020},
			{Code: `function foo() { new.target; }`, Options: deny("new", "target"), LanguageOptions: es2015},

			// ---- upstream valid: import attribute keys ----
			{Code: `import foo from 'foo.json' with { type: 'json' }`, Options: deny("type"), LanguageOptions: es2025},
			{Code: `export * from 'foo.json' with { type: 'json' }`, Options: deny("type"), LanguageOptions: es2025},
			{Code: `export { default } from 'foo.json' with { type: 'json' }`, Options: deny("type"), LanguageOptions: es2025},
			{Code: `import('foo.json', { with: { type: 'json' } })`, Options: deny("with", "type"), LanguageOptions: es2025},
			{Code: `import('foo.json', { 'with': { type: 'json' } })`, Options: deny("type"), LanguageOptions: es2025},
			{Code: `import('foo.json', { with: { type } })`, Options: deny("type"), LanguageOptions: es2025},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid: declarations, imports/exports, member writes, destructuring ----
			{Code: `foo = "bar"`, Options: deny("foo"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 1)}},
			{Code: `bar = "bar"`, Options: deny("bar"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 1)}},
			{Code: `foo = "bar"`, Options: deny("f", "fo", "foo", "bar"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 1)}},
			{Code: `function foo(){}`, Options: deny("f", "fo", "foo", "bar"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 10)}},
			{Code: `import foo from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 8)}},
			{Code: `import * as foo from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 13)}},
			{Code: `export * as foo from 'mod'`, Options: deny("foo"), LanguageOptions: es2020, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 13)}},
			{Code: `import { foo } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 10)}},
			{Code: `import { foo as bar } from 'mod'`, Options: deny("bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 17)}},
			{Code: `import { foo as bar } from 'mod'`, Options: deny("foo", "bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 17)}},
			{Code: `import { foo as foo } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 17)}},
			{Code: `import { foo, foo as bar } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 10)}},
			{Code: `import { foo as bar, foo } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 22)}},
			{Code: `import foo, { foo as bar } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 8)}},
			{Code: `var foo; export { foo as bar };`, Options: deny("bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 26)}},
			{Code: `var foo; export { foo };`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 5), restricted("foo", 1, 19)}},
			{Code: `var foo; export { foo as bar };`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 5), restricted("foo", 1, 19)}},
			{Code: `var foo; export { foo as foo };`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 5), restricted("foo", 1, 19), restricted("foo", 1, 26)}},
			{Code: `var foo; export { foo as bar };`, Options: deny("foo", "bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 5), restricted("foo", 1, 19), restricted("bar", 1, 26)}},
			{Code: `export { foo } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 10)}},
			{Code: `export { foo as bar } from 'mod'`, Options: deny("bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 17)}},
			{Code: `export { foo as bar } from 'mod'`, Options: deny("foo", "bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 17)}},
			{Code: `export { foo as foo } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 17)}},
			{Code: `export { foo, foo as bar } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 10)}},
			{Code: `export { foo as bar, foo } from 'mod'`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 22)}},
			{Code: `foo.bar()`, Options: deny("f", "fo", "foo", "b", "ba", "baz"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 1)}},
			{Code: `foo[bar] = baz;`, Options: deny("bar"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 5)}},
			{Code: `baz = foo[bar];`, Options: deny("bar"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 11)}},
			{Code: `var foo = bar.baz;`, Options: deny("f", "fo", "foo", "b", "ba", "barr", "bazz"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 5)}},
			{Code: `var foo = bar.baz;`, Options: deny("f", "fo", "fooo", "b", "ba", "bar", "bazz"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 11)}},
			{Code: `if (foo.bar) {}`, Options: deny("f", "fo", "foo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 5)}},
			{Code: `var obj = { key: foo.bar };`, Options: deny("obj"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("obj", 1, 5)}},
			{Code: `var obj = { key: foo.bar };`, Options: deny("key"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("key", 1, 13)}},
			{Code: `var obj = { key: foo.bar };`, Options: deny("foo"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 18)}},
			{Code: `var arr = [foo.bar];`, Options: deny("arr"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("arr", 1, 5)}},
			{Code: `var arr = [foo.bar];`, Options: deny("foo"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 12)}},
			{Code: `[foo.bar]`, Options: deny("f", "fo", "foo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 2)}},
			{Code: `if (foo.bar === bar.baz) { [bing.baz] }`, Options: deny("f", "fo", "foo", "b", "ba", "barr", "bazz", "bingg"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 5)}},
			{Code: `if (foo.bar === bar.baz) { [foo.bar] }`, Options: deny("f", "fo", "fooo", "b", "ba", "bar", "bazz", "bingg"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 17)}},
			{Code: `var myArray = new Array(); var myDate = new Date();`, Options: deny("array", "date", "myDate", "myarray", "new", "var"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("myDate", 1, 32)}},
			{Code: `var myArray = new Array(); var myDate = new Date();`, Options: deny("array", "date", "mydate", "myArray", "new", "var"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("myArray", 1, 5)}},
			{Code: `foo.bar = 1`, Options: deny("bar"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 5)}},
			{Code: `foo.bar.baz = 1`, Options: deny("bar", "baz"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("baz", 1, 9)}},
			{Code: `const {foo} = baz`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 8)}},
			{Code: `const {foo: bar} = baz`, Options: deny("foo", "bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 13)}},
			{Code: `const {[foo]: bar} = baz`, Options: deny("foo", "bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 9), restricted("bar", 1, 15)}},
			{Code: `const {foo: {bar: baz}} = qux`, Options: deny("foo", "bar", "baz"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("baz", 1, 19)}},
			{Code: `const {foo: {[bar]: baz}} = qux`, Options: deny("foo", "bar", "baz"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 15), restricted("baz", 1, 21)}},
			{Code: `const {[foo]: {[bar]: baz}} = qux`, Options: deny("foo", "bar", "baz"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 9), restricted("bar", 1, 17), restricted("baz", 1, 23)}},
			{Code: `function foo({ bar: baz }) {}`, Options: deny("bar", "baz"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("baz", 1, 21)}},
			{Code: `function foo({ bar: {baz: qux} }) {}`, Options: deny("bar", "baz", "qux"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("qux", 1, 27)}},
			{Code: `({foo: obj.bar} = baz);`, Options: deny("foo", "bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 12)}},
			{Code: `({foo: obj.bar.bar.bar.baz} = {});`, Options: deny("foo", "bar", "baz"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("baz", 1, 24)}},
			{Code: `({[foo]: obj.bar} = baz);`, Options: deny("foo", "bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 4), restricted("bar", 1, 14)}},
			{Code: `({foo: { a: obj.bar }} = baz);`, Options: deny("bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 17)}},
			{Code: `({a: obj.bar = baz} = qux);`, Options: deny("bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 10)}},
			{Code: `({a: obj.bar.bar.baz = obj.qux} = obj.qux);`, Options: deny("a", "bar", "baz", "qux"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("baz", 1, 18)}},
			{Code: `({a: obj[bar] = obj.qux} = obj.qux);`, Options: deny("a", "bar", "baz", "qux"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 10)}},
			{Code: `({a: [obj.bar] = baz} = qux);`, Options: deny("bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 11)}},
			{Code: `({foo: { a: obj.bar = baz}} = qux);`, Options: deny("bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 17)}},
			{Code: `({foo: { [a]: obj.bar }} = baz);`, Options: deny("bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 19)}},
			{Code: `({...obj.bar} = baz);`, Options: deny("bar"), LanguageOptions: es2018, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 10)}},
			{Code: `([obj.bar] = baz);`, Options: deny("bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 7)}},
			{Code: `const [bar] = baz;`, Options: deny("bar"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 8)}},

			// ---- upstream invalid: not a reference to a global variable, because it isn't a reference to a variable ----
			{Code: `foo.undefined = 1;`, Options: deny("undefined"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("undefined", 1, 5)}},
			{Code: `var foo = { undefined: 1 };`, Options: deny("undefined"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("undefined", 1, 13)}},
			{Code: `var foo = { undefined: undefined };`, Options: deny("undefined"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("undefined", 1, 13)}},
			{Code: `var foo = { Number() {} };`, Options: deny("Number"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 13)}},
			{Code: `class Foo { Number() {} }`, Options: deny("Number"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 13)}},
			{Code: `myGlobal: while(foo) { break myGlobal; } `, Options: deny("myGlobal"), LanguageOptions: es5, Globals: map[string]any{"myGlobal": "readonly"}, Errors: []rule_tester.InvalidTestCaseError{restricted("myGlobal", 1, 1), restricted("myGlobal", 1, 30)}},

			// ---- upstream invalid: globals declared in the given source code are not excluded from consideration ----
			{Code: `const foo = 1; bar = foo;`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 7), restricted("foo", 1, 22)}},
			{Code: `let foo; foo = bar;`, Options: deny("foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 5), restricted("foo", 1, 10)}},
			{Code: `bar = foo; var foo;`, Options: deny("foo"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 7), restricted("foo", 1, 16)}},
			{Code: `function foo() {} var bar = foo;`, Options: deny("foo"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 10), restricted("foo", 1, 29)}},
			{Code: `class Foo {} var bar = Foo;`, Options: deny("Foo"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("Foo", 1, 7), restricted("Foo", 1, 24)}},

			// ---- upstream invalid: redeclared globals are not excluded from consideration ----
			{Code: `let undefined; undefined = 1;`, Options: deny("undefined"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("undefined", 1, 5), restricted("undefined", 1, 16)}},
			{Code: `foo = undefined; var undefined;`, Options: deny("undefined"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("undefined", 1, 7), restricted("undefined", 1, 22)}},
			{Code: `function undefined(){} x = undefined;`, Options: deny("undefined"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("undefined", 1, 10), restricted("undefined", 1, 28)}},
			{Code: `class Number {} x = Number.NaN;`, Options: deny("Number"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 7), restricted("Number", 1, 21)}},

			// ---- upstream invalid: assignment to a property with a restricted name creates a global with that name ----
			{Code: `/* globals myGlobal */ window.myGlobal = 5; foo = myGlobal;`, Options: deny("myGlobal"), LanguageOptions: es5, Globals: map[string]any{"window": "readonly"}, Errors: []rule_tester.InvalidTestCaseError{restricted("myGlobal", 1, 31)}},

			// ---- upstream invalid: disabled global variables ----
			{Code: `var foo = undefined;`, Options: deny("undefined"), LanguageOptions: es5, Globals: map[string]any{"undefined": "off"}, Errors: []rule_tester.InvalidTestCaseError{restricted("undefined", 1, 11)}},
			{Code: `/* globals Number: off */ Number.parseInt()`, Options: deny("Number"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 27)}},
			{Code: `var foo = [Map];`, Options: deny("Map"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("Map", 1, 12)}},

			// ---- upstream invalid: shadowed global variables ----
			{Code: `if (foo) { let undefined; bar = undefined; }`, Options: deny("undefined"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("undefined", 1, 16), restricted("undefined", 1, 33)}},
			{Code: `function foo(Number) { var x = Number.NaN; }`, Options: deny("Number"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 14), restricted("Number", 1, 32)}},
			{Code: `function foo() { var myGlobal; x = myGlobal; }`, Options: deny("myGlobal"), LanguageOptions: es5, Globals: map[string]any{"myGlobal": "readonly"}, Errors: []rule_tester.InvalidTestCaseError{restricted("myGlobal", 1, 22), restricted("myGlobal", 1, 36)}},
			{Code: `function foo(bar) { return Number.parseInt(bar); } const Number = 1;`, Options: deny("Number"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 28), restricted("Number", 1, 58)}},
			{Code: `import Number from 'myNumber'; const foo = Number.parseInt(bar);`, Options: deny("Number"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 8), restricted("Number", 1, 44)}},
			{Code: `var foo = function undefined() {};`, Options: deny("undefined"), LanguageOptions: es5, Errors: []rule_tester.InvalidTestCaseError{restricted("undefined", 1, 20)}},

			// ---- upstream invalid: a reference to a global variable that also creates a property with a restricted name ----
			{Code: `var foo = { undefined }`, Options: deny("undefined"), LanguageOptions: es2015, Errors: []rule_tester.InvalidTestCaseError{restricted("undefined", 1, 13)}},

			// ---- upstream invalid: class fields ----
			{Code: `class C { camelCase; #camelCase; #camelCase2() {} }`, Options: deny("camelCase"), LanguageOptions: es2022, Errors: []rule_tester.InvalidTestCaseError{restricted("camelCase", 1, 11), restrictedPrivate("camelCase", 1, 22)}},
			{Code: `class C { snake_case; #snake_case() {}; #snake_case2() {} }`, Options: deny("snake_case"), LanguageOptions: es2022, Errors: []rule_tester.InvalidTestCaseError{restricted("snake_case", 1, 11), restrictedPrivate("snake_case", 1, 23)}},

			// ---- upstream invalid: not an import attribute key ----
			{Code: `import('foo.json', { with: { [type]: 'json' } })`, Options: deny("type"), LanguageOptions: es2025, Errors: []rule_tester.InvalidTestCaseError{restricted("type", 1, 31)}},
			{Code: `import('foo.json', { with: { type: json } })`, Options: deny("json"), LanguageOptions: es2025, Errors: []rule_tester.InvalidTestCaseError{restricted("json", 1, 36)}},
		},
	)
}
