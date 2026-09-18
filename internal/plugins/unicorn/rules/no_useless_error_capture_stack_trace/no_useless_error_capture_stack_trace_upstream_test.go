// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-useless-error-capture-stack-trace.js
package no_useless_error_capture_stack_trace_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_error_capture_stack_trace"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessErrorCaptureStackTraceUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_error_capture_stack_trace.NoUselessErrorCaptureStackTraceRule, []rule_tester.ValidTestCase{
		{Code: "class MyError {constructor() {Error.captureStackTrace(this, MyError)}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends NotABuiltinError {constructor() {Error.captureStackTrace(this, MyError)}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(not_this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, NotClassName)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError, ...extraArguments)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(..._, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, ..._)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(...[this, MyError])\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tNotError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.not_captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tnew Error.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError?.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, this?.constructor)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, this.notConstructor)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, import.meta)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tfunction foo() {\n\tError.captureStackTrace(this, MyError)\n}\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tnotConstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tfunction foo() {\n\t\t\tError.captureStackTrace(this, MyError)\n\t\t}\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor(MyError) {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tstatic {\n\t\tError.captureStackTrace(this, MyError)\n\n\t\tfunction foo() {\n\t\t\tError.captureStackTrace(this, MyError)\n\t\t}\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tclass NotAErrorSubclass {\n\t\t\tconstructor() {\n\t\t\t\tError.captureStackTrace(this, new.target)\n\t\t\t}\n\t\t}\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class Error {}\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class Error {}\nclass MyError extends RangeError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class MyError extends Error {\n\tconstructor(): void;\n\tstatic {\n\t\tError.captureStackTrace(this, MyError)\n\n\t\tfunction foo() {\n\t\t\tError.captureStackTrace(this, MyError)\n\t\t}\n\t}\n}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, MyError);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, this.constructor);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, this.constructor);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, new.target);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, new.target);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends EvalError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends EvalError {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends RangeError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends RangeError {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends ReferenceError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends ReferenceError {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends SyntaxError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends SyntaxError {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends TypeError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends TypeError {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends URIError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends URIError {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends AggregateError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends AggregateError {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends SuppressedError {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends SuppressedError {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tconst foo = () => {\n\t\t\tError.captureStackTrace(this, MyError)\n\t\t}\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class MyError extends Error {\n\tconstructor() {\n\t\tconst foo = () => {\n\t\t\t\n\t\t}\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 4, Column: 4, EndLine: 4, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tif (a) Error.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 10, EndLine: 3, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tconst x = () => Error.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 19, EndLine: 3, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class MyError extends Error {\n\tconstructor() {\n\t\tvoid Error.captureStackTrace(this, MyError)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 8, EndLine: 3, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export default class extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, new.target)\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"export default class extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 3, Column: 3, EndLine: 3, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "export default (\n\tclass extends Error {\n\t\tconstructor() {\n\t\t\tError.captureStackTrace(this, new.target)\n\t\t}\n\t}\n)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"export default (\n\tclass extends Error {\n\t\tconstructor() {\n\t\t\t\n\t\t}\n\t}\n)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 4, Column: 4, EndLine: 4, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 4, Column: 3, EndLine: 4, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, MyError);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 4, Column: 3, EndLine: 4, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, this.constructor);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 4, Column: 3, EndLine: 4, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, this.constructor);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 4, Column: 3, EndLine: 4, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, new.target);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 4, Column: 3, EndLine: 4, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace?.(this, new.target);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nclass MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 4, Column: 3, EndLine: 4, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
