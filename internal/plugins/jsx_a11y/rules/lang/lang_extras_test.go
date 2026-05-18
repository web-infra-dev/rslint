package lang

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicAsSettings exercises the `polymorphicPropName` settings entry
// without the componentMap remap — `<Foo as="html">` resolves to nodeType
// "html" and falls through to the lang check. Upstream's own test file uses
// the combined settings (componentMap + polymorphic); this isolates the
// polymorphic-only path.
var polymorphicAsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// htmlRemapSettings remaps a custom component name to "html" without
// touching polymorphic settings. Locks in the componentMap-only path.
var htmlRemapSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"HTMLRoot": "html",
		},
	},
}

// TestLangExtras locks in branches that upstream's test file doesn't exercise
// but are reachable through the rule's listener gate. Each case carries an
// inline comment pointing at the specific upstream branch / Dimension 4 edge
// shape / BCP-47 validation case it covers. These tests protect against
// silent regressions during refactors of either the rule, its shared
// helpers, or the underlying language.Parse semantics.
func TestLangExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &LangRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Dimension 1 — AST shapes of the lang value (string-typed)
		// ============================================================

		// ---- Valid BCP-47 tags beyond the upstream suite: every common
		//      shape with language / script / region / variant / extension /
		//      private-use subtags. ----
		{Code: `<html lang="en-GB" />`, Tsx: true},
		{Code: `<html lang="fr-CA" />`, Tsx: true},
		{Code: `<html lang="de-DE" />`, Tsx: true},
		{Code: `<html lang="pt-BR" />`, Tsx: true},
		// Three-letter language subtag (ISO 639-2/3) — `yue` is Cantonese.
		{Code: `<html lang="yue" />`, Tsx: true},
		// Script subtag (`Cyrl` Cyrillic).
		{Code: `<html lang="sr-Cyrl" />`, Tsx: true},
		// Region subtag as UN M.49 numeric code (`419` = Latin America).
		{Code: `<html lang="es-419" />`, Tsx: true},
		// Variant subtag (`gascon` Gascon Occitan).
		{Code: `<html lang="oc-gascon" />`, Tsx: true},
		// Extension subtag — `u-co-phonebk` German phonebook collation.
		{Code: `<html lang="de-DE-u-co-phonebk" />`, Tsx: true},
		// Private-use subtag (`x-*`).
		// Private-use subtags must be 1-8 chars each; "comp" satisfies that.
		// `en-x-comp` has a primary language → valid (vs `x-private` which
		// is private-use-only and rejected; see invalid section below).
		{Code: `<html lang="en-x-comp" />`, Tsx: true},

		// ---- Case-insensitive ATTRIBUTE name (jsx-ast-utils `getProp`
		//      default `ignoreCase: true`). Locks in every case variant of
		//      the lang attribute. ----
		{Code: `<html LANG="en" />`, Tsx: true},
		{Code: `<html Lang="en" />`, Tsx: true},
		{Code: `<html lAnG="en" />`, Tsx: true},

		// ---- Locks in upstream `getLiteralPropValue` null path — every
		//      shape that returns null (= upstream `value === null` → SKIP).
		//      Includes Identifier-non-undefined, member, call, ternary, etc.
		//      None of these are statically resolvable as literals. ----
		{Code: `<html lang={lang} />`, Tsx: true},
		{Code: `<html lang={i18n} />`, Tsx: true},
		{Code: `<html lang={i18n.locale} />`, Tsx: true},
		{Code: `<html lang={config["lang"]} />`, Tsx: true},
		{Code: `<html lang={getLang()} />`, Tsx: true},
		{Code: `<html lang={i18n?.locale} />`, Tsx: true},
		{Code: `<html lang={getLang?.()} />`, Tsx: true},
		{Code: `<html lang={cond ? "en" : "de"} />`, Tsx: true},
		{Code: `<html lang={lang || "en"} />`, Tsx: true},
		{Code: `<html lang={lang ?? "en"} />`, Tsx: true},
		{Code: `<html lang={cond && "en"} />`, Tsx: true},

		// ---- TemplateExpression with substitutions — upstream synthesizes
		//      a placeholder string. Upstream's `tags.check("en-{x}")` returns
		//      false, so we'd expect REPORT in upstream. BUT note: upstream
		//      uses LITERAL_TYPES (getLiteralPropValue), which DOES extract
		//      template literals through TemplateLiteral.js. The result is a
		//      non-empty placeholder like `en-{region}` for substitution
		//      Identifiers. That fails BCP-47 → REPORT.
		//
		//      However, a NoSubstitutionTemplateLiteral like `\`en\`` is just
		//      the string "en" — valid BCP-47. Lock in that variant. ----
		{Code: "<html lang={`en`} />", Tsx: true},
		{Code: "<html lang={`en-US`} />", Tsx: true},
		{Code: "<html lang={`zh-Hans`} />", Tsx: true},

		// ---- String literal in JsxExpression — `<html lang={"en"} />`.
		//      Should be valid (same as direct StringLiteral). ----
		{Code: `<html lang={"en"} />`, Tsx: true},
		{Code: `<html lang={"en-US"} />`, Tsx: true},

		// ---- TS expression wrappers on the lang value. jsx-ast-utils'
		//      LITERAL_TYPES has NO entry for TSAsExpression, TSNonNull,
		//      SatisfiesExpression — they fall through to noop → null. So
		//      `getLiteralPropValue({"en" as string})` returns null,
		//      upstream `value === null` → SKIP. Mirror that. ----
		{Code: `<html lang={"en" as string} />`, Tsx: true},
		{Code: `<html lang={lang!} />`, Tsx: true},
		{Code: `<html lang={"foo" as any} />`, Tsx: true},
		// satisfies — TS5+ syntax.
		{Code: `<html lang={"en" satisfies string} />`, Tsx: true},

		// ---- Parenthesized expressions — tsgo preserves parens; ESTree
		//      flattens. literalPropValue strips parens. ----
		{Code: `<html lang={("en")} />`, Tsx: true},
		{Code: `<html lang={(("en"))} />`, Tsx: true},

		// ---- Literal-spread of an object with lang. FindAttributeByName
		//      walks ObjectLiteral spreads. ----
		{Code: `<html {...{lang: "en"}} />`, Tsx: true},

		// ---- Other HTML tags must NOT be checked — listener gate filters
		//      to "html" (or remapped) only. ----
		{Code: `<head lang="foo" />`, Tsx: true},
		{Code: `<body lang="foo" />`, Tsx: true},
		{Code: `<div lang="foo" />`, Tsx: true},
		{Code: `<span lang="zz-LL" />`, Tsx: true},

		// ---- Case-sensitive HTML matching — upstream's `type !== 'html'`
		//      is case-sensitive. Capitalized or differently-cased custom
		//      tag names short-circuit. ----
		{Code: `<Html lang="foo" />`, Tsx: true},
		{Code: `<HtMl lang="zz-LL" />`, Tsx: true},
		{Code: `<HTML lang="foo" />`, Tsx: true},

		// ---- Namespaced JSX names — type "svg:html" is truthy and !=
		//      "html" → SKIP. ----
		{Code: `<svg:html lang="foo" />`, Tsx: true},

		// ---- Member-access tag — type "Foo.html" → SKIP. ----
		{Code: `<Foo.html lang="zz-LL" />`, Tsx: true},

		// ============================================================
		// Dimension 2 — Nesting / traversal boundaries
		// ============================================================

		// ---- Paired form `<html>...</html>` with valid lang. ----
		{Code: `<html lang="en"><body /></html>`, Tsx: true},
		{Code: `<html lang="en"></html>`, Tsx: true},

		// ---- Multi-line opening tag formatting. ----
		{Code: "<html\n  lang=\"en\"\n/>", Tsx: true},
		{Code: "<html\n  lang=\"en\"\n>\n</html>", Tsx: true},

		// ---- html inside JsxFragment / sibling JSX. ----
		{Code: `<><div /><html lang="en" /></>`, Tsx: true},

		// ---- html inside JsxExpression child of another element. Listener
		//      walks the entire tree. ----
		{Code: `<div>{<html lang="en" />}</div>`, Tsx: true},

		// ---- html inside an attribute value (render prop). ----
		{Code: `<X render={() => <html lang="en" />} />`, Tsx: true},

		// ---- Nested html — both have lang. Listener fires twice; both
		//      pass independently. ----
		{Code: `<html lang="en"><html lang="de" /></html>`, Tsx: true},

		// ============================================================
		// Settings-driven branches (componentMap, polymorphic)
		// ============================================================

		// ---- componentMap remap with valid lang. ----
		{
			Code:     `<HTMLRoot lang="en" />`,
			Tsx:      true,
			Settings: htmlRemapSettings,
		},
		{
			Code:     `<HTMLRoot lang="zh-Hant-HK" />`,
			Tsx:      true,
			Settings: htmlRemapSettings,
		},

		// ---- polymorphicPropName-only — `<Foo as="html" lang="en" />`. ----
		{
			Code:     `<Foo as="html" lang="en" />`,
			Tsx:      true,
			Settings: polymorphicAsSettings,
		},
		// Polymorphic with non-html target — skipped via the type-gate.
		{
			Code:     `<Foo as="div" lang="foo" />`,
			Tsx:      true,
			Settings: polymorphicAsSettings,
		},

		// ---- componentMap remap targeting a DIFFERENT html-equivalent
		//      component, with valid lang. Multiple components in map. ----
		{
			Code: `<HtmlRoot lang="en" />`,
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

		// ---- polymorphicAllowList that INCLUDES the rawType — `<Foo as="html" lang="en">`
		//      gets remapped to "html", lang is valid → no report. ----
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

		// ---- polymorphicAllowList that does NOT include rawType — `<Bar as="html">`
		//      keeps rawType "Bar", which is truthy and != "html" → SKIP. ----
		{
			Code: `<Bar as="html" lang="foo" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Foo"},
				},
			},
		},

		// ---- Settings remap "html" → "div" — disables the rule for the
		//      literal `<html />` tag. ----
		{
			Code: `<html lang="foo" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"html": "div"},
				},
			},
		},

		// ============================================================
		// Real-world layout patterns
		// ============================================================

		// ---- Next.js 13+ App Router root layout. ----
		{Code: `<html lang="en"><body>{children}</body></html>`, Tsx: true},
		{Code: `<html lang="en-US"><head /><body>{children}</body></html>`, Tsx: true},

		// ---- i18n routing — lang from route params. Identifier / member /
		//      optional chain → unresolvable → skip. ----
		{Code: `<html lang={locale}><body>{children}</body></html>`, Tsx: true},
		{Code: `<html lang={params.locale}><body /></html>`, Tsx: true},
		{Code: `<html lang={i18n.language}><body /></html>`, Tsx: true},
		{Code: `<html lang={params?.locale}><body /></html>`, Tsx: true},
		{Code: `<html lang={t("htmlLang")}><body /></html>`, Tsx: true},

		// ---- Class component returning html. ----
		{Code: `class Root { render() { return <html lang="en"><body /></html>; } }`, Tsx: true},

		// ---- Functional component returning html. ----
		{Code: `function Root({ children }) { return <html lang="en">{children}</html>; }`, Tsx: true},

		// ---- Arrow component with implicit return. ----
		{Code: `const Root = ({ children }) => <html lang="en">{children}</html>;`, Tsx: true},

		// ============================================================
		// BCP-47 edge cases (locked into language.Parse acceptance)
		// ============================================================

		// ---- Uppercase tag — language.Parse normalizes case, accepts. ----
		{Code: `<html lang="EN" />`, Tsx: true},
		{Code: `<html lang="EN-US" />`, Tsx: true},
		// Mixed case — accepted.
		{Code: `<html lang="En-uS" />`, Tsx: true},

		// ---- Production tags from real i18n stacks. All accepted by
		//      both rslint and `language-tags` (verified by side-by-side
		//      registry comparison). ----
		{Code: `<html lang="fil-PH" />`, Tsx: true},    // Filipino in Philippines
		{Code: `<html lang="haw-US" />`, Tsx: true},    // Hawaiian in US
		{Code: `<html lang="ar-001" />`, Tsx: true},    // World Arabic (UN M.49 region code 001)
		{Code: `<html lang="ckb" />`, Tsx: true},       // Central Kurdish (ISO 639-3)
		{Code: `<html lang="be-tarask" />`, Tsx: true}, // Belarusian classical orthography variant

		// ---- Single quote string literal `<html lang='en' />` — same AST
		//      shape as double-quoted. ----
		{Code: `<html lang='en' />`, Tsx: true},
		{Code: `<html lang='zh-Hant-HK' />`, Tsx: true},

		// ---- Leading / trailing whitespace inside the value. `language-tags`
		//      trims via internal subtag splitting; we explicitly TrimSpace
		//      before delegating to language.Parse so the acceptance set
		//      matches upstream. Internal whitespace is still rejected (both
		//      libraries reject "en US"). ----
		{Code: `<html lang=" en" />`, Tsx: true},
		{Code: `<html lang="en " />`, Tsx: true},
		{Code: `<html lang="  en  " />`, Tsx: true},
		{Code: `<html lang=" en-US " />`, Tsx: true},

		// ---- Trailing / repeated hyphens. `language-tags`' subtag iterator
		//      skips empty codes, so "en-" parses as "en" and "en--US"
		//      parses as "en-US" — both valid. We strip empty subtags
		//      before language.Parse to mirror this. Leading hyphen
		//      (`-en`) remains invalid because both libraries require a
		//      non-empty primary language subtag at position 0. ----
		{Code: `<html lang="en-" />`, Tsx: true},
		{Code: `<html lang="en--US" />`, Tsx: true},
		{Code: `<html lang="en-US-" />`, Tsx: true},

		// ---- RFC 5646 conformance (rslint MORE spec-conformant than
		//      `language-tags`): variants with declared prefix relationship.
		//      The IANA registry records `fonxsamp` with Prefix `en-fonipa`,
		//      so `en-fonipa-fonxsamp` is well-formed. `language-tags`
		//      rejects this incorrectly; rslint accepts via language.Parse. ----
		{Code: `<html lang="en-fonipa-fonxsamp" />`, Tsx: true},

		// ============================================================
		// Dimension 4 — JSX attribute-syntax variants
		// ============================================================

		// ---- Namespaced attribute name (`xml:lang`). jsx-ast-utils'
		//      propName returns the composite "xml:lang" string;
		//      uppercase != "LANG" → SKIP. The rule applies ONLY to the
		//      unqualified `lang` attribute. ----
		{Code: `<html xml:lang="foo" />`, Tsx: true},
		{Code: `<html xml:lang="invalid" />`, Tsx: true},
		// Even on non-html, xml:lang is irrelevant.
		{Code: `<div xml:lang="foo" />`, Tsx: true},

		// ---- Whitespace around `=` is parser-allowed in JSX. ----
		{Code: `<html lang ="en" />`, Tsx: true},
		{Code: `<html lang= "en" />`, Tsx: true},
		{Code: `<html lang = "en" />`, Tsx: true},
		{Code: "<html lang =\n\"en\" />", Tsx: true},

		// ---- JSX comments between attributes don't disturb the lang
		//      attribute resolution. ----
		{Code: `<html /* before */ lang="en" /* after */ />`, Tsx: true},

		// ---- HTML-entity-encoded attribute values are decoded before
		//      validation (jsx-ast-utils + jsxtransforms DecodeEntities).
		//      `&#101;` is `e`, `&#110;` is `n` → "en". ----
		{Code: `<html lang="&#101;&#110;" />`, Tsx: true},
		{Code: `<html lang="&#x65;&#x6E;" />`, Tsx: true}, // hex entities
		{Code: `<html lang="e&#110;-&#85;S" />`, Tsx: true},

		// ---- Multiple lang attributes — both valid, no report. tsgo
		//      preserves both; the listener fires twice independently.
		//      JSX runtime warns about duplicate props but it's still
		//      legal syntax. ----
		{Code: `<html lang="en" lang="en-US" />`, Tsx: true},

		// ---- Spread + explicit lang (both orderings). The spread does
		//      NOT fire the JsxAttribute listener (different Kind); the
		//      explicit attribute fires once and validates. ----
		{Code: `<html {...props} lang="en" />`, Tsx: true},
		{Code: `<html lang="en" {...props} />`, Tsx: true},

		// ---- Literal-spread of an object containing lang with an
		//      INVALID value — upstream's JSXAttribute listener never
		//      fires on JSXSpreadAttribute, so neither does ours. The
		//      lang inside the spread literal is invisible to the rule
		//      by design (matches upstream exactly). ----
		{Code: `<html {...{lang: "foo"}} />`, Tsx: true},
		{Code: `<html {...{lang: "zz-LL"}} />`, Tsx: true},
		{Code: `<html {...{lang: undefined}} />`, Tsx: true},

		// ---- Empty spread / spread of non-literal — opaque. ----
		{Code: `<html {...{}} />`, Tsx: true},
		{Code: `<html {...props} />`, Tsx: true},

		// ============================================================
		// TS-specific shapes that must NOT trip the rule
		// ============================================================

		// ---- TS const assertion — `as const` is a TS expression wrapper
		//      with no LITERAL_TYPES entry → null → SKIP. ----
		{Code: `<html lang={"foo" as const} />`, Tsx: true},

		// ---- TS angle-bracket type assertion (`<string>"foo"`) is not
		//      legal in TSX (the `<string>` is parsed as JSX) and we
		//      can't mix it with surrounding JSX. Skipping this shape
		//      since the rule never sees it in real TSX projects. ----

		// ---- Generic component — type "Foo" ≠ "html" → SKIP. ----
		{Code: `<Foo<string> lang="foo" />`, Tsx: true},

		// ---- TS satisfies on the value side (already in Dimension 1) +
		//      satisfies in tag-name position (not legal). Skip the second. ----

		// ============================================================
		// Deep nesting / scope crossing
		// ============================================================

		// ---- 3+ levels deep — listener walks JsxAttribute through
		//      JsxElement → JsxExpression → ArrowFunction → JsxElement … ----
		{Code: `<X r1={() => <Y r2={() => <html lang="en" />} />} />`, Tsx: true},

		// ---- html nested inside conditional/logical short-circuit. The
		//      tag itself is a JsxElement child of a JsxExpression; the
		//      listener walks both branches. ----
		{Code: `<>{cond && <html lang="en" />}</>`, Tsx: true},
		{Code: `<>{cond ? <html lang="en" /> : <div />}</>`, Tsx: true},
		{Code: `<>{cond ? <div /> : <html lang="en" />}</>`, Tsx: true},

		// ---- Array of JSX (e.g. `.map` output) containing valid html. ----
		{Code: `<>{[<html lang="en" key="x" />]}</>`, Tsx: true},

		// ---- IIFE pattern returning html with valid lang. ----
		{Code: `(() => <html lang="en" />)()`, Tsx: true},

		// ============================================================
		// Real-world i18n setups (production patterns from popular libs)
		// ============================================================

		// ---- next-intl: Server component reading locale at request time. ----
		{Code: `async function RootLayout({ children }) { const locale = await getLocale(); return <html lang={locale}>{children}</html>; }`, Tsx: true},

		// ---- next-intl: Static locale from segment params. ----
		{Code: `export default function Layout({ children, params: { locale } }) { return <html lang={locale}>{children}</html>; }`, Tsx: true},

		// ---- react-i18next: `useTranslation().i18n.language` pattern. ----
		{Code: `function Root() { const { i18n } = useTranslation(); return <html lang={i18n.language}><body /></html>; }`, Tsx: true},

		// ---- @lingui/react: `_locale` from `i18n` instance. ----
		{Code: `function Root() { const { i18n } = useLingui(); return <html lang={i18n.locale}><body /></html>; }`, Tsx: true},

		// ---- Plain React Helmet pattern (htmlAttributes). The Helmet
		//      component renders the html element — `<Helmet>` itself
		//      type is "Helmet" ≠ "html" → SKIP. ----
		{Code: `<Helmet htmlAttributes={{ lang: "foo" }} />`, Tsx: true},

		// ---- Gatsby SSR: html element in onRenderBody callback. ----
		{Code: `function onRenderBody({ setHtmlAttributes }) { setHtmlAttributes({ lang: "en" }); return <html lang="en" />; }`, Tsx: true},

		// ============================================================
		// componentMap remapping a custom component AWAY from html
		// ============================================================

		// ---- ComponentMap remaps `Foo: 'div'` — `<Foo lang="foo" />`
		//      resolves to "div", type-gate skips. Locks in the "remap to
		//      non-html disables the rule" path. ----
		{
			Code: `<Foo lang="foo" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"Foo": "div"},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Invalid BCP-47 tags
		// ============================================================

		// ---- Single unknown letter / two-letter unknown language. ----
		{
			Code: `<html lang="qq" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="xx-YY" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Empty string — language.Parse rejects "tag is not well-formed". ----
		{
			Code: `<html lang="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Underscore separator — explicitly rejected for parity with
		//      language-tags (which doesn't accept `_` per BCP-47 ABNF).
		//      Without our explicit pre-check, language.Parse silently
		//      rewrites to `-` and accepts. Lock in the divergence guard. ----
		{
			Code: `<html lang="en_US" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="zh_Hans" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Malformed tag syntax — leading hyphen and illegal characters.
		//      Both rslint and language-tags reject these. Trailing/double
		//      hyphens are NOT in this group: language-tags accepts them by
		//      skipping empty subtags, and we mirror that — see the valid
		//      section above. ----
		{
			Code: `<html lang="-en" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="en@US" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Internal whitespace — both rslint and language-tags reject.
		//      (Surrounding whitespace is trimmed; see valid section.) ----
		{
			Code: `<html lang="en US" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- StringLiteral expression `<html lang={"foo"} />`. ----
		{
			Code: `<html lang={"foo"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- NoSubstitutionTemplateLiteral with invalid tag. ----
		{
			Code: "<html lang={`zz-LL`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- TemplateExpression with substitution synthesizes placeholder
		//      `"en-{region}"` — fails BCP-47 (curly braces are illegal). ----
		{
			Code: "<html lang={`en-${region}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Case-insensitive attribute name with invalid value. ----
		{
			Code: `<html LANG="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ============================================================
		// ExtractOther — non-string literal values
		//
		// Upstream `tags.check(non-string)` throws TypeError. We surface
		// these as normal reports rather than crashing. Documented as a
		// "more permissive than upstream" intentional divergence in the
		// rule's `.md`.
		// ============================================================

		// ---- Boolean attribute form `<html lang />` — upstream extract
		//      null-attr-value path returns boolean true → throws. ----
		{
			Code: `<html lang />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Literal true / false — upstream throws. ----
		{
			Code: `<html lang={true} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang={false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- StringLiteral "true" / "false" — jsxAstUtilsLiteralCoerce
		//      maps these case-insensitively to bools → throws upstream. ----
		{
			Code: `<html lang="true" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="FALSE" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- NumericLiteral — upstream throws. ----
		{
			Code: `<html lang={1} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang={0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- BigInt literal. ----
		{
			Code: `<html lang={1n} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Array / Object literal. ----
		{
			Code: `<html lang={["en"]} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ============================================================
		// ExtractString edge cases — literal-resolves-to-string, BCP-47 fails
		// ============================================================

		// ---- `<html lang={null} />` — upstream LITERAL_TYPES.Literal special-
		//      cases null to the magic string "null". tags.check("null") false
		//      → REPORT. We mirror via literalPropValue's NullKeyword arm. ----
		{
			Code: `<html lang={null} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ============================================================
		// Nested / boundary edge cases
		// ============================================================

		// ---- Two nested html elements, both invalid. Listener fires twice
		//      independently — two reports. Locks in stateless listener.
		//      Note "qq" / "qq-XX" are deliberately chosen so each is well-
		//      formed but unknown to the IANA registry; bare 3-letter
		//      placeholders like "bar" / "baz" happen to be registered ISO
		//      639-3 codes and would accidentally pass validation. ----
		{
			Code: `<html lang="qq"><html lang="qq-XX" /></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 23},
			},
		},

		// ---- html inside Fragment / sibling — listener walks deeply. ----
		{
			Code: `<><div /><html lang="zz-LL" /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 16},
			},
		},

		// ---- Multi-line opening tag with invalid value — position still
		//      points at the lang attribute. ----
		{
			Code: "<html\n  lang=\"foo\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 2, Column: 3},
			},
		},

		// ---- componentMap remap reports with proper column. ----
		{
			Code:     `<HTMLRoot lang="zz-LL" />`,
			Tsx:      true,
			Settings: htmlRemapSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 11},
			},
		},

		// ============================================================
		// Listener gate — wrong-attribute / wrong-element edge cases
		// already covered as valid above. These ensure the rule reports
		// on cases that DO trip the gate.
		// ============================================================

		// ---- Class component returning html with invalid lang. ----
		{
			Code: `function Root() { return <html lang="bogus" />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 32},
			},
		},

		// ============================================================
		// Dimension 4 — JSX attribute-syntax variants (invalid path)
		// ============================================================

		// ---- Single-quoted invalid value — same as double-quote. ----
		{
			Code: `<html lang='foo' />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Whitespace around `=` with invalid value. ----
		{
			Code: `<html lang = "foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- JSX comment surrounding invalid lang attribute. ----
		{
			Code: `<html /* a */ lang="foo" /* b */ />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 15},
			},
		},

		// ---- HTML-entity-encoded INVALID value: `&#x66;oo` = "foo". The
		//      direct StringLiteral path runs entity decoding before tag
		//      validation, so the decoded "foo" reaches isValidBCP47Tag. ----
		{
			Code: `<html lang="&#x66;oo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		// Decimal-encoded `&#102;&#111;&#111;` = "foo".
		{
			Code: `<html lang="&#102;&#111;&#111;" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Multibyte / non-ASCII chars — language.Parse rejects as
		//      not well-formed (BCP-47 only allows ASCII letters / digits
		//      / hyphens). ----
		{
			Code: `<html lang="日本語" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="中文" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		// Latin-1 supplement char (`é`) — invalid for BCP-47.
		{
			Code: `<html lang="café" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		// Zero-width space (U+200B) hidden inside an otherwise valid-looking tag.
		{
			Code: "<html lang=\"en\u200b\" />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Subtag too long (BCP-47 caps subtags at 8 chars). ----
		{
			Code: `<html lang="english" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="en-aaaaaaaaa" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Single-char primary language (must be 2-3 chars). ----
		{
			Code: `<html lang="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="a" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- RFC 5646 conformance (rslint MORE spec-conformant than
		//      `language-tags`): extension/private-use singletons MUST
		//      carry a 1-8 alphanum payload subtag (RFC 5646 §2.2.6 /
		//      §2.2.7). `language-tags` accepts the malformed lone-
		//      singleton tail; rslint rejects via language.Parse. ----
		{
			Code: `<html lang="en-x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="en-u" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="en-x-" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Numeric-only primary language — BCP-47 requires alpha. ----
		{
			Code: `<html lang="123" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Multiple lang attributes — both invalid. tsgo preserves
		//      duplicates; the listener fires twice → two reports. ----
		{
			Code: `<html lang="foo" lang="zz-LL" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 18},
			},
		},

		// ---- Multiple lang attributes — one valid, one invalid. Only
		//      the invalid one reports. ----
		{
			Code: `<html lang="en" lang="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 17},
			},
		},

		// ---- Spread + explicit INVALID lang. Explicit attribute fires
		//      and reports; spread is opaque. ----
		{
			Code: `<html {...props} lang="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 18},
			},
		},
		{
			Code: `<html lang="zz-LL" {...props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ============================================================
		// Deep nesting + boundary edge cases (invalid path)
		// ============================================================

		// ---- 3+ levels of nesting through render-props. The innermost
		//      html is reached by the linter's DFS childVisitor. ----
		{
			Code: `<X r1={() => <Y r2={() => <html lang="bogus" />} />} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 33},
			},
		},

		// ---- html inside conditional / logical short-circuit. ----
		{
			Code: `<>{cond && <html lang="foo" />}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 18},
			},
		},
		{
			Code: `<>{cond ? <html lang="foo" /> : <div />}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 17},
			},
		},
		// Both branches of ternary contain html, both invalid. "qq" and
		// "zz-LL" are both well-formed but unknown to the IANA registry.
		{
			Code: `<>{cond ? <html lang="qq" /> : <html lang="zz-LL" />}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 17},
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 38},
			},
		},

		// ---- Array of JSX (e.g. .map output) with invalid html. ----
		{
			Code: `<>{[<html lang="foo" key="x" />]}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 11},
			},
		},

		// ============================================================
		// Real-world production mistakes
		// ============================================================

		// ---- Hardcoded misspelled locale in Next.js layout (common
		//      copy-paste accident). ----
		{
			Code: `export default function RootLayout({ children }) { return <html lang="enUS"><body>{children}</body></html>; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 65},
			},
		},

		// ---- Mistaken POSIX-style locale: "en.US" — `.` is illegal. ----
		{
			Code: `<html lang="en.US"><body /></html>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Forward slash separator (file-extension style mistake). ----
		{
			Code: `<html lang="en/US" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Full English word as lang — common newcomer error. ----
		{
			Code: `<html lang="english" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="english-us" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ============================================================
		// componentMap + polymorphic combined gates (invalid path)
		// ============================================================

		// ---- polymorphicAllowList includes Foo, `as="html"` resolves
		//      to "html"; the lang value is invalid → REPORT. ----
		{
			Code: `<Foo as="html" lang="foo" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Foo"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 16},
			},
		},

		// ---- componentMap chain via JSON-decoded interface map. ----
		{
			Code: `<HtmlRoot lang="bogus" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"Foo":      "div",
						"HtmlRoot": "html",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 11},
			},
		},

		// ============================================================
		// BCP-47 alignment with `language-tags` beyond plain registry
		// membership. Each of these would pass `language.Parse` alone
		// but is rejected by `language-tags`; rslint's extra checks in
		// bcp47.go bring them back into alignment.
		// ============================================================

		// ---- Deprecated grandfathered tags — `language-tags` reports
		//      these via `.deprecated()` truthy. The IANA registry
		//      records the preferred replacement, so the user should
		//      migrate (e.g. `i-klingon` → `tlh`). ----
		{
			Code: `<html lang="i-klingon" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="en-gb-oed" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="zh-min-nan" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="no-bok" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		// Case-insensitive match — uppercase variants of the grandfathered
		// form are still rejected.
		{
			Code: `<html lang="I-Klingon" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Suppress-Script violations. RFC 5646 §4.1: a script subtag
		//      should be omitted when it matches the language's default
		//      script as recorded in the IANA registry. `language-tags`
		//      flags this with "The script subtag 'X' is the same as the
		//      language suppress-script." ----
		{
			Code: `<html lang="hi-Deva" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="ja-Jpan" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="en-Latn" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="fr-Latn" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		// With region after Suppress-Script — the violation still triggers
		// because the explicit script subtag is the suppressed one.
		{
			Code: `<html lang="en-Latn-US" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Private-use-only tags (no primary language). Both `x-foo`
		//      and bare `x` trigger language-tags' "Missing language tag"
		//      error. A combined tag with primary language (e.g.
		//      `en-x-comp` — see valid section) is accepted. ----
		{
			Code: `<html lang="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="x-foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
		{
			Code: `<html lang="x-private" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- en-US-POSIX — `POSIX` is a legacy CLDR variant that is NOT
		//      a registered variant subtag in the IANA registry, so
		//      language-tags reports "Unknown code 'posix'". Note this
		//      diverges from how some browsers / tooling tolerate it. ----
		{
			Code: `<html lang="en-US-POSIX" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidLangValue", Message: errorMessage, Line: 1, Column: 7},
			},
		},
	})
}
