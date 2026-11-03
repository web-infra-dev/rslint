// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-useless-fallback-in-spread.js
package no_useless_fallback_in_spread_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_fallback_in_spread"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessFallbackInSpreadUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_fallback_in_spread.NoUselessFallbackInSpreadRule, []rule_tester.ValidTestCase{
		// Snapshot cases
		{Code: "const array = [...(foo || [])]", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const array = [...(foo || {})]", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const array = [...(foo && {})]", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo && {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...({} || foo)}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...({} && foo)}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...({} ?? foo)}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...foo}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo ?? ({} || {}))}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const {...foo} = object", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "function foo({...bar}){}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo || {}).toString()}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...fn(foo || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = call({}, ...(foo || {}))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo || {not: \"empty\"})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo || {...{}})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo ? {bar: true} : {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(!foo ? {} : {bar: true})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo ? {a: 1} : {b: 2})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo ? {} : {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo ? foo : {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(!foo ? {} : foo)}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo.bar ? foo.bar : {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const array = [...(foo ? {a: 1} : {})]", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo ? {a: 1} : {}).bar}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Documentation examples
		{Code: "const object = {...foo};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const object = {...(foo && {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const array = [...(foo || [])];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		// Comment-preserving fixes
		{Code: "const object = {...(getObject(/* keep */) || {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...getObject(/* keep */)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 46, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((/* keep */ foo) ?? {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(/* keep */ foo)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 41, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...({/* keep */ value: 1} || {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...{/* keep */ value: 1}};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 46, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(getObject(// keep\n) ?? {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...getObject(// keep\n)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 2, Column: 6, EndLine: 2, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(getObject(// keep\r\n) ?? {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...getObject(// keep\r\n)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 2, Column: 6, EndLine: 2, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// TypeScript and comments blocking fixes
		{Code: "const object = {...((foo as object /* keep */) ?? {})};", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(foo as object /* keep */)};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 51, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo || {/* keep */})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 28, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((foo) /* keep */ ?? {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 41, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// Snapshot cases
		{Code: "const object = {...(foo || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo ?? {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo ?? (( {} )))}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 31, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((( foo )) ?? (( {} )))}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(( foo ))}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 37, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(( (( foo )) ?? (( {} )) ))}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {... (( foo )) }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 39, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "async ()=> ({...((await foo) || {})})", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"async ()=> ({...(await foo)})"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 33, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(0 || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...0}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 26, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((-0) || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(-0)}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 29, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(.0 || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {....0}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(0n || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...0n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(false || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...false}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 30, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(null || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...null}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 29, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(undefined || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...undefined}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 34, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((a && b) || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(a && b)}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 33, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(NaN || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...NaN}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(\"\" || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...\"\"}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...([] || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...[]}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...({} || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...{}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo || {}),}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo,}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo /* keep */ || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 39, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo || /* keep */ {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 39, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((foo ?? {}) || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(foo ?? {})}", "const object = {...foo}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 36, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...((foo && {}) || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...(foo && {})}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 36, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo && {} || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo && {}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 34, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...({...(foo || {})})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...({...foo})}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 33, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...({...((0, foo) || {})})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...({...(0, foo)})}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 38, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function foo(a = {...(bar || {})}){}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function foo(a = {...bar}){}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 30, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(document.all || {})}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...document.all}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 37, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// Runtime-equivalence regression
		{Code: "JSON.stringify({...((value /* keep */) || {})});", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"JSON.stringify({...(value /* keep */)});"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 43, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "JSON.stringify({...((value /* keep */) ?? {})});", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"JSON.stringify({...(value /* keep */)});"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 43, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		// Documentation examples
		{Code: "const object = {...(foo || {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const object = {...(foo ?? {})};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const object = {...foo};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-fallback-in-spread", Message: "The empty object is useless.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
