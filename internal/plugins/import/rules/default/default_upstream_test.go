package default_rule_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	default_rule "github.com/web-infra-dev/rslint/internal/plugins/import/rules/default"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const noDefaultFromNamedExports = `No default export found in imported module "./named-exports".`

// TestDefaultUpstream migrates the valid/invalid suite from upstream
// tests/src/rules/default.js 1:1 where tsgo supports the syntax and framework
// hooks. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in default_extras_test.go.
func TestDefaultUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&default_rule.DefaultRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid: no default specifier ----
			{Code: `import "./malformed.js"`},
			{Code: `import { foo } from "./default-export";`},
			{Code: `export {a} from "./named-exports"`},

			// ---- upstream valid: unresolved / non-ES modules return no export map ----
			{Code: `import foo from "./empty-folder";`},
			{Code: `import crypto from "crypto";`},
			{Code: `import common from "./common";`},

			// ---- upstream valid: direct default exports ----
			{Code: `import foo from "./default-export";`},
			{Code: `import foo from "./mixed-exports";`},
			{Code: `import bar from "./default-export";`},
			{Code: `import CoolClass from "./default-class";`},
			{Code: `import bar, { baz } from "./default-export";`},
			{Code: `import connectedApp from "./redux"`},
			{Code: `import MyCoolComponent from "./jsx/MyCoolComponent.jsx"`, TSConfig: "tsconfig.allow-js.json"},
			{Code: `import App from "./jsx/App"`, TSConfig: "tsconfig.allow-js.json"},
			{Code: `import Foo from './jsx/FooES7.js';`, TSConfig: "tsconfig.allow-js.json"},

			// ---- upstream valid: default is exported under the name default ----
			{Code: `import foo from "./named-default-export"`},

			// ---- upstream valid: default re-export chains ----
			{Code: `import twofer from "./trampoline"`},
			{Code: `import bar from "./default-export-from";`},
			{Code: `import bar from "./default-export-from-named";`},

			// ---- upstream valid: TypeScript export assignment forms ----
			{Code: `import foobar from "./typescript-default"`},
			{Code: `import foobar from "./typescript-export-assign-default"`},
			{Code: `import foobar from "./typescript-export-assign-function"`},
			{Code: `import foobar from "./typescript-export-assign-mixed"`},
			{Code: `import foobar from "./typescript-export-assign-default-reexport"`},
			{Code: `import foobar from "./typescript-export-assign-property"`},

			// ---- upstream valid: named default re-export is not this rule's default specifier branch ----
			{Code: `export { default as bar } from "./bar"`},
			{Code: `export { default as bar, foo } from "./bar"`},
			{Code: `export { "default" as bar } from "./bar"`},

			// SKIP: rslint does not parse Babel's default re-export proposal `export foo from`.
			{Code: `export bar from "./bar"`, Skip: true},
			// SKIP: rslint does not parse Babel's default re-export proposal `export foo, { bar } from`.
			{Code: `export bar, { foo } from "./bar"`, Skip: true},
			// SKIP: rslint does not parse Babel's default re-export proposal `export foo, * as ns from`.
			{Code: `export bar, * as names from "./bar"`, Skip: true},
			{Code: `import bar from './default-export-from-ignored.js';`, Settings: map[string]interface{}{"import/ignore": []interface{}{"common"}}, TSConfig: "tsconfig.allow-js.json"},
			// SKIP: rslint does not parse Babel's default re-export proposal `export foo from`.
			{Code: `export bar from './default-export-from-ignored.js';`, Skip: true},
			{Code: `import React from "./typescript-export-assign-default-namespace"`},
			{Code: `import Foo from "./typescript-export-as-default-namespace"`},
			{Code: `import Foo from "./typescript-export-react-test-renderer"`},
			{Code: `import Foo from "./typescript-extended-config"`},
			// SKIP: upstream only runs this case on case-insensitive filesystems.
			{Code: `import foo from "./jsx/MyUncoolComponent.jsx"`, Skip: true},

			// ---- upstream valid: SYNTAX_CASES parser smoke cases with no default-import diagnostics ----
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
			{Code: `import json from "./data.json"`},
			{Code: `import foo from "./foobar.json";`},
			{Code: `import foo from "./foobar";`},
			{Code: `import { foo } from "./issue-370-commonjs-namespace/bar"`},
			{Code: `export * from "./issue-370-commonjs-namespace/bar"`},
			{Code: `import * as a from "./commonjs-namespace/a"; a.b`},
			{Code: `import { foo } from "./ignore.invalid.extension"`},
		},
		[]rule_tester.InvalidTestCase{
			// SKIP: upstream's default parser rejects class fields in FooES7.js; tsgo parses that fixture.
			{
				Code: `import Foo from './jsx/FooES7.js';`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault"},
				},
			},
			{
				Code:     `import baz from "./named-exports";`,
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noDefault",
						Message:   noDefaultFromNamedExports,
						Line:      1,
						Column:    8,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},
			// SKIP: rslint does not parse Babel's default re-export proposal `export foo from`.
			{
				Code: `export baz from "./named-exports"`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault"},
				},
			},
			// SKIP: rslint does not parse Babel's default re-export proposal `export foo, { bar } from`.
			{
				Code: `export baz, { bar } from "./named-exports"`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault"},
				},
			},
			// SKIP: rslint does not parse Babel's default re-export proposal `export foo, * as ns from`.
			{
				Code: `export baz, * as names from "./named-exports"`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault"},
				},
			},
			{
				Code:     `import twofer from "./broken-trampoline"`,
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: `No default export found in imported module "./broken-trampoline".`, Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: `import barDefault from "./re-export"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: `No default export found in imported module "./re-export".`, Line: 1, Column: 8, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:     `import foobar from "./typescript"`,
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: `No default export found in imported module "./typescript".`, Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
				},
			},
			// SKIP: upstream only runs this case on case-insensitive filesystems.
			{
				Code: `import bar from "./Named-Exports"`,
				Skip: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault"},
				},
			},
			// Upstream invalid branch without parserOptions.tsconfigRootDir.
			{
				Code:     `import React from "./typescript-export-assign-default-namespace"`,
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: `No default export found in imported module "./typescript-export-assign-default-namespace".`, Line: 1, Column: 8, EndLine: 1, EndColumn: 13},
				},
			},
			// Upstream invalid branch without parserOptions.tsconfigRootDir.
			{
				Code:     `import FooBar from "./typescript-export-as-default-namespace"`,
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: `No default export found in imported module "./typescript-export-as-default-namespace".`, Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
				},
			},
			// Upstream invalid branch with tsconfigRootDir pointing at a config without compiler options.
			{
				Code:     `import Foo from "./typescript-export-as-default-namespace"`,
				TSConfig: "tsconfig.no-interop.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDefault", Message: `No default export found in imported module "./typescript-export-as-default-namespace".`, Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
		},
	)
}
