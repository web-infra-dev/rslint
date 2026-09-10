package prefer_set_size_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_set_size"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferSetSizeExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_set_size.PreferSetSizeRule,
		[]rule_tester.ValidTestCase{
			// Static bracket access and optional access are distinct from upstream's
			// non-optional dot-member shape.
			{Code: "Array['from'](new Set(values)).length", FileName: "file.js"},
			{Code: "Array.from(new Set(values))?.length", FileName: "file.js"},
			{Code: "[...new Set(values)]?.length", FileName: "file.js"},
			// A TS assertion on the conversion itself is visible to upstream and does
			// not count as one of its two conversion shapes.
			{Code: "([...(set as Set<string>)] as string[]).length", FileName: "file.ts"},
		},
		[]rule_tester.InvalidTestCase{
			// Const aliases recurse, while `let` deliberately did not match in the
			// JavaScript upstream parser.
			invalid("const first = new Set(values); const second = first; [...second].length", "const first = new Set(values); const second = first; second.size", "file.js"),
			// A Set construction without constructor parentheses needs a wrapper once
			// it becomes the object of a member expression.
			invalid("[...new Set].length", "(new Set).size", "file.js"),
			invalid("[...(flag ? new Set() : new Set())].length", "(flag ? new Set() : new Set()).size", "file.js"),
			invalid("[...(sideEffect(), new Set())].length", "(sideEffect(), new Set()).size", "file.js"),
			invalid("function size(value: unknown) { return [...(new Set() satisfies Set)].length; }", "function size(value: unknown) { return (new Set() satisfies Set).size; }", "file.ts"),
			// Comments nested in Set survive; an outer conversion comment suppresses
			// the whole fix rather than dropping author text.
			invalid("[...new /* retained */ Set(values)].length", "new /* retained */ Set(values).size", "file.js"),
			{Code: "Array.from(/* outer */ new /* retained */ Set(values)).length", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{errorAtLength("Array.from(/* outer */ new /* retained */ Set(values)).length", 0)}},
			// Authored TypeScript wrappers remain visible in the replacement and are
			// parenthesized before they become a member-expression object.
			{Code: "function size(value: unknown) { return Array.from(value satisfies Set<string>).length; }", FileName: "file.ts"},
		},
	)
}
