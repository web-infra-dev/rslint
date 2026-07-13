// TestJsxPropsNoSpreadMultiExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
//
// Dimension 4 rows that don't apply to this rule (it inspects JSX spread
// attribute expressions only):
//   - N/A: JSX attribute names, object/class keys, PrivateIdentifier, and
//     element access as a member key; the rule never reads prop names or keys.
//   - N/A: class / function declaration & container forms; the rule targets
//     JSX opening elements only.
//   - N/A: ancestor scope walks (getThisContainer / FindEnclosingScope); none
//     performed.
//   - N/A: RestElement in a binding pattern; JSX spread attributes are the
//     only spread/rest form inspected.
package jsx_props_no_spread_multi

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxPropsNoSpreadMultiExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxPropsNoSpreadMultiRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: TS expression wrappers do NOT match upstream's direct Identifier check ----
		{Code: `<App {...props!} {...props} />`, Tsx: true},
		{Code: `<App {...(props as any)} {...props} />`, Tsx: true},
		{Code: `type Props = {}; const props = {}; <App {...(props satisfies Props)} {...props} />`, Tsx: true},

		// ---- Dimension 4: non-Identifier spread expressions are ignored ----
		{Code: `<App {...this.props} {...this.props} />`, Tsx: true},
		{Code: `<App {...props.foo} {...props.foo} />`, Tsx: true},
		{Code: `<App {...props["foo"]} {...props["foo"]} />`, Tsx: true},
		{Code: `<App {...props[foo]} {...props[foo]} />`, Tsx: true},
		{Code: `<App {...getProps()} {...getProps()} />`, Tsx: true},
		{Code: `<App {...{ props }} {...{ props }} />`, Tsx: true},
		{Code: `<App {...props?.bag} {...props?.bag} />`, Tsx: true},
		{Code: `<App {...(props?.bag)} {...(props?.bag)} />`, Tsx: true},

		// ---- Dimension 4: graceful degradation with empty / non-spread attributes ----
		{Code: `<App />`, Tsx: true},
		{Code: `<App attr="x" />`, Tsx: true},
		{Code: `<App {...props} attr="x" />`, Tsx: true},
		{Code: `<App {...props} {...other} />`, Tsx: true},

		// ---- Dimension 2: per-element tracking does not bleed across nested elements ----
		{Code: `<Outer {...props}><Inner {...props} /></Outer>`, Tsx: true},
		{Code: `<><App {...props} /><App {...props} /></>`, Tsx: true},

		// ---- Dimension 4: identifier names are case-sensitive ----
		{Code: `<App {...props} {...Props} />`, Tsx: true},

		// ---- Real-user: issue #3962 "When Not To Use It" dynamic expression shape ----
		{Code: `<App {...buildProps()} {...buildProps()} />`, Tsx: true},
		// ---- Real-user: PR #3724 explicitly limits the rule to identifier prop bags ----
		{Code: `<App {...props.common} {...props.common} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: parenthesized identifiers match because ESTree flattens parens ----
		{
			Code: `<App {...(props)} {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 19),
			},
		},
		{
			Code: `<App {...((props))} {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 21),
			},
		},
		// Locks in upstream filter arm: ignored TS wrappers do not reset identifier tracking.
		{
			Code: `<App {...props} {...(props as any)} {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 37),
			},
		},
		// Locks in upstream filter arm: parenthesized identifier duplicates report before later direct duplicates.
		{
			Code: `<App {...props} {...(props)} {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 17),
				noMultiSpreadingError(1, 30),
			},
		},
		// Locks in upstream JSXOpeningElement listener arm: non-self-closing elements are checked.
		{
			Code: `<App {...props} {...props}></App>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMultiSpreading", Message: noMultiSpreadingMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 27},
			},
		},
		// Locks in upstream JSXOpeningElement listener arm: self-closing elements are checked.
		{
			Code: `<App {...props} {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 17),
			},
		},
		// ---- Dimension 4: tag-name forms do not affect spread tracking ----
		{
			Code: `type Props = {}; <App<Props> {...props} {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 41),
			},
		},
		{
			Code: `<UI.Button {...props} {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 23),
			},
		},
		{
			Code: `<svg:path {...props} {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 22),
			},
		},
		// Locks in upstream filter arm: ordinary JSX attributes do not reset duplicate tracking.
		{
			Code: `<App {...props} attr="x" {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 26),
			},
		},
		// Locks in upstream Set branch: duplicate tracking is per identifier name.
		{
			Code: `<App {...a} {...b} {...a} {...b} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 20),
				noMultiSpreadingError(1, 27),
			},
		},
		// Locks in upstream forEach branch: every duplicate after the first reports.
		{
			Code: `<App {...a} {...a} {...a} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 13),
				noMultiSpreadingError(1, 20),
			},
		},
		// Locks in upstream branch ordering: duplicates are reported in attribute order.
		{
			Code: `<App {...a} {...b} {...a} {...a} {...b} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 20),
				noMultiSpreadingError(1, 27),
				noMultiSpreadingError(1, 34),
			},
		},
		// ---- Dimension 2: nested duplicate reports only on the inner element ----
		{
			Code: `<Outer {...props}><Inner {...props} {...props} /></Outer>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 37),
			},
		},
		// ---- Dimension 2: nested elements maintain independent duplicate sets ----
		{
			Code: `<Outer {...props} {...props}><Inner {...props} {...props} /></Outer>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 19),
				noMultiSpreadingError(1, 48),
			},
		},
		// ---- Real-user: PR #3724 duplicated prop bag around an overriding prop ----
		{
			Code: `<Button {...props} disabled {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 29),
			},
		},
		// ---- Real-user: PR #3724 duplicate prop bags around explicit overrides ----
		{
			Code: `<FormField {...fieldProps} disabled={isDisabled} {...fieldProps} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 50),
			},
		},
		{
			Code: `<Route {...routeProps} element={<Page />} {...routeProps} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 43),
			},
		},
		// ---- Dimension 4: comments and trivia between attributes do not affect tracking ----
		{
			Code: `<App {...props} /* keep override */ {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				noMultiSpreadingError(1, 37),
			},
		},
		// ---- Dimension 4: multi-line position uses the second duplicate spread ----
		{
			Code: `<App
  {...props}
  attr="x"
  {...props}
/>`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noMultiSpreading", Message: noMultiSpreadingMessage, Line: 4, Column: 3, EndLine: 4, EndColumn: 13},
			},
		},
	})
}
