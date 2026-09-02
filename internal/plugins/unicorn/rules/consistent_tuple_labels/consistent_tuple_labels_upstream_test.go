// TestConsistentTupleLabelsUpstream migrates the full valid/invalid suite from
// upstream test/consistent-tuple-labels.js in eslint-plugin-unicorn v74.0.0
// 1:1. Position assertions cover every invalid case. rslint-specific lock-in
// cases live in consistent_tuple_labels_extras_test.go.
package consistent_tuple_labels_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	consistent_tuple_labels "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_tuple_labels"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID = "consistent-tuple-labels"
	message   = "This tuple element should have a label, just like the other elements."
)

func expectedError(code, target string, occurrence int) rule_tester.InvalidTestCaseError {
	searchFrom := 0
	start := -1
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], target)
		if relative < 0 {
			panic("target not found in test input: " + target)
		}
		start = searchFrom + relative
		searchFrom = start + 1
	}

	lineStart := strings.LastIndex(code[:start], "\n") + 1
	line := strings.Count(code[:start], "\n") + 1
	column := start - lineStart + 1
	end := start + len(target)
	endLineStart := strings.LastIndex(code[:end], "\n") + 1
	endLine := strings.Count(code[:end], "\n") + 1
	endColumn := end - endLineStart + 1

	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts", Tsx: false}
}

func tsInvalid(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, FileName: "file.ts", Tsx: false, Errors: errors}
}

func TestConsistentTupleLabelsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_tuple_labels.ConsistentTupleLabelsRule,
		[]rule_tester.ValidTestCase{
			// ---- All labeled ----
			tsValid(`type Foo = [a: string, b: number];`),
			tsValid(`type Foo = [a: string, b: number, c: boolean];`),

			// ---- None labeled ----
			tsValid(`type Foo = [string, number];`),
			tsValid(`type Foo = [string, number, boolean];`),

			// ---- Fewer than two elements ----
			tsValid(`type Foo = [];`),
			tsValid(`type Foo = [string];`),
			tsValid(`type Foo = [a: string];`),

			// ---- Optional labeled elements ----
			tsValid(`type Foo = [a?: string, b?: number];`),
			tsValid(`type Foo = [a: string, b?: number];`),

			// ---- Optional unlabeled elements (all unlabeled) ----
			tsValid(`type Foo = [string?, number?];`),

			// ---- Labeled rest element (counts as labeled) ----
			tsValid(`type Foo = [a: string, ...b: number[]];`),

			// ---- Unlabeled rest element (all unlabeled) ----
			tsValid(`type Foo = [string, ...number[]];`),

			// ---- Nested tuples are evaluated independently ----
			tsValid(`type Foo = [[a: number, b: number]];`),
			tsValid(`type Foo = [a: [x: number], b: [y: number]];`),

			// ---- Not a tuple ----
			tsValid(`type Foo = string[];`),
			tsValid(`type Foo = readonly string[];`),
			tsValid(`type Foo = {a: string; b: number};`),
		},
		[]rule_tester.InvalidTestCase{
			tsInvalid(`type Foo = [a: string, number];`, expectedError(`type Foo = [a: string, number];`, `number`, 0)),
			tsInvalid(`type Foo = [string, b: number];`, expectedError(`type Foo = [string, b: number];`, `string`, 0)),
			tsInvalid(`type Foo = [a: string, b: number, c: boolean, d];`, expectedError(`type Foo = [a: string, b: number, c: boolean, d];`, `d`, 0)),
			tsInvalid(`type Foo = [a: string, number, c: boolean];`, expectedError(`type Foo = [a: string, number, c: boolean];`, `number`, 0)),

			// ---- Multiple unlabeled elements are each reported ----
			tsInvalid(
				`type Foo = [a: string, number, boolean];`,
				expectedError(`type Foo = [a: string, number, boolean];`, `number`, 0),
				expectedError(`type Foo = [a: string, number, boolean];`, `boolean`, 0),
			),

			// ---- Optional unlabeled element mixed with a labeled one ----
			tsInvalid(`type Foo = [a?: string, number?];`, expectedError(`type Foo = [a?: string, number?];`, `number?`, 0)),
			tsInvalid(`type Foo = [string?, b: number];`, expectedError(`type Foo = [string?, b: number];`, `string?`, 0)),

			// ---- Mixed via a labeled rest element ----
			tsInvalid(`type Foo = [string, ...rest: number[]];`, expectedError(`type Foo = [string, ...rest: number[]];`, `string`, 0)),

			// ---- Labeled head with an unlabeled rest element ----
			tsInvalid(`type Foo = [a: string, ...number[]];`, expectedError(`type Foo = [a: string, ...number[]];`, `...number[]`, 0)),

			// ---- Optional combined with a missing label ----
			tsInvalid(`type Foo = [a?: string, number];`, expectedError(`type Foo = [a?: string, number];`, `number`, 0)),

			// ---- Readonly tuple ----
			tsInvalid(`type Foo = readonly [a: string, number];`, expectedError(`type Foo = readonly [a: string, number];`, `number`, 0)),

			// ---- Nested tuple is inconsistent on its own ----
			tsInvalid(`type Foo = [[a: number, number]];`, expectedError(`type Foo = [[a: number, number]];`, `number`, 1)),

			// ---- Comments inside the tuple ----
			tsInvalid(`type Foo = [a: string, /* unlabeled */ number];`, expectedError(`type Foo = [a: string, /* unlabeled */ number];`, `number`, 0)),
		},
	)
}
