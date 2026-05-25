// cspell:ignore Apppp Appp appp

package jsx_no_undef

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxNoUndefRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoUndefRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `var React, App; React.render(<App />);`, Tsx: true},
		{Code: `var React; React.render(<img />);`, Tsx: true},
		{Code: `var React; React.render(<x-gif />);`, Tsx: true},
		{Code: `var React, app; React.render(<app.Foo />);`, Tsx: true},
		{Code: `var React, app; React.render(<app.foo.Bar />);`, Tsx: true},
		{Code: `var React; React.render(<Apppp:Foo />);`, Tsx: true},
		{
			Code: `
				var React;
				class Hello extends React.Component {
					render() {
						return <this.props.tag />
					}
				}
			`,
			Tsx: true,
		},
		// SKIP: rslint does not support ESLint's `globals` option.
		{
			Code: `var React, Text; React.render(<Text />);`,
			Tsx:  true,
			Skip: true,
		},
		{
			Code: `
				import Text from "cool-module";
				const TextWrapper = function (props) {
					return (
						<Text />
					);
				};
			`,
			Tsx:     true,
			Options: map[string]interface{}{"allowGlobals": false},
		},

		// ---- ThisKeyword forms ----
		// Bare `<this />` — ThisKeyword tag, not an Identifier reference.
		{Code: `var x = <this />;`, Tsx: true},
		// Nested member on `this`.
		{Code: `var x = <this.Foo />;`, Tsx: true},
		{Code: `var x = <this.a.b.c />;`, Tsx: true},

		// ---- DOM / lowercase shapes ----
		// Hyphenated lowercase tag.
		{Code: `var x = <foo-bar />;`, Tsx: true},
		// Web-component-style tag with digits.
		{Code: `var x = <my-elem-2 />;`, Tsx: true},

		// ---- Member tags with declared base ----
		{Code: `const foo: any = {}; var x = <foo.Bar />;`, Tsx: true},
		// Leftmost is only validated once — deeply nested is fine when the base resolves.
		{Code: `const a: any = {}; var x = <a.b.c.d.e.f />;`, Tsx: true},

		// ---- Declaration kinds recognized by IsShadowed ----
		{Code: `const _Foo = () => null; var x = <_Foo />;`, Tsx: true},
		{Code: `let Foo: any; var x = <Foo />;`, Tsx: true},
		{Code: `const Foo: any = null; var x = <Foo />;`, Tsx: true},
		{Code: `class Foo {} var x = <Foo />;`, Tsx: true},
		{Code: `function Foo() { return null; } var x = <Foo />;`, Tsx: true},
		{Code: `enum Foo { A } var x = <Foo />;`, Tsx: true},
		{Code: `namespace Foo { export const Bar: any = null; } var x = <Foo.Bar />;`, Tsx: true},
		{Code: `import { Foo } from "cool-module"; var x = <Foo />;`, Tsx: true},
		{Code: `import * as NS from "cool-module"; var x = <NS.Foo />;`, Tsx: true},
		{Code: `import Foo from "cool-module"; var x = <Foo />;`, Tsx: true},

		// ---- Nested scopes: declaration in an outer function is visible ----
		{
			Code: `function render() { const App = () => null; return <App />; }`,
			Tsx:  true,
		},
		{
			Code: `function outer() { const App: any = null; return function inner() { return <App />; }; }`,
			Tsx:  true,
		},
		// Parameter binding for the leftmost identifier of a member tag.
		{Code: `function render(app: any) { return <app.Foo />; }`, Tsx: true},
		// Catch-clause variable.
		{Code: `try {} catch (e: any) { var x = <e.Foo />; }`, Tsx: true},
		// for-let loop init.
		{Code: `for (let Foo: any; false; ) { var x = <Foo />; }`, Tsx: true},
		// for-of let.
		{
			Code: `const xs: any[] = []; for (const Foo of xs) { var x = <Foo />; }`,
			Tsx:  true,
		},

		// ---- Fragment is not a tag name ----
		{Code: `var x = <></>;`, Tsx: true},
		{Code: `var x = <><div /></>;`, Tsx: true},

		// ---- JSX appearing in non-top positions still resolves correctly ----
		// JSX as an attribute value; the outer tag's declared base passes.
		{Code: `const Outer: any = null; const Inner: any = null; var x = <Outer attr={<Inner />} />;`, Tsx: true},
		// JSX in ternary branches.
		{Code: `const A: any = null; const B: any = null; var x = true ? <A /> : <B />;`, Tsx: true},
		// JSX passed as a function argument.
		{Code: `const f: any = null; const A: any = null; f(<A />);`, Tsx: true},
		// JSX as the value of an object literal property.
		{Code: `const A: any = null; var x = { k: <A /> };`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `var React; React.render(<App />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'App' is not defined.", Line: 1, Column: 26, EndLine: 1, EndColumn: 29},
			},
		},
		{
			Code: `var React; React.render(<Appp.Foo />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'Appp' is not defined.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30},
			},
		},
		{
			Code: `var React; React.render(<appp.Foo />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'appp' is not defined.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30},
			},
		},
		{
			Code: `var React; React.render(<appp.foo.Bar />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'appp' is not defined.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30},
			},
		},
		{
			Code: `
				const TextWrapper = function (props) {
					return (
						<Text />
					);
				};
				export default TextWrapper;
			`,
			Tsx:     true,
			Options: map[string]interface{}{"allowGlobals": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'Text' is not defined."},
			},
		},
		{
			Code: `var React; React.render(<Foo />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'Foo' is not defined.", Line: 1, Column: 26, EndLine: 1, EndColumn: 29},
			},
		},

		// ---- Additional edge cases ----
		// Paired (non-self-closing) element — JsxOpeningElement listener fires.
		{
			Code: `<Bar></Bar>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'Bar' is not defined.", Line: 1, Column: 2, EndLine: 1, EndColumn: 5},
			},
		},
		// Multi-line: position points at the tag-name identifier.
		{
			Code: "var React;\nReact.render(\n\t<Missing />\n);",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'Missing' is not defined.", Line: 3, Column: 3, EndLine: 3, EndColumn: 10},
			},
		},
		// `allowGlobals: true` is accepted but has no effect in rslint —
		// undeclared names are still reported. Documented as a difference
		// from ESLint.
		{
			Code:    `var x = <Undeclared />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowGlobals": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'Undeclared' is not defined.", Line: 1, Column: 10, EndLine: 1, EndColumn: 20},
			},
		},
		// Member tag: only the leftmost identifier is reported, not the whole
		// dotted path.
		{
			Code: `var x = <Outer.Inner />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'Outer' is not defined.", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
			},
		},
		// Leading-underscore user component still reports when undeclared.
		{
			Code: `var x = <_Undeclared />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'_Undeclared' is not defined.", Line: 1, Column: 10, EndLine: 1, EndColumn: 21},
			},
		},

		// ---- Complex positional cases ----
		// Both the outer attribute-value tag and the nested child are checked —
		// two independent reports.
		{
			Code: `const Outer: any = null; var x = <Outer attr={<InnerBad />}><ChildBad /></Outer>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'InnerBad' is not defined.", Line: 1, Column: 48, EndLine: 1, EndColumn: 56},
				{MessageId: "undefined", Message: "'ChildBad' is not defined.", Line: 1, Column: 62, EndLine: 1, EndColumn: 70},
			},
		},
		// Ternary — both branches are checked when neither is declared.
		{
			Code: `var x = true ? <A /> : <B />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'A' is not defined.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				{MessageId: "undefined", Message: "'B' is not defined.", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
			},
		},
		// JSX in Fragment children — only the inner tag is checked (Fragment
		// itself has no tag name).
		{
			Code: `var x = <><Undef /></>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'Undef' is not defined.", Line: 1, Column: 12, EndLine: 1, EndColumn: 17},
			},
		},
		// Declaration in a sibling scope does NOT leak — inner scope cannot see
		// a declaration that only exists in a parallel function body.
		{
			Code: `function other() { const App: any = null; } function render() { return <App />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'App' is not defined.", Line: 1, Column: 73, EndLine: 1, EndColumn: 76},
			},
		},
		// catch-clause variable `e` is scoped to the catch body only — using
		// it after the `try/catch` is an unresolved reference.
		{
			Code: `try {} catch (e: any) {} var x = <e.Foo />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'e' is not defined.", Line: 1, Column: 35, EndLine: 1, EndColumn: 36},
			},
		},
		// for-let loop binding is scoped to the loop — using it after the loop
		// is an unresolved reference.
		{
			Code: `for (let Foo: any; false; ) {} var x = <Foo />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "undefined", Message: "'Foo' is not defined.", Line: 1, Column: 41, EndLine: 1, EndColumn: 44},
			},
		},
	})
}
