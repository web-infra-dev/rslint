// TestNoReturnAwaitExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
//
// N/A: the Dimension 4 access/key-form rows (identifier vs string vs numeric vs
// private vs computed keys, element access) don't apply — this rule never
// inspects property names.
package no_return_await

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoReturnAwaitExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoReturnAwaitRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS non-null assertion wraps the await ----
			{Code: `async function foo() {
	return (await bar())!;
}`},
			// ---- Dimension 4: TS `as` wraps the await ----
			{Code: `async function foo() {
	return (await bar()) as unknown;
}`},
			// ---- Dimension 4: TS `satisfies` wraps the await ----
			{Code: `async function foo() {
	return (await bar()) satisfies unknown;
}`},
			// ---- Dimension 4: optional member access wraps the await ----
			{Code: `async function foo() {
	return (await bar())?.x;
}`},
			// ---- Dimension 4: try block nested in a catch clause ----
			{Code: `async function foo() {
	try {}
	catch (e) {
		try {
			return await bar();
		} catch (e) {}
	}
}`},
			// ---- Dimension 4: bare return ----
			{Code: `async function foo() {
	return;
}`},
			// ---- Dimension 4: empty function body ----
			{Code: `async function foo() {}`},
			// ---- Dimension 4: empty arrow body ----
			{Code: `async () => {};`},
			// ---- Dimension 4: `for await` is not an await expression ----
			{Code: `async function foo() {
	for await (const x of y) {}
}`},
			// ---- Dimension 4: top-level await is not in a tail call position ----
			{Code: `await bar();`},
			// ---- Dimension 4: spread assignment in an object literal ----
			{Code: `async function foo() {
	return { ...(await bar()) };
}`},
			// ---- Dimension 4: spread element in an array literal ----
			{Code: `async function foo() {
	return [...(await bar())];
}`},
			// Locks in upstream isInTailCallPosition() arm 4: `??` left operand is not in tail position
			{Code: `async function foo() {
	return (await bar() ?? a);
}`},
			// Locks in upstream isInTailCallPosition() arm 5: a non-final comma operand is not in tail position
			{Code: `async function foo() {
	return (a, await bar(), b);
}`},
			// Locks in upstream isInTailCallPosition() default arm: a call argument is not in tail position
			{Code: `async function foo() {
	return baz(await bar());
}`},
			// Locks in upstream isInTailCallPosition() default arm: a callee is not in tail position
			{Code: `async function foo() {
	return (await bar())();
}`},
			// Locks in upstream isInTailCallPosition() default arm: an arithmetic operand is not in tail position
			{Code: `async function foo() {
	return (a + await bar());
}`},
			// Locks in upstream isInTailCallPosition() default arm: a unary operand is not in tail position
			{Code: `async function foo() {
	return void await bar();
}`},
			// Locks in upstream isInTailCallPosition() default arm: a template substitution is not in tail position
			{Code: "async function foo() {\n\treturn `${await bar()}`;\n}"},
			// Locks in upstream isInTailCallPosition() default arm: an array element is not in tail position
			{Code: `async function foo() {
	return [await bar()];
}`},
			// Locks in upstream isInTailCallPosition() default arm: a property value is not in tail position
			{Code: `async function foo() {
	return { a: await bar() };
}`},
			// Locks in upstream isInTailCallPosition() default arm: an assignment target is not in tail position
			{Code: `async function foo() {
	let a;
	return (a = await bar());
}`},
			// Locks in upstream isInTailCallPosition() default arm: `&&=` is an assignment, not a logical expression
			{Code: `async function foo() {
	let a;
	return (a &&= await bar());
}`},
			// Locks in upstream isInTailCallPosition() default arm: `||=` is an assignment, not a logical expression
			{Code: `async function foo() {
	let a;
	return (a ||= await bar());
}`},
			// Locks in upstream isInTailCallPosition() default arm: `??=` is an assignment, not a logical expression
			{Code: `async function foo() {
	let a;
	return (a ??= await bar());
}`},
			// Locks in upstream isInTailCallPosition() default arm: a yielded await is not in tail position
			{Code: `async function* foo() {
	yield await bar();
}`},
			// Locks in upstream isInTailCallPosition() arm 3: the conditional test is not in tail position
			{Code: `async () => (await bar() ? a : b)`},
			// Locks in upstream hasErrorHandler(): a catch clause without a finalizer keeps the report
			{Code: `async function foo() {
	try {}
	catch (e) {
		return await bar();
	}
	finally {}
}`},
			// Locks in upstream hasErrorHandler(): a try block followed by only a finalizer is an error handler
			{Code: `async function foo() {
	try {
		return await bar();
	} finally {}
}`},
			// ---- Real-user: eslint#15447 — `yield await` stays unreported ----
			{Code: `async function* foo() {
	yield await promise;
}`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized await as a concise arrow body ----
			{
				Code: `async () => (await bar())`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      1,
						Column:    14,
						EndLine:   1,
						EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async () => (bar())`},
						},
					},
				},
			},
			// ---- Dimension 4: multi-level parenthesized await as a concise arrow body ----
			{
				Code: `async () => ((await bar()))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      1,
						Column:    15,
						EndLine:   1,
						EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async () => ((bar()))`},
						},
					},
				},
			},
			// ---- Dimension 4: parenthesized await as a return argument ----
			{
				Code: `async function foo() {
	return (await bar());
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    10,
						EndLine:   2,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return (bar());
}`},
						},
					},
				},
			},
			// ---- Dimension 4: multi-level parenthesized await as a return argument ----
			{
				Code: `async function foo() {
	return ((await bar()));
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    11,
						EndLine:   2,
						EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return ((bar()));
}`},
						},
					},
				},
			},
			// ---- Dimension 4: parenthesized await as the last comma operand ----
			{
				Code: `async function foo() {
	return (a, (await bar()));
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    14,
						EndLine:   2,
						EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return (a, (bar()));
}`},
						},
					},
				},
			},
			// ---- Dimension 4: parenthesized comma sequence as a return argument ----
			{
				Code: `async function foo() {
	return ((a, await bar()));
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    14,
						EndLine:   2,
						EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return ((a, bar()));
}`},
						},
					},
				},
			},
			// ---- Dimension 4: parenthesized comma sequence as a concise arrow body ----
			{
				Code: `async () => ((a, await bar()))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      1,
						Column:    18,
						EndLine:   1,
						EndColumn: 29,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async () => ((a, bar()))`},
						},
					},
				},
			},
			// ---- Dimension 4: TS non-null assertion inside the await operand ----
			{
				Code: `async function foo() {
	return await bar()!;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    9,
						EndLine:   2,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return bar()!;
}`},
						},
					},
				},
			},
			// ---- Dimension 4: optional call inside the await operand ----
			{
				Code: `async function foo() {
	return await bar?.();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    9,
						EndLine:   2,
						EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return bar?.();
}`},
						},
					},
				},
			},
			// ---- Dimension 4: async function expression ----
			{
				Code: `const foo = async function () {
	return await bar();
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    9,
						EndLine:   2,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `const foo = async function () {
	return bar();
};`},
						},
					},
				},
			},
			// ---- Dimension 4: async method ----
			{
				Code: `class C {
	async m() {
		return await bar();
	}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      3,
						Column:    10,
						EndLine:   3,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `class C {
	async m() {
		return bar();
	}
}`},
						},
					},
				},
			},
			// ---- Dimension 4: async static method ----
			{
				Code: `class C {
	static async m() {
		return await bar();
	}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      3,
						Column:    10,
						EndLine:   3,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `class C {
	static async m() {
		return bar();
	}
}`},
						},
					},
				},
			},
			// ---- Dimension 4: class field holding an async arrow ----
			{
				Code: `class C {
	m = async () => await bar();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    18,
						EndLine:   2,
						EndColumn: 29,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `class C {
	m = async () => bar();
}`},
						},
					},
				},
			},
			// ---- Dimension 4: async object-literal method ----
			{
				Code: `const o = {
	async m() {
		return await bar();
	},
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      3,
						Column:    10,
						EndLine:   3,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `const o = {
	async m() {
		return bar();
	},
};`},
						},
					},
				},
			},
			// ---- Dimension 4: async generator ----
			{
				Code: `async function* foo() {
	return await bar();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    9,
						EndLine:   2,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function* foo() {
	return bar();
}`},
						},
					},
				},
			},
			// ---- Dimension 4: async method inside a try block is its own function boundary ----
			{
				Code: `async function foo() {
	try {
		class C {
			async m() {
				return await bar();
			}
		}
	} catch (e) {}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      5,
						Column:    12,
						EndLine:   5,
						EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	try {
		class C {
			async m() {
				return bar();
			}
		}
	} catch (e) {}
}`},
						},
					},
				},
			},
			// ---- Dimension 4: async object method inside a try block is its own function boundary ----
			{
				Code: `async function foo() {
	try {
		const o = {
			async m() {
				return await bar();
			},
		};
	} catch (e) {}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      5,
						Column:    12,
						EndLine:   5,
						EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	try {
		const o = {
			async m() {
				return bar();
			},
		};
	} catch (e) {}
}`},
						},
					},
				},
			},
			// ---- Dimension 4: three levels of nested conditionals ----
			{
				Code: `async function foo() {
	return (a ? (b ? (c ? await bar() : d) : e) : f);
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    24,
						EndLine:   2,
						EndColumn: 35,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return (a ? (b ? (c ? bar() : d) : e) : f);
}`},
						},
					},
				},
			},
			// ---- Dimension 4: rest element in a binding pattern ----
			{
				Code: `async function foo({ ...rest }) {
	return await bar();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    9,
						EndLine:   2,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo({ ...rest }) {
	return bar();
}`},
						},
					},
				},
			},
			// ---- Dimension 4: overload signature has no body ----
			{
				Code: `function foo(): Promise<void>;
async function foo(): Promise<void> {
	return await bar();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      3,
						Column:    9,
						EndLine:   3,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `function foo(): Promise<void>;
async function foo(): Promise<void> {
	return bar();
}`},
						},
					},
				},
			},
			// ---- Dimension 4: abstract member has no body ----
			{
				Code: `abstract class C {
	abstract m(): Promise<void>;
	async n() {
		return await bar();
	}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      4,
						Column:    10,
						EndLine:   4,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `abstract class C {
	abstract m(): Promise<void>;
	async n() {
		return bar();
	}
}`},
						},
					},
				},
			},
			// ---- Dimension 4: only the outer function body is in tail position ----
			{
				Code: `async function foo() {
	const g = async () => {
		await bar();
	};
	return await baz();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      5,
						Column:    9,
						EndLine:   5,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	const g = async () => {
		await bar();
	};
	return baz();
}`},
						},
					},
				},
			},
			// Locks in upstream isInTailCallPosition() arm 4: `??` right operand recurses
			{
				Code: `async function foo() {
	return (a ?? await bar());
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    15,
						EndLine:   2,
						EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return (a ?? bar());
}`},
						},
					},
				},
			},
			// Locks in upstream isInTailCallPosition() arm 5: only the final comma operand is reported
			{
				Code: `async function foo() {
	return (await bar(), await baz());
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    23,
						EndLine:   2,
						EndColumn: 34,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return (await bar(), baz());
}`},
						},
					},
				},
			},
			// Locks in upstream isInTailCallPosition() default arm: an inner await is not in tail position
			{
				Code: `async function foo() {
	return await await bar();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    9,
						EndLine:   2,
						EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return await bar();
}`},
						},
					},
				},
			},
			// Locks in upstream isInTailCallPosition() arm 3: both conditional branches are reported
			{
				Code: `async function foo() {
	return (a ? await bar() : await baz());
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    14,
						EndLine:   2,
						EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return (a ? bar() : await baz());
}`},
						},
					},
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    28,
						EndLine:   2,
						EndColumn: 39,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return (a ? await bar() : baz());
}`},
						},
					},
				},
			},
			// Locks in upstream hasErrorHandler(): a finally block is not an error handler
			{
				Code: `async function foo() {
	try {
		baz();
	} finally {
		return await bar();
	}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      5,
						Column:    10,
						EndLine:   5,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	try {
		baz();
	} finally {
		return bar();
	}
}`},
						},
					},
				},
			},
			// Locks in upstream hasErrorHandler(): only the return inside the catch clause is reported
			{
				Code: `async function foo() {
	try {
		return await bar();
	} catch (e) {
		return await baz();
	}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      5,
						Column:    10,
						EndLine:   5,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	try {
		return await bar();
	} catch (e) {
		return baz();
	}
}`},
						},
					},
				},
			},
			// ---- Dimension 3: a tab after the keyword is not trimmed ----
			{
				Code: `async function foo() {
	return await	bar();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    9,
						EndLine:   2,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return 	bar();
}`},
						},
					},
				},
			},
			// ---- Dimension 3: only one of two spaces after the keyword is trimmed ----
			{
				Code: `async function foo() {
	return await  bar();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    9,
						EndLine:   2,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return  bar();
}`},
						},
					},
				},
			},
			// ---- Dimension 3: a comment right after the keyword is preserved ----
			{
				Code: `async function foo() {
	return await/* c */bar();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    9,
						EndLine:   2,
						EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return /* c */bar();
}`},
						},
					},
				},
			},
			// ---- Dimension 3: a line break between the keyword and its operand drops the suggestion ----
			{
				Code: `async function foo() {
	return await
		bar();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    9,
						EndLine:   3,
						EndColumn: 8,
					},
				},
			},
			// ---- Real-user: eslint#8255 — an earlier `await` in the body does not exempt the return ----
			{
				Code: `async function foo() {
	await baz();
	return await bar();
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      3,
						Column:    9,
						EndLine:   3,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	await baz();
	return bar();
}`},
						},
					},
				},
			},
			// ---- Real-user: eslint#10835 — only the right operand of `||` is redundant ----
			{
				Code: `async function foo(a, b) {
	return await a || await b;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    20,
						EndLine:   2,
						EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo(a, b) {
	return await a || b;
}`},
						},
					},
				},
			},
			// ---- Real-user: eslint#7594 — comma sequence starting with a non-await operand ----
			{
				Code: `async function foo() {
	return (0, await bar());
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    13,
						EndLine:   2,
						EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return (0, bar());
}`},
						},
					},
				},
			},
			// ---- Real-user: eslint#7594 — conditional alternate holding a comma sequence ----
			{
				Code: `async function foo() {
	return (0 ? 1 : (2, 3, await bar()));
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Message:   "Redundant use of `await` on a return value.",
						Line:      2,
						Column:    25,
						EndLine:   2,
						EndColumn: 36,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `async function foo() {
	return (0 ? 1 : (2, 3, bar()));
}`},
						},
					},
				},
			},
		},
	)
}

func TestNoReturnAwaitEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"async function suggested() {\n\treturn await bar();\n}\nasync function multiline() {\n\treturn await\n\t\tbaz();\n}\n",
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      program,
			File:         sourceFile.FileName(),
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     NoReturnAwaitRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoReturnAwaitRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		for index, diagnostic := range diagnostics {
			if diagnostic.Message.Id != "redundantUseOfAwait" {
				t.Fatalf(
					"demand %d diagnostic %d: message id = %q, want redundantUseOfAwait",
					demand,
					index,
					diagnostic.Message.Id,
				)
			}
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for index, allEditsDiagnostic := range allEdits {
		for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly[index],
			rule.EditDemandAutofix:    autofixOnly[index],
			rule.EditDemandSuggestion: suggestionOnly[index],
		} {
			if got, want := withoutEdits(diagnostic), withoutEdits(allEditsDiagnostic); !reflect.DeepEqual(got, want) {
				t.Errorf(
					"diagnostic %d demand %d changed identity:\ngot  %#v\nwant %#v",
					index,
					demand,
					got,
					want,
				)
			}
		}
		if diagnosticsOnly[index].Suggestions != nil || autofixOnly[index].Suggestions != nil {
			t.Errorf("diagnostic %d: non-suggestion demand materialized suggestions", index)
		}
		if !reflect.DeepEqual(suggestionOnly[index].Suggestions, allEditsDiagnostic.Suggestions) {
			t.Errorf("diagnostic %d: suggestion and all-edits demands produced different suggestions", index)
		}
	}

	if allEdits[0].Suggestions == nil || len(*allEdits[0].Suggestions) != 1 {
		t.Error("same-line await did not produce exactly one suggestion")
	}
	if allEdits[1].Suggestions != nil {
		t.Error("await split across lines unexpectedly produced a suggestion")
	}
	for _, diagnostics := range [][]rule.RuleDiagnostic{
		diagnosticsOnly,
		autofixOnly,
		suggestionOnly,
		allEdits,
	} {
		for index, diagnostic := range diagnostics {
			if diagnostic.FixesPtr != nil {
				t.Errorf("diagnostic %d: suggestion-only rule materialized autofixes", index)
			}
		}
	}
}
