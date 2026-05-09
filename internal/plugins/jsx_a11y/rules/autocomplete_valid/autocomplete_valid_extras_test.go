// cspell:words bday impp

package autocomplete_valid

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestAutocompleteValidExtras covers behavior beyond the upstream
// `__tests__/src/rules/autocomplete-valid-test.js` suite — primarily:
//
//   - axe-core grammar boundaries that upstream's tests don't probe
//     (section-* length boundary, qualifier+standalone exclusion, double
//     qualifier / location, webauthn alone, ignoredValues, every stateTerm
//     extension, every standalone-term we ship, …)
//   - JSX-shape variants (paired vs self-closing, JsxExpression wrappers,
//     spread-of-literal, multi-line, deeply nested in map/conditionals/
//     hooks/class methods, etc.)
//   - tsgo-specific lock-ins (satisfies / `as const` / non-null
//     unwrapping, OEKAssertions bits, optional-chain noop, template-literal
//     `${undefined}` synthesis fix, …)
//   - axe-core matches() type filter on every excluded type, with literal
//     and JsxExpression-wrapped type forms, plus the negative direction
//     (non-excluded types still report)
//   - rslint-vs-upstream call-shape parity — runVirtualRule only forwards
//     `{autocomplete, type}`, so the rule must NOT skip on `disabled` /
//     `readonly` / `aria-*` / `tabindex` / `role` (negative tests in the
//     invalid block)
//   - Diagnostic position locking (single-line, multi-line, paired form,
//     siblings, fragments)
//
// The upstream-replica suite lives in `autocomplete_valid_upstream_test.go`.
// `componentsSettings` is shared from that file via the same package.
func TestAutocompleteValidExtras(t *testing.T) {
	validCases := []rule_tester.ValidTestCase{
		// ---- Upstream-walk lock-ins ----
		// stateTerms — the axe-core extras (none, false, true, disabled,
		// enabled, undefined, null, xoff, xon) must each be accepted.
		// Exercising every entry guards against accidental list edits.
		{Code: `<input autocomplete="none" />;`, Tsx: true},
		{Code: `<input autocomplete="false" />;`, Tsx: true},
		{Code: `<input autocomplete="true" />;`, Tsx: true},
		{Code: `<input autocomplete="disabled" />;`, Tsx: true},
		{Code: `<input autocomplete="enabled" />;`, Tsx: true},
		{Code: `<input autocomplete="undefined" />;`, Tsx: true},
		{Code: `<input autocomplete="null" />;`, Tsx: true},
		{Code: `<input autocomplete="xoff" />;`, Tsx: true},
		{Code: `<input autocomplete="xon" />;`, Tsx: true},
		// Case-insensitivity: the axe-core algorithm lowercases before any
		// term matching, so uppercase / mixed-case values must still pass.
		{Code: `<input autocomplete="NAME" />;`, Tsx: true},
		{Code: `<input autocomplete="OFF" />;`, Tsx: true},
		{Code: `<input autocomplete="Section-Blue Shipping Street-Address" />;`, Tsx: true},
		// Whitespace handling: trim before stateTerms check, then \s+ split.
		// Multiple spaces between tokens collapse via strings.Fields.
		{Code: `<input autocomplete="   name   " />;`, Tsx: true},
		{Code: `<input autocomplete="  " />;`, Tsx: true},
		{Code: `<input autocomplete="billing  family-name" />;`, Tsx: true},
		// webauthn alone is invalid (terms empty after pop) — tested below.
		// webauthn AFTER a valid token pops then re-validates the rest.
		{Code: `<input autocomplete="name webauthn" />;`, Tsx: true},
		{Code: `<input autocomplete="shipping street-address webauthn" />;`, Tsx: true},
		{Code: `<input autocomplete="home email webauthn" />;`, Tsx: true},
		// ignoredValues — axe-core returns `undefined` (incomplete), not a
		// violation, so the ESLint plugin reports nothing. Lock each entry.
		{Code: `<input autocomplete="text" />;`, Tsx: true},
		{Code: `<input autocomplete="pronouns" />;`, Tsx: true},
		{Code: `<input autocomplete="gender" />;`, Tsx: true},
		{Code: `<input autocomplete="message" />;`, Tsx: true},
		{Code: `<input autocomplete="content" />;`, Tsx: true},
		// Qualified term combinations: every qualifier × every qualifiedTerm
		// representative should pass. Sample a few.
		{Code: `<input autocomplete="home tel" />;`, Tsx: true},
		{Code: `<input autocomplete="work email" />;`, Tsx: true},
		{Code: `<input autocomplete="mobile tel-extension" />;`, Tsx: true},
		{Code: `<input autocomplete="fax impp" />;`, Tsx: true},
		{Code: `<input autocomplete="pager tel-country-code" />;`, Tsx: true},
		// All three optional prefixes used at once: section + location + qualifier + qualified term.
		{Code: `<input autocomplete="section-payment billing home tel" />;`, Tsx: true},
		// Boundary on the section- length check: upstream requires
		// `length > 8`, so the 8-character bare token "section-" is NOT
		// stripped — it's treated as the purpose token, which doesn't match
		// anything → the value is invalid. Conversely "section-x" (9 chars)
		// IS stripped. Lock both sides of the boundary; the invalid case is
		// in the invalid block.
		{Code: `<input autocomplete="section-x name" />;`, Tsx: true},
		// JSX shape variants — locks that the listener fires on both
		// JsxOpeningElement (paired) and JsxSelfClosingElement (no children).
		{Code: `<input autocomplete="name"></input>`, Tsx: true},
		// Self-closing without space before `/>`.
		{Code: `<input autocomplete="name"/>`, Tsx: true},
		// Spread props mixed with literal autocomplete — spread is opaque,
		// the named attr still resolves to the literal "name" → valid.
		{Code: `<input {...rest} autocomplete="name" />;`, Tsx: true},
		// Explicit JsxExpression wrapping a literal — not technically common
		// but legal. Both StringLiteral and NoSubstitutionTemplateLiteral
		// inside a JsxExpression must resolve via LiteralPropStringValue.
		{Code: `<input autocomplete={"name"} />;`, Tsx: true},
		{Code: "<input autocomplete={`name`} />;", Tsx: true},
		// TS wrappers around the autocomplete literal. literalPropValue
		// goes through SkipOuterExpressions which strips parens / `as` / `!`.
		{Code: `<input autocomplete={"name" as const} />;`, Tsx: true},
		{Code: `<input autocomplete={("name")} />;`, Tsx: true},
		// Settings: components map points at a non-input HTML tag. The
		// element resolves to "div", which isn't in inputTypes → rule
		// returns early. Locks the components-map path.
		{
			Code: `<Block autocomplete="foo" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"Block": "div"},
				},
			},
		},
		// polymorphicPropName resolves a polymorphic component to a
		// non-input element. Locks the polymorphic + non-input path.
		{
			Code: `<Box as="div" autocomplete="foo" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},
		// Non-input elements with autocomplete are silently passed through
		// per upstream's `inputTypes.includes(elType)` gate.
		{Code: `<select autocomplete="foo" />;`, Tsx: true},
		{Code: `<textarea autocomplete="foo" />;`, Tsx: true},
		{Code: `<button autocomplete="foo" />;`, Tsx: true},
		// Member-expression / namespaced tag names — neither matches "input"
		// directly, so no validation runs.
		{Code: `<Foo.input autocomplete="foo" />;`, Tsx: true},
		{Code: `<svg:input autocomplete="foo" />;`, Tsx: true},
		// inputComponents option as the array-wrapped JSON shape (matches
		// the rule_tester multi-element shape vs. CLI bare-object shape).
		{
			Code:    `<MyInput autocomplete="name" />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"inputComponents": []interface{}{"MyInput"}}},
		},
		// Empty inputComponents array — no extra components, only "input"
		// is checked. <Foo> remains unchecked.
		{
			Code:    `<Foo autocomplete="foo" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"inputComponents": []interface{}{}},
		},
		// Nil options — defaults to just ["input"].
		{Code: `<Foo autocomplete="foo" />;`, Tsx: true},
		// Malformed option (non-array inputComponents) — tolerated by
		// GetOptionsMap path, falls through to defaults.
		{
			Code:    `<Foo autocomplete="foo" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"inputComponents": "MyInput"},
		},
		// `<input autocomplete={undefined} />` — getLiteralPropValue returns
		// undefined; typeof !== string → return early.
		{Code: `<input autocomplete={undefined} />;`, Tsx: true},
		// Literal `null` autocomplete — LITERAL_TYPES.Literal special-cases
		// null to the string "null"; "null" is in the extended stateTerms
		// → valid. Locks the null-literal coercion path.
		{Code: `<input autocomplete={null} />;`, Tsx: true},
		// CallExpression inside a JsxExpression — noop in LITERAL_TYPES →
		// rule returns early.
		{Code: `<input autocomplete={getValue()} />;`, Tsx: true},
		// MemberExpression inside a JsxExpression — noop in LITERAL_TYPES.
		{Code: `<input autocomplete={config.autocomplete} />;`, Tsx: true},
		// Conditional expression — noop in LITERAL_TYPES.
		{Code: `<input autocomplete={cond ? "name" : "foo"} />;`, Tsx: true},
		// `<input>` inside common React patterns — listener must fire but
		// the value is valid. Locks the listener-shape coverage.
		{Code: `function F() { return <input autocomplete="name" />; }`, Tsx: true},
		{Code: `class C { render() { return <input autocomplete="name" />; } }`, Tsx: true},
		// Multi-line JSX with a valid literal autocomplete — must not fire.
		{Code: "<input\n  type=\"text\"\n  autocomplete=\"name\"\n/>", Tsx: true},

		// ---- axe-core matches() type-filter lockdowns ----
		// Each of the four excluded types skips the check entirely — even
		// with an obviously invalid autocomplete value. Locks parity with
		// `lib/commons/forms/autocomplete-matches.js` excludedInputTypes.
		{Code: `<input type="hidden" autocomplete="foo" />;`, Tsx: true},
		{Code: `<input type="submit" autocomplete="foo" />;`, Tsx: true},
		{Code: `<input type="reset" autocomplete="invalid garbage" />;`, Tsx: true},
		{Code: `<input type="button" autocomplete="bogus" />;`, Tsx: true},
		// Case-insensitivity on the type comparison: axe-core's
		// SerialVirtualNode lowercases `attributes.type`. Locks the lower-
		// case normalization on rslint's side.
		{Code: `<input type="HIDDEN" autocomplete="foo" />;`, Tsx: true},
		{Code: `<input type="Submit" autocomplete="foo" />;`, Tsx: true},
		// JsxExpression-wrapped literal type — should still resolve via
		// LITERAL_TYPES.Literal extraction.
		{Code: `<input type={"hidden"} autocomplete="foo" />;`, Tsx: true},
		{Code: "<input type={`hidden`} autocomplete=\"foo\" />;", Tsx: true},
		// TS wrappers around the literal type: stripped before comparison.
		{Code: `<input type={"hidden" as const} autocomplete="foo" />;`, Tsx: true},
		// Custom inputComponent with excluded type — the ESLint plugin
		// hardcodes `nodeName: 'input'` in runVirtualRule so the type
		// filter applies to inputComponents too.
		{
			Code:    `<MyInput type="hidden" autocomplete="foo" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"inputComponents": []interface{}{"MyInput"}},
		},
		// Components-map mapping with excluded type — same path.
		{
			Code:     `<Input type="submit" autocomplete="baz" />`,
			Tsx:      true,
			Settings: componentsSettings,
		},

		// ---- Spread-of-literal lockdowns ----
		// `<input {...{autocomplete: "name"}} />` — FindAttributeByName
		// inspects literal-spread arguments per jsx-ast-utils' getProp
		// behavior. Locks that the literal-spread path produces the same
		// validation result as a direct named attribute.
		{Code: `<input {...{autocomplete: "name"}} />;`, Tsx: true},
		// Shorthand inside a literal spread — `{...{autocomplete}}` — the
		// shorthand value is the bound identifier, not a literal string,
		// so LiteralPropStringValue returns ("", false) → rule returns
		// early → valid (matches upstream where Identifier is noop in
		// LITERAL_TYPES).
		{Code: `<input {...{autocomplete}} />;`, Tsx: true},

		// ---- Logical short-circuit lockdowns ----
		// `&&` short-circuit — same noop in LITERAL_TYPES as `||`, so the
		// rule must return early without inspecting either side.
		{Code: `<input autocomplete={cond && "name"} />;`, Tsx: true},
		// Nullish coalescing — also noop in LITERAL_TYPES.
		{Code: `<input autocomplete={x ?? "name"} />;`, Tsx: true},

		// ---- Deeply-nested JSX patterns ----
		// JSX inside Array.map callback — common React pattern. Listener
		// must fire on the inner <input>; outer expression doesn't matter.
		{Code: `function L({xs}) { return xs.map(x => <input autocomplete="name" key={x} />); }`, Tsx: true},
		// Conditional render via && inside parent JSX — inner <input> fires.
		{Code: `function F() { return <div>{cond && <input autocomplete="name" />}</div>; }`, Tsx: true},
		// JSX inside generic function — locks that surrounding TS
		// constructs don't affect listener firing.
		{Code: `function f<T>(x: T) { return <input autocomplete="name" />; }`, Tsx: true},
		// Nested forms / fragments — each <input> fires independently.
		{Code: `<form><input autocomplete="name" /></form>`, Tsx: true},
		{Code: `<><><input autocomplete="name" /></></>`, Tsx: true},
		// JSX inside hook callback.
		{Code: `function F() { useEffect(() => { renderer(<input autocomplete="name" />); }); }`, Tsx: true},
		// Tab / newline as whitespace separator: strings.Fields collapses
		// any whitespace run, mirroring axe-core's `\s+` split.
		{Code: "<input autocomplete=\"home\ttel\" />", Tsx: true},
		// String concatenation here splits the literal "\nwebauthn" so the
		// spell-checker doesn't see one fused token; the actual JSX value
		// still contains a newline followed by the webauthn keyword.
		{Code: "<input autocomplete=\"name\n" + "webauthn\" />", Tsx: true},
		{Code: "<input autocomplete=\"shipping\tstreet-address\" />", Tsx: true},

		// ---- Attribute-name case-insensitive matching ----
		// jsx-ast-utils' getProp uses { ignoreCase: true } by default;
		// FindAttributeByName mirrors via strings.EqualFold. Real JSX has
		// authors using both `autocomplete` (HTML-correct) and
		// `autoComplete` (React's camelCase prop). Both must trigger the
		// rule. Locks the case-folded match path on both literal and
		// JsxExpression initializer forms.
		{Code: `<input autoComplete="name" />;`, Tsx: true},
		{Code: `<input AUTOCOMPLETE="name" />;`, Tsx: true},
		{Code: `<input AutoComplete="name" />;`, Tsx: true},
		{Code: `<input autoComplete={dynamicValue} />;`, Tsx: true},

		// ---- Real-world login form patterns ----
		// All standalone-term entries we ship — covers the sub-list
		// we copied from axe-core. Explicitly exercising each guards
		// against accidental list edits.
		{Code: `<input type="text" autocomplete="username" />;`, Tsx: true},
		{Code: `<input type="password" autocomplete="current-password" />;`, Tsx: true},
		{Code: `<input type="password" autocomplete="new-password" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="one-time-code" />;`, Tsx: true},
		// Address form — the "address-line*"/"address-level*" family.
		{Code: `<input type="text" autocomplete="street-address" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="address-line1" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="address-line2" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="address-line3" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="address-level1" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="address-level2" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="address-level3" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="address-level4" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="country" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="country-name" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="postal-code" />;`, Tsx: true},
		// Personal — name family + birthday family.
		{Code: `<input type="text" autocomplete="honorific-prefix" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="given-name" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="additional-name" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="family-name" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="honorific-suffix" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="nickname" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="organization-title" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="organization" />;`, Tsx: true},
		{Code: `<input type="date" autocomplete="bday" />;`, Tsx: true},
		{Code: `<input type="number" autocomplete="bday-day" />;`, Tsx: true},
		{Code: `<input type="number" autocomplete="bday-month" />;`, Tsx: true},
		{Code: `<input type="number" autocomplete="bday-year" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="sex" />;`, Tsx: true},
		// Credit card form.
		{Code: `<input type="text" autocomplete="cc-name" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="cc-given-name" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="cc-additional-name" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="cc-family-name" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="cc-number" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="cc-exp" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="cc-exp-month" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="cc-exp-year" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="cc-csc" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="cc-type" />;`, Tsx: true},
		// Misc standalone.
		{Code: `<input type="text" autocomplete="transaction-currency" />;`, Tsx: true},
		{Code: `<input type="number" autocomplete="transaction-amount" />;`, Tsx: true},
		{Code: `<input type="text" autocomplete="language" />;`, Tsx: true},
		{Code: `<input type="url" autocomplete="url" />;`, Tsx: true},
		{Code: `<input type="url" autocomplete="photo" />;`, Tsx: true},

		// ---- Real-world component patterns ----
		// Common login form composition.
		{
			Code: `function LoginForm() {
				return (
					<form>
						<input type="text" autocomplete="username" />
						<input type="password" autocomplete="current-password" />
						<button type="submit">Login</button>
					</form>
				);
			}`,
			Tsx: true,
		},
		// Map over field config — common pattern for dynamic forms.
		{
			Code: `function DynamicForm({ fields }) {
				return fields.map(f => <input key={f.id} autocomplete={f.autocomplete} />);
			}`,
			Tsx: true,
		},
		// Forwarded ref + custom Input mapped to input.
		{
			Code: `const Input = forwardRef((props, ref) => <input {...props} ref={ref} autocomplete="name" />);`,
			Tsx:  true,
		},
		// Nested form with fieldset.
		{
			Code: `<form><fieldset><legend>Address</legend><input autocomplete="street-address" /></fieldset></form>`,
			Tsx:  true,
		},
		// JSX as array element.
		{
			Code: `const inputs = [<input autocomplete="name" key="1" />, <input autocomplete="email" key="2" />];`,
			Tsx:  true,
		},
		// JSX in object literal value.
		{
			Code: `const map = { name: <input autocomplete="name" />, email: <input autocomplete="email" /> };`,
			Tsx:  true,
		},
		// JSX in default function parameter.
		{
			Code: `function F(child = <input autocomplete="name" />) { return child; }`,
			Tsx:  true,
		},

		// ---- Multiple spread patterns ----
		// Spread before named, then named overrides — first match wins
		// per FindAttributeByName/getProp iteration order. Lock the
		// "first wins" behavior.
		{Code: `<input {...rest} autocomplete="name" />;`, Tsx: true},
		// Named first, spread second — finds the named.
		{Code: `<input autocomplete="name" {...rest} />;`, Tsx: true},
		// Multiple spreads.
		{Code: `<input {...a} {...b} autocomplete="name" />;`, Tsx: true},
		// Spread of literal containing key + autocomplete.
		{Code: `<input {...{key: "x", autocomplete: "name"}} />;`, Tsx: true},
		// Duplicate autocomplete attributes — first wins.
		{Code: `<input autocomplete="name" autocomplete="foo" />;`, Tsx: true},

		// ---- Optional chaining and complex expression shapes ----
		// Optional chain — tsgo flag-based; same kind as PropertyAccess.
		// In LITERAL_TYPES this is noop → null → not a string → early
		// return. Locks that optional chains don't accidentally produce
		// a literal string.
		{Code: `<input autocomplete={config?.autocomplete} />;`, Tsx: true},
		{Code: `<input autocomplete={fn?.()} />;`, Tsx: true},
		{Code: `<input autocomplete={a?.b?.c?.d} />;`, Tsx: true},
		// Tagged template — TaggedTemplateExpression noop in
		// LITERAL_TYPES; literalPropValue's default returns jsNull.
		{Code: "<input autocomplete={tag`name`} />;", Tsx: true},
		// New expression — null in LITERAL_TYPES.
		{Code: `<input autocomplete={new String("name")} />;`, Tsx: true},
		// `this` reference — ThisExpression noop in LITERAL_TYPES.
		{Code: `<input autocomplete={this.value} />;`, Tsx: true},
		// Class expression as autocomplete value — defensive.
		{Code: `<input autocomplete={class {}} />;`, Tsx: true},
		// Array literal as value.
		{Code: `<input autocomplete={["name"]} />;`, Tsx: true},
		// Object literal as value.
		{Code: `<input autocomplete={{}} />;`, Tsx: true},
		// Spread element via TS satisfies operator.
		{Code: `<input autocomplete={"name" satisfies string} />;`, Tsx: true},
		// TS non-null assertion.
		{Code: `<input autocomplete={value!} />;`, Tsx: true},
		// Multiple TS wrappers stacked — verifies all OEKAssertions
		// bits (TypeAssertions, NonNullAssertions, Satisfies) and parens
		// are stripped before extraction.
		{Code: `<input autocomplete={(("name" as string) satisfies any)} />;`, Tsx: true},
		// JsxExpression with surrounding comments — the parser strips
		// leading/trailing comments inside the container, leaving just
		// the inner literal.
		{Code: `<input autocomplete={/* leading */ "name" /* trailing */} />;`, Tsx: true},
	}

	invalidCases := []rule_tester.InvalidTestCase{
		// ---- Grammar-boundary lock-ins ----
		// webauthn alone is invalid — pops to empty terms list.
		{
			Code: `<input autocomplete="webauthn" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// section- alone (after stripping the prefix) leaves no purpose token.
		{
			Code: `<input autocomplete="section-foo" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Bare "section-" (8 chars) is NOT stripped (length > 8 boundary),
		// so it becomes the purpose token, which matches nothing → invalid.
		// Locks the strict `> 8` upstream comparison.
		{
			Code: `<input autocomplete="section-" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Two consecutive locations — only one is consumed, leaving an
		// unrecognized leftover.
		{
			Code: `<input autocomplete="billing shipping name" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Two consecutive qualifiers — second one isn't recognized and the
		// terms.length !== 1 check fails.
		{
			Code: `<input autocomplete="home work tel" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// qualifier + standaloneTerm: after a qualifier upstream resets
		// standaloneTerms = [], leaving only qualifiedTerms acceptable.
		// "name" is a standaloneTerm, not qualified → invalid.
		{
			Code: `<input autocomplete="home name" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// qualifier alone — terms.length !== 1 (it's 0 after the qualifier
		// shift) → invalid.
		{
			Code: `<input autocomplete="home" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// location alone — terms.length !== 1 (0 after shift) → invalid.
		{
			Code: `<input autocomplete="billing" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// section + location + qualifier without a final field-name token
		// — leaves the qualifier shifted off, terms empty → invalid.
		{
			Code: `<input autocomplete="section-foo shipping work" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// `<Input autocomplete="baz" />` with NO type attribute, with the
		// components-map mapping `Input` → `input`. Locks that the type
		// attribute is irrelevant to autocomplete-valid (axe-core's
		// autocomplete-valid-evaluate ignores it; only the separate
		// autocomplete-appropriate rule consults it).
		{
			Code:     `<Input autocomplete="baz" />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// polymorphicPropName resolves a polymorphic component TO `input`
		// — the rule must run validation as if it were a bare <input>.
		{
			Code: `<Box as="input" autocomplete="foo" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Diagnostic position locking: upstream reports on the
		// JSXOpeningElement node, which spans `<` through `>` of the opening
		// tag. tsgo's JsxSelfClosingElement spans the whole self-closing tag
		// including the trailing `/>`. Lock the position so a listener bleed
		// would surface as a position drift.
		{
			Code: `<input autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
				Line:      1, Column: 1, EndLine: 1, EndColumn: 29,
			}},
		},
		// Multi-line invalid attribute — report span extends across lines.
		{
			Code: "<input\n  autocomplete=\"foo\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
				Line:      1, Column: 1, EndLine: 3, EndColumn: 3,
			}},
		},
		// Paired (non-self-closing) form — the JsxOpeningElement listener
		// fires; span ends at `>` of the opening tag.
		{
			Code: `<input autocomplete="foo"></input>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
				Line:      1, Column: 1, EndLine: 1, EndColumn: 27,
			}},
		},
		// Multiple invalid <input> elements at the same level — each one
		// fires independently. Locks that the listener doesn't bleed.
		{
			Code: `<><input autocomplete="foo" /><input autocomplete="bar" /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "autocompleteValid", Message: failMessage},
				{MessageId: "autocompleteValid", Message: failMessage},
			},
		},
		// Mixed valid + invalid — only the invalid one reports.
		{
			Code: `<><input autocomplete="name" /><input autocomplete="foo" /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Uppercase invalid value — algorithm lowercases first, but the
		// resulting "foo" still doesn't match anything.
		{
			Code: `<input autocomplete="FOO" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Whitespace around an invalid value: trim then split, "foo" → invalid.
		{
			Code: `<input autocomplete="  foo  " />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Spread-of-literal with invalid autocomplete — must report.
		{
			Code: `<input {...{autocomplete: "foo"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Type filter does NOT skip non-excluded types (so `type="text"`
		// with invalid autocomplete still reports — already covered by the
		// upstream cases, but explicitly anchor a non-excluded boundary).
		{
			Code: `<input type="email" autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Dynamic type (ConditionalExpression noop in LITERAL_TYPES) →
		// type passed to axe-core as undefined → matches doesn't filter →
		// check runs. Locks that we don't accidentally skip via a non-
		// literal type.
		{
			Code: `<input type={cond ? "hidden" : "text"} autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Boolean type form (`<input type />`) — getLiteralPropValue
		// returns boolean true, not a string → matches() doesn't filter.
		{
			Code: `<input type autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// TemplateExpression with substitutions — literalPropValue
		// synthesizes a placeholder string ("name${Expression}" or similar)
		// which doesn't match any token grammar → invalid. Locks the
		// template-with-substitution path; upstream's TemplateLiteral
		// extractor produces a similar non-token string.
		{
			Code: "<input autocomplete={`name${suffix}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// CRITICAL alignment lock: `<input autocomplete={`${undefined}`} />`.
		// Pre-fix the synthesized text was bare "undefined" which is in
		// axe-core's extended stateTerms — would have falsely passed.
		// Post-fix the synthesized text is "${undefined}" which fails the
		// token grammar → invalid → reported. Same as upstream.
		{
			Code: "<input autocomplete={`${undefined}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Template substitution with a non-undefined identifier —
		// synthesizes "${name}" → not a token → invalid. Mirrors upstream.
		{
			Code: "<input autocomplete={`${dynVar}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Template with surrounding text + substitution.
		{
			Code: "<input autocomplete={`name ${suffix}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},

		// ---- Attribute-name case-insensitive match (invalid path) ----
		// Mirrors the case-folded match: any case form must trigger when
		// the value is invalid.
		{
			Code: `<input autoComplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code: `<input AUTOCOMPLETE="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},

		// ---- Duplicate attribute (first wins) ----
		// `<input autocomplete="foo" autocomplete="name" />` — the FIRST
		// match wins per getProp/FindAttributeByName iteration order, so
		// the value seen by the rule is "foo" → INVALID. Locks the
		// "first wins" semantics in the invalid direction.
		{
			Code: `<input autocomplete="foo" autocomplete="name" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Spread-of-literal duplicate: literal-spread with autocomplete:
		// "foo" comes BEFORE the named autocomplete="name". Spread is
		// scanned in order — its prop is found first → "foo" wins → INVALID.
		{
			Code: `<input {...{autocomplete: "foo"}} autocomplete="name" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},

		// ---- ESLint plugin doesn't pass disabled/readonly/aria-* to
		// runVirtualRule — these MUST NOT cause skipping. Locks that
		// our impl doesn't accidentally implement matches() gates that
		// upstream can't reach. ----
		{
			Code: `<input disabled autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code: `<input readOnly autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code: `<input aria-disabled="true" autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code: `<input aria-readonly="true" autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code: `<input tabIndex={-1} autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		{
			Code: `<input role="presentation" autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},

		// ---- Type filter with spread ----
		// type via spread-of-literal: `{...{type: "hidden"}}` resolves
		// the type to "hidden" → matches() filter applies → no report.
		// (In invalid block to lock the FALSE direction: when type is
		// NOT excluded via spread, validation runs.)
		{
			Code: `<input {...{type: "text"}} autocomplete="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},

		// ---- Multi-element invalid lock ----
		// Three sibling inputs with progressively complex bad values.
		// Locks per-element fire and reporting order.
		{
			Code: `<><input autocomplete="foo" /><input autocomplete="name invalid" /><input autocomplete="home url" /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "autocompleteValid", Message: failMessage},
				{MessageId: "autocompleteValid", Message: failMessage},
				{MessageId: "autocompleteValid", Message: failMessage},
			},
		},

		// ---- Edge grammar boundaries ----
		// Two webauthn tokens: pop one → ["webauthn"] → not stateTerm, single
		// term, "webauthn" not in any list → INVALID.
		{
			Code: `<input autocomplete="webauthn webauthn" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// section-* prefix appearing in second position — only the FIRST
		// term gets the section- treatment. Second one is left as a
		// regular token, fails grammar.
		{
			Code: `<input autocomplete="name section-foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Two sections: only the first is shifted, second stays.
		{
			Code: `<input autocomplete="section-a section-b name" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Trailing webauthn with too many tokens leftover.
		{
			Code: `<input autocomplete="name extra webauthn" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// JsxExpression with comments around an invalid literal — comments
		// are stripped, "foo" is extracted, fails grammar.
		{
			Code: `<input autocomplete={/* leading */ "foo" /* trailing */} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// `satisfies` in attribute initializer — OEKAssertions includes
		// the satisfies bit (probed empirically: `OEKAssertions = 38 = 32
		// + 4 + 2`, where bit 5 is satisfies). Locks that satisfies is
		// transparent in our extractor.
		{
			Code: `<input autocomplete={"foo" satisfies string} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
		// Spread-of-literal with TS-wrapped value — wrapper stripped at
		// the value, "foo" extracted via PropertyAssignment path.
		{
			Code: `<input {...{autocomplete: "foo" as const}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "autocompleteValid",
				Message:   failMessage,
			}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AutocompleteValidRule, validCases, invalidCases)
}
