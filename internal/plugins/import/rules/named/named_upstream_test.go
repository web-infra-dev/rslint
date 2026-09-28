package named_test

// cspell:ignore zoob flowtypes jsnext Extfield Typess Snorlax Doesnt

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/named"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func namedRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "named-rule")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("empty fixture archive")
	}
	files := make(map[string]string, len(names))
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(root.Dir, name)] = string(data)
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	return root
}

// All semantic cases from eslint-plugin-import v2.32.0 tests/src/rules/named.js,
// expanded SYNTAX_CASES, and docs/rules/named.md. Parser/version repeats collapse
// into one native case. Package entry fixtures use relative paths; the Flow and
// export-default-from dependency fixtures use equivalent TypeScript/ES syntax.
// Upstream reports literal messages, so MessageId is intentionally empty.

// Upstream group: named
func TestNamedUpstream(t *testing.T) {
	root := namedRoot(t, "testdata/upstream.txtar")
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &named.NamedRule,
		[]rule_tester.ValidTestCase{
			// The test Program rejects malformed dependencies before rule execution.
			{Code: `import "./malformed.js"`, Skip: true},
			{Code: `import { foo } from "./bar"`},
			{Code: `import { foo } from "./empty-module"`},
			{Code: `import bar from "./bar.js"`},
			{Code: `import bar, { foo } from "./bar.js"`},
			{Code: `import {a, b, d} from "./named-exports"`},
			{Code: `import {ExportedClass} from "./named-exports"`},
			{Code: `import { destructingAssign } from "./named-exports"`},
			{Code: `import { destructingRenamedAssign } from "./named-exports"`},
			{Code: `import { ActionTypes } from "./qc"`},
			{Code: `import {a, b, c, d} from "./re-export"`},
			{Code: `import {a, b, c} from "./re-export-common-star"`},
			{Code: `import {RuleTester} from "./re-export-node_modules"`},
			{Code: `import { jsxFoo } from "./jsx/AnotherComponent"`, Settings: map[string]any{"import/resolve": map[string]any{"extensions": []any{".js", ".jsx"}}}},
			{Code: `import {a, b, d} from "./common"; // eslint-disable-line import/named`},
			{Code: `import { foo, bar } from "./re-export-names"`},
			{Code: `import { foo, bar } from "./common"`, Settings: map[string]any{"import/ignore": []any{"common"}}},
			{Code: `import { foo } from "crypto"`},
			{Code: `import { zoob } from "a"`},
			{Code: `import { someThing } from "./test-module"`},
			{Code: `export { foo } from "./bar"`},
			{Code: `export { foo as bar } from "./bar"`},
			{Code: `export { foo } from "./does-not-exist"`},
			// Babel export-default-from declarations are not parsed by tsgo.
			{Code: `export bar, { foo } from "./bar"`, Skip: true},
			{Code: `import { foo, bar } from "./named-trampoline"`},
			{Code: `let foo; export { foo as bar }`},
			{Code: `import { destructuredProp } from "./named-exports"`},
			{Code: `import { arrayKeyProp } from "./named-exports"`},
			{Code: `import { deepProp } from "./named-exports"`},
			{Code: `import { deepSparseElement } from "./named-exports"`},
			{Code: `import type { MissingType } from "./flowtypes"`},
			// Flow typeof imports are not parsed by tsgo.
			{Code: `import typeof { MissingType } from "./flowtypes"`, Skip: true},
			{Code: `import type { MyOpaqueType } from "./flowtypes"`},
			// Flow typeof imports are not parsed by tsgo.
			{Code: `import typeof { MyOpaqueType } from "./flowtypes"`, Skip: true},
			{Code: `import { type MyOpaqueType, MyClass } from "./flowtypes"`},
			// Flow typeof imports are not parsed by tsgo.
			{Code: `import { typeof MyOpaqueType, MyClass } from "./flowtypes"`, Skip: true},
			// Flow typeof imports are not parsed by tsgo.
			{Code: `import typeof MissingType from "./flowtypes"`, Skip: true},
			// Flow typeof imports are not parsed by tsgo.
			{Code: `import typeof * as MissingType from "./flowtypes"`, Skip: true},
			{Code: `export type { MissingType } from "./flowtypes"`},
			{Code: `export type { MyOpaqueType } from "./flowtypes"`},
			{Code: `/*jsnext*/ import { createStore } from "./redux"`, Settings: map[string]any{"import/ignore": []any{}}},
			{Code: `/*jsnext*/ import { createStore } from "./redux"`},
			{Code: `import { foo } from "./es6-module"`},
			{Code: `import { me, soGreat } from "./narcissist"`},
			{Code: `import { foo, bar, baz } from "./re-export-default"`},
			{Code: `import { common } from "./re-export-default"`, Settings: map[string]any{"import/ignore": []any{"common"}}},
			{Code: `import {a, b, d} from "./common"`},
			{Code: `import { baz } from "./bar"`, Settings: map[string]any{"import/ignore": []any{"bar"}}},
			{Code: `import { common } from "./re-export-default"`},
			{Code: `const { destructuredProp } = require("./named-exports")`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `let { arrayKeyProp } = require("./named-exports")`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `const { deepProp } = require("./named-exports")`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `const { foo, bar } = require("./re-export-names")`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `const { baz } = require("./bar")`},
			{Code: `const { baz } = require("./bar")`, Options: []any{map[string]any{"commonjs": false}}},
			{Code: `const { default: defExport } = require("./bar")`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `for (let { foo, bar } of baz) {}`},
			{Code: `for (let [ foo, bar ] of baz) {}`},
			{Code: `const { x, y } = bar`},
			{Code: `const { x, y, ...z } = bar`},
			{Code: `let x; export { x }`},
			{Code: `let x; export { x as y }`},
			{Code: `export const x = null`},
			{Code: `export var x = null`},
			{Code: `export let x = null`},
			{Code: `export default x`},
			{Code: `export default class x {}`},
			{Code: `import json from "./data.json"`, Settings: map[string]any{"import/extensions": []any{".js"}}},
			{Code: `import foo from "./foobar.json";`, Settings: map[string]any{"import/extensions": []any{".js"}}},
			{Code: `import foo from "./foobar";`, Settings: map[string]any{"import/extensions": []any{".js"}}},
			{Code: `import { foo } from "./issue-370-commonjs-namespace/bar"`, Settings: map[string]any{"import/ignore": []any{"foo"}}},
			{Code: `export * from "./issue-370-commonjs-namespace/bar"`, Settings: map[string]any{"import/ignore": []any{"foo"}}},
			{Code: `import * as a from "./commonjs-namespace/a"; a.b`},
			{Code: `import { foo } from "./ignore.invalid.extension"`},
			{Code: `import { ExtfieldModel, Extfield2Model } from './models';`, FileName: "export-star/downstream.js"},
			{Code: `const { something } = require("./dynamic-import-in-commonjs")`, Options: []any{map[string]any{"commonjs": true}}},
			{Code: `import { something } from "./dynamic-import-in-commonjs"`},
			{Code: `import { "foo" as foo } from "./bar"`},
			{Code: `import { "foo" as foo } from "./empty-module"`},
		}, []rule_tester.InvalidTestCase{
			{Code: `import { somethingElse } from "./test-module"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "somethingElse not found in './test-module'", Line: 1, Column: 10, EndLine: 1, EndColumn: 23}}},
			{Code: `import { baz } from "./bar"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "baz not found in './bar'", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}}},
			{Code: `import { baz, bop } from "./bar"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "baz not found in './bar'", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}, {Message: "bop not found in './bar'", Line: 1, Column: 15, EndLine: 1, EndColumn: 18}}},
			{Code: `import {a, b, c} from "./named-exports"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "c not found in './named-exports'", Line: 1, Column: 15, EndLine: 1, EndColumn: 16}}},
			{Code: `import { a } from "./default-export"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "a not found in './default-export'", Line: 1, Column: 10, EndLine: 1, EndColumn: 11}}},
			{Code: `import { ActionTypess } from "./qc"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "ActionTypess not found in './qc'", Line: 1, Column: 10, EndLine: 1, EndColumn: 22}}},
			{Code: `import {a, b, c, d, e} from "./re-export"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "e not found in './re-export'", Line: 1, Column: 21, EndLine: 1, EndColumn: 22}}},
			{Code: `import { a } from "./re-export-names"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "a not found in './re-export-names'", Line: 1, Column: 10, EndLine: 1, EndColumn: 11}}},
			{Code: `export { bar } from "./bar"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "bar not found in './bar'", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}}},
			// Babel export-default-from declarations are not parsed by tsgo.
			{Code: `export bar2, { bar } from "./bar"`, Skip: true, Errors: []rule_tester.InvalidTestCaseError{{Message: "bar not found in './bar'", Line: 1, Column: 16, EndLine: 1, EndColumn: 19}}},
			{Code: `import { foo, bar, baz } from "./named-trampoline"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "baz not found in './named-trampoline'", Line: 1, Column: 20, EndLine: 1, EndColumn: 23}}},
			{Code: `import { baz } from "./broken-trampoline"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "baz not found via broken-trampoline.js -> named-exports.js", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}}},
			{Code: `const { baz } = require("./bar")`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "baz not found in './bar'", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
			{Code: `let { baz } = require("./bar")`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "baz not found in './bar'", Line: 1, Column: 7, EndLine: 1, EndColumn: 10}}},
			{Code: `const { baz: bar, bop } = require("./bar"), { a } = require("./re-export-names")`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "baz not found in './bar'", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}, {Message: "bop not found in './bar'", Line: 1, Column: 19, EndLine: 1, EndColumn: 22}, {Message: "a not found in './re-export-names'", Line: 1, Column: 47, EndLine: 1, EndColumn: 48}}},
			{Code: `const { default: defExport } = require("./named-exports")`, Options: []any{map[string]any{"commonjs": true}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './named-exports'", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
			{Code: `import  { type MyOpaqueType, MyMissingClass } from "./flowtypes"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "MyMissingClass not found in './flowtypes'", Line: 1, Column: 30, EndLine: 1, EndColumn: 44}}},
			{Code: `/*jsnext*/ import { createSnorlax } from "./redux"`, Settings: map[string]any{"import/ignore": []any{}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "createSnorlax not found in './redux'", Line: 1, Column: 21, EndLine: 1, EndColumn: 34}}},
			{Code: `/*jsnext*/ import { createSnorlax } from "./redux"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "createSnorlax not found in './redux'", Line: 1, Column: 21, EndLine: 1, EndColumn: 34}}},
			{Code: `import { baz } from "./es6-module"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "baz not found in './es6-module'", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}}},
			{Code: `import { foo, bar, bap } from "./re-export-default"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "bap not found in './re-export-default'", Line: 1, Column: 20, EndLine: 1, EndColumn: 23}}},
			{Code: `import { default as barDefault } from "./re-export"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './re-export'", Line: 1, Column: 10, EndLine: 1, EndColumn: 17}}},
			{Code: `import { "somethingElse" as somethingElse } from "./test-module"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "somethingElse not found in './test-module'", Line: 1, Column: 10, EndLine: 1, EndColumn: 25}}},
			{Code: `import { "baz" as baz, "bop" as bop } from "./bar"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "baz not found in './bar'", Line: 1, Column: 10, EndLine: 1, EndColumn: 15}, {Message: "bop not found in './bar'", Line: 1, Column: 24, EndLine: 1, EndColumn: 29}}},
			{Code: `import { "default" as barDefault } from "./re-export"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "default not found in './re-export'", Line: 1, Column: 10, EndLine: 1, EndColumn: 19}}},
		})
}

// Upstream group: named (path case-insensitivity)
func TestNamedUpstreamCaseInsensitive(t *testing.T) {
	root := namedRoot(t, "testdata/upstream.txtar")
	if root.FS.UseCaseSensitiveFileNames() {
		t.Skip("upstream case-insensitive filesystem group")
	}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &named.NamedRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { b } from "./Named-Exports"`},
		}, []rule_tester.InvalidTestCase{
			{Code: `import { foo } from "./Named-Exports"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "foo not found in './Named-Exports'", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}}},
		})
}

// Upstream group: named (export *)
func TestNamedUpstreamExportStar(t *testing.T) {
	root := namedRoot(t, "testdata/upstream.txtar")
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &named.NamedRule,
		[]rule_tester.ValidTestCase{
			{Code: `import { foo } from "./export-all"`},
		}, []rule_tester.InvalidTestCase{
			{Code: `import { bar } from "./export-all"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "bar not found in './export-all'", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}}},
		})
}

// Upstream group: named [TypeScript]
func TestNamedUpstreamTypeScript(t *testing.T) {
	root := namedRoot(t, "testdata/upstream.txtar")
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &named.NamedRule,
		[]rule_tester.ValidTestCase{
			{Code: `import x from './typescript-export-assign-object'`},
			{Code: `import { MyType } from "./typescript"`},
			{Code: `import { Foo } from "./typescript"`},
			{Code: `import { Bar } from "./typescript"`},
			{Code: `import { getFoo } from "./typescript"`},
			{Code: `import { MyEnum } from "./typescript"`},
			{Code: `
              import { MyModule } from "./typescript"
              MyModule.ModuleFunction()
            `},
			{Code: `
              import { MyNamespace } from "./typescript"
              MyNamespace.NSModule.NSModuleFunction()
            `},
			{Code: `import { MyType } from "./typescript-declare"`},
			{Code: `import { Foo } from "./typescript-declare"`},
			{Code: `import { Bar } from "./typescript-declare"`},
			{Code: `import { getFoo } from "./typescript-declare"`},
			{Code: `import { MyEnum } from "./typescript-declare"`},
			{Code: `
              import { MyModule } from "./typescript-declare"
              MyModule.ModuleFunction()
            `},
			{Code: `
              import { MyNamespace } from "./typescript-declare"
              MyNamespace.NSModule.NSModuleFunction()
            `},
			{Code: `import { MyType } from "./typescript-export-assign-namespace"`},
			{Code: `import { Foo } from "./typescript-export-assign-namespace"`},
			{Code: `import { Bar } from "./typescript-export-assign-namespace"`},
			{Code: `import { getFoo } from "./typescript-export-assign-namespace"`},
			{Code: `import { MyEnum } from "./typescript-export-assign-namespace"`},
			{Code: `
              import { MyModule } from "./typescript-export-assign-namespace"
              MyModule.ModuleFunction()
            `},
			{Code: `
              import { MyNamespace } from "./typescript-export-assign-namespace"
              MyNamespace.NSModule.NSModuleFunction()
            `},
			{Code: `import { MyType } from "./typescript-export-assign-namespace-merged"`},
			{Code: `import { Foo } from "./typescript-export-assign-namespace-merged"`},
			{Code: `import { Bar } from "./typescript-export-assign-namespace-merged"`},
			{Code: `import { getFoo } from "./typescript-export-assign-namespace-merged"`},
			{Code: `import { MyEnum } from "./typescript-export-assign-namespace-merged"`},
			{Code: `
              import { MyModule } from "./typescript-export-assign-namespace-merged"
              MyModule.ModuleFunction()
            `},
			{Code: `
              import { MyNamespace } from "./typescript-export-assign-namespace-merged"
              MyNamespace.NSModule.NSModuleFunction()
            `},
		}, []rule_tester.InvalidTestCase{
			{Code: `import { NotExported } from './typescript-export-assign-object'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "NotExported not found in './typescript-export-assign-object'", Line: 1, Column: 10, EndLine: 1, EndColumn: 21}}},
			{Code: `import { FooBar } from './typescript-export-assign-object'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "FooBar not found in './typescript-export-assign-object'", Line: 1, Column: 10, EndLine: 1, EndColumn: 16}}},
			{Code: `import { MissingType } from "./typescript"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "MissingType not found in './typescript'", Line: 1, Column: 10, EndLine: 1, EndColumn: 21}}},
			{Code: `import { NotExported } from "./typescript"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "NotExported not found in './typescript'", Line: 1, Column: 10, EndLine: 1, EndColumn: 21}}},
			{Code: `import { MissingType } from "./typescript-declare"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "MissingType not found in './typescript-declare'", Line: 1, Column: 10, EndLine: 1, EndColumn: 21}}},
			{Code: `import { NotExported } from "./typescript-declare"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "NotExported not found in './typescript-declare'", Line: 1, Column: 10, EndLine: 1, EndColumn: 21}}},
			{Code: `import { MissingType } from "./typescript-export-assign-namespace"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "MissingType not found in './typescript-export-assign-namespace'", Line: 1, Column: 10, EndLine: 1, EndColumn: 21}}},
			{Code: `import { NotExported } from "./typescript-export-assign-namespace"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "NotExported not found in './typescript-export-assign-namespace'", Line: 1, Column: 10, EndLine: 1, EndColumn: 21}}},
			{Code: `import { MissingType } from "./typescript-export-assign-namespace-merged"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "MissingType not found in './typescript-export-assign-namespace-merged'", Line: 1, Column: 10, EndLine: 1, EndColumn: 21}}},
			{Code: `import { NotExported } from "./typescript-export-assign-namespace-merged"`, Errors: []rule_tester.InvalidTestCaseError{{Message: "NotExported not found in './typescript-export-assign-namespace-merged'", Line: 1, Column: 10, EndLine: 1, EndColumn: 21}}},
		})
}

// Upstream group: documentation
func TestNamedDocumentation(t *testing.T) {
	root := namedRoot(t, "testdata/upstream.txtar")
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &named.NamedRule,
		[]rule_tester.ValidTestCase{
			{Code: `export const foo = "I'm so foo"`},
			{Code: `import { foo } from './foo'`},
			{Code: `export { foo as bar } from './foo'`},
			// The docs assume React's CommonJS entry is ignored. Explicitly
			// ignore packages so an installed declaration entry cannot change it.
			{Code: `import { SomeNonsenseThatDoesntExist } from 'react'`, Settings: map[string]any{"import/ignore": []any{"node_modules"}}},
			{Code: `import { notWhatever } from './whatever'`, Settings: map[string]any{"import/ignore": []any{"node_modules", "\\.coffee$"}}},
		}, []rule_tester.InvalidTestCase{
			{Code: `import { notFoo } from './foo'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "notFoo not found in './foo'", Line: 1, Column: 10, EndLine: 1, EndColumn: 16}}},
			{Code: `export { notFoo as defNotBar } from './foo'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "notFoo not found in './foo'", Line: 1, Column: 10, EndLine: 1, EndColumn: 16}}},
			{Code: `import { dontCreateStore } from './redux'`, Errors: []rule_tester.InvalidTestCaseError{{Message: "dontCreateStore not found in './redux'", Line: 1, Column: 10, EndLine: 1, EndColumn: 25}}},
		})
}
