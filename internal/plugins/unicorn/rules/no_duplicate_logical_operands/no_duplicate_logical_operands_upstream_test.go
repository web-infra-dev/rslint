// Ported from eslint-plugin-unicorn v80.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/test/no-duplicate-logical-operands.js
// Positions and edits of the test.snapshot cases follow the upstream snapshot
// report (test/snapshots/no-duplicate-logical-operands.js.md); positions of the
// other upstream cases, which assert no ranges upstream, are computed from source.
package no_duplicate_logical_operands_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_duplicate_logical_operands"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID    = "no-duplicate-logical-operands"
	suggestionID = "no-duplicate-logical-operands/suggestion"
	message      = "This operand duplicates the left operand."
)

func valid(code, fileName string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: fileName}
}

// diagnostic asserts the complete range of the reported right operand.
func diagnostic(line, column, endLine, endColumn int, suggestions ...rule_tester.InvalidTestCaseSuggestion) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID, Message: message,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
		Suggestions: suggestions,
	}
}

// fixed is an autofixed single-line diagnostic.
func fixed(code, fileName, output string, column, endColumn int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code, FileName: fileName, Output: []string{output},
		Errors: []rule_tester.InvalidTestCaseError{diagnostic(1, column, 1, endColumn)},
	}
}

// suggested is a single-line diagnostic with one suggestion and no autofix.
func suggested(code, fileName, output string, column, endColumn int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code, FileName: fileName,
		Errors: []rule_tester.InvalidTestCaseError{
			diagnostic(1, column, 1, endColumn, rule_tester.InvalidTestCaseSuggestion{MessageId: suggestionID, Output: output}),
		},
	}
}

// reported is a single-line diagnostic without an autofix or a suggestion.
func reported(code, fileName string, column, endColumn int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code, FileName: fileName,
		Errors: []rule_tester.InvalidTestCaseError{diagnostic(1, column, 1, endColumn)},
	}
}

func TestNoDuplicateLogicalOperandsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_duplicate_logical_operands.NoDuplicateLogicalOperandsRule,
		[]rule_tester.ValidTestCase{
			// Different operands.
			valid(`foo && bar`, "file.js"),
			valid(`foo || bar`, "file.js"),

			// Non-adjacent duplicates are ignored.
			valid(`foo && bar && foo`, "file.js"),
			valid(`foo && (foo && bar)`, "file.js"),

			// Not logical AND/OR.
			valid(`foo ?? foo`, "file.js"),

			// Calls and other complex expressions can have side effects.
			valid(`getValue() && getValue()`, "file.js"),
			valid(`getValue() || getValue()`, "file.js"),
			valid(`foo.bar() && foo.bar()`, "file.js"),
			valid(`foo + bar && foo + bar`, "file.js"),
			valid(`await foo && await foo`, "file.js"),
			valid(`foo++ && foo++`, "file.js"),
			valid(`(foo = bar) && (foo = bar)`, "file.js"),

			// Computed keys with complex expressions can have side effects.
			valid(`foo[bar()] && foo[bar()]`, "file.js"),
			valid(`foo[bar + baz] || foo[bar + baz]`, "file.js"),

			// Optional chains are skipped.
			valid(`foo?.bar && foo?.bar`, "file.js"),
			valid(`foo?.[bar] || foo?.[bar]`, "file.js"),

			// Constants are handled by ESLint core `no-constant-binary-expression`.
			valid(`true && true`, "file.js"),
			valid(`0 || 0`, "file.js"),

			// Identifier reads inside `with` can trigger proxy traps or getters.
			{Code: `with (scope) {foo && foo;}`, FileName: "file.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},

			// Static string keys must be untagged templates without expressions.
			valid("foo[tag`bar`] || foo.bar;", "file.js"),
			valid("foo[`bar${value}`] || foo.bar;", "file.js"),
			// Optional members are never simple references.
			valid("foo?.[`bar`] || foo.bar;", "file.js"),

			// Documentation: correct examples.
			valid(`if (isReady) {}`, "file.js"),
			valid(`const result = value;`, "file.js"),
			valid(`if (object.isReady) {}`, "file.js"),
			valid(`object.isReady;`, "file.js"),
		},
		[]rule_tester.InvalidTestCase{
			// Identifiers are autofixed.
			fixed(`foo && foo`, "file.js", `foo`, 8, 11),
			fixed(`foo || foo`, "file.js", `foo`, 8, 11),
			fixed(`foo && bar && bar`, "file.js", `foo && bar`, 15, 18),
			fixed(`foo || bar || bar`, "file.js", `foo || bar`, 15, 18),
			fixed(`(foo) && (foo)`, "file.js", `(foo)`, 11, 14),
			fixed(`this && this`, "file.js", `this`, 9, 13),

			// Member expressions are suggestions because property reads can be observable.
			suggested(`foo.bar && foo.bar`, "file.js", `foo.bar`, 12, 19),
			suggested(`foo.bar || foo.bar`, "file.js", `foo.bar`, 12, 19),
			suggested(`foo.bar.baz && foo.bar.baz`, "file.js", `foo.bar.baz`, 16, 27),
			suggested(`foo && bar.baz && bar.baz`, "file.js", `foo && bar.baz`, 19, 26),
			suggested(`this.foo && this.foo`, "file.js", `this.foo`, 13, 21),
			suggested(
				`class Foo {#foo; method() {return this.#foo && this.#foo;}}`, "file.js",
				`class Foo {#foo; method() {return this.#foo;}}`, 48, 57,
			),
			suggested(
				`class Foo extends Bar {method() {return super.foo && super.foo;}}`, "file.js",
				`class Foo extends Bar {method() {return super.foo;}}`, 54, 63,
			),
			suggested(`foo[bar] && foo[bar]`, "file.js", `foo[bar]`, 13, 21),
			suggested(`foo["bar"] || foo.bar`, "file.js", `foo["bar"]`, 15, 22),

			// TypeScript wrappers around references are ignored for matching.
			suggested(`(foo as boolean) && (foo as boolean)`, "file.ts", `(foo as boolean)`, 22, 36),
			suggested(`(foo as Foo) && (foo as Bar)`, "file.ts", `(foo as Foo)`, 18, 28),
			suggested(`foo && (foo as Foo)`, "file.ts", `foo`, 9, 19),
			suggested(`(<boolean>foo) && (<boolean>foo)`, "file.ts", `(<boolean>foo)`, 20, 32),
			suggested(`foo! || foo!`, "file.ts", `foo!`, 9, 13),
			suggested(`foo! && foo`, "file.ts", `foo!`, 9, 12),
			suggested(`foo && foo!`, "file.ts", `foo`, 8, 12),
			suggested(`(foo satisfies boolean) && (foo satisfies boolean)`, "file.ts", `(foo satisfies boolean)`, 29, 50),
			suggested(`foo && (foo satisfies Foo)`, "file.ts", `foo`, 9, 26),

			// Comments outside the retained left operand disable fixes and suggestions.
			reported(`foo /* keep */ && foo`, "file.js", 19, 22),
			reported(`foo && /* keep */ foo`, "file.js", 19, 22),
			reported(`foo && bar /* keep */ && bar`, "file.js", 26, 29),
			reported(`foo && (foo /* keep */)`, "file.js", 9, 12),

			// Comments inside the retained left operand are preserved.
			fixed(`(foo /* keep */) && foo`, "file.js", `(foo /* keep */)`, 21, 24),

			// Static string keys, including templates, compare by decoded value.
			suggested("foo[`bar`] || foo.bar;", "file.js", "foo[`bar`];", 15, 22),
			suggested("foo[\"bar\"] && foo[`b\\u0061r`];", "file.js", "foo[\"bar\"];", 15, 30),
			suggested("foo[``] || foo[\"\"];", "file.js", "foo[``];", 12, 19),

			// Replacing a reference-context operand could change `this`, the tag
			// receiver, direct eval or delete, so no edit is offered.
			reported(`(object.method || object.method)()`, "file.js", 19, 32),
			reported("(object.tag && object.tag)`text`", "file.js", 16, 26),
			reported(`delete (object.value || object.value)`, "file.js", 25, 37),
			reported(`delete (value || value)`, "file.js", 18, 23),
			reported(`function check() { const local = 1; return (eval || eval)("typeof local"); }`, "file.js", 53, 57),
			reported(`(object.method || object.method)!()`, "file.ts", 19, 32),
			reported(`delete ((object.value || object.value) as unknown)`, "file.ts", 26, 38),
			// A constructor callee is not a reference context.
			fixed(`new (Constructor || Constructor)()`, "file.js", `new (Constructor)()`, 21, 32),

			// Documentation: incorrect examples.
			fixed(`if (isReady && isReady) {}`, "file.js", `if (isReady) {}`, 16, 23),
			fixed(`const result = value || value;`, "file.js", `const result = value;`, 25, 30),
			suggested(`if (object.isReady && object.isReady) {}`, "file.js", `if (object.isReady) {}`, 23, 37),
			suggested(`object.isReady && object.isReady;`, "file.js", `object.isReady;`, 19, 33),
		},
	)
}
