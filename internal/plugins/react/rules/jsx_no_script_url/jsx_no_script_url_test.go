package jsx_no_script_url

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxNoScriptUrl(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoScriptUrlRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `<a href="https://reactjs.org"></a>`, Tsx: true},
		{Code: `<a href="mailto:foo@bar.com"></a>`, Tsx: true},
		{Code: `<a href="#"></a>`, Tsx: true},
		{Code: `<a href=""></a>`, Tsx: true},
		{Code: `<a name="foo"></a>`, Tsx: true},
		// Expression container — upstream only checks Literal, not JSXExpressionContainer
		{Code: `<a href={"javascript:"}></a>`, Tsx: true},
		// User component without config — not matched
		{Code: `<Foo href="javascript:"></Foo>`, Tsx: true},
		// Boolean shorthand — no value
		{Code: `<a href />`, Tsx: true},
		// Custom component with non-matching attribute name
		{
			Code:    `<Foo other="javascript:"></Foo>`,
			Options: []interface{}{map[string]interface{}{"name": "Foo", "props": []interface{}{"to", "href"}}},
			Tsx:     true,
		},
		// Settings cases — includeFromSettings defaults to false
		{
			Code: `<Foo href="javascript:"></Foo>`,
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "Foo", "linkAttribute": []interface{}{"to", "href"}},
				},
			},
			Tsx: true,
		},
		// includeFromSettings explicitly false — settings not read
		{
			Code:    `<Foo href="javascript:"></Foo>`,
			Options: map[string]interface{}{"includeFromSettings": false},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "Foo", "linkAttribute": []interface{}{"to", "href"}},
				},
			},
			Tsx: true,
		},
		// includeFromSettings true, but attribute doesn't match
		{
			Code:    `<Foo other="javascript:"></Foo>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "Foo", "linkAttribute": []interface{}{"to", "href"}},
				},
			},
			Tsx: true,
		},

		// Empty legacy array + includeFromSettings false — settings not used
		{
			Code: `<Foo href="javascript:"></Foo>`,
			Options: []interface{}{
				[]interface{}{},
				map[string]interface{}{"includeFromSettings": false},
			},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "Foo", "linkAttribute": []interface{}{"to", "href"}},
				},
			},
			Tsx: true,
		},
		// Empty legacy array + includeFromSettings true, non-matching attribute
		{
			Code: `<Foo other="javascript:"></Foo>`,
			Options: []interface{}{
				[]interface{}{},
				map[string]interface{}{"includeFromSettings": true},
			},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "Foo", "linkAttribute": []interface{}{"to", "href"}},
				},
			},
			Tsx: true,
		},
		// Legacy option redefines "a" with only "to" — "href" is replaced, not appended
		{
			Code:    `<a href="javascript:"></a>`,
			Options: []interface{}{map[string]interface{}{"name": "a", "props": []interface{}{"to"}}},
			Tsx:     true,
		},

		// ---- Additional edge cases ----
		// Template literal — not a Literal node, not flagged
		{Code: "<a href={`javascript:`}></a>", Tsx: true},
		// Number value — not a string literal
		{Code: `<a href={0}></a>`, Tsx: true},
		// Null/undefined — no value
		{Code: `<a href={null}></a>`, Tsx: true},
		// Member expression tag — upstream doesn't match these
		{Code: `<Foo.Bar href="javascript:"></Foo.Bar>`, Tsx: true},
		// Spread attribute — not a JsxAttribute
		{Code: `const x: any = {href: "javascript:"}; <a {...x}></a>;`, Tsx: true},
		// Non-javascript protocol that starts with "j"
		{Code: `<a href="jot:something"></a>`, Tsx: true},
		// Close but not javascript: protocol
		{Code: `<a href="javas:void(0)"></a>`, Tsx: true},
		// Uppercase element name — config is case-sensitive, "A" ≠ "a"
		{Code: `<A href="javascript:"></A>`, Tsx: true},
		// Paren-wrapped expression container — NOT a direct StringLiteral
		{Code: `<a href={("javascript:")}></a>`, Tsx: true},
		// Settings override "a" with only "to" — "href" is replaced
		{
			Code:    `<a href="javascript:"></a>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "a", "linkAttribute": "to"},
				},
			},
			Tsx: true,
		},
		// Settings as single object (not array)
		{
			Code:    `<Foo other="javascript:"></Foo>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": map[string]interface{}{"name": "Foo", "linkAttribute": "to"},
			},
			Tsx: true,
		},
		// Settings as single string (not array)
		{
			Code:    `<Link other="javascript:"></Link>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": "Link",
			},
			Tsx: true,
		},
		// Malformed legacy option: missing props
		{
			Code:    `<Foo href="javascript:"></Foo>`,
			Options: []interface{}{map[string]interface{}{"name": "Foo"}},
			Tsx:     true,
		},
		// Malformed legacy option: missing name
		{
			Code:    `<Foo href="javascript:"></Foo>`,
			Options: []interface{}{map[string]interface{}{"props": []interface{}{"href"}}},
			Tsx:     true,
		},
		// Malformed legacy option: props is string instead of array
		{
			Code:    `<Foo href="javascript:"></Foo>`,
			Options: []interface{}{map[string]interface{}{"name": "Foo", "props": "href"}},
			Tsx:     true,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		// Defaults — with full position assertion (EndLine/EndColumn)
		{
			Code:   `<a href="javascript:"></a>`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4, EndLine: 1, EndColumn: 22}},
			Tsx:    true,
		},
		{
			Code:   `<a href="javascript:void(0)"></a>`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4}},
			Tsx:    true,
		},
		{
			Code:   "<a href=\"j\n\n\na\rv\tascript:\"></a>",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4}},
			Tsx:    true,
		},
		// With component passed by options
		{
			Code:    `<Foo to="javascript:"></Foo>`,
			Options: []interface{}{map[string]interface{}{"name": "Foo", "props": []interface{}{"to", "href"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 6}},
			Tsx:     true,
		},
		{
			Code:    `<Foo href="javascript:"></Foo>`,
			Options: []interface{}{map[string]interface{}{"name": "Foo", "props": []interface{}{"to", "href"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 6}},
			Tsx:     true,
		},
		// Default "a" is still checked even when custom components are specified
		{
			Code:    `<a href="javascript:void(0)"></a>`,
			Options: []interface{}{map[string]interface{}{"name": "Foo", "props": []interface{}{"to", "href"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4}},
			Tsx:     true,
		},
		// With components passed by settings (includeFromSettings: true)
		{
			Code:    `<Foo to="javascript:"></Foo>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "Foo", "linkAttribute": "to"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 6}},
			Tsx:    true,
		},
		{
			Code:    `<Foo href="javascript:"></Foo>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "Foo", "linkAttribute": []interface{}{"to", "href"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 6}},
			Tsx:    true,
		},
		// Settings + legacy options combined
		{
			Code: `
			<div>
				<Foo href="javascript:"></Foo>
				<Bar link="javascript:"></Bar>
			</div>
			`,
			Options: []interface{}{
				[]interface{}{map[string]interface{}{"name": "Bar", "props": []interface{}{"link"}}},
				map[string]interface{}{"includeFromSettings": true},
			},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "Foo", "linkAttribute": []interface{}{"to", "href"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noScriptURL"},
				{MessageId: "noScriptURL"},
			},
			Tsx: true,
		},
		// Settings without includeFromSettings — only legacy options apply
		{
			Code: `
			<div>
				<Foo href="javascript:"></Foo>
				<Bar link="javascript:"></Bar>
			</div>
			`,
			Options: []interface{}{
				[]interface{}{map[string]interface{}{"name": "Bar", "props": []interface{}{"link"}}},
			},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "Foo", "linkAttribute": []interface{}{"to", "href"}},
				},
			},
			// Only Bar fires — Foo is in settings but includeFromSettings is false
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noScriptURL"},
			},
			Tsx: true,
		},

		// ---- Additional edge cases ----
		// Case insensitive
		{
			Code:   `<a href="JAVASCRIPT:void(0)"></a>`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4}},
			Tsx:    true,
		},
		{
			Code:   `<a href="JavaScript:void(0)"></a>`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4}},
			Tsx:    true,
		},
		// Control chars before javascript:
		{
			Code:   "<a href=\"\x01javascript:\"></a>",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4}},
			Tsx:    true,
		},
		// Spaces before javascript:
		{
			Code:   `<a href="  javascript:"></a>`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4}},
			Tsx:    true,
		},
		// Tab and newline between javascript: letters
		{
			Code:   "<a href=\"j\ta\nv\ra\tscript:\"></a>",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4}},
			Tsx:    true,
		},
		// Message text assertion
		{
			Code: `<a href="javascript:"></a>`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noScriptURL",
				Message:   noScriptURLMessage,
				Line:      1,
				Column:    4,
			}},
			Tsx: true,
		},
		// Namespaced tag — upstream resolves local part "a", matches default config
		{
			Code:   `<ns:a href="javascript:"></ns:a>`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 7}},
			Tsx:    true,
		},
		// Multiple custom components
		{
			Code: `<Link to="javascript:"></Link>`,
			Options: []interface{}{
				map[string]interface{}{"name": "Link", "props": []interface{}{"to"}},
				map[string]interface{}{"name": "Button", "props": []interface{}{"href"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 7}},
			Tsx:    true,
		},
		// Options via JSON path (map format — matches config.go single-option unwrap)
		{
			Code:    `<Foo to="javascript:"></Foo>`,
			Options: []interface{}{map[string]interface{}{"name": "Foo", "props": []interface{}{"to"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 6}},
			Tsx:     true,
		},
		// Multiline — full position assertion including EndLine/EndColumn
		{
			Code:   "<a\n\thref=\"javascript:void(0)\">\n</a>",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 2, Column: 2, EndLine: 2, EndColumn: 27}},
			Tsx:    true,
		},
		// Nested JSX — both fire independently
		{
			Code: `<div><a href="javascript:"><a href="javascript:void(0)"></a></a></div>`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noScriptURL"},
				{MessageId: "noScriptURL"},
			},
			Tsx: true,
		},
		// Settings with string-only component entry
		{
			Code:    `<Link href="javascript:"></Link>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{"Link"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 7}},
			Tsx:    true,
		},
		// Legacy option redefines "a" with ["href", "to"] — "to" now fires too
		{
			Code:    `<a to="javascript:"></a>`,
			Options: []interface{}{map[string]interface{}{"name": "a", "props": []interface{}{"href", "to"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4}},
			Tsx:     true,
		},
		// Settings with linkAttribute as string (not array)
		{
			Code:    `<Foo to="javascript:"></Foo>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "Foo", "linkAttribute": "to"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 6}},
			Tsx:    true,
		},
		// Settings as single object (not array) — triggers addOne directly
		{
			Code:    `<Foo to="javascript:"></Foo>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": map[string]interface{}{"name": "Foo", "linkAttribute": "to"},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 6}},
			Tsx:    true,
		},
		// Settings as single string — component gets default attr "href"
		{
			Code:    `<Link href="javascript:"></Link>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": "Link",
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 7}},
			Tsx:    true,
		},
		// Settings override "a" to only check "to" — "href" no longer flagged, "to" is
		{
			Code:    `<a to="javascript:"></a>`,
			Options: map[string]interface{}{"includeFromSettings": true},
			Settings: map[string]interface{}{
				"linkComponents": []interface{}{
					map[string]interface{}{"name": "a", "linkAttribute": "to"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 4}},
			Tsx:    true,
		},
		// Multiple legacy options for same name — last wins (set semantics)
		{
			Code: `<Foo to="javascript:"></Foo>`,
			Options: []interface{}{
				map[string]interface{}{"name": "Foo", "props": []interface{}{"href"}},
				map[string]interface{}{"name": "Foo", "props": []interface{}{"to"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noScriptURL", Line: 1, Column: 6}},
			Tsx:    true,
		},
	})
}
