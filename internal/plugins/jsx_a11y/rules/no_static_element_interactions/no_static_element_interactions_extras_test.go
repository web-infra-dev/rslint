package no_static_element_interactions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoStaticElementInteractionsPositions locks the diagnostic anchor: the
// rule reports on the JSX opening element node (paired form) or
// self-closing element node (single-tag form). Upstream's listener is
// `JSXOpeningElement` (in ESTree, fires once per element regardless of
// form), so the report range covers the opening tag — `<` through the
// matching `>` / `/>`. Two cases per shape: a single-line baseline and a
// multi-line variant where the opening tag spans several lines.
func TestNoStaticElementInteractionsPositions(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Self-closing — report spans <div … />.
		{
			Code: `<div onClick={() => {}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noStaticElementInteractions",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 1, EndColumn: 27,
			}},
		},
		// Paired form — report spans <div …> only (not the whole element).
		{
			Code: `<div onClick={() => {}}>label</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noStaticElementInteractions",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 1, EndColumn: 25,
			}},
		},
		// Multi-line self-closing.
		{
			Code: "<div\n  onClick={() => {}}\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noStaticElementInteractions",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 3, EndColumn: 3,
			}},
		},
		// Multi-line paired form.
		{
			Code: "<div\n  onClick={() => {}}\n>label</div>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noStaticElementInteractions",
				Message:   errorMessage,
				Line:      1, Column: 1, EndLine: 3, EndColumn: 2,
			}},
		},
	})
}

// TestNoStaticElementInteractionsListenerBoundary locks that the listener
// fires independently for each JSX opening element — nested JSX
// hierarchies produce one report per qualifying ancestor, with each
// report anchored at its own opening tag. Without this, a regression that
// "bleeds" the outer report into the inner traversal (or vice versa)
// would silently fold them into one diagnostic.
func TestNoStaticElementInteractionsListenerBoundary(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Three levels of nesting; outer + middle + inner each report.
		{
			Code: "<div onClick={() => {}}>\n  <span onClick={() => {}}>\n    <a onClick={() => {}} />\n  </span>\n</div>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1},
				{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 2, Column: 3},
				{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 3, Column: 5},
			},
		},
	})
}

// TestNoStaticElementInteractionsHandlersOption locks the `handlers` option
// branches that upstream's test suite doesn't directly exercise (apart from
// the implicit recommended-vs-strict comparison). Covers:
//   - User-supplied handler list (the directly-named handler matches; the
//     un-listed default handler does NOT).
//   - Explicit empty `handlers: []` — upstream's `[].some(...)` short-circuits
//     to false, so no handler ever matches and the rule never reports.
//   - Options JSON path: both bare-object (single-option CLI shape) and
//     array-wrapped (multi-element rule_tester shape).
func TestNoStaticElementInteractionsHandlersOption(t *testing.T) {
	customHandlersBareObject := map[string]interface{}{
		"handlers": []interface{}{"onCustomClick"},
	}
	customHandlersArrayWrapped := []interface{}{
		map[string]interface{}{
			"handlers": []interface{}{"onCustomClick"},
		},
	}
	emptyHandlersBareObject := map[string]interface{}{
		"handlers": []interface{}{},
	}
	emptyHandlersArrayWrapped := []interface{}{
		map[string]interface{}{
			"handlers": []interface{}{},
		},
	}

	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// onClick is NOT in the custom handler list → no interactive prop.
			{Code: `<div onClick={() => {}} />`, Tsx: true, Options: customHandlersBareObject},
			{Code: `<div onClick={() => {}} />`, Tsx: true, Options: customHandlersArrayWrapped},
			// Empty handlers list → upstream `[].some(...)` is false → no report.
			{Code: `<div onClick={() => {}} />`, Tsx: true, Options: emptyHandlersBareObject},
			{Code: `<div onClick={() => {}} />`, Tsx: true, Options: emptyHandlersArrayWrapped},
			{Code: `<div onClick={() => {}} onKeyDown={() => {}} />`, Tsx: true, Options: emptyHandlersBareObject},
		},
		[]rule_tester.InvalidTestCase{
			// Custom handler matches the rename.
			{
				Code:    `<div onCustomClick={() => {}} />`,
				Tsx:     true,
				Options: customHandlersBareObject,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticElementInteractions",
					Message:   errorMessage,
					Line:      1, Column: 1,
				}},
			},
			{
				Code:    `<div onCustomClick={() => {}} />`,
				Tsx:     true,
				Options: customHandlersArrayWrapped,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noStaticElementInteractions",
					Message:   errorMessage,
					Line:      1, Column: 1,
				}},
			},
		},
	)
}

// TestNoStaticElementInteractionsAllowExpressionValuesMatrix locks the
// `allowExpressionValues` branch upstream tests on a single fixture
// (`role={ROLE_BUTTON}`), expanded to the role-expression shapes
// IsNonLiteralProperty actually classifies.
//
// Upstream IsNonLiteralProperty:
//   - StringLiteral attribute value `role="button"`         → false (literal)
//   - JsxExpression Identifier `role={x}`                   → true (non-literal)
//   - JsxExpression Identifier `role={undefined}`           → false (literal-equivalent)
//   - JsxExpression JsxText (unreachable in practice)       → false
//   - JsxExpression anything-else (incl. Literal, Member,
//     Call, Conditional, BinaryExpression, ...)             → true (non-literal)
//
// Under allowExpressionValues=true, every "true" classification skips.
// Under allowExpressionValues=false (or absent), the rule continues to
// IsInteractiveRole / IsNonInteractiveRole / IsAbstractRole, which only
// resolve literal-typed string values — so all non-literal `role` shapes
// fall through to REPORT under defaults.
func TestNoStaticElementInteractionsAllowExpressionValuesMatrix(t *testing.T) {
	allowTrue := map[string]interface{}{"allowExpressionValues": true}
	allowFalse := map[string]interface{}{"allowExpressionValues": false}

	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// Identifier → non-literal → allowed.
			{Code: `<div role={ROLE} onClick={() => {}} />`, Tsx: true, Options: allowTrue},
			// MemberExpression → non-literal → allowed.
			{Code: `<div role={obj.role} onClick={() => {}} />`, Tsx: true, Options: allowTrue},
			// CallExpression → non-literal → allowed.
			{Code: `<div role={getRole()} onClick={() => {}} />`, Tsx: true, Options: allowTrue},
			// ConditionalExpression w/ two literals → non-literal under
			// upstream's IsNonLiteralProperty (Conditional ≠ Literal/Identifier-undefined/JSXText).
			{Code: `<div role={x ? "button" : "link"} onClick={() => {}} />`, Tsx: true, Options: allowTrue},
			// BinaryExpression → non-literal → allowed.
			{Code: `<div role={"button" + ""} onClick={() => {}} />`, Tsx: true, Options: allowTrue},
			// LogicalExpression → non-literal → allowed.
			{Code: `<div role={a || "button"} onClick={() => {}} />`, Tsx: true, Options: allowTrue},
			// `{null}` inside JsxExpression — upstream's IsNonLiteralProperty
			// treats anything other than StringLiteral / `undefined` / JsxText
			// as non-literal, so `null` also falls into the skip arm.
			{Code: `<div role={null} onClick={() => {}} />`, Tsx: true, Options: allowTrue},
		},
		[]rule_tester.InvalidTestCase{
			// allowExpressionValues=false — Identifier role can't resolve to
			// a literal, so IsInteractiveRole / IsNonInteractiveRole / IsAbstractRole
			// all return false → REPORT.
			{
				Code:    `<div role={ROLE} onClick={() => {}} />`,
				Tsx:     true,
				Options: allowFalse,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `<div role={undefined}>` — upstream's IsNonLiteralProperty
			// short-circuits on `Identifier 'undefined'` to FALSE (literal-equivalent).
			// allowExpressionValues=true cannot exempt this — it falls through
			// to IsInteractiveRole etc., none of which match → REPORT.
			{
				Code:    `<div role={undefined} onClick={() => {}} />`,
				Tsx:     true,
				Options: allowTrue,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `role="..."` (direct StringLiteral attribute) — IsNonLiteralProperty
			// short-circuits to FALSE (literal). The string is "button"
			// resolves through IsInteractiveRole; "notarole" doesn't. Pick
			// "notarole" so we trip REPORT regardless of allowExpressionValues.
			{
				Code:    `<div role="notarole" onClick={() => {}} />`,
				Tsx:     true,
				Options: allowTrue,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}

// TestNoStaticElementInteractionsEdgeShapes locks the Dimension-4 universal
// edge shapes our prep walk identified.
func TestNoStaticElementInteractionsEdgeShapes(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// `onClick={undefined}` — upstream `getPropValue` resolves to JS
			// undefined → `!= null` is false → no interactive prop → exempt.
			{Code: `<div onClick={undefined} />`, Tsx: true},
			// `<div onClick={null} />` — upstream getPropValue → JS null →
			// `!= null` is false → exempt. (Also in upstream suite.)
			{Code: `<div onClick={null} />`, Tsx: true},
			// `onClick={void 0}` — JS `void 0` evaluates to undefined; under
			// jsx-ast-utils' staticEval `void` produces undefined → `!= null`
			// false → exempt.
			{Code: `<div onClick={void 0} />`, Tsx: true},
			// `aria-hidden` (boolean-form) on a static element with click —
			// upstream short-circuits via isHiddenFromScreenReader.
			{Code: `<div onClick={() => {}} aria-hidden />`, Tsx: true},
			// Custom component (not in dom set) → exempt step 1.
			{Code: `<MyComp onClick={() => {}} />`, Tsx: true},
			// `<div role="presentation" onClick={...}>` — IsPresentationRole.
			{Code: `<div role="presentation" onClick={() => {}} />`, Tsx: true},
			// `<div role="none" onClick={...}>` — same as presentation.
			{Code: `<div role="none" onClick={() => {}} />`, Tsx: true},
			// Spread-only handler — hasProp / getProp default spreadStrict=true,
			// so `{...{onClick: () => {}}}` is INVISIBLE to the rule.
			{Code: `<div {...{onClick: () => {}}} />`, Tsx: true},
			// Direct handler + spread elsewhere — spread doesn't change the
			// resolution; onClick still resolves to the literal.
			{Code: `<div onClick={null} {...props} />`, Tsx: true},
			// Boolean-form handler — but with a literal value that resolves
			// `null`-ish. Boolean-form `<div onClick />` → upstream extractValue
			// resolves to JS `true`. `true != null` → TRUE → REPORTS. Locked
			// in invalid section below.
			//
			// Inherent-interactive container: `<th>` with onClick is exempt
			// via IsInteractiveElement.
			{Code: `<th onClick={() => {}} />`, Tsx: true},
			// Mixed-case role — IsInteractiveRole lowercases before lookup;
			// `<div role="BUTTON" onClick={...}>` → exempt.
			{Code: `<div role="BUTTON" onClick={() => {}} />`, Tsx: true},
			// Role template literal — `role={`button`}` → LiteralPropStringValue
			// extracts "button" → IsInteractiveRole returns true → exempt.
			{Code: "<div role={`button`} onClick={() => {}} />", Tsx: true},
			// Role wrapped StringLiteral `role={"button"}` — literalPropValue
			// returns "button" → exempt via IsInteractiveRole.
			{Code: `<div role={"button"} onClick={() => {}} />`, Tsx: true},
			// Space-separated roles where the first valid role is interactive.
			{Code: `<div role="button heading" onClick={() => {}} />`, Tsx: true},
			// Space-separated roles where the first valid role is non-interactive
			// (article) — IsNonInteractiveRole short-circuits before
			// IsAbstractRole, but the result is still exempt.
			{Code: `<div role="article button" onClick={() => {}} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Boolean-form handler `<div onClick />` — upstream extractValue
			// resolves to JS `true`; `true != null` → matches → REPORT.
			{
				Code:   `<div onClick />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// onFocus is in the default focus+keyboard+mouse list (only
			// reported under :strict / no-options).
			{
				Code:   `<div onFocus={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `<header>` always reports — IsNonInteractiveElement returns
			// false (upstream's banner-landmark-context guard).
			{
				Code:   `<header onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Space-separated roles where the first valid role is unknown
			// — falls through every classification → REPORT.
			{
				Code:   `<div role="notarole" onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Multiple handlers on the same element — ONE diagnostic per
			// element (rule reports on the element node, not per-handler).
			{
				Code:   `<div onClick={() => {}} onKeyDown={() => {}} onMouseUp={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}

// TestNoStaticElementInteractionsOptionParsing locks the Options JSON path —
// the rule MUST accept both the bare-object form (matches the single-option
// CLI shape after config.go unwraps the array) and the array-wrapped form
// (matches the multi-element rule_tester shape). Also covers nil / empty /
// malformed shapes so a future refactor of parseOptions can't regress them.
func TestNoStaticElementInteractionsOptionParsing(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// nil options — defaults: full focus+keyboard+mouse handlers,
			// allowExpressionValues=falsy. Validates the "options absent"
			// path doesn't short-circuit to no-default-handlers.
			{Code: `<TestComponent onClick={() => {}} />`, Tsx: true},
			// Empty options object — same as nil. allowExpressionValues
			// stays false; handlers stays as default.
			{Code: `<TestComponent onClick={() => {}} />`, Tsx: true, Options: map[string]interface{}{}},
			{Code: `<TestComponent onClick={() => {}} />`, Tsx: true, Options: []interface{}{}},
			// Malformed handlers value — non-array shape should be treated
			// as "user provided something" but produce an empty list →
			// never reports.
			{Code: `<div onClick={() => {}} />`, Tsx: true, Options: map[string]interface{}{"handlers": "not-an-array"}},
			{Code: `<div onClick={() => {}} />`, Tsx: true, Options: map[string]interface{}{"handlers": 0}},
			// Malformed allowExpressionValues value — non-bool ignored →
			// stays falsy. Rule path falls through (no skip), but onClick on
			// <button> is exempt via IsInteractiveElement → valid.
			{Code: `<button onClick={() => {}} />`, Tsx: true, Options: map[string]interface{}{"allowExpressionValues": "true"}},
		},
		[]rule_tester.InvalidTestCase{
			// Single-option bare-object form (CLI-shape after config.go unwrap).
			{
				Code:    `<div onClick={() => {}} />`,
				Tsx:     true,
				Options: map[string]interface{}{"allowExpressionValues": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Array-wrapped form (multi-element rule_tester shape).
			{
				Code:    `<div onClick={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"allowExpressionValues": false}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `handlers: ["onCustomClick"]` matches and reports.
			{
				Code:    `<div onCustomClick={() => {}} />`,
				Tsx:     true,
				Options: map[string]interface{}{"handlers": []interface{}{"onCustomClick"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}

// TestNoStaticElementInteractionsTSWrappers locks tsgo-only AST shapes that
// upstream's ESTree parser doesn't produce — TS non-null assertion, `as`
// cast, satisfies, parenthesized expressions on the handler / role values.
// Upstream's staticEval (via PropValueIsNullish / LiteralPropStringValue)
// is expected to look through these wrappers; the explicit cases protect
// against a regression where the wrapper unwrap is skipped.
func TestNoStaticElementInteractionsTSWrappers(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// Function-cast handler resolves to a non-null function → upstream
			// reports, but the element is `<button>` (inherently interactive) → exempt.
			{Code: `<button onClick={(() => {}) as any} />`, Tsx: true},
			// `role` value through `as` cast → LiteralPropStringValue
			// returns "" (literalPropValue routes through OEKNonNullAssertions
			// strip; `as` is dropped via SkipOuterExpressions). Confirm:
			// `<div role={"button" as any} onClick={...}>` should exempt via
			// IsInteractiveRole if our LiteralPropStringValue handles `as`.
			//
			// Upstream's `getLiteralPropValue` maps TSAsExpression to noop → null,
			// so the role wouldn't resolve to "button" upstream. We mirror
			// upstream's stricter behavior — REPORT this case (locked in
			// invalid section below).
			//
			// Parenthesized handler resolves through the paren — `<button>`
			// is inherently interactive → exempt.
			{Code: `<button onClick={((() => {}))} />`, Tsx: true},
			// Parenthesized role value `role={("button")}` → LiteralPropStringValue
			// handles paren unwrap → exempt via IsInteractiveRole.
			{Code: `<div role={("button")} onClick={() => {}} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// `role` through `as` cast → upstream's `getLiteralPropValue`
			// returns null → IsInteractiveRole returns false → REPORT.
			{
				Code:   `<div role={"button" as any} onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `role` through non-null assertion → same as `as` cast.
			{
				Code:   `<div role={"button"!} onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Parenthesized handler on static element — resolves through
			// the paren → onClick != null → REPORT.
			{
				Code:   `<div onClick={((() => {}))} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}

// TestNoStaticElementInteractionsUpstreamBranchLockIns locks behavioral
// branches that exist in the upstream source but are not covered by the
// upstream test file. Each test below names the upstream code branch it
// pins; a future refactor that flips the branch behavior fails these tests.
func TestNoStaticElementInteractionsUpstreamBranchLockIns(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream `hasInteractiveProps` short-circuit arm
			// `getPropValue(...) != null` — even though `<div onChange={x} />`
			// has a non-null `onChange`, `onChange` is NOT in any of the
			// default focus/keyboard/mouse lists → handler.some(...) is false.
			{Code: `<div onChange={() => {}} />`, Tsx: true},

			// Locks in upstream `isAbstractRole` arm — `role="widget"` etc.
			// short-circuit before `isInteractiveRole` / `isNonInteractiveRole`.
			// `widget` is abstract but isn't in either set, so without the
			// IsAbstractRole check the rule would fall through and REPORT.
			{Code: `<div role="widget" onClick={() => {}} />`, Tsx: true},

			// Locks in upstream IsHiddenFromScreenReader's `<input
			// type="hidden">` branch (the type-aware path, not aria-hidden).
			// `<input>` is inherently interactive too — IsInteractiveElement
			// short-circuits before reaching the hidden-screen-reader check —
			// so flip to `<input type="hidden" onCopy={...} />` (onCopy is
			// not in any default handler list, but onClick is). Use onClick
			// here to ensure IsHiddenFromScreenReader is what saves us.
			{Code: `<input type="hidden" onClick={() => {}} />`, Tsx: true},

			// Locks in upstream `isInteractiveElement` first-match arm —
			// `<input type="button">` matches the `{name: "input",
			// attributes: [{name: "type", value: "button"}]}` schema and
			// exits before IsInteractiveRole even runs.
			{Code: `<input type="button" onClick={() => {}} />`, Tsx: true},

			// Locks in upstream `isInteractiveElement`'s no-attribute schema
			// path — `<button onClick={...}>` matches `{name: "button"}` with
			// no required attributes (vacuous every-true predicate).
			{Code: `<button onClick={() => {}} />`, Tsx: true},

			// Locks in upstream IsNonInteractiveRole's "first valid role" walk
			// — `role="invalid article"` skips "invalid" (not in allRolesSet)
			// then matches "article" as non-interactive.
			{Code: `<div role="invalid article" onClick={() => {}} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream's behavior on `<div role="">` (empty string
			// — no valid role found). IsInteractiveRole / IsNonInteractiveRole /
			// IsAbstractRole all see "" as the literal value; none match → REPORT.
			{
				Code:   `<div role="" onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},

			// Locks in upstream's IsInteractiveRole "first valid role" walk
			// — `role="invalid notarole"` skips both ("invalid" / "notarole"
			// not in allRolesSet) → returns false. Same for non-interactive
			// and abstract → REPORT.
			{
				Code:   `<div role="invalid notarole" onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},

			// Locks the listener distinguishing self-closing from paired form
			// without a "wrapper" indirection that would hide the report on
			// JsxOpeningElement when the parent is a JsxElement.
			{
				Code:   `<div onClick={() => {}}></div>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 25}},
			},
		},
	)
}

// TestNoStaticElementInteractionsRealHandlerShapes locks the staticEval-driven
// classification of handler values that appear in real React codebases —
// member access (`this.handleClick`), call expressions (`useCallback(...)`),
// conditional expressions, logical operators, nullish coalescing, async /
// generator / class-field arrow function values, and bound methods. Each
// shape is checked under the "handler is statically non-null" gate
// (`PropValueIsNullish`), so a regression in staticEval that flips one of
// these classifications would silently change the rule's behavior on the
// most common React patterns.
//
// Reference: jsxa11yutil/static_eval.go covers Identifier (jvString of name),
// PropertyAccessExpression / ElementAccessExpression (jvString "(member)"),
// CallExpression (jvString "(call)"), ArrowFunction / FunctionExpression /
// ClassExpression (jvFn), BinaryExpression with `&&` / `||` / `??`,
// ConditionalExpression (eager-eval the matched arm), and NonNullExpression
// (string of inner + "!"). NewExpression / Object / Array map to truthy.
func TestNoStaticElementInteractionsRealHandlerShapes(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// `cond ? null : null` — both arms null; ConditionalExpression
			// statically evaluates the truthy arm (cond resolves truthy as
			// Identifier jvString), result is jvNull → nullish → no handler match.
			{Code: `<div onClick={cond ? null : null} />`, Tsx: true},
			// Truthy condition + null arm: cond is Identifier (truthy) → eager
			// eval whenTrue → jvNull → nullish → no match.
			{Code: `<div onClick={cond ? null : someHandler} />`, Tsx: true},
			// Truthy condition + undefined arm: same as above with jvUndef.
			{Code: `<div onClick={cond ? undefined : someHandler} />`, Tsx: true},
			// Logical `&&` with falsy left — left is falsy (false literal),
			// staticEvalBinary returns left (jvBool false) → !jsTruthy → nullish? NO!
			// jvBool(false) is NOT jvNull/jvUndef/jvUnknown, so PropValueIsNullish
			// returns false → has handler → REPORT.
			//
			// (Locked under invalid below.)
			//
			// Logical `||` with truthy left: returns left → jvString from Identifier → REPORT.
			//
			// Nullish coalesce where left is non-nullable: returns left.
			//
			// `<a href>` boolean form — schema only checks attribute existence,
			// not value, so this matches interactive schema → exempt.
			{Code: `<a href onClick={() => {}} />`, Tsx: true},
			// `<a href="">` — empty href string still satisfies the
			// "attribute exists" schema check → interactive → exempt.
			{Code: `<a href="" onClick={() => {}} />`, Tsx: true},
			// `<a href={undefined}>` — attribute exists, schema matches → exempt.
			{Code: `<a href={undefined} onClick={() => {}} />`, Tsx: true},
			// `<a href={null}>` — same as above, attribute existence only.
			{Code: `<a href={null} onClick={() => {}} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Member access — staticEval returns jvString "(member)" → truthy
			// → not nullish → REPORT. Locks the most common real-world handler shape.
			{
				Code:   `<div onClick={this.handleClick} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Deep member access (handlers.section.onClick).
			{
				Code:   `<div onClick={handlers.section.onClick} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Computed member access (`handlers[name]`).
			{
				Code:   `<div onClick={handlers[name]} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Call expression (the useCallback / useMemo / withErrorBoundary
			// pattern). staticEval returns jvString "(call)" → REPORT.
			{
				Code:   `<div onClick={getHandler()} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			{
				Code:   `<div onClick={useCallback(() => {}, [])} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Bound method — CallExpression of `.bind(this)`.
			{
				Code:   `<div onClick={this.handler.bind(this)} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Optional chain on member — tsgo flags the optional, kind is
			// still PropertyAccessExpression → jvString → REPORT.
			{
				Code:   `<div onClick={handlers?.onClick} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Async arrow function — ArrowFunction kind → jvFn → REPORT.
			{
				Code:   `<div onClick={async () => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Function expression.
			{
				Code:   `<div onClick={function() {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Named function expression.
			{
				Code:   `<div onClick={function named() {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Identifier alone (most common real-world shape — destructured
			// handler bound at the top of the component).
			{
				Code:   `<div onClick={memoizedHandler} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Conditional with non-null arms — both arms are truthy via
			// Identifier resolution; ConditionalExpression eagerly evaluates
			// the matched arm (cond is Identifier, jvTruthy → eval whenTrue).
			{
				Code:   `<div onClick={cond ? handler1 : handler2} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Logical `&&` with non-null left — staticEvalBinary's `&&`
			// returns the right when left truthy; Identifier left → truthy
			// → right Identifier → jvString → REPORT.
			{
				Code:   `<div onClick={cond && handler} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Logical `||` — returns left if truthy; Identifier left truthy
			// → returns left → jvString → REPORT.
			{
				Code:   `<div onClick={handler || defaultHandler} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Nullish coalesce — left is non-null Identifier (jvString) →
			// staticEvalBinary returns left → REPORT.
			{
				Code:   `<div onClick={handler ?? defaultHandler} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// New expression — staticEval returns jsTruthy → not nullish.
			{
				Code:   `<div onClick={new Handler()} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Inline arrow with side-effect body.
			{
				Code:   `<div onClick={(e) => { e.stopPropagation(); handler(e); }} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}

// TestNoStaticElementInteractionsPolymorphicSettings locks the
// `jsx-a11y.polymorphicPropName` / `polymorphicAllowList` settings: a custom
// component is resolved through its polymorphic prop (`as`, `is`, `forwardedAs`, ...)
// before the DOM-set membership check, mirroring upstream getElementType's
// behavior. Real-world design systems rely on this resolution pattern to
// expose a single component that renders to different underlying HTML tags.
func TestNoStaticElementInteractionsPolymorphicSettings(t *testing.T) {
	asPolymorphic := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"polymorphicPropName": "as",
		},
	}
	asPolymorphicAllowList := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"polymorphicPropName":  "as",
			"polymorphicAllowList": []interface{}{"Box"},
		},
	}
	isPolymorphic := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"polymorphicPropName": "is",
		},
	}

	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// `<Foo as="button">` — resolves to "button" (interactive) → exempt.
			{Code: `<Foo as="button" onClick={() => {}} />`, Tsx: true, Settings: asPolymorphic},
			// `<Foo as="article">` — non-interactive element → exempt.
			{Code: `<Foo as="article" onClick={() => {}} />`, Tsx: true, Settings: asPolymorphic},
			// Different polymorphic prop name (`is` instead of `as`).
			{Code: `<Foo is="button" onClick={() => {}} />`, Tsx: true, Settings: isPolymorphic},
			// AllowList: `Box` is in the list, gets polymorphic resolution.
			//   `<Box as="button">` → "button" → exempt.
			{Code: `<Box as="button" onClick={() => {}} />`, Tsx: true, Settings: asPolymorphicAllowList},
			// AllowList: `Foo` is NOT in the list, polymorphic resolution
			// is SKIPPED → rawType stays "Foo" (custom component) → exempt
			// step 1 (not in dom).
			{Code: `<Foo as="div" onClick={() => {}} />`, Tsx: true, Settings: asPolymorphicAllowList},
			// `<Foo>` without polymorphic prop set — no rewrite, stays "Foo"
			// (custom component) → exempt.
			{Code: `<Foo onClick={() => {}} />`, Tsx: true, Settings: asPolymorphic},
			// Polymorphic prop with non-literal value — can't resolve → stays "Foo" → exempt.
			{Code: `<Foo as={tagName} onClick={() => {}} />`, Tsx: true, Settings: asPolymorphic},
		},
		[]rule_tester.InvalidTestCase{
			// `<Foo as="div">` — resolves to "div" (static) → REPORT.
			{
				Code:     `<Foo as="div" onClick={() => {}} />`,
				Tsx:      true,
				Settings: asPolymorphic,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `<Foo as="span">` — span is static → REPORT.
			{
				Code:     `<Foo as="span" onClick={() => {}} />`,
				Tsx:      true,
				Settings: asPolymorphic,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// AllowList: Box is in the list, resolves to "div" → REPORT.
			{
				Code:     `<Box as="div" onClick={() => {}} />`,
				Tsx:      true,
				Settings: asPolymorphicAllowList,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Custom polymorphic prop name `is`.
			{
				Code:     `<Foo is="span" onClick={() => {}} />`,
				Tsx:      true,
				Settings: isPolymorphic,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}

// TestNoStaticElementInteractionsComponentsRemapping locks the
// `jsx-a11y.components` map: a JSX component name is rewritten to its
// configured HTML element before the DOM-set + interactivity checks. This
// is the canonical way teams say "our `<TextLink>` is an `<a href>`" — the
// rule must honor the mapping in both directions (remap to interactive ⇒
// exempt; remap to static ⇒ report).
func TestNoStaticElementInteractionsComponentsRemapping(t *testing.T) {
	componentsMap := map[string]interface{}{
		"jsx-a11y": map[string]interface{}{
			"components": map[string]interface{}{
				"Box":      "div",
				"Heading":  "h1",
				"TextLink": "a",
				"MyButton": "button",
				"Input":    "input",
				"Stack":    "section",
			},
		},
	}

	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// MyButton → button (inherently interactive).
			{Code: `<MyButton onClick={() => {}} />`, Tsx: true, Settings: componentsMap},
			// Input → input (inherently interactive).
			{Code: `<Input onClick={() => {}} />`, Tsx: true, Settings: componentsMap},
			// TextLink with href (interactive `<a>` schema).
			{Code: `<TextLink href="/" onClick={() => {}} />`, Tsx: true, Settings: componentsMap},
			// Heading → h1 (non-interactive HTML element).
			{Code: `<Heading onClick={() => {}} />`, Tsx: true, Settings: componentsMap},
			// Box → div with interactive role → exempt.
			{Code: `<Box role="button" onClick={() => {}} />`, Tsx: true, Settings: componentsMap},
			// Box with presentation role → exempt.
			{Code: `<Box role="presentation" onClick={() => {}} />`, Tsx: true, Settings: componentsMap},
			// Stack → section with aria-label → exempt (non-interactive schema).
			{Code: `<Stack aria-label="A" onClick={() => {}} />`, Tsx: true, Settings: componentsMap},
		},
		[]rule_tester.InvalidTestCase{
			// Box → div (static, no role) → REPORT.
			{
				Code:     `<Box onClick={() => {}} />`,
				Tsx:      true,
				Settings: componentsMap,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// TextLink → a (no href, no role) → REPORT.
			{
				Code:     `<TextLink onClick={() => {}} />`,
				Tsx:      true,
				Settings: componentsMap,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Stack → section (without aria-label, falls through non-interactive
			// schema) → REPORT.
			{
				Code:     `<Stack onClick={() => {}} />`,
				Tsx:      true,
				Settings: componentsMap,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}

// TestNoStaticElementInteractionsRoleSemantics locks the role-attribute
// classification semantics: case-insensitive role lookup, space-separated
// multi-role first-wins, the FIRST valid (in allRolesSet) role wins, and
// unknown roles fall through to REPORT.
//
// These branches exist in upstream's IsInteractiveRole / IsNonInteractiveRole /
// IsAbstractRole but the upstream test file only covers them at the
// per-role level — not the cross-class interaction (e.g. abstract role
// listed first vs second) or the case-sensitivity boundary.
func TestNoStaticElementInteractionsRoleSemantics(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// Mixed-case roles — upstream lowercases before lookup.
			{Code: `<div role="Button" onClick={() => {}} />`, Tsx: true},
			{Code: `<div role="BUTTON" onClick={() => {}} />`, Tsx: true},
			{Code: `<div role="BuTtOn" onClick={() => {}} />`, Tsx: true},
			// Multi-role, first valid is interactive — exempt via IsInteractiveRole.
			{Code: `<div role="button heading" onClick={() => {}} />`, Tsx: true},
			{Code: `<div role="link article" onClick={() => {}} />`, Tsx: true},
			// Multi-role, first valid is non-interactive (article) — exempt
			// via IsNonInteractiveRole; the trailing button does not "rescue".
			{Code: `<div role="article button" onClick={() => {}} />`, Tsx: true},
			// Skip invalid role names in the space-split, find first valid.
			{Code: `<div role="invalid button" onClick={() => {}} />`, Tsx: true},
			{Code: `<div role="invalid notarole article" onClick={() => {}} />`, Tsx: true},
			// First valid is abstract → exempt via IsAbstractRole.
			{Code: `<div role="widget button" onClick={() => {}} />`, Tsx: true},
			// Cross-class: inherently interactive element with non-interactive
			// role attribute — IsInteractiveElement fires first.
			{Code: `<button role="article" onClick={() => {}} />`, Tsx: true},
			// Inherently interactive `<a href>` with abstract role on it.
			{Code: `<a href="/" role="widget" onClick={() => {}} />`, Tsx: true},
			// Inherently non-interactive `<article>` with interactive role.
			{Code: `<article role="button" onClick={() => {}} />`, Tsx: true},
			// `role` with leading/trailing whitespace — String(value)
			// .toLowerCase().split(' ') preserves an empty leading entry
			// from " button" which fails allRolesSet, then finds "button".
			{Code: `<div role=" button" onClick={() => {}} />`, Tsx: true},
			// Multi-role with abstract first then non-interactive.
			{Code: `<div role="widget article" onClick={() => {}} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Empty role attribute — split → [""], filter to valid → empty →
			// none of the classifications match → REPORT.
			{
				Code:   `<div role="" onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// All invalid role names.
			{
				Code:   `<div role="bogus another-bogus" onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Whitespace-only role.
			{
				Code:   `<div role="   " onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}

// TestNoStaticElementInteractionsAriaHiddenVariants locks the
// IsHiddenFromScreenReader gate's value classification — `=== true` is
// upstream's strict comparison, so only the actual JS boolean true (or
// jsxAstUtilsLiteralCoerce'd "true"/boolean-form) exempts.
func TestNoStaticElementInteractionsAriaHiddenVariants(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// Boolean form, expression, string-coerced — all exempt.
			{Code: `<div onClick={() => {}} aria-hidden />`, Tsx: true},
			{Code: `<div onClick={() => {}} aria-hidden={true} />`, Tsx: true},
			{Code: `<div onClick={() => {}} aria-hidden="true" />`, Tsx: true},
			// String "TRUE" case-insensitively coerced — exempt.
			{Code: `<div onClick={() => {}} aria-hidden="TRUE" />`, Tsx: true},
			// Wrapped string literal in expression — staticEval routes through
			// "true" coercion → boolean true.
			{Code: `<div onClick={() => {}} aria-hidden={"true"} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// `aria-hidden={false}` — strict `=== true` fails → NOT hidden → REPORT.
			{
				Code:   `<div onClick={() => {}} aria-hidden={false} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `aria-hidden="false"` — string coerces to bool false → NOT hidden → REPORT.
			{
				Code:   `<div onClick={() => {}} aria-hidden="false" />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `aria-hidden={1}` — number 1 is NOT strict `=== true` → NOT hidden → REPORT.
			{
				Code:   `<div onClick={() => {}} aria-hidden={1} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `aria-hidden={someVar}` — Identifier resolves to jvString (not bool)
			// → NOT hidden → REPORT.
			{
				Code:   `<div onClick={() => {}} aria-hidden={isHidden} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `aria-hidden={null}` — null is NOT strict true → NOT hidden → REPORT.
			{
				Code:   `<div onClick={() => {}} aria-hidden={null} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}

// TestNoStaticElementInteractionsInputSchemaVariants locks every
// interactiveElementRoleSchemas entry for `<input>` — the most-tested
// element shape upstream covers, but the cross-product with the rule's
// "static handler" check (vs interactive-supports-focus's "interactive
// role" check) deserves an independent suite.
//
// Each input type in the interactive schema must be exempt under
// no-static-element-interactions because IsInteractiveElement returns
// true → step 3 of the rule short-circuits.
func TestNoStaticElementInteractionsInputSchemaVariants(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// `<input>` with no `type` — the input schema is matched via
			// `{name: "input"}` schemas (multiple schemas with attribute
			// predicates; the bare `<input>` falls through every type-aware
			// schema. BUT upstream considers `<input onClick>` valid; the
			// listener matches the broader input schema in aria-query.
			//
			// In our impl, the broader `<input>` (no type) does NOT match
			// the type-aware schemas. Whether it ends up exempt depends on
			// other classifications. Upstream's test:
			//   { code: '<input onClick={() => void 0} />' }
			// is in alwaysValid, so we expect exempt.
			{Code: `<input onClick={() => {}} />`, Tsx: true},
			// All `type` values that appear in interactiveElementRoleSchemas
			// or are general inputs (covered by interactive schemas).
			{Code: `<input type="text" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="email" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="search" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="tel" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="url" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="password" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="button" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="checkbox" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="image" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="number" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="radio" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="range" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="reset" onClick={() => {}} />`, Tsx: true},
			{Code: `<input type="submit" onClick={() => {}} />`, Tsx: true},
			// `type="hidden"` — exempt via IsHiddenFromScreenReader BEFORE
			// the interactive-element check.
			{Code: `<input type="hidden" onClick={() => {}} />`, Tsx: true},
			// Direct-string type attribute case-insensitivity.
			{Code: `<input type="TEXT" onClick={() => {}} />`, Tsx: true},
			// Type as wrapped expression `type={"button"}` — LiteralPropStringValue
			// resolves through JsxExpression → "button" → schema matches.
			{Code: `<input type={"button"} onClick={() => {}} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

// TestNoStaticElementInteractionsMultiHandlerCombinations locks the
// outer-loop-over-handlers semantics: per handler, the FIRST matching
// direct JsxAttribute decides whether that handler counts. Mixed null /
// non-null combinations across different handler names exercise the
// short-circuit boundary.
func TestNoStaticElementInteractionsMultiHandlerCombinations(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// All handlers null — no match anywhere.
			{Code: `<div onClick={null} onMouseDown={null} onKeyDown={null} />`, Tsx: true},
			// All handlers undefined.
			{Code: `<div onClick={undefined} onMouseDown={undefined} onKeyDown={undefined} />`, Tsx: true},
			// Mix of null + undefined + void 0.
			{Code: `<div onClick={null} onMouseDown={undefined} onKeyDown={void 0} />`, Tsx: true},
			// Null handler + non-handler-list prop (onChange not in default).
			{Code: `<div onClick={null} onChange={() => {}} />`, Tsx: true},
			// Null handler + non-handler-list prop (multiple).
			{Code: `<div onClick={null} onChange={() => {}} onScroll={() => {}} onSubmit={() => {}} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// First handler non-null wins → REPORT.
			{
				Code:   `<div onClick={() => {}} onMouseDown={null} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Later handler non-null also wins (handler iteration order).
			{
				Code:   `<div onClick={null} onMouseDown={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// All handlers non-null — only one report on the element.
			{
				Code:   `<div onClick={() => {}} onKeyDown={() => {}} onMouseDown={() => {}} onKeyUp={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Mixed null/non-null where the FIRST listed in attrs is null,
			// later one is non-null.
			{
				Code:   `<div onKeyDown={null} onClick={null} onMouseUp={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Non-handler-list prop is non-null but a handler-list prop is also
			// non-null → REPORT.
			{
				Code:   `<div onChange={() => {}} onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}

// TestNoStaticElementInteractionsJsxStructuralEdgeCases locks JSX
// structural shapes that real codebases use but upstream's test file does
// not exercise — fragments wrapping reportable elements, attribute-level
// comments, generic JSX component invocations, and ternary-rendered JSX
// children. Each pattern should report only on the qualifying inner JSX
// element, leaving the wrapper / fragment / Generic / conditional shell
// untouched.
func TestNoStaticElementInteractionsJsxStructuralEdgeCases(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// Empty fragment — listener doesn't fire on fragment opening.
			{Code: `<></>`, Tsx: true},
			// Fragment with non-reportable children.
			{Code: `<><div /><button onClick={() => {}} /></>`, Tsx: true},
			// Generic JSX component invocation — not in dom set → exempt.
			{Code: `<Component<string> onClick={() => {}} />`, Tsx: true},
			{Code: `<Generic<A, B> onClick={() => {}} />`, Tsx: true},
			// JSX as render-prop child (the prop's value is JSX, the OUTER
			// is React.Fragment — neither qualifying).
			{Code: `<><div /></>`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Fragment wrapping a reportable inner element — only the inner reports.
			{
				Code:   `<><div onClick={() => {}} /></>`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 3}},
			},
			// Fragment with two reportable children — two reports.
			{
				Code: `<><div onClick={() => {}} /><span onClick={() => {}} /></>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 3},
					{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 29},
				},
			},
			// Ternary-rendered JSX where only the static branch qualifies.
			//   `<button>` is interactive (exempt).
			//   `<div onClick>` qualifies → REPORT.
			{
				Code: `<>{cond ? <button onClick={() => {}} /> : <div onClick={() => {}} />}</>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 43},
				},
			},
			// Attribute-level comment between handler attrs — comments are
			// trivia, AST walking must still find the onClick attribute.
			{
				Code:   "<div /* a */ onClick={() => {}} /* b */ />",
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// Sibling reportable elements at different lines.
			{
				Code: "<div>\n  <div onClick={() => {}} />\n  <span onClick={() => {}} />\n</div>",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 2, Column: 3},
					{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 3, Column: 3},
				},
			},
			// Conditional rendering: `{cond && <span onClick={...} />}` inside
			// a container — only the span reports.
			{
				Code: `<div>{cond && <span onClick={() => {}} />}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 15},
				},
			},
			// Logical OR rendering pattern.
			{
				Code: `<div>{primary || <span onClick={() => {}} />}</div>`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 18},
				},
			},
		},
	)
}

// TestNoStaticElementInteractionsAnchorSchemaVariants locks the `<a>`
// interactive schema — the `{name: "a", attributes: [{name: "href"}]}`
// predicate matches on "attribute exists" regardless of value, so
// `<a href={anything}>` is inherently interactive. Upstream tests only
// cover `<a href="http://...">` and bare `<a>`; the value-shape boundary
// (empty string, null/undefined value, boolean form, expression) is on us.
func TestNoStaticElementInteractionsAnchorSchemaVariants(t *testing.T) {
	rule_tester.RunRuleTesterBatched(fixtures.GetRootDir(), "tsconfig.json", t, &NoStaticElementInteractionsRule,
		[]rule_tester.ValidTestCase{
			// Plain URL.
			{Code: `<a href="/" onClick={() => {}} />`, Tsx: true},
			{Code: `<a href="http://example.com" onClick={() => {}} />`, Tsx: true},
			{Code: `<a href="#anchor" onClick={() => {}} />`, Tsx: true},
			// Empty href — schema only checks attribute existence.
			{Code: `<a href="" onClick={() => {}} />`, Tsx: true},
			// Boolean form `<a href>` — JsxAttribute exists, schema matches.
			{Code: `<a href onClick={() => {}} />`, Tsx: true},
			// Expression-typed href (Identifier).
			{Code: `<a href={url} onClick={() => {}} />`, Tsx: true},
			// Expression with null/undefined.
			{Code: `<a href={null} onClick={() => {}} />`, Tsx: true},
			{Code: `<a href={undefined} onClick={() => {}} />`, Tsx: true},
			// Template literal href.
			{Code: "<a href={`/path/${id}`} onClick={() => {}} />", Tsx: true},
			// Ternary href.
			{Code: `<a href={cond ? "/a" : "/b"} onClick={() => {}} />`, Tsx: true},
			// `<area href>` — same interactive schema as `<a href>`.
			{Code: `<area href="/" onClick={() => {}} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// `<a>` with no href — not in interactive schema → REPORT.
			{
				Code:   `<a onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
			// `<area>` with no href — same.
			{
				Code:   `<area onClick={() => {}} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStaticElementInteractions", Message: errorMessage, Line: 1, Column: 1}},
			},
		},
	)
}
