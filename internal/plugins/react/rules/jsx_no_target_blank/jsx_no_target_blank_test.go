package jsx_no_target_blank

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxNoTargetBlankRule(t *testing.T) {
	defaultErrors := []rule_tester.InvalidTestCaseError{
		{MessageId: "noTargetBlankWithoutNoreferrer"},
	}
	allowReferrerErrors := []rule_tester.InvalidTestCaseError{
		{MessageId: "noTargetBlankWithoutNoopener"},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoTargetBlankRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `<a href="foobar"></a>;`, Tsx: true},
		{Code: `<a randomTag></a>;`, Tsx: true},
		{Code: `<a target />;`, Tsx: true},
		{Code: `<a href="foobar" target="_blank" rel="noopener noreferrer"></a>;`, Tsx: true},
		{Code: `<a href="foobar" target="_blank" rel="noreferrer"></a>;`, Tsx: true},
		{Code: `<a href="foobar" target="_blank" rel={"noopener noreferrer"}></a>;`, Tsx: true},
		{Code: `<a href="foobar" target="_blank" rel={"noreferrer"}></a>;`, Tsx: true},
		{Code: `<a href={"foobar"} target={"_blank"} rel={"noopener noreferrer"}></a>;`, Tsx: true},
		{Code: `<a href={"foobar"} target={"_blank"} rel={"noreferrer"}></a>;`, Tsx: true},
		{Code: `<a href={'foobar'} target={'_blank'} rel={'noopener noreferrer'}></a>;`, Tsx: true},
		{Code: `<a href={'foobar'} target={'_blank'} rel={'noreferrer'}></a>;`, Tsx: true},
		{Code: "<a href={`foobar`} target={`_blank`} rel={`noopener noreferrer`}></a>;", Tsx: true},
		{Code: "<a href={`foobar`} target={`_blank`} rel={`noreferrer`}></a>;", Tsx: true},
		{Code: `<a target="_blank" {...spreadProps} rel="noopener noreferrer"></a>;`, Tsx: true},
		{Code: `<a target="_blank" {...spreadProps} rel="noreferrer"></a>;`, Tsx: true},
		{Code: `<a {...spreadProps} target="_blank" rel="noopener noreferrer" href="https://example.com">s</a>;`, Tsx: true},
		{Code: `<a {...spreadProps} target="_blank" rel="noreferrer" href="https://example.com">s</a>;`, Tsx: true},
		{Code: `<a target="_blank" rel="noopener noreferrer" {...spreadProps}></a>;`, Tsx: true},
		{Code: `<a target="_blank" rel="noreferrer" {...spreadProps}></a>;`, Tsx: true},
		{Code: `<p target="_blank"></p>;`, Tsx: true},
		{Code: `<a href="foobar" target="_BLANK" rel="NOOPENER noreferrer"></a>;`, Tsx: true},
		{Code: `<a href="foobar" target="_BLANK" rel="NOREFERRER"></a>;`, Tsx: true},
		{Code: `<a target="_blank" rel={relValue}></a>;`, Tsx: true},
		{Code: `<a target={targetValue} rel="noopener noreferrer"></a>;`, Tsx: true},
		{Code: `<a target={targetValue} rel="noreferrer"></a>;`, Tsx: true},
		{Code: `<a target={targetValue} rel={"noopener noreferrer"}></a>;`, Tsx: true},
		{Code: `<a target={targetValue} rel={"noreferrer"}></a>;`, Tsx: true},
		{Code: `<a target={targetValue} href="relative/path"></a>;`, Tsx: true},
		{Code: `<a target={targetValue} href="/absolute/path"></a>;`, Tsx: true},
		{Code: `<a target={'targetValue'} href="/absolute/path"></a>;`, Tsx: true},
		{Code: `<a target={"targetValue"} href="/absolute/path"></a>;`, Tsx: true},
		{Code: `<a target={null} href="//example.com"></a>;`, Tsx: true},
		{
			Code:    `<a {...someObject} href="/absolute/path"></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "always", "warnOnSpreadAttributes": true},
		},
		{
			Code:    `<a {...someObject} rel="noreferrer"></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "always", "warnOnSpreadAttributes": true},
		},
		{
			Code:    `<a {...someObject} rel="noreferrer" target="_blank"></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "always", "warnOnSpreadAttributes": true},
		},
		{
			Code:    `<a {...someObject} href="foobar" target="_blank"></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "always", "warnOnSpreadAttributes": true},
		},
		{
			Code:    `<a target="_blank" href={ dynamicLink }></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "never"},
		},
		{
			Code:    `<a target={"_blank"} href={ dynamicLink }></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "never"},
		},
		{
			Code:    `<a target={'_blank'} href={ dynamicLink }></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "never"},
		},
		{
			Code:     `<Link target="_blank" href={ dynamicLink }></Link>;`,
			Tsx:      true,
			Options:  map[string]interface{}{"enforceDynamicLinks": "never"},
			Settings: map[string]interface{}{"linkComponents": []interface{}{"Link"}},
		},
		{
			Code:    `<Link target="_blank" to={ dynamicLink }></Link>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "never"},
			Settings: map[string]interface{}{
				"linkComponents": map[string]interface{}{"name": "Link", "linkAttribute": "to"},
			},
		},
		{
			Code:    `<Link target="_blank" to={ dynamicLink }></Link>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "never"},
			Settings: map[string]interface{}{
				"linkComponents": map[string]interface{}{"name": "Link", "linkAttribute": []interface{}{"to"}},
			},
		},
		{
			Code:    `<a href="foobar" target="_blank" rel="noopener"></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowReferrer": true},
		},
		{
			Code:    `<a href="foobar" target="_blank" rel="noreferrer"></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowReferrer": true},
		},
		{Code: `<a target={3} />;`, Tsx: true},
		{Code: `<a href="some-link" {...otherProps} target="some-non-blank-target"></a>;`, Tsx: true},
		{Code: `<a href="some-link" target="some-non-blank-target" {...otherProps}></a>;`, Tsx: true},
		{
			Code:    `<a target="_blank" href="/absolute/path"></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": false},
		},
		{
			Code:    `<a target="_blank" href="/absolute/path"></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": false, "links": true},
		},
		{Code: `<form action="https://example.com" target="_blank"></form>;`, Tsx: true},
		{
			Code:    `<form action="https://example.com" target="_blank" rel="noopener noreferrer"></form>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true},
		},
		{
			Code:    `<form action="https://example.com" target="_blank" rel="noopener noreferrer"></form>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true, "links": false},
		},
		{Code: `<a href target="_blank"/>;`, Tsx: true},
		{Code: `<a href={href} target={isExternal ? "_blank" : undefined} rel="noopener noreferrer" />;`, Tsx: true},
		{Code: `<a href={href} target={isExternal ? undefined : "_blank"} rel={isExternal ? "noreferrer" : "noopener noreferrer"} />;`, Tsx: true},
		{Code: `<a href={href} target={isExternal ? undefined : "_blank"} rel={isExternal ? "noreferrer noopener" : "noreferrer"} />;`, Tsx: true},
		{
			Code:    `<a href={href} target="_blank" rel={isExternal ? "noreferrer" : "noopener"} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowReferrer": true},
		},
		{Code: `<a href={href} target={isExternal ? "_blank" : undefined} rel={isExternal ? "noreferrer" : undefined} />;`, Tsx: true},
		{Code: `<a href={href} target={isSelf ? "_self" : "_blank"} rel={isSelf ? undefined : "noreferrer"} />;`, Tsx: true},
		{Code: `<a href={href} target={isSelf ? "_self" : ""} rel={isSelf ? undefined : ""} />;`, Tsx: true},
		{Code: `<a href={href} target={isExternal ? "_blank" : undefined} rel={isExternal ? "noopener noreferrer" : undefined} />;`, Tsx: true},
		{
			Code:    `<form action={action} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true},
		},
		{
			Code:    `<form action={action} {...spread} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true},
		},

		// ---- Additional edge cases ----
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
		// ---- Upstream invalid cases ----
		{
			Code:   `<a target="_blank" href="https://example.com/1"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href="https://example.com/1" rel="noreferrer"></a>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noTargetBlankWithoutNoreferrer", Line: 1, Column: 1, EndLine: 1, EndColumn: 49},
			},
		},
		{
			Code:   `<a target="_blank" rel="" href="https://example.com/2"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" rel="noreferrer" href="https://example.com/2"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" rel={0} href="https://example.com/3"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" rel="noreferrer" href="https://example.com/3"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" rel={1} href="https://example.com/3"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" rel="noreferrer" href="https://example.com/3"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" rel={false} href="https://example.com/4"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" rel="noreferrer" href="https://example.com/4"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" rel={null} href="https://example.com/5"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" rel="noreferrer" href="https://example.com/5"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" rel="noopenernoreferrer" href="https://example.com/6"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" rel="noopener noreferrer" href="https://example.com/6"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" rel="no referrer" href="https://example.com/7"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" rel="no referrer noreferrer" href="https://example.com/7"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_BLANK" href="https://example.com/8"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_BLANK" href="https://example.com/8" rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" href="//example.com/9"></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href="//example.com/9" rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" href="//example.com/10" rel={true}></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href="//example.com/10" rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" href="//example.com/11" rel={3}></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href="//example.com/11" rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" href="//example.com/12" rel={null}></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href="//example.com/12" rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		// Non-literal rel expression — no autofix (same as upstream: fix
		// returns null, diagnostic still reports).
		{
			Code:   `<a target="_blank" href="//example.com/13" rel={getRel()}></a>;`,
			Tsx:    true,
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" href="//example.com/14" rel={"noopenernoreferrer"}></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href="//example.com/14" rel={"noopener noreferrer"}></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target={"_blank"} href={"//example.com/15"} rel={"noopenernoreferrer"}></a>;`,
			Tsx:    true,
			Output: []string{`<a target={"_blank"} href={"//example.com/15"} rel={"noopener noreferrer"}></a>;`},
			Errors: defaultErrors,
		},
		{
			// cspell:disable-next-line
			Code:   `<a target={"_blank"} href={"//example.com/16"} rel={"noopenernoreferrernoreferrernoreferrernoreferrernoreferrer"}></a>;`,
			Tsx:    true,
			Output: []string{`<a target={"_blank"} href={"//example.com/16"} rel={"noopener noreferrer"}></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" href="//example.com/17" rel></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href="//example.com/17" rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target="_blank" href={ dynamicLink }></a>;`,
			Tsx:    true,
			Output: []string{`<a target="_blank" href={ dynamicLink } rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target={'_blank'} href="//example.com/18"></a>;`,
			Tsx:    true,
			Output: []string{`<a target={'_blank'} href="//example.com/18" rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:   `<a target={"_blank"} href="//example.com/19"></a>;`,
			Tsx:    true,
			Output: []string{`<a target={"_blank"} href="//example.com/19" rel="noreferrer"></a>;`},
			Errors: defaultErrors,
		},
		{
			Code:    `<a href="https://example.com/20" target="_blank" rel></a>;`,
			Tsx:     true,
			Output:  []string{`<a href="https://example.com/20" target="_blank" rel="noopener"></a>;`},
			Options: map[string]interface{}{"allowReferrer": true},
			Errors:  allowReferrerErrors,
		},
		{
			Code:    `<a href="https://example.com/20" target="_blank"></a>;`,
			Tsx:     true,
			Output:  []string{`<a href="https://example.com/20" target="_blank" rel="noopener"></a>;`},
			Options: map[string]interface{}{"allowReferrer": true},
			Errors:  allowReferrerErrors,
		},
		{
			Code:    `<a target="_blank" href={ dynamicLink }></a>;`,
			Tsx:     true,
			Output:  []string{`<a target="_blank" href={ dynamicLink } rel="noreferrer"></a>;`},
			Options: map[string]interface{}{"enforceDynamicLinks": "always"},
			Errors:  defaultErrors,
		},
		{
			Code:    `<a {...someObject}></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "always", "warnOnSpreadAttributes": true},
			Errors:  defaultErrors,
		},
		{
			Code:    `<a {...someObject} target="_blank"></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "always", "warnOnSpreadAttributes": true},
			Errors:  defaultErrors,
		},
		{
			Code:    `<a href="foobar" {...someObject} target="_blank"></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "always", "warnOnSpreadAttributes": true},
			Errors:  defaultErrors,
		},
		{
			Code:    `<a href="foobar" target="_blank" rel="noreferrer" {...someObject}></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "always", "warnOnSpreadAttributes": true},
			Errors:  defaultErrors,
		},
		{
			Code:    `<a href="foobar" target="_blank" {...someObject}></a>;`,
			Tsx:     true,
			Options: map[string]interface{}{"enforceDynamicLinks": "always", "warnOnSpreadAttributes": true},
			Errors:  defaultErrors,
		},
		{
			Code:     `<Link target="_blank" href={ dynamicLink }></Link>;`,
			Tsx:      true,
			Output:   []string{`<Link target="_blank" href={ dynamicLink } rel="noreferrer"></Link>;`},
			Options:  map[string]interface{}{"enforceDynamicLinks": "always"},
			Settings: map[string]interface{}{"linkComponents": []interface{}{"Link"}},
			Errors:   defaultErrors,
		},
		{
			Code:    `<Link target="_blank" to={ dynamicLink }></Link>;`,
			Tsx:     true,
			Output:  []string{`<Link target="_blank" to={ dynamicLink } rel="noreferrer"></Link>;`},
			Options: map[string]interface{}{"enforceDynamicLinks": "always"},
			Settings: map[string]interface{}{
				"linkComponents": map[string]interface{}{"name": "Link", "linkAttribute": "to"},
			},
			Errors: defaultErrors,
		},
		{
			Code:    `<a href="some-link" {...otherProps} target="some-non-blank-target"></a>;`,
			Tsx:     true,
			Errors:  defaultErrors,
			Options: map[string]interface{}{"warnOnSpreadAttributes": true},
		},
		{
			Code:    `<a href="some-link" target="some-non-blank-target" {...otherProps}></a>;`,
			Tsx:     true,
			Errors:  defaultErrors,
			Options: map[string]interface{}{"warnOnSpreadAttributes": true},
		},
		{
			Code:    `<a target="_blank" href="//example.com" rel></a>;`,
			Tsx:     true,
			Output:  []string{`<a target="_blank" href="//example.com" rel="noreferrer"></a>;`},
			Options: map[string]interface{}{"links": true},
			Errors:  defaultErrors,
		},
		{
			Code:    `<a target="_blank" href="//example.com" rel></a>;`,
			Tsx:     true,
			Output:  []string{`<a target="_blank" href="//example.com" rel="noreferrer"></a>;`},
			Options: map[string]interface{}{"links": true, "forms": true},
			Errors:  defaultErrors,
		},
		{
			Code:    `<a target="_blank" href="//example.com" rel></a>;`,
			Tsx:     true,
			Output:  []string{`<a target="_blank" href="//example.com" rel="noreferrer"></a>;`},
			Options: map[string]interface{}{"links": true, "forms": false},
			Errors:  defaultErrors,
		},
		{
			Code:    `<form method="POST" action="https://example.com" target="_blank"></form>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true},
			Errors:  defaultErrors,
		},
		{
			Code:    `<form method="POST" action="https://example.com" rel="" target="_blank"></form>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true},
			Errors:  defaultErrors,
		},
		{
			Code:    `<form method="POST" action="https://example.com" rel="noopenernoreferrer" target="_blank"></form>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true},
			Errors:  defaultErrors,
		},
		{
			Code:    `<form method="POST" action="https://example.com" rel="noopenernoreferrer" target="_blank"></form>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true, "links": false},
			Errors:  defaultErrors,
		},
		{
			Code:   `<a href={href} target="_blank" rel={isExternal ? "undefined" : "undefined"} />;`,
			Tsx:    true,
			Errors: defaultErrors,
		},
		{
			Code:   `<a href={href} target="_blank" rel={isExternal ? "noopener" : undefined} />;`,
			Tsx:    true,
			Errors: defaultErrors,
		},
		{
			Code:   `<a href={href} target="_blank" rel={isExternal ? "undefined" : "noopener"} />;`,
			Tsx:    true,
			Errors: defaultErrors,
		},
		{
			Code:   `<a href={href} target={isExternal ? "_blank" : undefined} rel={isExternal ? undefined : "noopener noreferrer"} />;`,
			Tsx:    true,
			Errors: defaultErrors,
		},
		{
			Code:   `<a href={href} target="_blank" rel={isExternal ? 3 : "noopener noreferrer"} />;`,
			Tsx:    true,
			Errors: defaultErrors,
		},
		{
			Code:   `<a href={href} target="_blank" rel={isExternal ? "noopener noreferrer" : "3"} />;`,
			Tsx:    true,
			Errors: defaultErrors,
		},
		{
			Code:    `<a href={href} target="_blank" rel={isExternal ? "noopener" : "2"} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowReferrer": true},
			Errors:  allowReferrerErrors,
		},
		{
			Code:    `<form action={action} target="_blank" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowReferrer": true, "forms": true},
			Errors:  allowReferrerErrors,
		},
		{
			Code:    `<form action={action} target="_blank" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true},
			Errors:  defaultErrors,
		},
		{
			Code:    `<form action={action} {...spread} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forms": true, "warnOnSpreadAttributes": true},
			Errors:  defaultErrors,
		},

		// ---- Additional edge cases ----
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
