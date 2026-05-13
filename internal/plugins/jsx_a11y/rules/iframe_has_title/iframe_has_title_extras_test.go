package iframe_has_title

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// polymorphicSettings exercises the `polymorphicPropName` settings entry —
// `<Foo as="iframe">` resolves to nodeType "iframe" and falls through to the
// title check. Upstream's own test file doesn't cover this path for
// iframe-has-title.
var polymorphicSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
	},
}

// TestIframeHasTitleExtras locks in branches that upstream's test file doesn't
// exercise but are reachable through the rule's listener gate. Each case
// carries an inline comment pointing at the specific upstream branch /
// Dimension 4 edge shape it covers. These tests protect against silent
// regressions during refactors of either the rule or its shared helpers.
//
// The defining difference from html-has-lang is the second guard:
//
//	if (title && typeof title === 'string') return;
//
// — truthy alone is not enough. Most of the cases below either lock in the
// "synthesizes-string" upstream extractors that DO satisfy `typeof string`
// (CallExpression, MemberExpression, JsxElement, JsxFragment,
// TaggedTemplateExpression, AssignmentExpression, etc.) or assert that
// non-string truthy shapes (Object / Array / New / RegExp / function /
// number / boolean / bigint) correctly REPORT despite being truthy.
func TestIframeHasTitleExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &IframeHasTitleRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: paired form `<iframe>...</iframe>`. The listener
		//      fires on KindJsxOpeningElement once for the opening tag;
		//      title is truthy string → no report. ----
		{Code: `<iframe title="x"></iframe>`, Tsx: true},
		{Code: `<iframe title="x">child</iframe>`, Tsx: true},

		// ---- Dimension 4: tag-name forms — case-sensitive iframe matching ----
		// Upstream's `type !== 'iframe'` check is case-sensitive. Capitalized
		// or differently-cased tag names are NOT iframe for this rule.
		{Code: `<IFRAME />`, Tsx: true},
		{Code: `<Iframe />`, Tsx: true},
		{Code: `<IFrame />`, Tsx: true},

		// ---- Dimension 4: namespaced JSX names — type "svg:iframe" is truthy
		//      and != "iframe" → SKIP. ----
		{Code: `<svg:iframe />`, Tsx: true},

		// ---- Dimension 4: member-access tag — type "Foo.iframe" → SKIP. ----
		{Code: `<Foo.iframe />`, Tsx: true},

		// ---- Other HTML tags must NOT be reported — the rule only applies
		//      to <iframe> itself. ----
		{Code: `<div title="x" />`, Tsx: true},
		{Code: `<span />`, Tsx: true},
		{Code: `<frame />`, Tsx: true}, // <frame> is its own element; not <iframe>

		// ---- Dimension 4: case-insensitive ATTRIBUTE name (jsx-ast-utils
		//      `getProp` default `ignoreCase: true`). ----
		{Code: `<iframe TITLE="x" />`, Tsx: true},
		{Code: `<iframe Title="x" />`, Tsx: true},
		{Code: `<iframe tItLe="x" />`, Tsx: true},

		// ============================================================
		// Synthesized-string upstream extractors — these all return
		// `typeof === 'string'` per jsx-ast-utils, so iframe-has-title
		// treats them as VALID. The defining iframe-has-title VALID
		// surface that goes BEYOND a literal string.
		// ============================================================

		// ---- CallExpression / OptionalCallExpression — upstream synthesizes
		//      "callee(args)" string. ----
		{Code: `<iframe title={getTitle()} />`, Tsx: true},
		{Code: `<iframe title={i18n.t("title")} />`, Tsx: true},
		{Code: `<iframe title={getTitle?.()} />`, Tsx: true},

		// ---- MemberExpression / OptionalMemberExpression / ElementAccess —
		//      upstream synthesizes "obj.prop" / "obj?.prop" / "obj[key]". ----
		{Code: `<iframe title={config.title} />`, Tsx: true},
		{Code: `<iframe title={config.title.value} />`, Tsx: true},
		{Code: `<iframe title={config?.title} />`, Tsx: true},
		{Code: `<iframe title={config["title"]} />`, Tsx: true},
		{Code: `<iframe title={titles[0]} />`, Tsx: true},

		// ---- JsxElement / JsxSelfClosingElement / JsxFragment —
		//      upstream synthesizes "<Tag />" / "<Tag>...</Tag>" / "<></>"
		//      strings. (Atypical title value but legal TSX.) ----
		{Code: `<iframe title={<span>title</span>} />`, Tsx: true},
		{Code: `<iframe title={<X />} />`, Tsx: true},
		{Code: `<iframe title={<></>} />`, Tsx: true},

		// ---- TaggedTemplateExpression — upstream redirects to inner
		//      template literal, so a non-empty quasi yields a non-empty
		//      string. ----
		{Code: "<iframe title={tag`hello`} />", Tsx: true},
		{Code: "<iframe title={tag`hello ${name}`} />", Tsx: true},

		// ---- this keyword — upstream returns the magic string "this". ----
		{Code: `<iframe title={this} />`, Tsx: true},

		// ---- Identifier (non-reserved) — upstream returns the bare name
		//      string ("title"). ----
		{Code: `<iframe title={title} />`, Tsx: true},
		{Code: `<iframe title={someTitle} />`, Tsx: true},

		// ---- TemplateExpression with substitutions — synthesized string
		//      via TemplateLiteral extractor (head/tail + placeholders). ----
		{Code: "<iframe title={`Frame ${name}`} />", Tsx: true},
		{Code: "<iframe title={`prefix-${id}-suffix`} />", Tsx: true},

		// ---- NoSubstitutionTemplate — string literal value. ----
		{Code: "<iframe title={`hello`} />", Tsx: true},

		// ---- ConditionalExpression — recurse into the chosen branch. ----
		{Code: `<iframe title={cond ? "a" : "b"} />`, Tsx: true},
		{Code: `<iframe title={cond ? getName() : "fallback"} />`, Tsx: true},
		{Code: `<iframe title={cond ? config.title : "fallback"} />`, Tsx: true},

		// ---- Logical `||` — short-circuits when left is truthy and returns
		//      the left side; an Identifier left is a non-empty string. ----
		{Code: `<iframe title={title || "Default"} />`, Tsx: true},
		{Code: `<iframe title={getTitle() || "Default"} />`, Tsx: true},

		// ---- Logical `??` — falls through to the right side when left is
		//      null/undefined; otherwise returns left. Identifier left is
		//      a string. ----
		{Code: `<iframe title={title ?? "Default"} />`, Tsx: true},

		// ---- Logical `&&` — when left is truthy, returns the right side. ----
		{Code: `<iframe title={cond && "Title"} />`, Tsx: true},

		// ---- BinaryExpression `+` string concat — string concat path
		//      makes the result a jvString. ----
		{Code: `<iframe title={"prefix-" + name} />`, Tsx: true},
		{Code: `<iframe title={"a" + "b"} />`, Tsx: true},
		{Code: `<iframe title={"" + getTitle()} />`, Tsx: true},

		// ---- AssignmentExpression — upstream's extractor returns
		//      "${left} ${op} ${right}" string. tsgo collapses ESTree's
		//      AssignmentExpression into BinaryExpression with an
		//      assignment operator; staticEval treats this kind as
		//      jvTruthy (truthy generic), but iframe-has-title's typeof
		//      gate would actually need a string. Lock the current
		//      behavior in below in the INVALID block as a known
		//      divergence — see comment there. ----

		// ---- Dimension 4: TS-only expression wrappers on the title value.
		//      jsx-ast-utils' `extract` while-loops past TSAsExpression and
		//      TSNonNullExpression — staticEval mirrors via skipTransparent. ----
		{Code: `<iframe title={"x" as string} />`, Tsx: true},
		{Code: `<iframe title={title!} />`, Tsx: true},
		{Code: `<iframe title={(title) as any} />`, Tsx: true},

		// ---- Dimension 4: parenthesized expressions on title value
		//      (single + multi-level). tsgo preserves Parens; ESTree
		//      flattens them. ----
		{Code: `<iframe title={("x")} />`, Tsx: true},
		{Code: `<iframe title={(("x"))} />`, Tsx: true},

		// ---- Locks in upstream `getProp`'s LITERAL spread walk: a literal
		//      ObjectLiteral spread carrying `title` matches as if the prop
		//      were declared directly. ----
		{Code: `<iframe {...{title: "x"}} />`, Tsx: true},
		{Code: `<iframe {...{ title }} />`, Tsx: true}, // shorthand → Identifier "title" → string

		// ---- Polymorphic-prop branch via `getElementType`:
		//      `<Foo as="iframe" title="x" />` resolves to nodeType "iframe",
		//      title truthy string → no report. ----
		{
			Code:     `<Foo as="iframe" title="x" />`,
			Tsx:      true,
			Settings: polymorphicSettings,
		},
		// Polymorphic with non-iframe target — `<Foo as="div" />` is treated
		// as a div by getElementType → not matched.
		{
			Code:     `<Foo as="div" />`,
			Tsx:      true,
			Settings: polymorphicSettings,
		},

		// ---- componentMap with literal-spread + title inside the spread. ----
		{
			Code:     `<FooComponent {...{title: "x"}} />`,
			Tsx:      true,
			Settings: componentsSettings,
		},

		// ---- Multi-line opening tag formatting must still resolve. ----
		{Code: "<iframe\n  title=\"x\"\n/>", Tsx: true},
		{Code: "<iframe\n  title=\"x\"\n>\n</iframe>", Tsx: true},

		// ---- Mixed siblings: only the iframe element is checked. ----
		{Code: `<><div /><iframe title="x" /></>`, Tsx: true},

		// ---- Settings-driven disable: remap "iframe" → "div" silences the
		//      rule for the literal `<iframe />` tag. Locks in the listener
		//      gate's "skipped when nodeType is not iframe" path. ----
		{
			Code:     `<iframe />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"components": map[string]interface{}{"iframe": "div"}}},
		},

		// ---- Other props on iframe besides title — only title is inspected. ----
		{Code: `<iframe title="x" src="https://example.com" />`, Tsx: true},
		{Code: `<iframe src="https://example.com" title="x" width="500" />`, Tsx: true},

		// ---- Spread + explicit title. Spread is opaque under upstream's
		//      `getProp` for non-literal arguments, but the explicit title
		//      attribute is found and is a non-empty string → no report. ----
		{Code: `<iframe {...props} title="x" />`, Tsx: true},
		// Reverse order — explicit title BEFORE spread. Same outcome.
		{Code: `<iframe title="x" {...props} />`, Tsx: true},

		// ---- Dimension 2: nested iframe — outer has title, inner has title.
		//      Listener fires twice; both pass independently. ----
		{Code: `<iframe title="outer"><iframe title="inner" /></iframe>`, Tsx: true},

		// ---- Dimension 2: iframe inside JsxExpression child of another
		//      element. Listener walks the entire tree; the inner iframe
		//      fires through KindJsxSelfClosingElement. ----
		{Code: `<div>{<iframe title="x" />}</div>`, Tsx: true},

		// ---- Dimension 2: iframe inside an attribute value (arrow body /
		//      render prop). ----
		{Code: `<X render={() => <iframe title="x" />} />`, Tsx: true},
		{Code: `<X children={<iframe title="x" />} />`, Tsx: true},

		// ---- Dimension 2: iframe inside JsxFragment. ----
		{Code: `<><iframe title="a" /><iframe title="b" /></>`, Tsx: true},

		// ============================================================
		// Real-world embed patterns (YouTube / Vimeo / Maps / Stripe)
		// ============================================================

		// ---- YouTube embed with explicit title. ----
		{Code: `<iframe title="YouTube video player" src="https://www.youtube.com/embed/dQw4w9WgXcQ" />`, Tsx: true},
		// ---- Vimeo embed. ----
		{Code: `<iframe title="Vimeo player" src="https://player.vimeo.com/video/12345" />`, Tsx: true},
		// ---- Google Maps embed. ----
		{Code: `<iframe title="Map of office" src="https://www.google.com/maps/embed/..." />`, Tsx: true},
		// ---- Stripe payment iframe. ----
		{Code: `<iframe title="Stripe payment form" src="https://js.stripe.com/..." />`, Tsx: true},

		// ---- Dynamic title from props/state — Identifier resolves to a
		//      non-empty string. ----
		{Code: `<iframe title={props.frameTitle} src={props.src} />`, Tsx: true},
		{Code: `<iframe title={frameTitle} />`, Tsx: true},

		// ---- Common i18n patterns — CallExpression / MemberExpression both
		//      synthesize strings upstream. ----
		{Code: `<iframe title={t("video.title")} src={src} />`, Tsx: true},
		{Code: `<iframe title={i18n.t("frame.title")} />`, Tsx: true},
		{Code: `<iframe title={i18n.translate("title")} />`, Tsx: true},

		// ---- Title with template substitution. ----
		{Code: "<iframe title={`Video: ${videoName}`} src={src} />", Tsx: true},

		// ---- Functional component returning iframe. ----
		{Code: `function Video({ src, title }) { return <iframe title={title} src={src} />; }`, Tsx: true},
		{Code: `const Video = ({ title }) => <iframe title={title} />;`, Tsx: true},

		// ---- Functional component with default title. ----
		{Code: `function Video({ title = "Default" }) { return <iframe title={title} />; }`, Tsx: true},

		// ---- Title resolved at runtime via fallback chain. Identifier
		//      "videoTitle" is a non-empty string upstream. ----
		{Code: `<iframe title={videoTitle ?? "Untitled video"} />`, Tsx: true},
		{Code: `<iframe title={videoTitle || "Untitled video"} />`, Tsx: true},

		// ---- Memoized / wrapped in HoCs — the inner iframe still fires. ----
		{Code: `memo(({ title }) => <iframe title={title} />)`, Tsx: true},
		{Code: `forwardRef((props, ref) => <iframe title={props.title} ref={ref} />)`, Tsx: true},

		// ============================================================
		// Settings: components / polymorphicAllowList / multi-key remaps
		// ============================================================

		// ---- Multi-component remap. Only matching keys remap. ----
		{
			Code: `<MyVideo title="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"Foo":     "div",
						"MyVideo": "iframe",
						"Bar":     "span",
					},
				},
			},
		},

		// ---- polymorphicAllowList includes the rawType — `<Foo as="iframe">`
		//      gets remapped to "iframe" and title="x" → no report. ----
		{
			Code: `<Foo as="iframe" title="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Foo"},
				},
			},
		},

		// ---- polymorphicAllowList does NOT include rawType — `<Bar as="iframe">`
		//      keeps rawType "Bar", which is truthy and != "iframe" → SKIP. ----
		{
			Code: `<Bar as="iframe" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{"Foo"},
				},
			},
		},

		// ---- Tab whitespace inside the opening tag must still resolve. ----
		{Code: "<iframe\ttitle=\"x\"\t/>", Tsx: true},

		// ---- CRLF line endings inside multi-line opening tag. ----
		{Code: "<iframe\r\n  title=\"x\"\r\n/>", Tsx: true},

		// ---- Numeric edge cases that go through `+` string concat — the
		//      `+` operator promotes to string when either side is a
		//      string. ----
		{Code: `<iframe title={"v" + 42} />`, Tsx: true}, // "v" + 42 → "v42"
		{Code: `<iframe title={"" + name + 1} />`, Tsx: true},

		// ---- Multi-key literal spread; "key" is excluded per upstream
		//      `getProp`. The rest of the keys are walked. ----
		{Code: `<iframe {...{key: someKey, title: "x"}} />`, Tsx: true},
		{Code: `<iframe {...{title: "x", key: someKey}} />`, Tsx: true},

		// ---- Multi-attribute mixing direct + literal-spread. ----
		{Code: `<iframe className="x" {...{title: "x"}} />`, Tsx: true},

		// ---- Nullish chain with Identifier left arm — staticEval for
		//      `a ?? "x"` evaluates `a` as Identifier "a" → returns
		//      string "a" which is non-null/undef → returns left → "a"
		//      → typeof string → no report. ----
		{Code: `<iframe title={a ?? b ?? ""} />`, Tsx: true},
		// `||` chain with leading Identifier — same logic; "a" truthy → returns left.
		{Code: `<iframe title={a || "" || false} />`, Tsx: true},

		// ---- Conditional with one branch synthesized-string. The truthy
		//      Identifier "cond" → recurse into the WhenTrue branch. ----
		{Code: `<iframe title={cond ? title : ""} />`, Tsx: true},

		// ---- Empty components-map (unconfigured remap): falls through to
		//      raw element type. <iframe> stays as iframe → checks title. ----
		{
			Code: `<iframe title="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"components": map[string]interface{}{}},
			},
		},

		// ---- Settings: malformed values must not crash — gracefully fall
		//      back to bare element-type resolution. ----
		// Non-string polymorphicPropName → ignored.
		{
			Code: `<iframe title="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"polymorphicPropName": 42},
			},
		},
		// Non-array polymorphicAllowList → treated as no-allowlist.
		{
			Code: `<Foo as="iframe" title="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": "Foo",
				},
			},
		},
		// Non-map components → ignored.
		{
			Code: `<iframe title="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{"components": "iframe"},
			},
		},

		// ============================================================
		// Async / generator / RSC / IIFE / HoC container patterns —
		// the listener walks every JsxOpeningElement regardless of
		// surrounding container shape. Locks in that the rule is
		// purely AST-driven, not control-flow-aware.
		// ============================================================

		// ---- Async functional component (RSC-style root). ----
		{Code: `async function Frame({ title }) { return <iframe title={title} />; }`, Tsx: true},
		{Code: `async function Frame() { const t = await getTitle(); return <iframe title={t} />; }`, Tsx: true},

		// ---- Async arrow with implicit return. ----
		{Code: `const Frame = async ({ title }) => <iframe title={title} />;`, Tsx: true},

		// ---- Generator yields iframe. ----
		{Code: `function* gen() { yield <iframe title="x" />; }`, Tsx: true},
		{Code: `async function* asyncGen() { yield <iframe title="x" />; }`, Tsx: true},

		// ---- React.lazy / Suspense / ErrorBoundary wrapping. ----
		{Code: `<Suspense fallback={null}><iframe title="x" /></Suspense>`, Tsx: true},
		{Code: `<ErrorBoundary fallback={null}><iframe title="x" /></ErrorBoundary>`, Tsx: true},

		// ---- IIFE returning iframe. ----
		{Code: `(() => <iframe title="x" />)()`, Tsx: true},
		{Code: `(function () { return <iframe title="x" />; })()`, Tsx: true},

		// ---- iframe inside .map / forEach iteration. The listener
		//      walks into arrow bodies inside CallExpression args. ----
		{Code: `<>{frames.map(f => <iframe key={f.id} title={f.title} src={f.src} />)}</>`, Tsx: true},
		{Code: `<>{frames.map(({id, title, src}) => <iframe key={id} title={title} src={src} />)}</>`, Tsx: true},

		// ---- iframe inside switch / case via IIFE. ----
		{Code: `(() => { switch (kind) { case "a": return <iframe title="A" />; default: return null; } })()`, Tsx: true},

		// ---- iframe inside try / catch via IIFE. ----
		{Code: `(() => { try { return <iframe title="x" />; } catch { return null; } })()`, Tsx: true},

		// ---- React.createElement is NOT a JsxOpeningElement; the rule
		//      doesn't fire on this CallExpression-based form even when
		//      the args describe an iframe. The rule only inspects JSX. ----
		{Code: `React.createElement("iframe");`, Tsx: true},
		{Code: `React.createElement("iframe", null, null);`, Tsx: true},

		// ============================================================
		// Real-world embed providers (broader coverage). All follow
		// the canonical "embed with explicit title" pattern that
		// teams should adopt for accessibility compliance.
		// ============================================================

		{Code: `<iframe title="Twitch stream" src="https://player.twitch.tv/?channel=foo" />`, Tsx: true},
		{Code: `<iframe title="Calendly scheduling" src="https://calendly.com/foo" />`, Tsx: true},
		{Code: `<iframe title="Spotify track player" src="https://open.spotify.com/embed/track/x" />`, Tsx: true},
		{Code: `<iframe title="SoundCloud player" src="https://w.soundcloud.com/player/?url=x" />`, Tsx: true},
		{Code: `<iframe title="CodePen demo" src="https://codepen.io/x/embed/y" />`, Tsx: true},
		{Code: `<iframe title="JSFiddle embed" src="https://jsfiddle.net/x/embedded/" />`, Tsx: true},
		{Code: `<iframe title="Twitter / X post" src="https://platform.twitter.com/embed/x" />`, Tsx: true},
		{Code: `<iframe title="Reddit thread" src="https://www.redditmedia.com/embed/x" />`, Tsx: true},
		{Code: `<iframe title="Embedded PDF document" src="/docs/manual.pdf" />`, Tsx: true},
		{Code: `<iframe title="Auth0 Universal Login" src="https://example.auth0.com/login" />`, Tsx: true},
		{Code: `<iframe title="reCAPTCHA challenge" src="https://www.google.com/recaptcha/api2/bframe" />`, Tsx: true},
		{Code: `<iframe title="Embed.ly preview" src="https://cdn.embedly.com/widgets/x" />`, Tsx: true},

		// ---- Modern attribute combos: lazy-loading / sandbox /
		//      allowFullScreen / referrerPolicy. None affect the title
		//      check; only `title` is inspected. ----
		{Code: `<iframe title="Lazy-loaded video" src="https://example.com/video" loading="lazy" />`, Tsx: true},
		{Code: `<iframe title="Sandbox embed" src="https://example.com" sandbox="allow-scripts" />`, Tsx: true},
		{Code: `<iframe title="Fullscreen-capable" src="https://example.com" allowFullScreen />`, Tsx: true},
		{Code: `<iframe title="Cross-origin embed" src="https://example.com" referrerPolicy="no-referrer" />`, Tsx: true},

		// ============================================================
		// TS-specific composition shapes — chained type assertions,
		// non-null + member chains, generic JSX, const assertion,
		// template-literal types. All are stripped or unwrapped via
		// skipTransparent so the inner expression's type-classification
		// drives the typeof-string gate.
		// ============================================================

		// ---- Chained `as A as B`. ----
		{Code: `<iframe title={value as unknown as string} />`, Tsx: true},
		{Code: `<iframe title={value as string as any} />`, Tsx: true},

		// ---- TS non-null + member chains. ----
		{Code: `<iframe title={props!.title} />`, Tsx: true},
		{Code: `<iframe title={props!.title!} />`, Tsx: true},
		{Code: `<iframe title={props!.title!.trim()} />`, Tsx: true},
		{Code: `<iframe title={(props.title)!} />`, Tsx: true},
		{Code: `<iframe title={(props!.title)!} />`, Tsx: true},

		// ---- Generic JSX (TSX). ----
		{Code: `<Box<HTMLProps> title="x"><iframe title="y" /></Box>`, Tsx: true},

		// ---- TS const assertion. ----
		{Code: `<iframe title={"x" as const} />`, Tsx: true},
		{Code: `<iframe title={value as const} />`, Tsx: true},

		// ---- TS template literal type — `as` is stripped, value
		//      extracts to its underlying expression's classification. ----
		{Code: "<iframe title={value as `prefix-${string}`} />", Tsx: true},

		// ============================================================
		// JSX content variants (paired form children). The listener
		// fires on the OPENING tag regardless of children shape.
		// ============================================================

		{Code: `<iframe title="x"></iframe>`, Tsx: true},
		{Code: `<iframe title="x">child text</iframe>`, Tsx: true},
		{Code: `<iframe title="x">{children}</iframe>`, Tsx: true},
		{Code: `<iframe title="x">{/* comment */}</iframe>`, Tsx: true},
		{Code: `<iframe title="x"><span /></iframe>`, Tsx: true},
		{Code: `<iframe title="x"><span /><div /></iframe>`, Tsx: true},
		{Code: `<iframe title="x"><>{frag}</></iframe>`, Tsx: true},
		{Code: `<iframe title="x">{cond && <span />}</iframe>`, Tsx: true},

		// ============================================================
		// Multi-spread + mixed ordering with explicit attrs. getProp
		// walks attributes left-to-right and returns the FIRST match
		// (whether direct or via literal-spread walk).
		// ============================================================

		// ---- Multiple non-literal spreads + explicit title (after).
		//      Spreads are opaque → walk continues → finds explicit. ----
		{Code: `<iframe {...a} {...b} title="x" />`, Tsx: true},
		// ---- Explicit before opaque spreads. ----
		{Code: `<iframe title="x" {...a} {...b} />`, Tsx: true},
		// ---- Two literal spreads, only second has title. ----
		{Code: `<iframe {...{className: "x"}} {...{title: "y"}} />`, Tsx: true},
		// ---- First literal spread has title — stops there. ----
		{Code: `<iframe {...{title: "x"}} {...{title: "y"}} />`, Tsx: true},
		// ---- Explicit before a literal spread that ALSO has title. ----
		{Code: `<iframe title="x" {...{title: "y"}} />`, Tsx: true},
		// ---- Literal spread with multiple attributes including title. ----
		{Code: `<iframe {...{src: "u", title: "x", width: 500}} />`, Tsx: true},
		// ---- Empty literal spread + explicit title. ----
		{Code: `<iframe {...{}} title="x" />`, Tsx: true},
		// ---- Spread containing nested SpreadAssignment + valid title.
		//      The walker matches title via PropertyAssignment; the
		//      SpreadAssignment sibling is harmlessly skipped. ----
		{Code: `<iframe {...{...rest, title: "x"}} />`, Tsx: true},

		// ---- Computed-key spread where the inner expression is a bare
		//      Identifier. jsx-ast-utils' getProp guard is
		//      `property.key.type === 'Identifier'` WITHOUT checking
		//      ESTree's `computed` flag — so `{[title]: "x"}` (computed
		//      Identifier key) is matched the same as `{title: "x"}`. Lock
		//      in this upstream-quirky-but-load-bearing alignment after
		//      jsxa11yutil.FindAttributeByName unwraps tsgo's
		//      ComputedPropertyName when its inner is an Identifier. ----
		{Code: `<iframe {...{[title]: "x"}} />`, Tsx: true},
		// Case-insensitive match through the computed-Identifier path.
		{Code: `<iframe {...{[Title]: "x"}} />`, Tsx: true},
		{Code: `<iframe {...{[TITLE]: "x"}} />`, Tsx: true},
		// Parens inside computed brackets are flattened by ESTree at parse
		// time; tsgo preserves them, so the helper strips parens to match.
		{Code: `<iframe {...{[(title)]: "x"}} />`, Tsx: true},
		{Code: `<iframe {...{[((title))]: "x"}} />`, Tsx: true},

		// ============================================================
		// Template literal exhaustive forms. Substitutions that are
		// Identifier / Call / Member / Conditional / Logical produce
		// non-empty placeholder strings → the whole template is a
		// truthy string per upstream's TemplateLiteral extractor.
		// ============================================================

		// ---- Multi-substitution Identifier. ----
		{Code: "<iframe title={`${a}-${b}-${c}`} />", Tsx: true},
		// ---- Back-to-back substitutions with no quasi between. ----
		{Code: "<iframe title={`${a}${b}`} />", Tsx: true},
		// ---- Substitution at the very start. ----
		{Code: "<iframe title={`${prefix} text`} />", Tsx: true},
		// ---- Substitution at the very end. ----
		{Code: "<iframe title={`text ${suffix}`} />", Tsx: true},
		// ---- Quasi with raw escape sequences. ----
		{Code: "<iframe title={`line1\\nline2`} />", Tsx: true},
		{Code: "<iframe title={`tab\\there`} />", Tsx: true},
		// ---- Substitution containing CallExpression / MemberExpression. ----
		{Code: "<iframe title={`${getName()}`} />", Tsx: true},
		{Code: "<iframe title={`${obj.prop}`} />", Tsx: true},
		// ---- Nested template inside substitution. ----
		{Code: "<iframe title={`outer ${`inner ${x}`}`} />", Tsx: true},
		// ---- Template with `${undefined}` — upstream specially returns
		//      "undefined" name (non-empty truthy string). ----
		{Code: "<iframe title={`${undefined}`} />", Tsx: true},
		// ---- Tagged template with non-empty quasi + Identifier sub. ----
		{Code: "<iframe title={tag`hello ${name}`} />", Tsx: true},
		// ---- String literal with single-quote (legal in JSX expression). ----
		{Code: `<iframe title={'x'} />`, Tsx: true},
		// ---- String literal with leading/trailing whitespace — non-
		//      empty per JS truthiness; the rule doesn't trim. ----
		{Code: `<iframe title=" " />`, Tsx: true},
		{Code: `<iframe title={" "} />`, Tsx: true},
		{Code: `<iframe title="\n" />`, Tsx: true},
		// ---- Unicode escape inside string literal. ----
		{Code: `<iframe title="X" />`, Tsx: true},
		// ---- Backslash literal. ----
		{Code: `<iframe title="\\" />`, Tsx: true},

		// ============================================================
		// Settings — empty / missing / unusual configurations.
		// ============================================================

		// ---- Empty settings object. ----
		{Code: `<iframe title="x" />`, Tsx: true, Settings: map[string]interface{}{}},

		// ---- Settings with unrelated keys (no 'jsx-a11y'). ----
		{
			Code:     `<iframe title="x" />`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18"}},
		},

		// ---- jsx-a11y is empty map. ----
		{
			Code:     `<iframe title="x" />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{}},
		},

		// ---- jsx-a11y components is nil (treated as missing). ----
		{
			Code:     `<iframe title="x" />`,
			Tsx:      true,
			Settings: map[string]interface{}{"jsx-a11y": map[string]interface{}{"components": nil}},
		},

		// ---- polymorphicAllowList containing non-string entries —
		//      those entries are silently filtered. ----
		{
			Code: `<Bar as="iframe" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"polymorphicPropName":  "as",
					"polymorphicAllowList": []interface{}{42, true, "Foo"},
				},
			},
		},

		// ============================================================
		// Component remap variations — non-cycling chain semantics.
		// ============================================================

		// ---- Components map: MyFrame → "iframe", iframe → "div".
		//      Lookup is single-shot (no chain), so MyFrame becomes
		//      "iframe" and the rule still checks. ----
		{
			Code: `<MyFrame title="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"MyFrame": "iframe",
						"iframe":  "div",
					},
				},
			},
		},

		// ---- Multi-key components — many custom names map to "iframe". ----
		{
			Code: `<VideoEmbed title="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{
						"VideoEmbed": "iframe",
						"MapEmbed":   "iframe",
						"ChartFrame": "iframe",
					},
				},
			},
		},

		// ---- Unrelated component remap leaves <iframe> unchanged. ----
		{
			Code: `<iframe title="x" />`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"jsx-a11y": map[string]interface{}{
					"components": map[string]interface{}{"OtherComp": "img"},
				},
			},
		},

		// ============================================================
		// Deep nesting (3+ levels). The listener walks every JSX
		// node regardless of depth.
		// ============================================================

		// ---- Three levels nested, all with titles. ----
		{Code: `<iframe title="L1"><iframe title="L2"><iframe title="L3" /></iframe></iframe>`, Tsx: true},
		// ---- Four levels nested. ----
		{Code: `<iframe title="A"><iframe title="B"><iframe title="C"><iframe title="D" /></iframe></iframe></iframe>`, Tsx: true},
		// ---- Multiple fragment levels containing iframe. ----
		{Code: `<><><><iframe title="x" /></></></>`, Tsx: true},
		// ---- Conditional inside conditional inside JsxExpression. ----
		{Code: `<div>{cond1 ? (cond2 ? <iframe title="A" /> : <iframe title="B" />) : null}</div>`, Tsx: true},
		// ---- iframe inside arrow inside attribute inside iframe. ----
		{Code: `<iframe title="outer"><X attr={() => <iframe title="inner" />} /></iframe>`, Tsx: true},

		// ============================================================
		// Title value AST shapes — broader real-code coverage.
		// ============================================================

		// ---- Conditional with member-access branches. ----
		{Code: `<iframe title={cond ? props.title : "fallback"} />`, Tsx: true},
		{Code: `<iframe title={cond ? props.title : props.fallback} />`, Tsx: true},

		// ---- Optional-chain with nullish fallback. ----
		{Code: `<iframe title={user?.profile?.name ?? "Anonymous"} />`, Tsx: true},

		// ---- Deep optional chain. ----
		{Code: `<iframe title={a?.b?.c?.d?.title} />`, Tsx: true},

		// ---- Chained calls. ----
		{Code: `<iframe title={getTitle().toUpperCase()} />`, Tsx: true},
		{Code: `<iframe title={user.getProfile().getDisplayName()} />`, Tsx: true},
		{Code: `<iframe title={obj.toString()} />`, Tsx: true},

		// ---- Title from typeof === comparison guard. Counterintuitive
		//      but upstream-aligned: jsx-ast-utils' UnaryExpression
		//      extractor maps `typeof x` to undefined, so the test arm
		//      `typeof title === "string"` evaluates to false → recurse
		//      into the alternate "Default" → typeof string → VALID. ----
		{Code: `<iframe title={typeof title === "string" ? title : "Default"} />`, Tsx: true},

		// ---- Common runtime-coercion patterns. CallExpressions
		//      synthesize strings → typeof string → VALID. ----
		{Code: `<iframe title={Array.isArray(t) ? t.join(", ") : t} />`, Tsx: true},
		{Code: `<iframe title={Object.values(titles)[0]} />`, Tsx: true},
		{Code: `<iframe title={JSON.stringify(meta)} />`, Tsx: true},
		{Code: `<iframe title={String(value)} />`, Tsx: true},
		{Code: `<iframe title={Boolean(x).toString()} />`, Tsx: true},

		// ---- BinaryExpression `+` with various right-hand kinds. ----
		{Code: `<iframe title={"v-" + obj.id} />`, Tsx: true},
		{Code: `<iframe title={obj.prefix + "-" + obj.suffix} />`, Tsx: true},
		{Code: `<iframe title={1 + ""} />`, Tsx: true}, // 1+"" → "1" → string

		// ============================================================
		// Identifier name variations — all non-reserved Identifier
		// names extract to their bare name as a non-empty string.
		// ============================================================

		{Code: `<iframe title={t} />`, Tsx: true},
		{Code: `<iframe title={$t} />`, Tsx: true},
		{Code: `<iframe title={_t} />`, Tsx: true},
		{Code: `<iframe title={t1} />`, Tsx: true},
		{Code: `<iframe title={titleProp} />`, Tsx: true},
		{Code: `<iframe title={title_prop} />`, Tsx: true},
		{Code: `<iframe title={TitleProp} />`, Tsx: true},
		{Code: `<iframe title={async} />`, Tsx: true}, // contextual keyword as identifier

		// ============================================================
		// JSX comment / multi-line variants.
		// ============================================================

		{Code: "<iframe /* before */ title=\"x\" /* after */ />", Tsx: true},
		{Code: "<iframe\n  /* sep */\n  title=\"x\"\n/>", Tsx: true},
		{Code: "<iframe title=\"x\"\n  src=\"https://example.com\"\n  width={500}\n/>", Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// typeof-string gate: non-string TRUTHY values report despite
		// being truthy. This is the defining iframe-has-title behavior
		// that distinguishes it from html-has-lang.
		// ============================================================

		// ---- Non-zero number — truthy but typeof "number" → REPORT.
		//      Locks in the typeof-string gate via the staticEval Number
		//      branch. ----
		{
			Code: `<iframe title={1} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={3.14} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Hex / octal / binary numerics — all jvNumber.
		{
			Code: `<iframe title={0xff} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={1e2} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- BigInt — truthy but typeof "bigint" → REPORT. ----
		{
			Code: `<iframe title={1n} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={123n} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Object literal — truthy but typeof "object" → REPORT. ----
		{
			Code: `<iframe title={{}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={{ value: "x" }} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Array literal — truthy but typeof "object" → REPORT. ----
		{
			Code: `<iframe title={[]} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={["x"]} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Function expression / arrow / class expression — typeof
		//      "function" / "object" → REPORT. ----
		{
			Code: `<iframe title={() => "x"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={function () { return "x"; }} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={class {}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- NewExpression — upstream returns `{}` (empty object), typeof
		//      "object" → REPORT. ----
		{
			Code: `<iframe title={new String("x")} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={new Foo()} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- RegExp literal — typeof "object" → REPORT. ----
		{
			Code: `<iframe title={/regex/} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- JS_RESERVED globals: Math is an object (not a function).
		//      Array / Date / Number / Object / String are constructor
		//      functions (typeof "function"). Infinity is a number.
		//      All fail typeof string. ----
		{
			Code: `<iframe title={Math} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={Array} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={Number} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={Object} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={String} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={Date} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={Infinity} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={-Infinity} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Falsy values — primitive falsy classifications.
		// ============================================================

		// ---- null literal → falsy. ----
		{
			Code: `<iframe title={null} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Numeric 0 → falsy. ----
		{
			Code: `<iframe title={0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Negative zero is falsy.
		{
			Code: `<iframe title={-0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Negative non-zero number — truthy but typeof "number" → REPORT. ----
		{
			Code: `<iframe title={-1} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- String literals "true" / "false" — jsx-ast-utils' Literal
		//      extractor case-insensitively coerces to BOOLEAN. So
		//      `<iframe title="true" />` evaluates to boolean true →
		//      typeof "boolean" → REPORT. Locks in the
		//      jsxAstUtilsLiteralCoerce path inside staticEval. ----
		{
			Code: `<iframe title="true" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title="false" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title="True" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title="TRUE" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title="FALSE" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Empty JsxExpression `{}` — tsgo synthesizes for malformed
		//      input. attributeInnerExpression returns nil →
		//      PropValueIsTruthyString false. ----
		{
			Code: `<iframe title={} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- typeof / void → undefined → not string → REPORT. ----
		{
			Code: `<iframe title={typeof x} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={void 0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- delete → boolean → not string → REPORT. ----
		{
			Code: `<iframe title={delete x.y} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Bang / not operators on title — boolean → REPORT. ----
		{
			Code: `<iframe title={!""} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={!!"x"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Tilde (bitwise not) → number → REPORT. ----
		{
			Code: `<iframe title={~0} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Postfix increment on a non-numeric Identifier evaluates
		//      to NaN (number) → not string → REPORT. ----
		{
			Code: `<iframe title={x++} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- TS-wrapped undefined still falsy — staticEval's
		//      skipTransparent strips `as` / `!` / parens before
		//      checking the inner Identifier. ----
		{
			Code: `<iframe title={undefined as any} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={(undefined)} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- `satisfies` is intentionally opaque — jsx-ast-utils' TYPES
		//      table has no entry for TSSatisfiesExpression, so getPropValue
		//      returns null → not string → REPORT. Locks in that satisfies
		//      is excluded from skipTransparent in staticEval. ----
		{
			Code: `<iframe title={"x" satisfies string} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Position assertions — multi-container line/column lock-ins.
		// ============================================================

		// ---- Bare self-closing covers cols 1..10 (10 chars). ----
		{
			Code: `<iframe />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
			},
		},

		// ---- Indented: two-space indent shifts the report column to 3.
		//      EndColumn is 1-based exclusive end of `<iframe />`. ----
		{
			Code: `  <iframe />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage,
					Line: 1, Column: 3, EndLine: 1, EndColumn: 13},
			},
		},

		// ---- Paired form: report lands on the JsxOpeningElement only —
		//      `<iframe>` is 8 chars, EndColumn 9. ----
		{
			Code: `<iframe></iframe>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
			},
		},

		// ---- Multi-line position: report covers entire opening tag
		//      from `<` to `>` across multiple lines. ----
		{
			Code: "<iframe\n  className=\"x\"\n>\n</iframe>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage,
					Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
			},
		},

		// ============================================================
		// Spread argument variants — non-literal spreads are opaque.
		// ============================================================

		// Spread of CallExpression result.
		{
			Code: `<iframe {...spreadFn()} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Spread of MemberExpression.
		{
			Code: `<iframe {...obj.spread} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Spread of NewExpression.
		{
			Code: `<iframe {...new Foo()} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Literal-key spread: upstream `getProp` has a STRICT
		// `property.key.type === 'Identifier'` guard. String-literal,
		// numeric-literal, and computed-literal keys all return
		// undefined → REPORT. ----
		// ============================================================

		// ---- String-literal key in spread → upstream undefined → REPORT. ----
		{
			Code: `<iframe {...{"title": "x"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Computed key — even StringLiteral inside `[...]` doesn't match. ----
		{
			Code: `<iframe {...{["title"]: "x"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Spread carrying title with falsy / non-string value inside.
		// ============================================================

		// ---- Spread that LITERALLY contains title={undefined}. ----
		{
			Code: `<iframe {...{title: undefined}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Spread carrying title="" inside literal spread. ----
		{
			Code: `<iframe {...{title: ""}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Spread carrying title={42} (non-string truthy). ----
		{
			Code: `<iframe {...{title: 42}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Spread carrying title={true} (non-string truthy). ----
		{
			Code: `<iframe {...{title: true}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// componentMap remap + various negative cases.
		// ============================================================

		// ---- componentMap remap + spread → REPORT (spread is opaque). ----
		{
			Code:     `<FooComponent {...props} />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- componentMap remap + title={false}. ----
		{
			Code:     `<FooComponent title={false} />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- componentMap remap + title={42}. ----
		{
			Code:     `<FooComponent title={42} />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Polymorphic-prop branch invalid case ----
		{
			Code:     `<Foo as="iframe" />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code:     `<Foo as="iframe" title={undefined} />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code:     `<Foo as="iframe" title={42} />`,
			Tsx:      true,
			Settings: polymorphicSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Listener is stateless: each <iframe> in the same file is
		//      independently validated. ----
		{
			Code: `<><iframe /><iframe title="x" /><iframe /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 3},
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 33},
			},
		},

		// ---- Conditional rendering still visits both branches. ----
		{
			Code: `cond ? <iframe /> : <iframe title="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 8},
			},
		},

		// ---- Dimension 2: nested iframe — outer no title, inner no title.
		//      Both fire independently → 2 reports. ----
		{
			Code: `<iframe><iframe /></iframe>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 9},
			},
		},
		// ---- Dimension 2: nested — outer has title, inner missing. ----
		{
			Code: `<iframe title="x"><iframe /></iframe>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 19},
			},
		},

		// ---- Dimension 2: iframe inside JsxExpression child of another
		//      element — listener still fires on the inner iframe. ----
		{
			Code: `<div>{<iframe />}</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 7},
			},
		},

		// ---- Dimension 2: iframe inside an arrow body that is itself the
		//      value of a JsxAttribute (render-prop pattern). ----
		{
			Code: `<X render={() => <iframe />} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 18},
			},
		},

		// ---- Dimension 2: iframe inside JsxFragment — independent fires. ----
		{
			Code: `<><iframe /><iframe /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 3},
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 13},
			},
		},

		// ============================================================
		// Templates with literal-only substitutions: upstream's
		// jsx-ast-utils TemplateLiteral.js extractor falls through to
		// `return ''` for literal substitution types. So `\`${"en"}\``
		// extracts to `""` (empty string) → REPORT. Same for numeric /
		// boolean / null / regex / JSX literals.
		// ============================================================
		// String-literal substitution → "" → falsy.
		{
			Code: "<iframe title={`${\"en\"}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Numeric-literal substitution → "" → falsy.
		{
			Code: "<iframe title={`${0}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Boolean-literal substitution → "" → falsy.
		{
			Code: "<iframe title={`${true}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Null-literal substitution → "" → falsy.
		{
			Code: "<iframe title={`${null}`} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Empty template (no substitutions, no quasi text) → "" → falsy.
		{
			Code: "<iframe title={``} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// Tagged template with empty template → empty inner → "" → falsy.
		{
			Code: "<iframe title={tag``} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Real-world bug patterns
		// ============================================================

		// ---- Common bug: forgot to provide a title at all. ----
		{
			Code: `<iframe src="https://example.com" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Common bug: nullish title with empty fallback. ----
		{
			Code: `<iframe title={undefined ?? ""} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Common bug: title on a parent / sibling element. ----
		{
			Code: `<div title="x"><iframe /></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 16},
			},
		},

		// ---- Functional component returning <iframe /> (no title). ----
		{
			Code: `function Video() { return <iframe />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 27},
			},
		},

		// ---- Arrow component with implicit return of bare iframe. ----
		{
			Code: `const Video = () => <iframe />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 21},
			},
		},

		// ---- Title resolves to non-string at runtime. Identifier `title`
		//      gives a string upstream, but explicit `false`/`true`/number
		//      literals don't. ----
		{
			Code: `<iframe src="https://example.com" title={false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Boolean attribute form `<iframe title />` — upstream's
		//      null-attribute-value path returns boolean true → typeof
		//      "boolean" → REPORT. Not in upstream test file but locks in
		//      the AttributeIsBooleanForm guard in PropValueIsTruthyString. ----
		{
			Code: `<iframe title />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Async / await / yield container patterns. Upstream's
		// extractor has no entry for AwaitExpression / YieldExpression
		// → fallthrough to the `console.error / null` arm → falsy and
		// not a string → REPORT. Locks in the await/yield arms of
		// staticEval which return jsNull.
		// ============================================================

		{
			Code: `<iframe title={await getTitle()} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={await Promise.resolve("x")} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Yield inside generator function body. Listener fires on
		//      the iframe inside the function. ----
		{
			Code: `function* gen() { return <iframe title={yield getTitle()} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 26},
			},
		},

		// ============================================================
		// Numeric literal forms — every numeric notation (octal /
		// binary / hex / float / numeric-separator / exponential /
		// bigint) classifies as typeof "number" or "bigint", never
		// string → REPORT.
		// ============================================================

		{
			Code: `<iframe title={0o10} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={0b10} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={1_000} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={1.5} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={0xffn} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// BigInt zero — falsy AND not string.
		{
			Code: `<iframe title={0n} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Counter-patterns — other a11y attributes do NOT substitute
		// for `title`. The rule strictly inspects the `title` prop;
		// developers occasionally try to "label" the iframe via other
		// attributes which doesn't satisfy the rule's intent.
		// ============================================================

		{
			Code: `<iframe aria-label="Frame label" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe aria-labelledby="label-id" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe aria-describedby="desc-id" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// `name` attribute is a frame-target name, not an accessible label.
		{
			Code: `<iframe name="frame-name" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// `alt` is for img/area/input, not iframe.
		{
			Code: `<iframe alt="alt text" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// `id` doesn't help.
		{
			Code: `<iframe id="frame-id" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// `role` doesn't help.
		{
			Code: `<iframe role="application" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Complex AST shapes that resolve to non-string truthy.
		// ============================================================

		// ---- Comma sequence (SequenceExpression). Upstream returns an
		//      array of values → typeof "object" → REPORT. tsgo
		//      collapses to BinaryExpression with comma op; staticEval
		//      treats it as jvTruthy (sentinel) → not jvString → REPORT. ----
		{
			Code: `<iframe title={(0, "x")} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={(a, b, c)} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Object literal with spread inside. ObjectLiteralExpression
		//      always classifies as truthy non-string regardless of
		//      contents. ----
		{
			Code: `<iframe title={{ ...x }} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={{ ...x, ...y }} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Array literal with elements. ----
		{
			Code: `<iframe title={[1, 2, 3]} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={["a", "b"]} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Generator / async function expressions. ----
		{
			Code: `<iframe title={function* () { yield 1; }} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={async function () { return "x"; }} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={async () => "x"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Computed-key spread variants. Upstream's getProp guards
		// `property.key.type === 'Identifier'` (with `computed: false`
		// implicit). Numeric / template-literal / mixed-case string
		// computed keys all miss the guard → no title found → REPORT.
		// ============================================================

		// ---- Numeric computed key. ----
		{
			Code: `<iframe {...{[1]: "x"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Template literal computed key. ----
		{
			Code: "<iframe {...{[`title`]: \"x\"}} />",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Mixed-case String literal key — upstream's strict guard
		//      misses string-literal keys regardless of case. ----
		{
			Code: `<iframe {...{"TITLE": "x"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Symbol-style computed key. The inner expression is a
		//      MemberExpression+Call, NOT a bare Identifier — upstream's
		//      `property.key.type === 'Identifier'` guard rejects, my
		//      ComputedPropertyName-unwrap also rejects (inner kind is
		//      PropertyAccessExpression). REPORT. ----
		{
			Code: `<iframe {...{[Symbol.for("title")]: "x"}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ---- Computed-Identifier-key spread with falsy / non-string
		//      value. After matching via the new unwrap, the value is
		//      extracted and fails the typeof-string gate. ----
		{
			Code: `<iframe {...{[title]: undefined}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe {...{[title]: 42}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe {...{[title]: ""}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Spread interaction with explicit attrs — left-to-right
		// walk semantics. The FIRST match wins, even if a later
		// match would be valid.
		// ============================================================

		// ---- Literal spread with title=undefined comes BEFORE the
		//      explicit title="x". getProp returns the first match
		//      (the spread's undefined) → REPORT. ----
		{
			Code: `<iframe {...{title: undefined}} title="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Two literal spreads, both with title=undefined. ----
		{
			Code: `<iframe {...{title: undefined}} {...{title: undefined}} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Literal spread title=42 (truthy non-string) BEFORE
		//      explicit title="x". The first-match (42) wins → REPORT. ----
		{
			Code: `<iframe {...{title: 42}} title="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Literal spread with title=false BEFORE explicit. ----
		{
			Code: `<iframe {...{title: false}} title="x" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Multi-iframe scenarios — listener fires independently per
		// element with correct line/column.
		// ============================================================

		// ---- iframe inside .map iteration without title. ----
		{
			Code: `<>{frames.map(f => <iframe key={f.id} src={f.src} />)}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 20},
			},
		},
		// ---- Mixed valid + invalid in fragment. ----
		{
			Code: `<><iframe /><iframe title="ok" /><iframe /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 3},
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 34},
			},
		},
		// ---- Three invalid in a row. ----
		{
			Code: `<><iframe /><iframe /><iframe /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 3},
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 13},
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 23},
			},
		},

		// ============================================================
		// Deep nesting — missing title at various depths.
		// ============================================================

		// ---- iframe deep inside non-iframe wrappers. ----
		{
			Code: `<div><section><article><iframe /></article></section></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 24},
			},
		},
		// ---- Outer iframe valid, inner missing. ----
		{
			Code: `<><iframe title="A"><iframe /></iframe></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 21},
			},
		},
		// ---- Three-level nesting, only middle missing. ----
		{
			Code: `<iframe title="L1"><iframe><iframe title="L3" /></iframe></iframe>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 20},
			},
		},

		// ============================================================
		// Real-world bug shapes worth locking in.
		// ============================================================

		// ---- Empty title via fallback chain. ----
		{
			Code: `<iframe title={null ?? ""} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={undefined ?? null} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- Conditional that resolves to a non-string in chosen branch. ----
		{
			Code: `<iframe title={true ? 42 : "x"} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={false ? "x" : 42} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- LogicalAnd resolves to non-string. ----
		{
			Code: `<iframe title={true && 42} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<iframe title={"x" && false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},

		// ============================================================
		// Multi-line / whitespace position assertions.
		// ============================================================

		// ---- Title attribute split across lines yet still missing
		//      a usable value (empty string). The position covers the
		//      entire opening tag from `<` to `>`. ----
		{
			Code: "<iframe\n  title=\"\"\n/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage,
					Line: 1, Column: 1, EndLine: 3, EndColumn: 3},
			},
		},
		// ---- Tab-only opening tag. ----
		{
			Code: "<iframe\t/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 1},
			},
		},
		// ---- iframe followed immediately by another token. ----
		{
			Code: `<><iframe /></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "iframeHasTitle", Message: errorMessage, Line: 1, Column: 3},
			},
		},
	})
}
