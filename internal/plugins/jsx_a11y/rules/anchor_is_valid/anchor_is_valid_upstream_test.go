package anchor_is_valid

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsOption mirrors upstream's `[{ components: ['Anchor', 'Link'] }]`.
var componentsOption = []interface{}{
	map[string]interface{}{
		"components": []interface{}{"Anchor", "Link"},
	},
}

// specialLinkOption mirrors upstream's `[{ specialLink: ['hrefLeft', 'hrefRight'] }]`.
var specialLinkOption = []interface{}{
	map[string]interface{}{
		"specialLink": []interface{}{"hrefLeft", "hrefRight"},
	},
}

var noHrefAspectOption = []interface{}{
	map[string]interface{}{
		"aspects": []interface{}{"noHref"},
	},
}

var invalidHrefAspectOption = []interface{}{
	map[string]interface{}{
		"aspects": []interface{}{"invalidHref"},
	},
}

var preferButtonAspectOption = []interface{}{
	map[string]interface{}{
		"aspects": []interface{}{"preferButton"},
	},
}

var noHrefInvalidHrefAspectOption = []interface{}{
	map[string]interface{}{
		"aspects": []interface{}{"noHref", "invalidHref"},
	},
}

var noHrefPreferButtonAspectOption = []interface{}{
	map[string]interface{}{
		"aspects": []interface{}{"noHref", "preferButton"},
	},
}

var preferButtonInvalidHrefAspectOption = []interface{}{
	map[string]interface{}{
		"aspects": []interface{}{"preferButton", "invalidHref"},
	},
}

var componentsAndSpecialLinkOption = []interface{}{
	map[string]interface{}{
		"components":  []interface{}{"Anchor"},
		"specialLink": []interface{}{"hrefLeft"},
	},
}

var componentsAndSpecialLinkAndInvalidHrefAspectOption = []interface{}{
	map[string]interface{}{
		"components":  []interface{}{"Anchor"},
		"specialLink": []interface{}{"hrefLeft"},
		"aspects":     []interface{}{"invalidHref"},
	},
}

var componentsAndSpecialLinkAndNoHrefAspectOption = []interface{}{
	map[string]interface{}{
		"components":  []interface{}{"Anchor"},
		"specialLink": []interface{}{"hrefLeft"},
		"aspects":     []interface{}{"noHref"},
	},
}

// componentsSettings mirrors upstream's
// `{ 'jsx-a11y': { components: { Anchor: 'a', Link: 'a' } } }`.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Anchor": "a",
			"Link":   "a",
		},
	},
}

// TestAnchorIsValidUpstream covers the full valid/invalid suite migrated
// 1:1 from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/anchor-is-valid-test.js`. rslint-specific lock-ins
// (semantic-walk branches, Dimension 4 universal edge shapes, tsgo AST
// quirks) live in anchor_is_valid_extras_test.go.
func TestAnchorIsValidUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AnchorIsValidRule, []rule_tester.ValidTestCase{
		// ---- DEFAULT ELEMENT 'a' TESTS ----
		{Code: `<Anchor />`, Tsx: true},
		{Code: `<a {...props} />`, Tsx: true},
		{Code: `<a href="foo" />`, Tsx: true},
		{Code: `<a href={foo} />`, Tsx: true},
		{Code: `<a href="/foo" />`, Tsx: true},
		{Code: `<a href="https://foo.bar.com" />`, Tsx: true},
		{Code: `<div href="foo" />`, Tsx: true},
		{Code: `<a href="javascript" />`, Tsx: true},
		{Code: `<a href="javascriptFoo" />`, Tsx: true},
		{Code: "<a href={`#foo`}/>", Tsx: true},
		{Code: `<a href={"foo"}/>`, Tsx: true},
		{Code: `<a href={"javascript"}/>`, Tsx: true},
		{Code: "<a href={`#javascript`}/>", Tsx: true},
		{Code: `<a href="#foo" />`, Tsx: true},
		{Code: `<a href="#javascript" />`, Tsx: true},
		{Code: `<a href="#javascriptFoo" />`, Tsx: true},
		{Code: `<UX.Layout>test</UX.Layout>`, Tsx: true},
		{Code: `<a href={this} />`, Tsx: true},

		// ---- CUSTOM ELEMENT TEST FOR ARRAY OPTION ----
		{Code: `<Anchor {...props} />`, Tsx: true, Options: componentsOption},
		{Code: `<Anchor href="foo" />`, Tsx: true, Options: componentsOption},
		{Code: `<Anchor href={foo} />`, Tsx: true, Options: componentsOption},
		{Code: `<Anchor href="/foo" />`, Tsx: true, Options: componentsOption},
		{Code: `<Anchor href="https://foo.bar.com" />`, Tsx: true, Options: componentsOption},
		{Code: `<div href="foo" />`, Tsx: true, Options: componentsOption},
		{Code: "<Anchor href={`#foo`}/>", Tsx: true, Options: componentsOption},
		{Code: `<Anchor href={"foo"}/>`, Tsx: true, Options: componentsOption},
		{Code: `<Anchor href="#foo" />`, Tsx: true, Options: componentsOption},
		{Code: `<Link {...props} />`, Tsx: true, Options: componentsOption},
		{Code: `<Link href="foo" />`, Tsx: true, Options: componentsOption},
		{Code: `<Link href={foo} />`, Tsx: true, Options: componentsOption},
		{Code: `<Link href="/foo" />`, Tsx: true, Options: componentsOption},
		{Code: `<Link href="https://foo.bar.com" />`, Tsx: true, Options: componentsOption},
		{Code: `<div href="foo" />`, Tsx: true, Options: componentsOption},
		{Code: "<Link href={`#foo`}/>", Tsx: true, Options: componentsOption},
		{Code: `<Link href={"foo"}/>`, Tsx: true, Options: componentsOption},
		{Code: `<Link href="#foo" />`, Tsx: true, Options: componentsOption},
		{Code: `<Link href="#foo" />`, Tsx: true, Settings: componentsSettings},

		// ---- CUSTOM PROP TESTS ----
		{Code: `<a {...props} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft="foo" />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft={foo} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft="/foo" />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft="https://foo.bar.com" />`, Tsx: true, Options: specialLinkOption},
		{Code: `<div hrefLeft="foo" />`, Tsx: true, Options: specialLinkOption},
		{Code: "<a hrefLeft={`#foo`}/>", Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft={"foo"}/>`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft="#foo" />`, Tsx: true, Options: specialLinkOption},
		{Code: `<UX.Layout>test</UX.Layout>`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight={this} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a {...props} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight="foo" />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight={foo} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight="/foo" />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight="https://foo.bar.com" />`, Tsx: true, Options: specialLinkOption},
		{Code: `<div hrefRight="foo" />`, Tsx: true, Options: specialLinkOption},
		{Code: "<a hrefRight={`#foo`}/>", Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight={"foo"}/>`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight="#foo" />`, Tsx: true, Options: specialLinkOption},
		{Code: `<UX.Layout>test</UX.Layout>`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight={this} />`, Tsx: true, Options: specialLinkOption},

		// ---- CUSTOM BOTH COMPONENTS AND SPECIAL LINK TESTS ----
		{Code: `<Anchor {...props} />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft="foo" />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft={foo} />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft="/foo" />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft="https://foo.bar.com" />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<div hrefLeft="foo" />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: "<Anchor hrefLeft={`#foo`}/>", Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft={"foo"}/>`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft="#foo" />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<UX.Layout>test</UX.Layout>`, Tsx: true, Options: componentsAndSpecialLinkOption},

		// ---- WITH ON CLICK — DEFAULT ELEMENT 'a' TESTS ----
		{Code: `<a {...props} onClick={() => void 0} />`, Tsx: true},
		{Code: `<a href="foo" onClick={() => void 0} />`, Tsx: true},
		{Code: `<a href={foo} onClick={() => void 0} />`, Tsx: true},
		{Code: `<a href="/foo" onClick={() => void 0} />`, Tsx: true},
		{Code: `<a href="https://foo.bar.com" onClick={() => void 0} />`, Tsx: true},
		{Code: `<div href="foo" onClick={() => void 0} />`, Tsx: true},
		{Code: "<a href={`#foo`} onClick={() => void 0} />", Tsx: true},
		{Code: `<a href={"foo"} onClick={() => void 0} />`, Tsx: true},
		{Code: `<a href="#foo" onClick={() => void 0} />`, Tsx: true},
		{Code: `<a href={this} onClick={() => void 0} />`, Tsx: true},

		// ---- WITH ON CLICK — CUSTOM ELEMENT TEST FOR ARRAY OPTION ----
		{Code: `<Anchor {...props} onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Anchor href="foo" onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Anchor href={foo} onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Anchor href="/foo" onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Anchor href="https://foo.bar.com" onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: "<Anchor href={`#foo`} onClick={() => void 0} />", Tsx: true, Options: componentsOption},
		{Code: `<Anchor href={"foo"} onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Anchor href="#foo" onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Link {...props} onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Link href="foo" onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Link href={foo} onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Link href="/foo" onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Link href="https://foo.bar.com" onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<div href="foo" onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: "<Link href={`#foo`} onClick={() => void 0} />", Tsx: true, Options: componentsOption},
		{Code: `<Link href={"foo"} onClick={() => void 0} />`, Tsx: true, Options: componentsOption},
		{Code: `<Link href="#foo" onClick={() => void 0} />`, Tsx: true, Options: componentsOption},

		// ---- WITH ON CLICK — CUSTOM PROP TESTS ----
		{Code: `<a {...props} onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft="foo" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft={foo} onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft="/foo" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft href="https://foo.bar.com" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<div hrefLeft="foo" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: "<a hrefLeft={`#foo`} onClick={() => void 0} />", Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft={"foo"} onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefLeft="#foo" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight={this} onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a {...props} onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight="foo" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight={foo} onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight="/foo" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight href="https://foo.bar.com" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<div hrefRight="foo" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: "<a hrefRight={`#foo`} onClick={() => void 0} />", Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight={"foo"} onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight="#foo" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},
		{Code: `<a hrefRight={this} onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption},

		// ---- WITH ON CLICK — CUSTOM BOTH COMPONENTS AND SPECIAL LINK TESTS ----
		{Code: `<Anchor {...props} onClick={() => void 0} />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft="foo" onClick={() => void 0} />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft={foo} onClick={() => void 0} />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft="/foo" onClick={() => void 0} />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft href="https://foo.bar.com" onClick={() => void 0} />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: "<Anchor hrefLeft={`#foo`} onClick={() => void 0} />", Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft={"foo"} onClick={() => void 0} />`, Tsx: true, Options: componentsAndSpecialLinkOption},
		{Code: `<Anchor hrefLeft="#foo" onClick={() => void 0} />`, Tsx: true, Options: componentsAndSpecialLinkOption},

		// ---- WITH ASPECTS TESTS — NO HREF ----
		{Code: `<a />`, Tsx: true, Options: invalidHrefAspectOption},
		{Code: `<a href={undefined} />`, Tsx: true, Options: invalidHrefAspectOption},
		{Code: `<a href={null} />`, Tsx: true, Options: invalidHrefAspectOption},
		{Code: `<a />`, Tsx: true, Options: preferButtonAspectOption},
		{Code: `<a href={undefined} />`, Tsx: true, Options: preferButtonAspectOption},
		{Code: `<a href={null} />`, Tsx: true, Options: preferButtonAspectOption},
		{Code: `<a />`, Tsx: true, Options: preferButtonInvalidHrefAspectOption},
		{Code: `<a href={undefined} />`, Tsx: true, Options: preferButtonInvalidHrefAspectOption},
		{Code: `<a href={null} />`, Tsx: true, Options: preferButtonInvalidHrefAspectOption},

		// ---- WITH ASPECTS TESTS — INVALID HREF ----
		{Code: `<a href="" />;`, Tsx: true, Options: preferButtonAspectOption},
		{Code: `<a href="#" />`, Tsx: true, Options: preferButtonAspectOption},
		{Code: `<a href={"#"} />`, Tsx: true, Options: preferButtonAspectOption},
		{Code: `<a href="javascript:void(0)" />`, Tsx: true, Options: preferButtonAspectOption},
		{Code: `<a href={"javascript:void(0)"} />`, Tsx: true, Options: preferButtonAspectOption},
		{Code: `<a href="" />;`, Tsx: true, Options: noHrefAspectOption},
		{Code: `<a href="#" />`, Tsx: true, Options: noHrefAspectOption},
		{Code: `<a href={"#"} />`, Tsx: true, Options: noHrefAspectOption},
		{Code: `<a href="javascript:void(0)" />`, Tsx: true, Options: noHrefAspectOption},
		{Code: `<a href={"javascript:void(0)"} />`, Tsx: true, Options: noHrefAspectOption},
		{Code: `<a href="" />;`, Tsx: true, Options: noHrefPreferButtonAspectOption},
		{Code: `<a href="#" />`, Tsx: true, Options: noHrefPreferButtonAspectOption},
		{Code: `<a href={"#"} />`, Tsx: true, Options: noHrefPreferButtonAspectOption},
		{Code: `<a href="javascript:void(0)" />`, Tsx: true, Options: noHrefPreferButtonAspectOption},
		{Code: `<a href={"javascript:void(0)"} />`, Tsx: true, Options: noHrefPreferButtonAspectOption},

		// ---- WITH ASPECTS TESTS — SHOULD BE BUTTON ----
		{Code: `<a onClick={() => void 0} />`, Tsx: true, Options: invalidHrefAspectOption},
		{Code: `<a href="#" onClick={() => void 0} />`, Tsx: true, Options: noHrefAspectOption},
		{Code: `<a href="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Options: noHrefAspectOption},
		{Code: `<a href={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Options: noHrefAspectOption},

		// ---- CUSTOM COMPONENTS AND SPECIAL LINK AND ASPECT ----
		{Code: `<Anchor hrefLeft={undefined} />`, Tsx: true, Options: componentsAndSpecialLinkAndInvalidHrefAspectOption},
		{Code: `<Anchor hrefLeft={null} />`, Tsx: true, Options: componentsAndSpecialLinkAndInvalidHrefAspectOption},
	}, []rule_tester.InvalidTestCase{
		// ---- DEFAULT ELEMENT 'a' TESTS — NO HREF ----
		{Code: `<a />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={undefined} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={null} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		// ---- DEFAULT ELEMENT 'a' TESTS — INVALID HREF ----
		{Code: `<a href="" />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="#" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"#"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="javascript:void(0)" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)"} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// ---- DEFAULT ELEMENT 'a' TESTS — SHOULD BE BUTTON ----
		{Code: `<a onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="#" onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},

		// ---- CUSTOM ELEMENT TEST FOR ARRAY OPTION — NO HREF ----
		{Code: `<Link />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Link href={undefined} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Link href={null} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		// ---- CUSTOM ELEMENT TEST FOR ARRAY OPTION — INVALID HREF ----
		{Code: `<Link href="" />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Link href="#" />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Link href={"#"} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Link href="javascript:void(0)" />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Link href={"javascript:void(0)"} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor href="" />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor href="#" />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor href={"#"} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor href="javascript:void(0)" />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor href={"javascript:void(0)"} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// ---- CUSTOM ELEMENT TEST FOR ARRAY OPTION — SHOULD BE BUTTON ----
		{Code: `<Link onClick={() => void 0} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Link href="#" onClick={() => void 0} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Link href="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Link href={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor onClick={() => void 0} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor href="#" onClick={() => void 0} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor href="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor href={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Options: componentsOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Link href="#" onClick={() => void 0} />`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},

		// ---- CUSTOM PROP TESTS — NO HREF ----
		{Code: `<a hrefLeft={undefined} />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a hrefLeft={null} />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		// ---- CUSTOM PROP TESTS — INVALID HREF ----
		{Code: `<a hrefLeft="" />;`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a hrefLeft="#" />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a hrefLeft={"#"} />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a hrefLeft="javascript:void(0)" />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a hrefLeft={"javascript:void(0)"} />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// ---- CUSTOM PROP TESTS — SHOULD BE BUTTON ----
		{Code: `<a hrefLeft="#" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a hrefLeft="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a hrefLeft={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Options: specialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},

		// ---- CUSTOM BOTH COMPONENTS AND SPECIAL LINK TESTS — NO HREF ----
		// Upstream's curious `Anchor={undefined}` (note: not hrefLeft) — tests
		// that the rule treats unrelated props as no-href (Anchor doesn't
		// match propsToValidate, so values is [undefined, undefined] for
		// [href, hrefLeft] → noHref).
		{Code: `<Anchor Anchor={undefined} />`, Tsx: true, Options: componentsAndSpecialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor hrefLeft={null} />`, Tsx: true, Options: componentsAndSpecialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		// ---- CUSTOM BOTH COMPONENTS AND SPECIAL LINK TESTS — INVALID HREF ----
		{Code: `<Anchor hrefLeft="" />;`, Tsx: true, Options: componentsAndSpecialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor hrefLeft="#" />`, Tsx: true, Options: componentsAndSpecialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor hrefLeft={"#"} />`, Tsx: true, Options: componentsAndSpecialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor hrefLeft="javascript:void(0)" />`, Tsx: true, Options: componentsAndSpecialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor hrefLeft={"javascript:void(0)"} />`, Tsx: true, Options: componentsAndSpecialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		// ---- CUSTOM BOTH COMPONENTS AND SPECIAL LINK TESTS — SHOULD BE BUTTON ----
		{Code: `<Anchor hrefLeft="#" onClick={() => void 0} />`, Tsx: true, Options: componentsAndSpecialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor hrefLeft="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Options: componentsAndSpecialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor hrefLeft={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Options: componentsAndSpecialLinkOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},

		// ---- WITH ASPECTS TESTS — NO HREF ----
		{Code: `<a />`, Tsx: true, Options: noHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a />`, Tsx: true, Options: noHrefPreferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a />`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={undefined} />`, Tsx: true, Options: noHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={undefined} />`, Tsx: true, Options: noHrefPreferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={undefined} />`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={null} />`, Tsx: true, Options: noHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={null} />`, Tsx: true, Options: noHrefPreferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={null} />`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- WITH ASPECTS TESTS — INVALID HREF ----
		{Code: `<a href="" />;`, Tsx: true, Options: invalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="" />;`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="" />;`, Tsx: true, Options: preferButtonInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="#" />;`, Tsx: true, Options: invalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="#" />;`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="#" />;`, Tsx: true, Options: preferButtonInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"#"} />;`, Tsx: true, Options: invalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"#"} />;`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"#"} />;`, Tsx: true, Options: preferButtonInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="javascript:void(0)" />;`, Tsx: true, Options: invalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="javascript:void(0)" />;`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="javascript:void(0)" />;`, Tsx: true, Options: preferButtonInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)"} />;`, Tsx: true, Options: invalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)"} />;`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)"} />;`, Tsx: true, Options: preferButtonInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- WITH ASPECTS TESTS — SHOULD BE BUTTON ----
		{Code: `<a onClick={() => void 0} />`, Tsx: true, Options: preferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a onClick={() => void 0} />`, Tsx: true, Options: preferButtonInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a onClick={() => void 0} />`, Tsx: true, Options: noHrefPreferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		// onClick + only-noHref: the `(onClick && !preferButton)` arm of the
		// `||` gates the noHref report; preferButton off → noHref reported.
		{Code: `<a onClick={() => void 0} />`, Tsx: true, Options: noHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a onClick={() => void 0} />`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="#" onClick={() => void 0} />`, Tsx: true, Options: preferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="#" onClick={() => void 0} />`, Tsx: true, Options: noHrefPreferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="#" onClick={() => void 0} />`, Tsx: true, Options: preferButtonInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="#" onClick={() => void 0} />`, Tsx: true, Options: invalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="#" onClick={() => void 0} />`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Options: preferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Options: noHrefPreferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Options: preferButtonInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Options: invalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href="javascript:void(0)" onClick={() => void 0} />`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Options: preferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Options: noHrefPreferButtonAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Options: preferButtonInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferButton", Message: preferButtonErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Options: invalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<a href={"javascript:void(0)"} onClick={() => void 0} />`, Tsx: true, Options: noHrefInvalidHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidHref", Message: invalidHrefErrorMessage, Line: 1, Column: 1}}},

		// ---- CUSTOM COMPONENTS AND SPECIAL LINK AND ASPECT ----
		{Code: `<Anchor hrefLeft={undefined} />`, Tsx: true, Options: componentsAndSpecialLinkAndNoHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
		{Code: `<Anchor hrefLeft={null} />`, Tsx: true, Options: componentsAndSpecialLinkAndNoHrefAspectOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noHref", Message: noHrefErrorMessage, Line: 1, Column: 1}}},
	})
}
