package no_floating_promises

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoFloatingPromisesExtras(t *testing.T) {
	safeCallOpts := map[string]interface{}{
		"allowForKnownSafeCalls": []interface{}{
			map[string]interface{}{"from": "file", "name": []interface{}{"safeCall", "safeArrayCall"}, "path": "file.ts"},
		},
	}
	safePromiseOpts := map[string]interface{}{
		"allowForKnownSafePromises": []interface{}{
			map[string]interface{}{"from": "file", "name": []interface{}{"SafePromise"}},
		},
	}
	voidAwait := func(voidOut, awaitOut string) []rule_tester.InvalidTestCaseSuggestion {
		return []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "floatingFixVoid", Output: voidOut},
			{MessageId: "floatingFixAwait", Output: awaitOut},
		}
	}
	awaitOnly := func(awaitOut string) []rule_tester.InvalidTestCaseSuggestion {
		return []rule_tester.InvalidTestCaseSuggestion{
			{MessageId: "floatingFixAwait", Output: awaitOut},
		}
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoFloatingPromisesRule,
		[]rule_tester.ValidTestCase{
			// allowForKnownSafeCalls with floating, parenthesized, promise-array, non-floating, and handled calls.
			{
				Code: `
declare function safeCall(): Promise<void>;
declare function safeArrayCall(): Promise<void>[];
declare function syncCall(): void;
declare function unsafeCall(): Promise<void>;
safeCall();
(safeCall());
safeArrayCall();
syncCall();
unsafeCall().catch(() => {});
`,
				FileName: "file.ts",
				Options:  safeCallOpts,
			},
			// ignoreVoid: true with parenthesized and nested void expressions.
			{
				Code: `
declare const promiseArray: Promise<number>[];
declare const cond: boolean;
(void Promise.resolve());
((void (Promise.resolve())));
(void promiseArray);
123, void Promise.resolve();
void Promise.resolve(), 123;
Promise.resolve().catch(() => {}) || void Promise.resolve();
cond ? Promise.resolve().catch(() => {}) : void Promise.resolve();
`,
				Options: map[string]interface{}{"ignoreVoid": true},
			},
			// Non-promise intrinsic types and non-promise arrays/tuples under checkThenables.
			{
				Code: `
declare function voidFn(): void;
declare function neverFn(): never;
declare const anyVal: any;
declare const maybeStr: string | undefined;
declare const anyArr: any[];
declare const nullableTuple: [null, undefined];
voidFn(); neverFn(); anyVal; undefined; null; maybeStr; anyArr; nullableTuple;
`,
				Options: map[string]interface{}{"checkThenables": true},
			},
			// allowForKnownSafePromises allows arrays of safe promises.
			{
				Code: `
class SafePromise<T> extends Promise<T> {}
declare const safeArr: SafePromise<number>[];
safeArr;
`,
				Options: safePromiseOpts,
			},
			// Generic constrained to void becomes void after getApparentType under checkThenables.
			{
				Code:    "const f = <T extends void>(x: T) => { x; };",
				Options: map[string]interface{}{"checkThenables": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			// allowForKnownSafeCalls still reports unsafe floating calls and chained calls on safe functions.
			{
				Code:     "declare function unsafeCall(): Promise<void>;\nunsafeCall();",
				FileName: "file.ts",
				Options:  safeCallOpts,
				Errors: []rule_tester.InvalidTestCaseError{{
					Line: 2, MessageId: "floatingVoid",
					Suggestions: voidAwait("declare function unsafeCall(): Promise<void>;\nvoid unsafeCall();", "declare function unsafeCall(): Promise<void>;\nawait unsafeCall();"),
				}},
			},
			{
				Code:     "declare function safeCall(): Promise<void>;\nsafeCall().catch(123 as any);",
				FileName: "file.ts",
				Options:  safeCallOpts,
				Errors: []rule_tester.InvalidTestCaseError{{
					Line: 2, MessageId: "floatingUselessRejectionHandlerVoid",
					Suggestions: voidAwait("declare function safeCall(): Promise<void>;\nvoid safeCall().catch(123 as any);", "declare function safeCall(): Promise<void>;\nawait safeCall().catch(123 as any);"),
				}},
			},
			// ignoreVoid: false with parenthesized and nested void expressions.
			{
				Code:    "(void Promise.resolve());",
				Options: map[string]interface{}{"ignoreVoid": false},
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, MessageId: "floating", Suggestions: awaitOnly("(await Promise.resolve());")}},
			},
			{
				Code:    "declare const promiseArray: Promise<number>[];\n(void promiseArray);",
				Options: map[string]interface{}{"ignoreVoid": false},
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 2, MessageId: "floatingPromiseArray"}},
			},
			{
				Code:    "123, void Promise.resolve();",
				Options: map[string]interface{}{"ignoreVoid": false},
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, MessageId: "floating", Suggestions: awaitOnly("await (123, void Promise.resolve());")}},
			},
			{
				Code:    "Promise.resolve().catch(() => {}) || void Promise.resolve();",
				Options: map[string]interface{}{"ignoreVoid": false},
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, MessageId: "floating", Suggestions: awaitOnly("await (Promise.resolve().catch(() => {}) || void Promise.resolve());")}},
			},
			// Awaited promise arrays, tuples with nullable promise elements, and constrained promise-array generics.
			{
				Code: `
async function test<A extends Promise<number>[]>(
  tupleWithNullablePromise: [string, Promise<number> | null],
  getPromiseArray: () => Promise<Promise<number>[]>,
  genericArr: A,
) {
  tupleWithNullablePromise;
  await getPromiseArray();
  genericArr;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{Line: 7, MessageId: "floatingPromiseArrayVoid"},
					{Line: 8, MessageId: "floatingPromiseArrayVoid"},
					{Line: 9, MessageId: "floatingPromiseArrayVoid"},
				},
			},
			// Union containing void is not skipped by the intrinsic fast path.
			{
				Code: "declare const p: Promise<number> | void;\np;",
				Errors: []rule_tester.InvalidTestCaseError{{
					Line: 2, MessageId: "floatingVoid",
					Suggestions: voidAwait("declare const p: Promise<number> | void;\nvoid p;", "declare const p: Promise<number> | void;\nawait p;"),
				}},
			},
			// Generic constrained to Promise.
			{
				Code: "function f<T extends Promise<number>>(p: T) { p; }",
				Errors: []rule_tester.InvalidTestCaseError{{
					Line: 1, MessageId: "floatingVoid",
					Suggestions: voidAwait("function f<T extends Promise<number>>(p: T) { void p; }", "function f<T extends Promise<number>>(p: T) { await p; }"),
				}},
			},
			// Nullable thenable under checkThenables skips undefined and checks the thenable branch.
			{
				Code:    "interface T { then(a: () => void, b: () => void): T }\ndeclare const x: T | undefined;\nx;",
				Options: map[string]interface{}{"checkThenables": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					Line: 3, MessageId: "floatingVoid",
					Suggestions: voidAwait("interface T { then(a: () => void, b: () => void): T }\ndeclare const x: T | undefined;\nvoid x;", "interface T { then(a: () => void, b: () => void): T }\ndeclare const x: T | undefined;\nawait x;"),
				}},
			},
			// allowForKnownSafePromises requires all union constituents to match, so SafePromise | undefined still reports.
			{
				Code:    "class SafePromise<T> extends Promise<T> {}\ndeclare const maybeSafe: SafePromise<number> | undefined;\nmaybeSafe;",
				Options: safePromiseOpts,
				Errors: []rule_tester.InvalidTestCaseError{{
					Line: 3, MessageId: "floatingVoid",
					Suggestions: voidAwait("class SafePromise<T> extends Promise<T> {}\ndeclare const maybeSafe: SafePromise<number> | undefined;\nvoid maybeSafe;", "class SafePromise<T> extends Promise<T> {}\ndeclare const maybeSafe: SafePromise<number> | undefined;\nawait maybeSafe;"),
				}},
			},
			// unknown under strictNullChecks: false is not skipped by nonPromiseTypeFlags and sees augmented Object.then.
			{
				Code:     "interface Object { then(a: () => void, b: () => void): void }\ndeclare const u: unknown;\nu;",
				TSConfig: "tsconfig.strict-with-null-checks-off.json",
				Options:  map[string]interface{}{"checkThenables": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					Line: 3, MessageId: "floatingVoid",
					Suggestions: voidAwait("interface Object { then(a: () => void, b: () => void): void }\ndeclare const u: unknown;\nvoid u;", "interface Object { then(a: () => void, b: () => void): void }\ndeclare const u: unknown;\nawait u;"),
				}},
			},
		},
	)
}
