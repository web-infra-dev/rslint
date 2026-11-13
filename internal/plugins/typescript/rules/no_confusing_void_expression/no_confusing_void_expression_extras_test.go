package no_confusing_void_expression

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConfusingVoidExpressionExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoConfusingVoidExpressionRule, []rule_tester.ValidTestCase{
		// JSX: arrow function with body in void-returning prop
		{
			Code:    `const foo = <button onClick={() => { console.log(); return console.log(); }} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreVoidReturningFunctions": true},
		},
		// JSX: custom component with void-returning callback prop
		{
			Code: `
declare function Comp(props: { handler: () => void }): any;
const foo = <Comp handler={() => console.log()} />;
`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreVoidReturningFunctions": true},
		},
		// JSX: union type prop including void
		{
			Code: `
declare function Comp(props: { handler: (() => void) | (() => number) }): any;
const foo = <Comp handler={() => console.log()} />;
`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreVoidReturningFunctions": true},
		},
	}, []rule_tester.InvalidTestCase{
		// JSX: without ignoreVoidReturningFunctions, should still report
		{
			Code:   `const foo = <button onClick={() => console.log()} />;`,
			Output: []string{`const foo = <button onClick={() => { console.log(); }} />;`},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 36, EndLine: 1, EndColumn: 49},
			},
		},
		// JSX: non-void returning prop (e.g. returns string) should still report
		{
			Code: `
declare function Comp(props: { getValue: () => string }): any;
const foo = <Comp getValue={() => console.log()} />;
`,
			Output: []string{`
declare function Comp(props: { getValue: () => string }): any;
const foo = <Comp getValue={() => { console.log(); }} />;
`},
			Tsx:     true,
			Options: map[string]interface{}{"ignoreVoidReturningFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 3, Column: 35, EndLine: 3, EndColumn: 48},
			},
		},
	})
}

// These boundary cases were checked against upstream v8.67.0. The multiple
// parentheses case intentionally expects a safe fix (see Differences from upstream).
func TestNoConfusingVoidExpressionDifferential(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoConfusingVoidExpressionRule, []rule_tester.ValidTestCase{
		{Code: "function cool(input: string) { return input, console.log(input), input; }"},
		{Code: "function cool(input: string) { return input, console.log(input), console.log(input), input; }"},
		{Code: "function cool(input: string) { return input, (console.log(input), input); }"},
		{Code: "async function f() { await Promise.resolve(); }"},
		{Code: "declare function tag(s: TemplateStringsArray): void; tag`hello`;"},
		{Code: "declare function f(): never; const x = f();"},
		{Code: "declare function f(): void | number; const x = f();"},
		{Code: "const f: (() => void) & { a?: number } = () => console.log();", Options: map[string]any{"ignoreVoidReturningFunctions": true}},
		{Code: "const f: () => void = function () { return console.log(); };", Options: map[string]any{"ignoreVoidReturningFunctions": true}}}, []rule_tester.InvalidTestCase{
		{Code: "function cool(input: string) { return (input, console.log(input)), input; }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExpr", Message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.", Line: 1, Column: 47, EndLine: 1, EndColumn: 65}}},
		{Code: "function cool(input: string) { return console.log(input), console.log(input); }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExpr", Message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.", Line: 1, Column: 59, EndLine: 1, EndColumn: 77}}},
		{Code: "async function f() { const x = await Promise.resolve(); }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExpr", Message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.", Line: 1, Column: 32, EndLine: 1, EndColumn: 55}}},
		{Code: "async function f() { return await Promise.resolve(); }", Output: []string{"async function f() { await Promise.resolve(); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 29, EndLine: 1, EndColumn: 52}}},
		{Code: "async function f() { if (true) return await Promise.resolve(); }", Output: []string{"async function f() { if (true) { await Promise.resolve(); return; } }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 39, EndLine: 1, EndColumn: 62}}},
		{Code: "declare function tag(s: TemplateStringsArray): void; const x = tag`hello`;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExpr", Message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.", Line: 1, Column: 64, EndLine: 1, EndColumn: 74}}},
		{Code: "declare function tag(s: TemplateStringsArray): void; const f = () => tag`hello`;", Output: []string{"declare function tag(s: TemplateStringsArray): void; const f = () => { tag`hello`; };"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 70, EndLine: 1, EndColumn: 80}}},
		{Code: "declare function tag(s: TemplateStringsArray): void; function f() { return tag`hello`; }", Output: []string{"declare function tag(s: TemplateStringsArray): void; function f() { tag`hello`; }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 76, EndLine: 1, EndColumn: 86}}},
		{Code: "declare function f<T extends void>(): T; const x = f();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExpr", Message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.", Line: 1, Column: 52, EndLine: 1, EndColumn: 55}}},
		{Code: "declare function f(): undefined; const x = f();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExpr", Message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.", Line: 1, Column: 44, EndLine: 1, EndColumn: 47}}},
		{Code: "const f = () => (console.log());", Output: []string{"const f = () => { console.log(); };"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 18, EndLine: 1, EndColumn: 31}}},
		{Code: "const f = () => ((console.log()));", Output: []string{"const f = () => { console.log(); };"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 19, EndLine: 1, EndColumn: 32}}},
		{Code: "const f = () => /*comment*/ console.log();", Output: []string{"const f = () => /*comment*/ { console.log(); };"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 29, EndLine: 1, EndColumn: 42}}},
		{Code: "function f() { return (console.log()); }", Output: []string{"function f() { console.log(); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 24, EndLine: 1, EndColumn: 37}}},
		{Code: "function f() { return /*comment*/ console.log(); }", Output: []string{"function f() { console.log(); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 35, EndLine: 1, EndColumn: 48}}},
		{Code: "function f() { if (true) return (console.log()); }", Output: []string{"function f() { if (true) { console.log(); return; } }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 34, EndLine: 1, EndColumn: 47}}},
		{Code: "function f() { if (true) { return console.log(); } }", Output: []string{"function f() { if (true) { console.log(); return; } }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturn", Message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.", Line: 1, Column: 35, EndLine: 1, EndColumn: 48}}},
		{Code: "function f() { return console.log() }", Output: []string{"function f() { console.log(); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 23, EndLine: 1, EndColumn: 36}}},
		{Code: "const f = () => (true ? console.log(1) : console.log(2));", Output: []string{"const f = () => { true ? console.log(1) : console.log(2); };"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 25, EndLine: 1, EndColumn: 39}, {MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 42, EndLine: 1, EndColumn: 56}}},
		{Code: "class C { method() { return console.log(); } }", Output: []string{"class C { method() { console.log(); } }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 29, EndLine: 1, EndColumn: 42}}},
		{Code: "const obj = { method() { return console.log(); } };", Output: []string{"const obj = { method() { console.log(); } };"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturnLast", Message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.", Line: 1, Column: 33, EndLine: 1, EndColumn: 46}}},
		{Code: "const f = () => console?.log();", Output: []string{"const f = () => { console?.log(); };"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprArrow", Message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.", Line: 1, Column: 17, EndLine: 1, EndColumn: 31}}},
		{Code: "console?.log()?.toString();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExpr", Message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15}}},
		{Code: "const x = console[\"log\"]();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExpr", Message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 27}}},
		{Code: "const x = console.log!();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExpr", Message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25}}},
		{Code: "async function f() { const x = await Promise.resolve(); }", Options: map[string]any{"ignoreVoidOperator": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprWrapVoid", Message: "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.", Line: 1, Column: 32, EndLine: 1, EndColumn: 55, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "voidExprWrapVoid", Output: "async function f() { const x = void await Promise.resolve(); }"}}}}},
		{Code: "declare function tag(s: TemplateStringsArray): void; const x = tag`hello`;", Options: map[string]any{"ignoreVoidOperator": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprWrapVoid", Message: "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.", Line: 1, Column: 64, EndLine: 1, EndColumn: 74, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "voidExprWrapVoid", Output: "declare function tag(s: TemplateStringsArray): void; const x = void tag`hello`;"}}}}},
		{Code: "const f = () => (console.log());", Options: map[string]any{"ignoreVoidOperator": true}, Output: []string{"const f = () => (void console.log());"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprArrowWrapVoid", Message: "Void expressions returned from an arrow function shorthand must be marked explicitly with the `void` operator.", Line: 1, Column: 18, EndLine: 1, EndColumn: 31}}},
		{Code: "function f() { return (console.log()); }", Options: map[string]any{"ignoreVoidOperator": true}, Output: []string{"function f() { return (void console.log()); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidVoidExprReturnWrapVoid", Message: "Void expressions returned from a function must be marked explicitly with the `void` operator.", Line: 1, Column: 24, EndLine: 1, EndColumn: 37}}}})
}
