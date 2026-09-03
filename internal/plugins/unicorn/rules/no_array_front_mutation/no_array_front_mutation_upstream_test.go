// TestNoArrayFrontMutationUpstream migrates the full valid/invalid suite from
// upstream test/no-array-front-mutation.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// no_array_front_mutation_extras_test.go.
package no_array_front_mutation_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_front_mutation"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "no-array-front-mutation"

func expectedMessage(method string) string {
	return "Avoid front-of-array mutation with `Array#" + method + "()`."
}

func methodError(code, method string, occurrence int) rule_tester.InvalidTestCaseError {
	needle := "." + method
	offset := -1
	searchFrom := 0
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], needle)
		if relative < 0 {
			panic("method not found in no-array-front-mutation test: " + method)
		}
		offset = searchFrom + relative + 1
		searchFrom = offset + len(method)
	}

	prefix := code[:offset]
	line := strings.Count(prefix, "\n") + 1
	lastNewline := strings.LastIndex(prefix, "\n")
	column := offset + 1
	if lastNewline >= 0 {
		column = offset - lastNewline
	}

	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   expectedMessage(method),
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len(method),
	}
}

func valid(code, fileName string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: fileName}
}

func invalid(code, fileName string, methods ...string) rule_tester.InvalidTestCase {
	seen := map[string]int{}
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(methods))
	for _, method := range methods {
		errors = append(errors, methodError(code, method, seen[method]))
		seen[method]++
	}
	return rule_tester.InvalidTestCase{Code: code, FileName: fileName, Errors: errors}
}

func TestNoArrayFrontMutationUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_array_front_mutation.NoArrayFrontMutationRule,
		[]rule_tester.ValidTestCase{
			// ---- Known non-array receiver (type information) ----
			valid(`function f(foo: Set<number>) { foo.shift(); }`, "file.ts"),
			valid(`function f(foo: Set<number>) { foo.unshift(value); }`, "file.ts"),

			valid(`array.shift`, "file.mjs"),
			valid(`array.unshift`, "file.mjs"),
			valid(`array.shift?.()`, "file.mjs"),
			valid(`array.unshift?.(value)`, "file.mjs"),
			valid(`array?.shift?.()`, "file.mjs"),
			valid(`array?.unshift?.(value)`, "file.mjs"),
			valid(`array["shift"]()`, "file.mjs"),
			valid(`array["unshift"](value)`, "file.mjs"),
			valid(`shift(array)`, "file.mjs"),
			valid(`unshift(array, value)`, "file.mjs"),
			valid(`Array.prototype.shift.call(array)`, "file.mjs"),
			valid(`Array.prototype.unshift.call(array, value)`, "file.mjs"),
			valid(`stream.unshift(chunk)`, "file.mjs"),
			valid(`this.unshift(chunk)`, "file.mjs"),
			valid(`this.stream.unshift(chunk)`, "file.mjs"),
			valid(`process.stdin.unshift(chunk)`, "file.mjs"),
			valid(`process.stdout.unshift(chunk)`, "file.mjs"),
			valid(`process.stderr.unshift(chunk)`, "file.mjs"),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`array.shift()`, "file.mjs", "shift"),
			invalid(`array.shift(extraArgument)`, "file.mjs", "shift"),
			invalid(`array?.shift()`, "file.mjs", "shift"),
			invalid(`array.unshift()`, "file.mjs", "unshift"),
			invalid(`array.unshift(value)`, "file.mjs", "unshift"),
			invalid(`array.unshift(...values)`, "file.mjs", "unshift"),
			invalid(`array?.unshift(value)`, "file.mjs", "unshift"),
			invalid(`stream.shift()`, "file.mjs", "shift"),
			invalid(`const item = array.shift()`, "file.mjs", "shift"),
			invalid(`const length = array.unshift(value)`, "file.mjs", "unshift"),
			invalid(`function getItem() { return array.shift(); }`, "file.mjs", "shift"),
			invalid(`while (array.shift()) {}`, "file.mjs", "shift"),
			invalid(`if (array.unshift(value)) {}`, "file.mjs", "unshift"),
			invalid(`for (; array.shift(); ) {}`, "file.mjs", "shift"),
			invalid(`(array as string[]).shift()`, "file.ts", "shift"),
			invalid(`array!.unshift(value)`, "file.ts", "unshift"),
		},
	)
}
