package no_danger_with_children

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDangerWithChildrenRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDangerWithChildrenRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `<div>Children</div>;`, Tsx: true},
		{Code: `const props: any = {}; <div {...props} />;`, Tsx: true},
		{Code: `<div dangerouslySetInnerHTML={{ __html: "HTML" }} />;`, Tsx: true},
		{Code: `<div children="Children" />;`, Tsx: true},
		{
			Code: `
				const props = { dangerouslySetInnerHTML: { __html: "HTML" } };
				const x = <div {...props} />;
			`,
			Tsx: true,
		},
		{
			Code: `
				const moreProps = { className: "eslint" };
				const props = { children: "Children", ...moreProps };
				const x = <div {...props} />;
			`,
			Tsx: true,
		},
		{
			Code: `
				const otherProps = { children: "Children" };
				const { a, b, ...props } = otherProps as any;
				const x = <div {...props} />;
			`,
			Tsx: true,
		},
		{Code: `<Hello>Children</Hello>;`, Tsx: true},
		{Code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }} />;`, Tsx: true},
		{
			Code: `
				<Hello dangerouslySetInnerHTML={{ __html: "HTML" }}>
				</Hello>;
			`,
			Tsx: true,
		},
		{Code: `React.createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } });`, Tsx: true},
		{Code: `React.createElement("div", {}, "Children");`, Tsx: true},
		{Code: `React.createElement("Hello", { dangerouslySetInnerHTML: { __html: "HTML" } });`, Tsx: true},
		{Code: `React.createElement("Hello", {}, "Children");`, Tsx: true},
		{Code: `<Hello {...undefined}>Children</Hello>;`, Tsx: true},
		{Code: `React.createElement("Hello", undefined, "Children");`, Tsx: true},
		{
			Code: `
				declare const shallow: any;
				declare const TaskEditableTitle: any;
				const props = { ...props, scratch: { mode: 'edit' } } as any;
				const component = shallow(<TaskEditableTitle {...props} />);
			`,
			Tsx: true,
		},

		// ---- Additional edge cases ----
		// Empty element, no props, no children.
		{Code: `<div />;`, Tsx: true},
		// Whitespace-only multi-line children with only dangerouslySetInnerHTML — not reported
		// because upstream treats a newline-only first child as a line break.
		{
			Code: `
				<div dangerouslySetInnerHTML={{ __html: "HTML" }}>
				</div>;
			`,
			Tsx: true,
		},
		// dangerouslySetInnerHTML is case-sensitive (lowercase h doesn't trigger).
		{Code: `<div dangerouslySetInnerHtml={{ __html: "HTML" }}>Children</div>;`, Tsx: true},
		// Bare createElement (destructured import) — upstream only inspects
		// `x.createElement(...)` member calls, so the bare form is not checked.
		{Code: `createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children");`, Tsx: true},
		// Computed createElement access — upstream's `'name' in property`
		// guard skips computed keys, so `x["createElement"](...)` is ignored.
		{Code: `React["createElement"]("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children");`, Tsx: true},
		// Call with fewer than 2 args.
		{Code: `React.createElement("div");`, Tsx: true},
		// Children-only prop, no dangerous — valid.
		{Code: `<div children="x">also</div>;`, Tsx: true},
		// Empty JSX text between tags (only whitespace on same line) counts as empty.
		{
			Code: `
				<div dangerouslySetInnerHTML={{ __html: "HTML" }}>

				</div>;
			`,
			Tsx: true,
		},
		// String-literal key (`"children": "x"`) matches neither upstream's
		// `prop.key.name === propName` nor our Identifier-only check — so a
		// spread of an object that carries a string-literal children key
		// is NOT treated as children. Locks parity with upstream.
		{
			Code: `
				const props = { "children": "Children", dangerouslySetInnerHTML: { __html: "HTML" } };
				const x = <div {...props} />;
			`,
			Tsx: true,
		},
		// Inline object spread (`{...{...}}`) isn't followed by upstream
		// (upstream checks `attribute.argument.name`, which only exists on
		// Identifier arguments). We intentionally keep parity — treated as
		// opaque, so nothing fires.
		{
			Code: `<div {...{ dangerouslySetInnerHTML: { __html: "HTML" }, children: "x" }} />;`,
			Tsx:  true,
		},
		// Self-closing JSX with dangerouslySetInnerHTML alone and no children attribute.
		{Code: `<Foo dangerouslySetInnerHTML={{ __html: "HTML" }} />;`, Tsx: true},
		// createElement with props as a parenthesized object literal — the
		// paren wrapper must be transparent.
		{Code: `React.createElement("div", ({ dangerouslySetInnerHTML: { __html: "HTML" } }));`, Tsx: true},
		// createElement where the 2nd arg is an Identifier that does NOT
		// resolve to an object literal (e.g. declared but never initialized).
		{
			Code: `
				let props: any;
				React.createElement("div", props, "Children");
			`,
			Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `
				<div dangerouslySetInnerHTML={{ __html: "HTML" }}>
					Children
				</div>;
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dangerWithChildren",
					Message:   "Only set one of `children` or `props.dangerouslySetInnerHTML`",
					Line:      2,
					Column:    5,
				},
			},
		},
		{
			Code: `<div dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 1, Column: 1},
			},
		},
		{
			Code: `
				const props = { dangerouslySetInnerHTML: { __html: "HTML" } };
				const x = <div {...props}>Children</div>;
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 3},
			},
		},
		{
			Code: `
				const props = { children: "Children", dangerouslySetInnerHTML: { __html: "HTML" } };
				const x = <div {...props} />;
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 3},
			},
		},
		{
			Code: `
				<Hello dangerouslySetInnerHTML={{ __html: "HTML" }}>
					Children
				</Hello>;
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 2, Column: 5},
			},
		},
		{
			Code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 1, Column: 1},
			},
		},
		{
			Code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }}> </Hello>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 1, Column: 1},
			},
		},
		{
			Code: `
				React.createElement(
					"div",
					{ dangerouslySetInnerHTML: { __html: "HTML" } },
					"Children"
				);
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 2},
			},
		},
		{
			Code: `
				React.createElement(
					"div",
					{
						dangerouslySetInnerHTML: { __html: "HTML" },
						children: "Children",
					}
				);
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 2},
			},
		},
		{
			Code: `
				React.createElement(
					"Hello",
					{ dangerouslySetInnerHTML: { __html: "HTML" } },
					"Children"
				);
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 2},
			},
		},
		{
			Code: `
				React.createElement(
					"Hello",
					{
						dangerouslySetInnerHTML: { __html: "HTML" },
						children: "Children",
					}
				);
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 2},
			},
		},
		{
			Code: `
				const props = { dangerouslySetInnerHTML: { __html: "HTML" } };
				React.createElement("div", props, "Children");
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 3},
			},
		},
		{
			Code: `
				const props = { children: "Children", dangerouslySetInnerHTML: { __html: "HTML" } };
				React.createElement("div", props);
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 3},
			},
		},
		{
			Code: `
				const moreProps = { children: "Children" };
				const otherProps = { ...moreProps };
				const props = { ...otherProps, dangerouslySetInnerHTML: { __html: "HTML" } };
				React.createElement("div", props);
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 5},
			},
		},

		// ---- Additional edge cases ----
		// JSX expression container as children (not a JsxText) — counts as children.
		{
			Code: `const x = <div dangerouslySetInnerHTML={{ __html: "HTML" }}>{children}</div>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 1, Column: 11},
			},
		},
		// Any `x.createElement(...)` member call is inspected — upstream
		// doesn't restrict to the React pragma, so non-React objects must
		// also trigger.
		{
			Code: `declare const createFoo: any; createFoo.createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children");`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 1, Column: 31},
			},
		},
		// Nested-member callee object (Preact-style `h.createElement` via
		// alias) — still a MemberExpression with property name `createElement`,
		// so upstream flags it.
		{
			Code: `declare const lib: { h: { createElement: (...a: any[]) => any } }; lib.h.createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children");`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren"},
			},
		},
		// Full range assertion on self-closing — covers the entire element.
		{
			Code: `<div dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dangerWithChildren",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 73,
				},
			},
		},
		// Shorthand property in the resolved spread target: the upstream
		// `prop.key.name === propName` check matches shorthand bindings just
		// like regular Identifier keys.
		{
			Code: `
				const dangerouslySetInnerHTML = { __html: "HTML" };
				const children = "Children";
				const props = { dangerouslySetInnerHTML, children };
				const x = <div {...props} />;
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 5},
			},
		},
		// Three-level spread chain resolved through the TypeChecker.
		{
			Code: `
				const a = { dangerouslySetInnerHTML: { __html: "HTML" } };
				const b = { ...a };
				const c = { ...b };
				React.createElement("div", c, "Children");
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 5},
			},
		},
		// Parenthesized initializer must be transparent on the spread target.
		{
			Code: `
				const props = ({ dangerouslySetInnerHTML: { __html: "HTML" } });
				const x = <div {...props}>Children</div>;
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 3},
			},
		},
		// Self-closing JSX with children attr via spread + direct dangerously prop.
		{
			Code: `
				const extras = { children: "Children" };
				const x = <div dangerouslySetInnerHTML={{ __html: "HTML" }} {...extras} />;
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 3},
			},
		},
		// createElement where props is a parenthesized inline object.
		{
			Code: `
				React.createElement(
					"div",
					({ dangerouslySetInnerHTML: { __html: "HTML" }, children: "Children" })
				);
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 2},
			},
		},
		// Nested JSX: outer element has neither prop, inner element has both.
		// Only the inner should report.
		{
			Code: `
				const x = (
					<div>
						<span dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />
					</div>
				);
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 4},
			},
		},
		// Direct dangerously prop co-located with a spread whose target has
		// children — both paths must be inspected.
		{
			Code: `
				const extras = { children: "Children" };
				const x = <div {...extras} dangerouslySetInnerHTML={{ __html: "HTML" }} />;
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dangerWithChildren", Line: 3},
			},
		},
	})
}
