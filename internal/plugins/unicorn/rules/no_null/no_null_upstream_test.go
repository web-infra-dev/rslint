// TestNoNullUpstream migrates the complete valid/invalid suite from Unicorn
// v74.0.0 test/no-null.js. rslint-specific edge shapes and branch lock-ins live
// in no_null_extras_test.go.
package no_null_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_null "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_null"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const noNullMessage = "Use `undefined` instead of `null`."

var ignoreArguments = []any{map[string]any{"checkArguments": false}}

func nullOccurrenceRange(code string, occurrence int) (line, column, endLine, endColumn int) {
	searchFrom := 0
	offset := -1
	for index := 1; index <= occurrence; index++ {
		found := strings.Index(code[searchFrom:], "null")
		if found < 0 {
			panic("null occurrence not found in: " + code)
		}
		offset = searchFrom + found
		searchFrom = offset + len("null")
	}

	before := code[:offset]
	line = strings.Count(before, "\n") + 1
	lastNewline := strings.LastIndex(before, "\n")
	column = offset - lastNewline
	return line, column, line, column + len("null")
}

func replaceNullOccurrence(code string, occurrence int, replacement string) string {
	remaining := code
	offset := 0
	for index := 1; index <= occurrence; index++ {
		found := strings.Index(remaining, "null")
		if found < 0 {
			panic("null occurrence not found in: " + code)
		}
		offset += found
		if index == occurrence {
			return code[:offset] + replacement + code[offset+len("null"):]
		}
		offset += len("null")
		remaining = code[offset:]
	}
	panic("unreachable")
}

func noNullError(code string, occurrence int, suggestions ...rule_tester.InvalidTestCaseSuggestion) rule_tester.InvalidTestCaseError {
	line, column, endLine, endColumn := nullOccurrenceRange(code, occurrence)
	return rule_tester.InvalidTestCaseError{
		MessageId:   "error",
		Message:     noNullMessage,
		Line:        line,
		Column:      column,
		EndLine:     endLine,
		EndColumn:   endColumn,
		Suggestions: suggestions,
	}
}

func replacementCase(code string, options any, occurrences ...int) rule_tester.InvalidTestCase {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(occurrences))
	for _, occurrence := range occurrences {
		errors = append(errors, noNullError(code, occurrence, rule_tester.InvalidTestCaseSuggestion{
			MessageId: "replace",
			Output:    replaceNullOccurrence(code, occurrence, "undefined"),
		}))
	}
	return rule_tester.InvalidTestCase{Code: code, FileName: "file.js", Options: options, Errors: errors}
}

func fixedCase(code string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Output:   []string{replaceNullOccurrence(code, 1, "undefined")},
		Errors:   []rule_tester.InvalidTestCaseError{noNullError(code, 1)},
	}
}

func returnCase(code string, options any) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Options:  options,
		Errors: []rule_tester.InvalidTestCaseError{noNullError(code, 1,
			rule_tester.InvalidTestCaseSuggestion{
				MessageId: "remove",
				Output:    replaceNullOccurrence(code, 1, ""),
			},
			rule_tester.InvalidTestCaseSuggestion{
				MessageId: "replace",
				Output:    replaceNullOccurrence(code, 1, "undefined"),
			},
		)},
	}
}

func mutableVariableCase(code, removedOutput string) rule_tester.InvalidTestCase {
	return mutableVariableCaseAt(code, removedOutput, 1)
}

func mutableVariableCaseAt(code, removedOutput string, occurrence int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors: []rule_tester.InvalidTestCaseError{noNullError(code, occurrence,
			rule_tester.InvalidTestCaseSuggestion{MessageId: "remove", Output: removedOutput},
			rule_tester.InvalidTestCaseSuggestion{MessageId: "replace", Output: replaceNullOccurrence(code, occurrence, "undefined")},
		)},
	}
}

func TestNoNullUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_null.NoNullRule,
		[]rule_tester.ValidTestCase{
			{Code: `let foo`, FileName: "file.js"},
			{Code: `Object.create(null)`, FileName: "file.js"},
			{Code: `Object.create(null, {foo: {value:1}})`, FileName: "file.js"},
			{Code: `let insertedNode = parentNode.insertBefore(newNode, null)`, FileName: "file.js"},
			{Code: `let insertedNode = parentNode?.insertBefore(newNode, null)`, FileName: "file.js"},
			{Code: `const foo = "null";`, FileName: "file.js"},
			{Code: `Object.create()`, FileName: "file.js"},
			{Code: `Object.create(bar)`, FileName: "file.js"},
			{Code: `Object.create("null")`, FileName: "file.js"},
			{Code: `useRef(null)`, FileName: "file.js"},
			{Code: `React.useRef(null)`, FileName: "file.js"},
			{Code: `if (foo === null) {}`, FileName: "file.js"},
			{Code: `if (null === foo) {}`, FileName: "file.js"},
			{Code: `if (foo !== null) {}`, FileName: "file.js"},
			{Code: `if (null !== foo) {}`, FileName: "file.js"},
			{Code: `if (foo === null) {}`, FileName: "file.js", Options: []any{map[string]any{"checkStrictEquality": false}}},
			{Code: `if (null === foo) {}`, FileName: "file.js", Options: []any{map[string]any{"checkStrictEquality": false}}},
			{Code: `if (foo !== null) {}`, FileName: "file.js", Options: []any{map[string]any{"checkStrictEquality": false}}},
			{Code: `if (null !== foo) {}`, FileName: "file.js", Options: []any{map[string]any{"checkStrictEquality": false}}},
			{Code: `foo(null)`, FileName: "file.js", Options: ignoreArguments},
			{Code: `foo(bar, null)`, FileName: "file.js", Options: ignoreArguments},
			{Code: `drawingManager.setMap(null)`, FileName: "file.js", Options: ignoreArguments},
			{Code: `markers[index].setMap(null)`, FileName: "file.js", Options: ignoreArguments},
			{Code: `object?.method?.(null)`, FileName: "file.js", Options: ignoreArguments},
			{Code: `foo?.(null)`, FileName: "file.js", Options: ignoreArguments},
			{Code: `new HttpResponse(null)`, FileName: "file.js", Options: ignoreArguments},
			{Code: `new HttpResponse(body, null)`, FileName: "file.js", Options: ignoreArguments},
			{
				Code:            `foo = Object.create(null)`,
				FileName:        "file.js",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2019},
			},
		},
		[]rule_tester.InvalidTestCase{
			replacementCase(`const foo = null`, nil, 1),
			replacementCase(`foo(null)`, nil, 1),
			fixedCase(`if (foo == null) {}`),
			fixedCase(`if (foo != null) {}`),
			fixedCase(`if (null == foo) {}`),
			fixedCase(`if (null != foo) {}`),
			returnCase("function foo() {\n\treturn null;\n}", nil),
			mutableVariableCase(`let foo = null;`, `let foo;`),
			mutableVariableCase(`var foo = null;`, `var foo;`),
			mutableVariableCase(`var foo = 1, bar = null, baz = 2;`, `var foo = 1, bar, baz = 2;`),
			replacementCase(`const foo = null;`, nil, 1),
			replacementCase(`const foo = null;`, ignoreArguments, 1),
			returnCase("function foo() {\n\treturn null;\n}", ignoreArguments),
			replacementCase(`if (foo === null) {}`, []any{map[string]any{"checkArguments": false, "checkStrictEquality": true}}, 1),
			replacementCase(`foo([null])`, ignoreArguments, 1),
			replacementCase(`foo(bar ?? null)`, ignoreArguments, 1),
			replacementCase(`foo(...[null])`, ignoreArguments, 1),
			replacementCase(`new HttpResponse([null])`, ignoreArguments, 1),
			replacementCase(`if (foo === null) {}`, []any{map[string]any{"checkStrictEquality": true}}, 1),
			replacementCase(`if (null === foo) {}`, []any{map[string]any{"checkStrictEquality": true}}, 1),
			replacementCase(`if (foo !== null) {}`, []any{map[string]any{"checkStrictEquality": true}}, 1),
			replacementCase(`if (null !== foo) {}`, []any{map[string]any{"checkStrictEquality": true}}, 1),
			replacementCase(`new Object.create(null)`, nil, 1),
			replacementCase(`new foo.insertBefore(bar, null)`, nil, 1),
			replacementCase(`create(null)`, nil, 1),
			replacementCase(`insertBefore(bar, null)`, nil, 1),
			replacementCase(`Object["create"](null)`, nil, 1),
			replacementCase(`foo["insertBefore"](bar, null)`, nil, 1),
			replacementCase(`Object[create](null)`, nil, 1),
			replacementCase(`foo[insertBefore](bar, null)`, nil, 1),
			replacementCase(`Object[null](null)`, nil, 1, 2),
			replacementCase(`Object.notCreate(null)`, nil, 1),
			replacementCase(`foo.notInsertBefore(foo, null)`, nil, 1),
			replacementCase(`NotObject.create(null)`, nil, 1),
			replacementCase(`lib.Object.create(null)`, nil, 1),
			replacementCase(`Object.create(...[null])`, nil, 1),
			replacementCase(`Object.create(null, bar, extraArgument)`, nil, 1),
			replacementCase(`foo.insertBefore(null)`, nil, 1),
			replacementCase(`foo.insertBefore(foo, null, bar)`, nil, 1),
			replacementCase(`foo.insertBefore(...[foo], null)`, nil, 1),
			replacementCase(`foo.insertBefore(null, bar)`, nil, 1),
			replacementCase(`Object.create(bar, null)`, nil, 1),
		},
	)
}
