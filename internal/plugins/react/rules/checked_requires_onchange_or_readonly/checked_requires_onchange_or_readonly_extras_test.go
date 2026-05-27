// TestCheckedRequiresOnchangeOrReadonlyExtras locks in branches and edge shapes
// that the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Dimension 4 rows that don't apply to this rule (it inspects JSX tag names,
// attribute names, and createElement string/object arguments — never a member
// receiver, class/function declaration, private field, or ancestor scope):
//   - N/A: JSX tag names cannot carry paren / non-null / `as` / optional-chain wrappers.
//   - N/A: element-access (`X['y']`) — the rule never reads dotted member access on user input.
//   - N/A: PrivateIdentifier (`#x`) keys — illegal in object literals.
//   - N/A: class / function declaration & container forms — the rule targets neither.
//   - N/A: ancestor scope walks (getThisContainer / FindEnclosingScope) — none performed.
//   - N/A: RestElement in a binding pattern — the rule never destructures (object SpreadAssignment IS covered below).
package checked_requires_onchange_or_readonly

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCheckedRequiresOnchangeOrReadonlyExtras(t *testing.T) {
	ignoreMissing := map[string]interface{}{"ignoreMissingProperties": true}
	ignoreExclusive := map[string]interface{}{"ignoreExclusiveCheckedAttribute": true}
	pragmaFoo := map[string]interface{}{"react": map[string]interface{}{"pragma": "Foo"}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &CheckedRequiresOnchangeOrReadonlyRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: object key forms that must NOT match ----
		// Upstream reads `prop.key.name`; a string-literal key exposes `.value`,
		// not `.name`, so it is invisible — even when its text is "checked".
		{Code: `React.createElement('input', { 'checked': true })`, Tsx: true},
		{Code: `React.createElement('input', { ['checked']: true })`, Tsx: true},
		{Code: `React.createElement('input', { 'checked': true, 'onChange': noop })`, Tsx: true},
		{Code: `React.createElement('input', { 0: true })`, Tsx: true},

		// ---- Dimension 4: graceful degradation — no `checked` key present ----
		{Code: `React.createElement('input', {})`, Tsx: true},
		{Code: `React.createElement('input', { ...rest })`, Tsx: true},
		{Code: `React.createElement('input', { onChange: noop })`, Tsx: true},
		// JSX spread only / spread plus a non-checked target → no `checked` → ok.
		{Code: `<input {...props} />`, Tsx: true},
		{Code: `<input {...props} onChange={noop} />`, Tsx: true},

		// ---- Branch lock-in: elementType != "input" → listener bails ----
		{Code: `<select checked />`, Tsx: true},
		{Code: `<Input checked />`, Tsx: true},     // capitalized → user component, not "input"
		{Code: `<input.Foo checked />`, Tsx: true}, // member tag → "input.Foo" != "input"
		{Code: `<INPUT checked />`, Tsx: true},     // elementType is case-sensitive

		// ---- Branch lock-in: not a recognized createElement → CallExpression bails ----
		{Code: `document.createElement('input', { checked: true })`, Tsx: true},
		{Code: `notReact.createElement('input', { checked: true })`, Tsx: true},
		// TS-wrapped pragma is NOT recognized (matches upstream — `.object.name`
		// is undefined on a `TSAsExpression`). Differential-verified: 0 reports.
		{Code: `(React as any).createElement('input', { checked: true })`, Tsx: true},
		// A parenthesized optional chain freezes into a ChainExpression callee
		// upstream → NOT recognized. Differential-verified: 0 reports. (Contrast
		// the bare `React?.createElement(...)` below, which IS recognized.)
		{Code: `(React?.createElement)('input', { checked: true })`, Tsx: true},
		{Code: `((React?.createElement))('input', { checked: true })`, Tsx: true},
		// Bare (destructured) createElement is not recognized by the member-form
		// detector — documented Limitation, shared with sibling react rules.
		{Code: `createElement('input', { checked: true })`, Tsx: true},

		// ---- Branch lock-in: createElement first arg not the string "input" ----
		{Code: `React.createElement('select', { checked: true })`, Tsx: true},
		{Code: `React.createElement(Component, { checked: true })`, Tsx: true},
		{Code: "React.createElement(`input`, { checked: true })", Tsx: true}, // template literal != string Literal
		{Code: `React.createElement(42, { checked: true })`, Tsx: true},

		// ---- Branch lock-in: createElement second arg not an ObjectExpression ----
		{Code: `React.createElement('input')`, Tsx: true},
		{Code: `React.createElement('input', null)`, Tsx: true},
		{Code: `React.createElement('input', 'checked')`, Tsx: true},
		{Code: `React.createElement('input', [checked])`, Tsx: true},

		// ---- Dimension 4: namespaced JSX attribute name is not an Identifier → never matched ----
		{Code: `<input ns:checked />`, Tsx: true},

		// ---- Branch lock-in: no `checked` → checkAttributesAndReport returns early ----
		{Code: `<input readOnly />`, Tsx: true},
		{Code: `<input defaultChecked onChange={noop} />`, Tsx: true},
		{Code: `React.createElement('input', { defaultChecked: true })`, Tsx: true},
		{Code: `React.createElement('input', { onChange: noop, readOnly: true })`, Tsx: true},

		// ---- Branch lock-in: onChange OR readOnly — either arm satisfies ----
		{Code: `<input checked onChange={noop} readOnly />`, Tsx: true},

		// ---- Branch lock-in: configured pragma — default React no longer matches ----
		{
			Code:     `React.createElement('input', { checked: true })`,
			Settings: pragmaFoo,
			Tsx:      true,
		},

		// ---- Real-user: correctly controlled inputs (the happy path the rule guards) ----
		{Code: `<input type="checkbox" checked={isChecked} onChange={handleChange} />`, Tsx: true},
		{Code: `<input type="radio" checked={value === 'a'} onChange={onChange} />`, Tsx: true},
		// Real-user #3711: ignoreMissingProperties allows checked-only on the createElement path too.
		{
			Code:    `React.createElement('input', { checked: true })`,
			Options: ignoreMissing,
			Tsx:     true,
		},

		// ---- Options shape: array-wrapped config exercises GetOptionsMap's array branch ----
		{
			Code:    `<input type="checkbox" checked />`,
			Options: []interface{}{map[string]interface{}{"ignoreMissingProperties": true}},
			Tsx:     true,
		},

		// ---- Dimension 4: namespaced JSX *tag* → elementType "svg:input" != "input" ----
		{Code: `<svg:input checked />`, Tsx: true},

		// ---- Branch lock-in: createElement second arg is a variable (not an ObjectExpression) ----
		{Code: `React.createElement('input', props)`, Tsx: true},

		// ---- Dimension 4: computed key whose inner expr is NOT an Identifier → no `.name` ----
		{Code: `React.createElement('input', { [obj.checked]: true })`, Tsx: true},

		// ---- Robustness: duplicate attributes de-dupe like upstream's Set ----
		{Code: `<input checked onChange={a} onChange={b} />`, Tsx: true},

		// ---- Branch lock-in: defaultChecked without checked never reports (early return) ----
		{Code: `<input defaultChecked readOnly />`, Tsx: true},

		// ---- Real-user: a fully-wired controlled radio ----
		{Code: `<input type="radio" name="group" value="a" checked={sel === 'a'} onChange={onSel} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: createElement callee / argument paren wrappers (ESTree flattens parens) ----
		{
			Code: `(React).createElement('input', { checked: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `(React.createElement)('input', { checked: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		// Optional-chain pragma access — recognized, matching upstream (espree
		// callee is a MemberExpression with the optional flag). Differential-
		// verified against eslint-plugin-react: this reports.
		{
			Code: `React?.createElement('input', { checked: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		// Optional *call* `React.createElement?.(...)` — bare (no wrapping paren),
		// so the chain isn't frozen; recognized. Differential-verified: reports.
		{
			Code: `React.createElement?.('input', { checked: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement(('input'), { checked: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement('input', ({ checked: true }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Dimension 4: object key forms that DO match upstream `prop.key.name` ----
		// Computed Identifier key: upstream reads `.name` without gating on
		// `computed`, so `[checked]` matches exactly like `checked`.
		{
			Code: `React.createElement('input', { [checked]: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		// Shorthand property.
		{
			Code: `React.createElement('input', { checked })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Dimension 4: graceful degradation — spread present, `checked` still found ----
		{
			Code: `<input checked {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement('input', { checked: true, ...rest })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Dimension 2: nesting — listener fires on the inner input, not the container ----
		{
			Code: `<div><input type="checkbox" checked /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 6},
			},
		},
		{
			Code: `React.createElement('div', null, React.createElement('input', { checked: true }))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 34},
			},
		},
		// Two siblings: only the one missing onChange/readOnly is reported (no bleed).
		{
			Code: `<div><input checked /><input checked readOnly /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 6},
			},
		},

		// ---- Dimension 4: paired (non-self-closing) input drives the JsxOpeningElement
		//      listener; the report targets the opening tag, not the whole element ----
		{
			Code: `<input checked></input>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
			},
		},

		// ---- Dimension 4: full position assertions (self-closing JSX + createElement) ----
		{
			Code: `<input type="radio" checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingProperty",
					Message:   "`checked` should be used with either `onChange` or `readOnly`.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 31,
				},
			},
		},
		{
			Code: `React.createElement('input', { checked: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
			},
		},

		// ---- Dimension 4: multi-line cases (start/end span across lines) ----
		{
			Code: "<input\n  type=\"checkbox\"\n  checked\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1, EndLine: 4, EndColumn: 3},
			},
		},
		{
			Code: "React.createElement('input', {\n  checked: true\n})",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1, EndLine: 3, EndColumn: 3},
			},
		},

		// ---- Locks in: the attribute value is ignored — only presence matters (even checked={false}) ----
		{
			Code: `<input checked={false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Locks in: the `type` attribute is NOT gated — any input with `checked` is checked ----
		{
			Code: `<input checked />`, // no type at all
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="text" checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="hidden" checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Locks in: configured `settings.react.pragma` recognizes Foo.createElement ----
		{
			Code:     `Foo.createElement('input', { checked: true })`,
			Settings: pragmaFoo,
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Real-user #3711: the two options isolate to their own diagnostic (non-inverted) ----
		// ignoreMissingProperties suppresses ONLY missingProperty; exclusive remains.
		{
			Code:    `React.createElement('input', { checked: true, defaultChecked: true })`,
			Options: ignoreMissing,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "exclusiveCheckedAttribute",
					Message:   "Use either `checked` or `defaultChecked`, but not both.",
					Line:      1, Column: 1,
				},
			},
		},
		// ignoreExclusiveCheckedAttribute suppresses ONLY exclusive; missing remains.
		{
			Code:    `React.createElement('input', { checked: true, defaultChecked: true })`,
			Options: ignoreExclusive,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Real-user: controlled checkbox missing its onChange handler (the canonical bug) ----
		{
			Code: `<input type="checkbox" checked={state.checked} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Options shape: array-wrapped config exercises GetOptionsMap's array branch ----
		{
			Code:    `<input type="checkbox" checked defaultChecked />`,
			Options: []interface{}{map[string]interface{}{"ignoreExclusiveCheckedAttribute": true}},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Dimension 4: multi-level parenthesized first arg (ESTree flattens all levels) ----
		{
			Code: `React.createElement((('input')), { checked: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Dimension 4: computed key wrapping a parenthesized Identifier → still `.name` "checked" ----
		{
			Code: `React.createElement('input', { [(checked)]: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Robustness: duplicate object key collapses to one `checked` (upstream Set) ----
		{
			Code: `React.createElement('input', { checked: true, checked: false })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Dimension 4: JSX spread BEFORE checked — order-independent, spread is opaque ----
		{
			Code: `<input {...props} checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Dimension 4: TS-only value wrappers don't affect name-presence detection ----
		{
			Code: `<input checked={x as boolean} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `<input checked={x!} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},

		// ---- Real-user: `disabled` is not a substitute for onChange/readOnly ----
		{
			Code: `<input type="checkbox" checked disabled />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
	})
}
