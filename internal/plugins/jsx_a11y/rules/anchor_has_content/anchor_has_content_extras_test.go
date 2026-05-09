package anchor_has_content

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsRuleOpts exercises the rule's `components` array option. Upstream's
// `typeCheck = ['a'].concat(componentOptions)` is appended AS-IS — names are
// matched against the *resolved* nodeType after `getElementType`'s polymorphic
// / componentMap step, but the option itself is not remapped.
var componentsRuleOpts = map[string]interface{}{
	"components": []interface{}{"Anchor", "Link"},
}

// polymorphicSettings exercises the `polymorphicPropName` settings entry —
// `<Foo as="a">` should be treated as `<a>`. This locks in the
// getElementType polymorphic-prop branch which upstream's own test file
// doesn't exercise for this rule.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// TestAnchorHasContentExtras locks in branches that upstream's test file
// doesn't exercise but are reachable through the rule's listener gate. Each
// case carries an inline comment pointing at the specific upstream branch /
// Dimension 4 edge shape it covers. These tests protect against silent
// regressions during refactors of either the rule or its shared helpers.
func TestAnchorHasContentExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AnchorHasContentRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: paired vs self-closing forms ----
		// Both shapes must be checked (KindJsxOpeningElement +
		// KindJsxSelfClosingElement listeners). `<a>x</a>` is the paired form.
		{Code: `<a>x</a>`, Tsx: true},

		// ---- Dimension 4: tag-name forms — case-sensitive HTML matching ----
		// Upstream's `typeCheck.indexOf(nodeType) === -1` is case-sensitive.
		// `<A />` resolves to nodeType "A" which is NOT in ['a'], so the rule
		// short-circuits and does NOT report — even though it has no content.
		{Code: `<A />`, Tsx: true},

		// ---- Dimension 4: namespaced JSX names → not matched ----
		// `<svg:a />` resolves to nodeType "svg:a", not "a" — skipped.
		{Code: `<svg:a />`, Tsx: true},

		// ---- Dimension 4: member-access tag → not matched ----
		// `<Foo.a />` resolves to "Foo.a", not "a" — skipped.
		{Code: `<Foo.a />`, Tsx: true},

		// ---- Dimension 4: case-insensitive attribute name (jsx-ast-utils
		//      `hasProp` / `getProp` default `ignoreCase: true`) ----
		// Locks in the title / aria-label match against ALL case variants.
		{Code: `<a TITLE="x" />`, Tsx: true},
		{Code: `<a Title="x" />`, Tsx: true},
		{Code: `<a ARIA-LABEL="x" />`, Tsx: true},
		{Code: `<a aria-Label="x" />`, Tsx: true},

		// ---- Dimension 4: boolean-attribute form ----
		// `<a title />` — upstream's `hasAnyProp` matches by NAME only, not
		// value. Boolean form still counts as "has title" → valid.
		{Code: `<a title />`, Tsx: true},
		{Code: `<a aria-label />`, Tsx: true},

		// ---- Dimension 4: title with empty / undefined value ----
		// `hasAnyProp` is value-blind — only the prop's presence matters.
		{Code: `<a title="" />`, Tsx: true},
		{Code: `<a title={undefined} />`, Tsx: true},
		{Code: `<a aria-label={null} />`, Tsx: true},

		// ---- Locks in upstream switch arm: Literal child (StringLiteral as
		//      JSX child) ---- — uncommon but legal (e.g. via Babel output).
		{Code: `<a>{"hello"}</a>`, Tsx: true},

		// ---- Locks in upstream switch arm: JSXExpressionContainer → non-
		//      Identifier expression. `0`, `null`, `false` are NOT
		//      Identifiers, so the switch falls through to `return true`.
		//      This matches a known upstream quirk — JS-falsy literals as
		//      JSX expressions still count as "accessible content". ----
		{Code: `<a>{0}</a>`, Tsx: true},
		{Code: `<a>{null}</a>`, Tsx: true},
		{Code: `<a>{false}</a>`, Tsx: true},

		// ---- Locks in upstream switch arm: JSXExpressionContainer →
		//      Identifier other than 'undefined' (e.g. variable reference).
		//      Returns true — runtime value unknown, assumed accessible. ----
		{Code: `<a>{someVar}</a>`, Tsx: true},

		// ---- Locks in upstream switch arm: JSXElement child whose opening
		//      element is NOT hidden from screen readers ----
		{Code: `<a><span>x</span></a>`, Tsx: true},
		{Code: `<a><span /></a>`, Tsx: true},

		// ---- Locks in upstream isHiddenFromScreenReader branches:
		//      aria-hidden="false" → ariaHidden !== true → not hidden ----
		{Code: `<a><span aria-hidden="false">x</span></a>`, Tsx: true},

		// ---- Locks in upstream isHiddenFromScreenReader: input[type=hidden]
		//      hides the element. Sibling text content keeps the anchor
		//      accessible regardless of the input. ----
		{Code: `<a><input type="hidden" />x</a>`, Tsx: true},

		// ---- Locks in upstream `hasAnyProp` fallback path inside
		//      hasAccessibleChild: `dangerouslySetInnerHTML` boolean form ----
		{Code: `<a dangerouslySetInnerHTML />`, Tsx: true},
		{Code: `<a children={null} />`, Tsx: true},

		// ---- Rule-options branch: `components` adds tag names to typeCheck ----
		{Code: `<Anchor>x</Anchor>`, Tsx: true, Options: componentsRuleOpts},
		{Code: `<Link>x</Link>`, Tsx: true, Options: componentsRuleOpts},
		// Tags NOT in components option are skipped entirely (not just
		// reported) — `<Other />` has no content but is unmatched, so valid.
		{Code: `<Other />`, Tsx: true, Options: componentsRuleOpts},

		// ---- Polymorphic-prop branch via `getElementType`:
		//      `<Foo as="a">x</Foo>` resolves to nodeType "a" → matched and
		//      passes via children. Locks in the `polymorphicPropName`
		//      settings path, which upstream's own anchor-has-content tests
		//      don't cover. ----
		{Code: `<Foo as="a">x</Foo>`, Tsx: true, Settings: polymorphicSettings},

		// ---- Multi-line content ----
		{Code: "<a>\n  <span>x</span>\n</a>", Tsx: true},

		// ---- Locks in upstream JSXText quirk: whitespace JsxText counts as
		//      accessible content (`!!child.value` is truthy for any non-empty
		//      string, including `"\n  "`). So a multi-line anchor whose only
		//      "real" child is `{undefined}` or a hidden element is STILL
		//      considered accessible because of the surrounding whitespace
		//      text nodes. This matches upstream verbatim. ----
		{Code: "<a>\n  {undefined}\n</a>", Tsx: true},
		{Code: "<a>\n  <Bar aria-hidden />\n</a>", Tsx: true},

		// ---- Spread-then-explicit-title (spread is opaque, but explicit
		//      attribute matches first). Locks in that explicit attrs are
		//      always seen even alongside non-literal spreads. ----
		{Code: `<a {...props} title="t" />`, Tsx: true},

		// ---- Dimension 2 nesting: outer anchor wraps an inner non-hidden
		//      element → outer is VALID. (Inner anchor without content would
		//      be reported separately on its own listener invocation, but
		//      this case puts content inside the inner anchor.) ----
		{Code: `<a><a>x</a></a>`, Tsx: true},

		// ---- Dimension 2: deeply nested ternary / conditional content.
		//      JsxExpression with a complex non-Identifier expression →
		//      upstream's switch returns true. ----
		{Code: `<a>{cond ? <span>x</span> : <span>y</span>}</a>`, Tsx: true},
		{Code: `<a>{cond && <span>x</span>}</a>`, Tsx: true},

		// ---- ComponentMap: explicitly remap 'a' → 'div' to disable the
		//      rule for native anchors. Locks in the listener gate's
		//      "skipped when nodeType is not in typeCheck" path. ----
		{Code: `<a />`, Tsx: true, Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"components": map[string]interface{}{"a": "div"}}}},

		// ---- ComponentMap: remap a custom name to 'a' AND give it content. ----
		{Code: `<Anchor>x</Anchor>`, Tsx: true, Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"components": map[string]interface{}{"Anchor": "a"}}}},

		// ---- Polymorphic-prop with allow-list: only `<Foo as=...>` is
		//      remapped, `<Bar as="a">` is NOT (rawType "Bar" not in allow). ----
		{Code: `<Bar as="a" />`, Tsx: true, Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as", "polymorphicAllowList": []interface{}{"Foo"}}}},

		// ============================================================
		// Real-world user patterns
		// ============================================================

		// ---- Common React Router / Next.js Link shapes (with componentMap). ----
		{Code: `<Link to="/about">About</Link>`, Tsx: true, Settings: linkSettings},
		{Code: `<Link to="/about"><Icon /> About</Link>`, Tsx: true, Settings: linkSettings},

		// ---- href + content + className (canonical anchor) ----
		{Code: `<a href="/about" className="link">About</a>`, Tsx: true},

		// ---- Children-via-prop: dynamic children value ----
		{Code: `<a href="/x">{children}</a>`, Tsx: true},
		{Code: `<a>{props.children}</a>`, Tsx: true},

		// ---- Icon-only anchor with explicit aria-label (canonical pattern). ----
		{Code: `<a href="/" aria-label="Home"><Icon aria-hidden="true" /></a>`, Tsx: true},
		{Code: `<a href="/" title="Home"><Icon aria-hidden="true" /></a>`, Tsx: true},

		// ---- Mixed content: hidden icon + visible text → text counts as
		//      accessible, regardless of the hidden icon. ----
		{Code: `<a href="/x"><i aria-hidden="true" className="icon" /> Click</a>`, Tsx: true},
		{Code: `<a href="/x"><i aria-hidden="true" /><span>Click</span></a>`, Tsx: true},

		// ---- Nested image with alt — img is non-hidden → accessible. ----
		{Code: `<a href="/x"><img src="x.png" alt="Home" /></a>`, Tsx: true},
		{Code: `<a href="/x"><img src="x.png" alt="" /></a>`, Tsx: true},
		{Code: `<a href="/x"><img src="x.png" /></a>`, Tsx: true},

		// ---- aria-hidden value variants on the child (lock in upstream's
		//      `getPropValue === true` semantic — both literal "true" string
		//      and JSX-expression boolean false / numeric / Identifier). ----
		{Code: `<a><span aria-hidden={false}>x</span></a>`, Tsx: true},
		{Code: `<a><span aria-hidden={1}>x</span></a>`, Tsx: true},
		// Identifier reference — not statically true → not hidden → accessible.
		{Code: `<a><span aria-hidden={someFlag}>x</span></a>`, Tsx: true},

		// ---- i18n / formatter call expressions as children. ----
		{Code: `<a>{t("hello")}</a>`, Tsx: true},
		{Code: `<a>{intl.formatMessage({id: "x"})}</a>`, Tsx: true},

		// ---- Logical / ternary expressions. ----
		{Code: `<a>{label || "Default"}</a>`, Tsx: true},
		{Code: `<a>{label ?? "Default"}</a>`, Tsx: true},
		{Code: `<a>{count > 0 ? "more" : "none"}</a>`, Tsx: true},

		// ---- Template literals and string concatenation. ----
		{Code: "<a>{`Hello ${name}`}</a>", Tsx: true},
		{Code: `<a>{"Hello " + name}</a>`, Tsx: true},

		// ---- Array map → JsxExpression with CallExpression → accessible. ----
		{Code: `<a>{items.map(i => <span key={i}>{i}</span>)}</a>`, Tsx: true},

		// ---- TypeScript-only expression wrappers on the child. Upstream's
		//      switch is `child.expression.type === 'Identifier'` — TS
		//      wrappers (`as` / `!` / `satisfies` / `<T>x`) are exposed as
		//      their own AST nodes in typescript-eslint, NOT as Identifier,
		//      so they fall through to `return true` (accessible). Even when
		//      the inner expression is the bare identifier `undefined`, the
		//      outer wrapper saves the JsxExpression from the rule. ----
		{Code: `<a>{x as string}</a>`, Tsx: true},
		{Code: `<a>{x!}</a>`, Tsx: true},
		{Code: `<a>{(x as any) satisfies string}</a>`, Tsx: true},
		{Code: `<a>{undefined as any}</a>`, Tsx: true},
		{Code: `<a>{(undefined)!}</a>`, Tsx: true},
		{Code: `<a>{undefined satisfies any}</a>`, Tsx: true},

		// ---- Spread + content (non-literal spread is opaque, content saves it). ----
		{Code: `<a {...rest}>x</a>`, Tsx: true},
		{Code: `<a {...rest}>{children}</a>`, Tsx: true},

		// ---- Multiple non-text children including some hidden ----
		{Code: `<a><span aria-hidden="true" /><span>x</span></a>`, Tsx: true},
		{Code: `<a>Click <span aria-hidden="true">→</span></a>`, Tsx: true},

		// ---- Deeply nested: outer always has a child, no need to recurse. ----
		{Code: `<a><b><i><u>deeply nested</u></i></b></a>`, Tsx: true},

		// ---- Children prop with explicit JSX value → fallback path. ----
		{Code: `<a children={<span>x</span>} />`, Tsx: true},
		{Code: `<a children={[<span key="1">x</span>]} />`, Tsx: true},
		{Code: `<a children="text" />`, Tsx: true},
		{Code: `<a children />`, Tsx: true},

		// ---- Whitespace-only text node still counts as content (upstream
		//      `!!child.value`; tsgo splits into KindJsxTextAllWhiteSpaces
		//      with non-empty Text). ----
		{Code: `<a>{" "}</a>`, Tsx: true},

		// ---- JSX comment children: pure-comment JsxExpression has no
		//      meaningful payload — but text/whitespace siblings keep the
		//      anchor accessible. (Pure-comment-only case is below in INVALID.) ----
		{Code: `<a>x{/* a comment */}</a>`, Tsx: true},
		{Code: `<a>{/* a comment */ x}</a>`, Tsx: true},

		// ---- Common DOM-attr-bearing anchors: onClick / download / target /
		//      rel / role — they don't affect accessibility, content does. ----
		{Code: `<a href="/x" onClick={handler}>Click</a>`, Tsx: true},
		{Code: `<a href="/x" download>Download</a>`, Tsx: true},
		{Code: `<a href="/x" target="_blank" rel="noopener">External</a>`, Tsx: true},
		{Code: `<a href="/x" role="button">Submit</a>`, Tsx: true},
		{Code: `<a href="/x" tabIndex={0}>x</a>`, Tsx: true},

		// ---- One undefined + one real text child → real text wins. ----
		{Code: `<a>{undefined}x</a>`, Tsx: true},
		{Code: `<a>x{undefined}</a>`, Tsx: true},

		// ---- Mixed text and inline elements (canonical anchor content). ----
		{Code: `<a>Click <strong>here</strong> now</a>`, Tsx: true},
		{Code: `<a><span>Read </span><span>more</span></a>`, Tsx: true},

		// ---- React.createElement-style rendering inside expression child ----
		{Code: `<a>{React.createElement("span", null, "x")}</a>`, Tsx: true},

		// ---- ReactDOM.createPortal call as expression child ----
		{Code: `<a>{ReactDOM.createPortal(content, target)}</a>`, Tsx: true},

		// ---- Locks in upstream quirk: ANY non-Identifier JsxExpression
		//      returns true unconditionally. Empty string, NaN, etc. are
		//      accepted even though they're falsy / empty at runtime. The
		//      upstream switch's `case 'JSXExpressionContainer'` short-
		//      circuits on `expression.type === 'Identifier'` — so non-
		//      Identifier (Literal, BinaryExpression, …) always return true. ----
		{Code: `<a>{""}</a>`, Tsx: true},
		{Code: `<a>{NaN}</a>`, Tsx: true},
		{Code: `<a>{void 0}</a>`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Paired empty form: `<a></a>` has no content (whitespace-free) ----
		{
			Code: `<a></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- JsxFragment child → upstream's switch has no JSXFragment arm
		//      and falls to `default: return false`. The fragment's text
		//      does NOT count as accessible content. ----
		{
			Code: `<a><>x</></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- All children are hidden elements → no accessible content ----
		{
			Code: `<a><span aria-hidden="true">x</span></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- input[type="hidden"] is hidden; no other children → invalid ----
		{
			Code: `<a><input type="hidden" /></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- JsxExpression with nothing inside (`{}` empty container) → no
		//      accessible content. Upstream's switch hits JSXExpressionContainer
		//      with no expression, which doesn't return early, but child.expression
		//      is undefined; in tsgo we synthesize an empty JsxExpression. ----
		{
			Code: `<a>{}</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Locks in upstream `node.parent` traversal: paired form
		//      reports on the JsxOpeningElement (column 1, not the full
		//      JsxElement). ----
		{
			Code: `  <a></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 3},
			},
		},

		// ---- Rule-options branch: custom component without content → invalid ----
		{
			Code:    `<Anchor />`,
			Tsx:     true,
			Options: componentsRuleOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Link />`,
			Tsx:     true,
			Options: componentsRuleOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Polymorphic-prop: `<Foo as="a" />` resolves to "a", no content. ----
		{
			Code:     `<Foo as="a" />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Dimension 2: nested anchors — only the inner empty `<a />` is
		//      reported. The outer wraps a non-hidden element so it's
		//      accessible. Locks in that listeners do NOT bleed across
		//      JsxElement boundaries. ----
		{
			Code: `<a><a /></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 4},
			},
		},

		// ---- ComponentMap: remap a custom name to 'a' WITHOUT content. ----
		{
			Code:     `<Anchor />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"components": map[string]interface{}{"Anchor": "a"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Polymorphic with allow-list including `<Foo>` — `<Foo as="a">`
		//      gets remapped to "a" → matched → no content → reported. ----
		{
			Code:     `<Foo as="a" />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as", "polymorphicAllowList": []interface{}{"Foo"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Real-world misuse patterns (high-value invalid cases)
		// ============================================================

		// ---- `aria-labelledby` is NOT in upstream's title/aria-label list,
		//      so an anchor with ONLY aria-labelledby and no content is
		//      reported. This is a non-obvious behavior — locked in
		//      explicitly to catch any future broadening of the prop list. ----
		{
			Code: `<a aria-labelledby="id1" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- href without content is the most common real-world misuse. ----
		{
			Code: `<a href="/about" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<a href="/about" className="link" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Icon-only anchor without aria-label — the icon's `aria-hidden`
		//      hides it from screen readers, leaving nothing accessible. ----
		{
			Code: `<a href="/"><Icon aria-hidden="true" /></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<a href="/"><i aria-hidden="true" className="icon-x" /></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Hidden `<input type="image" />` provides no text alternative
		//      to the parent. Even though images normally count, type=hidden
		//      classes them as hidden per upstream isHiddenFromScreenReader. ----
		{
			Code: `<a href="/"><input type="hidden" /></a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Pure JSX comment child: `<a>{/* xx */}</a>` — JsxExpression
		//      with no expression payload. Not accessible. Locks in the
		//      `expr == nil` branch in HasAccessibleChild. ----
		{
			Code: `<a>{/* a comment */}</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- TS wrappers around `undefined` are NOT stripped by upstream's
		//      switch — `child.expression.type === 'Identifier'` is checked
		//      directly. typescript-eslint exposes `as` / `!` / `satisfies`
		//      as their own AST nodes whose type is NOT 'Identifier', so the
		//      switch falls through to `return true` (= accessible). Pure
		//      parens, by contrast, are auto-flattened by ESTree → reach
		//      Identifier `undefined` → reported. ----
		{
			Code: `<a>{((undefined))}</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Spread-only anchor (non-literal spread is opaque). The whole
		//      anchor delegates content to runtime — but per upstream this
		//      is INVALID because the spread can't be statically verified. ----
		{
			Code: `<a {...this.props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Multiple anchors in one file: each is independently
		//      validated. Locks in that the listener is stateless. ----
		{
			Code: `<><a /><a>x</a><a /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 3},
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 16},
			},
		},

		// ---- Anchors inside conditional expressions are still visited. ----
		{
			Code: `cond ? <a /> : <a>x</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 8},
			},
		},

		// ---- Multiple `{undefined}` siblings — neither contributes content,
		//      and there's no whitespace/text/element sibling. ----
		{
			Code: `<a>{undefined}{undefined}</a>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Both spread AND empty children form. Spread doesn't help
		//      (non-literal, opaque), no other content. ----
		{
			Code: `<a {...rest} {...other} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Spread-strict alignment with upstream `hasAnyProp` defaults
		// ============================================================
		// Upstream uses `hasAnyProp` (default `spreadStrict: true`) for
		// title / aria-label and inside hasAccessibleChild's fallback for
		// dangerouslySetInnerHTML / children. Even literal ObjectLiteral
		// spreads are opaque — the prop must appear as a DIRECT JsxAttribute.
		// These cases lock that semantic in.

		// ---- title in literal spread → opaque → no title found → INVALID ----
		{
			Code: `<a {...{title: 'x'}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- aria-label in literal spread → opaque → INVALID ----
		{
			Code: `<a {...{"aria-label": 'x'}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- children in literal spread → opaque (hasAccessibleChild
		//      fallback uses spread-strict hasAnyProp) → INVALID ----
		{
			Code: `<a {...{children: 'x'}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- dangerouslySetInnerHTML in literal spread → opaque → INVALID ----
		{
			Code: `<a {...{dangerouslySetInnerHTML: {__html: "x"}}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- shorthand form `{...{title}}` is also opaque under strict ----
		{
			Code: `<a {...{title}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
	})
}

// TestAnchorHasContentOptionParsing exercises the JSON path that real CLI /
// JS configs take through `utils.GetOptionsMap` (vs. typed-struct shortcuts).
// Per the SKILL guidance, options coverage MUST include cases where Options
// is a bare map / array-wrapped map — typed structs short-circuit
// GetOptionsMap and never exercise the round-trip CLI tests rely on.
func TestAnchorHasContentOptionParsing(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AnchorHasContentRule, []rule_tester.ValidTestCase{
		// Single-option CLI shape — bare map.
		{Code: `<Anchor>x</Anchor>`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{"Anchor"}}},
		// rule_tester / multi-element shape — array-wrapped map.
		{Code: `<Anchor>x</Anchor>`, Tsx: true, Options: []interface{}{map[string]interface{}{"components": []interface{}{"Anchor"}}}},
		// Empty / nil options must not crash and must default to ['a'] only.
		{Code: `<Custom />`, Tsx: true},
		{Code: `<Custom />`, Tsx: true, Options: map[string]interface{}{}},
		// Empty components array — same as defaults.
		{Code: `<Custom />`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{}}},
		// Components with non-string entries are silently skipped — locks in
		// the type-guard inside parseOptions.
		{Code: `<Anchor>x</Anchor>`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{"Anchor", 42, true}}},
		// Malformed components (string instead of array) — must not crash;
		// falls back to defaults. Schema validation would reject this in
		// real ESLint usage, but the rule must still be defensive.
		{Code: `<Anchor />`, Tsx: true, Options: map[string]interface{}{"components": "Anchor"}},
		// Unknown options keys are ignored.
		{Code: `<Anchor />`, Tsx: true, Options: map[string]interface{}{"unknown": true, "components": []interface{}{}}},
		// Components includes "a" (duplicate) — typeCheck has duplicates but
		// slices.Contains short-circuits on first match.
		{Code: `<a>x</a>`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{"a"}}},
	}, []rule_tester.InvalidTestCase{
		// Bare-map option shape produces a report on the matched custom tag.
		{
			Code:    `<Anchor />`,
			Tsx:     true,
			Options: map[string]interface{}{"components": []interface{}{"Anchor"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anchorHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
	})
}
