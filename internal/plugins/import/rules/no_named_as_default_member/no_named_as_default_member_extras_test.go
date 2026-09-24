package no_named_as_default_member_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_named_as_default_member"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNamedAsDefaultMemberExtras(t *testing.T) {
	rule_tester.RunRuleTester(defaultMemberRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &no_named_as_default_member.NoNamedAsDefaultMemberRule,
		[]rule_tester.ValidTestCase{
			// Global namespace aliases are not named exports
			{Code: `import obj from './global-alias'; obj.foo;`},
			// Without interop, a global namespace declaration does not add exports
			{Code: `import obj from './global-members'; obj.Lib; obj.foo;`},
			// Non-identifier computed keys and object expressions are not evaluated
			{Code: `import obj from './base.js'; obj[foo!]; obj[foo as any]; obj[(foo, other)]; (obj, obj).foo;`},
			// Default-only modules cannot produce member reports
			{Code: `import obj from './default-only.js'; obj.foo; const {foo} = obj;`},
			// Loop patterns have no initializer
			{Code: `import obj from './base.js'; for (const {foo} of [obj]) {}`},
			// Upstream reports the dependency's parser error; rslint does not
			// turn dependency syntax errors into this rule's diagnostics.
			{Code: `import obj from './invalid-return.js'; obj.foo;`},
			{Code: `import obj from './base.js'; obj.foo;`, Settings: map[string]interface{}{"import/extensions": []interface{}{}}},
			// Only default specifiers are checked
			{Code: `import './base.js'; import { default as obj } from './base.js'; obj.foo;`},
			{Code: `import * as obj from './base.js'; obj.foo; const {foo} = obj;`},
			{Code: `const obj = require('./base.js'); obj.foo;`},
			{Code: `import obj = require('./base.js'); obj.foo;`},
			// Missing, ignored, empty and CommonJS modules
			{Code: `import obj from './missing.js'; obj.foo;`},
			{Code: `import obj from 'node:fs'; obj.foo;`},
			{Code: `import obj from './common.cjs'; obj.foo;`},
			{Code: `import obj from './empty.js'; obj.foo;`},
			{Code: `import obj from './base.js'; obj.foo;`, Settings: map[string]interface{}{"import/ignore": []interface{}{"base\\.js$"}}},
			{Code: `import obj from './base.js'; obj.foo;`, Settings: map[string]interface{}{"import/extensions": []interface{}{".ts"}}},
			// Re-exports do not belong to the local namespace
			{Code: `import obj from './reexport.js'; obj.foo;`},
			{Code: `import obj from './star.js'; obj.foo;`},
			{Code: `import obj from './unresolved.js'; obj.foo;`},
			// Literal and compound computed keys have no ESTree name
			{Code: "import obj from './base.js'; obj['foo']; obj[`foo`]; obj[0]; obj[foo + ''];"},
			{Code: "import obj from './base.js'; const {'foo': a, ['foo']: b, [`foo`]: c, 0: d} = obj;"},
			// Only the immediate identifier object and destructuring initializer count
			{Code: `import obj from './base.js'; const alias = obj; alias.foo; obj.unknown.foo; obj().foo;`},
			{Code: `import obj from './base.js'; const {foo} = obj.unknown; const [first] = obj; const {...rest} = obj;`},
			// Assignments and parameter patterns are not VariableDeclarator nodes
			{Code: `import obj from './base.js'; let foo; ({foo} = obj); function f({foo} = obj) {}`},
			// Authored TS expression wrappers remain visible
			{Code: `import obj from './base.js'; obj!.foo; (obj as any).foo; (obj satisfies any).foo; const {foo} = obj!;`},
			// JSX tag names and ordinary type names are not members
			{Code: `import obj from './base.js'; const view = <obj.foo />; type A = obj.foo; type B = typeof obj.foo;`, Tsx: true},
			// Default properties are excluded
			{Code: `import obj from './base.js'; obj.default; const {default: value} = obj;`},
		},
		[]rule_tester.InvalidTestCase{
			// rslint follows TypeScript's implicit interop default for NodeNext;
			// upstream requires esModuleInterop to be explicitly true.
			{Code: `import obj from './global-members'; obj.foo;`, FileName: "implicit/consumer.ts", TSConfig: "implicit/tsconfig.json", Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./global-members", 1, 37, 1, 44)}},
			// Property writes, deletions, updates and nested optional chains
			{Code: `import obj from './base.js'; obj.foo = 1; delete obj.other; ++obj.foo; obj?.foo?.value;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 30, 1, 37), defaultMemberError("obj", "other", "./base.js", 1, 50, 1, 59), defaultMemberError("obj", "foo", "./base.js", 1, 63, 1, 70), defaultMemberError("obj", "foo", "./base.js", 1, 72, 1, 80)}},
			// Object prototype names remain ordinary exported names
			{Code: `import obj from './special-names.js'; obj.__proto__; obj.constructor; obj.toString;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "__proto__", "./special-names.js", 1, 39, 1, 52), defaultMemberError("obj", "constructor", "./special-names.js", 1, 54, 1, 69), defaultMemberError("obj", "toString", "./special-names.js", 1, 71, 1, 83)}},
			// Unresolved namespace re-export still declares its alias
			{Code: `import obj from './namespace-unresolved'; obj.foo;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./namespace-unresolved", 1, 43, 1, 50)}},
			// Destructuring default initializers also contain ordinary member accesses
			{Code: `import obj from './base.js'; const {foo = obj.other} = obj;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 37, 1, 40), defaultMemberError("obj", "other", "./base.js", 1, 43, 1, 52)}},
			// A null extension setting uses defaults upstream
			{Code: `import obj from './null-extension'; obj.foo;`, Settings: map[string]interface{}{"import/extensions": nil}, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./null-extension", 1, 37, 1, 44)}},
			// Interop exposes the declared global namespace members
			{Code: `import obj from './global-members'; obj.foo; obj.Lib;`, FileName: "interop/consumer.ts", TSConfig: "interop/tsconfig.json", Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./global-members", 1, 37, 1, 44)}},
			// An explicitly configured parser's extensions augment the list.
			{Code: `import obj from './base.js'; obj.foo;`, Settings: map[string]interface{}{
				"import/extensions": []interface{}{},
				"import/parsers":    map[string]interface{}{"parser": []interface{}{".js"}},
			}, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 30, 1, 37)}},
			// Identifier computed keys are checked even though their runtime values may differ
			{Code: `import obj from './base.js'; obj[foo]; obj[(foo)];`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 30, 1, 38), defaultMemberError("obj", "foo", "./base.js", 1, 40, 1, 50)}},
			{Code: `import obj from './base.js'; const {[foo]: a, [(other)]: b} = obj;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 38, 1, 41), defaultMemberError("obj", "other", "./base.js", 1, 49, 1, 54)}},
			// Optional accesses and calls
			{Code: `import obj from './base.js'; obj?.foo; obj?.[foo]?.();`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 30, 1, 38), defaultMemberError("obj", "foo", "./base.js", 1, 40, 1, 50)}},
			// Parentheses around the object or initializer are transparent
			{Code: `import obj from './base.js'; ((obj)).foo; const {foo = 0, other: value = 1} = ((obj));`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 30, 1, 41), defaultMemberError("obj", "foo", "./base.js", 1, 50, 1, 53), defaultMemberError("obj", "other", "./base.js", 1, 59, 1, 64)}},
			// The immediate destructuring key is reported, not nested bindings
			{Code: `import obj from './base.js'; const {foo: {nested}, other: [item], ...rest} = obj;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 37, 1, 40), defaultMemberError("obj", "other", "./base.js", 1, 52, 1, 57)}},
			// Upstream intentionally matches shadowed identifiers by name
			{Code: `import obj from './base.js'; function f(obj) { return obj.foo; }`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 55, 1, 62)}},
			// Uses may precede the import
			{Code: `obj.foo; import obj from './base.js';`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 1, 1, 8)}},
			// A default export is not required
			{Code: `import obj from './named-only.js'; obj.foo;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./named-only.js", 1, 36, 1, 43)}},
			// Local export lists count even when the binding came from an import
			{Code: `import obj from './local.js'; obj.foo;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./local.js", 1, 31, 1, 38)}},
			// Namespace aliases count, without exposing their members
			{Code: `import obj from './namespace.js'; obj.foo; obj.other;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./namespace.js", 1, 35, 1, 42)}},
			// Quoted local export names are retained
			{Code: `import obj from './quoted.js'; obj.foo;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./quoted.js", 1, 32, 1, 39)}},
			// Cycles do not turn indirect names into local names
			{Code: `import obj from './cycle-a.js'; obj.foo; obj.other;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./cycle-a.js", 1, 33, 1, 40)}},
			// Named and namespace imports may accompany a default import
			{Code: `import obj, {foo as named} from './base.js'; import second, * as ns from './base.js'; obj.foo; second.other; ns.foo;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 87, 1, 94), defaultMemberError("second", "other", "./base.js", 1, 96, 1, 108)}},
			// PrivateIdentifier.name omits the hash
			{Code: `import obj from './base.js'; class C { #foo; read() { return obj.#foo; } }`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 62, 1, 70)}},
			// Member ranges include source parentheses and comments
			{Code: "import obj from './base.js';\n(obj) /* comment */\n .foo;\nconst {\n foo: local\n} = obj;", Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 2, 1, 3, 6), defaultMemberError("obj", "foo", "./base.js", 5, 2, 5, 5)}},
			// UTF-16 positions and decoded identifier names
			{Code: `import café from './base.js'; /* 😀 */ café.𐐀; café.café; café.\u0066oo;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("café", "𐐀", "./base.js", 1, 40, 1, 47), defaultMemberError("café", "café", "./base.js", 1, 49, 1, 58), defaultMemberError("café", "foo", "./base.js", 1, 60, 1, 73)}},
			// Explicit empty options
			{Code: `import obj from './base.js'; obj.foo;`, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 30, 1, 37)}},
			// JS and JSDoc casts
			{Code: `import obj from './base.js'; /** @type {any} */ (obj).foo;`, FileName: "consumer.js", Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 49, 1, 58)}},
			// JavaScript module extension
			{Code: `import obj from './base.js'; obj.foo;`, FileName: "consumer.mjs", Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 30, 1, 37)}},
			// TypeScript module extension
			{Code: `import obj from './base.js'; obj.foo;`, FileName: "consumer.mts", Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 30, 1, 37)}},
			// JSX expression members
			{Code: `import obj from './base.js'; const view = <div>{obj.foo}</div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 49, 1, 56)}},
			// Heritage names are ESTree members
			{Code: `import obj from './base.js'; interface A extends obj.foo {} class B implements obj.other {} class C extends obj.foo {}`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./base.js", 1, 50, 1, 57), defaultMemberError("obj", "other", "./base.js", 1, 80, 1, 89), defaultMemberError("obj", "foo", "./base.js", 1, 109, 1, 116)}},
			// Type exports and type-only default imports
			{Code: `import type obj from './types'; const a = obj.Shape; const {Alias} = obj; obj.Kind;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "Shape", "./types", 1, 43, 1, 52), defaultMemberError("obj", "Alias", "./types", 1, 61, 1, 66), defaultMemberError("obj", "Kind", "./types", 1, 75, 1, 83)}},
			// TypeScript namespace export assignment
			{Code: `import obj from './export-equals'; obj.foo;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("obj", "foo", "./export-equals", 1, 36, 1, 43)}},
		},
	)
}
