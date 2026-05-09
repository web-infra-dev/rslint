package img_redundant_alt

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// arrayOption mirrors upstream's `array` const — the combined components +
// custom-words option used across multiple invalid cases.
var arrayOption = map[string]interface{}{
	"components": []interface{}{"Image"},
	"words":      []interface{}{"Word1", "Word2"},
}

// componentsSettings mirrors upstream's `componentsSettings` — the
// jsx-a11y components map that aliases `Image` → `img`.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Image": "img",
		},
	},
}

// expectedError mirrors upstream's expected error shape. Every invalid case
// emits the same message; we centralize it here so future text tweaks live
// in one place. Shared with img_redundant_alt_extras_test.go.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "redundantAlt",
	Message:   errorMessage,
}

// TestImgRedundantAltUpstream covers the full valid/invalid suite migrated
// 1:1 from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/img-redundant-alt-test.js`. Order and grouping mirror
// the upstream file so a future audit can grep across both side-by-side.
//
// rslint-specific lock-ins (TS wrappers, Unicode whitespace, real-world
// React patterns, position assertions, listener boundaries, etc.) live in
// img_redundant_alt_extras_test.go.
func TestImgRedundantAltUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ImgRedundantAltRule, []rule_tester.ValidTestCase{
		{Code: `<img alt="foo" />;`, Tsx: true},
		{Code: `<img alt="picture of me taking a photo of an image" aria-hidden />`, Tsx: true},
		{Code: `<img aria-hidden alt="photo of image" />`, Tsx: true},
		{Code: `<img ALt="foo" />;`, Tsx: true},
		{Code: `<img {...this.props} alt="foo" />`, Tsx: true},
		{Code: `<img {...this.props} alt={"foo"} />`, Tsx: true},
		{Code: `<img {...this.props} alt={alt} />`, Tsx: true},
		{Code: `<a />`, Tsx: true},
		{Code: `<img />`, Tsx: true},
		{Code: `<IMG />`, Tsx: true},
		{Code: `<img alt={undefined} />`, Tsx: true},
		{Code: "<img alt={`this should pass for ${now}`} />", Tsx: true},
		{Code: "<img alt={`this should pass for ${photo}`} />", Tsx: true},
		{Code: "<img alt={`this should pass for ${image}`} />", Tsx: true},
		{Code: "<img alt={`this should pass for ${picture}`} />", Tsx: true},
		{Code: "<img alt={`${photo}`} />", Tsx: true},
		{Code: "<img alt={`${image}`} />", Tsx: true},
		{Code: "<img alt={`${picture}`} />", Tsx: true},
		{Code: `<img alt={"undefined"} />`, Tsx: true},
		{Code: `<img alt={() => {}} />`, Tsx: true},
		{Code: `<img alt={function(e){}} />`, Tsx: true},
		{Code: `<img aria-hidden={false} alt="Doing cool things." />`, Tsx: true},
		{Code: `<UX.Layout>test</UX.Layout>`, Tsx: true},
		{Code: `<img alt />`, Tsx: true},
		{Code: `<img alt={imageAlt} />`, Tsx: true},
		{Code: `<img alt={imageAlt.name} />`, Tsx: true},
		// Upstream gates the next two behind `semver.satisfies(eslintVersion,
		// '>= 6')`. tsgo always supports optional chaining, so we drop the
		// version check.
		{Code: `<img alt={imageAlt?.name} />`, Tsx: true},
		{Code: `<img alt="Doing cool things" aria-hidden={foo?.bar}/>`, Tsx: true},
		{Code: `<img alt="Photography" />;`, Tsx: true},
		{Code: `<img alt="ImageMagick" />;`, Tsx: true},
		{Code: `<Image alt="Photo of a friend" />`, Tsx: true},
		{Code: `<Image alt="Foo" />`, Tsx: true, Settings: componentsSettings},
		{
			Code:    `<img alt="画像" />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"words": []interface{}{"イメージ"}}},
		},
	}, []rule_tester.InvalidTestCase{
		{Code: `<img alt="Photo of friend." />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="Picture of friend." />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="Image of friend." />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="PhOtO of friend." />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt={"photo"} />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="piCTUre of friend." />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="imAGE of friend." />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="photo of cool person" aria-hidden={false} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="picture of cool person" aria-hidden={false} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="image of cool person" aria-hidden={false} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="photo" {...this.props} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="image" {...this.props} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="picture" {...this.props} />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		// TemplateExpression with substitution — extractValueFromTemplateLiteral
		// produces "picture doing {things}", which contains the "picture" token.
		{Code: "<img alt={`picture doing ${things}`} {...this.props} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<img alt={`photo doing ${things}`} {...this.props} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<img alt={`image doing ${things}`} {...this.props} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<img alt={`picture doing ${picture}`} {...this.props} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<img alt={`photo doing ${photo}`} {...this.props} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: "<img alt={`image doing ${image}`} {...this.props} />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Image alt="Photo of a friend" />`, Tsx: true, Settings: componentsSettings, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		// ---- TESTS FOR ARRAY OPTION TESTS ----
		{Code: `<img alt="Word1" />;`, Tsx: true, Options: []interface{}{arrayOption}, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<img alt="Word2" />;`, Tsx: true, Options: []interface{}{arrayOption}, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Image alt="Word1" />;`, Tsx: true, Options: []interface{}{arrayOption}, Errors: []rule_tester.InvalidTestCaseError{expectedError}},
		{Code: `<Image alt="Word2" />;`, Tsx: true, Options: []interface{}{arrayOption}, Errors: []rule_tester.InvalidTestCaseError{expectedError}},

		{
			Code:    `<img alt="イメージ" />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"words": []interface{}{"イメージ"}}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:    `<img alt="イメージです" />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"words": []interface{}{"イメージ"}}},
			Errors:  []rule_tester.InvalidTestCaseError{expectedError},
		},
	})
}
