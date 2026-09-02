// TestNoUnnecessaryArrayFlatDepthUpstream migrates the complete valid/invalid
// suite from eslint-plugin-unicorn v74.0.0. rslint-specific edge-shape and
// branch lock-in cases live in no_unnecessary_array_flat_depth_extras_test.go.
package no_unnecessary_array_flat_depth_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_unnecessary_array_flat_depth "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unnecessary_array_flat_depth"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID = "no-unnecessary-array-flat-depth"
	message   = "Passing `1` as the `depth` argument is unnecessary."
)

func expectedDepthError(code, target string, occurrence int) rule_tester.InvalidTestCaseError {
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

func depthInvalid(code, target, output, fileName string) rule_tester.InvalidTestCase {
	return depthInvalidAt(code, target, output, fileName, 0)
}

func depthInvalidAt(code, target, output, fileName string, occurrence int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: fileName,
		Output:   []string{output},
		Errors:   []rule_tester.InvalidTestCaseError{expectedDepthError(code, target, occurrence)},
	}
}

func TestNoUnnecessaryArrayFlatDepthUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_unnecessary_array_flat_depth.NoUnnecessaryArrayFlatDepthRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     `function f(foo: {flat(depth: number): void}) { foo.flat(1); }`,
				FileName: "file.ts",
			},
			{Code: `foo.flat()`, FileName: "file.js"},
			{Code: `foo.flat?.(1)`, FileName: "file.js"},
			{Code: `foo?.flat()`, FileName: "file.js"},
			{Code: `foo.flat(1, extra)`, FileName: "file.js"},
			{Code: `flat(1)`, FileName: "file.js"},
			{Code: `new foo.flat(1)`, FileName: "file.js"},
			{Code: `const ONE = 1; foo.flat(ONE)`, FileName: "file.js"},
			{Code: `foo.notFlat(1)`, FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			depthInvalid(`foo.flat(1)`, `1`, `foo.flat()`, "file.js"),
			depthInvalid(`foo.flat(1.0)`, `1.0`, `foo.flat()`, "file.js"),
			depthInvalid(`foo.flat(0b01)`, `0b01`, `foo.flat()`, "file.js"),
			depthInvalid(`foo?.flat(1)`, `1`, `foo?.flat()`, "file.js"),
			depthInvalid(
				`function f(foo: number[][]) { foo.flat(1); }`,
				`1`,
				`function f(foo: number[][]) { foo.flat(); }`,
				"file.ts",
			),
		},
	)
}
