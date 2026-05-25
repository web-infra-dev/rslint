package heading_has_content

// cspell:ignore Ελληνικά عنوان chakra

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicSettings exercises the `polymorphicPropName` settings entry —
// `<Foo as="h1">` should be treated as `<h1>`. Locks in the getElementType
// polymorphic-prop branch which upstream's own test file doesn't exercise
// for this rule.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// remapASettings remaps a custom component to a native heading via the
// componentMap settings path.
var remapASettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"MyHeader": "h2",
		},
	},
}

// remapHToDivSettings remaps native h1 to a non-heading via componentMap —
// disables the rule for raw `<h1>`.
var remapHToDivSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"h1": "div",
		},
	},
}

// TestHeadingHasContentExtras locks in branches that upstream's test file
// doesn't exercise but are reachable through the rule's listener gate. Each
// case carries an inline comment pointing at the specific upstream branch /
// Dimension 4 edge shape it covers.
func TestHeadingHasContentExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &HeadingHasContentRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: paired vs self-closing forms ----
		// Both shapes must be checked (KindJsxOpeningElement +
		// KindJsxSelfClosingElement listeners). `<h1>x</h1>` is the paired form.
		{Code: `<h1>x</h1>`, Tsx: true},

		// ---- Dimension 4: tag-name forms — case-sensitive HTML matching ----
		// Upstream's `typeCheck.indexOf(nodeType) === -1` is case-sensitive.
		// `<H1 />` resolves to nodeType "H1" which is NOT in headings, so the
		// rule short-circuits and does NOT report — even though it has no content.
		{Code: `<H1 />`, Tsx: true},

		// ---- Dimension 4: namespaced JSX names → not matched ----
		{Code: `<svg:h1 />`, Tsx: true},

		// ---- Dimension 4: member-access tag → not matched ----
		{Code: `<Foo.h1 />`, Tsx: true},

		// ---- All six heading levels are handled ----
		{Code: `<h1>x</h1>`, Tsx: true},
		{Code: `<h2>x</h2>`, Tsx: true},
		{Code: `<h3>x</h3>`, Tsx: true},
		{Code: `<h4>x</h4>`, Tsx: true},
		{Code: `<h5>x</h5>`, Tsx: true},
		{Code: `<h6>x</h6>`, Tsx: true},

		// ---- Non-heading tags outside h1–h6 are skipped — even with no content. ----
		{Code: `<h7 />`, Tsx: true},
		{Code: `<h0 />`, Tsx: true},
		{Code: `<heading />`, Tsx: true},

		// ---- aria-hidden on the heading exempts it from the rule ----
		// (upstream's third gate `isHiddenFromScreenReader(nodeType, attrs)`).
		{Code: `<h1 aria-hidden />`, Tsx: true},
		{Code: `<h1 aria-hidden="true" />`, Tsx: true},
		{Code: `<h1 aria-hidden={true} />`, Tsx: true},
		// Paired form with aria-hidden — also exempt.
		{Code: `<h2 aria-hidden></h2>`, Tsx: true},
		{Code: `<h2 aria-hidden="true"></h2>`, Tsx: true},

		// ---- Locks in upstream switch arm: Literal child (StringLiteral as
		//      JSX child) ---- — uncommon but legal (e.g. via Babel output).
		{Code: `<h1>{"hello"}</h1>`, Tsx: true},

		// ---- Locks in upstream switch arm: JSXExpressionContainer → non-
		//      Identifier expression. `0`, `null`, `false` are NOT
		//      Identifiers, so the switch falls through to `return true`. ----
		{Code: `<h1>{0}</h1>`, Tsx: true},
		{Code: `<h1>{null}</h1>`, Tsx: true},
		{Code: `<h1>{false}</h1>`, Tsx: true},

		// ---- Locks in upstream switch arm: JSXExpressionContainer →
		//      Identifier other than 'undefined' (e.g. variable reference). ----
		{Code: `<h1>{someVar}</h1>`, Tsx: true},

		// ---- Locks in upstream switch arm: JSXElement child whose opening
		//      element is NOT hidden from screen readers ----
		{Code: `<h1><span>x</span></h1>`, Tsx: true},
		{Code: `<h1><span /></h1>`, Tsx: true},

		// ---- Locks in upstream isHiddenFromScreenReader branches:
		//      aria-hidden="false" on a CHILD → not hidden → accessible. ----
		{Code: `<h1><span aria-hidden="false">x</span></h1>`, Tsx: true},

		// ---- Locks in upstream isHiddenFromScreenReader: input[type=hidden]
		//      hides only the input. Sibling text content keeps the heading
		//      accessible regardless. ----
		{Code: `<h1><input type="hidden" />x</h1>`, Tsx: true},

		// ---- Locks in upstream `hasAnyProp` fallback path inside
		//      hasAccessibleChild: dangerouslySetInnerHTML / children boolean form ----
		{Code: `<h1 dangerouslySetInnerHTML />`, Tsx: true},
		{Code: `<h1 children />`, Tsx: true},
		{Code: `<h1 children={null} />`, Tsx: true},

		// ---- Rule-options branch: components adds tag names to typeCheck ----
		{Code: `<Heading>x</Heading>`, Tsx: true, Options: componentsRuleOpts},
		{Code: `<Title>x</Title>`, Tsx: true, Options: componentsRuleOpts},
		// Tags NOT in components option are skipped entirely (not just
		// reported) — `<Other />` has no content but is unmatched, so valid.
		{Code: `<Other />`, Tsx: true, Options: componentsRuleOpts},

		// ---- Polymorphic-prop branch via getElementType:
		//      `<Foo as="h1">x</Foo>` resolves to nodeType "h1" → matched and
		//      passes via children. ----
		{Code: `<Foo as="h1">x</Foo>`, Tsx: true, Settings: polymorphicSettings},

		// ---- Multi-line content ----
		{Code: "<h1>\n  <span>x</span>\n</h1>", Tsx: true},

		// ---- Locks in upstream JSXText quirk: whitespace JsxText counts as
		//      accessible content. Multi-line heading w/ only `{undefined}` or
		//      hidden child STILL accessible because of surrounding whitespace
		//      text nodes. ----
		{Code: "<h1>\n  {undefined}\n</h1>", Tsx: true},
		{Code: "<h1>\n  <Bar aria-hidden />\n</h1>", Tsx: true},

		// ---- ComponentMap: explicitly remap 'h1' → 'div' to disable the
		//      rule for native h1. Locks in the listener gate's "skipped when
		//      nodeType is not in typeCheck" path. ----
		{Code: `<h1 />`, Tsx: true, Settings: remapHToDivSettings},

		// ---- ComponentMap: remap a custom name to 'h2' AND give it content. ----
		{Code: `<MyHeader>x</MyHeader>`, Tsx: true, Settings: remapASettings},

		// ---- Polymorphic-prop with allow-list: only `<Foo as=...>` is
		//      remapped, `<Bar as="h1">` is NOT (rawType "Bar" not in allow). ----
		{Code: `<Bar as="h1" />`, Tsx: true, Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as", "polymorphicAllowList": []interface{}{"Foo"}}}},

		// ============================================================
		// Real-world user patterns
		// ============================================================

		// ---- Headings with className, id, role, etc. ----
		{Code: `<h1 className="title">Hello</h1>`, Tsx: true},
		{Code: `<h2 id="section">About</h2>`, Tsx: true},
		{Code: `<h1 role="heading" aria-level={1}>Hello</h1>`, Tsx: true},

		// ---- Mixed content: hidden icon + visible text → text counts as
		//      accessible, regardless of the hidden icon. ----
		{Code: `<h1><i aria-hidden="true" className="icon" /> Title</h1>`, Tsx: true},
		{Code: `<h2><i aria-hidden="true" /><span>Section</span></h2>`, Tsx: true},

		// ---- Common React components ----
		{Code: `<h1>{t("page.title")}</h1>`, Tsx: true},
		{Code: `<h1>{intl.formatMessage({id: "x"})}</h1>`, Tsx: true},

		// ---- Logical / ternary expressions. ----
		{Code: `<h1>{label || "Default"}</h1>`, Tsx: true},
		{Code: `<h1>{label ?? "Default"}</h1>`, Tsx: true},
		{Code: `<h1>{count > 0 ? "more" : "none"}</h1>`, Tsx: true},

		// ---- Template literals and string concatenation. ----
		{Code: "<h1>{`Hello ${name}`}</h1>", Tsx: true},
		{Code: `<h1>{"Hello " + name}</h1>`, Tsx: true},

		// ---- TypeScript-only expression wrappers on the child. Upstream's
		//      switch is `child.expression.type === 'Identifier'` — TS
		//      wrappers are NOT Identifier nodes, so they fall through to
		//      `return true` (accessible). Even when the inner is `undefined`,
		//      the outer wrapper saves the JsxExpression from the rule. ----
		{Code: `<h1>{x as string}</h1>`, Tsx: true},
		{Code: `<h1>{x!}</h1>`, Tsx: true},
		{Code: `<h1>{undefined as any}</h1>`, Tsx: true},
		{Code: `<h1>{(undefined)!}</h1>`, Tsx: true},

		// ---- Spread + content (non-literal spread is opaque, content saves it). ----
		{Code: `<h1 {...rest}>x</h1>`, Tsx: true},
		{Code: `<h1 {...rest}>{children}</h1>`, Tsx: true},

		// ---- Whitespace-only text node still counts as content (upstream
		//      `!!child.value`; tsgo splits into KindJsxTextAllWhiteSpaces
		//      with non-empty Text). ----
		{Code: `<h1>{" "}</h1>`, Tsx: true},

		// ---- JSX comment children: text/whitespace siblings keep accessible. ----
		{Code: `<h1>x{/* a comment */}</h1>`, Tsx: true},
		{Code: `<h1>{/* a comment */ x}</h1>`, Tsx: true},

		// ---- Locks in upstream quirk: ANY non-Identifier JsxExpression
		//      returns true unconditionally. Empty string, NaN, etc. are
		//      accepted even though they're falsy / empty at runtime. ----
		{Code: `<h1>{""}</h1>`, Tsx: true},
		{Code: `<h1>{NaN}</h1>`, Tsx: true},
		{Code: `<h1>{void 0}</h1>`, Tsx: true},

		// ---- Mixed text and inline elements (canonical heading content). ----
		{Code: `<h1>Welcome to <strong>Site</strong></h1>`, Tsx: true},
		{Code: `<h2><span>Section </span><span>1</span></h2>`, Tsx: true},

		// ---- Multi-level heading with nested content ----
		{Code: `<h1><h2>nested</h2></h1>`, Tsx: true},

		// ---- One undefined + one real text child → real text wins. ----
		{Code: `<h1>{undefined}x</h1>`, Tsx: true},
		{Code: `<h1>x{undefined}</h1>`, Tsx: true},

		// ============================================================
		// tsgo AST shape essentials — empty / paren / wrapper variants
		// ============================================================

		// ---- Empty JsxExpression / pure-comment expression. Upstream's
		//      `JSXEmptyExpression` is NOT Identifier → fall through →
		//      return true (accessible). HasAccessibleChild mirrors. ----
		{Code: `<h1>{}</h1>`, Tsx: true},
		{Code: `<h1>{/* a comment */}</h1>`, Tsx: true},
		{Code: `<h1>{/* lots */ /* of */ /* comments */}</h1>`, Tsx: true},

		// ---- Multi-level parenthesized non-undefined Identifier — paren
		//      stripping must NOT confuse with `undefined`. ----
		{Code: `<h1>{(text)}</h1>`, Tsx: true},
		{Code: `<h1>{((text))}</h1>`, Tsx: true},
		{Code: `<h1>{(((text)))}</h1>`, Tsx: true},

		// ---- TS wrappers around non-undefined Identifier — wrappers are
		//      not Identifier nodes, switch returns true. Equivalent for
		//      undefined: `<h1>{undefined as any}</h1>` (already in upstream
		//      pattern set). ----
		{Code: `<h1>{(text as any)}</h1>`, Tsx: true},
		{Code: `<h1>{((text) as any)}</h1>`, Tsx: true},
		{Code: `<h1>{text!}</h1>`, Tsx: true},
		{Code: `<h1>{(text)!}</h1>`, Tsx: true},
		{Code: `<h1>{text satisfies string}</h1>`, Tsx: true},
		{Code: `<h1>{(x as any) satisfies string}</h1>`, Tsx: true},

		// ---- BinaryExpression / call / member / new — none of these is
		//      Identifier → accessible. ----
		{Code: `<h1>{a + b}</h1>`, Tsx: true},
		{Code: `<h1>{getTitle()}</h1>`, Tsx: true},
		{Code: `<h1>{store.title}</h1>`, Tsx: true},
		{Code: `<h1>{new Title()}</h1>`, Tsx: true},
		{Code: `<h1>{typeof title}</h1>`, Tsx: true},
		{Code: `<h1>{void title}</h1>`, Tsx: true},
		{Code: `<h1>{delete obj.x}</h1>`, Tsx: true},

		// ---- Tagged-template / generator / await / yield (in async/gen
		//      contexts these are syntactically valid as expressions
		//      provided the parser accepts them — the test code is wrapped
		//      via Tsx). All non-Identifier → accessible. ----
		{Code: "<h1>{tag`hi ${name}`}</h1>", Tsx: true},

		// ---- Numeric / BigInt / RegExp literals → JSXExpressionContainer
		//      with non-Identifier expression → accessible. ----
		{Code: `<h1>{42}</h1>`, Tsx: true},
		{Code: `<h1>{42n}</h1>`, Tsx: true},
		{Code: `<h1>{/regex/g}</h1>`, Tsx: true},

		// ---- Spread inside expression child position — JsxExpression
		//      payload is the spread itself, non-Identifier → accessible.
		//      (Note: this is `{...x}` inside JSX child, distinct from JSX
		//      attribute spread.) Actually `<h1>{...x}</h1>` is invalid JSX.
		//      Skip. ----

		// ============================================================
		// isHiddenFromScreenReader matrix on the heading itself
		// ============================================================

		// ---- aria-hidden="true" string literal — staticEval normalizes
		//      "true" / "false" to booleans → exempt. ----
		{Code: `<h1 aria-hidden="true">x</h1>`, Tsx: true},
		// Boolean form `<h1 aria-hidden></h1>` — upstream maps null-attr
		// value to `true`, so `=== true` exempts the heading. Empty form
		// is in valid here. Empty-string form (`aria-hidden=""`) is NOT
		// exempt — see invalid block below for that case.
		{Code: `<h1 aria-hidden></h1>`, Tsx: true},

		// ---- aria-hidden values that are NOT statically resolvable to
		//      `true` — heading is NOT exempt; falls back to content
		//      check. Content present → valid. ----
		{Code: `<h1 aria-hidden={someFlag}>x</h1>`, Tsx: true},
		{Code: `<h1 aria-hidden={cond ? true : false}>x</h1>`, Tsx: true},
		{Code: `<h1 aria-hidden="false">x</h1>`, Tsx: true},
		{Code: `<h1 aria-hidden={1}>x</h1>`, Tsx: true},

		// ============================================================
		// Components option permutations
		// ============================================================

		// ---- Components option with mixed native + custom: `<h1>` and
		//      custom both must trigger when empty (custom does, h1 does
		//      via builtin). Content via children covers both. ----
		{Code: `<h1>x</h1>`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{"Heading"}}},
		{Code: `<Heading>x</Heading>`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{"Heading"}}},

		// ---- Long components list — slices.Contains over many strings ----
		{Code: `<E>x</E>`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{"A", "B", "C", "D", "E", "F"}}},

		// ============================================================
		// Real-world React / Next.js / i18n patterns
		// ============================================================

		// ---- next/head, react-helmet-style heading content ----
		{Code: `<h1 className="text-4xl font-bold">{post.title}</h1>`, Tsx: true},
		{Code: `<h2 id="introduction" data-testid="intro">Welcome</h2>`, Tsx: true},
		{Code: `<h1 ref={headingRef}>Hello</h1>`, Tsx: true},

		// ---- i18n libraries (react-intl, react-i18next) ----
		{Code: `<h1><FormattedMessage id="home.title" /></h1>`, Tsx: true},
		{Code: `<h1>{t("home.title")}</h1>`, Tsx: true},
		{Code: `<h2>{i18n.t("section.heading")}</h2>`, Tsx: true},
		{Code: `<h3>{useTranslation().t("x")}</h3>`, Tsx: true},

		// ---- Brand / icon prefix patterns where icon is aria-hidden and
		//      visible text follows ----
		{Code: `<h1><svg aria-hidden="true" /> Hello</h1>`, Tsx: true},
		{Code: `<h2><Icon aria-hidden /> Section title</h2>`, Tsx: true},
		{Code: `<h3><span aria-hidden="true" className="icon">📚</span> Books</h3>`, Tsx: true},

		// ---- Server component / RSC patterns (async-style) ----
		{Code: `<h1>{await getTitle()}</h1>`, Tsx: true},

		// ---- Headings rendered through wrapper components — `<Heading>`
		//      mapped to native h2 via componentMap settings. ----
		{Code: `<MyHeader>Branded title</MyHeader>`, Tsx: true, Settings: remapASettings},

		// ---- Polymorphic-prop heading components (chakra-ui / radix style) ----
		{Code: `<Heading as="h1">Title</Heading>`, Tsx: true, Settings: polymorphicSettings},
		{Code: `<Box as="h2">Section</Box>`, Tsx: true, Settings: polymorphicSettings},

		// ---- Multi-line headings with wrapping content ----
		{Code: "<h1>\n  <span>First line</span>\n  <span>second line</span>\n</h1>", Tsx: true},
		{Code: "<h1>\n  Welcome\n  <strong>back</strong>\n</h1>", Tsx: true},

		// ---- Deeply nested structural wrappers — outer's direct child
		//      is non-hidden, accessibility gate passes without recursing
		//      into the wrapper. ----
		{Code: `<h1><div><span><strong>Deep</strong></span></div></h1>`, Tsx: true},
		{Code: `<h1><div className="hero"><span>Deep</span></div></h1>`, Tsx: true},

		// ---- Unicode / emoji content (multi-byte). Locks in Text != "" ----
		{Code: `<h1>📚 标题</h1>`, Tsx: true},
		{Code: `<h2>한국어 제목</h2>`, Tsx: true},
		{Code: `<h3>Ελληνικά</h3>`, Tsx: true},
		{Code: `<h4>عنوان</h4>`, Tsx: true},

		// ---- Heading with ONLY whitespace text content — tsgo's
		//      KindJsxTextAllWhiteSpaces.Text is non-empty (e.g. "\n  "),
		//      and upstream `!!child.value` is also truthy for whitespace
		//      strings. So `<h1>\n</h1>` is accessible per upstream. ----
		{Code: "<h1>\n</h1>", Tsx: true},
		{Code: "<h1>\n  \n</h1>", Tsx: true},
		{Code: "<h1>   </h1>", Tsx: true},

		// ---- Heading with React.Fragment via long form — note: long-form
		//      `<React.Fragment>` is a JSXElement (not JSXFragment), so it
		//      goes through the JSXElement arm and is NOT auto-skipped
		//      like `<>...</>`. Its opening element is non-hidden →
		//      accessible. Locks in JSXElement vs JSXFragment asymmetry. ----
		{Code: `<h1><React.Fragment>x</React.Fragment></h1>`, Tsx: true},
		{Code: `<h1><Fragment>x</Fragment></h1>`, Tsx: true},

		// ---- Boolean / falsy attribute children that contribute text-like
		//      content via wrapper. ----
		{Code: `<h1>{" "}{title}</h1>`, Tsx: true},

		// ---- Complex tsgo wrapper chain: paren + as + non-null. ----
		{Code: `<h1>{((x as string)!)}</h1>`, Tsx: true},
		{Code: `<h1>{(x! as string)}</h1>`, Tsx: true},

		// ---- Multiple distinct headings on different lines, all valid ----
		{Code: "<>\n  <h1>One</h1>\n  <h2>Two</h2>\n  <h3>Three</h3>\n</>", Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Paired empty form: `<h1></h1>` has no content (whitespace-free) ----
		{
			Code: `<h1></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<h2></h2>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<h6></h6>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Self-closing forms for every heading level ----
		{
			Code: `<h2 />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<h6 />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- JsxFragment child → upstream's switch has no JSXFragment arm
		//      and falls to `default: return false`. The fragment's text
		//      does NOT count as accessible content. ----
		{
			Code: `<h1><>x</></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- All children are hidden elements → no accessible content ----
		{
			Code: `<h1><span aria-hidden="true">x</span></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Locks in `node.parent` traversal: paired form reports on the
		//      JsxOpeningElement (column 1, not the full JsxElement). ----
		{
			Code: `  <h1></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 3},
			},
		},

		// ---- Polymorphic-prop: `<Foo as="h1" />` resolves to "h1", no content. ----
		{
			Code:     `<Foo as="h1" />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Dimension 2: nested empty headings — only the inner empty
		//      `<h2 />` is reported. The outer wraps a non-hidden element so
		//      it's accessible. Locks in that listeners do NOT bleed across
		//      JsxElement boundaries. ----
		{
			Code: `<h1><h2 /></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 5},
			},
		},

		// ---- Outer h1 wraps a self-closing `<h2 aria-hidden />`. The inner
		//      h2 is itself exempt (3rd gate: aria-hidden on the heading
		//      skips). The outer h1's only child is a hidden h2 → no
		//      accessible content → outer reported. Locks in the asymmetry:
		//      aria-hidden on the heading exempts THAT heading but DOES
		//      hide it from accessibility-child checks of an enclosing
		//      heading. ----
		{
			Code: `<h1><h2 aria-hidden /></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- ComponentMap: remap a custom name to 'h2' WITHOUT content. ----
		{
			Code:     `<MyHeader />`,
			Tsx:      true,
			Settings: remapASettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Real-world misuse patterns (high-value invalid cases)
		// ============================================================

		// ---- `aria-labelledby` is NOT a content substitute (this rule has
		//      no title/aria-label fallback like anchor-has-content does).
		//      An empty heading with only aria-labelledby is still reported. ----
		{
			Code: `<h1 aria-labelledby="id1" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<h1 aria-label="Title" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<h1 title="Title" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Hidden-input-only heading. The input doesn't provide alt text. ----
		{
			Code: `<h2><input type="hidden" /></h2>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Parens around `undefined` → ESTree flattens, tsgo strips parens.
		//      So `<h1>{(undefined)}</h1>` reaches Identifier `undefined` →
		//      reported. ----
		{
			Code: `<h1>{((undefined))}</h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Spread-only heading (non-literal spread is opaque). ----
		{
			Code: `<h1 {...this.props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Multiple headings in one file: each is independently validated. ----
		{
			Code: `<><h1 /><h2>x</h2><h3 /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 3},
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 19},
			},
		},

		// ---- Headings inside conditional expressions are still visited. ----
		{
			Code: `cond ? <h1 /> : <h1>x</h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 8},
			},
		},

		// ---- Multiple `{undefined}` siblings — neither contributes content. ----
		{
			Code: `<h1>{undefined}{undefined}</h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Spread-strict alignment with upstream `hasAnyProp` defaults
		// ============================================================
		// The hasAccessibleChild fallback for `dangerouslySetInnerHTML` /
		// `children` uses `hasAnyProp` (default `spreadStrict: true`),
		// so even literal ObjectLiteral spreads are opaque — the prop
		// must appear as a DIRECT JsxAttribute.

		// ---- children in literal spread → opaque → INVALID ----
		{
			Code: `<h1 {...{children: 'x'}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- dangerouslySetInnerHTML in literal spread → opaque → INVALID ----
		{
			Code: `<h1 {...{dangerouslySetInnerHTML: {__html: "x"}}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- shorthand form `{...{children}}` is also opaque under strict ----
		{
			Code: `<h1 {...{children}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Multi-line / position-precision invalid cases
		// ============================================================

		// ---- Multi-line file with several empty headings on different
		//      lines. Locks in column accuracy with leading whitespace
		//      (4-space indent) and per-line independent reporting. ----
		{
			Code: "<>\n    <h1 />\n    <h2 />\n    <h3>OK</h3>\n    <h4 />\n</>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 2, Column: 5},
				{MessageId: "headingHasContent", Message: errorMessage, Line: 3, Column: 5},
				{MessageId: "headingHasContent", Message: errorMessage, Line: 5, Column: 5},
			},
		},

		// ============================================================
		// isHiddenFromScreenReader on the heading itself — false-positive
		// guards. aria-hidden values that are NOT statically true must NOT
		// exempt the heading.
		// ============================================================

		// ---- aria-hidden="false" string → coerced to boolean false →
		//      heading NOT exempt → empty content reported. ----
		{
			Code: `<h1 aria-hidden="false" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- aria-hidden="" empty string → upstream getPropValue → ""
		//      → "" === true is false → NOT exempt → reported. Locks in
		//      that string truthiness is NOT consulted; only strict `=== true`. ----
		{
			Code: `<h1 aria-hidden=""></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- aria-hidden={false} JSX-expression boolean false → not
		//      exempt → reported. ----
		{
			Code: `<h1 aria-hidden={false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- aria-hidden={someVar} non-static → not statically true →
		//      not exempt → reported. ----
		{
			Code: `<h1 aria-hidden={someVar} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- aria-hidden={1} truthy non-boolean → upstream's `=== true`
		//      strict equality fails → not exempt → reported. Locks in
		//      strict-equality vs JS-truthy distinction. ----
		{
			Code: `<h1 aria-hidden={1} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ============================================================
		// Real-world misuse — high-value invalid cases observed in audits
		// ============================================================

		// ---- Heading wrapping ONLY an aria-hidden icon — common
		//      icon-only heading misuse. ----
		{
			Code: `<h1><Icon aria-hidden="true" /></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<h2><svg aria-hidden="true"><path /></svg></h2>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Empty heading with className / id / data-* / role only —
		//      none of these contribute accessibility text. ----
		{
			Code: `<h1 className="hero" id="page-title" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<h1 data-testid="title" role="heading" aria-level={1} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- aria-labelledby / aria-describedby (these point AT external
		//      content but don't satisfy this rule — distinct from
		//      anchor-has-content's title/aria-label exit). ----
		{
			Code: `<h1 aria-describedby="desc" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Heading rendered through component-options remap WITHOUT
		//      content. Mirrors a common refactor mistake. ----
		{
			Code:    `<MyHeader />`,
			Tsx:     true,
			Options: map[string]interface{}{"components": []interface{}{"MyHeader"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Multiple bare headings under one Fragment in a render —
		//      each gets reported independently. Locks in stateless listener. ----
		{
			Code: `<><h1 /><h2 /><h3 /><h4 /><h5 /><h6 /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 3},
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 9},
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 15},
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 21},
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 27},
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 33},
			},
		},

		// ---- Inside an iife / arrow render — listener still visits ----
		{
			Code: `() => <h1 />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `function Page() { return <h1 />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 26},
			},
		},

		// ---- Inside a class render method — listener still visits ----
		{
			Code: `class C { render() { return <h1 />; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 29},
			},
		},

		// ---- Heading with ONLY explicit `undefined` via TS-stripped paren
		//      chain. Pure parens are stripped → reaches Identifier
		//      `undefined` → reported. ----
		{
			Code: `<h1>{((((undefined))))}</h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
	})
}

// TestHeadingHasContentOptionParsing exercises the JSON path that real CLI /
// JS configs take through `utils.GetOptionsMap` (vs. typed-struct shortcuts).
// Per the SKILL guidance, options coverage MUST include cases where Options
// is a bare map / array-wrapped map.
func TestHeadingHasContentOptionParsing(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &HeadingHasContentRule, []rule_tester.ValidTestCase{
		// Single-option CLI shape — bare map.
		{Code: `<Heading>x</Heading>`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{"Heading"}}},
		// rule_tester / multi-element shape — array-wrapped map.
		{Code: `<Heading>x</Heading>`, Tsx: true, Options: []interface{}{map[string]interface{}{"components": []interface{}{"Heading"}}}},
		// Empty / nil options must not crash and must default to h1–h6 only.
		{Code: `<Custom />`, Tsx: true},
		{Code: `<Custom />`, Tsx: true, Options: map[string]interface{}{}},
		// Empty components array — same as defaults.
		{Code: `<Custom />`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{}}},
		// Components with non-string entries are silently skipped.
		{Code: `<Heading>x</Heading>`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{"Heading", 42, true}}},
		// Malformed components (string instead of array) — must not crash;
		// falls back to defaults.
		{Code: `<Heading />`, Tsx: true, Options: map[string]interface{}{"components": "Heading"}},
		// Unknown options keys are ignored.
		{Code: `<Heading />`, Tsx: true, Options: map[string]interface{}{"unknown": true, "components": []interface{}{}}},
		// Components includes "h1" (duplicate) — typeCheck has duplicates but
		// slices.Contains short-circuits on first match.
		{Code: `<h1>x</h1>`, Tsx: true, Options: map[string]interface{}{"components": []interface{}{"h1"}}},
	}, []rule_tester.InvalidTestCase{
		// Bare-map option shape produces a report on the matched custom tag.
		{
			Code:    `<Heading />`,
			Tsx:     true,
			Options: map[string]interface{}{"components": []interface{}{"Heading"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Array-wrapped shape — reports on a custom matched tag.
		{
			Code:    `<Title />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"components": []interface{}{"Title"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "headingHasContent", Message: errorMessage, Line: 1, Column: 1},
			},
		},
	})
}
