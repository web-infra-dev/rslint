// TestNoUnreadableNewExpressionUpstream migrates the complete valid/invalid
// suite from eslint-plugin-unicorn v74.0.0. Rslint-specific edge-shape and
// branch lock-in cases live in no_unreadable_new_expression_extras_test.go.
package no_unreadable_new_expression_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_unreadable_new_expression "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unreadable_new_expression"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	memberAccessMessageID       = "member-access"
	memberAccessMessage         = "Do not access members directly from a `new` expression."
	complexConstructorMessageID = "complex-constructor"
	complexConstructorMessage   = "Do not use a complex expression as a constructor."
)

func sourcePosition(code string, offset int) (int, int) {
	prefix := code[:offset]
	line := strings.Count(prefix, "\n") + 1
	lastNewline := strings.LastIndex(prefix, "\n")
	column := offset + 1
	if lastNewline >= 0 {
		column = offset - lastNewline
	}
	return line, column
}

func expectedError(code, target string, occurrence int, messageID, message string) rule_tester.InvalidTestCaseError {
	searchFrom := 0
	start := -1
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], target)
		if relative < 0 {
			panic("target not found in no-unreadable-new-expression test: " + target)
		}
		start = searchFrom + relative
		searchFrom = start + len(target)
	}

	line, column := sourcePosition(code, start)
	endLine, endColumn := sourcePosition(code, start+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func memberAccessError(code, target string, occurrence int) rule_tester.InvalidTestCaseError {
	return expectedError(code, target, occurrence, memberAccessMessageID, memberAccessMessage)
}

func complexConstructorError(code, target string, occurrence int) rule_tester.InvalidTestCaseError {
	return expectedError(code, target, occurrence, complexConstructorMessageID, complexConstructorMessage)
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.mjs"}
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts"}
}

func invalid(code, fileName string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, FileName: fileName, Errors: errors}
}

func TestNoUnreadableNewExpressionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_unreadable_new_expression.NoUnreadableNewExpressionRule,
		[]rule_tester.ValidTestCase{
			jsValid(`const foo = new Foo();`),
			jsValid(`const foo = new Foo;`),
			jsValid(`const foo = new Foo(bar);`),
			jsValid(`const foo = new Foo(...bar);`),
			jsValid(`const foo = new Foo.Bar;`),
			jsValid(`const bar = new foo.Bar();`),
			jsValid(`const bar = new foo.bar.Baz();`),
			jsValid(`const formatter = new Intl.ListFormat("en-US", {type: "disjunction"});`),
			jsValid(`const foo = Foo().Bar;`),
			jsValid(`const foo = Foo().Bar();`),
			jsValid(`const foo = Foo.Bar();`),
			jsValid(`const foo = new Foo(); foo.getBar();`),
			jsValid(`const foo = new Foo(); const Bar = foo.Bar;`),
			jsValid(`const {Bar} = foo; const bar = new Bar();`),
			jsValid(`const Bar = foo.Bar; const bar = new Bar();`),
			jsValid(`const bar = Foo ? new Foo() : foo.Bar;`),
			tsValid(`const foo = new Foo<Type>();`),
			tsValid(`const foo = new Foo<Type>;`),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`const bar = new Foo().getBar();`, "file.mjs", memberAccessError(`const bar = new Foo().getBar();`, "getBar", 0)),
			invalid(`const bar = (new Foo()).getBar();`, "file.mjs", memberAccessError(`const bar = (new Foo()).getBar();`, "getBar", 0)),
			invalid(`const Bar = new Foo().Bar;`, "file.mjs", memberAccessError(`const Bar = new Foo().Bar;`, "Bar", 1)),
			invalid(`const Bar = (new Foo()).Bar;`, "file.mjs", memberAccessError(`const Bar = (new Foo()).Bar;`, "Bar", 1)),
			invalid(`const Bar = (new Foo).Bar;`, "file.mjs", memberAccessError(`const Bar = (new Foo).Bar;`, "Bar", 1)),
			invalid(`const Bar = new Foo()[Bar];`, "file.mjs", memberAccessError(`const Bar = new Foo()[Bar];`, "Bar", 1)),
			invalid(`const Bar = (new Foo())[Bar];`, "file.mjs", memberAccessError(`const Bar = (new Foo())[Bar];`, "Bar", 1)),
			invalid(`const bar = (new Foo)?.getBar();`, "file.mjs", memberAccessError(`const bar = (new Foo)?.getBar();`, "getBar", 0)),
			invalid(`const baz = new Foo().bar.baz;`, "file.mjs", memberAccessError(`const baz = new Foo().bar.baz;`, "bar", 0)),
			invalid("new Foo().bar`x`;", "file.mjs", memberAccessError("new Foo().bar`x`;", "bar", 0)),
			invalid(`const Bar = new (Foo().Bar);`, "file.mjs", complexConstructorError(`const Bar = new (Foo().Bar);`, "Foo().Bar", 0)),
			invalid(`const bar = new foo[Bar]();`, "file.mjs", complexConstructorError(`const bar = new foo[Bar]();`, "foo[Bar]", 0)),
			invalid(`const bar = (new foo).Bar();`, "file.mjs", memberAccessError(`const bar = (new foo).Bar();`, "Bar", 0)),
			invalid(`const bar = new foo().Bar();`, "file.mjs", memberAccessError(`const bar = new foo().Bar();`, "Bar", 0)),
			invalid(`const bar = new (foo().Bar)();`, "file.mjs", complexConstructorError(`const bar = new (foo().Bar)();`, "foo().Bar", 0)),
			invalid(`const bar = new (foo())();`, "file.mjs", complexConstructorError(`const bar = new (foo())();`, "foo()", 0)),
			invalid(`const bar = new (foo ? Foo : Bar)();`, "file.mjs", complexConstructorError(`const bar = new (foo ? Foo : Bar)();`, "foo ? Foo : Bar", 0)),
			invalid(`const bar = new class {}();`, "file.mjs", complexConstructorError(`const bar = new class {}();`, "class {}", 0)),
			invalid(`const timestamp = new Date().getTime();`, "file.mjs", memberAccessError(`const timestamp = new Date().getTime();`, "getTime", 0)),
			invalid(`const formatted = new Intl.ListFormat("en-US", {type: "disjunction"}).format(words);`, "file.mjs", memberAccessError(`const formatted = new Intl.ListFormat("en-US", {type: "disjunction"}).format(words);`, "format", 1)),
			invalid(`const foo = new (Foo as typeof Bar)();`, "file.ts", complexConstructorError(`const foo = new (Foo as typeof Bar)();`, "Foo as typeof Bar", 0)),
			invalid(`const bar = (new Foo<Type>()).bar;`, "file.ts", memberAccessError(`const bar = (new Foo<Type>()).bar;`, "bar", 1)),
		},
	)
}
