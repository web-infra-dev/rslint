package anchor_is_valid

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicSettings exercises the `polymorphicPropName` jsx-a11y setting —
// `<Foo as="a">` should be treated as `<a>` after getElementType resolution.
// Locks in jsxa11yutil.GetElementType's polymorphic-prop branch which
// upstream's own test file doesn't directly exercise on this rule.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// TestAnchorIsValidExtras locks in branches that upstream's test file
// doesn't directly exercise but are reachable through the rule's listener
// gate. Each case carries an inline comment pointing at the specific
// upstream branch / Dimension 4 edge shape it covers.
func TestAnchorIsValidExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AnchorIsValidRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: paired form (KindJsxOpeningElement listener) ----
		// Upstream's `JSXOpeningElement` fires for both self-closing AND
		// paired tags. tsgo splits these into two kinds; we listen on
		// both. `<a href="foo">x</a>` exercises the paired form.
		{Code: `<a href="foo">x</a>`, Tsx: true},

		// ---- Dimension 4: tag-name forms — case-sensitive HTML matching ----
		// `typeCheck.indexOf(nodeType)` is case-sensitive. `<A />` resolves
		// to nodeType "A" which is NOT in ['a'] → rule short-circuits and
		// does NOT report, even with no href.
		{Code: `<A />`, Tsx: true},

		// ---- Dimension 4: namespaced JSX names → not matched ----
		{Code: `<svg:a />`, Tsx: true},

		// ---- Dimension 4: member-access tag → not matched ----
		// Even when the member's last segment is "a", the resolved type is
		// "Foo.a" — distinct from "a". Skipped.
		{Code: `<Foo.a />`, Tsx: true},

		// ---- Dimension 4: case-insensitive prop matching (jsx-ast-utils
		//      `getProp` default `ignoreCase: true`) ----
		// `HREF`, `Href`, etc. all match. Locks in FindAttributeByName's
		// strings.EqualFold lookup.
		{Code: `<a HREF="foo" />`, Tsx: true},
		{Code: `<a Href="foo" />`, Tsx: true},

		// ---- Dimension 4: boolean-attribute form — value is boolean true,
		//      not null/undefined. hasAnyHref=true; not a string so not in
		//      invalid list. Result: passes. ----
		{Code: `<a href />`, Tsx: true},

		// ---- Dimension 4: TS expression wrappers on the value (parens +
		//      `as` + `!`). staticEval strips these via skipTransparent
		//      (parens + type assertions + non-null), so the inner string
		//      literal "foo" is what staticEval sees. NOTE: `satisfies` is
		//      intentionally NOT in skipTransparent — see the satisfies
		//      lock-in in the invalid section below. ----
		{Code: `<a href={("foo")} />`, Tsx: true},
		{Code: `<a href={"foo" as string} />`, Tsx: true},
		{Code: `<a href={"foo" as const} />`, Tsx: true},
		{Code: `<a href={"foo"!} />`, Tsx: true},

		// ---- Dimension 4: spread attribute — non-literal spread is opaque,
		//      hasSpread=true suppresses the noHref/preferButton branches. ----
		{Code: `<a {...props} />`, Tsx: true},
		{Code: `<a {...props} onClick={() => 0} />`, Tsx: true},

		// ---- Dimension 4: spread of literal object that supplies a valid
		//      href. FindAttributeByName walks the literal-spread to find
		//      the property; staticEval extracts "foo" from the
		//      PropertyAssignment initializer. hasAnyHref=true,
		//      hasInvalidHref=false → no report. ----
		{Code: `<a {...{href: "foo"}} />`, Tsx: true},
		{Code: `<a {...{href: "/foo"}} />`, Tsx: true},

		// ---- Locks in upstream regex /^\W*?javascript:/ negative cases ----
		// Word-char prefix prevents the lazy `\W*?` from consuming, so the
		// pattern fails to anchor at "javascript:". These must remain valid.
		{Code: `<a href="xjavascript:" />`, Tsx: true},
		// "javascript" without the trailing colon never matches.
		{Code: `<a href="javascript" />`, Tsx: true},
		{Code: `<a href="javascript;" />`, Tsx: true},
		// "##" is not the exact "#" string upstream rejects.
		{Code: `<a href="##" />`, Tsx: true},
		// Non-empty whitespace-only strings are NOT rejected — they pass
		// every disjunct of the invalid filter (length>0, !== "#", no
		// "javascript:" anywhere). Locks in the literal value comparison.
		{Code: `<a href=" " />`, Tsx: true},

		// ---- Locks in upstream Identifier extractor → string of name.
		//      `href={foo}` ⇒ jsx-ast-utils returns "foo" (truthy non-empty
		//      non-"#" non-"javascript:" string) ⇒ valid. ----
		{Code: `<a href={someVar} />`, Tsx: true},

		// ---- Locks in upstream MemberExpression / CallExpression / etc.
		//      → synthesized non-empty string (or jsTruthy in our impl).
		//      typeof-string check upstream still passes for these because
		//      jsx-ast-utils returns "obj.prop" / "callee()" — but the
		//      synthesized strings don't match the invalid disjuncts. Our
		//      staticEval returns jsTruthy (Kind != jvString) so the typeof
		//      check fails. Same end result: not invalid. ----
		{Code: `<a href={obj.prop} />`, Tsx: true},
		{Code: `<a href={getUrl()} />`, Tsx: true},

		// ---- Polymorphic settings: `<Foo as="a">` resolves to "a" ----
		// FindAttributeByName + GetElementType's polymorphic step.
		{Code: `<Foo as="a" href="x" />`, Tsx: true, Settings: polymorphicSettings},

		// ---- Numeric / boolean values are non-string, so they bypass the
		//      typeof-string invalid filter. hasAnyHref=true. ----
		{Code: `<a href={5} />`, Tsx: true},
		{Code: `<a href={true} />`, Tsx: true},
		{Code: `<a href={false} />`, Tsx: true},

		// ---- One valid + one missing href-like prop → hasAnyHref=true via
		//      the valid one, hasInvalidHref=false. Locks in the OR-merge
		//      across propsToValidate. ----
		{Code: `<a href="foo" hrefLeft={null} />`, Tsx: true, Options: specialLinkOption},

		// ---- Empty aspects array: upstream's runtime treats this as "no
		//      aspects active", so even no-href elements pass. Locks in
		//      our explicit-aspects-array deactivation logic. ----
		{Code: `<a />`, Tsx: true, Options: []interface{}{
			map[string]interface{}{"aspects": []interface{}{}},
		}},

		// ---- staticEval `||` short-circuit: falsy left → take right ----
		// Upstream's jsx-ast-utils LogicalExpression extractor evaluates
		// `getValue(left) || getValue(right)`. Empty string is falsy → use
		// the truthy right value. Our staticEval mirrors via
		// `if !jsTruthy_(left) return staticEval(right)`.
		{Code: `<a href={"" || "/foo"} />`, Tsx: true},
		{Code: `<a href={null || "/foo"} />`, Tsx: true},
		{Code: `<a href={undefined || "/foo"} />`, Tsx: true},

		// ---- staticEval `||` short-circuit: truthy left → take left ----
		// `someVar || "/fallback"` — Identifier extractor returns "someVar"
		// (truthy) → returns "someVar". Not invalid. Realistic pattern.
		{Code: `<a href={someVar || "/fallback"} />`, Tsx: true},

		// ---- staticEval `&&` short-circuit with non-nullish falsy left ----
		// `false && "#"` — left is jvBool false (falsy but NOT null/undef)
		// → returns false. PropValueIsNullish=false → hasAnyHref=true.
		// PropStaticStringValue: jvBool != jvString → not invalid. Result:
		// passes. Lock in the boolean-false / number-zero distinction from
		// the null/undef cases below (which produce noHref reports).
		{Code: `<a href={false && "#"} />`, Tsx: true},
		// `0 && "#"` — number 0 (falsy, not null/undef) → returns 0. Same
		// classification as boolean false: contributes to hasAnyHref, not
		// a string, not invalid.
		{Code: `<a href={0 && "#"} />`, Tsx: true},

		// ---- staticEval `??` nullish coalescing: non-null left → take left ----
		// `"/" ?? "#"` — left is jvString "/" (not null/undef) → take left.
		// Valid.
		{Code: `<a href={"/" ?? "#"} />`, Tsx: true},
		// `someVar ?? "#"` — Identifier extractor returns string "someVar"
		// (non-null) → take left. Valid.
		{Code: `<a href={someVar ?? "#"} />`, Tsx: true},

		// ---- staticEval ConditionalExpression: take consequent or alternate
		//      based on test ----
		// `true ? "foo" : "#"` → "foo". `cond ? "foo" : "#"` where cond is
		// Identifier "cond" → truthy string → "foo".
		{Code: `<a href={true ? "/foo" : "#"} />`, Tsx: true},
		{Code: `<a href={cond ? "/foo" : "#"} />`, Tsx: true},

		// ---- staticEval string concatenation `+` ----
		// `"prefix-" + suffix` → "prefix-{suffix-name-string}" — synthesized
		// string, not invalid (no js: prefix, not "#" exactly, length>0).
		{Code: `<a href={"https://" + domain} />`, Tsx: true},
		{Code: `<a href={base + "/path"} />`, Tsx: true},

		// ---- staticEval TemplateExpression with variable substitutions ----
		// `` `/users/${id}` `` — staticEval emits "/users/{id}" placeholder.
		// Length > 0, not invalid. Realistic pattern.
		{Code: "<a href={`/users/${id}`} />", Tsx: true},
		{Code: "<a href={`${prefix}/${path}`} />", Tsx: true},

		// ---- Numeric / regex / new-expression / object literal as href
		//      values. None are strings → bypass typeof-string filter →
		//      contribute to hasAnyHref but never invalid. ----
		{Code: `<a href={5} />`, Tsx: true},
		{Code: `<a href={0} />`, Tsx: true},
		{Code: `<a href={5n} />`, Tsx: true},
		{Code: `<a href={new URL("/")} />`, Tsx: true},
		{Code: `<a href={["/foo"]} />`, Tsx: true},
		{Code: `<a href={{toString: () => "/"}} />`, Tsx: true},

		// ---- Optional chain on href value (`obj?.url`). tsgo flags this as
		//      a PropertyAccessExpression with the optional flag — same kind
		//      as plain `obj.prop`. staticEval returns jsTruthy → contributes
		//      to hasAnyHref, not invalid. ----
		{Code: `<a href={obj?.url} />`, Tsx: true},
		{Code: `<a href={fns.getUrl?.()} />`, Tsx: true},

		// ---- onClick prop case-insensitivity. jsx-ast-utils' getProp default
		//      `ignoreCase: true` matches the standard `onClick` against the
		//      PascalCase variant and the legacy lowercase HTML attribute. ----
		{Code: `<a href="/" OnClick={() => 0} />`, Tsx: true},
		{Code: `<a href="/" onclick={() => 0} />`, Tsx: true},

		// ---- onClick={undefined} still counts as "has onClick". The presence
		//      of the prop matters, not its value. With href present this
		//      passes (not invalid + hasAnyHref). ----
		{Code: `<a href="/" onClick={undefined} />`, Tsx: true},
		{Code: `<a href="/" onClick={null} />`, Tsx: true},

		// ---- Spread + onClick only: `<a {...{onClick: fn}}/>` — hasSpread
		//      from JsxSpreadAttribute, onClick found via literal-spread
		//      walk. !hasSpread is false → both noHref and preferButton
		//      gates suppressed in the no-href branch. Result: passes. ----
		{Code: `<a {...{onClick: () => 0}} />`, Tsx: true},

		// ---- Components option is case-sensitive — `["link"]` does NOT
		//      match JSX `<Link/>`. Locks in upstream's
		//      `typeCheck.indexOf(nodeType) === -1` exact-match behavior. ----
		{Code: `<Link />`, Tsx: true, Options: []interface{}{
			map[string]interface{}{"components": []interface{}{"link"}},
		}},

		// ---- specialLink with prop names not present on element — fall back
		//      to the default 'href' lookup. Locks in propsToValidate
		//      iteration when a configured prop is absent. ----
		{Code: `<a href="/foo" />`, Tsx: true, Options: []interface{}{
			map[string]interface{}{"specialLink": []interface{}{"hrefBottom"}},
		}},

		// ---- aspects array with an unknown name — silently ignored,
		//      matching upstream's `aspects.indexOf(aspect) !== -1`
		//      mechanism. All three known aspects stay false → no report. ----
		{Code: `<a />`, Tsx: true, Options: []interface{}{
			map[string]interface{}{"aspects": []interface{}{"madeUpAspect"}},
		}},

		// ---- aspects with all three names = same as default (all active) ----
		// Locks in the explicit-vs-default symmetry: same outcome.
		{Code: `<a href="/foo" />`, Tsx: true, Options: []interface{}{
			map[string]interface{}{"aspects": []interface{}{"noHref", "invalidHref", "preferButton"}},
		}},

		// ---- Real-world Nav-component flow: multiple anchors of varying
		//      states. Per-element listener firing must not bleed; valid
		//      anchors stay valid. ----
		{
			Code: `<nav>
  <a href="/">Home</a>
  <a href="/about">About</a>
</nav>`,
			Tsx: true,
		},

		// ---- Anchor inside Array.map callback — each rendered element gets
		//      its own listener invocation, with no shared state. ----
		{Code: `<ul>{items.map(it => <a key={it.id} href={it.url}>{it.label}</a>)}</ul>`, Tsx: true},

		// ---- Shorthand property in literal-spread: `{...{href}}` ≡
		//      `{...{href: href}}`. FindAttributeByName resolves the
		//      ShorthandPropertyAssignment via AttributeInitializer →
		//      bound Identifier "href" → staticEval returns jvString
		//      "href" (truthy non-empty) → not invalid. ----
		{Code: `<a {...{href}} />`, Tsx: true},

		// ---- Spread literal carries nullish href: `{...{href: null}}` ----
		// FindAttributeByName walks the spread, finds `href: null`,
		// PropValueIsNullish returns true (jsNull). hasAnyHref=false.
		// hasSpread=true (because the JsxSpreadAttribute is present). The
		// !hasSpread gate suppresses both noHref and preferButton in the
		// no-href branch → no report. Same shape as `<a href={null} {...x}/>`
		// upstream — the spread is opaque-enough to suppress.
		{Code: `<a {...{href: null}} />`, Tsx: true},
		{Code: `<a {...{href: undefined}} />`, Tsx: true},

		// ---- Wrapper-component pattern: explicit href={undefined} +
		//      spread that supplies a valid href. At runtime the spread
		//      overrides; getProp / FindAttributeByName return the FIRST
		//      match (the explicit one with undefined). hasAnyHref=false
		//      via the explicit, hasSpread=true → suppressed → no report.
		//      Locks in upstream's first-match quirk for a realistic
		//      higher-order-component wrapper shape. ----
		{Code: `<a href={undefined} {...{href: "/"}} />`, Tsx: true},

		// ---- Two spread attributes (non-literal + literal) — common in
		//      composition: `<a {...rest} {...defaults}/>`. hasSpread=true,
		//      no explicit href. Suppressed regardless of literal payload. ----
		{Code: `<a {...rest} {...{className: "x"}} />`, Tsx: true},

		// ---- Options as bare-object (single-option CLI shape) — exercises
		//      utils.GetOptionsMap's non-array branch. Without this case
		//      the suite would only cover the array-wrapped rule_tester
		//      shape, leaving the CLI-facing wiring untested. ----
		{Code: `<Anchor href="/foo" />`, Tsx: true, Options: map[string]interface{}{
			"components": []interface{}{"Anchor"},
		}},
	}, []rule_tester.InvalidTestCase{
		// ---- Locks in paired-form coverage (KindJsxOpeningElement) ----
		{Code: `<a>x</a>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Locks in upstream regex /^\W*?javascript:/ positive cases ----
		// 1+ non-word prefix (whitespace, `#`, etc.) still matches.
		{Code: `<a href=" javascript:void(0)" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="	javascript:void(0)" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// `#` is non-word in `\W`, so `#javascript:foo` matches the regex
		// after the lazy quantifier consumes the `#`. Locks in the
		// non-zero-prefix branch.
		{Code: `<a href="#javascript:foo" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Dimension 4: TS expression wrappers around invalid values
		//      must still be rejected. `("#" as const)` should reach the
		//      `value === '#'` disjunct. ----
		{Code: `<a href={("#")} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"#" as string} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"" as const} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Locks in NoSubstitutionTemplateLiteral coverage. tsgo splits
		//      template literals into a separate kind from StringLiteral;
		//      both must be staticEval-able to a string. `` `#` `` should
		//      hit the `value === '#'` disjunct. ----
		{Code: "<a href={`#`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: "<a href={`javascript:void(0)`} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Locks in spread-of-literal walking. `<a {...{href: ""}}/>`
		//      has hasSpread=true AND populates hasAnyHref via the literal
		//      spread walk. The hrefs-found branch then fires invalidHref. ----
		{Code: `<a {...{href: ""}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a {...{href: "#"}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Mixed-validity props: one prop is a valid href, another is
		//      invalid. Upstream's filter takes ANY invalid in the array →
		//      report. Locks in the OR-merge across propsToValidate. ----
		{Code: `<a href="foo" hrefLeft="" />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="" hrefLeft="foo" />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Polymorphic-as-a: `<Foo as="a" />` resolves to "a" → checked.
		//      Locks in the GetElementType polymorphic-prop branch with no
		//      href → noHref report. ----
		{Code: `<Foo as="a" />`, Tsx: true, Settings: polymorphicSettings, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Two consecutive anchors → both reported independently. Locks
		//      in the per-element listener firing (no state leak). ----
		{
			Code: `<><a /><a href="#" /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 3},
				{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 8},
			},
		},

		// ---- Nested anchors: outer and inner both invalid (illegal HTML
		//      but legal JSX). Locks in the absence of bleed-through; both
		//      should be reported. ----
		{
			Code: `<a><a /></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1},
				{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 4},
			},
		},

		// ---- staticEval `||` short-circuit: truthy left = invalid value ----
		// `"javascript:" || "/foo"` — left is truthy (non-empty string) →
		// take left = "javascript:" → matches /^\W*?javascript:/ → invalid.
		// Locks in the BinaryExpression-`||` arm of staticEvalBinary.
		{Code: `<a href={"javascript:" || "/foo"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// `"" || "#"` — left is falsy → take right = "#" → invalid.
		{Code: `<a href={"" || "#"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// `null || ""` — left null (falsy) → take right = "" → invalid.
		{Code: `<a href={null || ""} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- staticEval `&&` short-circuit: truthy left → take right ----
		// `"foo" && "#"` — left truthy → take right = "#" → invalid.
		{Code: `<a href={"foo" && "#"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// `1 && "javascript:void(0)"` — number 1 truthy → take right →
		// invalid. Locks in non-string truthy left.
		{Code: `<a href={1 && "javascript:void(0)"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- staticEval `??` nullish coalescing: null left → take right ----
		// `null ?? "#"` and `undefined ?? "javascript:foo"` — both fall to
		// the right operand. Locks in the null-or-undefined detection arm.
		{Code: `<a href={null ?? "#"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={undefined ?? "javascript:foo"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// `"" ?? "#"` — empty string is NOT null/undef → take left = "" →
		// invalid (length 0). Lock in distinction between `??` and `||`.
		{Code: `<a href={"" ?? "#"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- `&&` short-circuit producing nullish result → noHref ----
		// `null && X` returns null per JS semantics (and staticEval mirrors
		// via "if !jsTruthy_(left) return left"). null is nullish →
		// PropValueIsNullish=true → hasAnyHref=false → noHref. Same path
		// for `undefined && X`. Crucially distinct from `false && X` and
		// `0 && X` which produce non-nullish falsy values (covered in
		// valid above).
		{Code: `<a href={null && "/foo"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={undefined && "/foo"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- ConditionalExpression: literal true → consequent ----
		// `true ? "javascript:" : "/foo"` → "javascript:" → invalid.
		{Code: `<a href={true ? "javascript:foo" : "/"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// `false ? "/" : "javascript:..."` → take alternate → invalid.
		{Code: `<a href={false ? "/" : "javascript:foo"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// `cond ? "#" : "/"` — cond is Identifier (truthy string per
		// jsx-ast-utils Identifier extractor) → take consequent = "#" →
		// invalid.
		{Code: `<a href={cond ? "#" : "/"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- staticEval string concatenation: result matches js: prefix ----
		// `"java" + "script:" + "alert(1)"` — concatenated to
		// "javascript:alert(1)" → matches regex → invalid. Locks in the
		// `+` BinaryExpression arm with same-side string operands.
		{Code: `<a href={"java" + "script:" + "alert(1)"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// String concat producing exact "#" — `"#" + ""` → "#" → invalid.
		{Code: `<a href={"#" + ""} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Multi-line code: position assertions on each line. Verifies
		//      Line/Column report at the opening `<` of each anchor. ----
		{
			Code: `<div>
  <a />
  <a href="#" />
</div>`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noHref", Message: noHrefErrorMessage, Line: 2, Column: 3},
				{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 3, Column: 3},
			},
		},

		// ---- Real-world Nav-component flow with mixed states. Locks in
		//      independent per-element classification across a realistic
		//      structure. ----
		{
			Code: `<nav>
  <a href="/">Home</a>
  <a href="/about">About</a>
  <a href="#" onClick={() => 0}>Click</a>
  <a>Empty</a>
</nav>`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 4, Column: 3},
				{MessageId: "noHref", Message: noHrefErrorMessage, Line: 5, Column: 3},
			},
		},

		// ---- Mixed valid + invalid with onClick → preferButton wins ----
		// `<a href="/" hrefLeft="" onClick={fn}/>` with specialLink=
		// ['hrefLeft']. hasAnyHref=true (via /), hasInvalidHref=true (via
		// hrefLeft=""), onClick + preferButton active → preferButton report.
		{Code: `<a href="/" hrefLeft="" onClick={() => 0} />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},

		// ---- Spread + explicit invalid href: hasSpread=true does NOT
		//      suppress the invalidHref report (spread suppression only
		//      applies to the no-href branch). ----
		{Code: `<a {...props} href="" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a {...props} href="javascript:void(0)" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Spread literal walk + onClick via spread: `<a {...{href: "#",
		//      onClick: fn}}/>` — both found via single literal-spread
		//      walk. hasAnyHref=true, hasInvalidHref=true, onClick present,
		//      preferButton active → preferButton report. ----
		{Code: `<a {...{href: "#", onClick: () => 0}} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},

		// ---- Attribute order: literal-spread BEFORE explicit href —
		//      FindAttributeByName returns the FIRST match (the spread one),
		//      mirroring upstream's `getProp` first-match behavior. The
		//      spread's invalid value drives the report even though the
		//      runtime override would set href="foo". This is upstream's
		//      known quirk and we mirror it. ----
		{Code: `<a {...{href: "#"}} href="foo" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- staticEval Identifier extractor: bare `<a href={undefined}/>`
		//      vs identifier whose name happens to be a known-truthy global.
		//      The `undefined` Identifier returns jsUndef → nullish → no
		//      href → noHref report. Locks in the special-cased Identifier
		//      handling. ----
		{Code: `<a href={undefined} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Empty JsxExpression `<a href={}/>` — tsgo accepts this for
		//      error-recovery; jsx-ast-utils routes JSXEmptyExpression
		//      through its `TYPES[type] === undefined → return null`
		//      fallback. `null != null` is false → no href → noHref. Locks
		//      in PropValueIsNullish's `inner == nil` branch alignment. ----
		{Code: `<a href={} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- TS `satisfies` wrapper — upstream's TYPES table has NO
		//      `TSSatisfiesExpression` extractor, so `getPropValue` returns
		//      null for `<a href={"foo" satisfies string}/>`. staticEval
		//      mirrors by EXCLUDING `OEKSatisfies` from skipTransparent;
		//      satisfies-wrapped values land on the default `jsNull` arm
		//      and PropValueIsNullish reports nullish → noHref. Catches
		//      what would otherwise be silent under-reporting. ----
		{Code: `<a href={"foo" satisfies string} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)" satisfies string} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		// JsxExpression containing only a comment is the same shape as
		// empty: tsgo strips trivia, leaving Expression nil. Same outcome.
		{Code: `<a href={/* todo */} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- onClick prop case-insensitive lookup. With no href, the
		//      lowercase legacy HTML attribute should still trigger
		//      preferButton (per jsx-ast-utils ignoreCase=true). ----
		{Code: `<a onclick={() => 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},

		// ---- Boolean form for special link: `<a hrefLeft />` with
		//      specialLink=['hrefLeft']. Boolean form maps to value `true`
		//      (per upstream's null-attribute-value handling) → hasAnyHref
		//      = true, not a string → not invalid. But what about pure
		//      `<a hrefLeft/>` with NO other href? Should be valid (we have
		//      a href-like prop with a non-null/non-string-invalid value).
		//      No invalid case here — moved to valid section actually.
		//      Below is a different shape that IS invalid:
		//      `<a hrefLeft="" />` — explicit empty string, length 0,
		//      invalid. Already covered in upstream tests.

		// ---- aspects array with both ON and unknown — unknown silently
		//      ignored, "noHref" still triggers reports. ----
		{Code: `<a />`, Tsx: true, Options: []interface{}{
			map[string]interface{}{"aspects": []interface{}{"noHref", "madeUpAspect"}},
		}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Bare-object options shape (single-option CLI / lint() API
		//      shape) for the invalid path. Without an array wrap, the
		//      rule must still parse `components` correctly — exercises
		//      GetOptionsMap's `map[string]interface{}` branch. ----
		{Code: `<Link />`, Tsx: true, Options: map[string]interface{}{
			"components": []interface{}{"Link"},
		}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Bare-object options with aspects override ----
		// Locks in the CLI-shape parse for the aspects path too: only
		// `noHref` active should still report on a bare anchor.
		{Code: `<a />`, Tsx: true, Options: map[string]interface{}{
			"aspects": []interface{}{"noHref"},
		}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- Two onClick reports across same parent: each anchor's
		//      listener invocation is independent. Locks in the absence
		//      of cross-element shared state in the rule. ----
		{
			Code: `<div><a onClick={() => 0} /><a href="#" onClick={() => 0} /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 6},
				{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 29},
			},
		},

		// ---- Anchor inside a conditional render expression. The listener
		//      fires regardless of the surrounding container. ----
		{
			Code: `function App({cond}) { return cond && <a />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noHref", Message: noHrefErrorMessage},
			},
		},
	})
}
