package no_named_default_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_named_default"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNamedDefaultExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_named_default.NoNamedDefaultRule,
		[]rule_tester.ValidTestCase{
			{Code: `import "./bar"; import {} from "./bar";`},
			{Code: `import bar, * as namespace from "./bar";`},
			{Code: `import { value as defaultValue, "other" as named } from "./bar";`},
			{Code: `import { "" as empty, "Default" as capital, "default " as trailing } from "./bar";`},
			{Code: `import { type as localType, from as localFrom, as as localAs } from "./bar";`},
			{Code: `export { default as alias } from "./bar";`},
			{Code: `import type Foo from "./bar";`},
			{Code: `import { type "default" as Foo, type Model } from "./bar";`},
			{Code: `const { default: value } = await import("./bar"); const other = require("./bar").default;`},
			// JSDoc imports are comments, not ESTree import declarations.
			{
				Code:     `/** @import { default as Foo } from "./bar" */ const value = 1;`,
				FileName: "input.js",
				TSConfig: "tsconfig.allow-js.json",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Contextual keywords and internal-looking names are ordinary local bindings.
			{
				Code: `import { default as type } from "./bar"; import { default as from } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'type'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 25},
					{MessageId: "", Message: "Use default import syntax to import 'from'.", Line: 1, Column: 62, EndLine: 1, EndColumn: 66},
				},
			},
			{
				Code: `import { default as __proto__ } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import '__proto__'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: `import original, { default as alias } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'alias'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: `import { default as first, default as second } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'first'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 26},
					{MessageId: "", Message: "Use default import syntax to import 'second'.", Line: 1, Column: 39, EndLine: 1, EndColumn: 45},
				},
			},
			// Only inline type specifiers are exempt in upstream v2.32.0.
			{
				Code: `import type { default as Foo } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'Foo'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code: `import { type default as TypeOnly, default as value } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'value'.", Line: 1, Column: 47, EndLine: 1, EndColumn: 52},
				},
			},
			// Match decoded names, but report the local identifier's original range.
			{
				Code: `import { def\u0061ult as alias } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'alias'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code: `import { "def\u0061ult" as b\u0061r } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'bar'.", Line: 1, Column: 28, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: `/* 😀 */ import { default as 名𐐀 } from "./bar";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import '名𐐀'.", Line: 1, Column: 30, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code: "import {\n  default as /* local */\n    名称,\n} from \"./bar\";",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import '名称'.", Line: 3, Column: 5, EndLine: 3, EndColumn: 7},
				},
			},
			{
				Code: "import { default as /* block */ // line\r\n\t\u00A0\uFEFFn\\u0061me } from \"./bar\";",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'name'.", Line: 2, Column: 4, EndLine: 2, EndColumn: 13},
				},
			},
			{
				Code:     `import { default as data } from "./data.json" with { type: "json" };`,
				FileName: "input.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'data'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: `import { default as UI } from "./component"; const node = <UI />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'UI'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 23},
				},
			},
			// Ambient declarations and type import attributes keep the same import semantics.
			{
				Code: `declare module "pkg" { import { default as Foo } from "./bar"; export { Foo }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'Foo'.", Line: 1, Column: 44, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code: `import type { "default" as Foo } from "./bar" with { "resolution-mode": "import" };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: "Use default import syntax to import 'Foo'.", Line: 1, Column: 28, EndLine: 1, EndColumn: 31},
				},
			},
		},
	)
}
