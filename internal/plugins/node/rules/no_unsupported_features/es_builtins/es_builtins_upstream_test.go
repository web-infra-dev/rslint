package es_builtins_test

import (
	"maps"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_unsupported_features/es_builtins"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func unsupportedError(name, supported, version string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "not-supported-till",
		Message:   "The '" + name + "' is still an experimental feature and is not supported until Node.js " + supported + ". The configured version range is '" + version + "'.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

func runESBuiltinsTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &es_builtins.ESBuiltinsRule, valid, invalid)
}

// All 63 groups (161 valid, 92 invalid and 92 ignore variants) from:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unsupported-features/es-builtins.js
// Ranges and messages were checked with that release, including shadowed globals.
func TestESBuiltinsUpstream(t *testing.T) {
	for _, group := range []struct {
		keyword string
		valid   []rule_tester.ValidTestCase
		invalid []rule_tester.InvalidTestCase
	}{
		{keyword: "AggregateError",
			valid: []rule_tester.ValidTestCase{
				{Code: "if (error instanceof AggregateError) {}", Options: map[string]any{"version": "15.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "if (error instanceof AggregateError) {}", Options: map[string]any{"version": "14.0.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("AggregateError", "15.0.0", "14.0.0", 1, 22, 1, 36)}},
			},
		},
		{keyword: "Array.from",
			valid: []rule_tester.ValidTestCase{
				{Code: "Array.foo(a)", Options: map[string]any{"version": "3.9.9"}},
				{Code: "(function(Array) { Array.from(a) }(b))", Options: map[string]any{"version": "3.9.9"}},
				{Code: "Array.from(a)", Options: map[string]any{"version": "4.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Array.from(a)", Options: map[string]any{"version": "3.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Array.from", "4.0.0", "3.9.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Array.of",
			valid: []rule_tester.ValidTestCase{
				{Code: "Array.foo(a)", Options: map[string]any{"version": "3.9.9"}},
				{Code: "(function(Array) { Array.of(a) }(b))", Options: map[string]any{"version": "3.9.9"}},
				{Code: "Array.of(a)", Options: map[string]any{"version": "4.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Array.of(a)", Options: map[string]any{"version": "3.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Array.of", "4.0.0", "3.9.9", 1, 1, 1, 9)}},
			},
		},
		{keyword: "BigInt",
			valid: []rule_tester.ValidTestCase{
				{Code: "bigint", Options: map[string]any{"version": "10.3.0"}},
				{Code: "(function(BigInt) { BigInt }(b))", Options: map[string]any{"version": "10.3.0"}},
				{Code: "BigInt", Options: map[string]any{"version": "10.4.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "BigInt", Options: map[string]any{"version": "10.3.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("BigInt", "10.4.0", "10.3.0", 1, 1, 1, 7)}},
				{Code: "(function() { BigInt })()", Options: map[string]any{"version": "10.3.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("BigInt", "10.4.0", "10.3.0", 1, 15, 1, 21)}},
			},
		},
		{keyword: "FinalizationRegistry",
			valid: []rule_tester.ValidTestCase{
				{Code: "new FinalizationRegistry(() => {})", Options: map[string]any{"version": "14.6.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "new FinalizationRegistry(() => {})", Options: map[string]any{"version": "14.5.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("FinalizationRegistry", "14.6.0", "14.5.0", 1, 5, 1, 25)}},
			},
		},
		{keyword: "Map",
			valid: []rule_tester.ValidTestCase{
				{Code: "map", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Map) { Map }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Map", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Map", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Map", "0.12.0", "0.11.9", 1, 1, 1, 4)}},
				{Code: "(function() { Map })()", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Map", "0.12.0", "0.11.9", 1, 15, 1, 18)}},
			},
		},
		{keyword: "Math.acosh",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.acosh(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.acosh(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.acosh(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.acosh", "0.12.0", "0.11.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Math.asinh",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.asinh(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.asinh(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.asinh(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.asinh", "0.12.0", "0.11.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Math.atanh",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.atanh(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.atanh(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.atanh(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.atanh", "0.12.0", "0.11.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Math.cbrt",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.cbrt(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.cbrt(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.cbrt(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.cbrt", "0.12.0", "0.11.9", 1, 1, 1, 10)}},
			},
		},
		{keyword: "Math.clz32",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.clz32(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.clz32(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.clz32(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.clz32", "0.12.0", "0.11.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Math.cosh",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.cosh(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.cosh(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.cosh(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.cosh", "0.12.0", "0.11.9", 1, 1, 1, 10)}},
			},
		},
		{keyword: "Math.expm1",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.expm1(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.expm1(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.expm1(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.expm1", "0.12.0", "0.11.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Math.fround",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.fround(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.fround(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.fround(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.fround", "0.12.0", "0.11.9", 1, 1, 1, 12)}},
			},
		},
		{keyword: "Math.hypot",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.hypot(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.hypot(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.hypot(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.hypot", "0.12.0", "0.11.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Math.imul",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.imul(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.imul(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.imul(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.imul", "0.12.0", "0.11.9", 1, 1, 1, 10)}},
			},
		},
		{keyword: "Math.log10",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.log10(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.log10(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.log10(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.log10", "0.12.0", "0.11.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Math.log1p",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.log1p(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.log1p(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.log1p(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.log1p", "0.12.0", "0.11.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Math.log2",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.log2(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.log2(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.log2(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.log2", "0.12.0", "0.11.9", 1, 1, 1, 10)}},
			},
		},
		{keyword: "Math.sign",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.sign(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.sign(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.sign(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.sign", "0.12.0", "0.11.9", 1, 1, 1, 10)}},
			},
		},
		{keyword: "Math.sinh",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.sinh(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.sinh(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.sinh(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.sinh", "0.12.0", "0.11.9", 1, 1, 1, 10)}},
			},
		},
		{keyword: "Math.tanh",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.tanh(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.tanh(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.tanh(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.tanh", "0.12.0", "0.11.9", 1, 1, 1, 10)}},
			},
		},
		{keyword: "Math.trunc",
			valid: []rule_tester.ValidTestCase{
				{Code: "Math.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Math) { Math.trunc(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Math.trunc(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Math.trunc(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Math.trunc", "0.12.0", "0.11.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Number.isFinite",
			valid: []rule_tester.ValidTestCase{
				{Code: "Number.foo(a)", Options: map[string]any{"version": "0.9.9"}},
				{Code: "(function(Number) { Number.isFinite(a) }(b))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Number.isFinite(a)", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Number.isFinite(a)", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Number.isFinite", "0.10.0", "0.9.9", 1, 1, 1, 16)}},
			},
		},
		{keyword: "Number.isInteger",
			valid: []rule_tester.ValidTestCase{
				{Code: "Number.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Number) { Number.isInteger(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Number.isInteger(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Number.isInteger(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Number.isInteger", "0.12.0", "0.11.9", 1, 1, 1, 17)}},
			},
		},
		{keyword: "Number.isNaN",
			valid: []rule_tester.ValidTestCase{
				{Code: "Number.foo(a)", Options: map[string]any{"version": "0.9.9"}},
				{Code: "(function(Number) { Number.isNaN(a) }(b))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Number.isNaN(a)", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Number.isNaN(a)", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Number.isNaN", "0.10.0", "0.9.9", 1, 1, 1, 13)}},
			},
		},
		{keyword: "Number.isSafeInteger",
			valid: []rule_tester.ValidTestCase{
				{Code: "Number.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Number) { Number.isSafeInteger(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Number.isSafeInteger(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Number.isSafeInteger(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Number.isSafeInteger", "0.12.0", "0.11.9", 1, 1, 1, 21)}},
			},
		},
		{keyword: "Number.parseFloat",
			valid: []rule_tester.ValidTestCase{
				{Code: "Number.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Number) { Number.parseFloat(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Number.parseFloat(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Number.parseFloat(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Number.parseFloat", "0.12.0", "0.11.9", 1, 1, 1, 18)}},
			},
		},
		{keyword: "Number.parseInt",
			valid: []rule_tester.ValidTestCase{
				{Code: "Number.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Number) { Number.parseInt(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Number.parseInt(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Number.parseInt(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Number.parseInt", "0.12.0", "0.11.9", 1, 1, 1, 16)}},
			},
		},
		{keyword: "Object.assign",
			valid: []rule_tester.ValidTestCase{
				{Code: "Object.foo(a)", Options: map[string]any{"version": "3.9.9"}},
				{Code: "(function(Object) { Object.assign(a) }(b))", Options: map[string]any{"version": "3.9.9"}},
				{Code: "Object.assign(a)", Options: map[string]any{"version": "4.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Object.assign(a)", Options: map[string]any{"version": "3.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Object.assign", "4.0.0", "3.9.9", 1, 1, 1, 14)}},
			},
		},
		{keyword: "Object.getOwnPropertySymbols",
			valid: []rule_tester.ValidTestCase{
				{Code: "Object.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Object) { Object.getOwnPropertySymbols(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Object.getOwnPropertySymbols(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Object.getOwnPropertySymbols(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Object.getOwnPropertySymbols", "0.12.0", "0.11.9", 1, 1, 1, 29)}},
			},
		},
		{keyword: "Object.is",
			valid: []rule_tester.ValidTestCase{
				{Code: "Object.foo(a)", Options: map[string]any{"version": "0.9.9"}},
				{Code: "(function(Object) { Object.is(a) }(b))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Object.is(a)", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Object.is(a)", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Object.is", "0.10.0", "0.9.9", 1, 1, 1, 10)}},
			},
		},
		{keyword: "Object.setPrototypeOf",
			valid: []rule_tester.ValidTestCase{
				{Code: "Object.foo(a)", Options: map[string]any{"version": "0.11.9"}},
				{Code: "(function(Object) { Object.setPrototypeOf(a) }(b))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Object.setPrototypeOf(a)", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Object.setPrototypeOf(a)", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Object.setPrototypeOf", "0.12.0", "0.11.9", 1, 1, 1, 22)}},
			},
		},
		{keyword: "Promise",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Promise) { Promise }(a))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Promise", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Promise", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Promise", "0.12.0", "0.11.9", 1, 1, 1, 8)}},
				{Code: "function wrap() { Promise }", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Promise", "0.12.0", "0.11.9", 1, 19, 1, 26)}},
			},
		},
		{keyword: "Promise.allSettled",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Promise) { Promise.allSettled }(a))", Options: map[string]any{"version": "12.8.1"}},
				{Code: "Promise.allSettled", Options: map[string]any{"version": "12.9.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Promise.allSettled", Options: map[string]any{"version": "12.8.1"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Promise.allSettled", "12.9.0", "12.8.1", 1, 1, 1, 19)}},
				{Code: "function wrap() { Promise.allSettled }", Options: map[string]any{"version": "12.8.1"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Promise.allSettled", "12.9.0", "12.8.1", 1, 19, 1, 37)}},
			},
		},
		{keyword: "Promise.any",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Promise) { Promise.any }(a))", Options: map[string]any{"version": "14.0.0"}},
				{Code: "Promise.any", Options: map[string]any{"version": "15.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Promise.any", Options: map[string]any{"version": "14.0.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 1, 1, 12)}},
				{Code: "function wrap() { Promise.any }", Options: map[string]any{"version": "14.0.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Promise.any", "15.0.0", "14.0.0", 1, 19, 1, 30)}},
			},
		},
		{keyword: "Proxy",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Proxy) { Proxy }(a))", Options: map[string]any{"version": "5.9.9"}},
				{Code: "Proxy", Options: map[string]any{"version": "6.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Proxy", Options: map[string]any{"version": "5.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Proxy", "6.0.0", "5.9.9", 1, 1, 1, 6)}},
				{Code: "function wrap() { Proxy }", Options: map[string]any{"version": "5.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Proxy", "6.0.0", "5.9.9", 1, 19, 1, 24)}},
			},
		},
		{keyword: "Reflect",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Reflect) { Reflect }(a))", Options: map[string]any{"version": "5.9.9"}},
				{Code: "Reflect", Options: map[string]any{"version": "6.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Reflect", Options: map[string]any{"version": "5.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Reflect", "6.0.0", "5.9.9", 1, 1, 1, 8)}},
				{Code: "function wrap() { Reflect }", Options: map[string]any{"version": "5.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Reflect", "6.0.0", "5.9.9", 1, 19, 1, 26)}},
			},
		},
		{keyword: "Set",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Set) { Set }(a))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Set", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Set", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Set", "0.12.0", "0.11.9", 1, 1, 1, 4)}},
				{Code: "function wrap() { Set }", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Set", "0.12.0", "0.11.9", 1, 19, 1, 22)}},
			},
		},
		{keyword: "String.fromCodePoint",
			valid: []rule_tester.ValidTestCase{
				{Code: "String.foo(a)", Options: map[string]any{"version": "3.9.9"}},
				{Code: "(function(String) { String.fromCodePoint(a) }(b))", Options: map[string]any{"version": "3.9.9"}},
				{Code: "String.fromCodePoint(a)", Options: map[string]any{"version": "4.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "String.fromCodePoint(a)", Options: map[string]any{"version": "3.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("String.fromCodePoint", "4.0.0", "3.9.9", 1, 1, 1, 21)}},
			},
		},
		{keyword: "String.raw",
			valid: []rule_tester.ValidTestCase{
				{Code: "String.foo(a)", Options: map[string]any{"version": "3.9.9"}},
				{Code: "(function(String) { String.raw(a) }(b))", Options: map[string]any{"version": "3.9.9"}},
				{Code: "String.raw(a)", Options: map[string]any{"version": "4.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "String.raw(a)", Options: map[string]any{"version": "3.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("String.raw", "4.0.0", "3.9.9", 1, 1, 1, 11)}},
			},
		},
		{keyword: "Symbol",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Symbol) { Symbol }(a))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "Symbol", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Symbol", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Symbol", "0.12.0", "0.11.9", 1, 1, 1, 7)}},
				{Code: "function wrap() { Symbol }", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Symbol", "0.12.0", "0.11.9", 1, 19, 1, 25)}},
			},
		},
		{keyword: "Int8Array",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Int8Array) { Int8Array }(a))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Int8Array", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Int8Array", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Int8Array", "0.10.0", "0.9.9", 1, 1, 1, 10)}},
				{Code: "function wrap() { Int8Array }", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Int8Array", "0.10.0", "0.9.9", 1, 19, 1, 28)}},
			},
		},
		{keyword: "Uint8Array",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Uint8Array) { Uint8Array }(a))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Uint8Array", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Uint8Array", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Uint8Array", "0.10.0", "0.9.9", 1, 1, 1, 11)}},
				{Code: "function wrap() { Uint8Array }", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Uint8Array", "0.10.0", "0.9.9", 1, 19, 1, 29)}},
			},
		},
		{keyword: "Uint8ClampedArray",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Uint8ClampedArray) { Uint8ClampedArray }(a))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Uint8ClampedArray", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Uint8ClampedArray", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Uint8ClampedArray", "0.10.0", "0.9.9", 1, 1, 1, 18)}},
				{Code: "function wrap() { Uint8ClampedArray }", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Uint8ClampedArray", "0.10.0", "0.9.9", 1, 19, 1, 36)}},
			},
		},
		{keyword: "Int16Array",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Int16Array) { Int16Array }(a))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Int16Array", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Int16Array", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Int16Array", "0.10.0", "0.9.9", 1, 1, 1, 11)}},
				{Code: "function wrap() { Int16Array }", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Int16Array", "0.10.0", "0.9.9", 1, 19, 1, 29)}},
			},
		},
		{keyword: "Uint16Array",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Uint16Array) { Uint16Array }(a))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Uint16Array", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Uint16Array", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Uint16Array", "0.10.0", "0.9.9", 1, 1, 1, 12)}},
				{Code: "function wrap() { Uint16Array }", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Uint16Array", "0.10.0", "0.9.9", 1, 19, 1, 30)}},
			},
		},
		{keyword: "Int32Array",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Int32Array) { Int32Array }(a))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Int32Array", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Int32Array", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Int32Array", "0.10.0", "0.9.9", 1, 1, 1, 11)}},
				{Code: "function wrap() { Int32Array }", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Int32Array", "0.10.0", "0.9.9", 1, 19, 1, 29)}},
			},
		},
		{keyword: "Uint32Array",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Uint32Array) { Uint32Array }(a))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Uint32Array", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Uint32Array", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Uint32Array", "0.10.0", "0.9.9", 1, 1, 1, 12)}},
				{Code: "function wrap() { Uint32Array }", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Uint32Array", "0.10.0", "0.9.9", 1, 19, 1, 30)}},
			},
		},
		{keyword: "BigInt64Array",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(BigInt64Array) { BigInt64Array }(b))", Options: map[string]any{"version": "10.3.0"}},
				{Code: "BigInt64Array", Options: map[string]any{"version": "10.4.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "BigInt64Array", Options: map[string]any{"version": "10.3.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("BigInt64Array", "10.4.0", "10.3.0", 1, 1, 1, 14)}},
				{Code: "(function() { BigInt64Array })()", Options: map[string]any{"version": "10.3.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("BigInt64Array", "10.4.0", "10.3.0", 1, 15, 1, 28)}},
			},
		},
		{keyword: "BigUint64Array",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(BigUint64Array) { BigUint64Array }(b))", Options: map[string]any{"version": "10.3.0"}},
				{Code: "BigUint64Array", Options: map[string]any{"version": "10.4.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "BigUint64Array", Options: map[string]any{"version": "10.3.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("BigUint64Array", "10.4.0", "10.3.0", 1, 1, 1, 15)}},
				{Code: "(function() { BigUint64Array })()", Options: map[string]any{"version": "10.3.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("BigUint64Array", "10.4.0", "10.3.0", 1, 15, 1, 29)}},
			},
		},
		{keyword: "Float32Array",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Float32Array) { Float32Array }(a))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Float32Array", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Float32Array", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Float32Array", "0.10.0", "0.9.9", 1, 1, 1, 13)}},
				{Code: "function wrap() { Float32Array }", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Float32Array", "0.10.0", "0.9.9", 1, 19, 1, 31)}},
			},
		},
		{keyword: "Float64Array",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Float64Array) { Float64Array }(a))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "Float64Array", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Float64Array", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Float64Array", "0.10.0", "0.9.9", 1, 1, 1, 13)}},
				{Code: "function wrap() { Float64Array }", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Float64Array", "0.10.0", "0.9.9", 1, 19, 1, 31)}},
			},
		},
		{keyword: "DataView",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(DataView) { DataView }(a))", Options: map[string]any{"version": "0.9.9"}},
				{Code: "DataView", Options: map[string]any{"version": "0.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "DataView", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("DataView", "0.10.0", "0.9.9", 1, 1, 1, 9)}},
				{Code: "function wrap() { DataView }", Options: map[string]any{"version": "0.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("DataView", "0.10.0", "0.9.9", 1, 19, 1, 27)}},
			},
		},
		{keyword: "WeakMap",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(WeakMap) { WeakMap }(a))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "WeakMap", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "WeakMap", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("WeakMap", "0.12.0", "0.11.9", 1, 1, 1, 8)}},
				{Code: "function wrap() { WeakMap }", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("WeakMap", "0.12.0", "0.11.9", 1, 19, 1, 26)}},
			},
		},
		{keyword: "WeakRef",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(WeakRef) { WeakRef }(a))", Options: map[string]any{"version": "14.5.0"}},
				{Code: "WeakRef", Options: map[string]any{"version": "14.6.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "WeakRef", Options: map[string]any{"version": "14.5.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("WeakRef", "14.6.0", "14.5.0", 1, 1, 1, 8)}},
				{Code: "function wrap() { WeakRef }", Options: map[string]any{"version": "14.5.0"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("WeakRef", "14.6.0", "14.5.0", 1, 19, 1, 26)}},
			},
		},
		{keyword: "WeakSet",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(WeakSet) { WeakSet }(a))", Options: map[string]any{"version": "0.11.9"}},
				{Code: "WeakSet", Options: map[string]any{"version": "0.12.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "WeakSet", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("WeakSet", "0.12.0", "0.11.9", 1, 1, 1, 8)}},
				{Code: "function wrap() { WeakSet }", Options: map[string]any{"version": "0.11.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("WeakSet", "0.12.0", "0.11.9", 1, 19, 1, 26)}},
			},
		},
		{keyword: "Atomics",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(Atomics) { Atomics }(a))", Options: map[string]any{"version": "8.9.9"}},
				{Code: "Atomics", Options: map[string]any{"version": "8.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Atomics", Options: map[string]any{"version": "8.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Atomics", "8.10.0", "8.9.9", 1, 1, 1, 8)}},
				{Code: "function wrap() { Atomics }", Options: map[string]any{"version": "8.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Atomics", "8.10.0", "8.9.9", 1, 19, 1, 26)}},
			},
		},
		{keyword: "Object.values",
			valid: []rule_tester.ValidTestCase{
				{Code: "Object.foo(a)", Options: map[string]any{"version": "6.9.9"}},
				{Code: "(function(Object) { Object.values(a) }(b))", Options: map[string]any{"version": "6.9.9"}},
				{Code: "Object.values(a)", Options: map[string]any{"version": "7.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Object.values(a)", Options: map[string]any{"version": "6.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Object.values", "7.0.0", "6.9.9", 1, 1, 1, 14)}},
			},
		},
		{keyword: "Object.entries",
			valid: []rule_tester.ValidTestCase{
				{Code: "Object.foo(a)", Options: map[string]any{"version": "6.9.9"}},
				{Code: "(function(Object) { Object.entries(a) }(b))", Options: map[string]any{"version": "6.9.9"}},
				{Code: "Object.entries(a)", Options: map[string]any{"version": "7.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Object.entries(a)", Options: map[string]any{"version": "6.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Object.entries", "7.0.0", "6.9.9", 1, 1, 1, 15)}},
			},
		},
		{keyword: "Object.getOwnPropertyDescriptors",
			valid: []rule_tester.ValidTestCase{
				{Code: "Object.foo(a)", Options: map[string]any{"version": "6.9.9"}},
				{Code: "(function(Object) { Object.getOwnPropertyDescriptors(a) }(b))", Options: map[string]any{"version": "6.9.9"}},
				{Code: "Object.getOwnPropertyDescriptors(a)", Options: map[string]any{"version": "7.0.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "Object.getOwnPropertyDescriptors(a)", Options: map[string]any{"version": "6.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Object.getOwnPropertyDescriptors", "7.0.0", "6.9.9", 1, 1, 1, 33)}},
			},
		},
		{keyword: "SharedArrayBuffer",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(SharedArrayBuffer) { SharedArrayBuffer }(a))", Options: map[string]any{"version": "8.9.9"}},
				{Code: "SharedArrayBuffer", Options: map[string]any{"version": "8.10.0"}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "SharedArrayBuffer", Options: map[string]any{"version": "8.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("SharedArrayBuffer", "8.10.0", "8.9.9", 1, 1, 1, 18)}},
				{Code: "function wrap() { SharedArrayBuffer }", Options: map[string]any{"version": "8.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("SharedArrayBuffer", "8.10.0", "8.9.9", 1, 19, 1, 36)}},
			},
		},
		{keyword: "globalThis",
			valid: []rule_tester.ValidTestCase{
				{Code: "(function(globalThis) { globalThis }(a))", Options: map[string]any{"version": "12.0.0"}},
				{Code: "globalThis", Options: map[string]any{"version": "12.0.0"}},
				{Code: "globalThis", Settings: map[string]any{"node": map[string]any{"version": "12.0.0"}}},
			},
			invalid: []rule_tester.InvalidTestCase{
				{Code: "globalThis", Options: map[string]any{"version": "11.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("globalThis", "12.0.0", "11.9.9", 1, 1, 1, 11)}},
				{Code: "function wrap() { globalThis }", Options: map[string]any{"version": "11.9.9"}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("globalThis", "12.0.0", "11.9.9", 1, 19, 1, 29)}},
				{Code: "function wrap() { globalThis }", Settings: map[string]any{"node": map[string]any{"version": "11.9.9"}}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("globalThis", "12.0.0", "11.9.9", 1, 19, 1, 29)}},
				{Code: "function wrap() { globalThis }", Settings: map[string]any{"node": map[string]any{"version": ">=11.9.9"}}, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("globalThis", "12.0.0", ">=11.9.9", 1, 19, 1, 29)}},
			},
		},
	} {
		t.Run(group.keyword, func(t *testing.T) {
			// Upstream repeats every invalid case with its exact API name ignored.
			for _, item := range group.invalid {
				options, _ := item.Options.(map[string]any)
				options = maps.Clone(options)
				if options == nil {
					options = map[string]any{}
				}
				options["ignores"] = []any{group.keyword}
				group.valid = append(group.valid, rule_tester.ValidTestCase{Code: item.Code, Options: options, Settings: item.Settings})
			}
			runESBuiltinsTests(t, group.valid, group.invalid)
		})
	}
}

// Upstream documentation contains two configuration examples and no JavaScript
// snippets. Exercise both examples on supported and unsupported builtins.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-unsupported-features/es-builtins.md
func TestESBuiltinsDocumentation(t *testing.T) {
	for _, source := range []string{"options", "settings"} {
		t.Run(source, func(t *testing.T) {
			options := map[string]any{"ignores": []any{}}
			var settings map[string]any
			if source == "options" {
				options["version"] = ">=16.0.0"
			} else {
				settings = map[string]any{"node": map[string]any{"version": ">=16.0.0"}}
			}
			runESBuiltinsTests(t, []rule_tester.ValidTestCase{
				{Code: "Object.fromEntries(entries)", Options: options, Settings: settings},
			}, []rule_tester.InvalidTestCase{
				{Code: "Map.groupBy(items, key)", Options: options, Settings: settings, Errors: []rule_tester.InvalidTestCaseError{unsupportedError("Map.groupBy", "21.0.0", ">=16.0.0", 1, 1, 1, 12)}},
			})
		})
	}
}
