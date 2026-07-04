package prefer_numeric_literals

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferNumericLiteralsUpstream migrates the full valid/invalid suite from upstream tests/lib/rules/prefer-numeric-literals.js 1:1. Position assertions cover line/column for every invalid case. rslint-specific lock-in cases live in the prefer_numeric_literals_extras_test.go file.
func TestPreferNumericLiteralsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferNumericLiteralsRule,
		[]rule_tester.ValidTestCase{
			{Code: `parseInt(1);`},
			{Code: `parseInt(1, 3);`},
			{Code: `Number.parseInt(1);`},
			{Code: `Number.parseInt(1, 3);`},
			{Code: `0b111110111 === 503;`},
			{Code: `0o767 === 503;`},
			{Code: `0x1F7 === 503;`},
			{Code: `a[parseInt](1,2);`},
			{Code: `parseInt(foo);`},
			{Code: `parseInt(foo, 2);`},
			{Code: `Number.parseInt(foo);`},
			{Code: `Number.parseInt(foo, 2);`},
			{Code: `parseInt(11, 2);`},
			{Code: `Number.parseInt(1, 8);`},
			{Code: `parseInt(1e5, 16);`},
			{Code: `parseInt('11', '2');`},
			{Code: `Number.parseInt('11', '8');`},
			{Code: `parseInt(/foo/, 2);`},
			{Code: "parseInt(`11${foo}`, 2);"},
			{Code: `parseInt('11', 2n);`},
			{Code: `Number.parseInt('11', 8n);`},
			{Code: `parseInt('11', 16n);`},
			{Code: "parseInt(`11`, 16n);"},
			{Code: `parseInt(1n, 2);`},
			{Code: `class C { #parseInt; foo() { Number.#parseInt("111110111", 2); } }`},

			// Shadowed `parseInt` and `Number` should not be reported.
			{Code: `function foo(parseInt) { parseInt("111110111", 2); }`},
			{Code: `function foo() { var parseInt; parseInt("111110111", 2); }`},
			{Code: `function foo(Number) { Number.parseInt("111110111", 2); }`},
			{Code: `function foo() { var Number; Number.parseInt("111110111", 2); }`},
		},
		[]rule_tester.InvalidTestCase{
			preferNumericInvalidFixed(`parseInt("111110111", 2) === 503;`, `parseInt("111110111", 2)`, `0b111110111 === 503;`, "binary", "parseInt"),
			preferNumericInvalidFixed(`parseInt("767", 8) === 503;`, `parseInt("767", 8)`, `0o767 === 503;`, "octal", "parseInt"),
			preferNumericInvalidFixed(`parseInt("1F7", 16) === 255;`, `parseInt("1F7", 16)`, `0x1F7 === 255;`, "hexadecimal", "parseInt"),
			preferNumericInvalidFixed(`Number.parseInt("111110111", 2) === 503;`, `Number.parseInt("111110111", 2)`, `0b111110111 === 503;`, "binary", "Number.parseInt"),
			preferNumericInvalidFixed(`Number.parseInt("767", 8) === 503;`, `Number.parseInt("767", 8)`, `0o767 === 503;`, "octal", "Number.parseInt"),
			preferNumericInvalidFixed(`Number.parseInt("1F7", 16) === 255;`, `Number.parseInt("1F7", 16)`, `0x1F7 === 255;`, "hexadecimal", "Number.parseInt"),
			preferNumericInvalidNoFix(`parseInt('7999', 8);`, `parseInt('7999', 8)`, "octal", "parseInt"),
			preferNumericInvalidNoFix(`parseInt('1234', 2);`, `parseInt('1234', 2)`, "binary", "parseInt"),
			preferNumericInvalidNoFix(`parseInt('1234.5', 8);`, `parseInt('1234.5', 8)`, "octal", "parseInt"),
			preferNumericInvalidNoFix(`parseInt('\u0031\ufe0f\u20e3\u0033\ufe0f\u20e3\u0033\ufe0f\u20e3\u0037\ufe0f\u20e3', 16);`, `parseInt('\u0031\ufe0f\u20e3\u0033\ufe0f\u20e3\u0033\ufe0f\u20e3\u0037\ufe0f\u20e3', 16)`, "hexadecimal", "parseInt"),
			preferNumericInvalidNoFix(`Number.parseInt('7999', 8);`, `Number.parseInt('7999', 8)`, "octal", "Number.parseInt"),
			preferNumericInvalidNoFix(`Number.parseInt('1234', 2);`, `Number.parseInt('1234', 2)`, "binary", "Number.parseInt"),
			preferNumericInvalidNoFix(`Number.parseInt('1234.5', 8);`, `Number.parseInt('1234.5', 8)`, "octal", "Number.parseInt"),
			preferNumericInvalidNoFix(`Number.parseInt('\u0031\ufe0f\u20e3\u0033\ufe0f\u20e3\u0033\ufe0f\u20e3\u0037\ufe0f\u20e3', 16);`, `Number.parseInt('\u0031\ufe0f\u20e3\u0033\ufe0f\u20e3\u0033\ufe0f\u20e3\u0037\ufe0f\u20e3', 16)`, "hexadecimal", "Number.parseInt"),
			preferNumericInvalidFixed("parseInt(`111110111`, 2) === 503;", "parseInt(`111110111`, 2)", `0b111110111 === 503;`, "binary", "parseInt"),
			preferNumericInvalidFixed("parseInt(`767`, 8) === 503;", "parseInt(`767`, 8)", `0o767 === 503;`, "octal", "parseInt"),
			preferNumericInvalidFixed("parseInt(`1F7`, 16) === 255;", "parseInt(`1F7`, 16)", `0x1F7 === 255;`, "hexadecimal", "parseInt"),
			preferNumericInvalidNoFix(`parseInt('', 8);`, `parseInt('', 8)`, "octal", "parseInt"),
			preferNumericInvalidNoFix("parseInt(``, 8);", "parseInt(``, 8)", "octal", "parseInt"),
			preferNumericInvalidNoFix("parseInt(`7999`, 8);", "parseInt(`7999`, 8)", "octal", "parseInt"),
			preferNumericInvalidNoFix("parseInt(`1234`, 2);", "parseInt(`1234`, 2)", "binary", "parseInt"),
			preferNumericInvalidNoFix("parseInt(`1234.5`, 8);", "parseInt(`1234.5`, 8)", "octal", "parseInt"),

			// Adjacent tokens tests
			preferNumericInvalidFixed(`parseInt('11', 2)`, `parseInt('11', 2)`, `0b11`, "binary", "parseInt"),
			preferNumericInvalidFixed(`Number.parseInt('67', 8)`, `Number.parseInt('67', 8)`, `0o67`, "octal", "Number.parseInt"),
			preferNumericInvalidFixed(`5+parseInt('A', 16)`, `parseInt('A', 16)`, `5+0xA`, "hexadecimal", "parseInt"),
			preferNumericInvalidFixed(`function *f(){ yield(Number).parseInt('11', 2) }`, `(Number).parseInt('11', 2)`, `function *f(){ yield 0b11 }`, "binary", "(Number).parseInt"),
			preferNumericInvalidFixed(`function *f(){ yield(Number.parseInt)('67', 8) }`, `(Number.parseInt)('67', 8)`, `function *f(){ yield 0o67 }`, "octal", "Number.parseInt"),
			preferNumericInvalidFixed(`function *f(){ yield(parseInt)('A', 16) }`, `(parseInt)('A', 16)`, `function *f(){ yield 0xA }`, "hexadecimal", "parseInt"),
			preferNumericInvalidFixed(`function *f(){ yield Number.parseInt('11', 2) }`, `Number.parseInt('11', 2)`, `function *f(){ yield 0b11 }`, "binary", "Number.parseInt"),
			preferNumericInvalidFixed(`function *f(){ yield/**/Number.parseInt('67', 8) }`, `Number.parseInt('67', 8)`, `function *f(){ yield/**/0o67 }`, "octal", "Number.parseInt"),
			preferNumericInvalidFixed(`function *f(){ yield(parseInt('A', 16)) }`, `parseInt('A', 16)`, `function *f(){ yield(0xA) }`, "hexadecimal", "parseInt"),
			preferNumericInvalidFixed(`parseInt('11', 2)+5`, `parseInt('11', 2)`, `0b11+5`, "binary", "parseInt"),
			preferNumericInvalidFixed(`Number.parseInt('17', 8)+5`, `Number.parseInt('17', 8)`, `0o17+5`, "octal", "Number.parseInt"),
			preferNumericInvalidFixed(`parseInt('A', 16)+5`, `parseInt('A', 16)`, `0xA+5`, "hexadecimal", "parseInt"),
			preferNumericInvalidFixed(`parseInt('11', 2)in foo`, `parseInt('11', 2)`, `0b11 in foo`, "binary", "parseInt"),
			preferNumericInvalidFixed(`Number.parseInt('17', 8)in foo`, `Number.parseInt('17', 8)`, `0o17 in foo`, "octal", "Number.parseInt"),
			preferNumericInvalidFixed(`parseInt('A', 16)in foo`, `parseInt('A', 16)`, `0xA in foo`, "hexadecimal", "parseInt"),
			preferNumericInvalidFixed(`parseInt('11', 2) in foo`, `parseInt('11', 2)`, `0b11 in foo`, "binary", "parseInt"),
			preferNumericInvalidFixed(`Number.parseInt('17', 8)/**/in foo`, `Number.parseInt('17', 8)`, `0o17/**/in foo`, "octal", "Number.parseInt"),
			preferNumericInvalidFixed(`(parseInt('A', 16))in foo`, `parseInt('A', 16)`, `(0xA)in foo`, "hexadecimal", "parseInt"),

			// Should not autofix if it would remove comments
			preferNumericInvalidFixedNoMessage(`/* comment */Number.parseInt('11', 2);`, `Number.parseInt('11', 2)`, `/* comment */0b11;`),
			preferNumericInvalidNoFixNoMessage(`Number/**/.parseInt('11', 2);`, `Number/**/.parseInt('11', 2)`),
			preferNumericInvalidNoFixNoMessage("Number//\n.parseInt('11', 2);", "Number//\n.parseInt('11', 2)"),
			preferNumericInvalidNoFixNoMessage(`Number./**/parseInt('11', 2);`, `Number./**/parseInt('11', 2)`),
			preferNumericInvalidNoFixNoMessage(`Number.parseInt(/**/'11', 2);`, `Number.parseInt(/**/'11', 2)`),
			preferNumericInvalidNoFixNoMessage(`Number.parseInt('11', /**/2);`, `Number.parseInt('11', /**/2)`),
			preferNumericInvalidFixedNoMessage(`Number.parseInt('11', 2)/* comment */;`, `Number.parseInt('11', 2)`, `0b11/* comment */;`),
			preferNumericInvalidNoFixNoMessage(`parseInt/**/('11', 2);`, `parseInt/**/('11', 2)`),
			preferNumericInvalidNoFixNoMessage("parseInt(//\n'11', 2);", "parseInt(//\n'11', 2)"),
			preferNumericInvalidNoFixNoMessage(`parseInt('11'/**/, 2);`, `parseInt('11'/**/, 2)`),
			preferNumericInvalidNoFixNoMessage("parseInt(`11`/**/, 2);", "parseInt(`11`/**/, 2)"),
			preferNumericInvalidNoFixNoMessage(`parseInt('11', 2 /**/);`, `parseInt('11', 2 /**/)`),
			preferNumericInvalidFixedNoMessage("parseInt('11', 2)//comment\n;", "parseInt('11', 2)", "0b11//comment\n;"),

			// Optional chaining
			preferNumericInvalidFixed(`parseInt?.("1F7", 16) === 255;`, `parseInt?.("1F7", 16)`, `0x1F7 === 255;`, "hexadecimal", "parseInt"),
			preferNumericInvalidFixed(`Number?.parseInt("1F7", 16) === 255;`, `Number?.parseInt("1F7", 16)`, `0x1F7 === 255;`, "hexadecimal", "Number?.parseInt"),
			preferNumericInvalidFixed(`Number?.parseInt?.("1F7", 16) === 255;`, `Number?.parseInt?.("1F7", 16)`, `0x1F7 === 255;`, "hexadecimal", "Number?.parseInt"),
			preferNumericInvalidFixed(`(Number?.parseInt)("1F7", 16) === 255;`, `(Number?.parseInt)("1F7", 16)`, `0x1F7 === 255;`, "hexadecimal", "Number?.parseInt"),
			preferNumericInvalidFixed(`(Number?.parseInt)?.("1F7", 16) === 255;`, `(Number?.parseInt)?.("1F7", 16)`, `0x1F7 === 255;`, "hexadecimal", "Number?.parseInt"),

			// `parseInt` doesn't support numeric separators. The rule shouldn't autofix in those cases.
			preferNumericInvalidNoFix(`parseInt('1_0', 2);`, `parseInt('1_0', 2)`, "binary", "parseInt"),
			preferNumericInvalidNoFix(`Number.parseInt('5_000', 8);`, `Number.parseInt('5_000', 8)`, "octal", "Number.parseInt"),
			preferNumericInvalidNoFix(`parseInt('0_1', 16);`, `parseInt('0_1', 16)`, "hexadecimal", "parseInt"),
			preferNumericInvalidNoFix(`Number.parseInt('0_0', 16);`, `Number.parseInt('0_0', 16)`, "hexadecimal", "Number.parseInt"),
		},
	)
}

func preferNumericInvalidFixed(code string, target string, output string, system string, functionName string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:   code,
		Output: []string{output},
		Errors: []rule_tester.InvalidTestCaseError{
			preferNumericError(code, target, system, functionName, true),
		},
	}
}

func preferNumericInvalidNoFix(code string, target string, system string, functionName string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			preferNumericError(code, target, system, functionName, true),
		},
	}
}

func preferNumericInvalidFixedTwo(
	code string,
	firstTarget string,
	firstSystem string,
	firstFunctionName string,
	secondTarget string,
	secondSystem string,
	secondFunctionName string,
	output string,
) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:   code,
		Output: []string{output},
		Errors: []rule_tester.InvalidTestCaseError{
			preferNumericError(code, firstTarget, firstSystem, firstFunctionName, true),
			preferNumericError(code, secondTarget, secondSystem, secondFunctionName, true),
		},
	}
}

func preferNumericInvalidFixedNoMessage(code string, target string, output string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:   code,
		Output: []string{output},
		Errors: []rule_tester.InvalidTestCaseError{
			preferNumericError(code, target, "", "", false),
		},
	}
}

func preferNumericInvalidNoFixNoMessage(code string, target string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			preferNumericError(code, target, "", "", false),
		},
	}
}

func preferNumericError(code string, target string, system string, functionName string, checkMessage bool) rule_tester.InvalidTestCaseError {
	start := strings.Index(code, target)
	if start < 0 {
		panic("target not found in prefer-numeric-literals test: " + target)
	}
	end := start + len(target)
	line, column := utf16LineColumn(code, start)
	endLine, endColumn := utf16LineColumn(code, end)
	err := rule_tester.InvalidTestCaseError{
		MessageId: "useLiteral",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
	if checkMessage {
		err.Message = "Use " + system + " literals instead of " + functionName + "()."
	}
	return err
}

func utf16LineColumn(text string, offset int) (int, int) {
	line := 1
	column := 1
	for i, r := range text {
		if i >= offset {
			break
		}
		if r == '\n' {
			line++
			column = 1
			continue
		}
		if r >= 0x10000 {
			column += 2
		} else {
			column++
		}
	}
	return line, column
}
