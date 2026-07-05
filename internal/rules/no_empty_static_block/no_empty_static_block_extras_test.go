// TestNoEmptyStaticBlockExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
//
// Dimension 4 walk (rows that don't apply to no-empty-static-block, with
// reasons):
//   - N/A receiver / expression wrappers ((X).y, X!.y, X as T, X satisfies T,
//     X?.y, X?.()): the rule only inspects class static block declarations, not
//     expression receivers.
//   - N/A access / key forms (identifier, string, numeric, private, computed,
//     element access): static blocks have no key.
//   - N/A function container variants (function declaration/expression/arrow,
//     methods, accessors, async/generator): the rule is specific to class static
//     blocks.
//   - N/A autofix boundaries: the rule has suggestions only, not an autofix.
//   - N/A body-absent class member forms (overload signatures, abstract,
//     declare): static blocks always have a body.
package no_empty_static_block

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyStaticBlockExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyStaticBlockRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: non-empty static block bodies do not report ----
			{Code: `class Foo { static { ; } }`},
			{Code: `class Foo { static { {} } }`},
			{Code: `class Foo { static { label: ; } }`},

			// ---- Dimension 4: comments inside otherwise empty static blocks are allowed ----
			{Code: `class Foo { static { /* intentionally empty */ } }`},
			{Code: "class Foo { static {\n  // intentionally empty\n} }"},
			{Code: "class Foo { static {\n  /* intentionally empty */\n} }"},

			// ---- Dimension 4: declaration/container forms ----
			{Code: `const Foo = class { static { foo(); } };`},
			{Code: `export default class Foo { static { foo(); } }`, FileName: "file.ts"},
			{Code: `class Foo extends Bar { static { super.name; } }`},
			{Code: `const Foo = class extends Bar { static { this.name; } };`},
			{Code: `class Foo { static value = 1; static { value; } }`},

			// ---- Dimension 4: graceful degradation around empty class/member forms ----
			{Code: `class Foo {}`},
			{Code: `class Foo { ; ; }`},
			{Code: `abstract class Foo { abstract method(): void }`},
			{Code: `class Foo { static /* before */ { /* inside */ } }`},

			// Locks in upstream StaticBlock() arm 1: body length > 0 short-circuits before comment checks.
			{Code: `class Foo { static { doWork(); /* trailing */ } }`},
			{Code: `class Foo { static { "use strict"; } }`},
			{Code: `class Foo { static { (() => {}); } }`},
			{Code: `class Foo { static { class Inner { static { /* ok */ } } } }`},
			{Code: `class Foo { static { const Inner = class { static { /* ok */ } }; } }`},

			// Locks in upstream StaticBlock() arm 2: comments before the closing brace make an empty block valid.
			{Code: `class Foo { static { /* block comment before close */ } }`},
			{Code: "class Foo { static {\n  // line comment before close\n} }"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: class expression static blocks are inspected ----
			invalidStaticBlockCase(`const Foo = class { static {} };`, 1, 28, 1, 30, `const Foo = class { static { /* empty */ } };`),
			invalidStaticBlockCase(`const Foo = class extends Bar { static {} };`, 1, 40, 1, 42, `const Foo = class extends Bar { static { /* empty */ } };`),

			// ---- Dimension 4: nested static blocks each keep their own traversal boundary ----
			{
				Code: `class Outer { static { class Inner { static {} } } static {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					staticBlockError(1, 45, 1, 47, `class Outer { static { class Inner { static { /* empty */ } } } static {} }`),
					staticBlockError(1, 59, 1, 61, `class Outer { static { class Inner { static {} } } static { /* empty */ } }`),
				},
			},
			invalidStaticBlockCase(`class Outer { static { if (flag) { class Inner { static {} } } } }`, 1, 57, 1, 59, `class Outer { static { if (flag) { class Inner { static { /* empty */ } } } } }`),
			invalidStaticBlockCase(`class Foo { static { const Inner = class { static {} }; } }`, 1, 51, 1, 53, `class Foo { static { const Inner = class { static { /* empty */ } }; } }`),

			// ---- Dimension 4: multiple empty static blocks in one class each report ----
			{
				Code: `class Foo { static {} static { work(); } static {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					staticBlockError(1, 20, 1, 22, `class Foo { static { /* empty */ } static { work(); } static {} }`),
					staticBlockError(1, 49, 1, 51, `class Foo { static {} static { work(); } static { /* empty */ } }`),
				},
			},

			// ---- Dimension 4: comments outside the braces do not make the block non-empty ----
			invalidStaticBlockCase(`class Foo { /* before */ static {} }`, 1, 33, 1, 35, `class Foo { /* before */ static { /* empty */ } }`),
			invalidStaticBlockCase(`class Foo { static {} /* after */ }`, 1, 20, 1, 22, `class Foo { static { /* empty */ } /* after */ }`),
			invalidStaticBlockCase("class Foo { static /* before */\n{} }", 2, 1, 2, 3, "class Foo { static /* before */\n{ /* empty */ } }"),

			// ---- Real-user: eslint#16318 proposed multiline empty class static blocks ----
			invalidStaticBlockCase("class Foo {\n  static {\n  }\n}", 2, 10, 3, 4, "class Foo {\n  static { /* empty */ }\n}"),

			// ---- Real-user: eslint#20056 highlights only braces and suggests inserting a comment ----
			invalidStaticBlockCase("class Foo {\n  static // setup hook intentionally blank while prototyping\n  {}\n}", 3, 3, 3, 5, "class Foo {\n  static // setup hook intentionally blank while prototyping\n  { /* empty */ }\n}"),

			// Locks in upstream StaticBlock() arm 3: empty body + no comment reports.
			invalidStaticBlockCase(`export default class Foo { static {} }`, 1, 35, 1, 37, `export default class Foo { static { /* empty */ } }`),
			invalidStaticBlockCase(`export default class { static {} }`, 1, 31, 1, 33, `export default class { static { /* empty */ } }`),
			invalidStaticBlockCase(`export class Foo { static {} }`, 1, 27, 1, 29, `export class Foo { static { /* empty */ } }`),
			invalidStaticBlockCase(`class Foo extends Bar { static {} }`, 1, 32, 1, 34, `class Foo extends Bar { static { /* empty */ } }`),
			invalidStaticBlockCase(`class Foo<T> { static {} }`, 1, 23, 1, 25, `class Foo<T> { static { /* empty */ } }`),
			invalidStaticBlockCase(`const Foo = mixin(class { static {} });`, 1, 34, 1, 36, `const Foo = mixin(class { static { /* empty */ } });`),
			invalidStaticBlockCase(`class Outer extends mixin(class { static {} }) {}`, 1, 42, 1, 44, `class Outer extends mixin(class { static { /* empty */ } }) {}`),
			invalidStaticBlockCase(`class Outer { [class { static {} }]() {} }`, 1, 31, 1, 33, `class Outer { [class { static { /* empty */ } }]() {} }`),
			invalidStaticBlockCase(`class Outer { static field = class { static {} }; }`, 1, 45, 1, 47, `class Outer { static field = class { static { /* empty */ } }; }`),
			invalidStaticBlockCase(`class Foo { static { /* ok */ } static {} }`, 1, 40, 1, 42, `class Foo { static { /* ok */ } static { /* empty */ } }`),
			invalidStaticBlockCase("class Foo { static\n{} }", 2, 1, 2, 3, "class Foo { static\n{ /* empty */ } }"),
		},
	)
}

func staticBlockError(line int, column int, endLine int, endColumn int, output string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unexpected",
		Message:   "Unexpected empty static block.",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{
			{
				MessageId: "suggestComment",
				Output:    output,
			},
		},
	}
}
