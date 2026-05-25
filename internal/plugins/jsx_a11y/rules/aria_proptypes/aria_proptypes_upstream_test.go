// cspell:ignore abcaria descnding foobar removalss

package aria_proptypes

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// upstreamErrorMessage mirrors the test file's local `errorMessage(name)`
// helper byte-for-byte. It looks up the ARIA property's type / values in
// aria-query's map, then formats the same template the rule itself uses.
// Defined at package scope so aria_proptypes_extras_test.go can reuse.
//
// `${permittedValues}` in upstream's template literal hits
// `Array.prototype.toString`, which joins with a bare comma (no space).
// Booleans stringify as "true" / "false". Mirror byte-for-byte.
func upstreamErrorMessage(name string) string {
	def := jsxa11yutil.AriaPropertyDefinitions[strings.ToLower(name)]
	return errorMessage(name, def.Type, def.Values)
}

// TestValidityCheckDefault locks in upstream's tape unit test:
//
//	test('validityCheck', (t) => {
//	  t.equal(
//	    validityCheck(null, null),
//	    false,
//	    'is false for an unknown expected type',
//	  );
//	  t.end();
//	});
//
// Upstream runs this OUTSIDE the RuleTester suite as a direct assertion on
// the exported `validityCheck` function. We mirror by feeding the helper a
// NoLit value with a bogus type tag — both arms land on validityCheck's
// `default: return false`.
func TestValidityCheckDefault(t *testing.T) {
	if got := validityCheck(jsxa11yutil.AriaLiteralValue{Kind: jsxa11yutil.AriaLiteralNoLit}, "unknownType", nil); got != false {
		t.Fatalf("expected false for unknown type, got %v", got)
	}
	// Empty-string type — corresponds to upstream's literal `null` passed
	// as the second arg; same default-arm hit.
	if got := validityCheck(jsxa11yutil.AriaLiteralValue{Kind: jsxa11yutil.AriaLiteralNoLit}, "", nil); got != false {
		t.Fatalf("expected false for empty type, got %v", got)
	}
}

// TestAriaProptypesUpstream mirrors the full valid/invalid suite from
// upstream's `__tests__/src/rules/aria-proptypes-test.js`, 1:1 and in
// upstream order so a future audit can grep across both side-by-side.
//
// Anything NOT in upstream's test file — case-insensitive attribute names,
// parenthesized / TS-wrapped values, BigInt literals, aria-haspopup
// boolean tokens, aria-current cross-class values, aria-orientation's
// "undefined" string token, real-user patterns — lives in
// aria_proptypes_extras_test.go.
func TestAriaProptypesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AriaProptypesRule,
		[]rule_tester.ValidTestCase{
			// Non-aria-* and unrecognized aria-* attrs — gate-1 skip.
			{Code: `<div aria-foo="true" />`, Tsx: true},
			{Code: `<div abcaria-foo="true" />`, Tsx: true},

			// aria-hidden (boolean, allowUndefined). Boolean literals,
			// string-coerced "true"/"false", boolean attribute form,
			// PrefixUnary booleans, Identifier / Member (non-literal →
			// step-3 skip), null / undefined (step-1 skip), JSX
			// (non-literal → step-3 skip).
			{Code: `<div aria-hidden={true} />`, Tsx: true},
			{Code: `<div aria-hidden="true" />`, Tsx: true},
			{Code: `<div aria-hidden={"false"} />`, Tsx: true},
			{Code: `<div aria-hidden={!false} />`, Tsx: true},
			{Code: `<div aria-hidden />`, Tsx: true},
			{Code: `<div aria-hidden={false} />`, Tsx: true},
			{Code: `<div aria-hidden={!true} />`, Tsx: true},
			{Code: `<div aria-hidden={!"yes"} />`, Tsx: true},
			{Code: `<div aria-hidden={foo} />`, Tsx: true},
			{Code: `<div aria-hidden={foo.bar} />`, Tsx: true},
			{Code: `<div aria-hidden={null} />`, Tsx: true},
			{Code: `<div aria-hidden={undefined} />`, Tsx: true},
			{Code: `<div aria-hidden={<div />} />`, Tsx: true},

			// aria-label (string). String literals, NoSubstitutionTemplate,
			// Identifier / Member (skip), null / undefined (skip).
			{Code: `<div aria-label="Close" />`, Tsx: true},
			{Code: "<div aria-label={`Close`} />", Tsx: true},
			{Code: `<div aria-label={foo} />`, Tsx: true},
			{Code: `<div aria-label={foo.bar} />`, Tsx: true},
			{Code: `<div aria-label={null} />`, Tsx: true},
			{Code: `<div aria-label={undefined} />`, Tsx: true},

			// aria-invalid (token: ['grammar', false, 'spelling', true]) via
			// ConditionalExpression — getLiteralPropValue returns null for
			// Conditional → step-3 skip.
			{Code: `<input aria-invalid={error ? "true" : "false"} />`, Tsx: true},
			{Code: `<input aria-invalid={undefined ? "true" : "false"} />`, Tsx: true},

			// aria-checked (tristate). Boolean, "true"/"false" coerced,
			// PrefixUnary, "mixed" (string + template), null / undefined.
			{Code: `<div aria-checked={true} />`, Tsx: true},
			{Code: `<div aria-checked="true" />`, Tsx: true},
			{Code: `<div aria-checked={"false"} />`, Tsx: true},
			{Code: `<div aria-checked={!false} />`, Tsx: true},
			{Code: `<div aria-checked />`, Tsx: true},
			{Code: `<div aria-checked={false} />`, Tsx: true},
			{Code: `<div aria-checked={!true} />`, Tsx: true},
			{Code: `<div aria-checked={!"yes"} />`, Tsx: true},
			{Code: `<div aria-checked={foo} />`, Tsx: true},
			{Code: `<div aria-checked={foo.bar} />`, Tsx: true},
			{Code: `<div aria-checked="mixed" />`, Tsx: true},
			{Code: "<div aria-checked={`mixed`} />", Tsx: true},
			{Code: `<div aria-checked={null} />`, Tsx: true},
			{Code: `<div aria-checked={undefined} />`, Tsx: true},

			// aria-level (integer). Numeric / unary numeric / string
			// numerics / template numerics / Identifier (skip) / null /
			// undefined.
			{Code: `<div aria-level={123} />`, Tsx: true},
			{Code: `<div aria-level={-123} />`, Tsx: true},
			{Code: `<div aria-level={+123} />`, Tsx: true},
			{Code: `<div aria-level={~123} />`, Tsx: true},
			{Code: `<div aria-level={"123"} />`, Tsx: true},
			{Code: "<div aria-level={`123`} />", Tsx: true},
			{Code: `<div aria-level="123" />`, Tsx: true},
			{Code: `<div aria-level={foo} />`, Tsx: true},
			{Code: `<div aria-level={foo.bar} />`, Tsx: true},
			{Code: `<div aria-level={null} />`, Tsx: true},
			{Code: `<div aria-level={undefined} />`, Tsx: true},

			// aria-valuemax (number) — same shape as aria-level.
			{Code: `<div aria-valuemax={123} />`, Tsx: true},
			{Code: `<div aria-valuemax={-123} />`, Tsx: true},
			{Code: `<div aria-valuemax={+123} />`, Tsx: true},
			{Code: `<div aria-valuemax={~123} />`, Tsx: true},
			{Code: `<div aria-valuemax={"123"} />`, Tsx: true},
			{Code: "<div aria-valuemax={`123`} />", Tsx: true},
			{Code: `<div aria-valuemax="123" />`, Tsx: true},
			{Code: `<div aria-valuemax={foo} />`, Tsx: true},
			{Code: `<div aria-valuemax={foo.bar} />`, Tsx: true},
			{Code: `<div aria-valuemax={null} />`, Tsx: true},
			{Code: `<div aria-valuemax={undefined} />`, Tsx: true},

			// aria-sort (token: ['ascending','descending','none','other']).
			// Case-insensitive token, template form (no boolean coerce),
			// Identifier (skip).
			{Code: `<div aria-sort="ascending" />`, Tsx: true},
			{Code: `<div aria-sort="ASCENDING" />`, Tsx: true},
			{Code: `<div aria-sort={"ascending"} />`, Tsx: true},
			{Code: "<div aria-sort={`ascending`} />", Tsx: true},
			{Code: `<div aria-sort="descending" />`, Tsx: true},
			{Code: `<div aria-sort={"descending"} />`, Tsx: true},
			{Code: "<div aria-sort={`descending`} />", Tsx: true},
			{Code: `<div aria-sort="none" />`, Tsx: true},
			{Code: `<div aria-sort={"none"} />`, Tsx: true},
			{Code: "<div aria-sort={`none`} />", Tsx: true},
			{Code: `<div aria-sort="other" />`, Tsx: true},
			{Code: `<div aria-sort={"other"} />`, Tsx: true},
			{Code: "<div aria-sort={`other`} />", Tsx: true},
			{Code: `<div aria-sort={foo} />`, Tsx: true},
			{Code: `<div aria-sort={foo.bar} />`, Tsx: true},

			// aria-invalid (token: ['grammar', false, 'spelling', true]).
			// Heterogeneous list — booleans AND strings both valid.
			{Code: `<div aria-invalid={true} />`, Tsx: true},
			{Code: `<div aria-invalid="true" />`, Tsx: true},
			{Code: `<div aria-invalid={false} />`, Tsx: true},
			{Code: `<div aria-invalid="false" />`, Tsx: true},
			{Code: `<div aria-invalid="grammar" />`, Tsx: true},
			{Code: `<div aria-invalid="spelling" />`, Tsx: true},
			{Code: `<div aria-invalid={null} />`, Tsx: true},
			{Code: `<div aria-invalid={undefined} />`, Tsx: true},

			// aria-relevant (tokenlist: ['additions','all','removals','text']).
			// Single token / multi-token / repeated / out-of-order / template
			// / Identifier / null / undefined.
			{Code: `<div aria-relevant="additions" />`, Tsx: true},
			{Code: `<div aria-relevant={"additions"} />`, Tsx: true},
			{Code: "<div aria-relevant={`additions`} />", Tsx: true},
			{Code: `<div aria-relevant="additions removals" />`, Tsx: true},
			{Code: `<div aria-relevant="additions additions" />`, Tsx: true},
			{Code: `<div aria-relevant={"additions removals"} />`, Tsx: true},
			{Code: "<div aria-relevant={`additions removals`} />", Tsx: true},
			{Code: `<div aria-relevant="additions removals text" />`, Tsx: true},
			{Code: `<div aria-relevant={"additions removals text"} />`, Tsx: true},
			{Code: "<div aria-relevant={`additions removals text`} />", Tsx: true},
			{Code: `<div aria-relevant="additions removals text all" />`, Tsx: true},
			{Code: `<div aria-relevant={"additions removals text all"} />`, Tsx: true},
			{Code: "<div aria-relevant={`removals additions text all`} />", Tsx: true},
			{Code: `<div aria-relevant={foo} />`, Tsx: true},
			{Code: `<div aria-relevant={foo.bar} />`, Tsx: true},
			{Code: `<div aria-relevant={null} />`, Tsx: true},
			{Code: `<div aria-relevant={undefined} />`, Tsx: true},

			// aria-activedescendant (id). ANY string is valid — upstream's
			// validityCheck for "id" only inspects typeof. Token-looking
			// strings included by upstream to assert the absence of token
			// validation for the id type.
			{Code: `<div aria-activedescendant="ascending" />`, Tsx: true},
			{Code: `<div aria-activedescendant="ASCENDING" />`, Tsx: true},
			{Code: `<div aria-activedescendant={"ascending"} />`, Tsx: true},
			{Code: "<div aria-activedescendant={`ascending`} />", Tsx: true},
			{Code: `<div aria-activedescendant="descending" />`, Tsx: true},
			{Code: `<div aria-activedescendant={"descending"} />`, Tsx: true},
			{Code: "<div aria-activedescendant={`descending`} />", Tsx: true},
			{Code: `<div aria-activedescendant="none" />`, Tsx: true},
			{Code: `<div aria-activedescendant={"none"} />`, Tsx: true},
			{Code: "<div aria-activedescendant={`none`} />", Tsx: true},
			{Code: `<div aria-activedescendant="other" />`, Tsx: true},
			{Code: `<div aria-activedescendant={"other"} />`, Tsx: true},
			{Code: "<div aria-activedescendant={`other`} />", Tsx: true},
			{Code: `<div aria-activedescendant={foo} />`, Tsx: true},
			{Code: `<div aria-activedescendant={foo.bar} />`, Tsx: true},
			{Code: `<div aria-activedescendant={null} />`, Tsx: true},
			{Code: `<div aria-activedescendant={undefined} />`, Tsx: true},

			// aria-labelledby (idlist). ANY string is valid — `idlist`'s
			// inner `validityCheck(token, 'id', [])` reduces to "is string"
			// since split always returns string tokens.
			{Code: `<div aria-labelledby="additions" />`, Tsx: true},
			{Code: `<div aria-labelledby={"additions"} />`, Tsx: true},
			{Code: "<div aria-labelledby={`additions`} />", Tsx: true},
			{Code: `<div aria-labelledby="additions removals" />`, Tsx: true},
			{Code: `<div aria-labelledby="additions additions" />`, Tsx: true},
			{Code: `<div aria-labelledby={"additions removals"} />`, Tsx: true},
			{Code: "<div aria-labelledby={`additions removals`} />", Tsx: true},
			{Code: `<div aria-labelledby="additions removals text" />`, Tsx: true},
			{Code: `<div aria-labelledby={"additions removals text"} />`, Tsx: true},
			{Code: "<div aria-labelledby={`additions removals text`} />", Tsx: true},
			{Code: `<div aria-labelledby="additions removals text all" />`, Tsx: true},
			{Code: `<div aria-labelledby={"additions removals text all"} />`, Tsx: true},
			{Code: "<div aria-labelledby={`removals additions text all`} />", Tsx: true},
			{Code: `<div aria-labelledby={foo} />`, Tsx: true},
			{Code: `<div aria-labelledby={foo.bar} />`, Tsx: true},
			{Code: `<div aria-labelledby={null} />`, Tsx: true},
			{Code: `<div aria-labelledby={undefined} />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// aria-hidden (boolean) — non-coerced string, number, template
			// with substitution (placeholder string).
			{
				Code: `<div aria-hidden="yes" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-hidden")}},
			},
			{
				Code: `<div aria-hidden="no" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-hidden")}},
			},
			{
				Code: `<div aria-hidden={1234} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-hidden")}},
			},
			{
				Code: "<div aria-hidden={`${abc}`} />", Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-hidden")}},
			},

			// aria-label (string) — boolean form, boolean literals, number,
			// PrefixUnary boolean.
			{
				Code: `<div aria-label />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-label")}},
			},
			{
				Code: `<div aria-label={true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-label")}},
			},
			{
				Code: `<div aria-label={false} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-label")}},
			},
			{
				Code: `<div aria-label={1234} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-label")}},
			},
			{
				Code: `<div aria-label={!true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-label")}},
			},

			// aria-checked (tristate) — non-mixed string, number, template
			// with substitution.
			{
				Code: `<div aria-checked="yes" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-checked")}},
			},
			{
				Code: `<div aria-checked="no" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-checked")}},
			},
			{
				Code: `<div aria-checked={1234} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-checked")}},
			},
			{
				Code: "<div aria-checked={`${abc}`} />", Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-checked")}},
			},

			// aria-level (integer) — non-numeric strings, boolean, boolean-form,
			// coerced "false" boolean, PrefixUnary boolean.
			{
				Code: `<div aria-level="yes" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-level")}},
			},
			{
				Code: `<div aria-level="no" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-level")}},
			},
			{
				Code: "<div aria-level={`abc`} />", Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-level")}},
			},
			{
				Code: `<div aria-level={true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-level")}},
			},
			{
				Code: `<div aria-level />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-level")}},
			},
			{
				Code: `<div aria-level={"false"} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-level")}},
			},
			{
				Code: `<div aria-level={!"false"} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-level")}},
			},

			// aria-valuemax (number) — same shape as aria-level.
			{
				Code: `<div aria-valuemax="yes" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-valuemax")}},
			},
			{
				Code: `<div aria-valuemax="no" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-valuemax")}},
			},
			{
				Code: "<div aria-valuemax={`abc`} />", Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-valuemax")}},
			},
			{
				Code: `<div aria-valuemax={true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-valuemax")}},
			},
			{
				Code: `<div aria-valuemax />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-valuemax")}},
			},
			{
				Code: `<div aria-valuemax={"false"} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-valuemax")}},
			},
			{
				Code: `<div aria-valuemax={!"false"} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-valuemax")}},
			},

			// aria-sort (token) — empty string, typo, boolean form, boolean
			// literal, coerced "false" boolean, multi-token string.
			{
				Code: `<div aria-sort="" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-sort")}},
			},
			{
				Code: `<div aria-sort="descnding" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-sort")}},
			},
			{
				Code: `<div aria-sort />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-sort")}},
			},
			{
				Code: `<div aria-sort={true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-sort")}},
			},
			{
				Code: `<div aria-sort={"false"} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-sort")}},
			},
			{
				Code: `<div aria-sort="ascending descending" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-sort")}},
			},

			// aria-relevant (tokenlist) — empty string, unknown token, boolean
			// form, boolean literal, coerced "false" boolean, typo, trailing
			// space (split yields empty trailing token).
			{
				Code: `<div aria-relevant="" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-relevant")}},
			},
			{
				Code: `<div aria-relevant="foobar" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-relevant")}},
			},
			{
				Code: `<div aria-relevant />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-relevant")}},
			},
			{
				Code: `<div aria-relevant={true} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-relevant")}},
			},
			{
				Code: `<div aria-relevant={"false"} />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-relevant")}},
			},
			{
				Code: `<div aria-relevant="additions removalss" />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-relevant")}},
			},
			{
				Code: `<div aria-relevant="additions removalss " />`, Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidAriaPropType", Message: upstreamErrorMessage("aria-relevant")}},
			},
		})
}
