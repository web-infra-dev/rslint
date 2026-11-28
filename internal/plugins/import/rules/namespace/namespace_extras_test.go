package namespace_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/namespace"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

// TestNamespaceExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
func TestNamespaceExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&namespace.NamespaceRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver wrappers, multi-level parenthesized exported member remains valid ----
			{Code: `import * as names from "./named-exports"; console.log(((names)).a);`},
			// ---- Dimension 4: receiver wrappers, TS non-null expression is not an Identifier receiver ----
			{Code: `import * as names from "./named-exports"; console.log(names!.c);`},
			// ---- Dimension 4: receiver wrappers, TS as-expression is not an Identifier receiver ----
			{Code: `import * as names from "./named-exports"; console.log((names as any).c);`},
			// ---- Dimension 4: receiver wrappers, TS satisfies-expression is not an Identifier receiver ----
			{Code: `import * as names from "./named-exports"; console.log((names satisfies typeof names).c);`},
			// ---- Dimension 4: receiver wrappers, export-all namespace aliases are not recursively checked ----
			{Code: `import * as names from "./named-exports"; console.log(((names.deep)).e);`},
			// ---- Dimension 4: access/key forms, computed access is allowed when configured ----
			{Code: `import * as names from "./named-exports"; console.log(names[key]);`, Options: map[string]interface{}{"allowComputed": true}},
			// ---- Dimension 4: access/key forms, computed access after an export-all namespace alias is not checked ----
			{Code: `import * as names from "./named-exports"; console.log(names.deep["e"]);`},
			{Code: `import * as names from "./named-exports"; console.log(names.deep[key]);`, Options: map[string]interface{}{"allowComputed": true}},
			// ---- Dimension 4: access/key forms, private identifiers are N/A outside classes and cannot be namespace member syntax ----
			// ---- Dimension 4: declaration/container forms, default import points at an export-all namespace alias ----
			{Code: `import b from "./deep-namespace-chain/default"; console.log(b.c.d.e);`},
			// ---- Dimension 4: declaration/container forms, unresolved star exports keep the namespace open-ended ----
			{Code: `import * as unknown from "./unresolved-star-export"; console.log(unknown.anything.deep);`},
			// ---- Dimension 4: declaration/container forms, TS namespace declarations are not recursively checked ----
			{Code: `import * as ts from "./typescript-ambient-namespace"; ts.ambient.implicit(); ts.ambient.explicit();`},
			{Code: `import * as ts from "./typescript-ambient-namespace"; ts.ambient.missing();`},
			// ---- Dimension 4: declaration/container forms, string-named export-all default is not recursively checked ----
			{Code: `import names from "./default-export-namespace-string"; console.log(names.baz);`},
			// ---- Dimension 4: declaration/container forms, string-named export-all default is not recursively checked ----
			{Code: `import { "default" as names } from "./default-export-namespace-string"; console.log(names.baz);`},
			// ---- Real-user: default export that forwards a namespace import is validated, but ordinary parameter shadowing still wins ----
			{Code: `import binding from "./default-from-namespace-import"; function f(binding) { return binding.c; }`},
			// ---- Upstream parity: default import chains do not preserve namespace metadata for another default export ----
			{Code: `import binding from "./default-chain-from-namespace-import"; console.log(binding.c);`},
			// ---- Upstream parity: local named-import chains do not preserve namespace metadata for another local export ----
			{Code: `import { forwardedLocal } from "./namespace-import-local-reexport-chain"; console.log(forwardedLocal.c);`},
			// ---- Dimension 4: nesting/traversal boundaries, shadowed namespace in nested arrow is ignored ----
			{Code: `import * as names from "./named-exports"; const fn = (names) => names.c;`},
			// ---- Dimension 4: nesting/traversal boundaries, block-scoped shadowing is ignored ----
			{Code: `import * as names from "./named-exports"; { let names = {}; names.c; }`},
			// ---- Dimension 4: nesting/traversal boundaries, catch binding shadowing is ignored ----
			{Code: `import * as names from "./named-exports"; try { throw 1 } catch (names) { names.c; }`},
			// ---- Dimension 4: nesting/traversal boundaries, class declaration shadowing is ignored ----
			{Code: `import * as names from "./named-exports"; class names { method() { return names.c; } }`},
			// ---- Dimension 4: nesting/traversal boundaries, for-loop binding shadowing is ignored ----
			{Code: `import * as names from "./named-exports"; for (let names of []) { names.c; }`},
			// ---- Dimension 4: graceful degradation, rest binding is skipped without masking sibling checks ----
			{Code: `import * as names from "./named-exports"; const { a, ...rest } = names;`},
			// ---- Dimension 4: graceful degradation, array binding pattern is ignored ----
			{Code: `import * as names from "./named-exports"; const [a] = names;`},
			// ---- Dimension 4: graceful degradation, empty object binding pattern is ignored ----
			{Code: `import * as names from "./named-exports"; const {} = names;`},
			// ---- Dimension 4: assignment forms, update expressions are not AssignmentExpression in upstream ----
			{Code: `import * as foo from './bar'; foo.foo++;`},
			// ---- Dimension 4: assignment forms, deep member assignments are validated but not reported as namespace assignment ----
			{Code: `import * as foo from './bar'; foo.foo.bar = 'y';`},

			// Locks in upstream processBodyStatement branch 1: non-import declarations do not populate namespaces.
			{Code: `const names = {}; console.log(names.c);`},
			// Locks in upstream processBodyStatement branch 2: side-effect imports have no specifiers.
			{Code: `import "./named-exports"; console.log(names.c);`},
			// Locks in upstream MemberExpression branch 1: non-Identifier receiver is ignored.
			{Code: `import * as names from "./named-exports"; console.log(getNames().c);`},
			// Locks in upstream VariableDeclarator branch 1: missing initializer is ignored.
			{Code: `import * as names from "./named-exports"; let value;`},
			// Locks in upstream VariableDeclarator branch 2: non-Identifier initializer is ignored.
			{Code: `import * as names from "./named-exports"; const { c } = getNames();`},
			// Locks in upstream testKey branch: non-object binding pattern is ignored.
			{Code: `import * as names from "./named-exports"; const value = names;`},

			// ---- Real-user: multiple export namespace re-exports resolve without false positives ----
			{Code: `import * as reexports from './namespace-reexports'; console.log(reexports.abc.B); console.log(reexports.def.F);`, FileName: "multiple-namespace-reexports/consumer.ts"},
			// ---- Upstream parity: export-all namespace alias missing members are not recursively checked ----
			{Code: `import * as reexports from './namespace-reexports'; console.log(reexports.abc.Missing);`, FileName: "multiple-namespace-reexports/consumer.ts"},
			// ---- Real-user: #456 allowComputed permits dynamic namespace access ----
			{Code: `import * as Module from './named-exports'; function getItem(name) { return Module[name]; }`, Options: map[string]interface{}{"allowComputed": true}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `import * as names from "./named-exports" with { type: "text" }; console.log(names.a);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notFound", Message: `'a' not found in imported namespace 'names'.`,
				}},
			},
			// ---- Dimension 4: receiver wrappers, parenthesized receiver still reports missing member ----
			{
				Code: `import * as names from "./named-exports"; console.log(((names)).c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 65, EndLine: 1, EndColumn: 66},
				},
			},
			// ---- Dimension 4: access/key forms, string element access is reported as computed rather than validated ----
			{
				Code: `import * as names from "./named-exports"; console.log(names["a"]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "computed", Message: computedNames, Line: 1, Column: 61, EndLine: 1, EndColumn: 64},
				},
			},
			// ---- Dimension 4: access/key forms, numeric element access is reported as computed ----
			{
				Code: `import * as names from "./named-exports"; console.log(names[0]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "computed", Message: computedNames, Line: 1, Column: 61, EndLine: 1, EndColumn: 62},
				},
			},
			// ---- Dimension 4: options, empty options array keeps computed reporting enabled ----
			{
				Code:    `import * as names from "./named-exports"; console.log(names["a"]);`,
				Options: []interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "computed", Message: computedNames, Line: 1, Column: 61, EndLine: 1, EndColumn: 64},
				},
			},
			// ---- Dimension 4: access/key forms, non-Identifier destructuring key is rejected ----
			{
				Code: `import * as names from "./named-exports"; const { "a": a } = names`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "onlyTopLevel", Message: "Only destructure top-level names.", Line: 1, Column: 51, EndLine: 1, EndColumn: 57},
				},
			},
			// ---- Dimension 4: access/key forms, computed destructuring key is rejected ----
			{
				Code: `import * as names from "./named-exports"; const { [a]: value } = names`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "onlyTopLevel", Message: "Only destructure top-level names.", Line: 1, Column: 51, EndLine: 1, EndColumn: 61},
				},
			},
			// ---- Dimension 4: declaration/container forms, namespace import from an empty ES module reports at import ----
			{
				Code: `import * as empty from "./empty-module";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noExports", Message: `No exported names found in module './empty-module'.`, Line: 1, Column: 13, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- Dimension 4: declaration/container forms, export namespace from an empty ES module reports at export ----
			{
				Code: `export * as empty from "./empty-module";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noExports", Message: `No exported names found in module './empty-module'.`, Line: 1, Column: 8, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- Dimension 4: nesting/traversal boundaries, outer namespace remains checked after inner shadowing scope ----
			{
				Code: `import * as names from "./named-exports"; function f(names) { return names.c } console.log(names.c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 98, EndLine: 1, EndColumn: 99},
				},
			},
			// ---- Dimension 4: graceful degradation, rest binding is skipped but missing sibling is still reported ----
			{
				Code: `import * as names from "./named-exports"; const { a, c, ...rest } = names;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 54, EndLine: 1, EndColumn: 55},
				},
			},
			// Locks in upstream MemberExpression branch 5: missing direct member reports and stops.
			{
				Code: `import * as names from "./named-exports"; console.log(names.c.d);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 61, EndLine: 1, EndColumn: 62},
				},
			},
			// Locks in upstream AssignmentExpression branch for compound assignment operators.
			{
				Code: `import * as foo from './bar'; foo.foo += 'y';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignment", Message: `Assignment to member of namespace 'foo'.`, Line: 1, Column: 31, EndLine: 1, EndColumn: 45},
				},
			},
			// Locks in upstream ExportNamespaceSpecifier branch: empty export namespace reports without adding a local namespace.
			{
				Code: `export * as empty from "./empty-module"; console.log(empty.missing);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noExports", Message: `No exported names found in module './empty-module'.`, Line: 1, Column: 8, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- Real-user: default export that forwards a namespace import exposes the namespace to default import consumers ----
			{
				Code: `import binding from "./default-from-namespace-import"; console.log(binding.c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: `'c' not found in imported namespace 'binding'.`, Line: 1, Column: 76, EndLine: 1, EndColumn: 77},
				},
			},
			// ---- Real-user: declaration files that default-export a namespace import expose that namespace ----
			{
				Code: `import binding from "./default-from-namespace-import-declaration"; console.log(binding.c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: `'c' not found in imported namespace 'binding'.`, Line: 1, Column: 88, EndLine: 1, EndColumn: 89},
				},
			},
			// ---- Upstream parity: local export of a namespace import preserves namespace metadata ----
			{
				Code: `import { forwarded } from "./namespace-import-local-reexport"; console.log(forwarded.c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: `'c' not found in imported namespace 'forwarded'.`, Line: 1, Column: 86, EndLine: 1, EndColumn: 87},
				},
			},
			// ---- Upstream parity: namespace import of a local namespace re-export checks deeply ----
			{
				Code: `import * as wrapper from "./namespace-import-local-reexport"; console.log(wrapper.forwarded.c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: `'c' not found in deeply imported namespace 'wrapper.forwarded'.`, Line: 1, Column: 93, EndLine: 1, EndColumn: 94},
				},
			},
			// ---- Upstream parity: source re-export preserves remote namespace metadata ----
			{
				Code: `import { forwardedAgain } from "./namespace-import-source-reexport"; console.log(forwardedAgain.c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: `'c' not found in imported namespace 'forwardedAgain'.`, Line: 1, Column: 97, EndLine: 1, EndColumn: 98},
				},
			},
			// ---- Real-user: mirrors eslint-plugin-import when a same-name parameter's type annotation references the namespace ----
			{
				Code: `import binding from "./default-from-namespace-import"; class EntryData { private constructor(binding: binding.a) { this.value = binding.c; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notFound", Message: `'c' not found in imported namespace 'binding'.`, Line: 1, Column: 137, EndLine: 1, EndColumn: 138},
				},
			},
			// ---- Real-user: #456 default computed namespace access reports when allowComputed is false ----
			{
				Code: `import * as Module from './named-exports'; function getItem(name) { return Module[name]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "computed", Message: `Unable to validate computed reference to imported namespace 'Module'.`, Line: 1, Column: 83, EndLine: 1, EndColumn: 87},
				},
			},
		},
	)
}

func TestNamespaceHeritage(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &namespace.NamespaceRule, []rule_tester.ValidTestCase{
		{Code: `import * as names from './named-exports'; interface I extends names.a {}`},
		{Code: `import * as names from './named-exports'; class C implements names.a {}`},
		{Code: `import * as names from './named-exports'; type T = names.c;`},
		{Code: `import * as names from './named-exports'; type T = typeof names.c;`},
		{Code: `import * as names from './named-exports'; interface I extends Base<names.c> {}`},
		{Code: `import * as names from './named-exports'; function f(names) { class C implements names.c {} return C; }`},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `import * as names from './named-exports'; interface I extends names.c {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 69, EndLine: 1, EndColumn: 70}},
		},
		{
			Code:   `import * as names from './named-exports'; class C implements names.c {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 68, EndLine: 1, EndColumn: 69}},
		},
		{
			Code:   `import * as names from './named-exports'; class C extends names.c {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 65, EndLine: 1, EndColumn: 66}},
		},
		{
			Code:   `import * as names from './named-exports'; interface I extends names.c.d {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notFound", Message: notFoundNamesC, Line: 1, Column: 69, EndLine: 1, EndColumn: 70}},
		},
	})
}

func TestNamespaceCyclicBarrelReexports(t *testing.T) {
	for _, tc := range []struct {
		name  string
		index string
	}{
		{
			name: "b imported before a",
			index: `
import * as b from './b.js';
import * as a from './a.js';
export { a, b };
`,
		},
		{
			name: "a imported before b",
			index: `
import * as a from './a.js';
import * as b from './b.js';
export { a, b };
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := fixtures.GetRootDir()
			root.FS = rslint_utils.NewOverlayVFS(root.FS, map[string]string{
				tspath.ResolvePath(root.Dir, "cyclic-barrel/a.ts"): `
export * from './b.js';
export const value = 42;
`,
				tspath.ResolvePath(root.Dir, "cyclic-barrel/b.ts"): `
export * from './a.js';
`,
				tspath.ResolvePath(root.Dir, "cyclic-barrel/index.ts"): tc.index,
			})
			rule_tester.RunRuleTester(
				root,
				"tsconfig.json",
				t,
				&namespace.NamespaceRule,
				[]rule_tester.ValidTestCase{
					{
						Code:     `import { b } from './index.js'; console.log(b.value);`,
						FileName: "cyclic-barrel/consumer.ts",
					},
					{
						Code:     `import { a, b } from './index.js'; console.log(a.value, b.value);`,
						FileName: "cyclic-barrel/consumer.ts",
					},
				},
				[]rule_tester.InvalidTestCase{
					{
						Code:     `import { b } from './index.js'; console.log(b.missing);`,
						FileName: "cyclic-barrel/consumer.ts",
						Errors: []rule_tester.InvalidTestCaseError{
							{MessageId: "notFound", Message: `'missing' not found in imported namespace 'b'.`, Line: 1, Column: 47, EndLine: 1, EndColumn: 54},
						},
					},
					{
						Code:     `import { a } from './index.js'; console.log(a.missing);`,
						FileName: "cyclic-barrel/consumer.ts",
						Errors: []rule_tester.InvalidTestCaseError{
							{MessageId: "notFound", Message: `'missing' not found in imported namespace 'a'.`, Line: 1, Column: 47, EndLine: 1, EndColumn: 54},
						},
					},
				},
			)
		})
	}

	t.Run("cyclic namespace reference declared before value export", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			index string
		}{
			{
				name: "value re-export before namespace re-export",
				index: `
export { value } from './a.js';
export { c } from './b.js';
`,
			},
			{
				name: "namespace re-export before value re-export",
				index: `
export { c } from './b.js';
export { value } from './a.js';
`,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				root := fixtures.GetRootDir()
				root.FS = rslint_utils.NewOverlayVFS(root.FS, map[string]string{
					tspath.ResolvePath(root.Dir, "cyclic-ns-ref/a.ts"): `
export { value } from './c.js';
`,
					tspath.ResolvePath(root.Dir, "cyclic-ns-ref/b.ts"): `
import * as c from './c.js';
export { c };
export const value = 42;
export * from './a.js';
`,
					tspath.ResolvePath(root.Dir, "cyclic-ns-ref/c.ts"): `
export * from './b.js';
`,
					tspath.ResolvePath(root.Dir, "cyclic-ns-ref/index.ts"): tc.index,
				})
				rule_tester.RunRuleTester(
					root,
					"tsconfig.json",
					t,
					&namespace.NamespaceRule,
					[]rule_tester.ValidTestCase{
						{
							Code:     `import { c } from './index.js'; console.log(c.value);`,
							FileName: "cyclic-ns-ref/consumer.ts",
						},
						{
							Code:     `import { c } from './index.js'; console.log(c.c.value);`,
							FileName: "cyclic-ns-ref/consumer.ts",
						},
					},
					[]rule_tester.InvalidTestCase{
						{
							Code:     `import { c } from './index.js'; console.log(c.missing);`,
							FileName: "cyclic-ns-ref/consumer.ts",
							Errors: []rule_tester.InvalidTestCaseError{
								{MessageId: "notFound", Message: `'missing' not found in imported namespace 'c'.`, Line: 1, Column: 47, EndLine: 1, EndColumn: 54},
							},
						},
					},
				)
			})
		}
	})

	t.Run("cyclic star re-export does not override explicit binding", func(t *testing.T) {
		for _, order := range []struct {
			name string
			b    string
		}{
			{
				name: "explicit before star",
				b: `
export { alias as api } from './a.js';
export * from './a.js';
`,
			},
			{
				name: "star before explicit",
				b: `
export * from './a.js';
export { alias as api } from './a.js';
`,
			},
		} {
			t.Run("plain value source/"+order.name, func(t *testing.T) {
				root := fixtures.GetRootDir()
				root.FS = rslint_utils.NewOverlayVFS(root.FS, map[string]string{
					tspath.ResolvePath(root.Dir, "cyclic-explicit-vs-star/a.ts"): `
import * as b from './b.js';
export { b, b as api };
export { source as alias } from './b.js';
export const source = { good: 42 };
`,
					tspath.ResolvePath(root.Dir, "cyclic-explicit-vs-star/b.ts"): order.b,
				})
				rule_tester.RunRuleTester(
					root,
					"tsconfig.json",
					t,
					&namespace.NamespaceRule,
					[]rule_tester.ValidTestCase{
						{
							Code:     `import { b } from './a.js'; console.log(b.api.good);`,
							FileName: "cyclic-explicit-vs-star/consumer.ts",
						},
						{
							Code:     `import * as b from './b.js'; console.log(b.api.good);`,
							FileName: "cyclic-explicit-vs-star/consumer.ts",
						},
					},
					[]rule_tester.InvalidTestCase{
						{
							Code:     `import { b } from './a.js'; console.log(b.missing);`,
							FileName: "cyclic-explicit-vs-star/consumer.ts",
							Errors: []rule_tester.InvalidTestCaseError{
								{MessageId: "notFound", Message: `'missing' not found in imported namespace 'b'.`, Line: 1, Column: 43, EndLine: 1, EndColumn: 50},
							},
						},
					},
				)
			})

			t.Run("namespace source/"+order.name, func(t *testing.T) {
				root := fixtures.GetRootDir()
				root.FS = rslint_utils.NewOverlayVFS(root.FS, map[string]string{
					tspath.ResolvePath(root.Dir, "cyclic-explicit-vs-star-ns/source.ts"): `
export const good = 42;
`,
					tspath.ResolvePath(root.Dir, "cyclic-explicit-vs-star-ns/a.ts"): `
import * as b from './b.js';
import * as source from './source.js';
export { b, b as api, source };
export { source as alias } from './b.js';
`,
					tspath.ResolvePath(root.Dir, "cyclic-explicit-vs-star-ns/b.ts"): order.b,
				})
				rule_tester.RunRuleTester(
					root,
					"tsconfig.json",
					t,
					&namespace.NamespaceRule,
					[]rule_tester.ValidTestCase{
						{
							Code:     `import { b } from './a.js'; console.log(b.api.good);`,
							FileName: "cyclic-explicit-vs-star-ns/consumer.ts",
						},
						{
							Code:     `import * as b from './b.js'; console.log(b.api.good);`,
							FileName: "cyclic-explicit-vs-star-ns/consumer.ts",
						},
					},
					[]rule_tester.InvalidTestCase{
						{
							Code:     `import { b } from './a.js'; console.log(b.api.missing);`,
							FileName: "cyclic-explicit-vs-star-ns/consumer.ts",
							Errors: []rule_tester.InvalidTestCaseError{
								{MessageId: "notFound", Message: `'missing' not found in deeply imported namespace 'b.api'.`, Line: 1, Column: 47, EndLine: 1, EndColumn: 54},
							},
						},
						{
							Code:     `import * as b from './b.js'; console.log(b.api.missing);`,
							FileName: "cyclic-explicit-vs-star-ns/consumer.ts",
							Errors: []rule_tester.InvalidTestCaseError{
								{MessageId: "notFound", Message: `'missing' not found in deeply imported namespace 'b.api'.`, Line: 1, Column: 48, EndLine: 1, EndColumn: 55},
							},
						},
					},
				)
			})
		}

		// Equal-rank star exports in a JS cycle: both direct.js (which holds
		// the namespace binding) and indirect.js (a plain re-export without
		// namespace metadata) forward `api` into hub.js. Regardless of `export *`
		// order, hub.api must preserve leaf's Namespace so `direct.hub.api.missing`
		// is reported.
		for _, tc := range []struct {
			name string
			hub  string
		}{
			{
				name: "direct_then_indirect",
				hub:  "export * from './direct.js';\nexport * from './indirect.js';\n",
			},
			{
				name: "indirect_then_direct",
				hub:  "export * from './indirect.js';\nexport * from './direct.js';\n",
			},
		} {
			t.Run("equal_rank_star_preserves_namespace/"+tc.name, func(t *testing.T) {
				root := fixtures.GetRootDir()
				root.FS = rslint_utils.NewOverlayVFS(root.FS, map[string]string{
					tspath.ResolvePath(root.Dir, "cyclic-star-ns-merge/tsconfig.json"): `{"compilerOptions":{"allowJs":true,"module":"esnext","moduleResolution":"bundler","target":"esnext"}}`,
					tspath.ResolvePath(root.Dir, "cyclic-star-ns-merge/leaf.js"): `
export const good = 42;
`,
					tspath.ResolvePath(root.Dir, "cyclic-star-ns-merge/direct.js"): `
import * as hub from './hub.js';
export { hub };
import * as api from './leaf.js';
export { api };
`,
					tspath.ResolvePath(root.Dir, "cyclic-star-ns-merge/indirect.js"): `
import * as direct from './direct.js';
export { direct };
import { api } from './direct.js';
export { api };
`,
					tspath.ResolvePath(root.Dir, "cyclic-star-ns-merge/hub.js"): tc.hub,
				})
				rule_tester.RunRuleTester(
					root,
					"cyclic-star-ns-merge/tsconfig.json",
					t,
					&namespace.NamespaceRule,
					[]rule_tester.ValidTestCase{
						{
							Code:     `import { direct } from './indirect.js'; console.log(direct.hub.api.good);`,
							FileName: "cyclic-star-ns-merge/consumer.js",
						},
						{
							Code:     `import * as direct from './direct.js'; console.log(direct.hub.api.good);`,
							FileName: "cyclic-star-ns-merge/consumer.js",
						},
						{
							Code:     `import * as hub from './hub.js'; console.log(hub.api.good);`,
							FileName: "cyclic-star-ns-merge/consumer.js",
						},
					},
					[]rule_tester.InvalidTestCase{
						{
							Code:     `import { direct } from './indirect.js'; console.log(direct.hub.api.missing);`,
							FileName: "cyclic-star-ns-merge/consumer.js",
							Errors: []rule_tester.InvalidTestCaseError{
								{MessageId: "notFound", Message: `'missing' not found in deeply imported namespace 'direct.hub.api'.`, Line: 1, Column: 68, EndLine: 1, EndColumn: 75},
							},
						},
						{
							Code:     `import * as direct from './direct.js'; console.log(direct.hub.api.missing);`,
							FileName: "cyclic-star-ns-merge/consumer.js",
							Errors: []rule_tester.InvalidTestCaseError{
								{MessageId: "notFound", Message: `'missing' not found in deeply imported namespace 'direct.hub.api'.`, Line: 1, Column: 67, EndLine: 1, EndColumn: 74},
							},
						},
						{
							Code:     `import * as hub from './hub.js'; console.log(hub.api.missing);`,
							FileName: "cyclic-star-ns-merge/consumer.js",
							Errors: []rule_tester.InvalidTestCaseError{
								{MessageId: "notFound", Message: `'missing' not found in deeply imported namespace 'hub.api'.`, Line: 1, Column: 54, EndLine: 1, EndColumn: 61},
							},
						},
					},
				)
			})
		}

		// Conflicting `export *` bindings of the same name (`api`) in a barrel:
		// namespace.js exports a namespace binding while values.js exports a
		// distinct plain value binding. JavaScript omits the ambiguous name, so
		// neither star order may report a false missing-member error.
		for _, tc := range []struct {
			name   string
			barrel string
		}{
			{
				name:   "namespace_then_values",
				barrel: "export * from './namespace.js';\nexport * from './values.js';\n",
			},
			{
				name:   "values_then_namespace",
				barrel: "export * from './values.js';\nexport * from './namespace.js';\n",
			},
		} {
			t.Run("conflicting_star_exports_clear_namespace/"+tc.name, func(t *testing.T) {
				root := fixtures.GetRootDir()
				root.FS = rslint_utils.NewOverlayVFS(root.FS, map[string]string{
					tspath.ResolvePath(root.Dir, "conflicting-star-ns/tsconfig.json"): `{"compilerOptions":{"allowJs":true,"module":"esnext","moduleResolution":"bundler","target":"esnext"}}`,
					tspath.ResolvePath(root.Dir, "conflicting-star-ns/values.js"): `
export const api = 7;
`,
					tspath.ResolvePath(root.Dir, "conflicting-star-ns/namespace.js"): `
import * as api from './values.js';
export { api };
`,
					tspath.ResolvePath(root.Dir, "conflicting-star-ns/barrel.js"): tc.barrel,
				})
				rule_tester.RunRuleTester(
					root,
					"conflicting-star-ns/tsconfig.json",
					t,
					&namespace.NamespaceRule,
					[]rule_tester.ValidTestCase{
						{
							Code:     `import * as barrel from './barrel.js'; const { api: { good } = { good: 42 } } = barrel; console.log(good);`,
							FileName: "conflicting-star-ns/consumer.js",
						},
						{
							Code:     `import * as barrel from './barrel.js'; console.log(barrel.api.good);`,
							FileName: "conflicting-star-ns/consumer.js",
						},
					},
					[]rule_tester.InvalidTestCase{
						{
							Code:     `import * as barrel from './barrel.js'; console.log(barrel.missing);`,
							FileName: "conflicting-star-ns/consumer.js",
							Errors: []rule_tester.InvalidTestCaseError{
								{MessageId: "notFound", Message: `'missing' not found in imported namespace 'barrel'.`, Line: 1, Column: 59, EndLine: 1, EndColumn: 66},
							},
						},
					},
				)
			})
		}

		// Multi-hop same-binding star exports (direct -> mid -> indirect) still
		// resolve to the same binding identity and preserve leaf's Namespace.
		for _, tc := range []struct {
			name string
			hub  string
		}{
			{
				name: "direct_then_indirect",
				hub:  "export * from './direct.js';\nexport * from './indirect.js';\n",
			},
			{
				name: "indirect_then_direct",
				hub:  "export * from './indirect.js';\nexport * from './direct.js';\n",
			},
		} {
			t.Run("multi_hop_same_binding_star_preserves_namespace/"+tc.name, func(t *testing.T) {
				root := fixtures.GetRootDir()
				root.FS = rslint_utils.NewOverlayVFS(root.FS, map[string]string{
					tspath.ResolvePath(root.Dir, "multi-hop-star-ns/tsconfig.json"): `{"compilerOptions":{"allowJs":true,"module":"esnext","moduleResolution":"bundler","target":"esnext"}}`,
					tspath.ResolvePath(root.Dir, "multi-hop-star-ns/leaf.js"):       `export const good = 42;`,
					tspath.ResolvePath(root.Dir, "multi-hop-star-ns/direct.js"): `
import * as api from './leaf.js';
export { api };
`,
					tspath.ResolvePath(root.Dir, "multi-hop-star-ns/mid.js"): `
export { api } from './direct.js';
`,
					tspath.ResolvePath(root.Dir, "multi-hop-star-ns/indirect.js"): `
import { api } from './mid.js';
export { api };
`,
					tspath.ResolvePath(root.Dir, "multi-hop-star-ns/hub.js"): tc.hub,
				})
				rule_tester.RunRuleTester(
					root,
					"multi-hop-star-ns/tsconfig.json",
					t,
					&namespace.NamespaceRule,
					[]rule_tester.ValidTestCase{
						{
							Code:     `import * as hub from './hub.js'; console.log(hub.api.good);`,
							FileName: "multi-hop-star-ns/consumer.js",
						},
					},
					[]rule_tester.InvalidTestCase{
						{
							Code:     `import * as hub from './hub.js'; console.log(hub.api.missing);`,
							FileName: "multi-hop-star-ns/consumer.js",
							Errors: []rule_tester.InvalidTestCaseError{
								{MessageId: "notFound", Message: `'missing' not found in deeply imported namespace 'hub.api'.`, Line: 1, Column: 54, EndLine: 1, EndColumn: 61},
							},
						},
					},
				)
			})
		}

		// Local `import * as local` re-exported as `api` alongside `export * as local from './other.js'`
		// are distinct bindings; star-merging them must conflict and clear Namespace in both orders.
		for _, tc := range []struct {
			name   string
			barrel string
		}{
			{
				name:   "source_then_alias",
				barrel: "export * from './source.js';\nexport * from './alias.js';\n",
			},
			{
				name:   "alias_then_source",
				barrel: "export * from './alias.js';\nexport * from './source.js';\n",
			},
		} {
			t.Run("namespace_reexport_conflicts_with_same_named_local_import/"+tc.name, func(t *testing.T) {
				root := fixtures.GetRootDir()
				root.FS = rslint_utils.NewOverlayVFS(root.FS, map[string]string{
					tspath.ResolvePath(root.Dir, "ns-reexport-conflict/tsconfig.json"): `{"compilerOptions":{"allowJs":true,"module":"esnext","moduleResolution":"bundler","target":"esnext"}}`,
					tspath.ResolvePath(root.Dir, "ns-reexport-conflict/leaf.js"):       `export const left = 1;`,
					tspath.ResolvePath(root.Dir, "ns-reexport-conflict/other.js"):      `export const right = 2;`,
					tspath.ResolvePath(root.Dir, "ns-reexport-conflict/source.js"): `
import * as local from './leaf.js';
export { local as api };
export * as local from './other.js';
`,
					tspath.ResolvePath(root.Dir, "ns-reexport-conflict/alias.js"):  `export { local as api } from './source.js';`,
					tspath.ResolvePath(root.Dir, "ns-reexport-conflict/barrel.js"): tc.barrel,
				})
				rule_tester.RunRuleTester(
					root,
					"ns-reexport-conflict/tsconfig.json",
					t,
					&namespace.NamespaceRule,
					[]rule_tester.ValidTestCase{
						{
							Code: `
import * as barrel from './barrel.js';
const { api: { good } = { good: 42 } } = barrel;
console.log(good);
`,
							FileName: "ns-reexport-conflict/consumer.js",
						},
					},
					[]rule_tester.InvalidTestCase{},
				)
			})
		}
	})
}
