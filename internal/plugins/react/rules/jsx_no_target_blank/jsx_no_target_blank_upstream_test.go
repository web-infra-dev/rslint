package jsx_no_target_blank

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxNoTargetBlankUpstream(t *testing.T) {
	defaultErrors := []rule_tester.InvalidTestCaseError{
		{MessageId: "noTargetBlankWithoutNoreferrer"},
	}
	allowReferrerErrors := []rule_tester.InvalidTestCaseError{
		{MessageId: "noTargetBlankWithoutNoopener"},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoTargetBlankRule, []rule_tester.ValidTestCase{
		// eslint-plugin-react v7.37.5: all 63 valid and 50 invalid cases, in upstream order.
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
	}, []rule_tester.InvalidTestCase{
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
	})
}
