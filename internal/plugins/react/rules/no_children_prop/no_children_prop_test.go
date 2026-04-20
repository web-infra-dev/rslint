package no_children_prop

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoChildrenPropRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoChildrenPropRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `<div />;`, Tsx: true},
		{Code: `<div></div>;`, Tsx: true},
		{Code: `React.createElement("div", {});`, Tsx: true},
		{Code: `React.createElement("div", undefined);`, Tsx: true},
		{Code: `<div className="class-name"></div>;`, Tsx: true},
		{Code: `React.createElement("div", {className: "class-name"});`, Tsx: true},
		{Code: `<div>Children</div>;`, Tsx: true},
		{Code: `React.createElement("div", "Children");`, Tsx: true},
		{Code: `React.createElement("div", {}, "Children");`, Tsx: true},
		{Code: `React.createElement("div", undefined, "Children");`, Tsx: true},
		{Code: `<div className="class-name">Children</div>;`, Tsx: true},
		{Code: `React.createElement("div", {className: "class-name"}, "Children");`, Tsx: true},
		{Code: `<div><div /></div>;`, Tsx: true},
		{Code: `React.createElement("div", React.createElement("div"));`, Tsx: true},
		{Code: `React.createElement("div", {}, React.createElement("div"));`, Tsx: true},
		{Code: `React.createElement("div", undefined, React.createElement("div"));`, Tsx: true},
		{Code: `<div><div /><div /></div>;`, Tsx: true},
		{Code: `React.createElement("div", React.createElement("div"), React.createElement("div"));`, Tsx: true},
		{Code: `React.createElement("div", {}, React.createElement("div"), React.createElement("div"));`, Tsx: true},
		{Code: `React.createElement("div", undefined, React.createElement("div"), React.createElement("div"));`, Tsx: true},
		{Code: `React.createElement("div", [React.createElement("div"), React.createElement("div")]);`, Tsx: true},
		{Code: `React.createElement("div", {}, [React.createElement("div"), React.createElement("div")]);`, Tsx: true},
		{Code: `React.createElement("div", undefined, [React.createElement("div"), React.createElement("div")]);`, Tsx: true},
		{Code: `<MyComponent />`, Tsx: true},
		{Code: `React.createElement(MyComponent);`, Tsx: true},
		{Code: `React.createElement(MyComponent, {});`, Tsx: true},
		{Code: `React.createElement(MyComponent, undefined);`, Tsx: true},
		{Code: `<MyComponent>Children</MyComponent>;`, Tsx: true},
		{Code: `React.createElement(MyComponent, "Children");`, Tsx: true},
		{Code: `React.createElement(MyComponent, {}, "Children");`, Tsx: true},
		{Code: `React.createElement(MyComponent, undefined, "Children");`, Tsx: true},
		{Code: `<MyComponent className="class-name"></MyComponent>;`, Tsx: true},
		{Code: `React.createElement(MyComponent, {className: "class-name"});`, Tsx: true},
		{Code: `<MyComponent className="class-name">Children</MyComponent>;`, Tsx: true},
		{Code: `React.createElement(MyComponent, {className: "class-name"}, "Children");`, Tsx: true},
		{Code: `<MyComponent className="class-name" {...props} />;`, Tsx: true},
		{Code: `React.createElement(MyComponent, {className: "class-name", ...props});`, Tsx: true},

		// ---- allowFunctions: functions passed as `children` prop are allowed ----
		{
			Code:    `<MyComponent children={() => {}} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		{
			Code:    `<MyComponent children={function() {}} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		{
			Code:    `<MyComponent children={async function() {}} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		{
			Code:    `<MyComponent children={function* () {}} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		{
			Code:    `React.createElement(MyComponent, {children: () => {}});`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		{
			Code:    `React.createElement(MyComponent, {children: function() {}});`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		{
			Code:    `React.createElement(MyComponent, {children: async function() {}});`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		{
			Code:    `React.createElement(MyComponent, {children: function* () {}});`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},

		// ---- Additional edge cases ----
		// Without allowFunctions, a function child (as JSX children) is fine —
		// the rule only complains about function children when the user opted in.
		{Code: `<MyComponent>{() => {}}</MyComponent>;`, Tsx: true},
		// A 3rd-arg function without allowFunctions is also fine.
		{Code: `React.createElement(MyComponent, {}, () => {});`, Tsx: true},
		// Non-React.createElement call with `children` in options doesn't match.
		{Code: `someOther.createElement("div", {children: "x"});`, Tsx: true},
		// A plain function call (not React.createElement) is ignored.
		{Code: `foo({children: "x"});`, Tsx: true},
		// String-keyed `"children"` is treated as an Identifier key by tsgo
		// (both parse to a PropertyName), which matches ESLint's behavior on
		// tsgo's normalized AST — see the invalid case below that locks this.
		// Spread-only second arg has no `children` key.
		{Code: `React.createElement("div", {...props});`, Tsx: true},
		// Computed key is not recognized as the children prop.
		{Code: `const k = "children"; React.createElement("div", {[k]: "x"});`, Tsx: true},
		// String-keyed `"children"` is NOT matched — upstream's `'name' in prop.key`
		// guard excludes StringLiteral keys.
		{Code: `React.createElement("div", {"children": "x"});`, Tsx: true},
		// Numeric key is also not matched.
		{Code: `React.createElement("div", {1: "x"});`, Tsx: true},
		// JSX Fragment child is ignored — the rule only listens to JsxElement.
		{
			Code:    `<>{() => {}}</>;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		// allowFunctions + non-function JSX child expression — no `nestFunction`.
		{
			Code:    `<MyComponent>{someValue}</MyComponent>;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		// JSXElement with multiple children never triggers `nestFunction`.
		{
			Code:    `<MyComponent>{() => {}}{() => {}}</MyComponent>;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		// Paren-wrapped function in `children={...}` is unwrapped and allowed.
		{
			Code:    `<MyComponent children={(() => {})} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		// Paren-wrapped function in a createElement children prop is allowed.
		{
			Code:    `React.createElement(MyComponent, {children: (() => {})});`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
		},
		// Custom pragma via `settings.react.pragma` — `h.createElement(...)` is
		// recognized, and `React.createElement(...)` is NOT (so the `children`
		// prop here is just a plain property access, not a React call).
		{
			Code: `React.createElement("div", {children: "x"});`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"pragma": "h"},
			},
		},
		// Parenthesized createElement callee is still recognized — so a
		// valid call with no `children` prop must still pass cleanly.
		{Code: `(React.createElement)("div", {});`, Tsx: true},
		// Parenthesized props object — still recognized; no children here.
		{Code: `React.createElement("div", ({className: "x"}));`, Tsx: true},
		// Generic createElement: `React.createElement<Props>(...)` — type args
		// don't change the call shape.
		{Code: `React.createElement<{}>("div", {});`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `<div children />;`, // not a valid use case but make sure we don't crash
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Message: msgNestChildren, Line: 1, Column: 6},
			},
		},
		{
			Code: `<div children="Children" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Message: msgNestChildren, Line: 1, Column: 6},
			},
		},
		{
			Code: `<div children={<div />} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div children={[<div />, <div />]} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div children="Children">Children</div>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Line: 1, Column: 6},
			},
		},
		{
			Code: `React.createElement("div", {children: "Children"});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Message: msgPassChildrenAsArgs, Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("div", {children: "Children"}, "Children");`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("div", {children: React.createElement("div")});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("div", {children: [React.createElement("div"), React.createElement("div")]});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		{
			Code: `<MyComponent children="Children" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Line: 1, Column: 14},
			},
		},
		{
			Code: `React.createElement(MyComponent, {children: "Children"});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		{
			Code: `<MyComponent className="class-name" children="Children" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Line: 1, Column: 37},
			},
		},
		{
			Code: `React.createElement(MyComponent, {children: "Children", className: "class-name"});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		{
			Code: `<MyComponent {...props} children="Children" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Line: 1, Column: 25},
			},
		},
		{
			Code: `React.createElement(MyComponent, {...props, children: "Children"})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<MyComponent>{() => {}}</MyComponent>;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestFunction", Message: msgNestFunction, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<MyComponent>{function() {}}</MyComponent>;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestFunction", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<MyComponent>{async function() {}}</MyComponent>;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestFunction", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<MyComponent>{function* () {}}</MyComponent>;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestFunction", Line: 1, Column: 1},
			},
		},
		{
			Code:    `React.createElement(MyComponent, {}, () => {});`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passFunctionAsArgs", Message: msgPassFunctionAsArgs, Line: 1, Column: 1},
			},
		},
		{
			Code:    `React.createElement(MyComponent, {}, function() {});`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passFunctionAsArgs", Line: 1, Column: 1},
			},
		},
		{
			Code:    `React.createElement(MyComponent, {}, async function() {});`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passFunctionAsArgs", Line: 1, Column: 1},
			},
		},
		{
			Code:    `React.createElement(MyComponent, {}, function* () {});`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passFunctionAsArgs", Line: 1, Column: 1},
			},
		},

		// ---- Additional edge cases ----

		// Shorthand `{children}` is still a children-prop reference — reported.
		{
			Code: `const children = "x"; React.createElement("div", {children});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 23},
			},
		},
		// Multi-line JSX attribute position — reported at the attribute.
		{
			Code: "<div\n\tchildren=\"x\"\n/>;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Line: 2, Column: 2, EndLine: 2, EndColumn: 14},
			},
		},
		// Without allowFunctions, a function literal as children prop still
		// reports (the exemption requires the opt-in).
		{
			Code: `<MyComponent children={() => {}} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Line: 1, Column: 14},
			},
		},
		{
			Code: `React.createElement(MyComponent, {children: () => {}});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		// With allowFunctions, a non-function value (e.g. variable reference)
		// still reports.
		{
			Code:    `const fn = () => {}; <MyComponent children={fn} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Line: 1, Column: 35},
			},
		},
		// Parenthesized second arg — still unwrapped and inspected for children.
		{
			Code: `React.createElement("div", ({children: "x"}));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		// Parenthesized createElement callee — IsCreateElementCall unwraps it.
		{
			Code: `(React.createElement)("div", {children: "x"});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		// Custom pragma: `h.createElement(...)` is picked up when settings
		// declares it.
		{
			Code: `h.createElement("div", {children: "x"});`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"pragma": "h"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		// Nested createElement: the outer `{children: ...}` fires; the inner
		// call has its own props with no `children` so it does not fire.
		{
			Code: `React.createElement("div", {children: React.createElement("span", {})});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 1},
			},
		},
		// Deeply nested JSX: only the innermost `<Inner>` with a function
		// child triggers `nestFunction`; outer elements have element (not
		// expression) children and are unaffected.
		{
			Code:    `<Outer><Mid><Inner>{() => {}}</Inner></Mid></Outer>;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestFunction", Line: 1, Column: 13, EndLine: 1, EndColumn: 38},
			},
		},
		// Multiple createElement calls on one line: each is considered
		// independently and the one with a `children` prop reports.
		{
			Code: `React.createElement("a", {}); React.createElement("b", {children: "x"});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "passChildrenAsArgs", Line: 1, Column: 31},
			},
		},
		// Member-based user component `<Foo.Bar>` — `children` prop still reports.
		{
			Code: `const Foo: any = {}; Foo.Bar = () => null; const x = <Foo.Bar children="x" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nestChildren", Line: 1, Column: 63},
			},
		},
	})
}
