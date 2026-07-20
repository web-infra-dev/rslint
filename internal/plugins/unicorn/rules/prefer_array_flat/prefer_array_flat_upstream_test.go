package prefer_array_flat_test

import (
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const preferArrayFlatMessageID = "prefer-array-flat"

type upstreamSuite struct {
	valid   []rule_tester.ValidTestCase
	invalid []rule_tester.InvalidTestCase
}

type expectedDiagnostic struct {
	target      string
	description string
	occurrence  int
}

func (suite *upstreamSuite) addValid(options any, codes ...string) {
	for _, code := range codes {
		suite.valid = append(suite.valid, rule_tester.ValidTestCase{
			Code:    code,
			Options: options,
		})
	}
}

func (suite *upstreamSuite) addFixed(
	code string,
	target string,
	replacement string,
	description string,
	options any,
) {
	suite.addFixedOutput(
		code,
		strings.Replace(code, target, replacement, 1),
		options,
		expectedDiagnostic{target: target, description: description},
	)
}

func (suite *upstreamSuite) addFixedOutput(
	code string,
	output string,
	options any,
	diagnostics ...expectedDiagnostic,
) {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		errors = append(errors, upstreamError(
			code,
			diagnostic.target,
			diagnostic.description,
			diagnostic.occurrence,
		))
	}
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code:    code,
		Output:  []string{output},
		Errors:  errors,
		Options: options,
	})
}

func (suite *upstreamSuite) addNoFix(
	code string,
	target string,
	description string,
	options any,
) {
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code:    code,
		Errors:  []rule_tester.InvalidTestCaseError{upstreamError(code, target, description, 0)},
		Options: options,
	})
}

func upstreamError(
	code string,
	target string,
	description string,
	occurrence int,
) rule_tester.InvalidTestCaseError {
	start := nthIndex(code, target, occurrence)
	if start < 0 {
		panic("target not found in upstream test: " + target)
	}
	line, column := lineColumn(code, start)
	endLine, endColumn := lineColumn(code, start+len(target))
	return rule_tester.InvalidTestCaseError{
		MessageId: preferArrayFlatMessageID,
		Message: "Prefer `Array#flat()` over `" + description +
			"` to flatten an array.",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func nthIndex(text string, target string, occurrence int) int {
	offset := 0
	for index := 0; index <= occurrence; index++ {
		found := strings.Index(text[offset:], target)
		if found < 0 {
			return -1
		}
		if index == occurrence {
			return offset + found
		}
		offset += found + len(target)
	}
	return -1
}

func lineColumn(text string, offset int) (int, int) {
	line := 1
	lineStart := 0
	for index, character := range text[:offset] {
		if character == '\n' {
			line++
			lineStart = index + 1
		}
	}
	column := len(utf16.Encode([]rune(text[lineStart:offset]))) + 1
	return line, column
}

func (suite *upstreamSuite) run(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_array_flat.PreferArrayFlatRule,
		suite.valid,
		suite.invalid,
	)
}

// TestPreferArrayFlatUpstream migrates the flatMap and reduce groups from
// upstream test/prefer-array-flat.js 1:1. Position assertions cover line and
// column for every invalid case. The remaining upstream groups live in the
// prefer_array_flat_upstream_*_test.go files; rslint-specific lock-ins live in
// prefer_array_flat_extras_test.go.
func TestPreferArrayFlatUpstream(t *testing.T) {
	var suite upstreamSuite

	// ---- `array.flatMap(x => x)` ----
	suite.addValid(nil,
		`array.flatMap`,
		`new array.flatMap(x => x)`,
		`flatMap(x => x)`,
		`array.notFlatMap(x => x)`,
		`array[flatMap](x => x)`,
		`array.flatMap(x => x, thisArgument)`,
		`array.flatMap(...[x => x])`,
		`array.flatMap(function (x) { return x; })`,
		`array.flatMap(async x => x)`,
		`array.flatMap(function * (x) { return x;})`,
		`array.flatMap(() => x)`,
		`array.flatMap((x, y) => x)`,
		`array.flatMap((x) => { return x; })`,
		`array.flatMap(x => y)`,
	)
	suite.addFixed(
		`array.flatMap(x => x)`,
		`array.flatMap(x => x)`,
		`array.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`array?.flatMap(x => x)`,
		`array?.flatMap(x => x)`,
		`array?.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixedOutput(
		`function foo(){return[].flatMap(x => x)}`,
		`function foo(){return [].flat()}`,
		nil,
		expectedDiagnostic{
			target:      `[].flatMap(x => x)`,
			description: `Array#flatMap()`,
		},
	)
	suite.addFixedOutput(
		`foo.flatMap(x => x)instanceof Array`,
		`foo.flat() instanceof Array`,
		nil,
		expectedDiagnostic{
			target:      `foo.flatMap(x => x)`,
			description: `Array#flatMap()`,
		},
	)

	// ---- Obvious non-array flatMap receivers ----
	suite.addValid(nil,
		`const randomObject = {
			flatMap(function_) {
				function_();
			},
		};
		randomObject.flatMap(x => x);`,
		`Effects.flatMap(x => x)`,
		`const effects = {
			flatMap(function_) {
				function_();
			},
		};
		effects.flatMap(x => x);`,
		`const effects = new Set(); effects.flatMap(x => x);`,
		`const mapping = new Map(); mapping.flatMap(x => x);`,
		`const text = ""; text.flatMap(x => x);`,
		`const handler = () => {}; handler.flatMap(x => x);`,
		`const collection = new Foo(); collection.flatMap(x => x);`,
	)
	suite.addFixed(
		`array.flatMap((x) => x)`,
		`array.flatMap((x) => x)`,
		`array.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`Foo.bar.flatMap(x => x)`,
		`Foo.bar.flatMap(x => x)`,
		`Foo.bar.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`const values = getValues(); values.flatMap(x => x);`,
		`values.flatMap(x => x)`,
		`values.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`const values = []; values.flatMap(x => x);`,
		`values.flatMap(x => x)`,
		`values.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`const Items = []; Items.flatMap(x => x);`,
		`Items.flatMap(x => x)`,
		`Items.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`for (const value of values) {
			value.flatMap(x => x);
		}`,
		`value.flatMap(x => x)`,
		`value.flat()`,
		`Array#flatMap()`,
		nil,
	)

	// ---- `array.reduce((a, b) => a.concat(b), [])` ----
	suite.addValid(nil,
		`new array.reduce((a, b) => a.concat(b), [])`,
		`array.reduce`,
		`reduce((a, b) => a.concat(b), [])`,
		`array[reduce]((a, b) => a.concat(b), [])`,
		`array.notReduce((a, b) => a.concat(b), [])`,
		`array.reduce((a, b) => a.concat(b), [], EXTRA_ARGUMENT)`,
		`array.reduce((a, b) => a.concat(b), NOT_EMPTY_ARRAY)`,
		`array.reduce((a, b, extraParameter) => a.concat(b), [])`,
		`array.reduce((a,) => a.concat(b), [])`,
		`array.reduce(() => a.concat(b), [])`,
		`array.reduce((a, b) => {return a.concat(b); }, [])`,
		`array.reduce(function (a, b) { return a.concat(b); }, [])`,
		`array.reduce((a, b) => b.concat(b), [])`,
		`array.reduce((a, b) => a.concat(a), [])`,
		`array.reduce((a, b) => b.concat(a), [])`,
		`array.reduce((a, b) => a.notConcat(b), [])`,
		`array.reduce((a, b) => a.concat, [])`,
	)
	for _, code := range []string{
		`array.reduce((a, b) => a.concat(b), [])`,
		`array?.reduce((a, b) => a.concat(b), [])`,
		`function foo(){return[].reduce((a, b) => a.concat(b), [])}`,
		`function foo(){return[]?.reduce((a, b) => a.concat(b), [])}`,
	} {
		target := code
		output := `array.flat()`
		switch code {
		case `array?.reduce((a, b) => a.concat(b), [])`:
			output = `array?.flat()`
		case `function foo(){return[].reduce((a, b) => a.concat(b), [])}`:
			target = `[].reduce((a, b) => a.concat(b), [])`
			output = `function foo(){return [].flat()}`
		case `function foo(){return[]?.reduce((a, b) => a.concat(b), [])}`:
			target = `[]?.reduce((a, b) => a.concat(b), [])`
			output = `function foo(){return []?.flat()}`
		}
		suite.addFixedOutput(
			code,
			output,
			nil,
			expectedDiagnostic{target: target, description: `Array#reduce()`},
		)
	}

	// ---- `array.reduce((a, b) => [...a, ...b], [])` ----
	suite.addValid(nil,
		`new array.reduce((a, b) => [...a, ...b], [])`,
		`array[reduce]((a, b) => [...a, ...b], [])`,
		`reduce((a, b) => [...a, ...b], [])`,
		`array.notReduce((a, b) => [...a, ...b], [])`,
		`array.reduce((a, b) => [...a, ...b], [], EXTRA_ARGUMENT)`,
		`array.reduce((a, b) => [...a, ...b], NOT_EMPTY_ARRAY)`,
		`array.reduce((a, b, extraParameter) => [...a, ...b], [])`,
		`array.reduce((a,) => [...a, ...b], [])`,
		`array.reduce(() => [...a, ...b], [])`,
		`array.reduce((a, b) => {return [...a, ...b]; }, [])`,
		`array.reduce(function (a, b) { return [...a, ...b]; }, [])`,
		`array.reduce((a, b) => [...b, ...b], [])`,
		`array.reduce((a, b) => [...a, ...a], [])`,
		`array.reduce((a, b) => [...b, ...a], [])`,
		`array.reduce((a, b) => [a, ...b], [])`,
		`array.reduce((a, b) => [...a, b], [])`,
		`array.reduce((a, b) => [a, b], [])`,
		`array.reduce((a, b) => [...a, ...b, c], [])`,
		`array.reduce((a, b) => [...a, ...b,,], [])`,
		`array.reduce((a, b) => [,...a, ...b], [])`,
		`array.reduce((a, b) => [, ], [])`,
		`array.reduce((a, b) => [, ,], [])`,
	)
	suite.addFixed(
		`array.reduce((a, b) => [...a, ...b], [])`,
		`array.reduce((a, b) => [...a, ...b], [])`,
		`array.flat()`,
		`Array#reduce()`,
		nil,
	)
	suite.addFixed(
		`array.reduce((a, b) => [...a, ...b,], [])`,
		`array.reduce((a, b) => [...a, ...b,], [])`,
		`array.flat()`,
		`Array#reduce()`,
		nil,
	)
	suite.addFixedOutput(
		`function foo(){return[].reduce((a, b) => [...a, ...b,], [])}`,
		`function foo(){return [].flat()}`,
		nil,
		expectedDiagnostic{
			target:      `[].reduce((a, b) => [...a, ...b,], [])`,
			description: `Array#reduce()`,
		},
	)

	suite.run(t)
}
