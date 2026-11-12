package use_unknown_in_catch_callback_variable

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// These cases were independently checked against typescript-eslint v8.67.0.
// Cover computed catch/then keys, variable aliases and shadowing, optional and
// parenthesized calls, and rejection of reassigned or runtime-only keys.
func TestUseUnknownInCatchCallbackVariableComputedKeys(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UseUnknownInCatchCallbackVariableRule, []rule_tester.ValidTestCase{
		{Code: "let method = 'catch'; method = 'then'; Promise.resolve()[method](() => {}, (error: Error) => {});"},
		{Code: "let method = 'catch'; Promise.resolve()[method](error => {}); method = 'then';"},
		{Code: "declare const method: 'catch'; Promise.resolve()[method](error => {});"},
		{Code: "let method = Math.random() ? 'catch' : 'then'; Promise.resolve()[method](error => {});"},
		{Code: "let method = 'finally'; Promise.resolve()[method]((error: Error) => {});"},
		{Code: "let method = 'catch'; Promise.resolve()[method]((error: unknown) => {});"},
		{Code: "let method = 'catch'; declare const other: { catch: (f: Function) => void }; other[method]((error: Error) => {});"},
		{Code: "let method = 'then'; Promise.resolve()[method]((error: Error) => {});"},
		{Code: "let method = 'then'; const args: [() => void] = [() => {}]; Promise.resolve()[method](...args, error => {});"},
		{Code: "let method = 'catch'; const handler = (error: Error) => {}; Promise.resolve()[method](handler);"},
	}, []rule_tester.InvalidTestCase{
		{Code: "let method = 'then'; Promise.resolve()[method]((value: any) => {}, (error: Error) => {});", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `then` rejection callback variable.", Line: 1, Column: 69, EndLine: 1, EndColumn: 81, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "wrongTypeAnnotationSuggestion", Output: "let method = 'then'; Promise.resolve()[method]((value: any) => {}, (error: unknown) => {});"},
			}},
		}},
		{Code: "Promise.resolve()['ca' + 'tch']((error: Error) => {});", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 1, Column: 34, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "wrongTypeAnnotationSuggestion", Output: "Promise.resolve()['ca' + 'tch']((error: unknown) => {});"},
			}},
		}},
		{Code: "const suffix = 'tch'; Promise.resolve()[`ca${suffix}`]((error: Error) => {});", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 1, Column: 57, EndLine: 1, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "wrongTypeAnnotationSuggestion", Output: "const suffix = 'tch'; Promise.resolve()[`ca${suffix}`]((error: unknown) => {});"},
			}},
		}},
		{Code: "let key = 'catch'; const alias = key; Promise.resolve()[alias](error => {});", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 1, Column: 64, EndLine: 1, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "addUnknownTypeAnnotationSuggestion", Output: "let key = 'catch'; const alias = key; Promise.resolve()[alias]((error: unknown) => {});"},
			}},
		}},
		{Code: "let key = 'catch'; Promise.resolve()?.[key]?.(error => {});", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 1, Column: 47, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "addUnknownTypeAnnotationSuggestion", Output: "let key = 'catch'; Promise.resolve()?.[key]?.((error: unknown) => {});"},
			}},
		}},
		{Code: "let key = 'catch'; (Promise.resolve()[key])(error => {});", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 1, Column: 45, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "addUnknownTypeAnnotationSuggestion", Output: "let key = 'catch'; (Promise.resolve()[key])((error: unknown) => {});"},
			}},
		}},
		{Code: "let key = 'catch'; Promise.resolve()[key](error => {}, ...[]);", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 1, Column: 43, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "addUnknownTypeAnnotationSuggestion", Output: "let key = 'catch'; Promise.resolve()[key]((error: unknown) => {}, ...[]);"},
			}},
		}},
		{Code: "const key = 'catch'; { let key = 'then'; Promise.resolve()[key](() => {}, error => {}); } Promise.resolve()[key](error => {});", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `then` rejection callback variable.", Line: 1, Column: 75, EndLine: 1, EndColumn: 80, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "addUnknownTypeAnnotationSuggestion", Output: "const key = 'catch'; { let key = 'then'; Promise.resolve()[key](() => {}, (error: unknown) => {}); } Promise.resolve()[key](error => {});"},
			}},
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 1, Column: 114, EndLine: 1, EndColumn: 119, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "addUnknownTypeAnnotationSuggestion", Output: "const key = 'catch'; { let key = 'then'; Promise.resolve()[key](() => {}, error => {}); } Promise.resolve()[key]((error: unknown) => {});"},
			}},
		}},
		{Code: "const method = 'catch'; Promise.resolve()[method as string](error => {});", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 1, Column: 61, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "addUnknownTypeAnnotationSuggestion", Output: "const method = 'catch'; Promise.resolve()[method as string]((error: unknown) => {});"},
			}},
		}},
		{Code: "const method = 'catch'; Promise.resolve()[method!](error => {});", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 1, Column: 52, EndLine: 1, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "addUnknownTypeAnnotationSuggestion", Output: "const method = 'catch'; Promise.resolve()[method!]((error: unknown) => {});"},
			}},
		}},
	})
}
