package camelcase

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func parityError(column, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "notCamelCase",
		Line:      1,
		Column:    column,
		EndLine:   1,
		EndColumn: endColumn,
	}
}

func TestCamelcaseTypeScriptScopeManagerParity(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&CamelcaseRule,
		[]rule_tester.ValidTestCase{
			// TypeScript-only variables resolve locally even when their declaration
			// form has no camelcase declaration listener.
			{Code: `namespace snake_case {} snake_case`},
			{Code: `interface snake_case {}; export default (snake_case)`},
			{Code: `interface snake_case {}; export = (snake_case)`},
			// A signature is TSDeclareFunction upstream and never triggers the
			// runtime function listener.
			{Code: `declare function ok<snake_case>(snake_case: unknown): void;`},
			// scope-manager still merges invalid TypeScript declarations by lexical
			// scope and name. The enum makes each reference locally resolvable in
			// both value and type space even when the TypeScript binder rejects the
			// declaration pair.
			{Code: `interface snake_case{}; enum snake_case{X}; snake_case; type T=snake_case; type Q=typeof snake_case`},
			{Code: `function f(){type snake_case={};enum snake_case{};snake_case}`},
			// Both ignore options inspect the merged variable's anchor rather than
			// the syntax node that happened to trigger the listener.
			{
				Code:    `import { snake_case } from "m"; const { snake_case } = source;`,
				Options: map[string]any{"ignoreDestructuring": true, "ignoreImports": true},
			},
			{
				Code:    `const { snake_case } = source; import { snake_case } from "m";`,
				Options: map[string]any{"ignoreDestructuring": true, "ignoreImports": true},
			},
			// Configured globals merge with authored definitions only in an
			// actual global Program scope.
			{
				Code:            `interface snake_case {}; snake_case`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Globals:         map[string]any{"snake_case": "readonly"},
			},
			{
				Code:            `interface snake_case {}; snake_case`,
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Globals:         map[string]any{"snake_case": "readonly"},
				Options:         map[string]any{"ignoreGlobals": true},
			},
			// JavaScript casing is applied to the first UTF-16 code unit. These
			// BMP characters uppercase to a different string and are intrinsic.
			{
				Code: `const x = <><é_snake /><ß_snake /><ﬁ_snake /><ı_snake /><ǅ_snake /></>`,
				Tsx:  true,
			},
			// Namespace and enum declarations satisfy the value references which
			// TSESTree creates for both pieces of a namespaced TSX tag.
			{
				Code: `namespace snake_case {}; namespace other_name {}; const x = <snake_case:other_name />;`,
				Tsx:  true,
			},
			{
				Code: `enum snake_case {}; enum other_name {}; const x = <snake_case:other_name />;`,
				Tsx:  true,
			},
			// Espree does not create references for namespaced JSX tag pieces.
			{
				Code:     `const x = <snake_case:other_name />;`,
				FileName: "camelcase-namespaced.jsx",
				TSConfig: "tsconfig.allowJs.json",
			},
			// Namespaced attributes are metadata under both parsers.
			{Code: `const x = <Component snake_case:other_name="value" />;`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `interface snake_case {}; class snake_case {}`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(11, 21), parityError(32, 42)},
			},
			{
				Code:   `class snake_case {}; interface snake_case {}`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(7, 17)},
			},
			{
				Code:   `type snake_case = {}; const snake_case = 1; snake_case; type T = snake_case`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(6, 16), parityError(45, 55), parityError(66, 76)},
			},
			{
				Code:   `const snake_case = 1; type snake_case = {}; snake_case; type T = snake_case`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(7, 17), parityError(45, 55), parityError(66, 76)},
			},
			{
				Code:   `namespace snake_case {}; function snake_case() {}`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(11, 21)},
			},
			{
				Code:   `enum snake_case {}; class snake_case {}; snake_case`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(6, 16), parityError(27, 37), parityError(42, 52)},
			},
			{
				Code:   `namespace snake_case {}; const snake_case=1; snake_case; type T=snake_case`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(11, 21), parityError(46, 56), parityError(65, 75)},
			},
			{
				Code:   `function f(snake_case){namespace snake_case{};type T=snake_case}`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(12, 22), parityError(54, 64)},
			},
			// Class variables are type- and value-capable in scope-manager. Their
			// qualified roots stay attached to the class even when TypeScript asks
			// its binder for namespace meaning.
			{
				Code:   `class snake_case {}; type T=snake_case.Member; interface X extends snake_case.Member {}; class Y implements snake_case.Member {}`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(7, 17), parityError(29, 39), parityError(68, 78), parityError(109, 119)},
			},
			{
				Code:   `const snake_case={}; import Alias=snake_case.Member;`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(7, 17), parityError(35, 45)},
			},
			{
				Code:   `export default function snake_case(){return snake_case}`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(25, 35), parityError(45, 55)},
			},
			{
				Code:   `export default class snake_case{static x=snake_case}`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(22, 32), parityError(42, 52)},
			},
			// Runtime function scopes define parameters before type parameters,
			// even though the type parameter appears first in source text.
			{
				Code:   `function ok<snake_case>(snake_case: unknown) { snake_case }`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(25, 44), parityError(48, 58)},
			},
			{
				Code:   `const ok = <snake_case,>(snake_case: unknown) => snake_case`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{parityError(26, 45), parityError(50, 60)},
			},
			{
				Code:   `class C { ok<snake_case>(snake_case: unknown) { return snake_case } }`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(26, 45), parityError(56, 66)},
			},
			{
				Code:   `function ok<snake_case>() { var snake_case; snake_case }`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(13, 23), parityError(45, 55)},
			},
			// ImportDeclaration is the one camelcase listener which reports init
			// references. The typed initializer retains TSESTree's binding range.
			{
				Code:   `import { snake_case } from 'm'; const snake_case: number = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(10, 20), parityError(39, 57)},
			},
			{
				Code:   `import { snake_case } from 'm'; let snake_case;`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(10, 20)},
			},
			{
				Code:   `import { snake_case } from 'm'; let snake_case = 1, snake_case = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(10, 20), parityError(37, 47), parityError(53, 63)},
			},
			{
				Code:   `import { snake_case } from 'm'; for (var snake_case of source) {}`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(10, 20), parityError(42, 52)},
			},
			{
				Code:    `import { snake_case } from 'm'; const { source_name: snake_case = value } = source;`,
				Options: map[string]any{"ignoreDestructuring": true, "ignoreImports": true},
				Errors:  []rule_tester.InvalidTestCaseError{parityError(54, 64)},
			},
			{
				Code:    `import { snake_case } from 'm'; const snake_case = 1;`,
				Options: map[string]any{"ignoreImports": true},
				Errors:  []rule_tester.InvalidTestCaseError{parityError(10, 20), parityError(39, 49)},
			},
			{
				Code:    `import { snake_case } from 'm'; class snake_case {}`,
				Options: map[string]any{"ignoreDestructuring": true, "ignoreImports": true},
				Errors:  []rule_tester.InvalidTestCaseError{parityError(39, 49)},
			},
			// Import-equals contributes an anchor to the merged variable but has
			// no ImportDeclaration listener, so the const init write stays filtered.
			{
				Code:   `import snake_case = require('m'); const snake_case = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{parityError(8, 18)},
			},
			{
				Code:    `type snake_case = {}; const { snake_case } = source`,
				Options: map[string]any{"ignoreDestructuring": true},
				Errors:  []rule_tester.InvalidTestCaseError{parityError(6, 16)},
			},
			{
				Code:    `interface snake_case {}; import { snake_case } from "m"`,
				Options: map[string]any{"ignoreImports": true},
				Errors:  []rule_tester.InvalidTestCaseError{parityError(11, 21)},
			},
			{
				Code:            `interface snake_case {}; snake_case`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors:          []rule_tester.InvalidTestCaseError{parityError(26, 36)},
			},
			{
				Code:            `interface snake_case {}; snake_case`,
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Globals:         map[string]any{"snake_case": "readonly"},
				Errors:          []rule_tester.InvalidTestCaseError{parityError(26, 36)},
			},
			{
				Code:   `const é_snake = () => null; const x = <é_snake />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{parityError(7, 14)},
			},
			{
				Code:   `const x = <Snake_case-tag />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{parityError(12, 26)},
			},
			// An astral first letter is indexed as an unchanged high surrogate.
			{
				Code:   `const x = <𐐨_snake-tag />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{parityError(12, 24)},
			},
			// Espree visits only the opening JSX name, while typescript-eslint
			// visits both opening and closing names.
			{
				Code:     `const x=<Snake_case></Snake_case>`,
				FileName: "camelcase-closing.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{parityError(10, 20)},
			},
			{
				Code:   `const x=<Snake_case></Snake_case>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{parityError(10, 20), parityError(23, 33)},
			},
			{
				Code:     `const x=<Snake_case.Part></Snake_case.Part>`,
				FileName: "camelcase-member-closing.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{parityError(10, 20)},
			},
			// Interfaces do not satisfy the value references of a namespaced tag.
			{
				Code:   `interface snake_case {}; interface other_name {}; const x = <snake_case:other_name />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{parityError(62, 72), parityError(73, 83)},
			},
			{
				Code:   `const snake_case = 1, other_name = 2; const x = <snake_case:other_name />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{parityError(7, 17), parityError(23, 33), parityError(50, 60), parityError(61, 71)},
			},
			{
				Code:   `namespace snake_case {}; const x = <snake_case:part_bad></snake_case:part_bad>;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{parityError(48, 56), parityError(70, 78)},
			},
			{
				Code:   `const snake_case = 1; function f(){ namespace snake_case {}; const x=<snake_case:part_bad/> }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{parityError(7, 17), parityError(82, 90)},
			},
			{
				Code:   `const snake_case = 1; function f(){ interface snake_case {}; const x=<snake_case:part_bad/> }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{parityError(7, 17), parityError(71, 81), parityError(82, 90)},
			},
		},
	)
}
