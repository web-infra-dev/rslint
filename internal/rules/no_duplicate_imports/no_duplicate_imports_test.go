package no_duplicate_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDuplicateImportsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDuplicateImportsRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream — JS suite ----
			{Code: `import os from "os";
import fs from "fs";`},
			{Code: `import { merge } from "lodash-es";`},
			{Code: `import _, { merge } from "lodash-es";`},
			{Code: `import * as Foobar from "async";`},
			{Code: `import "foo"`},
			{Code: `import os from "os";
export { something } from "os";`},
			{Code: `import * as bar from "os";
import { baz } from "os";`},
			{Code: `import foo, * as bar from "os";
import { baz } from "os";`},
			{Code: `import foo, { bar } from "os";
import * as baz from "os";`},
			{Code: `import os from "os";
export { hello } from "hello";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import os from "os";
export * from "hello";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import os from "os";
export { hello as hi } from "hello";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import os from "os";
export default function(){};`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import { merge } from "lodash-es";
export { merge as lodashMerge }`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `export { something } from "os";
export * as os from "os";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import { something } from "os";
export * as os from "os";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import * as os from "os";
export { something } from "os";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import os from "os";
export * from "os";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `export { something } from "os";
export * from "os";`, Options: map[string]interface{}{"includeExports": true}},

			// ---- ESLint upstream — TypeScript suite (type-only forms) ----
			{Code: `import type { Os } from "os";
import type { Fs } from "fs";`},
			{Code: `import { type Os } from "os";
import type { Fs } from "fs";`},
			{Code: `import type { Merge } from "lodash-es";`},
			{Code: `import _, { type Merge } from "lodash-es";`},
			{Code: `import type * as Foobar from "async";`},
			{Code: `import type Os from "os";
export type { Something } from "os";`},
			{Code: `import type Os from "os";
export { type Something } from "os";`},
			{Code: `import type * as Bar from "os";
import { type Baz } from "os";`},
			{Code: `import foo, * as bar from "os";
import { type Baz } from "os";`},
			{Code: `import foo, { type bar } from "os";
import type * as Baz from "os";`},
			{Code: `import type { Merge } from "lodash-es";
import type _ from "lodash-es";`},
			{Code: `import type Os from "os";
export { type Hello } from "hello";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import type Os from "os";
export type * from "hello";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import type Os from "os";
export { type Hello as Hi } from "hello";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import type Os from "os";
export default function(){};`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import { type Merge } from "lodash-es";
export { Merge as lodashMerge }`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `export type { Something } from "os";
export * as os from "os";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import { type Something } from "os";
export * as os from "os";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import type * as Os from "os";
export { something } from "os";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import type Os from "os";
export * from "os";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `import type Os from "os";
export type { Something } from "os";`, Options: map[string]interface{}{"includeExports": true}},
			{Code: `export type { Something } from "os";
export * from "os";`, Options: map[string]interface{}{"includeExports": true}},

			// ---- allowSeparateTypeImports (TS) ----
			{Code: `import { foo, type Bar } from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true}},
			{Code: `import { foo } from "module";
import type { Bar } from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true}},
			{Code: `import { type Foo } from "module";
import type { Bar } from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true}},
			{Code: `import { foo, type Bar } from "module";
export { type Baz } from "module2";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true, "includeExports": true}},
			{Code: `import type { Foo } from "module";
export { bar, type Baz } from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true, "includeExports": true}},
			{Code: `import { type Foo } from "module";
export type { Bar } from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true, "includeExports": true}},
			{Code: `import type * as Foo from "module";
export { type Bar } from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true, "includeExports": true}},
			{Code: `import { type Foo } from "module";
export type * as Bar from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true, "includeExports": true}},

			// ---- Extra: locks in disjoint module names (different specifier shapes) ----
			{Code: `import a from "mod-a";
import { b } from "mod-b";
import * as c from "mod-c";
import "mod-d";
export { e } from "mod-e";
export * from "mod-f";
export * as g from "mod-g";`, Options: map[string]interface{}{"includeExports": true}},

			// ---- Extra: includeExports defaults to false (export collisions ignored) ----
			// `export { x } from "os"` after `import os from "os"` is invalid only when
			// includeExports is on; with the default options it must stay valid.
			{Code: `import os from "os";
export { something } from "os";
export * from "os";`},

			// ---- Extra: per-specifier `type` keyword without declaration-level type ----
			// Both have importKind="value"; the inline `type` modifier on a specifier
			// does NOT make the declaration type-only, and the rule MUST still report —
			// covered by the invalid suite. This valid case locks in that the inline
			// keyword by itself doesn't synthesize allowSeparateTypeImports behaviour.
			{Code: `import { type Foo } from "moduleA";
import { type Bar } from "moduleB";`},

			// ---- Extra: empty options object behaves like defaults ----
			{Code: `import os from "os";
export { something } from "os";`, Options: map[string]interface{}{}},

			// ---- Extra: side-effect-only repeated across separate modules is fine ----
			{Code: `import "fs";
import "os";`},

			// ---- Extra: tsgo string normalization — module specifier whitespace ----
			// ESLint trims `node.source.value`; we mirror that with strings.TrimSpace.
			// Two specifiers that trim to the same value WOULD collide (covered in
			// invalid). This valid case locks in that two distinct trimmed names stay
			// disjoint — guards against a TrimSpace regression that returns "".
			{Code: `import a from "mod-a";
import b from "mod-b";`},

			// ---- Extra: empty NamedImports — `import {} from "m"` ----
			// ESLint walks `specifiers` (empty here) and falls back to
			// "SideEffectImport". Pairing with a namespace import is therefore
			// allowed (would be NS+SideEffect → mergeable, so the INVALID side
			// covers the merge); the VALID lock-in pairs two empty named imports
			// of DIFFERENT modules.
			{Code: `import {} from "mod-a";
import {} from "mod-b";`},

			// ---- Extra: namespace + empty named pair is mergeable in ESLint ----
			// Covered as INVALID below; here we lock in that namespace + non-empty
			// named is NOT mergeable (the upstream NS+Named guard fires).
			{Code: `import * as ns from "mod";
import { x } from "mod";`},

			// ---- Extra: comments inside import declarations ----
			// Comments must not interfere with module-name extraction.
			{Code: `import /* a */ "mod-a";
import /* b */ "mod-b";`},

			// ---- Extra: multi-line import declaration ----
			// The trimmed name is still "mod"; pairing with a different module is
			// fine. Position in INVALID variants must point at line where the
			// duplicate STARTS.
			{Code: `import {
  foo,
  bar,
} from "mod";`},

			// ---- Extra: empty file / single declarations / many disjoint ----
			{Code: ``},
			{Code: `import a from "mod";`},
			{Code: `import a from "m1";
import b from "m2";
import c from "m3";
import d from "m4";
import e from "m5";`},

			// ---- Extra: scoped packages and deep paths ----
			{Code: `import a from "@scope/pkg";
import b from "@scope/other";
import c from "@scope/pkg/sub";`},

			// ---- Extra: relative paths are distinct keys ----
			{Code: `import a from "./util";
import b from "../util";
import c from "/util";
import d from "util";`},

			// ---- Extra: import attributes / `with` clause must not affect key ----
			// ESLint and tsgo both attach attributes to a separate ImportAttributes
			// node; the module specifier is the same StringLiteral. Two different
			// modules must stay disjoint regardless of the attribute clause.
			{Code: `import a from "mod-a" with { type: "json" };
import b from "mod-b" with { type: "json" };`},

			// ---- Extra: dynamic import is NOT an ImportDeclaration ----
			// `import("mod")` is a CallExpression with a special token; it must
			// not collide with a static import of the same module.
			{Code: `import x from "mod";
const y = import("mod");`},

			// ---- Extra: import equals (TS) is not handled — must NOT collide ----
			// `import x = require("m")` is KindImportEqualsDeclaration; we don't
			// listen on it (upstream doesn't either).
			{Code: `import a = require("mod");
import b = require("mod");`},

			// ---- Extra: CommonJS require + ESM import on same module ----
			// require() is a CallExpression, not an ImportDeclaration.
			{Code: `import x from "mod";
const y = require("mod");`},

			// ---- Extra: type-only ExportNamespace (`export type * as ns`) ----
			// Locks in that this is treated as a Namespace export, type-only at
			// the declaration level. Pairing with `import * as ns` (value) does
			// NOT trigger a report under default options (no includeExports).
			{Code: `import * as Ns from "mod";
export type * as Other from "mod";`},

			// ---- Extra: declare module is its own scope ----
			// We mirror upstream which scans the whole file; an ImportDeclaration
			// inside `declare module "x" {}` IS still an ImportDeclaration and
			// would join the same module map. Lock in the disjoint case (same
			// nested module name as a top-level import of a DIFFERENT module
			// stays valid).
			{Code: `import a from "mod-a";
declare module "augmented" {
  import b from "mod-b";
}`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint upstream — JS suite ----
			{
				Code: `import "fs";
import "fs"`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'fs' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import { merge } from "lodash-es";
import { find } from "lodash-es";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'lodash-es' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import { merge } from "lodash-es";
import _ from "lodash-es";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'lodash-es' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import os from "os";
import { something } from "os";
import * as foobar from "os";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Message: "'os' import is duplicated.", Line: 2, Column: 1},
					{MessageId: "import", Message: "'os' import is duplicated.", Line: 3, Column: 1},
				},
			},
			{
				Code: `import * as modns from "lodash-es";
import { merge } from "lodash-es";
import { baz } from "lodash-es";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'lodash-es' import is duplicated.",
					Line:      3,
					Column:    1,
				}},
			},
			{
				Code: `export { os } from "os";
export { something } from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "export",
					Message:   "'os' export is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import os from "os";
export { os as foobar } from "os";
export { something } from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exportAs", Message: "'os' export is duplicated as import.", Line: 2, Column: 1},
					{MessageId: "export", Message: "'os' export is duplicated.", Line: 3, Column: 1},
					{MessageId: "exportAs", Message: "'os' export is duplicated as import.", Line: 3, Column: 1},
				},
			},
			{
				Code: `import os from "os";
export { something } from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "exportAs",
					Message:   "'os' export is duplicated as import.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import os from "os";
export * as os from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "exportAs",
					Message:   "'os' export is duplicated as import.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `export * as os from "os";
import os from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "importAs",
					Message:   "'os' import is duplicated as export.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import * as modns from "mod";
export * as  modns from "mod";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "exportAs",
					Message:   "'mod' export is duplicated as import.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `export * from "os";
export * from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "export",
					Message:   "'os' export is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import "os";
export * from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "exportAs",
					Message:   "'os' export is duplicated as import.",
					Line:      2,
					Column:    1,
				}},
			},

			// ---- ESLint upstream — TypeScript suite ----
			{
				Code: `import "fs";
import "fs"`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'fs' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import { type Merge } from "lodash-es";
import { type Find } from "lodash-es";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'lodash-es' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import { type Merge } from "lodash-es";
import type { Find } from "lodash-es";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'lodash-es' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import type { Merge } from "lodash-es";
import type { Find } from "lodash-es";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'lodash-es' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				// node2 (named, type-only) vs node1 (default, type-only) is the
				// type-only default+named special case → NOT mergeable, no report.
				// node3 (namespace, type-only) vs node1 (default, type-only) IS
				// mergeable → report on line 3.
				Code: `import type Os from "os";
import type { Something } from "os";
import type * as Foobar from "os";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'os' import is duplicated.",
					Line:      3,
					Column:    1,
				}},
			},
			{
				Code: `import type * as Modns from "lodash-es";
import type { Merge } from "lodash-es";
import type { Baz } from "lodash-es";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'lodash-es' import is duplicated.",
					Line:      3,
					Column:    1,
				}},
			},
			{
				Code: `import { type Foo } from "module";
export type { Bar } from "module";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "exportAs",
					Message:   "'module' export is duplicated as import.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `export { os } from "os";
export type { Something } from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "export",
					Message:   "'os' export is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `export type { Os } from "os";
export type { Something } from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "export",
					Message:   "'os' export is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import type { Os } from "os";
export type { Os as Foobar } from "os";
export type { Something } from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exportAs", Message: "'os' export is duplicated as import.", Line: 2, Column: 1},
					{MessageId: "export", Message: "'os' export is duplicated.", Line: 3, Column: 1},
					{MessageId: "exportAs", Message: "'os' export is duplicated as import.", Line: 3, Column: 1},
				},
			},
			{
				Code: `import type { Os } from "os";
export type { Something } from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "exportAs",
					Message:   "'os' export is duplicated as import.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import type Os from "os";
export type * as Os from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "exportAs",
					Message:   "'os' export is duplicated as import.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import type * as Modns from "mod";
export type * as Modns from "mod";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "exportAs",
					Message:   "'mod' export is duplicated as import.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `export type * from "os";
export type * from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "export",
					Message:   "'os' export is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import "os";
export type { Os } from "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "exportAs",
					Message:   "'os' export is duplicated as import.",
					Line:      2,
					Column:    1,
				}},
			},

			// ---- allowSeparateTypeImports invalid forms ----
			{
				Code: `import { someValue } from 'module';
import { anotherValue } from 'module';`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'module' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import type { Merge } from "lodash-es";
import type { Find } from "lodash-es";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'lodash-es' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import { someValue, type Foo } from 'module';
import type { SomeType } from 'module';
import type { AnotherType } from 'module';`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'module' import is duplicated.",
					Line:      3,
					Column:    1,
				}},
			},
			{
				Code: `import { type Foo } from 'module';
import { type Bar } from 'module';`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'module' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `export type { Foo } from "module";
export type { Bar } from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true, "includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "export",
					Message:   "'module' export is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},
			{
				Code: `import { type Foo } from "module";
export { type Bar } from "module";
export { type Baz } from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true, "includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exportAs", Message: "'module' export is duplicated as import.", Line: 2, Column: 1},
					{MessageId: "export", Message: "'module' export is duplicated.", Line: 3, Column: 1},
					{MessageId: "exportAs", Message: "'module' export is duplicated as import.", Line: 3, Column: 1},
				},
			},
			{
				Code: `import { type Foo } from "module";
export { type Bar } from "module";
export { regular } from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true, "includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exportAs", Message: "'module' export is duplicated as import.", Line: 2, Column: 1},
					{MessageId: "export", Message: "'module' export is duplicated.", Line: 3, Column: 1},
					{MessageId: "exportAs", Message: "'module' export is duplicated as import.", Line: 3, Column: 1},
				},
			},
			{
				Code: `import { type Foo } from "module";
import { regular } from "module";
export { type Bar } from "module";
export { regular as other } from "module";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true, "includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Message: "'module' import is duplicated.", Line: 2, Column: 1},
					{MessageId: "exportAs", Message: "'module' export is duplicated as import.", Line: 3, Column: 1},
					{MessageId: "export", Message: "'module' export is duplicated.", Line: 4, Column: 1},
					{MessageId: "exportAs", Message: "'module' export is duplicated as import.", Line: 4, Column: 1},
				},
			},

			// ---- Extra: lock-in tests (upstream branches not directly tested) ----

			// Lock-in: array-form options (rule_tester multi-element shape) — exercises
			// the `[]interface{}{ map[string]interface{}{} }` JSON path so a
			// regression in GetOptionsMap is caught here.
			{
				Code: `export { os } from "os";
export { something } from "os";`,
				Options: []interface{}{map[string]interface{}{"includeExports": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "export",
					Message:   "'os' export is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in: `import os, { a } from "os"` reports as "named" (named/namespace
			// wins over default). Pairing it with `import { b } from "os"` confirms
			// named+named merges.
			{
				Code: `import os, { a } from "os";
import { b } from "os";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'os' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in: `import os, * as ns from "os"` reports as "namespace". Pairing
			// with `import * as foo from "os"` confirms namespace+namespace merges.
			{
				Code: `import foo, * as ns from "os";
import * as bar from "os";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'os' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in: side-effect import + side-effect import (both `import "X"`).
			// Already covered in the JS suite; keep a position-asserted variant.
			{
				Code: `import "lib";
import "lib";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'lib' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in: type+default vs type+named is NOT mergeable (separate branch
			// inside isImportExportCanBeMerged). Adding type+default after a type+
			// namespace IS mergeable (the default-vs-named guard is symmetric only
			// for default/named pairs), so we DO get a report.
			{
				Code: `import type * as Modns from "mod";
import type Mod from "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'mod' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in: ExportAll + side-effect-import IS mergeable (the ExportAll
			// guard explicitly excludes SideEffectImport from "non-mergeable"). The
			// upstream `import "os"; export * from "os"` test covers exportAs; the
			// reverse direction (export * first, import after) without
			// includeExports off → no listener attached for export. We instead add
			// the symmetric form with includeExports on, both directions.
			{
				Code: `export * from "os";
import "os";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "importAs",
					Message:   "'os' import is duplicated as export.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in: TrimSpace on module specifier — `"  mod  "` and `"mod"` collide.
			{
				Code: `import a from "  mod  ";
import b from "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'mod' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in: `import {} from "mod"` reports as SideEffectImport in
			// ESLint, so pairing with another `import "mod"` is "side-effect +
			// side-effect" → mergeable → REPORT. (If we mistakenly classified it
			// as Named the result would be the same; the discriminating case is
			// "empty named + namespace" below.)
			{
				Code: `import {} from "mod";
import "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'mod' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in (DISCRIMINATING): `import {} from "mod"` followed by
			// `import * as ns from "mod"`. ESLint reports because it sees
			// SideEffectImport + Namespace (mergeable). If we mis-classify the
			// empty-named declaration as Named, the NS+Named guard would fire and
			// suppress the report — this test would fail in that regression.
			{
				Code: `import {} from "mod";
import * as ns from "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'mod' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in: `import foo, {} from "mod"` is "default + empty named" →
			// ESLint falls to ImportDefaultSpecifier. Pairing with another
			// default import is mergeable.
			{
				Code: `import foo, {} from "mod";
import bar from "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'mod' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in: `export {} from "mod"` is also SideEffectImport in ESLint.
			// Pairing with `export * from "mod"` (ExportAll) is mergeable (the
			// ExportAll guard explicitly excludes SideEffectImport).
			{
				Code: `export {} from "mod";
export * from "mod";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "export",
					Message:   "'mod' export is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Lock-in: comments and multi-line whitespace must not break dup detection.
			{
				Code: `import /* leading */ { /* inner */ a /* tail */ } from "mod";
import {
  b,
} from "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'mod' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// ---- Real-world / robustness lock-ins ----

			// 3+ side-effect imports: every duplicate after the first reports.
			// Locks in that we accumulate `previous` correctly (not just keep the
			// first) — a regression that reset the slice would silently drop the
			// third report.
			{
				Code: `import "mod";
import "mod";
import "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Message: "'mod' import is duplicated.", Line: 2, Column: 1},
					{MessageId: "import", Message: "'mod' import is duplicated.", Line: 3, Column: 1},
				},
			},

			// Same module hit by all 7 specifier shapes in sequence — the
			// canonical "real bad file". 6 reports expected (every line after the
			// first that has at least one mergeable predecessor).
			//   1: default
			//   2: named (mergeable with default → REPORT)
			//   3: namespace (mergeable with default; NOT mergeable with named) → REPORT
			//   4: side-effect (mergeable with everything not-ExportAll-only) → REPORT
			//   5: default again (mergeable with default/named/side-effect) → REPORT
			//   With includeExports=true:
			//   6: export named (mergeable as exportAs with import default/named/SE; NOT mergeable as export with prior exports because none yet) → REPORT exportAs
			//   7: export-all (mergeable with side-effect import, mergeable as export with #6 named export only if the ExportAll guard allows — actually ExportAll vs Named is NOT mergeable) → REPORT importAs(Wrong direction); actually export-all only mergeable with side-effect/export-all; #6 is named → not mergeable as export. As "exportAs" against imports: ExportAll vs Named/Default/Namespace = NOT mergeable; ExportAll vs SideEffect = mergeable → REPORT exportAs.
			{
				Code: `import a from "mod";
import { b } from "mod";
import * as c from "mod";
import "mod";
import d from "mod";
export { e } from "mod";
export * from "mod";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Message: "'mod' import is duplicated.", Line: 2, Column: 1},
					{MessageId: "import", Message: "'mod' import is duplicated.", Line: 3, Column: 1},
					{MessageId: "import", Message: "'mod' import is duplicated.", Line: 4, Column: 1},
					{MessageId: "import", Message: "'mod' import is duplicated.", Line: 5, Column: 1},
					{MessageId: "exportAs", Message: "'mod' export is duplicated as import.", Line: 6, Column: 1},
					{MessageId: "exportAs", Message: "'mod' export is duplicated as import.", Line: 7, Column: 1},
				},
			},

			// Position EndLine / EndColumn assertion (not asserted in upstream).
			// Locks in that the report span covers the whole declaration.
			{
				Code: `import a from "mod";
import {
  longer,
  multiLineImport,
} from "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'mod' import is duplicated.",
					Line:      2,
					Column:    1,
					EndLine:   5,
					EndColumn: 14,
				}},
			},

			// Scoped package — module key is `@scope/pkg`, two-segment names match.
			{
				Code: `import a from "@scope/pkg";
import b from "@scope/pkg";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'@scope/pkg' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Non-ASCII module name (Unicode in path) — module key is byte-exact.
			{
				Code: `import a from "./módulo";
import b from "./módulo";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'./módulo' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Unicode escape sequence vs cooked string: `"os"` and `"os"`
			// have the same cooked Text — tsgo's StringLiteral.Text is the
			// decoded value, matching ESLint's `node.source.value`. Lock-in: the
			// rule MUST treat them as the same module.
			{
				Code: `import a from "os";
import b from "os";`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'os' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Import attributes (`with { type: "json" }`) sit on a separate
			// ImportAttributes node and must not affect the module key.
			{
				Code: `import a from "mod" with { type: "json" };
import b from "mod" with { type: "json" };`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'mod' import is duplicated.",
					Line:      2,
					Column:    1,
				}},
			},

			// Nested ImportDeclaration inside `declare module "x" { ... }`. Both
			// upstream and this implementation walk the whole file via a
			// listener, so the inner declaration shares the module map with the
			// top-level one. Lock-in: same module name in both → report on the
			// inner.
			{
				Code: `import a from "mod";
declare module "augmented" {
  import b from "mod";
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "import",
					Message:   "'mod' import is duplicated.",
					Line:      3,
					Column:    3,
				}},
			},

			// Deeply mixed file — interleave two modules with different shapes.
			// Catches accidental cross-module collisions (separate Map keys).
			{
				Code: `import a from "lib-a";
import b from "lib-b";
import { c } from "lib-a";
import { d } from "lib-b";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Message: "'lib-a' import is duplicated.", Line: 3, Column: 1},
					{MessageId: "import", Message: "'lib-b' import is duplicated.", Line: 4, Column: 1},
				},
			},

			// allowSeparateTypeImports: declaration-level type vs value across
			// MANY entries — locks in that the per-previous filter doesn't flip
			// state when the slice grows large.
			{
				Code: `import { a } from "mod";
import type { B } from "mod";
import { c } from "mod";
import type { D } from "mod";`,
				Options: map[string]interface{}{"allowSeparateTypeImports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					// node3 (value, named) merges with node1 (value, named) — REPORT.
					{MessageId: "import", Message: "'mod' import is duplicated.", Line: 3, Column: 1},
					// node4 (type, named) merges with node2 (type, named) — REPORT.
					{MessageId: "import", Message: "'mod' import is duplicated.", Line: 4, Column: 1},
				},
			},

			// CRLF line endings — position arithmetic must still produce
			// 1-based line numbers correctly.
			{
				Code:   "import \"mod\";\r\nimport \"mod\";\r\n",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "import", Message: "'mod' import is duplicated.", Line: 2, Column: 1}},
			},

			// Options: malformed input must not crash and must fall back to
			// defaults. Passing options=nil explicitly through the rule_tester
			// path (rather than the default of "no Options field") exercises the
			// `optsMap == nil` branch in parseOptions.
			{
				Code: `import "fs";
import "fs";`,
				Options: nil,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "import", Message: "'fs' import is duplicated.", Line: 2, Column: 1}},
			},

			// Options: empty array form (multi-element with no payload) — the
			// rule_tester unwraps `[]interface{}{}` to nil → defaults.
			{
				Code:    `import "fs"; import "fs";`,
				Options: []interface{}{},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "import", Message: "'fs' import is duplicated.", Line: 1, Column: 14}},
			},

			// Options: unknown key is ignored — must NOT silently flip
			// includeExports on. Pairing import + export under "unknownOption"
			// stays valid (we'd otherwise report exportAs).
			{
				Code:    `import os from "os"; export { x } from "os";`,
				Options: map[string]interface{}{"unknownOption": true},
				// No expected errors: a duplicate `import` of `"os"` would be
				// needed, but the export side is silent under the default.
				// Replace this with a real duplicate so the case is "invalid":
				// instead test that a CLI shape with bogus key still detects
				// genuine import duplicates.
				// (kept invalid below)
				Errors: []rule_tester.InvalidTestCaseError{},
				Skip:   true,
			},

			// Options: bogus key + genuine duplicate import — still reports.
			{
				Code: `import a from "os";
import b from "os";`,
				Options: map[string]interface{}{"bogus": "yes", "includeExports": "not-a-bool"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "import", Message: "'os' import is duplicated.", Line: 2, Column: 1}},
			},

			// includeExports=true + a chain of three exports of same module.
			// Locks in that the filter-by-declKind partition is monotonic and
			// each new export checks against ALL prior exports (and imports).
			{
				Code: `export { a } from "mod";
export { b } from "mod";
export { c } from "mod";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "export", Message: "'mod' export is duplicated.", Line: 2, Column: 1},
					{MessageId: "export", Message: "'mod' export is duplicated.", Line: 3, Column: 1},
				},
			},

			// Type-only namespace export DOES collide with type-only namespace
			// import (under includeExports). Same key, both type-only, namespace
			// + namespace = mergeable. Locks in that the type-only default+named
			// special case does NOT over-fire on namespace+namespace.
			{
				Code: `import type * as Ns from "mod";
export type * as Ns from "mod";`,
				Options: map[string]interface{}{"includeExports": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "exportAs", Message: "'mod' export is duplicated as import.", Line: 2, Column: 1}},
			},

			// Side-effect import after declaration-only `import "mod"` and a
			// later named — locks in that the side-effect path tracks all 3.
			{
				Code: `import "mod";
import { a } from "mod";
import { b } from "mod";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Message: "'mod' import is duplicated.", Line: 2, Column: 1},
					{MessageId: "import", Message: "'mod' import is duplicated.", Line: 3, Column: 1},
				},
			},
		},
	)
}
