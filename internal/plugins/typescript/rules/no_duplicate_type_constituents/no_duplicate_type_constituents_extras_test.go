package no_duplicate_type_constituents

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional AST, parameter, comment and fix boundaries checked against v8.72.2.
func TestNoDuplicateTypeConstituentsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateTypeConstituentsRule, []rule_tester.ValidTestCase{
		{Code: "type T = { a: 1 } | { a: 0x1 };"},
		{Code: "type T = { a: 'a' } | { a: \"a\" };"},
		{Code: "type T = { readonly a: string } | { a: string };"},
		{Code: "type T = { a?: string } | { a: string };"},
		{Code: "type T = ((x: string) => void) | ((y: string) => void);"},
		{Code: "type T = (new () => string) | (() => string);"},
		{Code: "type T = (readonly string[]) | string[];"},
		{Code: "type T = { [K in 'a']?: string } | { [K in 'a']: string };"},
		{Code: "type T = { [K in 'a']-?: string } | { [K in 'a']+?: string };"},
		{Code: "type T = ((x: unknown) => x is string) | ((x: unknown) => asserts x is string);"},
		{Code: "type T = [a?: string] | [a: string];"},
		{Code: "type T = [a: string] | [b: string];"},
		{Code: "function f({x}?: {x: string} | undefined) {}"},
		{Code: "class C { constructor(public x?: string | undefined) {} }"},
		{Code: "interface I { x?: string | undefined }"},
		{Code: "type T = Missing | Missing;"},
		{Code: "type T = { a: string } | { a: string };", Options: map[string]any{"ignoreUnions": true}},
		{Code: "type T = { a: string } & { a: string };", Options: map[string]any{"ignoreIntersections": true}},
		{Code: "type T = { [K in 'a']: K } | { [P in 'a']: P };"},
		{Code: "type T = { a: keyof string } | { a: readonly string[] };"},
		{Code: "type T = { a: -1 } | { a: 1 };"},
		{Code: "type T = { a: `a` } | { a: `b` };"},
		{Code: "class C { constructor(readonly x?: string | undefined) {} }"},
		{Code: "type T = { a: import(\"missing\").A } | { a: typeof import(\"missing\").A };"},
		{Code: "type T = { a: number } | { a: string };"},
		{Code: "interface I { [x: string | undefined]: unknown }"},
		{Code: "type T = { a: `${string}a` } | { a: `${string}b` };"},
		{Code: "type T = { readonly [K in \"a\"]: string } | { +readonly [K in \"a\"]: string };"},
		{Code: "type T = [string?] | [string | undefined];"},
		{Code: "type T = (<T extends string>() => T) | (<T = string>() => T);"},
	}, []rule_tester.InvalidTestCase{
		{Code: "type T = { a: string } | { a: (string) };", Output: []string{"type T = { a: string }  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { a: string }.", Line: 1, Column: 26, EndLine: 1, EndColumn: 41},
		}},
		{Code: "type T = { a: string; } | { a: string };", Output: []string{"type T = { a: string; }  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { a: string; }.", Line: 1, Column: 27, EndLine: 1, EndColumn: 40},
		}},
		{Code: "type T = (() => string) | (() => string);", Output: []string{"type T = (() => string)  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with () => string.", Line: 1, Column: 28, EndLine: 1, EndColumn: 41},
		}},
		{Code: "type T = { a: string } | /*keep*/ (( /*before*/ { a: string } /*after*/ ));", Output: []string{"type T = { a: string }  /*keep*/  /*before*/  /*after*/ ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { a: string }.", Line: 1, Column: 49, EndLine: 1, EndColumn: 75},
		}},
		{Code: "type T = string | ((string));", Output: []string{"type T = string  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 1, Column: 21, EndLine: 1, EndColumn: 29},
		}},
		{Code: "type T = string | (( /*comment*/ string /*comment*/ ));", Output: []string{"type T = string   /*comment*/  /*comment*/ ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 1, Column: 34, EndLine: 1, EndColumn: 55},
		}},
		{Code: "type T = number & string & (number & string);", Output: []string{"type T = number & string & (  string);", "type T = number & string    ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with number.", Line: 1, Column: 29, EndLine: 1, EndColumn: 35},
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with string.", Line: 1, Column: 38, EndLine: 1, EndColumn: 44},
		}},
		{Code: "function f(x?: ((undefined)) | string) {}", Output: []string{"function f(x?:   string) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 18, EndLine: 1, EndColumn: 29},
		}},
		{Code: "function f(x?: | ((undefined)) | string) {}", Output: []string{"function f(x?:   | string) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 20, EndLine: 1, EndColumn: 31},
		}},
		{Code: "function f(x?: (string | undefined)) {}", Output: []string{"function f(x?: (string  )) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 26, EndLine: 1, EndColumn: 35},
		}},
		{Code: "type T = { missing: Missing } | { missing: Missing };", Output: []string{"type T = { missing: Missing }  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { missing: Missing }.", Line: 1, Column: 33, EndLine: 1, EndColumn: 53},
		}},
		{Code: "type T = { a: `a` } | { a: `a` };", Output: []string{"type T = { a: `a` }  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { a: `a` }.", Line: 1, Column: 23, EndLine: 1, EndColumn: 33},
		}},
		{Code: "type T = { a: string } | { a: /*comment*/ string };", Output: []string{"type T = { a: string }  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { a: string }.", Line: 1, Column: 26, EndLine: 1, EndColumn: 51},
		}},
		{Code: "class C { constructor(x?: string | undefined) {} }", Output: []string{"class C { constructor(x?: string  ) {} }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 36, EndLine: 1, EndColumn: 45},
		}},
		{Code: "function f(x?: | ((undefined)) | string) {}", Output: []string{"function f(x?:   | string) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 20, EndLine: 1, EndColumn: 31},
		}},
		{Code: "function f(x?: ((undefined) | string)) {}", Output: []string{"function f(x?: (  string)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28},
		}},
		{Code: "function f(x?: undefined | (undefined | string)) {}", Output: []string{"function f(x?:   (  string)) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with undefined.", Line: 1, Column: 29, EndLine: 1, EndColumn: 38},
		}},
		{Code: "function f(x?: ((undefined | string)) | undefined) {}", Output: []string{"function f(x?: ((  string))  ) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 18, EndLine: 1, EndColumn: 27},
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with undefined.", Line: 1, Column: 41, EndLine: 1, EndColumn: 50},
		}},
		{Code: "function f(x?: undefined | undefined) {}", Output: []string{"function f(x?:   undefined) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with undefined.", Line: 1, Column: 28, EndLine: 1, EndColumn: 37},
		}},
		{Code: "type T = { a: string } | { \\u0061: string };", Output: []string{"type T = { a: string }  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { a: string }.", Line: 1, Column: 26, EndLine: 1, EndColumn: 44},
		}},
		{Code: "type T = { a: string } | { a: string, };", Output: []string{"type T = { a: string }  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { a: string }.", Line: 1, Column: 26, EndLine: 1, EndColumn: 40},
		}},
		{Code: "type T = (() => never) & (() => never);", Output: []string{"type T = (() => never)  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with () => never.", Line: 1, Column: 27, EndLine: 1, EndColumn: 39},
		}},
		{Code: "type T = { a: `${string}a` } | { a: `${string}a` };", Output: []string{"type T = { a: `${string}a` }  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { a: `${string}a` }.", Line: 1, Column: 32, EndLine: 1, EndColumn: 51},
		}},
		{Code: "type T = [string?] | [string?];", Output: []string{"type T = [string?]  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with [string?].", Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
		}},
		{Code: "type T = ((<T>() => T)) | (<T>() => T);", Output: []string{"type T = ((<T>() => T))  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with <T>() => T.", Line: 1, Column: 28, EndLine: 1, EndColumn: 39},
		}},
	})
}
