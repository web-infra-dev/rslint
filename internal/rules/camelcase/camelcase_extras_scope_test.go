// TestCamelcaseExtrasScope locks in typescript-eslint scope-manager behavior
// that the upstream core camelcase suite does not exercise: declaration
// merging, independent type/value references, Program globals, and JSX parser
// splits. The full upstream migration lives in camelcase_upstream_test.go;
// other rslint-specific edge shapes live in camelcase_extras_test.go.
package camelcase

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func scopeManagerError(column, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "notCamelCase",
		Line:      1,
		Column:    column,
		EndLine:   1,
		EndColumn: endColumn,
	}
}

func TestCamelcaseExtrasScope(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&CamelcaseRule,
		[]rule_tester.ValidTestCase{
			// ---- Scope resolution: a namespace value stays out of Program ----
			{Code: `namespace snake_case {} snake_case`},
			// ---- Scope resolution: parenthesized default export is a dual reference ----
			{Code: `interface snake_case {}; export default (snake_case)`},
			// ---- Scope resolution: parenthesized export-equals is a dual reference ----
			{Code: `interface snake_case {}; export = (snake_case)`},
			// ---- Invalid TS merge: interface plus enum resolves in both spaces ----
			{Code: `interface snake_case{}; enum snake_case{X}; snake_case; type T=snake_case; type Q=typeof snake_case`},
			// ---- Nested invalid TS merge: type alias plus enum resolves locally ----
			{Code: `function f(){type snake_case={};enum snake_case{};snake_case}`},
			// ---- Merge anchor: import-first order honors both ignore options ----
			{
				Code:    `import { snake_case } from "m"; const { snake_case } = source;`,
				Options: map[string]any{"ignoreDestructuring": true, "ignoreImports": true},
			},
			// ---- Merge anchor: destructuring-first order honors both ignore options ----
			{
				Code:    `const { snake_case } = source; import { snake_case } from "m";`,
				Options: map[string]any{"ignoreDestructuring": true, "ignoreImports": true},
			},
			// ---- Program global: script definition merges with a configured global ----
			{
				Code:            `interface snake_case {}; snake_case`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Globals:         map[string]any{"snake_case": "readonly"},
			},
			// ---- Program global: an inline readonly declaration merges with a script definition ----
			{
				Code:            `/* global snake_case:readonly */ interface snake_case {}; snake_case`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// ---- Program global: module definition remains separate and is ignored ----
			{
				Code:            `interface snake_case {}; snake_case`,
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Globals:         map[string]any{"snake_case": "readonly"},
				Options:         map[string]any{"ignoreGlobals": true},
			},
			// ---- JSX casing: BMP names use JavaScript first-code-unit uppercase ----
			{
				Code: `const x = <><é_snake /><ß_snake /><ﬁ_snake /><ı_snake /><ǅ_snake /></>`,
				Tsx:  true,
			},
			// ---- TSX namespace tag: namespace declarations satisfy both pieces ----
			{
				Code: `namespace snake_case {}; namespace other_name {}; const x = <snake_case:other_name />;`,
				Tsx:  true,
			},
			// ---- TSX namespace tag: enum declarations satisfy both pieces ----
			{
				Code: `enum snake_case {}; enum other_name {}; const x = <snake_case:other_name />;`,
				Tsx:  true,
			},
			// ---- JSX parser split: Espree creates no namespaced tag references ----
			{
				Code:     `const x = <snake_case:other_name />;`,
				FileName: "camelcase-namespaced.jsx",
				TSConfig: "tsconfig.allowJs.json",
			},
			// ---- JSX attributes: namespaced pieces are metadata, not references ----
			{Code: `const x = <Component snake_case:other_name="value" />;`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Merge anchor: interface-first reports both interface and class ----
			{
				Code:   `interface snake_case {}; class snake_case {}`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(11, 21), scopeManagerError(32, 42)},
			},
			// ---- Merge anchor: class-first reports only identifiers[0] ----
			{
				Code:   `class snake_case {}; interface snake_case {}`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(7, 17)},
			},
			// ---- Merge references: type-first anchor retains value and type reads ----
			{
				Code:   `type snake_case = {}; const snake_case = 1; snake_case; type T = snake_case`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(6, 16), scopeManagerError(45, 55), scopeManagerError(66, 76)},
			},
			// ---- Merge references: value-first anchor retains value and type reads ----
			{
				Code:   `const snake_case = 1; type snake_case = {}; snake_case; type T = snake_case`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(7, 17), scopeManagerError(45, 55), scopeManagerError(66, 76)},
			},
			// ---- Merge anchor: namespace precedes a runtime function ----
			{
				Code:   `namespace snake_case {}; function snake_case() {}`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(11, 21)},
			},
			// ---- Merge anchor: enum and class identifiers plus value read ----
			{
				Code:   `enum snake_case {}; class snake_case {}; snake_case`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(6, 16), scopeManagerError(27, 37), scopeManagerError(42, 52)},
			},
			// ---- Merge references: namespace plus const retains both spaces ----
			{
				Code:   `namespace snake_case {}; const snake_case=1; snake_case; type T=snake_case`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(11, 21), scopeManagerError(46, 56), scopeManagerError(65, 75)},
			},
			// ---- Nested merge: parameter plus namespace retains the type read ----
			{
				Code:   `function f(snake_case){namespace snake_case{};type T=snake_case}`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(12, 22), scopeManagerError(54, 64)},
			},
			// ---- Reference spaces: qualified type roots stay on the class variable ----
			{
				Code:   `class snake_case {}; type T=snake_case.Member; interface X extends snake_case.Member {}; class Y implements snake_case.Member {}`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(7, 17), scopeManagerError(29, 39), scopeManagerError(68, 78), scopeManagerError(109, 119)},
			},
			// ---- Reference spaces: import-equals RHS is a value reference ----
			{
				Code:   `const snake_case={}; import Alias=snake_case.Member;`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(7, 17), scopeManagerError(35, 45)},
			},
			// ---- Default export: named function body resolves to its declaration ----
			{
				Code:   `export default function snake_case(){return snake_case}`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(25, 35), scopeManagerError(45, 55)},
			},
			// ---- Default export: named class self-reference resolves to its class ----
			{
				Code:   `export default class snake_case{static x=snake_case}`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(22, 32), scopeManagerError(42, 52)},
			},
			// ---- Function ordering: declaration parameter precedes type parameter ----
			{
				Code:   `function ok<snake_case>(snake_case: unknown) { snake_case }`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(25, 44), scopeManagerError(48, 58)},
			},
			// ---- Function ordering: arrow parameter precedes type parameter ----
			{
				Code:   `const ok = <snake_case,>(snake_case: unknown) => snake_case`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(26, 45), scopeManagerError(50, 60)},
			},
			// ---- Function ordering: method parameter precedes type parameter ----
			{
				Code:   `class C { ok<snake_case>(snake_case: unknown) { return snake_case } }`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(26, 45), scopeManagerError(56, 66)},
			},
			// ---- Function ordering: type parameter anchors a body var merge ----
			{
				Code:   `function ok<snake_case>() { var snake_case; snake_case }`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(13, 23), scopeManagerError(45, 55)},
			},
			// ---- Import init: typed initializer keeps the ESTree binding range ----
			{
				Code:   `import { snake_case } from 'm'; const snake_case: number = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(10, 20), scopeManagerError(39, 57)},
			},
			// ---- Import init: an uninitialized merged declaration has no init read ----
			{
				Code:   `import { snake_case } from 'm'; let snake_case;`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(10, 20)},
			},
			// ---- Import init: every initialized declaration contributes an init read ----
			{
				Code:   `import { snake_case } from 'm'; let snake_case = 1, snake_case = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(10, 20), scopeManagerError(37, 47), scopeManagerError(53, 63)},
			},
			// ---- Import init: for-of variable declarations carry an init read ----
			{
				Code:   `import { snake_case } from 'm'; for (var snake_case of source) {}`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(10, 20), scopeManagerError(42, 52)},
			},
			// ---- Import init: destructuring defaults remain reportable when ignored ----
			{
				Code:    `import { snake_case } from 'm'; const { source_name: snake_case = value } = source;`,
				Options: map[string]any{"ignoreDestructuring": true, "ignoreImports": true},
				Errors:  []rule_tester.InvalidTestCaseError{scopeManagerError(54, 64)},
			},
			// ---- Ignore imports: a merged const still contributes its init report ----
			{
				Code:    `import { snake_case } from 'm'; const snake_case = 1;`,
				Options: map[string]any{"ignoreImports": true},
				Errors:  []rule_tester.InvalidTestCaseError{scopeManagerError(10, 20), scopeManagerError(39, 49)},
			},
			// ---- Ignore options: a merged class is still a reportable local binding ----
			{
				Code:    `import { snake_case } from 'm'; class snake_case {}`,
				Options: map[string]any{"ignoreDestructuring": true, "ignoreImports": true},
				Errors:  []rule_tester.InvalidTestCaseError{scopeManagerError(39, 49)},
			},
			// ---- Import-equals merge: it anchors the variable but has no init listener ----
			{
				Code:   `import snake_case = require('m'); const snake_case = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(8, 18)},
			},
			// ---- Ignore destructuring: a preceding type declaration remains the anchor ----
			{
				Code:    `type snake_case = {}; const { snake_case } = source`,
				Options: map[string]any{"ignoreDestructuring": true},
				Errors:  []rule_tester.InvalidTestCaseError{scopeManagerError(6, 16)},
			},
			// ---- Ignore imports: a preceding interface remains the anchor ----
			{
				Code:    `interface snake_case {}; import { snake_case } from "m"`,
				Options: map[string]any{"ignoreImports": true},
				Errors:  []rule_tester.InvalidTestCaseError{scopeManagerError(11, 21)},
			},
			// ---- Program reference: an unresolved script value is reported ----
			{
				Code:            `interface snake_case {}; snake_case`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors:          []rule_tester.InvalidTestCaseError{scopeManagerError(26, 36)},
			},
			// ---- Program global: a configured module global reports when not ignored ----
			{
				Code:            `interface snake_case {}; snake_case`,
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Globals:         map[string]any{"snake_case": "readonly"},
				Errors:          []rule_tester.InvalidTestCaseError{scopeManagerError(26, 36)},
			},
			// ---- Program global: inline off overrides configured readonly access ----
			{
				Code:            `/* global snake_case:off */ interface snake_case {}; snake_case`,
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Globals:         map[string]any{"snake_case": "readonly"},
				Errors:          []rule_tester.InvalidTestCaseError{scopeManagerError(54, 64)},
			},
			// ---- JSX casing: lowercase BMP local reports only its declaration ----
			{
				Code:   `const é_snake = () => null; const x = <é_snake />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(7, 14)},
			},
			// ---- JSX casing: uppercase hyphenated tag is a component reference ----
			{
				Code:   `const x = <Snake_case-tag />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(12, 26)},
			},
			// ---- JSX casing: astral first code unit is an unchanged high surrogate ----
			{
				Code:   `const x = <𐐨_snake-tag />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(12, 24)},
			},
			// ---- JSX parser split: Espree visits only the opening bare tag ----
			{
				Code:     `const x=<Snake_case></Snake_case>`,
				FileName: "camelcase-closing.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{scopeManagerError(10, 20)},
			},
			// ---- JSX parser split: typescript-eslint visits both bare tag names ----
			{
				Code:   `const x=<Snake_case></Snake_case>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(10, 20), scopeManagerError(23, 33)},
			},
			// ---- JSX parser split: Espree skips the closing member-expression root ----
			{
				Code:     `const x=<Snake_case.Part></Snake_case.Part>`,
				FileName: "camelcase-member-closing.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{scopeManagerError(10, 20)},
			},
			// ---- TSX namespace tag: interfaces do not satisfy value references ----
			{
				Code:   `interface snake_case {}; interface other_name {}; const x = <snake_case:other_name />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(62, 72), scopeManagerError(73, 83)},
			},
			// ---- TSX namespace tag: local values report declarations and both pieces ----
			{
				Code:   `const snake_case = 1, other_name = 2; const x = <snake_case:other_name />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(7, 17), scopeManagerError(23, 33), scopeManagerError(50, 60), scopeManagerError(61, 71)},
			},
			// ---- TSX namespace tag: unresolved local part reports on open and close ----
			{
				Code:   `namespace snake_case {}; const x = <snake_case:part_bad></snake_case:part_bad>;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(48, 56), scopeManagerError(70, 78)},
			},
			// ---- TSX shadowing: an inner namespace satisfies the outer-name piece ----
			{
				Code:   `const snake_case = 1; function f(){ namespace snake_case {}; const x=<snake_case:part_bad/> }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(7, 17), scopeManagerError(82, 90)},
			},
			// ---- TSX shadowing: a type-only interface falls through to the outer value; the local part stays unresolved ----
			{
				Code:   `const snake_case = 1; function f(){ interface snake_case {}; const x=<snake_case:part_bad/> }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{scopeManagerError(7, 17), scopeManagerError(71, 81), scopeManagerError(82, 90)},
			},
		},
	)
}
