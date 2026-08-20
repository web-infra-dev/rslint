package id_match

// TestIdMatchUpstream migrates the full valid/invalid suite from upstream
// tests/lib/rules/id-match.js 1:1. Position assertions cover line/column for
// every invalid case. rslint-specific lock-in cases live in the
// id_match_extras_test.go file.

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIdMatchUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&IdMatchRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    `__foo = "Matthieu"`,
				Options: []any{`^[a-z]+$`, map[string]any{"onlyDeclarations": true}},
			},
			{
				Code:    `firstname = "Matthieu"`,
				Options: []any{`^[a-z]+$`},
			},
			{
				Code:    `first_name = "Matthieu"`,
				Options: []any{`[a-z]+`},
			},
			{
				Code:    `firstname = "Matthieu"`,
				Options: []any{`^f`},
			},
			{
				Code:    `last_Name = "Larcher"`,
				Options: []any{`^[a-z]+(_[A-Z][a-z]+)*$`},
			},
			{
				Code:    `param = "none"`,
				Options: []any{`^[a-z]+(_[A-Z][a-z])*$`},
			},
			{
				Code:    `function noUnder(){}`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `no_under()`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `foo.no_under2()`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `var foo = bar.no_under3;`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `var foo = bar.no_under4.something;`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `foo.no_under5.qux = bar.no_under6.something;`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `if (bar.no_under7) {}`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `var obj = { key: foo.no_under8 };`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `var arr = [foo.no_under9];`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `[foo.no_under10]`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `var arr = [foo.no_under11.qux];`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `[foo.no_under12.nesting]`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `if (foo.no_under13 === boom.no_under14) { [foo.no_under15] }`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `var myArray = new Array(); var myDate = new Date();`,
				Options: []any{`^[a-z$]+([A-Z][a-z]+)*$`},
			},
			{
				Code:    `var x = obj._foo;`,
				Options: []any{`^[^_]+$`},
			},
			{
				Code:    `var obj = {key: no_under}`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true, "onlyDeclarations": true}},
			},
			{
				Code:            `var {key_no_under: key} = {}`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
			},
			{
				Code:            `var { category_id } = query;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true, "ignoreDestructuring": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
			},
			{
				Code:            `var { category_id: category_id } = query;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true, "ignoreDestructuring": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
			},
			{
				Code:            `var { category_id = 1 } = query;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true, "ignoreDestructuring": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
			},
			{
				Code:    `var o = {key: 1}`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
			},
			{
				Code:    `var o = {no_under16: 1}`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": false}},
			},
			{
				Code:    `obj.no_under17 = 2;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": false}},
			},
			{
				Code: `var obj = {
 no_under18: 1 
};
 obj.no_under19 = 2;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": false}},
			},
			{
				Code:    `obj.no_under20 = function(){};`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": false}},
			},
			{
				Code:    `var x = obj._foo2;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": false}},
			},
			// ---- Should not report for global references - https://github.com/eslint/eslint/issues/15395 ----
			{
				Code: `
            const foo = Object.keys(bar);
            const a = Array.from(b);
            const bar = () => Array;
            `,
				Options:         []any{`^\$?[a-z]+([A-Z0-9][a-z0-9]+)*$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			{
				Code: `
            const foo = {
                foo_one: 1,
                bar_one: 2,
                fooBar: 3
            };
            `,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": false}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			{
				Code: `
            const foo = {
                foo_one: 1,
                bar_one: 2,
                fooBar: 3
            };
            `,
				Options:         []any{`^[^_]+$`, map[string]any{"onlyDeclarations": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			{
				Code: `
            const foo = {
                foo_one: 1,
                bar_one: 2,
                fooBar: 3
            };
            `,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": false, "onlyDeclarations": false}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			{
				Code: `
            const foo = {
                [a]: 1,
            };
            `,
				Options:         []any{`^[^a]`, map[string]any{"properties": true, "onlyDeclarations": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			// ---- Class Methods ----
			{
				Code:            `class x { foo() {} }`,
				Options:         []any{`^[^_]+$`},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			{
				Code:            `class x { #foo() {} }`,
				Options:         []any{`^[^_]+$`},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			// ---- Class Fields ----
			{
				Code:            `class x { _foo = 1; }`,
				Options:         []any{`^[^_]+$`},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			{
				Code:            `class x { _foo = 1; }`,
				Options:         []any{`^[^_]+$`, map[string]any{"classFields": false}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			{
				Code:            `class x { #_foo = 1; }`,
				Options:         []any{`^[^_]+$`, map[string]any{"classFields": false}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			{
				Code:            `class x { #_foo = 1; }`,
				Options:         []any{`^[^_]+$`},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
			},
			// ---- Meta-properties ----
			{
				Code:    `import.meta`,
				Options: []any{`^$`},
			},
			{
				Code:    `function foo() { new.target; }`,
				Options: []any{`^foo$`},
			},
			// ---- Import attribute keys ----
			{
				Code:            `import foo from 'foo.json' with { type: 'json' }`,
				Options:         []any{`^foo`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
			},
			{
				Code:            `export * from 'foo.json' with { type: 'json' }`,
				Options:         []any{`^foo`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
			},
			{
				Code:            `export { default } from 'foo.json' with { type: 'json' }`,
				Options:         []any{`^def`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
			},
			{
				Code:            `import('foo.json', { with: { type: 'json' } })`,
				Options:         []any{`^foo`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
			},
			{
				Code:            `import('foo.json', { 'with': { type: 'json' } })`,
				Options:         []any{`^foo`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
			},
			{
				Code:            `import('foo.json', { with: { type } })`,
				Options:         []any{`^foo`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `var __foo = "Matthieu"`,
				Options: []any{`^[a-z]+$`, map[string]any{"onlyDeclarations": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier '__foo' does not match the pattern '^[a-z]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 10,
					},
				},
			},
			{
				Code:    `first_name = "Matthieu"`,
				Options: []any{`^[a-z]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'first_name' does not match the pattern '^[a-z]+$'.`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},
			{
				Code:    `first_name = "Matthieu"`,
				Options: []any{`^z`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'first_name' does not match the pattern '^z'.`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},
			{
				Code:    `Last_Name = "Larcher"`,
				Options: []any{`^[a-z]+(_[A-Z][a-z])*$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Last_Name' does not match the pattern '^[a-z]+(_[A-Z][a-z])*$'.`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 10,
					},
				},
			},
			{
				Code:    `var obj = {key: no_under}`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    17,
						EndLine:   1,
						EndColumn: 25,
					},
				},
			},
			{
				Code:    `function no_under21(){}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under21' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 20,
					},
				},
			},
			{
				Code:    `obj.no_under22 = function(){};`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under22' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			{
				Code:    `no_under23.foo = function(){};`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under23' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},
			{
				Code:    `[no_under24.baz]`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under24' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    2,
						EndLine:   1,
						EndColumn: 12,
					},
				},
			},
			{
				Code:    `if (foo.bar_baz === boom.bam_pow) { [no_under25.baz] }`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under25' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    38,
						EndLine:   1,
						EndColumn: 48,
					},
				},
			},
			{
				Code:    `foo.no_under26 = boom.bam_pow`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under26' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			{
				Code:    `var foo = { no_under27: boom.bam_pow }`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under27' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 23,
					},
				},
			},
			{
				Code:    `foo.qux.no_under28 = { bar: boom.bam_pow }`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under28' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    9,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code:    `var o = {no_under29: 1}`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under29' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 20,
					},
				},
			},
			{
				Code:    `obj.no_under30 = 2;`,
				Options: []any{`^[^_]+$`, map[string]any{"properties": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_under30' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    5,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			{
				Code:            `var { category_id: category_alias } = query;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'category_alias' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    20,
						EndLine:   1,
						EndColumn: 34,
					},
				},
			},
			{
				Code:            `var { category_id: category_alias } = query;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true, "ignoreDestructuring": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'category_alias' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    20,
						EndLine:   1,
						EndColumn: 34,
					},
				},
			},
			{
				Code:            `var { category_id: categoryId, ...other_props } = query;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true, "ignoreDestructuring": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2018},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'other_props' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    35,
						EndLine:   1,
						EndColumn: 46,
					},
				},
			},
			{
				Code:            `var { category_id } = query;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'category_id' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 18,
					},
				},
			},
			{
				Code:            `var { category_id = 1 } = query;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'category_id' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 18,
					},
				},
			},
			{
				Code:            `import no_camelcased from "external-module";`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    8,
						EndLine:   1,
						EndColumn: 21,
					},
				},
			},
			{
				Code:            `import * as no_camelcased from "external-module";`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 26,
					},
				},
			},
			{
				Code:            `export * as no_camelcased from "external-module";`,
				Options:         []any{`^[^_]+$`},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 26,
					},
				},
			},
			{
				Code:            `import { no_camelcased } from "external-module";`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 23,
					},
				},
			},
			{
				Code:            `import { no_camelcased as no_camel_cased } from "external module";`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camel_cased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    27,
						EndLine:   1,
						EndColumn: 41,
					},
				},
			},
			{
				Code:            `import { camelCased as no_camel_cased } from "external module";`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camel_cased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    24,
						EndLine:   1,
						EndColumn: 38,
					},
				},
			},
			{
				Code:            `import { camelCased, no_camelcased } from "external-module";`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    22,
						EndLine:   1,
						EndColumn: 35,
					},
				},
			},
			{
				Code:            `import { no_camelcased as camelCased, another_no_camelcased } from "external-module";`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'another_no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    39,
						EndLine:   1,
						EndColumn: 60,
					},
				},
			},
			{
				Code:            `import camelCased, { no_camelcased } from "external-module";`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    22,
						EndLine:   1,
						EndColumn: 35,
					},
				},
			},
			{
				Code:            `import no_camelcased, { another_no_camelcased as camelCased } from "external-module";`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    8,
						EndLine:   1,
						EndColumn: 21,
					},
				},
			},
			{
				Code:            `function foo({ no_camelcased }) {};`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    16,
						EndLine:   1,
						EndColumn: 29,
					},
				},
			},
			{
				Code:            `function foo({ no_camelcased = 'default value' }) {};`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    16,
						EndLine:   1,
						EndColumn: 29,
					},
				},
			},
			{
				Code:            `const no_camelcased = 0; function foo({ camelcased_value = no_camelcased }) {}`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    7,
						EndLine:   1,
						EndColumn: 20,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'camelcased_value' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    41,
						EndLine:   1,
						EndColumn: 57,
					},
				},
			},
			{
				Code:            `const { bar: no_camelcased } = foo;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 27,
					},
				},
			},
			{
				Code:            `function foo({ value_1: my_default }) {}`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'my_default' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    25,
						EndLine:   1,
						EndColumn: 35,
					},
				},
			},
			{
				Code:            `function foo({ isCamelcased: no_camelcased }) {};`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    30,
						EndLine:   1,
						EndColumn: 43,
					},
				},
			},
			{
				Code:            `var { foo: bar_baz = 1 } = quz;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'bar_baz' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 19,
					},
				},
			},
			{
				Code:            `const { no_camelcased = false } = bar;`,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'no_camelcased' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    9,
						EndLine:   1,
						EndColumn: 22,
					},
				},
			},
			// ---- https://github.com/eslint/eslint/issues/15395 ----
			{
				Code: `
            const foo_variable = 1;
            class MyClass {
            }
            let a = new MyClass();
            let b = {id: 1};
            let c = Object.keys(b);
            let d = Array.from(b);
            let e = (Object) => Object.keys(obj, prop); // not global Object
            let f = (Array) => Array.from(obj, prop); // not global Array
            foo.Array = 5; // not global Array
            `,
				Options:         []any{`^\$?[a-z]+([A-Z0-9][a-z0-9]+)*$`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'foo_variable' does not match the pattern '^\$?[a-z]+([A-Z0-9][a-z0-9]+)*$'.`,
						Line:      2,
						Column:    19,
						EndLine:   2,
						EndColumn: 31,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'MyClass' does not match the pattern '^\$?[a-z]+([A-Z0-9][a-z0-9]+)*$'.`,
						Line:      3,
						Column:    19,
						EndLine:   3,
						EndColumn: 26,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Object' does not match the pattern '^\$?[a-z]+([A-Z0-9][a-z0-9]+)*$'.`,
						Line:      9,
						Column:    22,
						EndLine:   9,
						EndColumn: 28,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Object' does not match the pattern '^\$?[a-z]+([A-Z0-9][a-z0-9]+)*$'.`,
						Line:      9,
						Column:    33,
						EndLine:   9,
						EndColumn: 39,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Array' does not match the pattern '^\$?[a-z]+([A-Z0-9][a-z0-9]+)*$'.`,
						Line:      10,
						Column:    22,
						EndLine:   10,
						EndColumn: 27,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Array' does not match the pattern '^\$?[a-z]+([A-Z0-9][a-z0-9]+)*$'.`,
						Line:      10,
						Column:    32,
						EndLine:   10,
						EndColumn: 37,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Array' does not match the pattern '^\$?[a-z]+([A-Z0-9][a-z0-9]+)*$'.`,
						Line:      11,
						Column:    17,
						EndLine:   11,
						EndColumn: 22,
					},
				},
			},
			// ---- Class Methods ----
			{
				Code:            `class x { _foo() {} }`,
				Options:         []any{`^[^_]+$`},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier '_foo' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			{
				Code:            `class x { #_foo() {} }`,
				Options:         []any{`^[^_]+$`},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatchPrivate",
						Message:   `Identifier '#_foo' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 16,
					},
				},
			},
			// ---- Class Fields ----
			{
				Code:            `class x { _foo = 1; }`,
				Options:         []any{`^[^_]+$`, map[string]any{"classFields": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier '_foo' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			{
				Code:            `class x { #_foo = 1; }`,
				Options:         []any{`^[^_]+$`, map[string]any{"classFields": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatchPrivate",
						Message:   `Identifier '#_foo' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 16,
					},
				},
			},
			// ---- https://github.com/eslint/eslint/issues/15123 ----
			{
				Code: `
            const foo = {
                foo_one: 1,
                bar_one: 2,
                fooBar: 3
            };
            `,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true, "onlyDeclarations": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'foo_one' does not match the pattern '^[^_]+$'.`,
						Line:      3,
						Column:    17,
						EndLine:   3,
						EndColumn: 24,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'bar_one' does not match the pattern '^[^_]+$'.`,
						Line:      4,
						Column:    17,
						EndLine:   4,
						EndColumn: 24,
					},
				},
			},
			{
				Code: `
            const foo = {
                foo_one: 1,
                bar_one: 2,
                fooBar: 3
            };
            `,
				Options:         []any{`^[^_]+$`, map[string]any{"properties": true, "onlyDeclarations": false}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'foo_one' does not match the pattern '^[^_]+$'.`,
						Line:      3,
						Column:    17,
						EndLine:   3,
						EndColumn: 24,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'bar_one' does not match the pattern '^[^_]+$'.`,
						Line:      4,
						Column:    17,
						EndLine:   4,
						EndColumn: 24,
					},
				},
			},
			{
				Code: `
            const foo = {
                [a]: 1,
            };
            `,
				Options:         []any{`^[^a]`, map[string]any{"properties": true, "onlyDeclarations": false}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a' does not match the pattern '^[^a]'.`,
						Line:      3,
						Column:    18,
						EndLine:   3,
						EndColumn: 19,
					},
				},
			},
			// ---- https://github.com/eslint/eslint/issues/15443 ----
			{
				Code: `
            const foo = {
                [a]: 1,
            };
            `,
				Options:         []any{`^[^a]`, map[string]any{"properties": false, "onlyDeclarations": false}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a' does not match the pattern '^[^a]'.`,
						Line:      3,
						Column:    18,
						EndLine:   3,
						EndColumn: 19,
					},
				},
			},
			// ---- Not an import attribute key ----
			{
				Code:            `import('foo.json', { with: { [type]: 'json' } })`,
				Options:         []any{`^foo`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'type' does not match the pattern '^foo'.`,
						Line:      1,
						Column:    31,
						EndLine:   1,
						EndColumn: 35,
					},
				},
			},
			{
				Code:            `import('foo.json', { with: { type: json } })`,
				Options:         []any{`^foo`, map[string]any{"properties": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'json' does not match the pattern '^foo'.`,
						Line:      1,
						Column:    36,
						EndLine:   1,
						EndColumn: 40,
					},
				},
			},
		},
	)
}
