// TestJsxFunctionCallNewlineExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in. Upstream mirrors live in jsx_function_call_newline_upstream_test.go.
//
// The upstream suite only covers JsxSelfClosingElement arguments under `fn(...)`
// and `new OBJ(...)`; it never exercises JSX fragments, JSX elements with
// children, optional calls, nested calls, spreads, type-wrapped JSX, the
// CLI/bare-string option shape, or the "next token on a later line" branch of
// needsClosingNewLine. Those all live here.
package jsx_function_call_newline

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxFunctionCallNewlineExtras(t *testing.T) {
	always := []interface{}{"always"}
	multiline := []interface{}{"multiline"}
	const msg = "Missing line break around JSX"

	// Dimension 4 rows that do not apply to this rule (the rule only inspects
	// whole call/new arguments, never member access or declarations):
	//   - N/A: access / key forms (`.prop`, `['x']`, `#x`, computed keys) — the
	//     rule never reads a property name; it matches the argument node itself.
	//   - N/A: declaration / container forms (class/function/arrow/method,
	//     async/generator) — the rule targets call arguments, not declarations.
	//   - N/A: optional chain ON the argument — a JSX element can't be an
	//     optional-chain expression; the optional chain that matters is on the
	//     call (`fn?.(...)`), covered in the invalid cases below.

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxFunctionCallNewlineRule, []rule_tester.ValidTestCase{
		// ---- Locks in handleCallExpression() arm: arguments.length === 0 ----
		{Code: "fn()", Tsx: true},
		{Code: "fn()", Tsx: true, Options: always},
		{Code: "new Foo()", Tsx: true},
		{Code: "new Foo()", Tsx: true, Options: always},
		// ---- Locks in NewExpression with no argument list (`new Foo`) ----
		{Code: "new Foo", Tsx: true, Options: always},
		// ---- Locks in check() arm: !isJSX(node) → non-JSX args never report ----
		{Code: "fn(1, 'x', foo)", Tsx: true, Options: always},
		{Code: "fn({\n  a: 1\n})", Tsx: true, Options: always}, // multiline object arg, not JSX
		{Code: "fn(foo(\n))", Tsx: true, Options: always},      // multiline call arg, not JSX
		{Code: "fn(`a\nb`)", Tsx: true, Options: always},       // multiline template arg, not JSX
		// ---- Dimension 4: graceful degradation — spread argument ----
		{Code: "fn(...items)", Tsx: true, Options: always},
		// A spread's array operand is not a direct call argument, so its JSX is
		// never inspected (the rule must not descend into it).
		{Code: "fn(...[<div\n/>])", Tsx: true},
		// ---- Dimension 4: TS type-expression wrapper — `as` makes the arg a
		// TSAsExpression, not JSX (ESTree parity: arg node is TSAsExpression) ----
		{Code: "fn((<div />) as any)", Tsx: true, Options: always},
		{Code: "fn((<div\n/>) as any)", Tsx: true, Options: always}, // even multiline JSX inside `as` is skipped
		// ---- multiline (default): single-line JSX is never touched ----
		{Code: "fn(<></>)", Tsx: true},                             // single-line fragment
		{Code: "fn(<div>x</div>)", Tsx: true},                      // single-line element with children
		{Code: "fn(<div />)", Tsx: true, Options: multiline},       // explicit multiline, array shape
		{Code: "fn(<div />)", Tsx: true, Options: []interface{}{}}, // empty options array → default
		// ---- Dimension 4: parenthesized argument — newlines INSIDE the parens
		// satisfy the rule because positions come from the inner JSX node ----
		{Code: "fn((\n<div\n/>\n))", Tsx: true},
		{Code: "fn(((\n<div\n/>\n)))", Tsx: true}, // multi-level parens
		// ---- always: properly broken multi-arg call stays valid ----
		{Code: "fn(\n<></>\n,\n<div>x</div>\n)", Tsx: true, Options: always},
		// ---- JSX as a sub-expression is NOT the argument node, so it is never
		// checked — even when the wrapping expression spans lines. ESTree's arg
		// is the ConditionalExpression / BinaryExpression, not a JSXElement;
		// SkipParentheses doesn't unwrap `?:` or `&&`/`||`, so isJSX stays false. ----
		{Code: "fn(cond ? <div\n/> : null)", Tsx: true, Options: always},
		{Code: "fn(a && <div\n/>)", Tsx: true, Options: always},
		{Code: "fn(x, y || <div\n/>)", Tsx: true, Options: always},
		// ---- Multi-byte content must not break column math or cause misfires.
		// (The invalid `fn('日本語', <span>…)` case below pins the exact column.) ----
		{Code: "fn(<div>{'中文テキスト'}</div>)", Tsx: true},                   // single-line → skipped
		{Code: "fn(\n<div>{'日本語'}</div>\n)", Tsx: true, Options: always}, // properly broken
		// ---- A JSX child of a fragment/element is not a call argument, so the
		// inner <a/> here is never checked; only the outer fragment arg is. ----
		{Code: "fn(\n<>\n<a />\n</>\n)", Tsx: true, Options: always},
		// ---- Tagged template: the JSX sits in a template interpolation, not a
		// call argument. TaggedTemplateExpression is neither Call nor New, so it
		// is never visited (matches upstream) — multiline JSX here stays valid. ----
		{Code: "html`${<div\n/>}`", Tsx: true, Options: always},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: JSX fragment (KindJsxFragment), multiline ----
		{
			Code:   "fn(<>\n</>)",
			Tsx:    true,
			Output: []string{"fn(\n<>\n</>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 4},
			},
		},
		// ---- Dimension 4: JSX fragment, always ----
		{
			Code:    "fn(<></>)",
			Tsx:     true,
			Options: always,
			Output:  []string{"fn(\n<></>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 1, EndColumn: 9},
			},
		},
		// ---- Dimension 4: JSX element with children (KindJsxElement), multiline ----
		{
			Code:   "fn(<div>\n</div>)",
			Tsx:    true,
			Output: []string{"fn(\n<div>\n</div>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 7},
			},
		},
		// ---- Dimension 4: optional call `fn?.(...)` is still a CallExpression ----
		{
			Code:   "fn?.(<div\n/>)",
			Tsx:    true,
			Output: []string{"fn?.(\n<div\n/>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 6, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Dimension 4: nested call — only the inner call's JSX arg reports;
		// the outer call's CallExpression arg is not JSX (listener must not bleed) ----
		{
			Code:   "g(fn(<div\n/>))",
			Tsx:    true,
			Output: []string{"g(fn(\n<div\n/>\n))"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 6, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Dimension 4: parenthesized argument hugging the JSX — the inner
		// `(` is the previous token, so a same-line open still needs a break ----
		{
			Code:   "fn((<div\n/>))",
			Tsx:    true,
			Output: []string{"fn((\n<div\n/>\n))"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 5, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Dimension 4: multi-level parens unwrap to the innermost JSX ----
		{
			Code:   "fn(((<div\n/>)))",
			Tsx:    true,
			Output: []string{"fn(((\n<div\n/>\n)))"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 6, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Dimension 4: NewExpression + hugging parens + multiline JSX ----
		{
			Code:   "new Foo((<div\n/>))",
			Tsx:    true,
			Output: []string{"new Foo((\n<div\n/>\n))"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 10, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Deep nesting: the listener descends into a JSX expression
		// container, so a call nested inside JSX children still has its JSX arg
		// checked. Here the outer <div> arg is already broken (no report); only
		// the inner g(<b/>) call's multiline JSX arg reports. The two JSX nodes
		// nest, but the outer doesn't fire, so the single fix doesn't overlap. ----
		{
			Code:   "fn(\n<div>{g(<b\n/>)}</div>\n)",
			Tsx:    true,
			Output: []string{"fn(\n<div>{g(\n<b\n/>\n)}</div>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 2, Column: 9, EndLine: 3, EndColumn: 3},
			},
		},
		// ---- Deep nesting, both fire: the outer <div> arg and the inner g()'s
		// <b/> arg are both unbroken, so both report (2 errors). Their fix ranges
		// nest/overlap, so the fixer applies the outer fix first and defers the
		// inner to the next pass — matching ESLint's multi-pass convergence
		// (hence two Output steps). ----
		{
			Code: "fn(<div>{g(<b\n/>)}</div>)",
			Tsx:  true,
			Output: []string{
				"fn(\n<div>{g(<b\n/>)}</div>\n)",
				"fn(\n<div>{g(\n<b\n/>\n)}</div>\n)",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 11},
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 12, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Option shape: explicit "multiline" via single-element array ----
		{
			Code:    "fn(<div\n/>)",
			Tsx:     true,
			Options: multiline,
			Output:  []string{"fn(\n<div\n/>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Option shape: bare string "always" (single-option CLI form) ----
		{
			Code:    "fn(<div />)",
			Tsx:     true,
			Options: "always",
			Output:  []string{"fn(\n<div />\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 1, EndColumn: 11},
			},
		},
		// ---- Option shape: bare string "multiline" (single-option CLI form) ----
		{
			Code:    "fn(<div\n/>)",
			Tsx:     true,
			Options: "multiline",
			Output:  []string{"fn(\n<div\n/>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Locks in needsClosingNewLine() arm: endWithComma → false, so only
		// the opening break is added even though the JSX spans lines ----
		{
			Code:   "fn(<div\n/>, x)",
			Tsx:    true,
			Output: []string{"fn(\n<div\n/>, x)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Locks in needsClosingNewLine() final arm: next token (`)`) is on a
		// LATER line and is not a comma → return false (closing already fine),
		// so only the opening break is added ----
		{
			Code:   "fn(<div\n/>\n)",
			Tsx:    true,
			Output: []string{"fn(\n<div\n/>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Comment in the gap: getTokenBefore (node.Pos()) and getTokenAfter
		// (scanner) skip trivia like ESLint; the leading comment stays put and
		// only the JSX text is wrapped ----
		{
			Code:   "fn(/* c */ <div\n/>)",
			Tsx:    true,
			Output: []string{"fn(/* c */ \n<div\n/>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 12, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Real-user shape: testing-library render() with a multiline element ----
		{
			Code:   "render(<App\n  prop={1}\n/>)",
			Tsx:    true,
			Output: []string{"render(\n<App\n  prop={1}\n/>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 8, EndLine: 3, EndColumn: 3},
			},
		},
		// ---- Real-user shape: createPortal(jsx, container) — only the JSX arg
		// reports; the trailing container identifier is not JSX ----
		{
			Code:   "createPortal(<Modal>\n</Modal>, container)",
			Tsx:    true,
			Output: []string{"createPortal(\n<Modal>\n</Modal>, container)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 14, EndLine: 2, EndColumn: 9},
			},
		},
		// ---- tsgo callee/type-arg shapes: a generic call carries a
		// TypeArgumentList between the callee and `(`; node.Pos() of the first
		// arg is still the `(`-end, so the open position stays correct ----
		{
			Code:   "fn<number>(<div\n/>)",
			Tsx:    true,
			Output: []string{"fn<number>(\n<div\n/>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 12, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Member-access callee (PropertyAccessExpression) still triggers the
		// CallExpression listener; trailing non-JSX arg is left alone ----
		{
			Code:   "ReactDOM.render(<App\n/>, root)",
			Tsx:    true,
			Output: []string{"ReactDOM.render(\n<App\n/>, root)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 17, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- NewExpression with a qualified (member) callee ----
		{
			Code:   "new ns.Foo(<div\n/>)",
			Tsx:    true,
			Output: []string{"new ns.Foo(\n<div\n/>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 12, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Multi-byte: a CJK string arg precedes the JSX. The reported column
		// (11) counts each CJK char as one unit, matching ESLint's UTF-16 columns
		// (a byte-based count would land at ~18). Locks in column parity. ----
		{
			Code:   "fn('日本語', <span>\n</span>)",
			Tsx:    true,
			Output: []string{"fn('日本語', \n<span>\n</span>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 11, EndLine: 2, EndColumn: 8},
			},
		},
		// ---- JSX with multiple attributes spanning lines (real component shape) ----
		{
			Code:   "fn(<div\n a={1}\n/>)",
			Tsx:    true,
			Output: []string{"fn(\n<div\n a={1}\n/>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 3, EndColumn: 3},
			},
		},
		// ---- JSX with a spread attribute ({...p}) spanning lines ----
		{
			Code:   "fn(<div {...p}\n/>)",
			Tsx:    true,
			Output: []string{"fn(\n<div {...p}\n/>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Fragment with children spanning lines (the inner <a/> child is not
		// a call argument, so only the outer fragment reports) ----
		{
			Code:   "fn(<>\n<a />\n</>)",
			Tsx:    true,
			Output: []string{"fn(\n<>\n<a />\n</>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 3, EndColumn: 4},
			},
		},
		// ---- Trailing comma after a multiline JSX arg: endWithComma → no closing
		// break, only the opening one is added ----
		{
			Code:   "fn(<div\n/>,)",
			Tsx:    true,
			Output: []string{"fn(\n<div\n/>,)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Real-user shape: JSX argument inside a chained assertion call ----
		{
			Code:   "expect(<C\n/>).toBe(x)",
			Tsx:    true,
			Output: []string{"expect(\n<C\n/>\n).toBe(x)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 8, EndLine: 2, EndColumn: 3},
			},
		},
		// ---- Decorator factory `@deco(...)` is a CallExpression, so a JSX arg
		// inside it is checked like any other call ----
		{
			Code:   "@deco(<div\n/>)\nclass X {}",
			Tsx:    true,
			Output: []string{"@deco(\n<div\n/>\n)\nclass X {}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 7, EndLine: 2, EndColumn: 3},
			},
		},
	})
}
