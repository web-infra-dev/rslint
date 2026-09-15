// TestPreferSetSizeUpstream migrates eslint-plugin-unicorn v74.0.0's complete
// prefer-set-size test suite. rslint-specific AST and fix-safety cases live in
// prefer_set_size_extras_test.go.
package prefer_set_size_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_set_size"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	preferSetSizeMessageID = "prefer-set-size"
	preferSetSizeMessage   = "Prefer using `Set#size` instead of `Array#length`."
)

func errorAtLength(code string, occurrence int) rule_tester.InvalidTestCaseError {
	start := -1
	searchFrom := 0
	for range occurrence + 1 {
		relative := strings.Index(code[searchFrom:], "length")
		if relative < 0 {
			panic("length not found")
		}
		start = searchFrom + relative
		searchFrom = start + 1
	}
	lineStart := strings.LastIndex(code[:start], "\n") + 1
	return rule_tester.InvalidTestCaseError{
		MessageId: preferSetSizeMessageID,
		Message:   preferSetSizeMessage,
		Line:      strings.Count(code[:start], "\n") + 1,
		Column:    start - lineStart + 1,
		EndLine:   strings.Count(code[:start+len("length")], "\n") + 1,
		EndColumn: start - lineStart + len("length") + 1,
	}
}

func invalid(code, output string, fileName string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code, FileName: fileName,
		Output: []string{output},
		Errors: []rule_tester.InvalidTestCaseError{errorAtLength(code, 0)},
	}
}

func TestPreferSetSizeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_set_size.PreferSetSizeRule,
		[]rule_tester.ValidTestCase{
			{Code: "new Set(foo).size", FileName: "file.js"},
			{Code: "for (const foo of bar) console.log([...foo].length)", FileName: "file.js"},
			{Code: "[...new Set(array), foo].length", FileName: "file.js"},
			{Code: "[foo, ...new Set(array), ].length", FileName: "file.js"},
			{Code: "[...new Set(array)].notLength", FileName: "file.js"},
			{Code: "[...new Set(array)]?.length", FileName: "file.js"},
			{Code: "[...new Set(array)][length]", FileName: "file.js"},
			{Code: "[...new Set(array)][\"length\"]", FileName: "file.js"},
			{Code: "[...new NotSet(array)].length", FileName: "file.js"},
			{Code: "[...Set(array)].length", FileName: "file.js"},
			{Code: "const foo = new NotSet([]);[...foo].length;", FileName: "file.js"},
			{Code: "let foo = new Set([]);[...foo].length;", FileName: "file.js"},
			{Code: "const {foo} = new Set([]);[...foo].length;", FileName: "file.js"},
			{Code: "const [foo] = new Set([]);[...foo].length;", FileName: "file.js"},
			{Code: "[...foo].length", FileName: "file.js"},
			{Code: "var foo = new Set(); var foo = new Set(); [...foo].length", FileName: "file.js"},
			{Code: "[,].length", FileName: "file.js"},
			{Code: "Array.from(foo).length", FileName: "file.js"},
			{Code: "Array.from(new NotSet(array)).length", FileName: "file.js"},
			{Code: "Array.from(Set(array)).length", FileName: "file.js"},
			{Code: "Array.from(new Set(array)).notLength", FileName: "file.js"},
			{Code: "Array.from(new Set(array))?.length", FileName: "file.js"},
			{Code: "Array.from(new Set(array))[length]", FileName: "file.js"},
			{Code: "Array.from(new Set(array))[\"length\"]", FileName: "file.js"},
			{Code: "Array.from(new Set(array), mapFn).length", FileName: "file.js"},
			{Code: "Array?.from(new Set(array)).length", FileName: "file.js"},
			{Code: "Array.from?.(new Set(array)).length", FileName: "file.js"},
			{Code: "const foo = new NotSet([]);Array.from(foo).length;", FileName: "file.js"},
			{Code: "let foo = new Set([]);Array.from(foo).length;", FileName: "file.js"},
			{Code: "const {foo} = new Set([]);Array.from(foo).length;", FileName: "file.js"},
			{Code: "const [foo] = new Set([]);Array.from(foo).length;", FileName: "file.js"},
			{Code: "var foo = new Set(); var foo = new Set(); Array.from(foo).length", FileName: "file.js"},
			{Code: "NotArray.from(new Set(array)).length", FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			invalid("[...new Set(array)].length", "new Set(array).size", "file.js"),
			invalid("const foo = new Set([]);\nconsole.log([...foo].length);", "const foo = new Set([]);\nconsole.log(foo.size);", "file.js"),
			invalid("function isUnique(array) {\n\treturn[...new Set(array)].length === array.length\n}", "function isUnique(array) {\n\treturn new Set(array).size === array.length\n}", "file.js"),
			invalid("[...new Set(array),].length", "new Set(array).size", "file.js"),
			invalid("[...(( new Set(array) ))].length", "new Set(array).size", "file.js"),
			invalid("(( [...new Set(array)] )).length", "(( new Set(array) )).size", "file.js"),
			invalid("foo\n;[...new Set(array)].length", "foo\n;new Set(array).size", "file.js"),
			{Code: "[/* comment */...new Set(array)].length", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{errorAtLength("[/* comment */...new Set(array)].length", 0)}},
			invalid("[...new /* comment */ Set(array)].length", "new /* comment */ Set(array).size", "file.js"),
			invalid("Array.from(new Set(array)).length", "new Set(array).size", "file.js"),
			invalid("const foo = new Set([]);\nconsole.log(Array.from(foo).length);", "const foo = new Set([]);\nconsole.log(foo.size);", "file.js"),
			invalid("Array.from((( new Set(array) ))).length", "new Set(array).size", "file.js"),
			invalid("(( Array.from(new Set(array)) )).length", "(( new Set(array) )).size", "file.js"),
			{Code: "Array.from(/* comment */ new Set(array)).length", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{errorAtLength("Array.from(/* comment */ new Set(array)).length", 0)}},
			invalid("Array.from(new /* comment */ Set(array)).length", "new /* comment */ Set(array).size", "file.js"),
			invalid("function isUnique(array) {\n\treturn Array.from(new Set(array)).length === array.length\n}", "function isUnique(array) {\n\treturn new Set(array).size === array.length\n}", "file.js"),
			invalid("function getSize(set: Set<string>) { return Array.from(set).length; }", "function getSize(set: Set<string>) { return set.size; }", "file.ts"),
			invalid("function getSize(set: ReadonlySet<string>) { return Array.from(set).length; }", "function getSize(set: ReadonlySet<string>) { return set.size; }", "file.ts"),
			invalid("function getSize(set: unknown) { return [...(set as Set<string>)].length; }", "function getSize(set: unknown) { return (set as Set<string>).size; }", "file.ts"),
			invalid("function getSize(set: unknown) { return Array.from(set as Set<string>).length; }", "function getSize(set: unknown) { return (set as Set<string>).size; }", "file.ts"),
		},
	)
}
