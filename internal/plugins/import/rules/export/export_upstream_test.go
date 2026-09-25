package export_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func exportRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "export-rule")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
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

// All active semantic cases (parser/version variants deduplicated) from
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/export.js
// and its docs/rules/export.md examples. SYNTAX_CASES are expanded in place.
// Positions and messages were checked with the pinned rule and the TS parser.
func TestExportUpstreamJavaScript(t *testing.T) {
	rule_tester.RunRuleTester(exportRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &export.ExportRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `import "./malformed.js"`,
				// The test Program rejects the malformed dependency before rules run.
				Skip: true,
			},
			{
				Code: `var foo = "foo"; export default foo;`,
			},
			{
				Code: `export var foo = "foo"; export var bar = "bar";`,
			},
			{
				Code: `export var foo = "foo", bar = "bar";`,
			},
			{
				Code: `export var { foo, bar } = object;`,
			},
			{
				Code: `export var [ foo, bar ] = array;`,
			},
			{
				Code: `let foo; export { foo, foo as bar }`,
			},
			{
				Code: `let bar; export { bar }; export * from "./export-all"`,
			},
			{
				Code: `export * from "./export-all"`,
			},
			{
				Code: `export * from "./does-not-exist"`,
			},
			{
				Code: `export default foo; export * from "./bar"`,
			},
			{
				Code: `for (let { foo, bar } of baz) {}`,
			},
			{
				Code: `for (let [ foo, bar ] of baz) {}`,
			},
			{
				Code: `const { x, y } = bar`,
			},
			{
				Code: `const { x, y, ...z } = bar`,
			},
			{
				Code: `let x; export { x }`,
			},
			{
				Code: `let x; export { x as y }`,
			},
			{
				Code: `export const x = null`,
			},
			{
				Code: `export var x = null`,
			},
			{
				Code: `export let x = null`,
			},
			{
				Code: `export default x`,
			},
			{
				Code: `export default class x {}`,
			},
			{
				Code:     `import json from "./data.json"`,
				Settings: map[string]interface{}{"import/extensions": []interface{}{".js"}},
			},
			{
				Code:     `import foo from "./foobar.json";`,
				Settings: map[string]interface{}{"import/extensions": []interface{}{".js"}},
			},
			{
				Code:     `import foo from "./foobar";`,
				Settings: map[string]interface{}{"import/extensions": []interface{}{".js"}},
			},
			{
				Code:     `import { foo } from "./issue-370-commonjs-namespace/bar"`,
				Settings: map[string]interface{}{"import/ignore": []interface{}{"foo"}},
			},
			{
				Code:     `export * from "./issue-370-commonjs-namespace/bar"`,
				Settings: map[string]interface{}{"import/ignore": []interface{}{"foo"}},
			},
			{
				Code: `import * as a from "./commonjs-namespace/a"; a.b`,
			},
			{
				Code: `import { foo } from "./ignore.invalid.extension"`,
			},
			{
				Code: `
        import * as A from './named-export-collision/a';
        import * as B from './named-export-collision/b';

        export { A, B };
      `,
			},
			{
				Code: `
        export * as A from './named-export-collision/a';
        export * as B from './named-export-collision/b';
      `,
			},
			{
				Code: `
        export default function foo(param: string): boolean;
        export default function foo(param: string, param1: number): boolean;
        export default function foo(param: string, param1?: number): boolean {
          return param && param1;
        }
      `,
			},
			{
				Code: `
        export default function foo(param: string): boolean;
        export default function foo(param: string, param1?: number): boolean {
          return param && param1;
        }
      `,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `let foo; export { foo }; export * from "./export-all"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'foo'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 22},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'foo'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 54},
				},
			},
			{
				Code: `export * from "./malformed.js"`,
				// tsgo accepts the top-level return rejected by Espree.
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "parseError", Message: "Parse errors in imported module './malformed.js': 'return' outside of function (1:1)"},
				},
			},
			{
				Code: `export * from "./default-export"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNamed", Message: "No named exports found in module './default-export'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code: `let foo; export { foo as "foo" }; export * from "./export-all"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'foo'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 31},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'foo'.", Line: 1, Column: 35, EndLine: 1, EndColumn: 63},
				},
			},
			{
				Code: `
        export default function a(): void;
        export default function a() {}
        export { x as default };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 3, Column: 9, EndLine: 3, EndColumn: 39},
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 4, Column: 23, EndLine: 4, EndColumn: 30},
				},
			},
		},
	)
}

func TestExportUpstreamTypeScript(t *testing.T) {
	rule_tester.RunRuleTester(exportRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &export.ExportRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `
            export const Foo = 1;
            export type Foo = number;
          `,
			},
			{
				Code: `
            export const Foo = 1;
            export interface Foo {}
          `,
			},
			{
				Code: `
            export function fff(a: string);
            export function fff(a: number);
          `,
			},
			{
				Code: `
            export function fff(a: string);
            export function fff(a: number);
            export function fff(a: string|number) {};
          `,
			},
			{
				Code: `
            export const Bar = 1;
            export namespace Foo {
              export const Bar = 1;
            }
          `,
			},
			{
				Code: `
            export type Bar = string;
            export namespace Foo {
              export type Bar = string;
            }
          `,
			},
			{
				Code: `
            export const Bar = 1;
            export type Bar = string;
            export namespace Foo {
              export const Bar = 1;
              export type Bar = string;
            }
          `,
			},
			{
				Code: `
            export namespace Foo {
              export const Foo = 1;
              export namespace Bar {
                export const Foo = 2;
              }
              export namespace Baz {
                export const Foo = 3;
              }
            }
          `,
			},
			{
				Code: `
              export class Foo { }
              export namespace Foo { }
              export namespace Foo {
                export class Bar {}
              }
            `,
			},
			{
				Code: `
              export function Foo();
              export namespace Foo { }
            `,
			},
			{
				Code: `
              export function Foo(a: string);
              export namespace Foo { }
            `,
			},
			{
				Code: `
              export function Foo(a: string);
              export function Foo(a: number);
              export namespace Foo { }
            `,
			},
			{
				Code: `
              export enum Foo { }
              export namespace Foo { }
            `,
			},
			{
				Code:     `export * from "./file1.ts"`,
				FileName: "typescript-d-ts/file-2.ts",
			},
			{
				Code: `
              export * as A from './named-export-collision/a';
              export * as B from './named-export-collision/b';
            `,
			},
			{
				Code: `
            declare module "a" {
              const Foo = 1;
              export {Foo as default};
            }
            declare module "b" {
              const Bar = 2;
              export {Bar as default};
            }
          `,
			},
			{
				Code: `
            declare module "a" {
              const Foo = 1;
              export {Foo as default};
            }
            const Bar = 2;
            export {Bar as default};
          `,
			},
			{
				Code: `
            export * from './module';
          `,
				FileName: "export-star-4/index.ts",
				Settings: map[string]interface{}{"import/extensions": []interface{}{".js", ".ts", ".jsx"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
            export type Foo = string;
            export type Foo = number;
          `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 2, Column: 25, EndLine: 2, EndColumn: 28},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 3, Column: 25, EndLine: 3, EndColumn: 28},
				},
			},
			{
				Code: `
            export const a = 1
            export namespace Foo {
              export const a = 2;
              export const a = 3;
            }
          `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 4, Column: 28, EndLine: 4, EndColumn: 29},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 5, Column: 28, EndLine: 5, EndColumn: 29},
				},
			},
			{
				Code: `
            declare module 'foo' {
              const Foo = 1;
              export default Foo;
              export default Foo;
            }
          `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 4, Column: 15, EndLine: 4, EndColumn: 34},
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 5, Column: 15, EndLine: 5, EndColumn: 34},
				},
			},
			{
				Code: `
            export namespace Foo {
              export namespace Bar {
                export const Foo = 1;
                export const Foo = 2;
              }
              export namespace Baz {
                export const Bar = 3;
                export const Bar = 4;
              }
            }
          `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 4, Column: 30, EndLine: 4, EndColumn: 33},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 5, Column: 30, EndLine: 5, EndColumn: 33},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Bar'.", Line: 8, Column: 30, EndLine: 8, EndColumn: 33},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Bar'.", Line: 9, Column: 30, EndLine: 9, EndColumn: 33},
				},
			},
			{
				Code: `
              export class Foo { }
              export class Foo { }
              export namespace Foo { }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 2, Column: 28, EndLine: 2, EndColumn: 31},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 3, Column: 28, EndLine: 3, EndColumn: 31},
				},
			},
			{
				Code: `
              export enum Foo { }
              export enum Foo { }
              export namespace Foo { }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 2, Column: 27, EndLine: 2, EndColumn: 30},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 3, Column: 27, EndLine: 3, EndColumn: 30},
				},
			},
			{
				Code: `
              export enum Foo { }
              export class Foo { }
              export namespace Foo { }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 2, Column: 27, EndLine: 2, EndColumn: 30},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 3, Column: 28, EndLine: 3, EndColumn: 31},
				},
			},
			{
				Code: `
              export const Foo = 'bar';
              export class Foo { }
              export namespace Foo { }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 2, Column: 28, EndLine: 2, EndColumn: 31},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 3, Column: 28, EndLine: 3, EndColumn: 31},
				},
			},
			{
				Code: `
              export function Foo() { };
              export class Foo { }
              export namespace Foo { }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 2, Column: 31, EndLine: 2, EndColumn: 34},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 3, Column: 28, EndLine: 3, EndColumn: 31},
				},
			},
			{
				Code: `
              export const Foo = 'bar';
              export function Foo() { };
              export namespace Foo { }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 2, Column: 28, EndLine: 2, EndColumn: 31},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 3, Column: 31, EndLine: 3, EndColumn: 34},
				},
			},
			{
				Code: `
              export const Foo = 'bar';
              export namespace Foo { }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 2, Column: 28, EndLine: 2, EndColumn: 31},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 3, Column: 32, EndLine: 3, EndColumn: 35},
				},
			},
			{
				Code: `
            declare module "a" {
              const Foo = 1;
              export {Foo as default};
            }
            const Bar = 2;
            export {Bar as default};
            const Baz = 3;
            export {Baz as default};
          `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 7, Column: 28, EndLine: 7, EndColumn: 35},
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 9, Column: 28, EndLine: 9, EndColumn: 35},
				},
			},
		},
	)
}

func TestExportUpstreamDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(exportRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &export.ExportRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code: `export default class MyClass { /*...*/ } // Multiple default exports.

function makeClass() { return new MyClass(...arguments) }

export default makeClass // Multiple default exports.`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 5, Column: 1, EndLine: 5, EndColumn: 25},
				},
			},
			{
				Code: `export const foo = function () { /*...*/ } // Multiple exports of name 'foo'.

function bar() { /*...*/ }
export { bar as foo } // Multiple exports of name 'foo'.`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'foo'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 17},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'foo'.", Line: 4, Column: 17, EndLine: 4, EndColumn: 20},
				},
			},
		},
	)
}
