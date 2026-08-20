package id_match

// TestIdMatchExtrasTypescript locks in the names TypeScript adds to a file that
// ESLint's own AST and scope model have no place for: the standard library's
// types, the `const` of an `as const`, and the class constructor that tsgo
// spells with a keyword instead of a name. Its siblings are
// id_match_extras_dim4_test.go, id_match_extras_branches_test.go and
// id_match_extras_realuser_test.go.

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIdMatchExtrasTypescript(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&IdMatchRule,
		[]rule_tester.ValidTestCase{
			// ---- A standard-library type is not the author's to name ----
			{
				Code:    `let x: Record<string, number>;`,
				Options: []any{`^x$`},
			},
			{
				Code:    `let x: ReadonlyArray<number>;`,
				Options: []any{`^x$`},
			},
			// ---- `as const` names no declaration ----
			{
				Code:    `const x = 1 as const;`,
				Options: []any{`^x$`},
			},
			// ---- A class constructor is checked like any other name ----
			{
				Code:    `class foo { constructor() {} }`,
				Options: []any{`^(foo|constructor)$`},
			},
			{
				Code:    `class foo { constructor() {} }`,
				Options: []any{`^foo$`, map[string]any{"onlyDeclarations": true}},
			},
			// ---- A catch parameter is not a variable declaration ----
			{
				Code:    `try {} catch (error_1) {}`,
				Options: []any{`^[^_]+$`, map[string]any{"onlyDeclarations": true}},
			},
			// ---- A string-literal-named constructor is a Literal to ESLint, not an identifier ----
			{
				Code:    `class foo { 'constructor'() {} }`,
				Options: []any{`^foo$`},
			},
			// ---- A string-literal-named constructor is a Literal to ESLint, not an identifier ----
			{
				Code:    `class foo { "constructor"() {} }`,
				Options: []any{`^foo$`},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- A class constructor is checked like any other name ----
			{
				Code:    `class foo { constructor() {} }`,
				Options: []any{`^foo$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'constructor' does not match the pattern '^foo$'.`,
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 24,
					},
				},
			},
			{
				Code:    `class foo { private constructor() {} }`,
				Options: []any{`^foo$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'constructor' does not match the pattern '^foo$'.`,
						Line:      1,
						Column:    21,
						EndLine:   1,
						EndColumn: 32,
					},
				},
			},
			{
				Code: `class foo {
  constructor(a: string);
  constructor(a?: string) {}
}`,
				Options: []any{`^(foo|a)$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'constructor' does not match the pattern '^(foo|a)$'.`,
						Line:      2,
						Column:    3,
						EndLine:   2,
						EndColumn: 14,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'constructor' does not match the pattern '^(foo|a)$'.`,
						Line:      3,
						Column:    3,
						EndLine:   3,
						EndColumn: 14,
					},
				},
			},
			// ---- A catch parameter is not a variable declaration ----
			{
				Code:    `try {} catch (error_1) {}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'error_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 22,
					},
				},
			},
			// ---- A type the file declares is the author's ----
			{
				Code: `interface Local_1 {}
let x: Local_1;`,
				Options: []any{`^x$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Local_1' does not match the pattern '^x$'.`,
						Line:      1,
						Column:    11,
						EndLine:   1,
						EndColumn: 18,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Local_1' does not match the pattern '^x$'.`,
						Line:      2,
						Column:    8,
						EndLine:   2,
						EndColumn: 15,
					},
				},
			},
			// ---- A type nothing declares is still the author's ----
			{
				Code:    `let x: Undeclared_1;`,
				Options: []any{`^x$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Undeclared_1' does not match the pattern '^x$'.`,
						Line:      1,
						Column:    8,
						EndLine:   1,
						EndColumn: 20,
					},
				},
			},
			// ---- A type the file imports is the author's too ----
			{
				Code: `import type { Foo_1 } from './m';
let x: Foo_1;`,
				Options: []any{`^x$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Foo_1' does not match the pattern '^x$'.`,
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 20,
					},
					{
						MessageId: "notMatch",
						Message:   `Identifier 'Foo_1' does not match the pattern '^x$'.`,
						Line:      2,
						Column:    8,
						EndLine:   2,
						EndColumn: 13,
					},
				},
			},
			// ---- A parameter is reported over its name, not over its type annotation ----
			{
				Code:    `function f(a_1: string) {}`,
				Options: []any{`^[^_]+$`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'a_1' does not match the pattern '^[^_]+$'.`,
						Line:      1,
						Column:    12,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			// ---- A name written in a JSDoc comment is not a name in the file ----
			{
				Code: `/** @typedef {{a_x: number}} T_x */
let w_x;`,
				Options:  []any{`^[^_]+$`},
				FileName: "jsdoc-id-match.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "notMatch",
						Message:   `Identifier 'w_x' does not match the pattern '^[^_]+$'.`,
						Line:      2,
						Column:    5,
						EndLine:   2,
						EndColumn: 8,
					},
				},
			},
		},
	)
}
