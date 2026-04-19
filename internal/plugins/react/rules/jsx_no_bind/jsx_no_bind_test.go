package jsx_no_bind

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxNoBindRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoBindRule, []rule_tester.ValidTestCase{
		// ---------- 1. Plain identifiers & non-banned expressions ----------
		{Code: `var x = <div onClick={this._handleClick}></div>`, Tsx: true},
		{Code: `var x = <Foo onClick={this._handleClick} />`, Tsx: true},
		{Code: `var x = <div meaningOfLife={42}></div>`, Tsx: true},
		{Code: `var x = <div onClick={getHandler()}></div>`, Tsx: true},
		{Code: `var x = <div onClick={obj.method}></div>`, Tsx: true},
		// `foo()` (not `.bind()`) is not flagged
		{Code: `var x = <div onClick={foo()}></div>`, Tsx: true},
		// Bracket-access bind (ESLint requires `.bind`, not `['bind']`)
		{Code: `var x = <div onClick={foo['bind'](this)}></div>`, Tsx: true},
		// Logical expressions are not checked (matches ESLint)
		{Code: `var x = <div onClick={cond && handler}></div>`, Tsx: true},
		{Code: `var x = <div onClick={cond || handler}></div>`, Tsx: true},
		{Code: `var x = <div onClick={handler ?? fallback}></div>`, Tsx: true},
		// JSX spread attribute is not checked
		{Code: `var x = <div {...props}></div>`, Tsx: true},
		// JSX child expression is not checked
		{Code: `var x = <div>{() => 1}</div>`, Tsx: true},

		// ---------- 2. ignoreRefs ----------
		{Code: `var x = <div ref={c => (this._input = c)}></div>`, Tsx: true, Options: map[string]interface{}{"ignoreRefs": true}},
		{Code: `var x = <div ref={this._refCallback.bind(this)}></div>`, Tsx: true, Options: map[string]interface{}{"ignoreRefs": true}},
		{Code: `var x = <div ref={function (c) { this._input = c; }}></div>`, Tsx: true, Options: map[string]interface{}{"ignoreRefs": true}},

		// ---------- 3. allow* options ----------
		{Code: `var x = <div onClick={this._handleClick.bind(this)}></div>`, Tsx: true, Options: map[string]interface{}{"allowBind": true}},
		{Code: `var x = <div onClick={() => alert("1337")}></div>`, Tsx: true, Options: map[string]interface{}{"allowArrowFunctions": true}},
		{Code: `var x = <div onClick={async () => alert("1337")}></div>`, Tsx: true, Options: map[string]interface{}{"allowArrowFunctions": true}},
		{Code: `var x = <div onClick={function () { alert("1337"); }}></div>`, Tsx: true, Options: map[string]interface{}{"allowFunctions": true}},
		{Code: `var x = <div onClick={function* () { alert("1337"); }}></div>`, Tsx: true, Options: map[string]interface{}{"allowFunctions": true}},
		{Code: `var x = <div onClick={async function () { alert("1337"); }}></div>`, Tsx: true, Options: map[string]interface{}{"allowFunctions": true}},

		// ---------- 4. ignoreDOMComponents ----------
		{Code: `var x = <div onClick={this._handleClick.bind(this)}></div>`, Tsx: true, Options: map[string]interface{}{"ignoreDOMComponents": true}},
		{Code: `var x = <div onClick={() => alert("1337")}></div>`, Tsx: true, Options: map[string]interface{}{"ignoreDOMComponents": true}},
		{Code: `var x = <div onClick={function () { alert("1337"); }}></div>`, Tsx: true, Options: map[string]interface{}{"ignoreDOMComponents": true}},
		// Namespaced DOM element is also intrinsic
		{Code: `var x = <svg:path onClick={() => 1} />`, Tsx: true, Options: map[string]interface{}{"ignoreDOMComponents": true}},
		// Property-access tag with a lowercase base — ESLint's isDOMComponent
		// tests the first character of elementType only, so `<foo.Bar>` is
		// classified as DOM and skipped under `ignoreDOMComponents: true`.
		{Code: `var x = <foo.Bar onClick={() => 1} />`, Tsx: true, Options: map[string]interface{}{"ignoreDOMComponents": true}},

		// ---------- 5. Bind-not-used-for-JSX-prop ----------
		{
			Code: `
				class Hello extends Component {
					render() {
						this.onTap.bind(this);
						return true;
					}
				}
			`,
			Tsx: true,
		},
		{
			Code: `
				class Hello extends Component {
					render() {
						const click = this.onTap.bind(this);
						return <div onClick={onClick}>Hello</div>;
					}
				};
			`,
			Tsx: true,
		},
		{
			Code: `
				class Hello extends Component {
					render() {
						return (<div>{
							this.props.list.map(this.wrap.bind(this, "span"))
						}</div>);
					}
				};
			`,
			Tsx: true,
		},
		// Arrow as map callback (a JSX child, not a prop)
		{
			Code: `
				class Hello extends Component {
					render() {
						return (<div>{
							this.props.list.map(item => <item hello="true"/>)
						}</div>);
					}
				};
			`,
			Tsx: true,
		},

		// ---------- 6. Non-const / non-identifier bindings are not tracked ----------
		{
			// let with arrow init — NOT tracked
			Code: `
				function C() {
					let click = () => 1;
					return <div onClick={click}>Hello</div>;
				}
			`,
			Tsx: true,
		},
		{
			// var with arrow init — NOT tracked
			Code: `
				function C() {
					var click = () => 1;
					return <div onClick={click}>Hello</div>;
				}
			`,
			Tsx: true,
		},
		{
			// Destructured const — name is not an Identifier, NOT tracked
			Code: `
				function C() {
					const { click } = obj;
					return <div onClick={click}>Hello</div>;
				}
			`,
			Tsx: true,
		},
		{
			// Uninitialized variable should not crash
			Code: `
				class Hello extends Component {
					render() {
						let click;
						return <div onClick={onClick}>Hello</div>;
					}
				}
			`,
			Tsx: true,
		},

		// ---------- 7. Top-level / cross-scope ----------
		{
			// Top-level function declarations are NOT tracked
			Code: `
				function click() { return true; }
				class Hello23 extends React.Component {
					renderDiv() {
						return <div onClick={click}>Hello</div>;
					}
				};
			`,
			Tsx: true,
		},
		{
			// for-of const binding is not tracked
			Code: `
				function C() {
					for (const handler of handlers) {
						return <div onClick={handler} />;
					}
				}
			`,
			Tsx: true,
		},
		{
			// for-in const binding is not tracked
			Code: `
				function C() {
					for (const key in handlers) {
						return <div onClick={handlers[key]} />;
					}
				}
			`,
			Tsx: true,
		},

		// ---------- 8. TS-only wrappers are opaque (matches ESLint behavior under
		// typescript-eslint; TSAsExpression/TSNonNullExpression/TSSatisfiesExpression
		// are distinct AST nodes the rule doesn't recurse into) ----------
		{
			// `as T` wrapping an arrow is not recognized → not flagged
			Code: `var x = <div onClick={(() => 1) as any}></div>`,
			Tsx:  true,
		},
		{
			// Non-null on an arrow is not recognized → not flagged
			Code: `var x = <div onClick={(() => 1)!}></div>`,
			Tsx:  true,
		},
		{
			// `satisfies T` wrapping an arrow is not recognized → not flagged
			Code: `var x = <div onClick={(() => 1) satisfies any}></div>`,
			Tsx:  true,
		},
		{
			// `as T` on the result of .bind() is opaque → not flagged
			Code: `var x = <div onClick={this.h.bind(this) as any}></div>`,
			Tsx:  true,
		},
		{
			// Non-null on an identifier of a tracked arrow is opaque → not flagged
			Code: `
function C() {
	const handler = () => 1;
	return <div onClick={handler!}>Hello</div>;
}
			`,
			Tsx: true,
		},

		// ---------- 9. Scope isolation across siblings ----------
		{
			// A const arrow in an unrelated function doesn't leak into siblings
			Code: `
				function A() {
					const click = () => 1;
					return true;
				}
				function B() {
					return <div onClick={click}>Hello</div>;
				}
			`,
			Tsx: true,
		},

		// ---------- 10. IIFE returning a function is a CallExpression result, not .bind ----------
		{
			// Value is the call's return value, not itself a violating expression
			Code: `var x = <div onClick={(() => handler)()}></div>`,
			Tsx:  true,
		},

		// ---------- 11. Async generators with allowFunctions ----------
		{
			Code:    `var x = <div onClick={async function* () { yield 1; }}></div>`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
	}, []rule_tester.InvalidTestCase{
		// ---------- 10. Direct violations ----------
		{
			Code: `var x = <div onClick={this._handleClick.bind(this)}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 1, Column: 14},
			},
		},
		{
			Code: `var x = <div onClick={someGlobalFunction.bind(this)}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 1, Column: 14},
			},
		},
		{
			Code: `var x = <div onClick={window.lol.bind(this)}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 1, Column: 14},
			},
		},
		{
			Code: `var x = <div ref={this._refCallback.bind(this)}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 1, Column: 14},
			},
		},
		{
			Code: `var x = <div onClick={() => alert("1337")}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 1, Column: 14},
			},
		},
		{
			Code: `var x = <div onClick={async () => alert("1337")}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 1, Column: 14},
			},
		},
		{
			Code: `var x = <div ref={c => (this._input = c)}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 1, Column: 14},
			},
		},
		{
			Code: `var x = <div onClick={function () { alert("1337"); }}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "func", Line: 1, Column: 14},
			},
		},
		{
			Code: `var x = <div onClick={function* () { alert("1337"); }}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "func", Line: 1, Column: 14},
			},
		},
		{
			Code: `var x = <div onClick={async function () { alert("1337"); }}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "func", Line: 1, Column: 14},
			},
		},

		// ---------- 11. Parentheses are invisible (ESTree parity) ----------
		{
			// Parenthesized arrow in attribute value
			Code: `var x = <div onClick={(() => 1)}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 1, Column: 14},
			},
		},
		{
			// Parenthesized .bind() callee — callee parens are stripped
			Code: `var x = <div onClick={(this.h.bind)(this)}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 1, Column: 14},
			},
		},
		{
			// Parenthesized const initializer
			Code: `
function C() {
	const click = (() => 1);
	return <div onClick={click}>Hello</div>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 4, Column: 14},
			},
		},

		// ---------- 12. Conditional expressions ----------
		{
			Code: `var x = <div onClick={cond ? onClick.bind(this) : handleClick()}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 1, Column: 14},
			},
		},
		{
			Code: `var x = <div onClick={cond ? handleClick() : this.onClick.bind(this)}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 1, Column: 14},
			},
		},
		{
			// Nested ternary — recursion finds the arrow in the inner WhenTrue
			Code: `var x = <div onClick={a ? (b ? () => 1 : c) : d}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 1, Column: 14},
			},
		},
		{
			// Ternary condition itself is a .bind()
			Code: `var x = <div onClick={returningBoolean.bind(this) ? handleClick() : onClick()}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 1, Column: 14},
			},
		},

		// ---------- 13. ignoreDOMComponents ----------
		{
			// User component is still flagged
			Code:    `var x = <Foo onClick={this._handleClick.bind(this)} />`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreDOMComponents": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 1, Column: 14},
			},
		},
		{
			// Property-access tag name with UPPERCASE base is a user component,
			// so `ignoreDOMComponents: true` still reports it.
			Code:    `var x = <Foo.Bar onClick={() => 1} />`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreDOMComponents": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 1, Column: 18},
			},
		},

		// ---------- 14. Variable tracking in function / method / arrow bodies ----------
		{
			Code: `
class Hello23 extends React.Component {
	render() {
		const click = this.someMethod.bind(this);
		return <div onClick={click}>Hello</div>;
	}
};
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 5, Column: 15},
			},
		},
		{
			Code: `
class Hello23 extends React.Component {
	render() {
		const click = () => true;
		return <div onClick={click}>Hello</div>;
	}
};
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 5, Column: 15},
			},
		},
		{
			Code: `
class Hello23 extends React.Component {
	render() {
		const click = function () { return true; };
		return <div onClick={click}>Hello</div>;
	}
};
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "func", Line: 5, Column: 15},
			},
		},
		{
			// Class-field arrow body is itself a Block scope
			Code: `
class Hello23 extends React.Component {
	renderDiv = () => {
		const click = this.doSomething.bind(this, "no");
		return <div onClick={click}>Hello</div>;
	}
};
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 5, Column: 15},
			},
		},
		{
			// Local FunctionDeclaration inside a method body
			Code: `
class Hello23 extends React.Component {
	renderDiv() {
		function click() { return true; }
		return <div onClick={click}>Hello</div>;
	}
};
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "func", Line: 5, Column: 15},
			},
		},

		// ---------- 15. Nested blocks & innermost-first resolution ----------
		{
			// Inner block shadows outer: innermost hit wins (bindCall for inner,
			// arrowFunc for outer)
			Code: `
class Hello23 extends React.Component {
	renderDiv() {
		const click = () => true;
		const renderStuff = () => {
			const click = this.doSomething.bind(this, "hey");
			return <div onClick={click} />;
		};
		return <div onClick={click}>Hello</div>;
	}
};
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 7, Column: 16},
				{MessageId: "arrowFunc", Line: 9, Column: 15},
			},
		},
		{
			// Closures across function boundaries: outer decl is found from inner body
			Code: `
function outer() {
	const click = () => 1;
	function inner() {
		return <div onClick={click}>Hello</div>;
	}
	inner();
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 5, Column: 15},
			},
		},
		{
			// If-block as an independent scope
			Code: `
function C() {
	if (cond) {
		const handler = () => 1;
		return <div onClick={handler}>Hello</div>;
	}
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 5, Column: 15},
			},
		},
		{
			// For-loop body as its own block
			Code: `
function C() {
	for (let i = 0; i < 1; i++) {
		const handler = () => 1;
		return <div onClick={handler}>Hello</div>;
	}
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 5, Column: 15},
			},
		},

		// ---------- 16. JSX nested inside JSX — multiple reports ----------
		{
			// Outer arrow prop is flagged AND inner bind prop is flagged
			Code: `var x = <Outer onClick={() => <Inner baz={fn.bind(this)} />}></Outer>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 1, Column: 16},
				{MessageId: "bindCall", Line: 1, Column: 38},
			},
		},

		// ---------- 17. Multiple VariableDeclarations in one list ----------
		{
			Code: `
function C() {
	const a = () => 1, b = this.x.bind(this);
	return <><div onClick={a} /><div onClick={b} /></>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 4, Column: 16},
				{MessageId: "bindCall", Line: 4, Column: 35},
			},
		},

		// ---------- 18. Non-violating inner shadow still flags outer (ESLint parity) ----------
		{
			// Inner `const click = handler;` (non-violating) does NOT override
			// the outer `click = () => 1` tracking — matches ESLint's behavior.
			Code: `
function C() {
	const click = () => 1;
	if (cond) {
		const click = normalHandler;
		return <div onClick={click}>Hello</div>;
	}
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 6, Column: 15},
			},
		},

		// ---------- 19. ref without ignoreRefs — flagged like any other prop ----------
		{
			Code: `var x = <div ref={() => 1}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 1, Column: 14},
			},
		},

		// ---------- 20. Optional chain `?.bind()` ----------
		{
			Code: `var x = <div onClick={foo?.bind(this)}></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bindCall", Line: 1, Column: 14},
			},
		},

		// ---------- 21. Const initialized with a conditional whose branch violates ----------
		{
			// First non-empty violation from the ternary is the tracked kind.
			Code: `
function C() {
	const click = cond ? () => 1 : other;
	return <div onClick={click} />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 4, Column: 14},
			},
		},

		// ---------- 22. catch-clause body is its own block ----------
		{
			Code: `
function C() {
	try {} catch (err) {
		const handler = () => 1;
		return <div onClick={handler} />;
	}
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 5, Column: 15},
			},
		},

		// ---------- 23. Object method shorthand body is a tracked block ----------
		{
			Code: `
const obj = {
	render() {
		const handler = () => 1;
		return <div onClick={handler} />;
	}
};
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 5, Column: 15},
			},
		},

		// ---------- 24. Constructor body is a tracked block ----------
		{
			Code: `
class C {
	constructor() {
		const handler = () => 1;
		this.el = <div onClick={handler} />;
	}
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 5, Column: 18},
			},
		},

		// ---------- 25. Getter body tracked ----------
		{
			Code: `
class C {
	get el() {
		const handler = () => 1;
		return <div onClick={handler} />;
	}
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 5, Column: 15},
			},
		},

		// ---------- 26. Raw (free-standing) block creates its own scope ----------
		{
			Code: `
function C() {
	{
		const handler = () => 1;
		return <div onClick={handler} />;
	}
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 5, Column: 15},
			},
		},

		// ---------- 27. Switch case with braced block ----------
		{
			Code: `
function C(x) {
	switch (x) {
		case 1: {
			const handler = () => 1;
			return <div onClick={handler} />;
		}
	}
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "arrowFunc", Line: 6, Column: 16},
			},
		},
	})
}
