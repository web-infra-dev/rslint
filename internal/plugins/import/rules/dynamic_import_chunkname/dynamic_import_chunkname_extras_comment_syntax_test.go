package dynamic_import_chunkname_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/dynamic_import_chunkname"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Magic comment bodies are evaluated by upstream with vm.runInNewContext. The
// cases below pin how rslint's approximation of that evaluation behaves.
// Expected diagnostics were recorded from the upstream v2.32.0 rule under
// ESLint 8.57.1 with typescript-eslint; a body whose key is webpackChunkName
// accepts any text after it, so the snippets are placed after a chunk name.

// Syntax that V8 rejects even in code that never runs: the comment is evaluated as one source text, so a function or class body is checked too.
func TestDynamicImportChunknameCommentSyntaxInFunctions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule,
		[]rule_tester.ValidTestCase{
			// the same shapes in plain JavaScript are accepted
			{Code: `import(
  /* webpackChunkName: "a", x: () => 1 */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: () => null */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: () => void 0 */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: function () { return this } */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: function* () { yield 1 } */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: async () => 1 */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: class extends Object {} */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: class { m() { return 1 } static p = 1 } */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: { m() { return 1 }, get g() { return 1 }, set s(v) {} } */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: (a, b = 1, ...c) => [a, b, c] */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: () => typeof foo */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: () => { let a = 1; const b = null; return a + b } */
  'm',
)`},
		},
		[]rule_tester.InvalidTestCase{
			// TypeScript-only syntax is a syntax error even where the code never runs
			{
				Code: `import(
  /* webpackChunkName: "a", webpackPrefetch: (() => (true as boolean))() */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: () => (1 as number) */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: function () { return 1 as number } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: function* () { yield 1 as number } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: async () => (1 as number) */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: class { m() { return 1 as number } } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: class { static p = 1 as number } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: { m() { return 1 as number } } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: { get g() { return 1 as number } } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: { set s(v) { v as number } } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: () => foo! */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: () => (1 satisfies number) */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: (a: string) => a */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: (a?) => a */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: function (a?) {} */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: (): void => {} */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: function f<T>() {} */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: class implements Foo {} */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: class { m<T>() {} } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: class { constructor(public a) {} } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: class { p: number } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: class<T> {} */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: () => { interface I {} } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: () => { type T = number } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: () => { enum E {} } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: () => { let a: number } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: () => { class A extends B<number> {} } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
		},
	)
}

// A regular expression literal is an early error when its pattern or flags are invalid, wherever it appears in the comment; the TypeScript parser accepts any terminated literal.
func TestDynamicImportChunknameCommentRegexLiterals(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule,
		[]rule_tester.ValidTestCase{
			// valid regular expression literals
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /\.json$/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackExclude: /^\.\/locale\/[a-z]+$/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /a{1,2}/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /{/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /]/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /\1/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /[\d-x]/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /\p{L}/u */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /(?<=a)b/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /a/dgimsuy */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /[\p{L}--[a-z]]/v */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", webpackInclude: /(?<a>x)|(?<a>y)/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: /[/]/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: () => /a/ */
  'm',
)`},
			{Code: `import(
  /* webpackChunkName: "a", x: 1 / 2 / 3 */
  'm',
)`},
		},
		[]rule_tester.InvalidTestCase{
			// invalid regular expression literals in include and exclude filters
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /(/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /[z-a]/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackExclude: /a{2,1}/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /)/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackExclude: /(?<a>x)(?<a>y)/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /\p{Foo}/u */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /[\d-x]/u */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /\u{110000}/u */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /a/gg */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /a/x */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /a/uv */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /\k<a>/u */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /(?<a/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /+a/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /(?=a){2}/u */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackInclude: /(/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackExclude: /[z-a]/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", webpackInclude: /(/, webpackExclude: /\.json$/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			// invalid regular expression literals anywhere in the comment
			{
				Code: `import(
  /* webpackChunkName: "a", x: /(/.test("a") */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: () => /(/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: function () { return /[z-a]/ } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: class { m() { return /a{2,1}/ } } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: { m() { return /(/ } } */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: [/(/] */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: "import(\n  /* webpackChunkName: \"a\", x: `${/(/}` */\n  'm',\n)",
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackChunkName: "a", x: (1, /(/) */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a "webpack" comment with valid syntax`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			// valid regular expression literals
			{
				Code: `import(
  /* webpackInclude: /\.json$/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			{
				Code: `import(
  /* webpackInclude: /(?<lang>en|fr)\.json$/ */
  'm',
)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: `dynamic imports require a leading comment in the form /*webpackChunkName: ["']([0-9a-zA-Z-_/.]|\[(request|index)\])+["'],? */`, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
		},
	)
}
