package html_has_lang

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicSettings exercises the `polymorphicPropName` settings entry —
// `<Foo as="html">` resolves to nodeType "html" and falls through to the lang
// check. Upstream's own test file doesn't cover this path for html-has-lang.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// TestHtmlHasLangExtras locks in branches that upstream's test file doesn't
// exercise but are reachable through the rule's listener gate. Each case
// carries an inline comment pointing at the specific upstream branch /
// Dimension 4 edge shape it covers. These tests protect against silent
// regressions during refactors of either the rule or its shared helpers.
func TestHtmlHasLangExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &HtmlHasLangRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: paired form `<html>...</html>`. The listener
		//      fires on KindJsxOpeningElement once for the opening tag;
		//      lang is truthy → no report. Locks in that the paired form
		//      is wired up alongside self-closing. ----
		{Code: `<html lang="en"><body /></html>`, Tsx: true},
		{Code: `<html lang="en"></html>`, Tsx: true},

		// ---- Dimension 4: tag-name forms — case-sensitive HTML matching ----
		// Upstream's `type !== 'html'` check is case-sensitive. Capitalized
		// or differently-cased tag names are NOT html for this rule.
		{Code: `<Html />`, Tsx: true},
		{Code: `<HtMl />`, Tsx: true},

		// ---- Dimension 4: namespaced JSX names — type "svg:html" is truthy
		//      and != "html" → SKIP. ----
		{Code: `<svg:html />`, Tsx: true},

		// ---- Dimension 4: member-access tag — type "Foo.html" → SKIP. ----
		{Code: `<Foo.html />`, Tsx: true},

		// ---- Dimension 4: case-insensitive ATTRIBUTE name (jsx-ast-utils
		//      `getProp` default `ignoreCase: true`). Lock in every case
		//      variant of the lang attribute. ----
		{Code: `<html LANG="en" />`, Tsx: true},
		{Code: `<html Lang="en" />`, Tsx: true},
		{Code: `<html lAnG="en" />`, Tsx: true},

		// ---- Other HTML tags must NOT be reported — the rule only applies
		//      to <html> itself, regardless of accessibility relevance. ----
		{Code: `<head />`, Tsx: true},
		{Code: `<body />`, Tsx: true},
		{Code: `<div lang="en" />`, Tsx: true},
		{Code: `<span />`, Tsx: true},

		// ---- Locks in upstream `getPropValue` extract path — every shape
		//      that returns a truthy value via staticEval. ----
		// String literal with hyphenated locale.
		{Code: `<html lang="zh-Hant-TW" />`, Tsx: true},
		// Identifier (any non-undefined name) → returns the bare name → truthy.
		{Code: `<html lang={lang} />`, Tsx: true},
		{Code: `<html lang={i18n} />`, Tsx: true},
		// Member access → upstream synthesizes "obj.prop" → truthy.
		{Code: `<html lang={i18n.locale} />`, Tsx: true},
		{Code: `<html lang={config.lang.value} />`, Tsx: true},
		// Element access → truthy.
		{Code: `<html lang={config["lang"]} />`, Tsx: true},
		// Call expression → truthy.
		{Code: `<html lang={getLang()} />`, Tsx: true},
		{Code: `<html lang={i18n.getLang()} />`, Tsx: true},
		// Optional chaining (member / call) — same kinds in tsgo, optional flag.
		{Code: `<html lang={i18n?.locale} />`, Tsx: true},
		{Code: `<html lang={getLang?.()} />`, Tsx: true},
		// Conditional → recurse into the truthy branch.
		{Code: `<html lang={cond ? "en" : "de"} />`, Tsx: true},
		// Logical `||` — falls through to the right side, which is truthy.
		{Code: `<html lang={lang || "en"} />`, Tsx: true},
		// Logical `??` — falls through to the right side.
		{Code: `<html lang={lang ?? "en"} />`, Tsx: true},
		// Logical `&&` — short-circuits when left is truthy and returns the right.
		{Code: `<html lang={cond && "en"} />`, Tsx: true},
		// Template with substitution → synthesized non-empty string → truthy.
		{Code: "<html lang={`en-${region}`} />", Tsx: true},
		// No-substitution template → string literal value.
		{Code: "<html lang={`en-US`} />", Tsx: true},
		// Numeric literal — non-zero is truthy. (Atypical for lang but the
		// path lands on the truthy branch of staticEval.)
		{Code: `<html lang={1} />`, Tsx: true},
		// `new` expression → upstream returns truthy.
		{Code: `<html lang={new String("en")} />`, Tsx: true},

		// ---- Dimension 4: TS-only expression wrappers on the lang value.
		//      jsx-ast-utils' `extract` while-loops past TSAsExpression and
		//      TSNonNullExpression — staticEval mirrors via skipTransparent,
		//      so wrapper-stripping happens transparently. `satisfies` is
		//      INTENTIONALLY excluded from skipTransparent (jsx-ast-utils
		//      has no TYPES entry for it → null), see the INVALID block
		//      below. ----
		{Code: `<html lang={"en" as string} />`, Tsx: true},
		{Code: `<html lang={lang!} />`, Tsx: true},
		{Code: `<html lang={(lang) as any} />`, Tsx: true},

		// ---- Dimension 4: parenthesized expressions on lang value
		//      (single + multi-level). tsgo preserves Parens; ESTree
		//      flattens them. SkipParentheses inside attributeInnerExpression
		//      handles both. ----
		{Code: `<html lang={("en")} />`, Tsx: true},
		{Code: `<html lang={(("en"))} />`, Tsx: true},

		// ---- Locks in upstream `getProp`'s LITERAL spread walk: a literal
		//      ObjectLiteral spread carrying `lang` matches as if the prop
		//      were declared directly. tsgo preserves parens around the
		//      object literal; FindAttributeByName strips them. ----
		{Code: `<html {...{lang: "en"}} />`, Tsx: true},
		{Code: `<html {...{ lang }} />`, Tsx: true},

		// ---- Polymorphic-prop branch via `getElementType`:
		//      `<Foo as="html" lang="en" />` resolves to nodeType "html",
		//      lang truthy → no report. ----
		{
			Code:     `<Foo as="html" lang="en" />`,
			Tsx:      true,
			Settings: polymorphicSettings,
		},
		// Polymorphic with non-html target — `<Foo as="div" />` is treated
		// as a div by getElementType → not matched.
		{
			Code:     `<Foo as="div" />`,
			Tsx:      true,
			Settings: polymorphicSettings,
		},

		// ---- componentMap with literal-spread + lang inside the spread.
		//      Locks in the joint behaviour of GetElementType + literal-spread
		//      walk in FindAttributeByName. ----
		{
			Code:     `<HTMLTop {...{lang: "en"}} />`,
			Tsx:      true,
			Settings: htmlTopSettings,
		},

		// ---- Multi-line opening tag formatting must still resolve. ----
		{Code: "<html\n  lang=\"en\"\n/>", Tsx: true},
		{Code: "<html\n  lang=\"en\"\n>\n</html>", Tsx: true},

		// ---- Mixed siblings: only the html element is checked, surrounding
		//      JSX is not relevant. ----
		{Code: `<><div /><html lang="en" /></>`, Tsx: true},

		// ---- Settings-driven disable: remap "html" → "div" silences the
		//      rule for the literal `<html />` tag. Locks in the listener
		//      gate's "skipped when nodeType is not html" path through the
		//      componentMap. ----
		{
			Code:     `<html />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"components": map[string]interface{}{"html": "div"}}},
		},

		// ---- Other props on html besides lang. The rule should only inspect
		//      lang; arbitrary attrs don't change the outcome. ----
		{Code: `<html lang="en" className="root" />`, Tsx: true},
		{Code: `<html className="root" lang="en" data-x={1} />`, Tsx: true},

		// ---- Spread + explicit lang. Spread is opaque under upstream
		//      `getProp` for non-literal arguments, BUT the explicit lang
		//      attribute is still found and is truthy → no report. ----
		{Code: `<html {...props} lang="en" />`, Tsx: true},
		// Reverse order — explicit lang BEFORE spread. Same outcome.
		{Code: `<html lang="en" {...props} />`, Tsx: true},

		// ---- Dimension 2: nested html — outer has lang, inner has lang.
		//      Listener fires twice; both pass independently. Locks in that
		//      neither check bleeds across the JsxElement boundary. (Two
		//      <html>s in one tree is invalid HTML, but the AST listener
		//      must remain stateless and check each independently.) ----
		{Code: `<html lang="en"><html lang="de" /></html>`, Tsx: true},

		// ---- Dimension 2: html inside JsxExpression child of another
		//      element. Listener walks the entire tree; the inner html
		//      fires through KindJsxSelfClosingElement. ----
		{Code: `<div>{<html lang="en" />}</div>`, Tsx: true},

		// ---- Dimension 2: html inside an attribute value (arrow body /
		//      render prop). Listener still walks deeply. ----
		{Code: `<X render={() => <html lang="en" />} />`, Tsx: true},
		{Code: `<X children={<html lang="en" />} />`, Tsx: true},

		// ---- Dimension 2: html inside JsxFragment, both have lang. ----
		{Code: `<><html lang="en" /><html lang="de" /></>`, Tsx: true},

		// ============================================================
		// Real-world layout patterns (Next.js / Remix / Gatsby / SSR)
		// ============================================================

		// ---- Next.js 13+ App Router root layout: most common pattern.
		//      JsxOpeningElement fires once on `<html lang="en">`. ----
		{Code: `<html lang="en"><body>{children}</body></html>`, Tsx: true},
		{Code: `<html lang="en"><head /><body>{children}</body></html>`, Tsx: true},

		// ---- Next.js with i18n routing: lang from route params. ----
		{Code: `<html lang={locale}><body>{children}</body></html>`, Tsx: true},
		{Code: `<html lang={params.locale}><body /></html>`, Tsx: true},
		{Code: `<html lang={params?.locale}><body /></html>`, Tsx: true},

		// ---- Common i18n provider patterns: upstream getPropValue
		//      returns truthy for member access / call / template. ----
		{Code: `<html lang={i18n.language}><body /></html>`, Tsx: true},
		{Code: `<html lang={i18n.resolvedLanguage}><body /></html>`, Tsx: true},
		{Code: `<html lang={t("htmlLang")}><body /></html>`, Tsx: true},
		{Code: `<html lang={getLocale(req)}><body /></html>`, Tsx: true},
		{Code: "<html lang={`${locale}-${region}`}><body /></html>", Tsx: true},

		// ---- Common fallback patterns. staticEval recurses into the
		//      truthy branch — and Identifiers are non-empty strings → truthy. ----
		{Code: `<html lang={locale ?? "en"}><body /></html>`, Tsx: true},
		{Code: `<html lang={locale || "en"}><body /></html>`, Tsx: true},
		{Code: `<html lang={locale && "en"}><body /></html>`, Tsx: true},

		// ---- Remix / Gatsby DOCTYPE-adjacent patterns: html inside a
		//      Fragment alongside other top-level nodes. ----
		{Code: `<><div>{/*doctype*/}</div><html lang="en"><body /></html></>`, Tsx: true},

		// ---- Class component returning html — listener walks render(). ----
		{Code: `class Root { render() { return <html lang="en"><body /></html>; } }`, Tsx: true},

		// ---- Functional component returning html. ----
		{Code: `function Root({ children }) { return <html lang="en">{children}</html>; }`, Tsx: true},

		// ---- Arrow component with implicit return. ----
		{Code: `const Root = ({ children }) => <html lang="en">{children}</html>;`, Tsx: true},

		// ---- Generic component / TS-typed wrapper still fires the
		//      inner html listener correctly. ----
		{Code: `const Root: React.FC = ({ children }) => <html lang="en">{children}</html>;`, Tsx: true},

		// ---- React.lazy / Suspense / ErrorBoundary wrapping html.
		//      None of those are in the JSX render path of the listener;
		//      they're regular CallExpressions. The html inside still fires. ----
		{Code: `<Suspense fallback={null}><html lang="en"><body /></html></Suspense>`, Tsx: true},

		// ---- Layout extracted into reusable component with lang prop. ----
		{Code: `function HtmlRoot({ lang, children }) { return <html lang={lang}>{children}</html>; }`, Tsx: true},

		// ---- HoC pattern: html is inside the wrapped component. ----
		{Code: `withRouter(({ children }) => <html lang="en">{children}</html>)`, Tsx: true},

		// ---- Multi-line opening tag with several props before lang. ----
		{Code: "<html\n  className=\"root\"\n  data-theme={theme}\n  lang={locale}\n>\n  <body />\n</html>", Tsx: true},

		// ---- Common Tailwind / class name patterns. ----
		{Code: `<html lang="en" className="dark antialiased"><body /></html>`, Tsx: true},
		{Code: `<html lang="en" className={cn("dark", "antialiased")}><body /></html>`, Tsx: true},

		// ============================================================
		// AST-shape edge cases (tsgo↔ESTree differences locked in)
		// ============================================================

		// ---- BigInt lang (atypical but legal): non-zero bigint → truthy. ----
		{Code: `<html lang={1n} />`, Tsx: true},

		// ---- Function reference / arrow as lang value. Both are
		//      synthesized as truthy by staticEval (function values). ----
		{Code: `<html lang={getLang} />`, Tsx: true},
		{Code: `<html lang={() => "en"} />`, Tsx: true},
		{Code: `<html lang={function() { return "en"; }} />`, Tsx: true},

		// ---- Object / Array literal as lang value: ObjectLiteralExpression
		//      / ArrayLiteralExpression → truthy in staticEval. ----
		{Code: `<html lang={{ value: "en" }} />`, Tsx: true},
		{Code: `<html lang={["en"]} />`, Tsx: true},

		// ---- JSX as lang value (uncommon but legal in TSX). JsxElement
		//      / JsxSelfClosingElement → truthy in staticEval. ----
		{Code: `<html lang={<span>en</span>} />`, Tsx: true},
		{Code: `<html lang={<></>} />`, Tsx: true},

		// ---- Tagged template / RegExp as lang value: both truthy. ----
		{Code: "<html lang={tag`en`} />", Tsx: true},
		{Code: `<html lang={/en/} />`, Tsx: true},

		// ---- Class expression as lang value: truthy. ----
		{Code: `<html lang={class {}} />`, Tsx: true},

		// ---- Postfix increment of a non-numeric identifier returns NaN
		//      via staticEval, but the value is still computed before the
		//      increment — for an Identifier the staticEval result of `x`
		//      is "x" (truthy). However `x++` evaluates to NaN (falsy),
		//      see invalid block. ----

		// ---- Settings: empty components / polymorphicAllowList — no
		//      remap, behaves as if not set. ----
		{
			Code: `<html lang="en" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"components": map[string]interface{}{}},
			},
		},
		{
			Code: `<html lang="en" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{},
				},
			},
		},

		// ---- Settings: multiple unrelated component remaps. Only the
		//      one matching `<html>`-equivalent matters. ----
		{
			Code: `<HtmlRoot lang="en"><body /></HtmlRoot>`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"Foo":      "div",
						"HtmlRoot": "html",
						"Bar":      "span",
					},
				},
			},
		},

		// ---- Settings: components key remaps html → div (effectively
		//      disables the rule for native html tags). ----
		{
			Code: `<html />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"html": "div"},
				},
			},
		},

		// ---- polymorphicAllowList includes the rawType — `<Foo as="html">`
		//      gets remapped to "html" and lang="en" → no report. ----
		{
			Code: `<Foo as="html" lang="en" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Foo"},
				},
			},
		},

		// ---- polymorphicAllowList does NOT include rawType — `<Bar as="html">`
		//      keeps rawType "Bar", which is truthy and != "html" → SKIP. ----
		{
			Code: `<Bar as="html" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Foo"},
				},
			},
		},

		// ---- Tab whitespace inside the opening tag must still resolve. ----
		{Code: "<html\tlang=\"en\"\t/>", Tsx: true},

		// ---- CRLF line endings inside multi-line opening tag. ----
		{Code: "<html\r\n  lang=\"en\"\r\n/>", Tsx: true},

		// ---- JSX comment inside opening tag (uncommon; expression-only
		//      comments). The lang attribute is still detected. ----
		{Code: "<html /* before */ lang=\"en\" /* after */ />", Tsx: true},

		// ---- Templates with NON-literal substitutions remain truthy
		//      under upstream's `{TypeName}` / `{name}` placeholder logic.
		//      Identifier substitutions: upstream `{name}` (single curly).
		//      Other expression substitutions: upstream `{ESTreeTypeName}`.
		//      All are non-empty → truthy → no report. ----
		{Code: "<html lang={`${someName}`} />", Tsx: true},
		{Code: "<html lang={`${getLang()}`} />", Tsx: true},
		{Code: "<html lang={`${i18n.locale}`} />", Tsx: true},
		{Code: "<html lang={`${cond ? \"en\" : \"de\"}`} />", Tsx: true},
		{Code: "<html lang={`${a + b}`} />", Tsx: true},
		// `${undefined}` produces literally "undefined" upstream (the
		// special-cased Identifier name), which is a non-empty truthy
		// string → no report. Aligned with upstream.
		{Code: "<html lang={`${undefined}`} />", Tsx: true},
		// Mixed quasi + Identifier substitution → non-empty (head + "{x}").
		{Code: "<html lang={`prefix-${x}`} />", Tsx: true},
		// Mixed quasi + Literal substitution → quasi text alone → truthy
		// because the head/tail provide non-empty content even though
		// the literal substitution contributes "".
		{Code: "<html lang={`prefix-${\"suffix\"}`} />", Tsx: true},

		// ---- Sequence-comma operator inside JsxExpression. Upstream's
		//      SequenceExpression extractor returns an array of all
		//      values (truthy regardless of contents). Mine treats the
		//      comma operator (KindCommaToken) as truthy directly. Both
		//      truthy → no report. Locks in alignment for `(a, b)`. ----
		{Code: `<html lang={(0, "en")} />`, Tsx: true},
		{Code: `<html lang={(false, true)} />`, Tsx: true},
		{Code: `<html lang={(undefined, undefined)} />`, Tsx: true}, // upstream: array → truthy

		// ---- Bang / not operators on lang. Aligned with upstream. ----
		{Code: `<html lang={!""} />`, Tsx: true},   // !"" = true → truthy
		{Code: `<html lang={!!"x"} />`, Tsx: true}, // !!"x" = true → truthy

		// ---- Negative number (non-zero) is truthy. ----
		{Code: `<html lang={-1} />`, Tsx: true},

		// ---- Tilde (bitwise not) of zero → -1 → truthy. ----
		{Code: `<html lang={~0} />`, Tsx: true},

		// ---- delete operator returns boolean true → truthy. ----
		{Code: `<html lang={delete x.y} />`, Tsx: true},

		// ---- Literal Infinity → Number(+Infinity) → truthy. ----
		{Code: `<html lang={Infinity} />`, Tsx: true},
		// Negative Infinity → Number(-Infinity) → truthy.
		{Code: `<html lang={-Infinity} />`, Tsx: true},

		// ---- JS reserved global identifiers extracted as functions
		//      (truthy). ----
		{Code: `<html lang={Array} />`, Tsx: true},
		{Code: `<html lang={String} />`, Tsx: true},
		{Code: `<html lang={Object} />`, Tsx: true},

		// ---- `this` keyword → upstream synthesizes "this" string → truthy. ----
		{Code: `<html lang={this} />`, Tsx: true},

		// ============================================================
		// Production-code patterns (React 18+ / Next.js 13+ / Remix /
		// React Server Components / i18next / react-intl)
		// ============================================================

		// ---- React Server Component with await on locale loader. ----
		{Code: `async function RootLayout({ children }) { const locale = await getLocale(); return <html lang={locale}>{children}</html>; }`, Tsx: true},
		// ---- Async layout with await member access in lang. ----
		{Code: `<html lang={(await loadConfig()).lang}>{children}</html>`, Tsx: true},
		// ---- Cookies-based locale (Next.js dynamic API). ----
		{Code: `<html lang={cookies().get("locale")?.value ?? "en"}>{children}</html>`, Tsx: true},
		// ---- Headers-based negotiation. ----
		{Code: `<html lang={headers().get("accept-language")?.split(",")[0] ?? "en"}>{children}</html>`, Tsx: true},
		// ---- searchParams-based locale. ----
		{Code: `<html lang={searchParams.locale ?? "en"}>{children}</html>`, Tsx: true},
		// ---- env-driven default lang. ----
		{Code: `<html lang={process.env.DEFAULT_LOCALE ?? "en"}>{children}</html>`, Tsx: true},

		// ---- React Redux connect-style HOC wrapping html layout. ----
		{Code: `connect(state => ({lang: state.locale}))(({lang, children}) => <html lang={lang}>{children}</html>)`, Tsx: true},
		// ---- forwardRef wrapping html. ----
		{Code: `forwardRef((props, ref) => <html lang={props.lang} ref={ref}>{props.children}</html>)`, Tsx: true},
		// ---- memo wrapping html. ----
		{Code: `memo(({ lang, children }) => <html lang={lang}>{children}</html>)`, Tsx: true},
		// ---- React.memo with displayName-style assignment. ----
		{Code: `const Layout = React.memo(({ lang, children }) => <html lang={lang}>{children}</html>);`, Tsx: true},

		// ---- i18next patterns. ----
		{Code: `<html lang={i18n.language}>{children}</html>`, Tsx: true},
		{Code: `<html lang={i18n.resolvedLanguage}>{children}</html>`, Tsx: true},
		{Code: `<html lang={i18next.language}>{children}</html>`, Tsx: true},
		{Code: `<html lang={t("htmlLang")}>{children}</html>`, Tsx: true},

		// ---- react-intl patterns. ----
		{Code: `<html lang={intl.locale}>{children}</html>`, Tsx: true},
		{Code: `<html lang={useIntl().locale}>{children}</html>`, Tsx: true},

		// ---- Browser-aware locale fallback (SSR-safe pattern). ----
		{Code: `<html lang={typeof window !== "undefined" ? navigator.language : "en"}>{children}</html>`, Tsx: true},
		// ---- localStorage / Cookies fallback chain. ----
		{Code: `<html lang={localStorage.getItem("lang") ?? Cookies.get("lang") ?? "en"}>{children}</html>`, Tsx: true},
		// ---- Deep optional chain to user preferences. ----
		{Code: `<html lang={user?.preferences?.locale?.code ?? defaultLocale}>{children}</html>`, Tsx: true},
		// ---- User preferences with bracket notation. ----
		{Code: `<html lang={user?.["preferences"]?.["locale"] ?? "en"}>{children}</html>`, Tsx: true},

		// ---- BCP-47 tags with all subtag depths (lang allows them all
		//      since the rule only checks truthiness, not validity). ----
		{Code: `<html lang="en" />`, Tsx: true},
		{Code: `<html lang="en-US" />`, Tsx: true},
		{Code: `<html lang="en-GB-oed" />`, Tsx: true},                        // grandfathered
		{Code: `<html lang="zh-Hant-TW" />`, Tsx: true},                       // script + region
		{Code: `<html lang="cmn-Hans-CN" />`, Tsx: true},                      // macrolanguage
		{Code: `<html lang="sl-rozaj-biske-1994" />`, Tsx: true},              // multi-variant
		{Code: `<html lang="en-US-x-twain" />`, Tsx: true},                    // private use
		{Code: `<html lang="x-pseudo" />`, Tsx: true},                         // private-only
		{Code: `<html lang="i-default" />`, Tsx: true},                        // grandfathered
		{Code: `<html lang="art-x-pirate" />`, Tsx: true},                     // artificial + private

		// ---- Common HTML attribute combinations on root. ----
		{Code: `<html lang="en" dir="ltr" />`, Tsx: true},
		{Code: `<html lang="ar" dir="rtl" />`, Tsx: true},
		{Code: `<html lang="en" prefix="og: https://ogp.me/ns#" />`, Tsx: true},
		{Code: `<html lang="en" itemScope itemType="https://schema.org/WebSite" />`, Tsx: true},
		{Code: `<html lang="en" data-theme="dark" data-density="compact" />`, Tsx: true},
		{Code: `<html lang="en" suppressHydrationWarning />`, Tsx: true}, // common Next.js pattern

		// ---- String concat producing truthy lang. ----
		{Code: `<html lang={"en" + "-" + "US"} />`, Tsx: true},
		{Code: `<html lang={base + locale} />`, Tsx: true},
		{Code: `<html lang={"" + lang} />`, Tsx: true},

		// ---- Numeric edge cases (lang doesn't validate, just truthy). ----
		{Code: `<html lang={0o10} />`, Tsx: true}, // octal 8
		{Code: `<html lang={0xff} />`, Tsx: true}, // hex 255
		{Code: `<html lang={0b10} />`, Tsx: true}, // binary 2
		{Code: `<html lang={1e2} />`, Tsx: true},  // 100
		{Code: `<html lang={1_000} />`, Tsx: true}, // numeric separator
		{Code: `<html lang={1.5} />`, Tsx: true},  // float

		// ---- BigInt non-zero variants. ----
		{Code: `<html lang={123n} />`, Tsx: true},
		{Code: `<html lang={0xffn} />`, Tsx: true},

		// ---- Bracket-access patterns. ----
		{Code: `<html lang={config["lang"]} />`, Tsx: true},
		{Code: `<html lang={config[langKey]} />`, Tsx: true},
		{Code: `<html lang={langs[0]} />`, Tsx: true},

		// ---- Arrow IIFE returning truthy. ----
		{Code: `<html lang={(() => "en")()} />`, Tsx: true},
		{Code: `<html lang={(function () { return "en"; })()} />`, Tsx: true},

		// ---- Polymorphic propName variants — non-default names. ----
		{
			Code: `<Box component="html" lang="en" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "component"},
			},
		},
		{
			Code: `<Box tag="html" lang="en" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "tag"},
			},
		},

		// ---- Settings: polymorphicPropName + componentMap + allowList
		//      all three combined — exercises the full GetElementType
		//      resolution chain. ----
		{
			Code: `<HtmlBox as="custom-html" lang="en" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"HtmlBox"},
					"components":           map[string]interface{}{"custom-html": "html"},
				},
			},
		},

		// ---- Settings: malformed values must not crash and must
		//      gracefully fall back. ----
		// Non-string polymorphicPropName → ignored.
		{
			Code: `<html lang="en" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": 42},
			},
		},
		// Non-array polymorphicAllowList → treated as no-allowlist.
		{
			Code: `<Foo as="html" lang="en" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": "Foo", // wrong type
				},
			},
		},
		// Non-map components → ignored.
		{
			Code: `<html lang="en" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"components": "html"}, // wrong type
			},
		},

		// ---- Multi-key literal spread; "key" is excluded per upstream
		//      `getProp`. The rest of the keys are walked. ----
		{Code: `<html {...{key: someKey, lang: "en"}} />`, Tsx: true},
		// Reverse order — same result.
		{Code: `<html {...{lang: "en", key: someKey}} />`, Tsx: true},

		// ---- Multi-attribute mixing direct + literal-spread. ----
		{Code: `<html className="x" {...{lang: "en"}} />`, Tsx: true},

		// ---- String coerced "true"/"false" literal — case sensitivity
		//      via jsxAstUtilsLiteralCoerce. ----
		// "True" → coerced to true (boolean) → truthy.
		{Code: `<html lang="True" />`, Tsx: true},
		{Code: `<html lang="TRUE" />`, Tsx: true},
		{Code: `<html lang="tRuE" />`, Tsx: true},

		// ---- String literal with leading/trailing whitespace — JS
		//      truthy because non-empty (the rule doesn't trim). ----
		{Code: `<html lang=" en " />`, Tsx: true},
		{Code: `<html lang="\nen\n" />`, Tsx: true},

		// ---- Unicode-escape inside string literal — decoded to "en". ----
		{Code: `<html lang="en" />`, Tsx: true},
		{Code: `<html lang="en" />`, Tsx: true},

		// ---- Generic component with polymorphic html. ----
		{
			Code: `<Box<HTMLProps> as="html" lang="en" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": "as"},
			},
		},

		// ---- Nullish chain with Identifier left arm — staticEval for
		//      `a ?? b ?? ""` evaluates `a` as Identifier "a" → returns
		//      string "a" which is non-null/undef → returns left → "a"
		//      → truthy. Aligned with upstream's `??` semantics on
		//      Identifier (returns the bare name string). NO REPORT. ----
		{Code: `<html lang={a ?? b ?? ""} />`, Tsx: true},
		{Code: `<html lang={a ?? "" ?? "fallback"} />`, Tsx: true},
		// `||` chain with leading Identifier — same logic; "a" truthy → returns left.
		{Code: `<html lang={a || "" || false} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Paired form `<html></html>` with no lang — the listener fires
		//      on the JsxOpeningElement (column 1), not on the JsxElement
		//      wrapper. ----
		{
			Code: `<html></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Paired form with non-lang attributes only. ----
		{
			Code: `<html className="root"></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Locks in upstream `if (lang)` JS truthiness — every falsy
		//      static value reports. ----
		// Empty string → falsy.
		{
			Code: `<html lang="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Empty string via JsxExpression-wrapped string literal.
		{
			Code: `<html lang={""} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// null literal → falsy.
		{
			Code: `<html lang={null} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// boolean false → falsy.
		{
			Code: `<html lang={false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Numeric 0 → falsy. (Atypical for lang but locks in the truthiness
		// check for the staticEval Number → 0 path.)
		{
			Code: `<html lang={0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// String literal "false" — jsx-ast-utils' Literal extractor
		// case-insensitively coerces to boolean false → falsy. Locks in
		// the jsxAstUtilsLiteralCoerce path inside staticEval.
		{
			Code: `<html lang="false" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<html lang="FALSE" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Empty JsxExpression `{}` — tsgo synthesizes for malformed input.
		// attributeInnerExpression returns nil → PropValueIsTruthy false.
		{
			Code: `<html lang={} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// `void 0` → undefined → falsy.
		{
			Code: `<html lang={void 0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- TS-wrapped undefined still falsy — staticEval's
		//      skipTransparent strips `as` / `!` / parens before
		//      checking the inner Identifier. ----
		{
			Code: `<html lang={undefined as any} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<html lang={(undefined)} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- `satisfies` is intentionally opaque — jsx-ast-utils' TYPES
		//      table has no entry for TSSatisfiesExpression, so getPropValue
		//      returns null → falsy → REPORT. Locks in that satisfies is
		//      excluded from skipTransparent in staticEval. Note this is the
		//      ONLY TS expression wrapper that does NOT pass through; `as` /
		//      `!` / parens all unwrap. ----
		{
			Code: `<html lang={"en" satisfies string} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Position assertion (Dimension: source-position locking).
		//      Two-space indent shifts the report column to 3. End range
		//      covers the whole self-closing element `<html />` (8 chars,
		//      cols 3..11). EndColumn is the exclusive 1-based end. ----
		{
			Code: `  <html />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage,
					Line: 1, Column: 3, EndLine: 1, EndColumn: 11},
			},
		},

		// ---- Bare self-closing covers cols 1..8 (8 chars). Locks in the
		//      JsxSelfClosingElement node range. ----
		{
			Code: `<html />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
			},
		},

		// ---- Paired form: report lands on the JsxOpeningElement only —
		//      `<html>` (6 chars), NOT the wrapping JsxElement that
		//      includes `</html>`. EndColumn 7 = position after `>`. ----
		{
			Code: `<html></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			},
		},

		// ---- Multi-line position: the report covers the entire opening
		//      tag from `<` to `>` across multiple lines. Locks in the
		//      multi-line range so a future regression in node.End()
		//      computation (e.g. dropping trailing trivia handling) is
		//      caught immediately. ----
		{
			Code: "<html\n  className=\"root\"\n>\n</html>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage,
					Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
			},
		},

		// ---- Polymorphic-prop branch invalid case ----
		{
			Code:     `<Foo as="html" />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code:     `<Foo as="html" lang={undefined} />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Listener is stateless: each <html> in the same file is
		//      independently validated. ----
		{
			Code: `<><html /><html lang="en" /><html /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 3},
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 29},
			},
		},

		// ---- Conditional rendering still visits both branches. ----
		{
			Code: `cond ? <html /> : <html lang="en" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 8},
			},
		},

		// ---- Spread that LITERALLY contains lang={undefined} — getProp
		//      walks the literal spread and finds the property. The value
		//      is the bare Identifier `undefined` → falsy → REPORT.
		//      Locks in literal-spread walking + Identifier-undefined
		//      coordination across FindAttributeByName + PropValueIsTruthy. ----
		{
			Code: `<html {...{lang: undefined}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Spread carrying lang="" inside literal spread — empty string
		//      via the literal spread walk. ----
		{
			Code: `<html {...{lang: ""}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- componentMap remap + spread → REPORT (spread is opaque). ----
		{
			Code:     `<HTMLTop {...props} />`,
			Tsx:      true,
			Settings: htmlTopSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Dimension 2: nested html — outer no lang, inner no lang.
		//      Both fire independently → 2 reports. Locks in stateless
		//      listener (no parent-state inheritance). ----
		{
			Code: `<html><html /></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Dimension 2: nested html — outer has lang, inner no lang.
		//      Only the inner reports. Locks in that the outer's truthy
		//      lang doesn't suppress the inner check. ----
		{
			Code: `<html lang="en"><html /></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 17},
			},
		},

		// ---- Dimension 2: nested html — outer no lang, inner has lang.
		//      Only the outer reports. Symmetric counterpart of the above. ----
		{
			Code: `<html><html lang="en" /></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Dimension 2: html inside a JsxExpression child of another
		//      element — listener still fires on the inner html. Locks in
		//      that the rule isn't gated on direct-root-of-tree position. ----
		{
			Code: `<div>{<html />}</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Dimension 2: html inside an arrow body that is itself the
		//      value of a JsxAttribute (render-prop pattern). The inner
		//      html still fires. ----
		{
			Code: `<X render={() => <html />} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 18},
			},
		},

		// ---- Dimension 2: html inside JsxFragment — listener fires
		//      independently on each fragment child. ----
		{
			Code: `<><html /><html /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 3},
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 11},
			},
		},

		// ============================================================
		// Literal-key spread: upstream `getProp` has a STRICT
		// `property.key.type === 'Identifier'` guard. String-literal,
		// numeric-literal, and computed-literal keys all return
		// undefined → REPORT. These lock in alignment with upstream's
		// jsx-ast-utils v3+ `getProp.js` (verified: src/getProp.js).
		// ============================================================

		// ---- String-literal key in spread → upstream undefined → REPORT. ----
		{
			Code: `<html {...{"lang": "en"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Single-quote string-literal key (legal TS object literal).
		//      Same guard. ----
		{
			Code: `<html {...{'lang': 'en'}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Computed key (literal expression inside `[...]`). Even
		//      when the inner expression is a StringLiteral matching
		//      "lang", upstream's strict Identifier guard misses it. ----
		{
			Code: `<html {...{["lang"]: "en"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Numeric-literal key (uncommon; matches upstream's
		//      Identifier-only guard). The numeric "0" doesn't match
		//      "lang" anyway, but lock in the AST-shape rejection. ----
		{
			Code: `<html {...{0: "en"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// AST-shape edge cases — falsy lang values via various
		// expression shapes that upstream getPropValue evaluates.
		// ============================================================

		// ---- `null ?? ""` → empty string → falsy → REPORT. Locks in
		//      the staticEval `??` short-circuit reaching the empty-string
		//      right operand. ----
		{
			Code: `<html lang={null ?? ""} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Conditional with both branches falsy → falsy. ----
		{
			Code: `<html lang={cond ? "" : null} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- `false || ""` → empty string → falsy → REPORT. Locks in
		//      the `||` short-circuit reaching empty string. ----
		{
			Code: `<html lang={false || ""} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- typeof / void are upstream-mapped to undefined / falsy.
		//      Locks in the staticEval typeof/void branches. ----
		{
			Code: `<html lang={typeof x} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Postfix increment on a non-numeric Identifier evaluates
		//      to NaN (number, falsy). Locks in staticEvalUpdate. ----
		{
			Code: `<html lang={x++} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Negative zero is falsy in JS. ----
		{
			Code: `<html lang={-0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Sequence-comma in a JsxExpression. tsgo collapses comma
		//      into BinaryExpression, which staticEval treats as truthy
		//      (matches ESTree's SequenceExpression extractor). So this
		//      is actually VALID upstream — but if the leftmost is the
		//      issue, the comma case actually returns truthy.
		//      ----
		// (Skipping comma-sequence case — it's truthy in upstream
		// ExpressionExtractor's SequenceExpression handler.)

		// ============================================================
		// Real-world bug patterns
		// ============================================================

		// ---- Common bug: forgot to provide a default fallback. ----
		{
			Code: `<html lang={undefined}><body /></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Common bug: nullish locale with empty fallback. ----
		{
			Code: `<html lang={undefined ?? ""}><body /></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Common bug: dropped lang in a refactor / template copy. ----
		{
			Code: `<html className="dark" data-theme={theme}><body /></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Common bug: lang on inner element instead of outer html. ----
		{
			Code: `<html><body lang="en">{children}</body></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Common bug: dynamic lang resolved at runtime, but
		//      `getLanguage()` could return undefined. staticEval treats
		//      CallExpression as truthy (synthesized "callee(args)"
		//      string), so this is actually VALID upstream — locked in
		//      via the valid block above. ----

		// ---- Functional component returning <html /> (no lang). ----
		{
			Code: `function Layout() { return <html />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 28},
			},
		},

		// ---- Arrow component with implicit return of bare html. ----
		{
			Code: `const Layout = () => <html />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 22},
			},
		},

		// ---- Class component returning bare html. ----
		{
			Code: `class Layout { render() { return <html />; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 34},
			},
		},

		// ---- HoC wrapping a component that returns bare html. ----
		{
			Code: `withRouter(({ children }) => <html>{children}</html>)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 30},
			},
		},

		// ---- Async functional component returning bare html (RSC pattern). ----
		{
			Code: `async function Layout() { return <html />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 34},
			},
		},

		// ---- Tab-whitespace inside the opening tag — the html element
		//      still parses; lang missing → REPORT. ----
		{
			Code: "<html\t/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Templates with literal-only substitutions: upstream's
		// jsx-ast-utils TemplateLiteral.js extractor has no branch for
		// `Literal` AST type — it falls through to `return ''`. So
		// `\`${"en"}\`` extracts to `""` (empty string), which is JS-
		// falsy → REPORT. Same for numeric / boolean / null / regex /
		// JSX literals. These tests lock in alignment with upstream
		// after the staticEvalTemplate fix.
		// ============================================================
		// String-literal substitution → "" → falsy.
		{
			Code: "<html lang={`${\"en\"}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Numeric-literal substitution → "" → falsy.
		{
			Code: "<html lang={`${0}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Boolean-literal substitution → "" → falsy.
		{
			Code: "<html lang={`${false}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: "<html lang={`${true}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Null-literal substitution → "" → falsy.
		{
			Code: "<html lang={`${null}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// JSX-element substitution → ESTree JSXElement, no "Expression" → "".
		{
			Code: "<html lang={`${<x />}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// JSX-fragment substitution → ESTree JSXFragment, no "Expression" → "".
		{
			Code: "<html lang={`${<></>}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Empty template (no substitutions, no quasi text) → "" → falsy.
		{
			Code: "<html lang={``} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Template with empty head/tail and pure-literal sub:
		// head="" + "" + tail="" = "" → falsy.
		{
			Code: "<html lang={`${\"\"}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Spread argument variants — non-literal spreads are opaque to
		// upstream `getProp`. All shapes below are "non-literal" → the
		// spread is opaque → no `lang` prop found → REPORT.
		// ============================================================

		// Spread of CallExpression result.
		{
			Code: `<html {...spreadFn()} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Spread of MemberAccessExpression.
		{
			Code: `<html {...obj.spread} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Spread of NewExpression.
		{
			Code: `<html {...new Foo()} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Spread of ElementAccessExpression.
		{
			Code: `<html {...arr[0]} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Spread of ConditionalExpression.
		{
			Code: `<html {...(cond ? a : b)} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Spread of LogicalExpression.
		{
			Code: `<html {...(props || defaults)} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Spread of TS-wrapped object — TSAsExpression is opaque per
		// upstream's strict ObjectExpression check (no TS-wrapper unwrap
		// in `getProp`).
		{
			Code: `<html {...({lang: "en"} as any)} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<html {...({lang: "en"})!} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Numeric edge falsy literals — staticEval normalizes via
		// utils.NormalizeNumericLiteral; all of these resolve to 0.
		// ============================================================
		{
			Code: `<html lang={0o0} />`, // octal zero
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<html lang={0x0} />`, // hex zero
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<html lang={0b0} />`, // binary zero
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Float zero.
		{
			Code: `<html lang={0.0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// BigInt zero.
		{
			Code: `<html lang={0n} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Arithmetic / coercion paths producing falsy.
		// ============================================================
		// String concat producing empty.
		{
			Code: `<html lang={"" + ""} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Subtraction yielding zero.
		{
			Code: `<html lang={1 - 1} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Multiplication yielding zero.
		{
			Code: `<html lang={5 * 0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Division yielding zero.
		{
			Code: `<html lang={0 / 5} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Modulo yielding zero.
		{
			Code: `<html lang={6 % 3} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Comparison yielding boolean false.
		{
			Code: `<html lang={1 > 2} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<html lang={2 < 1} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// `!` of truthy.
		{
			Code: `<html lang={!"x"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<html lang={!1} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Negation yielding negative zero (which is falsy).
		{
			Code: `<html lang={-0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Production-code BUG patterns — common mistakes.
		// ============================================================

		// Forgot lang on Next.js root layout (very common).
		{
			Code: `function RootLayout({ children }) { return <html><body>{children}</body></html>; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 44},
			},
		},

		// Async layout without lang (RSC pattern).
		{
			Code: `export default async function Layout({ children }) { return <html>{children}</html>; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 61},
			},
		},

		// Lang typo'd as "language" instead of "lang" — Identifier-only
		// match means "language" doesn't help.
		{
			Code: `<html language="en" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Lang on body instead of html.
		{
			Code: `<html><body lang="en">{children}</body></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Lang in style attribute instead of lang prop (very wrong).
		{
			Code: `<html style={{lang: "en"}}>{children}</html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Listener stateless lock-ins — exhaustive nested combinations.
		// ============================================================

		// Three-deep nested html — 3 listeners fire independently.
		{
			Code: `<html><html><html /></html></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 7},
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 13},
			},
		},

		// Mixed 3-deep — middle has lang, inner+outer don't.
		{
			Code: `<html><html lang="en"><html /></html></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 23},
			},
		},

		// html generated via array map — listener walks all expression
		// children. Each generated <html /> reports independently.
		{
			Code: `<>{["en", "de", "fr"].map(l => <html />)}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 32},
			},
		},

		// html inside try/catch returning JSX.
		{
			Code: `function f() { try { return <html />; } catch { return null; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 29},
			},
		},

		// html inside switch case (uncommon but legal).
		{
			Code: `function f(k) { switch (k) { case 0: return <html />; default: return null; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 45},
			},
		},

		// html as default export expression.
		{
			Code: `export default () => <html />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 22},
			},
		},

		// html inside generator yield (legal in TSX even if unusual).
		{
			Code: `function* g() { yield <html />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 23},
			},
		},

		// html inside async generator yield.
		{
			Code: `async function* g() { yield <html />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 29},
			},
		},

		// html inside conditional with both branches missing lang.
		{
			Code: `cond ? <html /> : <html />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 8},
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 19},
			},
		},

		// html inside logical && and ||.
		{
			Code: `cond && <html />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 9},
			},
		},
		{
			Code: `cond || <html />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 9},
			},
		},

		// html inside array literal.
		{
			Code: `[<html />, <html lang="en" />, <html />]`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 2},
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 32},
			},
		},

		// html inside object literal property value.
		{
			Code: `({ root: <html /> })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 10},
			},
		},

		// ============================================================
		// Falsy chain cases — staticEval through nullish/logical chains
		// to a falsy terminal.
		// ============================================================

		// `||` chain reaching false at the end.
		{
			Code: `<html lang={"" || "" || false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// Conditional with constant-false test reaching falsy branch.
		{
			Code: `<html lang={true ? "" : "en"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<html lang={false ? "en" : ""} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// Bang of truthy.
		{
			Code: `<html lang={!1} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// String coerced to false via Literal.js — case-insensitive.
		{
			Code: `<html lang="False" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<html lang="fAlSe" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Three-way attribute equivalence classes (Dimension 4) — lang
		// MUST be matched as a JsxAttribute name. Keys in literal spread
		// object are matched only when Identifier (per upstream guard).
		// All non-Identifier-key forms below produce no match → REPORT.
		// ============================================================

		// Computed key with dynamic name expression — even if the name
		// resolves to "lang" at runtime, upstream's static guard rejects.
		{
			Code: `<html {...{[propName]: "en"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Computed key referencing a constant.
		{
			Code: `<html {...{[LANG_KEY]: "en"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Spread of nested literal — outer's properties contain
		// SpreadAssignment (`...{lang: "en"}`), not a PropertyAssignment
		// for "lang". Upstream's getProp doesn't recurse, so this is
		// opaque → REPORT.
		{
			Code: `<html {...{...{lang: "en"}}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "htmlHasLang", Message: errorMessage, Line: 1, Column: 1},
			},
		},
	})
}
