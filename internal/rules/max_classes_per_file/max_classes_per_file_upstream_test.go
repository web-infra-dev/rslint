package max_classes_per_file

import (
	"strconv"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMaxClassesPerFileUpstream migrates the full valid/invalid suite from
// upstream eslint/tests/lib/rules/max-classes-per-file.js 1:1 (v10.8.1).
// Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in the max_classes_per_file_extras_test.go file.
func TestMaxClassesPerFileUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxClassesPerFileRule,
		[]rule_tester.ValidTestCase{
			// ---- valid ----
			{Code: "class Foo {}"},
			{Code: "var x = class {};"},
			{Code: "var x = 5;"},
			// SKIP: rule_tester exercises rule logic directly and does not
			// evaluate eslint-disable directive comments, which is a
			// framework-level (linter-pipeline) concern, not rule logic.
			{
				Code: "/* comment */\n/* eslint-disable rule-to-test/max-classes-per-file */\nclass A {}\nclass B {}",
				Skip: true,
			},
			{Code: "class Foo {}", Options: option(1)},
			{Code: "class Foo {}\nclass Bar {}", Options: option(2)},
			{Code: "class Foo {}", Options: option(map[string]interface{}{"max": 1})},
			{Code: "class Foo {}\nclass Bar {}", Options: option(map[string]interface{}{"max": 2})},
			{
				Code: `
                class Foo {}
                const myExpression = class {}
            `,
				Options: option(map[string]interface{}{"ignoreExpressions": true, "max": 1}),
			},
			{
				Code: `
                class Foo {}
                class Bar {}
                const myExpression = class {}
            `,
				Options: option(map[string]interface{}{"ignoreExpressions": true, "max": 2}),
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid ----
			{
				Code: "class Foo {}\nclass Bar {}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 13,
				}},
			},
			{
				Code: "class Foo {}\nconst myExpression = class {}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 30,
				}},
			},
			{
				Code: "var x = class {};\nvar y = class {};",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 18,
				}},
			},
			{
				Code: "class Foo {}\nvar x = class {};",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 18,
				}},
			},
			{
				Code:    "class Foo {} class Bar {}",
				Options: option(1),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 1, EndColumn: 26,
				}},
			},
			{
				Code:    "class Foo {} class Bar {} class Baz {}",
				Options: option(2),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(3, 2),
					Line:      1, Column: 1, EndLine: 1, EndColumn: 39,
				}},
			},
			{
				Code: `
                class Foo {}
                class Bar {}
                const myExpression = class {}
            `,
				Options: option(map[string]interface{}{"ignoreExpressions": true, "max": 1}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      2, Column: 17, EndLine: 4, EndColumn: 46,
				}},
			},
			{
				Code: `
                class Foo {}
                class Bar {}
                class Baz {}
                const myExpression = class {}
            `,
				Options: option(map[string]interface{}{"ignoreExpressions": true, "max": 2}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(3, 2),
					Line:      2, Column: 17, EndLine: 5, EndColumn: 46,
				}},
			},
			{
				Code: "/* comment */\nclass A {}\nclass B {}\n/* comment */",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      2, Column: 1, EndLine: 3, EndColumn: 11,
				}},
			},
		},
	)
}

func option(v any) []interface{} {
	return []interface{}{v}
}

func maximumExceededMessage(count, maxAllowed int) string {
	return "File has too many classes (" + strconv.Itoa(count) +
		"). Maximum allowed is " + strconv.Itoa(maxAllowed) + "."
}
