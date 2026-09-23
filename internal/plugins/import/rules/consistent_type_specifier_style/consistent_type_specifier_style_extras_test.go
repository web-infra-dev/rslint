package consistent_type_specifier_style_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/consistent_type_specifier_style"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTypeSpecifierStyleEditDemand(t *testing.T) {
	r := &consistent_type_specifier_style.ConsistentTypeSpecifierStyleRule
	for _, tc := range []struct {
		name, code, option string
		count              int
	}{
		{"inline", `import type {Foo, Bar} from 'module';`, "prefer-inline", 1},
		{"all type", `import {type Foo, type Bar} from 'module';`, "prefer-top-level", 1},
		{"mixed", `import {Value, type Foo, type Bar} from 'module';`, "prefer-top-level", 2},
		{"default", `import Value, {type Foo, type Bar} from 'module';`, "prefer-top-level", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/edit-demand.ts", Path: "/edit-demand.ts"}, tc.code, core.ScriptKindTS)
			options := rule_tester.ResolveTestCaseOptions(t, r, []any{tc.option})
			var all []rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
				var diagnostics []rule.RuleDiagnostic
				ctx := (rule.RuleContext{SourceFile: file}).WithDiagnosticConsumer(r.Name, rule.SeverityError, rule.DiagnosticConsumer{
					Demand: demand,
					Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
				})
				r.Run(ctx, options)[ast.KindImportDeclaration](file.Statements.Nodes[0])
				if len(diagnostics) != tc.count {
					t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(diagnostics), tc.count)
				}
				if demand == rule.EditDemandAll {
					all = diagnostics
				}
				for i, got := range diagnostics {
					if got.Suggestions != nil {
						t.Fatalf("demand %d: unexpected suggestions", demand)
					}
					if demand == rule.EditDemandAll || demand == rule.EditDemandAutofix {
						if got.FixesPtr == nil || len(*got.FixesPtr) == 0 || !reflect.DeepEqual(got.FixesPtr, all[i].FixesPtr) {
							t.Fatalf("demand %d: autofixes missing or differ from all edits", demand)
						}
					} else if got.FixesPtr != nil {
						t.Fatalf("demand %d: unexpected autofixes", demand)
					}
					want := all[i]
					got.FixesPtr, want.FixesPtr = nil, nil
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("demand %d changed diagnostic identity", demand)
					}
				}
			}
		})
	}
}

// Expected messages, ranges and fixes checked against eslint-plugin-import v2.32.0.

func TestConsistentTypeSpecifierStyleExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_type_specifier_style.ConsistentTypeSpecifierStyleRule,
		[]rule_tester.ValidTestCase{

			// Extras
			{
				Code: `import type { Foo } from 'Foo';`,
			},
			{
				Code:    `import type * as Foo from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code: `import Foo, {} from 'Foo';`,
			},
			{
				Code:    `import type Foo, {} from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import Foo, * as Bar from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code: `import { type, type as alias, type as typeValue } from 'Foo';`,
			},
			{
				Code: `import { typeofValue, typeValue as type } from 'Foo';`,
			},
			{
				Code:    `export type { Foo } from 'Foo'; export { type Bar } from 'Bar'; type T = import('Foo').Foo; import Foo = require('Foo');`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import { type Foo as Local } from 'Foo'; const x = <Local />;`,
				Options: []any{"prefer-inline"},
				Tsx:     true,
			},
		},
		[]rule_tester.InvalidTestCase{

			// Extras
			{
				Code:   `import {type Foo} from 'Foo';`,
				Output: []string{`import type {Foo} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    `import type {Foo as Bar} from 'Foo';`,
				Options: []any{"prefer-inline"},
				Output:  []string{`import  {type Foo as Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    `/* lead */ import /* type */ type /* kind */ { /* first */ Foo, /* second */ Bar as Baz } from 'Foo';`,
				Options: []any{"prefer-inline"},
				Output:  []string{`/* lead */ import /* type */  /* kind */ { /* first */ type Foo, /* second */ type Bar as Baz } from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 1, Column: 12, EndLine: 1, EndColumn: 102},
				},
			},
			{
				Code: `import type
{
  Foo,
  Bar as Baz,
} from 'Foo'`,
				Options: []any{"prefer-inline"},
				Output:  []string{"import \n{\n  type Foo,\n  type Bar as Baz,\n} from 'Foo'"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 1, Column: 1, EndLine: 5, EndColumn: 13},
				},
			},
			{
				Code:    `import type{Foo}from'Foo'`,
				Options: []any{"prefer-inline"},
				Output:  []string{`import {type Foo}from'Foo'`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    `import type { 'Foo' as Foo } from 'Foo';`,
				Options: []any{"prefer-inline"},
				Output:  []string{`import  { type 'Foo' as Foo } from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code:   `import {type Foo, type Bar as Baz,} from "F\u006fo"`,
				Output: []string{`import type {Foo, Bar as Baz} from "F\u006fo";`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code: `import {type Foo, Value, type Bar as Baz, Other, type Last,} from 'Foo';`,
				Output: []string{`import { Value,  Other } from 'Foo';
import type {Foo, Bar as Baz, Last} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17},
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 26, EndLine: 1, EndColumn: 41},
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 50, EndLine: 1, EndColumn: 59},
				},
			},
			{
				Code: `import Foo, {type Bar, type Baz as Local,} from 'Foo';`,
				Output: []string{`import Foo from 'Foo';
import type {Bar, Baz as Local} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22},
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 24, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code: `import Foo /*before comma*/, /*brace*/ {type Bar, /*tail*/} /*after*/ from 'Foo';`,
				Output: []string{`import Foo /*before comma*/ /*after*/ from 'Foo';
import type {Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 41, EndLine: 1, EndColumn: 49},
				},
			},
			{
				Code: `import {Value /*before comma*/, /*keep*/ type Foo /*after type*/, Other, type Bar} from 'Foo';`,
				Output: []string{`import {Value /*before comma*/, /*keep*/  /*after type*/ Other } from 'Foo';
import type {Foo, Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 42, EndLine: 1, EndColumn: 50},
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 74, EndLine: 1, EndColumn: 82},
				},
			},
			{
				Code: `import {Value, type Foo, Last,} from 'Foo';`,
				Output: []string{`import {Value,  Last} from 'Foo';
import type {Foo} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `import {type Foo, Value} from 'Foo'`,
				Output: []string{`import { Value} from 'Foo'
import type {Foo} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 9, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:   `import {type Foo, type Bar /*removed*/} from 'Foo';`,
				Output: []string{`import type {Foo, Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code:   `import {type default as Foo, type Bar as Bar} from 'Foo';`,
				Output: []string{`import type {default as Foo, Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 58},
				},
			},
			{
				Code:   `import {type \u0046oo as Bar} from 'Foo';`,
				Output: []string{`import type {Foo as Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code: `/*😀*/ import {值, type 类型 as 别名} from '模块';`,
				Output: []string{`/*😀*/ import {值 } from '模块';
import type {类型 as 别名} from '模块';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 19, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code: `import {
  值,
  type 类型 as 别名,
  type Other
} from '模块';`,
				Output: []string{"import {\n  值\n  \n  \n} from '模块';\nimport type {类型 as 别名, Other} from '模块';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 3, Column: 3, EndLine: 3, EndColumn: 16},
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 4, Column: 3, EndLine: 4, EndColumn: 13},
				},
			},
			{
				Code:   `import {type Foo} from 'Foo' with { type: 'json' };`,
				Output: []string{`import type {Foo} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code: `import Foo, {type Bar} from 'Foo' with { type: 'json' };`,
				Output: []string{`import Foo from 'Foo' with { type: 'json' };
import type {Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code: `import {Value, type Foo} from 'Foo'; // trailing
const x = 1;`,
				Output: []string{`import {Value } from 'Foo';
import type {Foo} from 'Foo'; // trailing
const x = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `import { type 'Foo' as Foo, type 'undefined' as undefined } from 'Foo';`,
				// Preserve the quoted names; upstream v2.32.0 substitutes undefined.
				Output: []string{`import type {'Foo' as Foo, 'undefined' as undefined} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 72},
				},
			},
			{
				Code:    `import type {Foo} from 'Foo'; const x = <Foo />;`,
				Options: []any{"prefer-inline"},
				Tsx:     true,
				Output:  []string{`import  {type Foo} from 'Foo'; const x = <Foo />;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: `import {type Foo} from 'Foo';
import {Value, type Bar} from 'Bar';`,
				Output: []string{`import type {Foo} from 'Foo';
import {Value } from 'Bar';
import type {Bar} from 'Bar';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 2, Column: 16, EndLine: 2, EndColumn: 24},
				},
			},
		},
	)
}

// Cover keyword identifiers, comments, non-BMP names, line endings and imports
// in ambient modules. Quoted names intentionally differ from upstream's unsafe fix.
func TestConsistentTypeSpecifierStyleASTEdges(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_type_specifier_style.ConsistentTypeSpecifierStyleRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    "import type from 'foo';",
				Options: []any{"prefer-inline"},
			},
			{
				Code:    "/** @import { Foo } from 'foo' */\nconst value = 1;",
				Options: []any{"prefer-inline"},
			},
			{
				Code: "import { type as type, type as from } from 'foo';",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "import {type type, type type as Alias, Value} from 'foo';",
				Output: []string{"import {  Value} from 'foo';\nimport type {type, type as Alias} from 'foo';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 20, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:    "\uFEFF// leading\r\nimport type { Foo, Bar as Baz } from 'foo';",
				Options: []any{"prefer-inline"},
				Output:  []string{"\uFEFF// leading\r\nimport  { type Foo, type Bar as Baz } from 'foo';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 2, Column: 1, EndLine: 2, EndColumn: 44},
				},
			},
			{
				Code:   "#!/usr/bin/env node\nimport {type Foo} from 'foo';",
				Output: []string{"#!/usr/bin/env node\nimport type {Foo} from 'foo';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 2, Column: 1, EndLine: 2, EndColumn: 30},
				},
			},
			{
				Code:   "import { type /*between*/ Foo /*alias*/ as /*name*/ Bar, Value /*last*/, } from 'foo';",
				Output: []string{"import {  Value /*last*/ } from 'foo';\nimport type {Foo as Bar} from 'foo';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 10, EndLine: 1, EndColumn: 56},
				},
			},
			{
				Code:    "/* lead\u2028*/import type /*kind*/{Foo as Bar}from'foo';",
				Options: []any{"prefer-inline"},
				Output:  []string{"/* lead\u2028*/import  /*kind*/{type Foo as Bar}from'foo';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 2, Column: 3, EndLine: 2, EndColumn: 45},
				},
			},
			{
				Code:   "import {type \\u{10400} as 𐐁, type café as cafe, 值} from '模块';",
				Output: []string{"import {  值} from '模块';\nimport type {𐐀 as 𐐁, café as cafe} from '模块';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 9, EndLine: 1, EndColumn: 29},
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 31, EndLine: 1, EndColumn: 48},
				},
			},
			{
				Code:   "declare module 'foo' {\n  import {type Foo, Value} from 'bar';\n  export {Foo, Value};\n}",
				Output: []string{"declare module 'foo' {\n  import { Value} from 'bar';\nimport type {Foo} from 'bar';\n  export {Foo, Value};\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 2, Column: 11, EndLine: 2, EndColumn: 19},
				},
			},
			{
				Code:    "declare module 'foo' {\n  import type {Foo} from 'bar';\n  export {Foo};\n}",
				Options: []any{"prefer-inline"},
				Output:  []string{"declare module 'foo' {\n  import  {type Foo} from 'bar';\n  export {Foo};\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 2, Column: 3, EndLine: 2, EndColumn: 32},
				},
			},
			{
				// Keep the original quoted names instead of upstream's undefined imports.
				Code:   "import {type '' as Empty, type 'a-b' as A} from 'foo';",
				Output: []string{"import type {'' as Empty, 'a-b' as A} from 'foo';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 55},
				},
			},
			{
				// Keep the original quoted names instead of upstream's undefined imports.
				Code:   "import Default, {Value, type 'f\\u006fo' as Foo, type 'a-b' as A} from 'foo';",
				Output: []string{"import Default, {Value  } from 'foo';\nimport type {'f\\u006fo' as Foo, 'a-b' as A} from 'foo';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 25, EndLine: 1, EndColumn: 47},
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 49, EndLine: 1, EndColumn: 64},
				},
			},
			{
				// Keep the original quoted names instead of upstream's undefined imports.
				Code:   "import Default, {type 'a-b' as A} from 'foo';",
				Output: []string{"import Default from 'foo';\nimport type {'a-b' as A} from 'foo';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 18, EndLine: 1, EndColumn: 33},
				},
			},
			{
				// Keep the original quoted names instead of upstream's undefined imports.
				Code:   "import { type 'a-b' as A } from 'module';",
				Output: []string{"import type {'a-b' as A} from 'module';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
				},
			},
		},
	)
}
