package alt_text

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// componentsSettings mirrors the upstream `componentsSettings` constant — the
// `as` polymorphic prop and an `Input` → `input` component map. Used for the
// settings-based valid/invalid cases from upstream's test file.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"polymorphicPropName": "as",
		"components": map[string]interface{}{
			"Input": "input",
		},
	},
}

// arrayOpts mirrors the upstream `array` constant — custom-component options
// mapping each DOM element to a list of accepted component names.
var arrayOpts = map[string]interface{}{
	"img":                 []interface{}{"Thumbnail", "Image"},
	"object":              []interface{}{"Object"},
	"area":                []interface{}{"Area"},
	`input[type="image"]`: []interface{}{"InputImage"},
}

// TestAltTextUpstream covers the full valid/invalid suite migrated 1:1
// from upstream eslint-plugin-jsx-a11y's `__tests__/src/rules/alt-text-test.js`.
// rslint-specific lock-ins (staticEval coverage, TS wrappers, real-world
// patterns, polymorphic edge cases, etc.) live in alt_text_extras_test.go.
func TestAltTextUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AltTextRule, []rule_tester.ValidTestCase{
		// ---- DEFAULT ELEMENT 'img' TESTS ----
		{Code: `<img alt="foo" />;`, Tsx: true},
		{Code: `<img alt={"foo"} />;`, Tsx: true},
		{Code: `<img alt={alt} />;`, Tsx: true},
		{Code: `<img ALT="foo" />;`, Tsx: true},
		{Code: "<img ALT={`This is the ${alt} text`} />;", Tsx: true},
		{Code: `<img ALt="foo" />;`, Tsx: true},
		{Code: `<img alt="foo" salt={undefined} />;`, Tsx: true},
		{Code: `<img {...this.props} alt="foo" />`, Tsx: true},
		{Code: `<a />`, Tsx: true},
		{Code: `<div />`, Tsx: true},
		{Code: `<img alt={function(e) {} } />`, Tsx: true},
		{Code: `<div alt={function(e) {} } />`, Tsx: true},
		{Code: `<img alt={() => void 0} />`, Tsx: true},
		{Code: `<IMG />`, Tsx: true},
		{Code: `<UX.Layout>test</UX.Layout>`, Tsx: true},
		{Code: `<img alt={alt || "Alt text" } />`, Tsx: true},
		{Code: `<img alt={photo.caption} />;`, Tsx: true},
		{Code: `<img alt={bar()} />;`, Tsx: true},
		{Code: `<img alt={foo.bar || ""} />`, Tsx: true},
		{Code: `<img alt={bar() || ""} />`, Tsx: true},
		{Code: `<img alt={foo.bar() || ""} />`, Tsx: true},
		{Code: `<img alt="" />`, Tsx: true},
		{Code: "<img alt={`${undefined}`} />", Tsx: true},
		{Code: `<img alt=" " />`, Tsx: true},
		{Code: `<img alt="" role="presentation" />`, Tsx: true},
		{Code: `<img alt="" role="none" />`, Tsx: true},
		{Code: "<img alt=\"\" role={`presentation`} />", Tsx: true},
		{Code: `<img alt="" role={"presentation"} />`, Tsx: true},
		{Code: `<img alt="this is lit..." role="presentation" />`, Tsx: true},
		{Code: `<img alt={error ? "not working": "working"} />`, Tsx: true},
		{Code: `<img alt={undefined ? "working": "not working"} />`, Tsx: true},
		{Code: `<img alt={plugin.name + " Logo"} />`, Tsx: true},
		{Code: `<img aria-label="foo" />`, Tsx: true},
		{Code: `<img aria-labelledby="id1" />`, Tsx: true},

		// ---- DEFAULT <object> TESTS ----
		{Code: `<object aria-label="foo" />`, Tsx: true},
		{Code: `<object aria-labelledby="id1" />`, Tsx: true},
		{Code: `<object>Foo</object>`, Tsx: true},
		{Code: `<object><p>This is descriptive!</p></object>`, Tsx: true},
		{Code: `<Object />`, Tsx: true},
		{Code: `<object title="An object" />`, Tsx: true},

		// ---- DEFAULT <area> TESTS ----
		{Code: `<area aria-label="foo" />`, Tsx: true},
		{Code: `<area aria-labelledby="id1" />`, Tsx: true},
		{Code: `<area alt="" />`, Tsx: true},
		{Code: `<area alt="This is descriptive!" />`, Tsx: true},
		{Code: `<area alt={altText} />`, Tsx: true},
		{Code: `<Area />`, Tsx: true},

		// ---- DEFAULT <input type="image"> TESTS ----
		{Code: `<input />`, Tsx: true},
		{Code: `<input type="foo" />`, Tsx: true},
		{Code: `<input type="image" aria-label="foo" />`, Tsx: true},
		{Code: `<input type="image" aria-labelledby="id1" />`, Tsx: true},
		{Code: `<input type="image" alt="" />`, Tsx: true},
		{Code: `<input type="image" alt="This is descriptive!" />`, Tsx: true},
		{Code: `<input type="image" alt={altText} />`, Tsx: true},
		{Code: `<InputImage />`, Tsx: true},
		{Code: `<Input type="image" alt="" />`, Tsx: true, Settings: componentsSettings},
		{Code: `<SomeComponent as="input" type="image" alt="" />`, Tsx: true, Settings: componentsSettings},

		// ---- CUSTOM ELEMENT TESTS FOR ARRAY OPTION TESTS ----
		{Code: `<Thumbnail alt="foo" />;`, Tsx: true, Options: arrayOpts},
		{Code: `<Thumbnail alt={"foo"} />;`, Tsx: true, Options: arrayOpts},
		{Code: `<Thumbnail alt={alt} />;`, Tsx: true, Options: arrayOpts},
		{Code: `<Thumbnail ALT="foo" />;`, Tsx: true, Options: arrayOpts},
		{Code: "<Thumbnail ALT={`This is the ${alt} text`} />;", Tsx: true, Options: arrayOpts},
		{Code: `<Thumbnail ALt="foo" />;`, Tsx: true, Options: arrayOpts},
		{Code: `<Thumbnail alt="foo" salt={undefined} />;`, Tsx: true, Options: arrayOpts},
		{Code: `<Thumbnail {...this.props} alt="foo" />`, Tsx: true, Options: arrayOpts},
		{Code: `<thumbnail />`, Tsx: true, Options: arrayOpts},
		{Code: `<Thumbnail alt={function(e) {} } />`, Tsx: true, Options: arrayOpts},
		{Code: `<div alt={function(e) {} } />`, Tsx: true, Options: arrayOpts},
		{Code: `<Thumbnail alt={() => void 0} />`, Tsx: true, Options: arrayOpts},
		{Code: `<THUMBNAIL />`, Tsx: true, Options: arrayOpts},
		{Code: `<Thumbnail alt={alt || "foo" } />`, Tsx: true, Options: arrayOpts},
		{Code: `<Image alt="foo" />;`, Tsx: true, Options: arrayOpts},
		{Code: `<Image alt={"foo"} />;`, Tsx: true, Options: arrayOpts},
		{Code: `<Image alt={alt} />;`, Tsx: true, Options: arrayOpts},
		{Code: `<Image ALT="foo" />;`, Tsx: true, Options: arrayOpts},
		{Code: "<Image ALT={`This is the ${alt} text`} />;", Tsx: true, Options: arrayOpts},
		{Code: `<Image ALt="foo" />;`, Tsx: true, Options: arrayOpts},
		{Code: `<Image alt="foo" salt={undefined} />;`, Tsx: true, Options: arrayOpts},
		{Code: `<Image {...this.props} alt="foo" />`, Tsx: true, Options: arrayOpts},
		{Code: `<image />`, Tsx: true, Options: arrayOpts},
		{Code: `<Image alt={function(e) {} } />`, Tsx: true, Options: arrayOpts},
		{Code: `<div alt={function(e) {} } />`, Tsx: true, Options: arrayOpts},
		{Code: `<Image alt={() => void 0} />`, Tsx: true, Options: arrayOpts},
		{Code: `<IMAGE />`, Tsx: true, Options: arrayOpts},
		{Code: `<Image alt={alt || "foo" } />`, Tsx: true, Options: arrayOpts},
		{Code: `<Object aria-label="foo" />`, Tsx: true, Options: arrayOpts},
		{Code: `<Object aria-labelledby="id1" />`, Tsx: true, Options: arrayOpts},
		{Code: `<Object>Foo</Object>`, Tsx: true, Options: arrayOpts},
		{Code: `<Object><p>This is descriptive!</p></Object>`, Tsx: true, Options: arrayOpts},
		{Code: `<Object title="An object" />`, Tsx: true, Options: arrayOpts},
		{Code: `<Area aria-label="foo" />`, Tsx: true, Options: arrayOpts},
		{Code: `<Area aria-labelledby="id1" />`, Tsx: true, Options: arrayOpts},
		{Code: `<Area alt="" />`, Tsx: true, Options: arrayOpts},
		{Code: `<Area alt="This is descriptive!" />`, Tsx: true, Options: arrayOpts},
		{Code: `<Area alt={altText} />`, Tsx: true, Options: arrayOpts},
		{Code: `<InputImage aria-label="foo" />`, Tsx: true, Options: arrayOpts},
		{Code: `<InputImage aria-labelledby="id1" />`, Tsx: true, Options: arrayOpts},
		{Code: `<InputImage alt="" />`, Tsx: true, Options: arrayOpts},
		{Code: `<InputImage alt="This is descriptive!" />`, Tsx: true, Options: arrayOpts},
		{Code: `<InputImage alt={altText} />`, Tsx: true, Options: arrayOpts},
	}, []rule_tester.InvalidTestCase{
		// ---- DEFAULT ELEMENT 'img' TESTS ----
		{
			Code: `<img />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt={undefined} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img src="xyz" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img role />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img {...this.props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt={false || false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt={undefined} role="presentation" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img alt role="presentation" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("img"), Line: 1, Column: 1},
			},
		},
		{
			Code: `<img role="presentation" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferAlt", Message: msgPreferAlt, Line: 1, Column: 1},
			},
		},
		{
			Code: `<img role="none" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferAlt", Message: msgPreferAlt, Line: 1, Column: 1},
			},
		},
		{
			Code: `<img aria-label={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ariaLabelValue", Message: msgAriaLabelEmpty, Line: 1, Column: 1},
			},
		},
		{
			Code: `<img aria-labelledby={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ariaLabelledByValue", Message: msgAriaLabelledByEmpty, Line: 1, Column: 1},
			},
		},
		{
			Code: `<img aria-label="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ariaLabelValue", Message: msgAriaLabelEmpty, Line: 1, Column: 1},
			},
		},
		{
			Code: `<img aria-labelledby="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ariaLabelledByValue", Message: msgAriaLabelledByEmpty, Line: 1, Column: 1},
			},
		},
		{
			Code:     `<SomeComponent as="img" aria-label="" />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ariaLabelValue", Message: msgAriaLabelEmpty, Line: 1, Column: 1},
			},
		},

		// ---- DEFAULT ELEMENT 'object' TESTS ----
		{
			Code: `<object />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code: `<object><div aria-hidden /></object>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code: `<object title={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code: `<object aria-label="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code: `<object aria-labelledby="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code: `<object aria-label={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code: `<object aria-labelledby={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},

		// ---- DEFAULT ELEMENT 'area' TESTS ----
		{
			Code: `<area />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code: `<area alt />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code: `<area alt={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code: `<area src="xyz" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code: `<area {...this.props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code: `<area aria-label="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code: `<area aria-label={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code: `<area aria-labelledby="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code: `<area aria-labelledby={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},

		// ---- DEFAULT ELEMENT 'input type="image"' TESTS ----
		{
			Code: `<input type="image" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="image" alt />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="image" alt={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="image">Foo</input>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="image" {...this.props} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="image" aria-label="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="image" aria-label={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="image" aria-labelledby="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="image" aria-labelledby={undefined} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},

		// ---- CUSTOM ELEMENT TESTS FOR ARRAY OPTION TESTS ----
		{
			Code:    `<Thumbnail />;`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("Thumbnail"), Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Thumbnail alt />;`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("Thumbnail"), Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Thumbnail alt={undefined} />;`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("Thumbnail"), Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Thumbnail src="xyz" />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("Thumbnail"), Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Thumbnail {...this.props} />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("Thumbnail"), Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Image />;`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("Image"), Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Image alt />;`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("Image"), Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Image alt={undefined} />;`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "altValue", Message: altValueMessage("Image"), Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Image src="xyz" />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("Image"), Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Image {...this.props} />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProp", Message: missingPropMessage("Image"), Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Object />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Object><div aria-hidden /></Object>`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Object title={undefined} />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "object", Message: msgObject, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Area />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Area alt />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Area alt={undefined} />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Area src="xyz" />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Area {...this.props} />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "area", Message: msgArea, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<InputImage />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<InputImage alt />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<InputImage alt={undefined} />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<InputImage>Foo</InputImage>`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<InputImage {...this.props} />`,
			Tsx:     true,
			Options: arrayOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
		{
			Code:     `<Input type="image" />`,
			Tsx:      true,
			Settings: componentsSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "inputImage", Message: msgInputImage, Line: 1, Column: 1},
			},
		},
	})
}
