package no_interactive_element_to_noninteractive_role

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoInteractiveElementToNoninteractiveRolePositions locks the
// diagnostic anchor: the rule reports on the JsxAttribute itself (NOT the
// opening element). Upstream's `context.report({ node: attribute, … })`
// makes the report range cover `role="…"`, so a multi-line opening tag
// still anchors the report on the role attribute's line / column.
func TestNoInteractiveElementToNoninteractiveRolePositions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInteractiveElementToNoninteractiveRoleRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Self-closing — report on `role="img"` between `href="…"` and `/>`.
		{
			Code: `<a href="http://x.y.z" role="img" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noInteractiveElementToNoninteractiveRole",
				Message:   errorMessage,
				Line:      1, Column: 24, EndLine: 1, EndColumn: 34,
			}},
		},
		// Paired form — report on `role="img"` only (not whole element).
		{
			Code: `<a href="http://x.y.z" role="img">x</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noInteractiveElementToNoninteractiveRole",
				Message:   errorMessage,
				Line:      1, Column: 24, EndLine: 1, EndColumn: 34,
			}},
		},
		// Multi-line opening tag — report still anchors to the role
		// attribute on its own line.
		{
			Code: "<a\n  href=\"http://x.y.z\"\n  role=\"img\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noInteractiveElementToNoninteractiveRole",
				Message:   errorMessage,
				Line:      3, Column: 3, EndLine: 3, EndColumn: 13,
			}},
		},
		// Tab-indented role — column counts each leading tab as one,
		// matching tsgo's char-offset semantics.
		{
			Code: "<a\n\thref=\"x\"\n\trole=\"img\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noInteractiveElementToNoninteractiveRole",
				Message:   errorMessage,
				Line:      3, Column: 2, EndLine: 3, EndColumn: 12,
			}},
		},
	})
}

// TestNoInteractiveElementToNoninteractiveRoleListenerBoundary locks
// that the listener fires independently for each JsxAttribute named
// `role` — nested JSX hierarchies produce one report per qualifying
// ancestor, with each report anchored on its own role attribute.
func TestNoInteractiveElementToNoninteractiveRoleListenerBoundary(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInteractiveElementToNoninteractiveRoleRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Two nested interactive-with-noninteractive-role elements — both
		// should report independently.
		{
			Code: "<a href=\"x\" role=\"img\">\n  <button role=\"img\" />\n</a>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage, Line: 1, Column: 13},
				{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage, Line: 2, Column: 11},
			},
		},
		// Three-level nesting — outer + middle + inner each report
		// against their own role attribute, never bleeding across.
		{
			Code: "<a href=\"x\" role=\"img\">\n  <button role=\"listitem\">\n    <select role=\"article\" />\n  </button>\n</a>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage, Line: 1, Column: 13},
				{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage, Line: 2, Column: 11},
				{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage, Line: 3, Column: 13},
			},
		},
	})
}

// TestNoInteractiveElementToNoninteractiveRoleEdgeShapes mirrors the
// Universal Edge Shapes checklist (Dimension 4) — spread / namespaced /
// boolean / non-literal / template shapes around the `role` attribute
// exercise corners upstream tests rarely cover.
func TestNoInteractiveElementToNoninteractiveRoleEdgeShapes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInteractiveElementToNoninteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Boolean form `<a href="x" role />` — `getLiteralPropValue`
			// returns boolean true, not a string. Both
			// `isNonInteractiveRole` and `isPresentationRole` return
			// false, so no report.
			{Code: `<a href="x" role />`, Tsx: true},
			// Non-literal role expressions — all noop in LITERAL_TYPES.
			{Code: `<a href="x" role={someRole} />`, Tsx: true},
			{Code: `<a href="x" role={cond ? "img" : "button"} />`, Tsx: true},
			{Code: `<a href="x" role={getRole()} />`, Tsx: true},
			{Code: `<a href="x" role={obj.role} />`, Tsx: true},
			{Code: `<a href="x" role={"img" + ""} />`, Tsx: true},
			{Code: `<a href="x" role={"img" || "button"} />`, Tsx: true},
			// `role={null}` — LITERAL_TYPES.Literal maps null to the
			// string "null", not a role name → no report.
			{Code: `<a href="x" role={null} />`, Tsx: true},
			// Capital `ROLE` is a different attribute name — upstream's
			// `propName(attr) !== 'role'` is case-sensitive.
			{Code: `<a href="x" ROLE="img" />`, Tsx: true},
			// Self-closing on a custom JSX component — not in dom map.
			{Code: `<MyButton role="img" />`, Tsx: true},
			// Namespaced role attribute — `propName` serializes to
			// `"ns:role"`, which isn't `"role"`.
			{Code: `<a href="x" ns:role="img" />`, Tsx: true},
			// Literal-spread that does NOT contain `role` — interactive
			// role on the JsxAttribute wins.
			{Code: `<a href="x" {...rest} role="button" />`, Tsx: true},
			// Template literal without substitutions resolves to "button".
			{Code: "<a href=\"x\" role={`button`} />", Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Template literal extracts to "img" — locks in the
			// template-literal arm.
			{
				Code:   "<a href=\"x\" role={`img`} />",
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Duplicate role attributes — both report. tsgo parses both,
			// and each listener invocation classifies via the FIRST
			// role attribute.
			{
				Code: `<a href="x" role="img" role="button" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage},
					{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage},
				},
			},
			// Multi-role `role="img button"` — `IsNonInteractiveRole`
			// takes the first VALID role (img) → non-interactive.
			{
				Code:   `<a href="x" role="img button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoInteractiveElementToNoninteractiveRoleOptionParsing locks the
// JSON-shape paths the CLI / JS-config feed:
//
//   - Single-option array: `["error", {tr:[...]}]` is unwrapped by
//     config.go to a bare map.
//   - Array-wrapped: matches rule_tester's multi-element shape.
//
// Without this suite, a regression where `parseOptions` only handles
// `[]interface{}` would silently fall back to defaults on every CLI
// invocation.
func TestNoInteractiveElementToNoninteractiveRoleOptionParsing(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInteractiveElementToNoninteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Bare map — CLI single-option shape.
			{
				Code:    `<tr role="presentation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tr": []interface{}{"none", "presentation"}},
			},
			// Array-wrapped — rule_tester multi-option shape.
			{
				Code:    `<canvas role="img" />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"canvas": []interface{}{"img"}}},
			},
			// Per-element override exempts only the configured roles.
			{
				Code:    `<tr role="none" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tr": []interface{}{"none", "presentation"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Override doesn't exempt other roles on the same tag.
			{
				Code:    `<tr role="img" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tr": []interface{}{"none", "presentation"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Override on `tr` doesn't help `<a href>`.
			{
				Code:    `<a href="x" role="img" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tr": []interface{}{"none", "presentation"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Empty options map — strict semantics.
			{
				Code:    `<canvas role="img" />`,
				Tsx:     true,
				Options: map[string]interface{}{},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Nil options.
			{
				Code:   `<canvas role="img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoInteractiveElementToNoninteractiveRoleRoleExpressionShapes locks
// the rule's behavior across every tsgo-specific wrapper around a
// `role={…}` expression value. tsgo preserves several AST shapes that
// ESTree flattens at parse time, and `LiteralPropStringValue` has to
// route them correctly into upstream's `getLiteralPropValue` semantics
// — a regression that strips the wrong wrapper would silently shift the
// rule's surface (over- or under-reporting). Each row pins one shape.
func TestNoInteractiveElementToNoninteractiveRoleRoleExpressionShapes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInteractiveElementToNoninteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// `as` / `satisfies` / non-null wrappers map to noop in
			// jsx-ast-utils' LITERAL_TYPES → null → no classification →
			// no report. tsgo preserves these wrappers as discrete AST
			// nodes (vs ESTree flattening some); the literal extractor
			// must NOT peel them.
			{Code: `<a href="x" role={"img" as const} />`, Tsx: true},
			{Code: `<a href="x" role={"img" as Role} />`, Tsx: true},
			{Code: `<a href="x" role={"img" satisfies string} />`, Tsx: true},
			// Identifier from a local binding — non-literal in
			// LITERAL_TYPES sense. tsgo's symbol resolution doesn't
			// participate in literal extraction.
			{
				Code: "const r = 'img';\nconst el = <a href=\"x\" role={r} />;",
				Tsx:  true,
			},
		},
		[]rule_tester.InvalidTestCase{
			// `role={"img"}` — JsxExpression wrapping a StringLiteral.
			// tsgo represents this as JsxExpression > StringLiteral
			// (no implicit unwrap); the extractor must recurse through.
			{
				Code:   `<a href="x" role={"img"} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// `role={("img")}` — single paren wrap. tsgo preserves
			// ParenthesizedExpression; ESTree flattens. The extractor
			// must strip parens for parity.
			{
				Code:   `<a href="x" role={("img")} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Multi-level parens.
			{
				Code:   `<a href="x" role={(("img"))} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// HTML entity in the raw string — `&#105;` decodes to `i`,
			// producing "img". tsgo's StringLiteral.Text preserves the
			// raw `&…;` source; the literal extractor decodes via
			// `jsxtransforms.DecodeEntities`.
			{
				Code:   `<a href="x" role="&#105;mg" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoInteractiveElementToNoninteractiveRoleRoleStringBoundaries locks
// `IsNonInteractiveRole`'s string-handling edges. Upstream does
// `String(value).toLowerCase().split(' ')` then takes the FIRST valid
// role. The combination of single-space split (not `\s+`) and "first
// valid role wins" produces several non-obvious classifications that a
// Go re-implementation could easily flip:
//
//   - case-folding: `"IMG"` and `"Img"` both classify as `img`
//   - empty string → no token → no classification
//   - leading/trailing space → split yields `""` token; skipped
//   - multi-space → empty internal tokens; skipped
//   - invalid-then-valid → invalid token skipped, valid wins
func TestNoInteractiveElementToNoninteractiveRoleRoleStringBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInteractiveElementToNoninteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Empty role string — no valid role token → no classification.
			{Code: `<a href="x" role="" />`, Tsx: true},
			// All-whitespace role — same.
			{Code: `<a href="x" role="   " />`, Tsx: true},
			// Unknown role name — not in role set, no classification.
			{Code: `<a href="x" role="foobar" />`, Tsx: true},
			// All-invalid space-separated — no token resolves.
			{Code: `<a href="x" role="xxx yyy zzz" />`, Tsx: true},
			// First valid is interactive (`button`) → no report.
			{Code: `<a href="x" role="xxx button img" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Case-folded — upstream `.toLowerCase()` normalizes.
			{
				Code:   `<a href="x" role="IMG" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			{
				Code:   `<a href="x" role="Img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Leading space — split produces `["", "img"]`; second
			// token wins. Locks in the "single-space split" semantics.
			{
				Code:   `<a href="x" role=" img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Trailing space — split produces `["img", ""]`; first wins.
			{
				Code:   `<a href="x" role="img " />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Multiple internal spaces — `["img", "", "button"]`; first
			// valid token "img" wins.
			{
				Code:   `<a href="x" role="img  button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// First invalid, second valid non-interactive.
			{
				Code:   `<a href="x" role="xxx img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoInteractiveElementToNoninteractiveRoleOptionRobustness locks
// `parseOptions`' behavior on misshapen / borderline option payloads
// that the JSON schema doesn't fully constrain at the rule layer. None
// of these should panic or report inconsistently:
//
//   - empty array value (`{tr: []}`) — no role exempted, strict semantics
//   - non-array value (`{tr: "presentation"}`) — silently ignored
//   - mixed-type array (`{tr: ["none", 123, null]}`) — non-strings dropped
//   - option key case mismatch (`{TR: [...]}` vs tag `<tr>`) — does NOT match
//   - option role-value case mismatch (`{tr: ["Presentation"]}` vs `role="presentation"`)
func TestNoInteractiveElementToNoninteractiveRoleOptionRobustness(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInteractiveElementToNoninteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Mixed-type array — only "presentation" survives, exempting
			// the case.
			{
				Code:    `<tr role="presentation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tr": []interface{}{"none", 123, nil, "presentation"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Empty array — no role exempted under that key.
			{
				Code:    `<tr role="presentation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tr": []interface{}{}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Non-array value — `StringSliceOption` returns nil, key is
			// silently dropped from `allowedRoles`.
			{
				Code:    `<tr role="presentation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tr": "presentation"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Mixed-type array containing only the WRONG roles — still
			// reports.
			{
				Code:    `<tr role="presentation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tr": []interface{}{"none", 123, nil}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Option key case mismatch — `getElementType` returns "tr"
			// (lowercased tag); option key "TR" doesn't match.
			{
				Code:    `<tr role="presentation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"TR": []interface{}{"presentation"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Option role-value case mismatch — `slices.Contains` is
			// case-sensitive; "Presentation" doesn't match "presentation".
			{
				Code:    `<tr role="presentation" />`,
				Tsx:     true,
				Options: map[string]interface{}{"tr": []interface{}{"Presentation"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoInteractiveElementToNoninteractiveRoleContainerForms locks
// listener triggering across real-world React patterns — JSX inside
// fragments / ternaries / array maps / arrow bodies / conditional
// renders. The rule must fire identically regardless of the surrounding
// expression context. A regression that hooked into `JsxOpeningElement`
// rather than `JsxAttribute` would silently drop these.
func TestNoInteractiveElementToNoninteractiveRoleContainerForms(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInteractiveElementToNoninteractiveRoleRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Inside a JsxFragment (`<>…</>`).
			{
				Code:   `<><a href="x" role="img" /></>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Inside a ternary returned from a JsxExpression child.
			{
				Code:   `<div>{cond ? <a href="x" role="img" /> : null}</div>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Inside an arrow function body, common in `.map(…)`.
			{
				Code:   `const r = items.map(x => <a href={x} role="img" />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Returned from a function declaration body.
			{
				Code:   `function R() { return <button role="article">x</button>; }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Paren-wrapped JSX expression — `({<a … />})` form.
			{
				Code:   `const x = (<a href="x" role="img" />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoInteractiveElementToNoninteractiveRoleInteractiveGate locks
// `IsInteractiveElement`'s ATTRIBUTE-DEPENDENT classification. `<a>` is
// interactive only when it has an `href` attribute — both
// `IsInteractiveElement` and our rule must respect that. A regression
// where the schema's required-attribute predicate is lost would either
// silently exempt `<a href="x" role="img" />` or wrongly flag
// `<a role="img" />`.
//
// The all-flavors-of-input cases live in upstream's neverValid; this
// suite locks `<a>`'s href-dependency separately because it's the most
// common case the schema gate decides on.
func TestNoInteractiveElementToNoninteractiveRoleInteractiveGate(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInteractiveElementToNoninteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// `<a>` without `href` — schema gate fails →
			// `IsInteractiveElement` returns false → no report.
			{Code: `<a role="img" />`, Tsx: true},
			{Code: `<a role="listitem" />`, Tsx: true},
			// `<a tabIndex>` without `href` — tabIndex alone doesn't
			// make `<a>` interactive in aria-query's schema.
			{Code: `<a tabIndex="0" role="img" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// `<a href="x">` — href present → interactive → report.
			{
				Code:   `<a href="x" role="img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// `<a href={x}>` — href present (value irrelevant to schema).
			{
				Code:   `<a href={someUrl} role="img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// `<a href />` — boolean form; `hasProp` matches regardless
			// of value, so `<a>` is interactive.
			{
				Code:   `<a href role="img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// `<a href="">` — empty string is still a present `href`.
			{
				Code:   `<a href="" role="img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoInteractiveElementToNoninteractiveRoleMixedRealWorld captures
// real-world JSX shapes — multiple a11y props, event handlers, spread
// + role ordering — that a rule operating only on simple attribute
// lists could mishandle. Each case is a "looks like real code" pattern
// users would actually write.
func TestNoInteractiveElementToNoninteractiveRoleMixedRealWorld(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInteractiveElementToNoninteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Interactive role + event handlers + className mixed — no
			// non-interactive role anywhere; not reported.
			{Code: `<a href="x" role="button" className="cta" onClick={f} aria-label="Go" />`, Tsx: true},
			// Spread before AND after a `role="button"` — interactive,
			// not reported.
			{Code: `<a href="x" {...a} role="button" {...b} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Real-world: developer slapping role="img" on a button +
			// onClick handler. Single report on the `role` attribute,
			// the rest of the attributes are noise.
			{
				Code: `<button role="img" onClick={f} aria-label="x" className="icon" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage},
				},
			},
			// Spread before role="img" on `<a href>` — spread is opaque,
			// the JsxAttribute fires normally.
			{
				Code:   `<a href="x" {...rest} role="img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Spread AFTER role="img" — same.
			{
				Code:   `<a href="x" role="img" {...rest} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
			// Multiple non-interactive props mixed in — only the `role`
			// triggers; one report.
			{
				Code:   `<select role="article" aria-hidden={false} onClick={f} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noInteractiveElementToNoninteractiveRole", Message: errorMessage}},
			},
		})
}
