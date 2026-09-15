// TestNoNewBufferUpstream migrates the complete valid/invalid suite from
// eslint-plugin-unicorn v74.0.0 test/no-new-buffer.js.
package no_new_buffer_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_new_buffer "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_new_buffer"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	errorMessageUnknown = "`new Buffer()` is deprecated, use `Buffer.alloc()` or `Buffer.from()` instead."
)

func newBufferError(code, expression, messageID, message string, suggestions ...rule_tester.InvalidTestCaseSuggestion) rule_tester.InvalidTestCaseError {
	offset := strings.Index(code, expression)
	if offset < 0 {
		panic("expression not found: " + expression)
	}
	line, column := ecmaLineColumnForOffset(code, offset)
	endLine, endColumn := ecmaLineColumnForOffset(code, offset+len(expression))
	return rule_tester.InvalidTestCaseError{
		MessageId:   messageID,
		Message:     message,
		Line:        line,
		Column:      column,
		EndLine:     endLine,
		EndColumn:   endColumn,
		Suggestions: suggestions,
	}
}

func ecmaLineColumnForOffset(code string, offset int) (line int, column int) {
	sourceFile := parser.ParseSourceFile(
		ast.SourceFileParseOptions{FileName: "/fixture.js"},
		code,
		core.ScriptKindJS,
	)
	lineIndex, columnIndex := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, offset)
	return lineIndex + 1, int(columnIndex) + 1
}

func TestNewBufferErrorUsesECMACharacterLocations(t *testing.T) {
	for _, test := range []struct {
		code            string
		line, col       int
		endLine, endCol int
	}{
		{`"😀"; new Buffer(1)`, 1, 7, 1, 20},
		{"const value = 1;\r\nnew Buffer(1)", 2, 1, 2, 14},
	} {
		diagnostic := newBufferError(test.code, "new Buffer(1)", "error", "error")
		if diagnostic.Line != test.line || diagnostic.Column != test.col || diagnostic.EndLine != test.endLine || diagnostic.EndColumn != test.endCol {
			t.Errorf("%q: location = %d:%d-%d:%d, want %d:%d-%d:%d", test.code, diagnostic.Line, diagnostic.Column, diagnostic.EndLine, diagnostic.EndColumn, test.line, test.col, test.endLine, test.endCol)
		}
	}
}

func fixedNewBufferCase(code, expression, method, output string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Output:   []string{output},
		Errors: []rule_tester.InvalidTestCaseError{
			newBufferError(code, expression, "error", "`new Buffer()` is deprecated, use `Buffer."+method+"()` instead."),
		},
	}
}

func suggestedNewBufferCase(code, expression string) rule_tester.InvalidTestCase {
	withoutNew := strings.Replace(expression, "new ", "", 1)
	from := strings.Replace(withoutNew, "Buffer", "Buffer.from", 1)
	alloc := strings.Replace(withoutNew, "Buffer", "Buffer.alloc", 1)
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.js",
		Errors: []rule_tester.InvalidTestCaseError{
			newBufferError(code, expression, "error-unknown", errorMessageUnknown,
				rule_tester.InvalidTestCaseSuggestion{MessageId: "suggestion", Output: strings.Replace(code, expression, from, 1)},
				rule_tester.InvalidTestCaseSuggestion{MessageId: "suggestion", Output: strings.Replace(code, expression, alloc, 1)},
			),
		},
	}
}

func TestNoNewBufferUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `const buffer = Buffer`, FileName: "file.js"},
		{Code: `const buffer = new NotBuffer(1)`, FileName: "file.js"},
		{Code: `const buffer = Buffer.from('buf')`, FileName: "file.js"},
		{Code: `const buffer = Buffer.from('7468697320697320612074c3a97374', 'hex')`, FileName: "file.js"},
		{Code: `const buffer = Buffer.from([0x62, 0x75, 0x66, 0x66, 0x65, 0x72])`, FileName: "file.js"},
		{Code: `const buffer = Buffer.alloc(10)`, FileName: "file.js"},
	}

	unknownConditional := `const modes = new Set(['foo']);` + "\n" + `modes.clear();` + "\n" + `new Buffer(modes.size ? 'x' : 1);`
	unknownLogical := `const modes = new Set(["foo"]); modes.clear(); new Buffer((modes.size && "x") || value);`
	unknownAlias := `const alias = condition; var condition = true; new Buffer(alias ? "x" : 1);`
	unknownProperty := `const object = {value: true}; Object.defineProperty(object, "value", {get() { return false; }}); new Buffer(object.value ? "x" : value);`
	unknownAssertion := `const modes = new Set(["foo"]); modes.clear(); new Buffer((modes.size ? 1 : "x") as string);`

	invalid := []rule_tester.InvalidTestCase{
		suggestedNewBufferCase(unknownConditional, `new Buffer(modes.size ? 'x' : 1)`),
		suggestedNewBufferCase(unknownLogical, `new Buffer((modes.size && "x") || value)`),
		suggestedNewBufferCase(unknownAlias, `new Buffer(alias ? "x" : 1)`),
		suggestedNewBufferCase(unknownProperty, `new Buffer(object.value ? "x" : value)`),
		withFileName(suggestedNewBufferCase(unknownAssertion, `new Buffer((modes.size ? 1 : "x") as string)`), "file.ts"),

		fixedNewBufferCase(`const buffer = new Buffer([0x62, 0x75, 0x66, 0x66, 0x65, 0x72])`, `new Buffer([0x62, 0x75, 0x66, 0x66, 0x65, 0x72])`, "from", `const buffer = Buffer.from([0x62, 0x75, 0x66, 0x66, 0x65, 0x72])`),
		fixedNewBufferCase(`const buffer = new Buffer([0x62, bar])`, `new Buffer([0x62, bar])`, "from", `const buffer = Buffer.from([0x62, bar])`),
		fixedNewBufferCase(`const array = [0x62];`+"\n"+`const buffer = new Buffer(array);`, `new Buffer(array)`, "from", `const array = [0x62];`+"\n"+`const buffer = Buffer.from(array);`),
		suggestedNewBufferCase(`const arrayBuffer = new ArrayBuffer(10);`+"\n"+`const buffer = new Buffer(arrayBuffer);`, `new Buffer(arrayBuffer)`),
		fixedNewBufferCase(`const arrayBuffer = new ArrayBuffer(10);`+"\n"+`const buffer = new Buffer(arrayBuffer, 0, );`, `new Buffer(arrayBuffer, 0, )`, "from", `const arrayBuffer = new ArrayBuffer(10);`+"\n"+`const buffer = Buffer.from(arrayBuffer, 0, );`),
		fixedNewBufferCase(`const arrayBuffer = new ArrayBuffer(10);`+"\n"+`const buffer = new Buffer(arrayBuffer, 0, 2);`, `new Buffer(arrayBuffer, 0, 2)`, "from", `const arrayBuffer = new ArrayBuffer(10);`+"\n"+`const buffer = Buffer.from(arrayBuffer, 0, 2);`),

		fixedNewBufferCase(`const buffer = new Buffer(10);`, `new Buffer(10)`, "alloc", `const buffer = Buffer.alloc(10);`),
		fixedNewBufferCase(`const size = 10;`+"\n"+`const buffer = new Buffer(size);`, `new Buffer(size)`, "alloc", `const size = 10;`+"\n"+`const buffer = Buffer.alloc(size);`),
		fixedNewBufferCase(`new Buffer(foo.length)`, `new Buffer(foo.length)`, "alloc", `Buffer.alloc(foo.length)`),
		fixedNewBufferCase(`new Buffer(Math.min(foo, bar))`, `new Buffer(Math.min(foo, bar))`, "alloc", `Buffer.alloc(Math.min(foo, bar))`),

		fixedNewBufferCase(`const buffer = new Buffer("string");`, `new Buffer("string")`, "from", `const buffer = Buffer.from("string");`),
		fixedNewBufferCase(`const buffer = new Buffer("7468697320697320612074c3a97374", "hex")`, `new Buffer("7468697320697320612074c3a97374", "hex")`, "from", `const buffer = Buffer.from("7468697320697320612074c3a97374", "hex")`),
		fixedNewBufferCase(`const string = "string";`+"\n"+`const buffer = new Buffer(string);`, `new Buffer(string)`, "from", `const string = "string";`+"\n"+`const buffer = Buffer.from(string);`),
		fixedNewBufferCase("const buffer = new Buffer(`${unknown}`)", "new Buffer(`${unknown}`)", "from", "const buffer = Buffer.from(`${unknown}`)"),

		suggestedNewBufferCase(`const buffer = new (Buffer)(unknown)`, `new (Buffer)(unknown)`),
		fixedNewBufferCase(`const buffer = new Buffer(unknown, 2)`, `new Buffer(unknown, 2)`, "from", `const buffer = Buffer.from(unknown, 2)`),
		suggestedNewBufferCase(`const buffer = new Buffer(...unknown)`, `new Buffer(...unknown)`),
		fixedNewBufferCase("() => {\n\treturn new // 1\n\t\tBuffer();\n}", "new // 1\n\t\tBuffer()", "from", "() => {\n\treturn ( // 1\n\t\tBuffer.from());\n}"),
		fixedNewBufferCase("() => {\n\treturn (\n\t\tnew // 2\n\t\t\tBuffer()\n\t);\n}", "new // 2\n\t\t\tBuffer()", "from", "() => {\n\treturn (\n\t\t// 2\n\t\t\tBuffer.from()\n\t);\n}"),
		fixedNewBufferCase("() => {\n\treturn new // 3\n\t\t(Buffer);\n}", "new // 3\n\t\t(Buffer)", "from", "() => {\n\treturn ( // 3\n\t\t(Buffer.from)());\n}"),
		fixedNewBufferCase("() => {\n\treturn new // 4\n\t\tBuffer;\n}", "new // 4\n\t\tBuffer", "from", "() => {\n\treturn ( // 4\n\t\tBuffer.from());\n}"),
		fixedNewBufferCase("() => {\n\treturn (\n\t\tnew // 5\n\t\t\tBuffer\n\t);\n}", "new // 5\n\t\t\tBuffer", "from", "() => {\n\treturn (\n\t\t// 5\n\t\t\tBuffer.from()\n\t);\n}"),
		fixedNewBufferCase("() => {\n\treturn (\n\t\tnew // 6\n\t\t\t(Buffer)\n\t);\n}", "new // 6\n\t\t\t(Buffer)", "from", "() => {\n\treturn (\n\t\t// 6\n\t\t\t(Buffer.from)()\n\t);\n}"),
		fixedNewBufferCase(`const buffer = new /* comment */ Buffer()`, `new /* comment */ Buffer()`, "from", `const buffer = /* comment */ Buffer.from()`),
		fixedNewBufferCase(`const buffer = new /* comment */ Buffer`, `new /* comment */ Buffer`, "from", `const buffer = /* comment */ Buffer.from()`),
		withFileName(fixedNewBufferCase(`new Buffer(input, encoding);`, `new Buffer(input, encoding)`, "from", `Buffer.from(input, encoding);`), "file.ts"),
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_buffer.NoNewBufferRule, valid, invalid)
}

func withFileName(testCase rule_tester.InvalidTestCase, fileName string) rule_tester.InvalidTestCase {
	testCase.FileName = fileName
	return testCase
}
