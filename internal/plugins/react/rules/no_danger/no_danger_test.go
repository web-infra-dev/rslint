package no_danger

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDangerRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDangerRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `<App />;`, Tsx: true},
		{Code: `<App dangerouslySetInnerHTML={{ __html: "" }} />;`, Tsx: true},
		{Code: `<div className="bar"></div>;`, Tsx: true},
		{
			Code: `<div className="bar"></div>;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"*"},
			},
		},
		{
			Code: `
				function App() {
					return <Title dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Home"},
			},
		},
		{
			Code: `
				function App() {
					return <TextMUI dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"MUI*"},
			},
		},

		// ---- Additional edge cases ----
		// Case-sensitive: only the exact `dangerouslySetInnerHTML` spelling is dangerous.
		{Code: `<div dangerouslySetInnerHtml={{ __html: "" }} />;`, Tsx: true},
		// Similar-looking but distinct attribute name.
		{Code: `<div innerHTML="<span />" />;`, Tsx: true},
		// Multiple patterns, none matching the user component.
		{
			Code: `<Widget dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Foo*", "Bar*"},
			},
		},
		// Empty `customComponentNames` means custom components are not checked.
		{
			Code: `<MyComponent dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{},
			},
		},
		// Options passed as a bare object (no array wrapper) — still works.
		{
			Code: `<MyComponent dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Other"},
			},
		},
		// Non-DOM React.Fragment-style tag is not checked by default.
		{Code: `<React.Fragment></React.Fragment>;`, Tsx: true},
		// Spread attribute alone on a DOM element is fine.
		{Code: `<div {...props} />;`, Tsx: true},
		// Single-character wildcard `?` — `MUI` pattern matches exactly 3 trailing chars,
		// so a 4-char suffix like `MUIx` must not match.
		{
			Code: `<TextMUIx dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Text???"},
			},
		},
		// Literal `.` in the pattern — exact-dot match only (no wildcard).
		{
			Code: `<Foo.Bar dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Foo.Baz"},
			},
		},
		// Member tag with UPPERCASE base is a user component — no pattern,
		// no report.
		{
			Code: `
				const Foo: any = {};
				Foo.Bar = () => null;
				function App() {
					return <Foo.Bar dangerouslySetInnerHTML={{ __html: "" }} />;
				}
			`,
			Tsx: true,
		},
		// Defensive: non-array `customComponentNames` is ignored silently.
		{
			Code: `<MyComponent dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": "MyComponent",
			},
		},
		// Defensive: non-string entries in the array are skipped.
		{
			Code: `<MyComponent dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{42, nil, false},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `<div dangerouslySetInnerHTML={{ __html: "" }}></div>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dangerousProp",
					Message:   "Dangerous property 'dangerouslySetInnerHTML' found",
					Line:      1,
					Column:    6,
				},
			},
		},
		{
			Code: `<App dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"*"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dangerousProp",
					Line:      1,
					Column:    6,
				},
			},
		},
		{
			Code: `
				function App() {
					return <Title dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Title"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dangerousProp",
					Line:      3,
					Column:    20,
				},
			},
		},
		{
			Code: `
				function App() {
					return <TextFoo dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"*Foo"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dangerousProp",
					Line:      3,
					Column:    22,
				},
			},
		},
		{
			Code: `
				function App() {
					return <FooText dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Foo*"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dangerousProp",
					Line:      3,
					Column:    22,
				},
			},
		},
		{
			Code: `
				function App() {
					return <TextMUI dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"*MUI"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dangerousProp",
					Line:      3,
					Column:    22,
				},
			},
		},
		{
			Code: `
				import type { ComponentProps } from "react";

				const Comp = "div";
				const Component = () => <></>;

				const NestedComponent = (_props: ComponentProps<"div">) => <></>;

				Component.NestedComponent = NestedComponent;

				function App() {
					return (
						<>
							<div dangerouslySetInnerHTML={{ __html: "<div>aaa</div>" }} />
							<Comp dangerouslySetInnerHTML={{ __html: "<div>aaa</div>" }} />

							<Component.NestedComponent
								dangerouslySetInnerHTML={{ __html: '<div>aaa</div>' }} />
						</>
					);
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"*"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 14},
				{MessageId: "dangerousProp", Line: 15},
				{MessageId: "dangerousProp", Line: 18},
			},
		},

		// ---- Additional edge cases ----

		// Boolean-shorthand attribute (no value) on a DOM element — still dangerous by name.
		{
			Code: `<div dangerouslySetInnerHTML />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 1, Column: 6},
			},
		},
		// Attribute on its own line — reported at the attribute position, not the tag's.
		{
			Code: "<div\n\tdangerouslySetInnerHTML={{ __html: '' }}\n/>;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 2, Column: 2},
			},
		},
		// Deeply-nested member tag (<A.B.C />) matched via `*` — the rule
		// uses raw source text for the tag name, so this works without a
		// hand-rolled member-expression flattener.
		{
			Code: `
				const A: any = {};
				A.B = { C: () => null };
				function App() {
					return <A.B.C dangerouslySetInnerHTML={{ __html: "" }} />;
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"*"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 5},
			},
		},
		// Any matching pattern in a multi-pattern list is enough.
		{
			Code: `<Widget dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Foo*", "Widget", "Bar*"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 1, Column: 9},
			},
		},
		// Spread attribute co-located with the dangerous prop — rule still fires
		// on the dangerous prop and leaves the spread alone.
		{
			Code: `<div {...props} dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 1, Column: 17},
			},
		},
		// `?` glob wildcard — matches exactly one character per `?`.
		{
			Code: `<TextMUI dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Text???"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 1, Column: 10},
			},
		},
		// Literal `.` in the pattern — exact-dot match on a member tag.
		{
			Code: `<Foo.Bar dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Foo.Bar"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 1, Column: 10},
			},
		},
		// Generic component `<Foo<T> ...>` — type arguments must not confuse
		// tag-name extraction.
		{
			Code: `
				function Foo<T>(_: { x: T; dangerouslySetInnerHTML?: unknown }) { return null; }
				function App() {
					return <Foo<string> dangerouslySetInnerHTML={{ __html: "" }} />;
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"Foo"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 4},
			},
		},
		// Full range assertion — EndLine/EndColumn cover the entire attribute node.
		{
			Code: `<div dangerouslySetInnerHTML={{ __html: "x" }} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dangerousProp",
					Line:      1,
					Column:    6,
					EndLine:   1,
					EndColumn: 47,
				},
			},
		},
		// JSX namespaced element `<svg:path>` — intrinsic (lowercase start), so DOM.
		{
			Code: `<svg:path dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 1, Column: 11},
			},
		},
		// Member tag with LOWERCASE base — `isDOMComponent` tests the first
		// character of `elementType(node)` only, so `<foo.bar>` classifies as
		// DOM and the rule fires without any customComponentNames entry.
		{
			Code: `
				const foo: any = {};
				foo.bar = () => null;
				function App() {
					return <foo.bar dangerouslySetInnerHTML={{ __html: "" }} />;
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 5},
			},
		},
		// `<this.Foo>` — `elementType` returns `"this.Foo"`, first char is
		// lowercase → DOM-classified by ESLint's regex. rslint matches.
		{
			Code: `
				class C {
					render() {
						return <this._Widget dangerouslySetInnerHTML={{ __html: "" }} />;
					}
				}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 4},
			},
		},
		// Locks "Differences from ESLint" #1: deep-member tag matched by its
		// full dotted name. eslint-plugin-react would compute `"undefined.C"`
		// here and NOT match either of these patterns; we use the source text
		// (`"A.B.C"`) so both `["A.B.C"]` and `["A.*"]` fire.
		{
			Code: `
				const A: any = {};
				A.B = { C: () => null };
				function App() {
					return <A.B.C dangerouslySetInnerHTML={{ __html: "" }} />;
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"A.B.C"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 5},
			},
		},
		{
			Code: `
				const A: any = {};
				A.B = { C: () => null };
				function App() {
					return <A.B.C dangerouslySetInnerHTML={{ __html: "" }} />;
				}
			`,
			Tsx: true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"A.*"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 5},
			},
		},
		// Locks "Differences from ESLint" #2: namespaced tag matched by an
		// explicit `"ns:name"` pattern. The element is already DOM (lowercase
		// first char) so it would be reported regardless — the point of this
		// case is that the custom-pattern branch can also match a namespaced
		// name, which eslint-plugin-react effectively cannot.
		{
			Code: `<svg:path dangerouslySetInnerHTML={{ __html: "" }} />;`,
			Tsx:  true,
			Options: map[string]interface{}{
				"customComponentNames": []interface{}{"svg:path"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerousProp", Line: 1, Column: 11},
			},
		},
	})
}
