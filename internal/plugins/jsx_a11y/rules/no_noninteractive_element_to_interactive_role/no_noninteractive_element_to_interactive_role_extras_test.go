// cspell:ignore utton buttoń Chakra reimplementation

package no_noninteractive_element_to_interactive_role

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoNoninteractiveElementToInteractiveRolePositions locks the
// diagnostic anchor: the rule reports on the JsxAttribute itself (NOT the
// opening element). Upstream's `context.report({ node: attribute, … })`
// makes the report range cover `role="…"`, so a multi-line opening tag
// still anchors the report on the role attribute's line / column.
func TestNoNoninteractiveElementToInteractiveRolePositions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Self-closing — report on `role="button"` between the tag name and `/>`.
		{
			Code: `<article role="button" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNoninteractiveElementToInteractiveRole",
				Message:   errorMessage,
				Line:      1, Column: 10, EndLine: 1, EndColumn: 23,
			}},
		},
		// Paired form — report on `role="button"` only (not whole element).
		{
			Code: `<article role="button">x</article>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNoninteractiveElementToInteractiveRole",
				Message:   errorMessage,
				Line:      1, Column: 10, EndLine: 1, EndColumn: 23,
			}},
		},
		// Multi-line opening tag — report still anchors to the role
		// attribute on its own line.
		{
			Code: "<article\n  className=\"foo\"\n  role=\"button\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNoninteractiveElementToInteractiveRole",
				Message:   errorMessage,
				Line:      3, Column: 3, EndLine: 3, EndColumn: 16,
			}},
		},
		// Tab-indented role — column counts each leading tab as one,
		// matching tsgo's char-offset semantics.
		{
			Code: "<article\n\tclassName=\"foo\"\n\trole=\"button\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noNoninteractiveElementToInteractiveRole",
				Message:   errorMessage,
				Line:      3, Column: 2, EndLine: 3, EndColumn: 15,
			}},
		},
	})
}

// TestNoNoninteractiveElementToInteractiveRoleListenerBoundary locks that
// the listener fires independently for each JsxAttribute named `role` —
// nested JSX hierarchies produce one report per qualifying ancestor, with
// each report anchored on its own role attribute.
func TestNoNoninteractiveElementToInteractiveRoleListenerBoundary(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Two nested non-interactive-with-interactive-role elements — both
		// should report independently.
		{
			Code: "<article role=\"button\">\n  <h1 role=\"menuitem\" />\n</article>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage, Line: 1, Column: 10},
				{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage, Line: 2, Column: 7},
			},
		},
		// Three-level nesting — outer + middle + inner each report against
		// their own role attribute, never bleeding across.
		{
			Code: "<article role=\"button\">\n  <h2 role=\"menuitem\">\n    <li role=\"link\" />\n  </h2>\n</article>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage, Line: 1, Column: 10},
				{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage, Line: 2, Column: 7},
				{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage, Line: 3, Column: 9},
			},
		},
	})
}

// TestNoNoninteractiveElementToInteractiveRoleEdgeShapes mirrors the
// Universal Edge Shapes checklist (Dimension 4) — spread / namespaced /
// boolean / non-literal / template shapes around the `role` attribute
// exercise corners upstream tests rarely cover.
func TestNoNoninteractiveElementToInteractiveRoleEdgeShapes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Boolean form `<article role />` — `getLiteralPropValue`
			// returns boolean true, not a string. `IsInteractiveRole`
			// returns false, so no report.
			{Code: `<article role />`, Tsx: true},
			// Non-literal role expressions — all noop in LITERAL_TYPES.
			{Code: `<article role={someRole} />`, Tsx: true},
			{Code: `<article role={cond ? "button" : "img"} />`, Tsx: true},
			{Code: `<article role={getRole()} />`, Tsx: true},
			{Code: `<article role={obj.role} />`, Tsx: true},
			{Code: `<article role={"button" + ""} />`, Tsx: true},
			{Code: `<article role={"button" || "img"} />`, Tsx: true},
			// `role={null}` — LITERAL_TYPES.Literal maps null to the
			// string "null", not a role name → no report.
			{Code: `<article role={null} />`, Tsx: true},
			// Capital `ROLE` is a different attribute name — upstream's
			// `propName(attr) !== 'role'` is case-sensitive.
			{Code: `<article ROLE="button" />`, Tsx: true},
			// Self-closing on a custom JSX component — not in dom map.
			{Code: `<MyArticle role="button" />`, Tsx: true},
			// Namespaced role attribute — `propName` serializes to
			// `"ns:role"`, which isn't `"role"`.
			{Code: `<article ns:role="button" />`, Tsx: true},
			// Literal-spread that does NOT contain `role` — non-interactive
			// role on the JsxAttribute (none here) wins.
			{Code: `<article {...rest} role="listitem" />`, Tsx: true},
			// Template literal without substitutions resolves to "listitem"
			// (non-interactive) — no report.
			{Code: "<article role={`listitem`} />", Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Template literal extracts to "button" — locks in the
			// template-literal arm.
			{
				Code:   "<article role={`button`} />",
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Duplicate role attributes — both report. tsgo parses both,
			// and each listener invocation classifies via the FIRST
			// role attribute.
			{
				Code: `<article role="button" role="listitem" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// Multi-role `role="button img"` — `IsInteractiveRole` takes
			// the first VALID role (button) → interactive.
			{
				Code:   `<article role="button img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleOptionParsing locks the
// JSON-shape paths the CLI / JS-config feed:
//
//   - Single-option array: `["error", {ul:[...]}]` is unwrapped by
//     config.go to a bare map.
//   - Array-wrapped: matches rule_tester's multi-element shape.
//
// Without this suite, a regression where `parseOptions` only handles
// `[]interface{}` would silently fall back to defaults on every CLI
// invocation.
func TestNoNoninteractiveElementToInteractiveRoleOptionParsing(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Bare map — CLI single-option shape.
			{
				Code:    `<ul role="menu" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"menu", "menubar"}},
			},
			// Array-wrapped — rule_tester multi-option shape.
			{
				Code:    `<ul role="menubar" />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"ul": []interface{}{"menu", "menubar"}}},
			},
			// Per-element override exempts only the configured roles.
			{
				Code:    `<table role="grid" />`,
				Tsx:     true,
				Options: map[string]interface{}{"table": []interface{}{"grid"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Override doesn't exempt other roles on the same tag.
			{
				Code:    `<ul role="tab" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"menu"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Override on `ul` doesn't help `<article>`.
			{
				Code:    `<article role="button" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"menu"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Empty options map — strict semantics.
			{
				Code:    `<ul role="menu" />`,
				Tsx:     true,
				Options: map[string]interface{}{},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Nil options.
			{
				Code:   `<ul role="menu" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleRoleExpressionShapes locks
// the rule's behavior across every tsgo-specific wrapper around a
// `role={…}` expression value. Each row pins one shape.
func TestNoNoninteractiveElementToInteractiveRoleRoleExpressionShapes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// `as` / `satisfies` / non-null wrappers map to noop in
			// jsx-ast-utils' LITERAL_TYPES → null → no classification.
			{Code: `<article role={"button" as const} />`, Tsx: true},
			{Code: `<article role={"button" as Role} />`, Tsx: true},
			{Code: `<article role={"button" satisfies string} />`, Tsx: true},
			// Identifier from a local binding — non-literal in
			// LITERAL_TYPES sense.
			{
				Code: "const r = 'button';\nconst el = <article role={r} />;",
				Tsx:  true,
			},
		},
		[]rule_tester.InvalidTestCase{
			// `role={"button"}` — JsxExpression wrapping a StringLiteral.
			{
				Code:   `<article role={"button"} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// `role={("button")}` — single paren wrap. tsgo preserves
			// ParenthesizedExpression; ESTree flattens. The extractor
			// must strip parens for parity.
			{
				Code:   `<article role={("button")} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Multi-level parens.
			{
				Code:   `<article role={(("button"))} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// HTML entity in the raw string — `&#98;` decodes to `b`,
			// producing "button".
			{
				Code:   `<article role="&#98;utton" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleRoleStringBoundaries locks
// `IsInteractiveRole`'s string-handling edges. Upstream does
// `String(value).toLowerCase().split(' ')` then takes the FIRST valid
// role. The combination of single-space split (not `\s+`) and "first
// valid role wins" produces several non-obvious classifications that a
// Go re-implementation could easily flip.
func TestNoNoninteractiveElementToInteractiveRoleRoleStringBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Empty role string — no valid role token → no classification.
			{Code: `<article role="" />`, Tsx: true},
			// All-whitespace role — same.
			{Code: `<article role="   " />`, Tsx: true},
			// Unknown role name — not in role set, no classification.
			{Code: `<article role="foobar" />`, Tsx: true},
			// All-invalid space-separated — no token resolves.
			{Code: `<article role="xxx yyy zzz" />`, Tsx: true},
			// First valid is non-interactive (`listitem`) → no report.
			{Code: `<article role="xxx listitem button" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Case-folded — upstream `.toLowerCase()` normalizes.
			{
				Code:   `<article role="BUTTON" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			{
				Code:   `<article role="Button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Leading space — split produces `["", "button"]`; second
			// token wins. Locks in the "single-space split" semantics.
			{
				Code:   `<article role=" button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Trailing space — split produces `["button", ""]`; first wins.
			{
				Code:   `<article role="button " />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Multiple internal spaces — `["button", "", "img"]`; first
			// valid token "button" wins.
			{
				Code:   `<article role="button  img" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// First invalid, second valid interactive.
			{
				Code:   `<article role="xxx button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleOptionRobustness locks
// `parseOptions`' behavior on misshapen / borderline option payloads:
//
//   - empty array value (`{ul: []}`) — no role exempted, strict semantics
//   - non-array value (`{ul: "menu"}`) — silently ignored
//   - mixed-type array (`{ul: ["menu", 123, null]}`) — non-strings dropped
//   - option key case mismatch (`{UL: [...]}` vs tag `<ul>`) — does NOT match
//   - option role-value case mismatch (`{ul: ["Menu"]}` vs `role="menu"`)
//   - allow-list value case mismatch with upstream's `getExplicitRole`
//     lowercasing — `role="Menu"` (lowercases to "menu") matches an
//     allow-list entry "menu" exactly.
func TestNoNoninteractiveElementToInteractiveRoleOptionRobustness(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Mixed-type array — only "menu" survives, exempting the case.
			{
				Code:    `<ul role="menu" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"menu", 123, nil, "menubar"}},
			},
			// Case-mismatched ROLE-attribute value vs lowercase allow-list
			// — upstream lowercases the role via `getExplicitRole`, so
			// `role="Menu"` matches allow-list entry "menu". Locks in the
			// upstream `getExplicitRole` lowercasing arm — a regression
			// that compared the raw role value would flip this to invalid.
			{
				Code:    `<ul role="Menu" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"menu"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Empty array — no role exempted under that key.
			{
				Code:    `<ul role="menu" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Non-array value — `StringSliceOption` returns nil, key is
			// silently dropped from `allowedRoles`.
			{
				Code:    `<ul role="menu" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": "menu"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Mixed-type array containing only the WRONG roles — still
			// reports.
			{
				Code:    `<ul role="menu" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"menubar", 123, nil}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Option key case mismatch — `getElementType` returns "ul"
			// (lowercased tag); option key "UL" doesn't match.
			{
				Code:    `<ul role="menu" />`,
				Tsx:     true,
				Options: map[string]interface{}{"UL": []interface{}{"menu"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Allow-list entry case mismatch — `slices.Contains` is
			// case-sensitive; lowercased "menu" doesn't match "Menu".
			{
				Code:    `<ul role="menu" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"Menu"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleContainerForms locks
// listener triggering across real-world React patterns — JSX inside
// fragments / ternaries / array maps / arrow bodies / conditional
// renders. The rule must fire identically regardless of the surrounding
// expression context.
func TestNoNoninteractiveElementToInteractiveRoleContainerForms(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// Inside a JsxFragment (`<>…</>`).
			{
				Code:   `<><article role="button" /></>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Inside a ternary returned from a JsxExpression child.
			{
				Code:   `<div>{cond ? <article role="button" /> : null}</div>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Inside an arrow function body, common in `.map(…)`.
			{
				Code:   `const r = items.map(x => <h1 role="button">{x}</h1>);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Returned from a function declaration body.
			{
				Code:   `function R() { return <article role="button">x</article>; }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Paren-wrapped JSX expression — `({<article … />})` form.
			{
				Code:   `const x = (<article role="button" />);`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleNonInteractiveGate locks
// `IsNonInteractiveElement`'s ATTRIBUTE-DEPENDENT classification. `<img>`
// is non-interactive only when `alt` is present; `<form>` is
// non-interactive only when `aria-label`/`aria-labelledby`/`name` is
// present. A regression where the schema's required-attribute predicate
// is lost would either silently flag `<img role="button" />` without alt
// or silently exempt `<img alt="x" role="button" />` (or the form variants).
func TestNoNoninteractiveElementToInteractiveRoleNonInteractiveGate(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// `<form>` without aria-label / aria-labelledby / name — schema
			// gate fails → `IsNonInteractiveElement` returns false (the
			// nonInteractiveElementRoleSchemas entries all require one of
			// those attrs); `nonInteractiveElementAXSchemas` has a bare
			// `form` entry though, so this actually IS non-interactive
			// via the AX path. Lock that as INVALID below — kept here
			// only as a sentinel during regression triage.
			//
			// Empty placeholder (intentionally none here): the AX schema
			// matches `<form>` unconditionally, so there's no "form without
			// any of the three attrs" valid case for this rule.
		},
		[]rule_tester.InvalidTestCase{
			// `<img alt="x">` — alt present → schema gate matches in
			// nonInteractiveElementRoleSchemas → non-interactive → report.
			{
				Code:   `<img alt="x" role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// `<img alt="">` — empty-alt explicit schema entry matches.
			{
				Code:   `<img alt="" role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// `<img>` — no alt, but matches `nonInteractiveElementAXSchemas`
			// bare `{Name: "img"}` entry → still non-interactive → report.
			{
				Code:   `<img role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// `<form aria-label="…">` — aria-label present → schema entry
			// matches → non-interactive → report.
			{
				Code:   `<form aria-label="Login" role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleMixedRealWorld captures
// real-world JSX shapes — multiple a11y props, event handlers, spread
// + role ordering — that a rule operating only on simple attribute
// lists could mishandle.
func TestNoNoninteractiveElementToInteractiveRoleMixedRealWorld(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Non-interactive role on a non-interactive element — fine.
			{Code: `<article role="article" className="card" onClick={f} aria-label="Card" />`, Tsx: true},
			// Spread before AND after a `role="article"` — not interactive,
			// not reported.
			{Code: `<article {...a} role="article" {...b} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Real-world: developer slapping role="button" on a heading +
			// onClick handler. Single report on the `role` attribute, the
			// rest of the attributes are noise.
			{
				Code: `<h1 role="button" onClick={f} aria-label="x" className="hdr" />`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// Spread before role="button" on `<article>` — spread is
			// opaque, the JsxAttribute fires normally.
			{
				Code:   `<article {...rest} role="button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Spread AFTER role="button" — same.
			{
				Code:   `<article role="button" {...rest} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Multiple interactive props mixed in — only the `role`
			// triggers; one report.
			{
				Code:   `<table role="menuitem" aria-hidden={false} onClick={f} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleAllowListSemantics locks
// the rule's allow-list lookup against the sibling
// `no-interactive-element-to-noninteractive-role` rule. Upstream uses
// DIFFERENT extraction helpers:
//
//   - this rule:    `includes(allowedRoles[type], getExplicitRole(...))`
//     — lowercased + rolesMap-validated literal value
//   - sibling rule: `includes(allowedRoles[type], getLiteralPropValue(...))`
//     — raw literal value, no lowercasing or rolesMap filter
//
// The observable consequences this suite locks in:
//
//  1. A `role` value that isn't a real ARIA role ("foobar") never
//     matches the allow-list here — `getExplicitRole` returns null and
//     the check collapses. Sibling rule would match if the allow-list
//     contained "foobar".
//  2. A multi-role `role="menu menubar"` likewise misses — `rolesMap.has`
//     checks the whole-string key, which fails on space-separated values.
//  3. `role="MENU"` (uppercase) matches an allow-list entry "menu" —
//     `getExplicitRole` lowercases before comparing. Sibling rule would
//     NOT match (raw-value comparison is case-sensitive).
func TestNoNoninteractiveElementToInteractiveRoleAllowListSemantics(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Uppercase role value — `getExplicitRole` lowercases →
			// matches "menu" in allow-list. Locks in the upstream
			// `getExplicitRole` lowercase arm.
			{
				Code:    `<ul role="MENU" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"menu"}},
			},
			// Mixed-case role value with allow-list entry in correct case
			// — `getExplicitRole` lowercases to "menubar" which matches.
			{
				Code:    `<ul role="MenuBar" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"menubar"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid role name in allow-list does NOT exempt — upstream
			// `getExplicitRole` returns null for "foobar" (rolesMap.has
			// is false), so the comparison fails. Locks in upstream's
			// rolesMap-validation arm. (Tag also reports because <ul>
			// + interactive role = report.)
			//
			// Note: `<ul role="foobar">` would NOT report on its own
			// because `IsInteractiveRole` also rejects "foobar". Use a
			// real interactive role for the actual report and only the
			// allow-list bypass for "foobar" in the option.
			{
				Code:    `<ul role="menu" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"foobar"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Multi-role attribute — `getExplicitRole` returns null
			// (whole-string rolesMap lookup fails on "menu menubar"),
			// allow-list comparison collapses, falls through to
			// `IsInteractiveRole` which takes first valid role "menu"
			// → interactive → report. Locks in the multi-role asymmetry
			// between getExplicitRole (whole-string) and
			// IsInteractiveRole (space-split).
			{
				Code:    `<ul role="menu menubar" />`,
				Tsx:     true,
				Options: map[string]interface{}{"ul": []interface{}{"menu", "menubar"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRolePolymorphicSettings locks
// the `polymorphicPropName` / `polymorphicAllowList` settings interaction
// — the canonical MUI `<Box as="article">` and Chakra
// `<Box as="article">` pattern. Without polymorphic settings the rule
// can't see through the custom-component name; with them, the resolved
// element type swaps to the polymorphic prop's literal string and the
// rule reports against that. A regression where the polymorphic
// extraction silently dropped (or always-on) would either under- or
// over-report on this very common real-world shape.
func TestNoNoninteractiveElementToInteractiveRolePolymorphicSettings(t *testing.T) {
	polySettings := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"polymorphicPropName": "as",
		},
	}
	polySettingsWithAllowList := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"polymorphicPropName":  "as",
			"polymorphicAllowList": []interface{}{"Box"},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Without settings — `<Box>` is a custom component, no dom.has match.
			{Code: `<Box as="article" role="button" />`, Tsx: true},
			// With settings, `as="div"` — div is generic, not non-interactive.
			{Code: `<Box as="div" role="button" />`, Tsx: true, Settings: polySettings},
			// With settings, no `role` attribute — listener never fires for `role`.
			{Code: `<Box as="article" />`, Tsx: true, Settings: polySettings},
			// With settings + allowList, `<Other>` not in allowList → stays "Other"
			// (custom component) → bails before role check.
			{Code: `<Other as="article" role="button" />`, Tsx: true, Settings: polySettingsWithAllowList},
			// With settings, polymorphic prop is non-literal — extractor returns
			// ("",false), rawType stays "Box" (custom component) → bail.
			{Code: `<Box as={someTag} role="button" />`, Tsx: true, Settings: polySettings},
			// With settings, `as` prop missing — rawType stays "Box" → bail.
			{Code: `<Box role="button" />`, Tsx: true, Settings: polySettings},
			// With settings, polymorphic resolves to a non-interactive element
			// + non-interactive role → no report.
			{Code: `<Box as="article" role="article" />`, Tsx: true, Settings: polySettings},
		},
		[]rule_tester.InvalidTestCase{
			// Canonical case: <Box as="article" role="button" /> → resolves to
			// "article" → non-interactive + interactive role = report.
			{
				Code:     `<Box as="article" role="button" />`,
				Tsx:      true,
				Settings: polySettings,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// `<Box as="h1" role="menuitem" />` resolves to h1 (non-interactive).
			{
				Code:     `<Box as="h1" role="menuitem" />`,
				Tsx:      true,
				Settings: polySettings,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// AllowList includes "Box" → polymorphic resolution applies → reports.
			{
				Code:     `<Box as="article" role="button" />`,
				Tsx:      true,
				Settings: polySettingsWithAllowList,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Polymorphic resolves to <table> with role="grid" — strict semantics
			// (without recommended allow-list for table→grid) → reports.
			{
				Code:     `<Box as="table" role="grid" />`,
				Tsx:      true,
				Settings: polySettings,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleComponentsMappingEdges
// locks corner cases of the `components` settings map — the
// non-polymorphic `<Article>` → `"article"` rewrite that lets users
// declare aliases for custom JSX components. Upstream's
// `getElementType` looks up by EXACT rawType key; the post-rewrite
// rawType then re-enters the rule pipeline.
func TestNoNoninteractiveElementToInteractiveRoleComponentsMappingEdges(t *testing.T) {
	caseSensitiveMap := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"components": map[string]interface{}{
				"Article": "article",
			},
		},
	}
	mapToNonDOM := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"components": map[string]interface{}{
				"Article": "MysteryThing",
			},
		},
	}
	mapToInteractive := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"components": map[string]interface{}{
				"MyButton": "button",
			},
		},
	}
	mapToGeneric := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"components": map[string]interface{}{
				"Wrapper": "div",
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Map key is case-sensitive — `<article>` (lower) doesn't match
			// "Article" key in components map; goes through with rawType "article"
			// (which IS dom — so this just hits role check directly).
			// `role="button"` on raw `<article>` would report, so use a
			// non-interactive role to verify the lowercase path isn't
			// remapped.
			{Code: `<article role="article" />`, Tsx: true, Settings: caseSensitiveMap},
			// Map case mismatch on JSX side — `<ARTICLE>` doesn't match
			// "Article" key → stays "ARTICLE" → not in dom set → bail.
			{Code: `<ARTICLE role="button" />`, Tsx: true, Settings: caseSensitiveMap},
			// Map → non-DOM name → bail at IsDOMElement.
			{Code: `<Article role="button" />`, Tsx: true, Settings: mapToNonDOM},
			// Map → interactive element + interactive role → no report.
			{Code: `<MyButton role="button" />`, Tsx: true, Settings: mapToInteractive},
			// Map → generic element (div) + interactive role → no report
			// (div isn't non-interactive).
			{Code: `<Wrapper role="button" />`, Tsx: true, Settings: mapToGeneric},
			// Custom component WITHOUT settings — bail at dom.has.
			{Code: `<Article role="button" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Map → non-interactive element + interactive role → report.
			{
				Code:     `<Article role="button" />`,
				Tsx:      true,
				Settings: caseSensitiveMap,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleHOCAndComponentBodies locks
// the rule across the common React component declaration / wrapper
// patterns. The listener fires on JsxAttribute regardless of the
// surrounding callable's shape; a regression that hooked into top-level
// statements (or skipped JSX inside callable bodies) would silently
// miss every JSX inside a real React codebase.
func TestNoNoninteractiveElementToInteractiveRoleHOCAndComponentBodies(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// forwardRef arrow body.
			{
				Code: "const C = React.forwardRef((props, ref) => <article role=\"button\" ref={ref} />);",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// React.memo wrapping an arrow.
			{
				Code: "const C = React.memo(({children}) => <article role=\"button\">{children}</article>);",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// Class component render method.
			{
				Code: "class C extends React.Component { render() { return <article role=\"button\" />; } }",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// Functional component arrow assigned to a typed const.
			{
				Code: "const C: React.FC = () => <h1 role=\"menuitem\">x</h1>;",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// Hook callback returning JSX.
			{
				Code: "function useThing() { return <ul role=\"menu\" />; }",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// JSX returned from `useMemo` callback.
			{
				Code: "const m = React.useMemo(() => <article role=\"button\" />, []);",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// JSX passed as a render prop.
			{
				Code: "const x = <Parent renderItem={() => <li role=\"tab\" />} />;",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleDeepNesting verifies the
// listener fires INDEPENDENTLY for each role attribute at every depth
// — no accumulated state across the listener invocations would silently
// mask sibling reports. The fixture mirrors a typical Card UI: outer
// container with a header, body, and footer, each containing its own
// nested JSX.
func TestNoNoninteractiveElementToInteractiveRoleDeepNesting(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// 4 violations across 5 nested levels (Card → CardHeader → h1;
			// Card → CardBody → article → h2; Card → CardFooter → ul → li).
			// Custom components (Card, CardHeader, CardBody, CardFooter) are
			// non-DOM and bail; the inner DOM elements fire.
			{
				Code: "<Card>\n" +
					"  <CardHeader>\n" +
					"    <h1 role=\"button\">Title</h1>\n" +
					"  </CardHeader>\n" +
					"  <CardBody>\n" +
					"    <article role=\"button\">\n" +
					"      <h2 role=\"menuitem\">Subtitle</h2>\n" +
					"    </article>\n" +
					"  </CardBody>\n" +
					"  <CardFooter>\n" +
					"    <ul role=\"button\">\n" +
					"      <li role=\"link\" />\n" +
					"    </ul>\n" +
					"  </CardFooter>\n" +
					"</Card>",
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}, // h1
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}, // article
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}, // h2
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}, // ul
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}, // li
				},
			},
			// 6-level deep flat chain — every level reports independently.
			{
				Code: "<article role=\"button\">" +
					"<h1 role=\"button\">" +
					"<table role=\"button\">" +
					"<tbody role=\"button\">" +
					"<tr role=\"button\">" +
					"<td>x</td>" +
					"</tr>" +
					"</tbody>" +
					"</table>" +
					"</h1>" +
					"</article>",
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
					// <tr> is interactive (inside table) — no report on tr.
					// Actually wait: <tr> IS in interactiveElementRoleSchemas
					// (just `{name: "tr"}` unconditional). So <tr role="button">
					// is interactive + interactive role → NOT non-interactive
					// → bail at IsNonInteractiveElement → no report.
				},
			},
			// Sibling JSX in fragment — both report independently.
			{
				Code: "<>\n" +
					"  <article role=\"button\" />\n" +
					"  <h1 role=\"menuitem\" />\n" +
					"  <ul role=\"tab\" />\n" +
					"</>",
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleConditionalRendering locks
// JSX inside the full menagerie of conditional / list rendering
// patterns React developers actually write. Each pattern produces one
// report per matching JsxAttribute regardless of the surrounding control
// flow's static reachability.
func TestNoNoninteractiveElementToInteractiveRoleConditionalRendering(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule, []rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// `&&` short-circuit.
			{
				Code:   `const x = cond && <article role="button" />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// `??` nullish coalescing — the rule fires on the JSX even though
			// it's only rendered when LHS is nullish.
			{
				Code:   `const x = maybe ?? <article role="button" />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Ternary with BOTH branches having matching role — 2 reports.
			{
				Code: `const x = cond ? <article role="button" /> : <h1 role="menuitem" />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// switch case JSX returns.
			{
				Code: "function R(x: number) {\n" +
					"  switch (x) {\n" +
					"    case 1: return <article role=\"button\" />;\n" +
					"    case 2: return <h1 role=\"menuitem\" />;\n" +
					"    default: return null;\n" +
					"  }\n" +
					"}",
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// `.map()` with multiple sibling props.
			{
				Code: `const xs = items.map((item, i) => <article key={i} className="card" role="button" onClick={item.fn}>{item.label}</article>);`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// `.filter().map()` chain.
			{
				Code: `const xs = items.filter(Boolean).map(item => <li role="link" key={item.id} />);`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// IIFE returning JSX.
			{
				Code:   `const x = (() => <article role="button" />)();`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// JSX inside try/catch.
			{
				Code: "function R() {\n" +
					"  try { return <article role=\"button\" />; }\n" +
					"  catch { return null; }\n" +
					"}",
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleTSXTypeSystemShapes locks
// TSX-specific shapes that tsgo preserves but ESTree flattens or doesn't
// expose. Each case probes a tsgo AST node that COULD have been peeled
// to a string by mistake (silently changing classification) — `as`,
// `satisfies`, JSX generic type parameters, TS const assertions, etc.
func TestNoNoninteractiveElementToInteractiveRoleTSXTypeSystemShapes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// `as` casting the whole JSX element — outer cast doesn't affect
			// rule (rule fires on attribute regardless).
			// But role inside is non-literal so doesn't classify.
			{Code: `const x = <article role={someRole as Role} />;`, Tsx: true},
			// `satisfies` on the JSX element.
			{Code: `const x = <article role={someRole} /> satisfies JSX.Element;`, Tsx: true},
			// JSX assigned to typed const — type annotation is irrelevant to
			// the rule.
			{Code: `const x: JSX.Element = <article role={someRole} />;`, Tsx: true},
			// JSX inside generic function — type params don't affect the rule.
			{Code: "function R<T>(props: { item: T }) { return <article role={someRole} />; }", Tsx: true},
			// Generic-typed JSX component — `<Foo<string>>` resolves to "Foo"
			// (custom component) → bail at dom.has.
			{Code: `const x = <Foo<string> role="button" />;`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// `as const` inside role expression — literal value is preserved
			// at the source level; tsgo's TSAsExpression wraps it.
			// jsx-ast-utils' LITERAL_TYPES doesn't unwrap → null → no
			// classification → no report from the role-classifier path.
			// But IF a regression DID unwrap `as const` and extract "button",
			// this would report. So we lock the upstream behavior (NO report)
			// by putting it in valid. Mirror with valid_only test above.
			//
			// This invalid case probes the OTHER direction: `as JSX.Element`
			// on the WHOLE JSX — outer cast doesn't peel attributes, the
			// inner literal role="button" still fires normally.
			{
				Code:   `const x = <article role="button" /> as JSX.Element;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// JSX inside a TS function with return type — rule fires.
			{
				Code: "function R(): JSX.Element { return <article role=\"button\" />; }",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// JSX inside a TS interface-typed callback.
			{
				Code: "const handler: (() => JSX.Element) = () => <h1 role=\"button\" />;",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
			// `<!-- TS namespace import scenario -->` — JSX inside a TS
			// `namespace` block. tsgo preserves namespace; rule should fire.
			{
				Code: "namespace UI { export const X = <article role=\"button\" />; }",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleUnicodeInRoleValue locks
// non-ASCII handling in the `role` attribute value. tsgo's string
// machinery is byte-based; ESLint's is JS-string (UTF-16). Each row
// probes a different unicode shape that could classify differently
// under a Go reimplementation.
func TestNoNoninteractiveElementToInteractiveRoleUnicodeInRoleValue(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// CJK in role — not a real ARIA role → no classification.
			{Code: `<article role="按钮" />`, Tsx: true},
			// Emoji in role — same.
			{Code: `<article role="🔘" />`, Tsx: true},
			// Real role with trailing CJK — split on ASCII space only;
			// "button按钮" is one token, not in role set → no classification.
			{Code: `<article role="button按钮" />`, Tsx: true},
			// NBSP (U+00A0) instead of ASCII space — single token, not split.
			// Result: "button img" not a valid role → no classification.
			{Code: "<article role=\"button img\" />", Tsx: true},
			// Real role with trailing combining mark — single token, not in
			// role set → no classification.
			{Code: "<article role=\"buttoń\" />", Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Multi-role with ASCII space separator + CJK token — first valid
			// role "button" wins.
			{
				Code:   `<article role="button 按钮" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// CJK in role attribute name doesn't matter — the attribute name
			// is `role` (ASCII), the value is the only thing we read.
			{
				Code:   `<article role="link 按钮" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Emoji first, real role second — emoji isn't a real role, "button"
			// wins as first VALID role.
			{
				Code:   `<article role="🔘 button" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleAllowListMultiTag locks
// the rule's behavior under realistic multi-tag allow-list configs
// (closer to the actual upstream `recommended` preset shape). A
// regression where the allowlist map lookup misbehaved (e.g., Go map
// iteration order leaked into the result) would surface here as
// inconsistent reporting across siblings.
func TestNoNoninteractiveElementToInteractiveRoleAllowListMultiTag(t *testing.T) {
	multiTagAllowList := map[string]interface{}{
		"ul":       []interface{}{"menu"},
		"ol":       []interface{}{"menubar"},
		"li":       []interface{}{"menuitem", "menuitemradio"},
		"table":    []interface{}{"grid"},
		"td":       []interface{}{"gridcell"},
		"fieldset": []interface{}{"radiogroup"},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// Each tag exempts only its own configured roles.
			{Code: `<ul role="menu" />`, Tsx: true, Options: multiTagAllowList},
			{Code: `<ol role="menubar" />`, Tsx: true, Options: multiTagAllowList},
			{Code: `<li role="menuitem" />`, Tsx: true, Options: multiTagAllowList},
			{Code: `<li role="menuitemradio" />`, Tsx: true, Options: multiTagAllowList},
			{Code: `<table role="grid" />`, Tsx: true, Options: multiTagAllowList},
			{Code: `<td role="gridcell" />`, Tsx: true, Options: multiTagAllowList},
			{Code: `<fieldset role="radiogroup" />`, Tsx: true, Options: multiTagAllowList},
		},
		[]rule_tester.InvalidTestCase{
			// Cross-tag role miss — `<ul role="menubar">` not in ul's allow-list.
			{
				Code:    `<ul role="menubar" />`,
				Tsx:     true,
				Options: multiTagAllowList,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Tag not in allow-list at all — `<h1>` ignored by the allow-list.
			{
				Code:    `<h1 role="button" />`,
				Tsx:     true,
				Options: multiTagAllowList,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Multiple violations in one element tree — order-independent
			// reporting.
			{
				Code: "<>\n" +
					"  <ul role=\"menu\" />\n" +     // exempted by allow-list
					"  <ol role=\"menu\" />\n" +     // NOT in ol's allow-list (which has "menubar")
					"  <li role=\"menuitem\" />\n" + // exempted
					"  <table role=\"button\" />\n" + // NOT in table's allow-list (which has "grid")
					"</>",
				Tsx:     true,
				Options: multiTagAllowList,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
					{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage},
				},
			},
		})
}

// TestNoNoninteractiveElementToInteractiveRoleExtendedRoleClassification
// locks classification of roles outside ARIA 1.x core — DPub
// (`doc-*`), Graphics (`graphics-*`), abstract roles
// (`composite` / `widget` / …), and roles in the aria-query "interactive"
// subset. A regression in [interactiveRolesSet] / [allRolesSet] (e.g.,
// loss of DPub interactive roles) would silently let these slip past.
func TestNoNoninteractiveElementToInteractiveRoleExtendedRoleClassification(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNoninteractiveElementToInteractiveRoleRule,
		[]rule_tester.ValidTestCase{
			// DPub non-interactive on non-interactive element — no report.
			{Code: `<article role="doc-chapter" />`, Tsx: true},
			// Graphics role — none are in interactive set → no report.
			{Code: `<article role="graphics-document" />`, Tsx: true},
			// Abstract role on non-interactive element — abstract not in
			// interactive set → no report.
			{Code: `<article role="widget" />`, Tsx: true},
			{Code: `<h1 role="composite" />`, Tsx: true},
			// Toolbar is in BOTH interactive and non-interactive sets (load-
			// bearing dual-classification per aria-query upstream). For THIS
			// rule, `IsInteractiveRole` returns true → would report on a
			// non-interactive element. So this needs to be invalid below.
			// Here we lock the OPPOSITE: toolbar on a NON-DOM tag (custom
			// component) bails before role check.
			{Code: `<MyComponent role="toolbar" />`, Tsx: true},
			// Allow-list with DPub role.
			{
				Code:    `<li role="doc-backlink" />`,
				Tsx:     true,
				Options: map[string]interface{}{"li": []interface{}{"doc-backlink"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// DPub interactive role on non-interactive element — `doc-backlink`
			// is in interactiveRolesSet.
			{
				Code:   `<article role="doc-backlink" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			{
				Code:   `<h1 role="doc-glossref" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
			// Toolbar on a non-interactive element — toolbar IS in
			// interactiveRolesSet (load-bearing) → reports.
			{
				Code:   `<article role="toolbar" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNoninteractiveElementToInteractiveRole", Message: errorMessage}},
			},
		})
}
