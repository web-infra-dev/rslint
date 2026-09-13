package jsx_no_target_blank

// cspell:words noreferre NOREFE noopen oreferrer noopenernoreferre

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxNoTargetBlankExtras(t *testing.T) {
	defaultErrors := []rule_tester.InvalidTestCaseError{
		{MessageId: "noTargetBlankWithoutNoreferrer"},
	}
	allowReferrerErrors := []rule_tester.InvalidTestCaseError{
		{MessageId: "noTargetBlankWithoutNoopener"},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoTargetBlankRule, []rule_tester.ValidTestCase{
		{Code: `<a href="https://example.com" target="&#x1005f;blank"/>;`, Tsx: true},
		{Code: `<a href="https://example.com" target="_blank" rel="noreferre&#00000000114;"/>;`, Tsx: true},
		// Decode character references only at the direct JSX string boundary.
		{Code: `<a href="https://example.com" target="_blank" rel="noreferre&#114;"/>;`, Tsx: true},
		{Code: `<a href="https://example.com" target="_blank" rel="NOREFE&#82;RER"/>;`, Tsx: true},
		{Code: `<a href="https://example.com" target="_blank" rel="noopen&#101;r"/>;`, Tsx: true, Options: map[string]interface{}{"allowReferrer": true}},
		{Code: `<a href="https://example.com" target={"&#95;blank"}/>;`, Tsx: true},
		{Code: `<a href={"https&#58;//example.com"} target="_blank"/>;`, Tsx: true, Options: map[string]interface{}{"enforceDynamicLinks": "never"}},
		{Code: `<a href="https&amp;colon;//example.com" target="_blank"/>;`, Tsx: true},
		{Code: `<form action="https&#58;//example.com" target="&#95;blank" rel="noreferre&#114;"/>;`, Tsx: true, Options: map[string]interface{}{"forms": true}},
		// Safe tokens must be bounded by literal spaces or the template edges.
		{Code: "<a href={url} target='_blank' rel={`noreferrer ${more}`}/>;", Tsx: true},
		{Code: "<a href={url} target='_blank' rel={`NOREFERRER ${more}`}/>;", Tsx: true},
		{Code: "<a href={url} target='_blank' rel={`noopener ${more}`}/>;", Tsx: true, Options: map[string]interface{}{"allowReferrer": true}},
		{Code: "<a {...props} href={url} target='_blank' rel={`noreferrer ${more}`}/>;", Tsx: true, Options: map[string]interface{}{"warnOnSpreadAttributes": true}},
		// Intentional upstream divergence: a complete token in a later quasi
		// is equally safe; upstream inspects only the first quasi.
		{Code: "<a href={url} target='_blank' rel={`${more} noreferrer`}/>;", Tsx: true},
		{Code: "<a href={url} target='_blank' rel={`${a} noreferrer ${b}`}/>;", Tsx: true},
		{Code: "<a href={url} target='_blank' rel={`${a}${b} noreferrer`}/>;", Tsx: true},
		// Both conditional template branches are provably safe, even though
		// upstream accesses Literal.value and misses TemplateLiteral branches.
		{Code: "<a href={url} target='_blank' rel={ok ? `noreferrer` : `noreferrer`}/>;", Tsx: true},
		{Code: "<a href={url} target='_blank' rel={ok ? `noreferrer ${a}` : `${b} noreferrer`}/>;", Tsx: true},
		{Code: "<a href={url} target={ok ? '_blank' : undefined} rel={ok ? `noreferrer ${a}` : undefined}/>;", Tsx: true},
		// Intentional upstream divergence: links:false must disable link checks.
		{Code: `<a href="https://example.com" target="_blank"/>;`, Tsx: true, Options: map[string]interface{}{"links": false}},
		// Non-link JSX elements are ignored regardless of target value.
		{Code: `<div target="_blank" href="https://example.com"></div>;`, Tsx: true},
		// The link branch does not apply to unknown link components unless
		// configured via settings.
		{Code: `<Link target="_blank" href="https://example.com" />;`, Tsx: true},
		// Member-access tag names (<Foo.Bar>) are not link components under
		// the default configuration — the check skips them entirely.
		{Code: `<Foo.Bar target="_blank" href="https://example.com" />;`, Tsx: true},
		// Namespaced tag — same as above, skipped.
		{Code: `<svg:a target="_blank" href="https://example.com" />;`, Tsx: true},
		// Fragments have no tag name → skipped.
		{Code: `<><a href="/safe">safe</a></>;`, Tsx: true},
		// `target=""` is not "_blank" — the case-insensitive comparison still
		// requires the exact token match.
		{Code: `<a target="" href="https://example.com"></a>;`, Tsx: true},
		// Parenthesized conditional branches still match — AST unwrap reaches
		// the inner literal even when tsgo preserves the paren node.
		{Code: `<a href={href} target={isExternal ? ("_blank") : undefined} rel={isExternal ? ("noreferrer") : undefined} />;`, Tsx: true},
		// Template literal in target is NOT treated as possibly blank —
		// matches upstream's `expr.type === 'Literal'` guard which excludes
		// TemplateLiteral. So this entire element is considered not checked.
		{Code: "<a target={`_blank`} href=\"https://example.com\"></a>;", Tsx: true},
		// Multiple distinct spread attributes with a secure rel after them
		// are still valid under default options.
		{Code: `<a {...a} {...b} target="_blank" href="https://example.com" rel="noreferrer"></a>;`, Tsx: true},
		// Nested anchor inside a component — inner anchor is independently
		// checked. Outer <div> is not a link component.
		{Code: `<div><a href="/safe">s</a></div>;`, Tsx: true},
		// Conditional where BOTH branches are non-_blank strings is not
		// possibly blank, so the element is skipped.
		{Code: `<a href={href} target={isSelf ? "_parent" : "_top"}></a>;`, Tsx: true},
		// Namespaced attribute name for target — upstream compares against
		// `.name.name` which is the local "name" object for JSXNamespacedName,
		// so it never equals 'target'. The rule treats `a:target` as an
		// unrelated attribute and skips the element.
		{Code: `<a a:target="_blank" href="https://example.com"></a>;`, Tsx: true},
		// Custom link component with linkAttribute: 'to', but the actual
		// `href` attribute is the one set externally — the rule only scans
		// the configured attribute, so this is valid.
		{
			Code:    `<Link target="_blank" href="https://example.com" to="/safe" rel="noreferrer"></Link>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "always"},
			Settings: map[string]interface{}{
				"linkComponents": map[string]interface{}{"name": "Link", "linkAttribute": "to"},
			},
		},
		// Default linkComponents plus a settings-configured one — both kinds
		// of elements are checked in the same pass.
		{
			Code:     `<a href="/safe" target="_blank" rel="noreferrer"></a>;`,
			Tsx:      true,
			Settings: map[string]interface{}{"linkComponents": []interface{}{"Link"}},
		},
		// Custom form component with `formAttribute` (NOT `linkAttribute`) —
		// regression for a field-name mismatch that silently fell back to
		// the default "action". Here the configured attribute `endpoint`
		// carries the external URL so the form must be checked; with a
		// secure rel it stays valid.
		{
			Code:    `<MyForm target="_blank" endpoint="https://example.com" rel="noopener noreferrer"></MyForm>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true},
			Settings: map[string]interface{}{
				"formComponents": map[string]interface{}{"name": "MyForm", "formAttribute": "endpoint"},
			},
		},
		// Custom component configured via `settings.linkComponents` with
		// multiple link attributes — `linkAttribute: ['to', 'href']`. The
		// rule scans both; here neither carries an external URL and href is
		// absent, so the element is valid.
		{
			Code: `<MultiLink target="_blank" to="/internal" rel="noreferrer"></MultiLink>;`,
			Tsx:  true,
			Settings: map[string]interface{}{
				"linkComponents": map[string]interface{}{"name": "MultiLink", "linkAttribute": []interface{}{"to", "href"}},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Espree truncates non-BMP numeric entities; preserve the JSX runtime value instead.
		{Code: `<a href="https://example.com" target="_blank" rel="&#x1006e;oreferrer"/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="_blank" rel="𐁮oreferrer noreferrer"/>;`}},
		{Code: `<a href="https://example.com" target="_blank" rel="&#x1F600;"/>;`, Tsx: true, Errors: defaultErrors, Output: []string{"<a href=\"https://example.com\" target=\"_blank\" rel=\"\U0001F600 noreferrer\"/>;"}},
		{Code: `<a href="https://example.com" target="&#00000000095;blank"/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="&#00000000095;blank" rel="noreferrer"/>;`}},
		{Code: `<a href="https://example.com" target="_blank" rel="&#00000000095;blank"/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="_blank" rel="_blank noreferrer"/>;`}},
		// Upstream emits unescaped expression values, which can change values or break syntax.
		{Code: `<a href="https://example.com" target="&#95;blank" rel={'a"b'}/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="&#95;blank" rel={"a\"b noreferrer"}/>;`}},
		{Code: `<a href="https://example.com" target="_blank" rel={'a\\b'}/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="_blank" rel={"a\\b noreferrer"}/>;`}},
		{Code: `<a href="https://example.com" target="_blank" rel={'a\nb'}/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="_blank" rel={"a\nb noreferrer"}/>;`}},
		{Code: `<a href="https://example.com" target="_blank" rel={'\uD800'}/>;`, Tsx: true, Errors: defaultErrors},
		{Code: `<a href="https://example.com" target="_blank" rel="&#xD800;"/>;`, Tsx: true, Errors: defaultErrors},
		// JSX entities participate in target/URL detection and rel fixes.
		{Code: `<a href="https://example.com" target="&#95;blank"/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="&#95;blank" rel="noreferrer"/>;`}},
		{Code: `<a href="https&#58;//example.com" target="_blank"/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https&#58;//example.com" target="_blank" rel="noreferrer"/>;`}},
		{Code: `<a href="https://example.com" target="_blank" rel="noopenernoreferre&#114;"/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="_blank" rel="noopener noreferrer"/>;`}},
		// Do not decode JavaScript expression strings or template quasis.
		{Code: `<a href="https://example.com" target="_blank" rel={"noreferre&#114;"}/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="_blank" rel={"noreferre&#114; noreferrer"}/>;`}},
		{Code: "<a href={url} target='_blank' rel={`noreferre&#114; ${more}`}/>;", Tsx: true, Errors: defaultErrors},
		// Decoded quotes and ampersands must be escaped on output. Upstream
		// inserts decoded text verbatim, yielding invalid JSX or decoding twice.
		{Code: `<a href="https://example.com" target="_blank" rel="&quot;&amp;quot;"/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="_blank" rel="&quot;&amp;quot; noreferrer"/>;`}},
		{Code: `<a href="https://example.com" target="_blank" rel="&amp;#114;"/>;`, Tsx: true, Errors: defaultErrors, Output: []string{`<a href="https://example.com" target="_blank" rel="&amp;#114; noreferrer"/>;`}},
		{Code: `<form action="https&#58;//example.com" target="&#95;blank"/>;`, Tsx: true, Options: map[string]interface{}{"forms": true}, Errors: defaultErrors},
		// Unlike upstream, do not trust a token that a substitution can extend.
		{Code: "<a href={url} target='_blank' rel={`noreferrer${more}`}/>;", Tsx: true, Errors: defaultErrors},
		{Code: "<a href={url} target='_blank' rel={`noopener${more}`}/>;", Tsx: true, Options: map[string]interface{}{"allowReferrer": true}, Errors: allowReferrerErrors},
		{Code: "<a href={url} target='_blank' rel={`${more}noreferrer`}/>;", Tsx: true, Errors: defaultErrors},
		{Code: "<a href={url} target='_blank' rel={`${a}noreferrer${b}`}/>;", Tsx: true, Errors: defaultErrors},
		{Code: "<a href={url} target='_blank' rel={`${a} noreferrer${b}`}/>;", Tsx: true, Errors: defaultErrors},
		{Code: "<a href={url} target='_blank' rel={`${a}noreferrer ${b}`}/>;", Tsx: true, Errors: defaultErrors},
		{Code: "<a href={url} target='_blank' rel={`noreferrer-extra ${more}`}/>;", Tsx: true, Errors: defaultErrors},
		{Code: "<a href={url} target='_blank' rel={`noopener ${more}`}/>;", Tsx: true, Errors: defaultErrors},
		{Code: "<a href={url} target='_blank' rel={`noreferrer\\t${more}`}/>;", Tsx: true, Errors: defaultErrors},
		{Code: "<a href={url} target='_blank' rel={ok ? `noreferrer ${a}` : `noreferrer${b}`}/>;", Tsx: true, Errors: defaultErrors},
		{Code: "<a href={url} target='_blank' rel={`noreferrer ${more}`} {...props}/>;", Tsx: true, Options: map[string]interface{}{"warnOnSpreadAttributes": true}, Errors: defaultErrors},
		{Code: "<form action={url} target='_blank' rel={`noopener ${more}`}/>;", Tsx: true, Options: map[string]interface{}{"forms": true, "allowReferrer": true}, Errors: allowReferrerErrors},
		// undefined is an Identifier in ESTree, not a literal, and may be
		// shadowed. Neither the global name nor any binding permits this fix.
		{Code: `<a href="https://example.com" target="_blank" rel={undefined}/>;`, Tsx: true, Errors: defaultErrors},
		{Code: `function C(undefined) { return <a href="https://example.com" target="_blank" rel={undefined}/>; } C("nofollow");`, Tsx: true, Errors: defaultErrors},
		{Code: `function C() { const undefined = "nofollow"; return <a href="https://example.com" target="_blank" rel={undefined}/>; }`, Tsx: true, Errors: defaultErrors},
		// Different member tests have no Identifier.name. Upstream compares
		// two absent names as equal and incorrectly trusts an unrelated branch.
		{Code: `<a href={url} target={a.x ? "_blank" : undefined} rel={b.x ? "noreferrer" : undefined}/>;`, Tsx: true, Errors: defaultErrors},
		// Intentional upstream divergence: allowReferrer fixes must add
		// noopener, not noreferrer (which would suppress the requested referrer).
		{Code: `<a href="https://example.com" target="_blank" rel="nofollow"/>;`, Tsx: true, Options: map[string]interface{}{"allowReferrer": true}, Errors: allowReferrerErrors, Output: []string{`<a href="https://example.com" target="_blank" rel="nofollow noopener"/>;`}},
		// Multi-line attributes — the reported range still covers the full
		// opening element across lines.
		{
			Code: "<a\n  target=\"_blank\"\n  href=\"https://example.com/ml\"\n></a>;",
			Tsx:  true,
			Output: []string{
				"<a\n  target=\"_blank\"\n  href=\"https://example.com/ml\" rel=\"noreferrer\"\n></a>;",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noTargetBlankWithoutNoreferrer", Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		// Self-closing anchor form — same semantics as non-self-closing for
		// the diagnostic; the reported range ends after the `/>`.
		{
			Code:   `<a target="_blank" href="https://example.com/sc" />;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href="https://example.com/sc" rel="noreferrer" />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noTargetBlankWithoutNoreferrer", Line: 1, Column: 1, EndLine: 1, EndColumn: 52},
			},
		},
		// Regression: parens around the whole conditional target expression
		// must not hide the `_blank` branch from detection. Before the
		// SkipParentheses fix, this was a silent false-negative.
		{
			Code:   `<a href="https://example.com/paren" target={(isExternal ? "_blank" : undefined)}></a>;`,
			Tsx:    true,
			Output: []string{`<a href="https://example.com/paren" target={(isExternal ? "_blank" : undefined)} rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		// Regression: parens around a direct string literal inside the JSX
		// expression container. Same failure mode as above.
		{
			Code:   `<a target={("_blank")} href="https://example.com/paren2"></a>;`,
			Tsx:    true,
			Output: []string{`<a target={("_blank")} href="https://example.com/paren2" rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		// Regression: parens around the rel conditional — secure-rel
		// extraction must unwrap parens on both target and rel to find the
		// matched-test shortcut, otherwise the conditional was treated as
		// non-literal and wrongly reported.
		{
			Code:   `<a href="https://example.com/relp" target="_blank" rel={(getRel())}></a>;`,
			Tsx:    true,
			Errors: defaultErrors,
		},
		// Case-insensitive rel token matching: the rule splits on whitespace
		// and compares case-insensitively, so the single-word joined form
		// "NOOPENERNOREFERRER" does NOT contain "noreferrer" as a token and
		// is reported. Matches upstream.
		{
			Code:   `<a target="_blank" href="https://example.com/ci" rel="NOOPENERNOREFERRER"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href="https://example.com/ci" rel="NOOPENERNOREFERRER noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		// Three-way conditional where the _blank branch is unguarded by the
		// rel's matched test — we fall back to examining every rel branch
		// and report when any branch is non-secure.
		{
			Code:   `<a href={href} target="_blank" rel={isExternal ? "noreferrer" : isInternal ? "noreferrer" : "nothing"} />;`,
			Tsx:    true,
			Errors: defaultErrors,
		},
		// Custom link component whose configured linkAttribute also happens
		// to be set to an external URL — still reported.
		{
			Code:     `<Link target="_blank" to="https://example.com/link"></Link>;`,
			Tsx:      true,
			Output:   []string{`<Link target="_blank" to="https://example.com/link" rel="noreferrer"></Link>;`},
			Options:  map[string]interface{}{"enforceDynamicLinks": "always"},
			Settings: map[string]interface{}{"linkComponents": map[string]interface{}{"name": "Link", "linkAttribute": "to"}},
			Errors:   defaultErrors,
		},
		// Two spreads, target/rel around them — spread position still
		// matters: rel BEFORE a spread with warnOnSpread true is not trusted.
		{
			Code:    `<a {...a} target="_blank" rel="noreferrer" {...b}></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"warnOnSpreadAttributes": true},
			Errors:  defaultErrors,
		},
		// Regression: custom form component using `formAttribute` with a
		// dynamic URL — forms don't autofix; the report still fires.
		{
			Code:    `<MyForm target="_blank" endpoint={url}></MyForm>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true},
			Settings: map[string]interface{}{
				"formComponents": map[string]interface{}{"name": "MyForm", "formAttribute": "endpoint"},
			},
			Errors: defaultErrors,
		},
		// Message text assertion — locks in the exact diagnostic string for
		// the default (noreferrer) message, matching upstream verbatim.
		{
			Code:   `<a target="_blank" href="https://example.com/msg"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href="https://example.com/msg" rel="noreferrer"></a>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noTargetBlankWithoutNoreferrer",
					Message:   `Using target="_blank" without rel="noreferrer" (which implies rel="noopener") is a security risk in older browsers: see https://mathiasbynens.github.io/rel-noopener/#recommendations`,
				},
			},
		},
		// Message text assertion — locks in the exact diagnostic string for
		// the allowReferrer (noopener) message.
		{
			Code:    `<a target="_blank" href="https://example.com/msg2"></a>;`,
			Tsx:     true,
			Output:  []string{`<a target="_blank" href="https://example.com/msg2" rel="noopener"></a>;`},
			Options: map[string]interface{}{"allowReferrer": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noTargetBlankWithoutNoopener",
					Message:   `Using target="_blank" without rel="noreferrer" or rel="noopener" (the former implies the latter and is preferred due to wider support) is a security risk: see https://mathiasbynens.github.io/rel-noopener/#recommendations`,
				},
			},
		},
	})
}
