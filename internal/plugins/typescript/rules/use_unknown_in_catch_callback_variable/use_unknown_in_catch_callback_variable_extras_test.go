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

// The original review reproduction and related cases were checked against v8.67.0.
func TestUseUnknownInCatchCallbackVariableCallMutations(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UseUnknownInCatchCallbackVariableRule, []rule_tester.ValidTestCase{
		// reviewer reproduction
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\nfunction changeMethod(holder: typeof methods) {\n  holder.current = 'then';\n}\nchangeMethod(methods);\n\nconst promise: Record<'then' | 'catch', (\n  callback: (value: number) => unknown,\n  onRejected?: (error: unknown) => unknown,\n) => unknown> = Promise.resolve(1);\n\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// alias
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\ndeclare const mutate: any;\nconst alias = methods; mutate(alias);\nconst promise: Record<'then' | 'catch', (callback: (value: number) => unknown, onRejected?: (error: unknown) => unknown) => unknown> = Promise.resolve(1);\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// nested container
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\ndeclare const mutate: any;\nmutate({nested: [methods]});\nconst promise: Record<'then' | 'catch', (callback: (value: number) => unknown, onRejected?: (error: unknown) => unknown) => unknown> = Promise.resolve(1);\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// apply
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\ndeclare const mutate: any;\nmutate.apply(null, [methods]);\nconst promise: Record<'then' | 'catch', (callback: (value: number) => unknown, onRejected?: (error: unknown) => unknown) => unknown> = Promise.resolve(1);\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// constructor
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\ndeclare const mutate: any;\nnew mutate(methods);\nconst promise: Record<'then' | 'catch', (callback: (value: number) => unknown, onRejected?: (error: unknown) => unknown) => unknown> = Promise.resolve(1);\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// optional call
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\ndeclare const mutate: any;\nmutate?.(methods);\nconst promise: Record<'then' | 'catch', (callback: (value: number) => unknown, onRejected?: (error: unknown) => unknown) => unknown> = Promise.resolve(1);\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// shallow freeze
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\ndeclare const mutate: any;\nmutate(Object.freeze({methods}));\nconst promise: Record<'then' | 'catch', (callback: (value: number) => unknown, onRejected?: (error: unknown) => unknown) => unknown> = Promise.resolve(1);\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// callback exposes object
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\ndeclare const mutate: any;\nmutate(() => methods);\nconst promise: Record<'then' | 'catch', (callback: (value: number) => unknown, onRejected?: (error: unknown) => unknown) => unknown> = Promise.resolve(1);\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
	}, []rule_tester.InvalidTestCase{
		{Code: "let method = 'catch';\nfunction change(value: string) { value = 'then'; }\nchange(method);\nPromise.resolve()[method]((error: Error) => {});\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 4, Column: 28, EndLine: 4, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrongTypeAnnotationSuggestion", Output: "let method = 'catch';\nfunction change(value: string) { value = 'then'; }\nchange(method);\nPromise.resolve()[method]((error: unknown) => {});\n"}}}}},
	})
}

func TestUseUnknownInCatchCallbackVariableCapturedValues(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UseUnknownInCatchCallbackVariableRule, []rule_tester.ValidTestCase{
		// later alias
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\nlet alias: typeof methods = { current: 'then' };\nalias = methods;\nfunction change(value: typeof methods) {\n  value.current = 'then';\n}\nchange(alias);\n\nconst promise: Record<'then' | 'catch', (\n  callback: (value: number) => unknown,\n  onRejected?: (error: unknown) => unknown,\n) => unknown> = Promise.resolve(1);\n\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// Object.assign
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\nconst holder: { methods?: typeof methods } = {};\nObject.assign(holder, { methods });\nfunction change(value: typeof holder) {\n  value.methods!.current = 'then';\n}\nchange(holder);\n\nconst promise: Record<'then' | 'catch', (\n  callback: (value: number) => unknown,\n  onRejected?: (error: unknown) => unknown,\n) => unknown> = Promise.resolve(1);\n\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// Object.defineProperty
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\nconst holder: { methods?: typeof methods } = {};\nObject.defineProperty(holder, 'methods', { value: methods });\nfunction change(value: typeof holder) {\n  value.methods!.current = 'then';\n}\nchange(holder);\n\nconst promise: Record<'then' | 'catch', (\n  callback: (value: number) => unknown,\n  onRejected?: (error: unknown) => unknown,\n) => unknown> = Promise.resolve(1);\n\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// descriptor alias
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\nconst holder: { methods?: typeof methods } = {};\nconst descriptor = { value: methods };\nObject.defineProperty(holder, 'methods', descriptor);\nfunction change(value: typeof holder) {\n  value.methods!.current = 'then';\n}\nchange(holder);\n\nconst promise: Record<'then' | 'catch', (\n  callback: (value: number) => unknown,\n  onRejected?: (error: unknown) => unknown,\n) => unknown> = Promise.resolve(1);\n\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// defineProperties
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\nconst holder: { methods?: typeof methods } = {};\nObject.defineProperties(holder, { methods: { value: methods } });\nfunction change(value: typeof holder) {\n  value.methods!.current = 'then';\n}\nchange(holder);\n\nconst promise: Record<'then' | 'catch', (\n  callback: (value: number) => unknown,\n  onRejected?: (error: unknown) => unknown,\n) => unknown> = Promise.resolve(1);\n\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
		// Reflect.set
		{Code: "const methods: { current: 'catch' | 'then' } = { current: 'catch' };\nconst holder: { methods?: typeof methods } = {};\nReflect.set(holder, 'methods', methods);\nfunction change(value: typeof holder) {\n  value.methods!.current = 'then';\n}\nchange(holder);\n\nconst promise: Record<'then' | 'catch', (\n  callback: (value: number) => unknown,\n  onRejected?: (error: unknown) => unknown,\n) => unknown> = Promise.resolve(1);\n\npromise[methods.current](value => value + 1, (error: unknown) => {});\n"},
	}, []rule_tester.InvalidTestCase{
		// copied primitive
		{Code: "const methods = { current: 'catch' as const };\nconst key = methods.current;\nfunction change(value: { current: string }) {\n  value.current = 'then';\n}\nchange(methods);\nPromise.resolve(1)[key]((error: any) => {});\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 7, Column: 26, EndLine: 7, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrongTypeAnnotationSuggestion", Output: "const methods = { current: 'catch' as const };\nconst key = methods.current;\nfunction change(value: { current: string }) {\n  value.current = 'then';\n}\nchange(methods);\nPromise.resolve(1)[key]((error: unknown) => {});\n"}}}}},
		// primitive alias
		{Code: "const methods = { current: 'catch' as const };\nconst key = methods.current;\nfunction change(value: { current: string }) {\n  value.current = 'then';\n}\nchange(methods);\nconst second = key;\nPromise.resolve(1)[second]((error: any) => {});\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 8, Column: 29, EndLine: 8, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrongTypeAnnotationSuggestion", Output: "const methods = { current: 'catch' as const };\nconst key = methods.current;\nfunction change(value: { current: string }) {\n  value.current = 'then';\n}\nchange(methods);\nconst second = key;\nPromise.resolve(1)[second]((error: unknown) => {});\n"}}}}},
		// derived primitive
		{Code: "const methods = { current: 'catch' as const };\nconst key = methods.current;\nfunction change(value: { current: string }) {\n  value.current = 'then';\n}\nchange(methods);\nconst second = key + '';\nPromise.resolve(1)[second]((error: any) => {});\n", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useUnknown", Message: "Prefer the safe `: unknown` for a `catch` callback variable.", Line: 8, Column: 29, EndLine: 8, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "wrongTypeAnnotationSuggestion", Output: "const methods = { current: 'catch' as const };\nconst key = methods.current;\nfunction change(value: { current: string }) {\n  value.current = 'then';\n}\nchange(methods);\nconst second = key + '';\nPromise.resolve(1)[second]((error: unknown) => {});\n"}}}}},
	})
}
