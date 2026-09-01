// TestNoInvalidHtmlAttributeRuleExtras locks in branches and edge shapes the
// upstream suite does not exercise. Every case identifies its Dimension 4,
// real-user, or upstream-branch purpose; migrated cases live in
// no_invalid_html_attribute_upstream_test.go.
package no_invalid_html_attribute

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidHtmlAttributeRuleExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInvalidHtmlAttributeRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized expression ----
		// Parentheses are transparent in ESTree at createElement call sites.
		{Code: `((React).createElement)(("a"), ({rel: "alternate"}))`, Tsx: true},
		{Code: `React.createElement(tag, {rel: "alternate"})`, Tsx: true},
		{Code: "React.createElement(`a`, {rel: \"alternate\"})", Tsx: true},
		{Code: `h.createElement("a", {rel: "invalid"})`, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "h"}}, Tsx: true},
		// ---- Dimension 4: JSX member and namespace tag names ----
		{Code: `var x = <Foo.Bar rel="invalid"/>`, Tsx: true},
		{Code: `var x = <svg:a rel="invalid"/>`, Tsx: true},
		// ---- Dimension 4: object members, arrays, spread, computed and shorthand ----
		{Code: `React.createElement("a", {rel: [call(), {value: "invalid"}]})`, Tsx: true},
		{Code: `React.createElement("a", {rel})`, Tsx: true},
		// ---- Dimension 4: option shape ----
		{Code: `var x = <a rel="invalid"/>`, Options: []any{[]interface{}{}}, Tsx: true},
		// N/A: optional chains, type wrappers, and computed JSX attributes cannot
		// be values of a JSX attribute in the upstream Literal-only branch.
		// ---- Real-user: eslint-plugin-react#3132 ----
		{Code: `var x = <link rel="apple-touch-icon" href="/icon.png"/>`, Tsx: true},
		// ---- Real-user: eslint-plugin-react#3132 ----
		{Code: `var x = <link rel="mask-icon" href="/pinned.svg"/>`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// Locks in upstream checkAttribute() arm: expression-container string literal.
		{Code: `var x = <a rel={"invalid"}/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `var x = <a rel={""}/>`}}}}},
		{Code: `var x = <a rel="noopener  noreferrer"/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spaceDelimited", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveWhitespaces", Output: `var x = <a rel="noopener noreferrer"/>`}}}}},
		{Code: `React.createElement("a", {rel: ["invalid", "canonical"]})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `React.createElement("a", {rel: ["", "canonical"]})`}}}, {MessageId: "notValidFor"}}},
		{Code: `React.createElement(1, {rel: "alternate"})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "onlyMeaningfulFor"}}},
		{Code: `var x = <a rel={undefined}/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "onlyStrings", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveDefault", Output: `var x = <a />`}}}}},
		// Locks in upstream checkCreateProps() arm: method takes precedence over values.
		{Code: `React.createElement("a", {rel() {}})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMethod"}}},
		{Code: `React.createElement("html", {rel})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "onlyMeaningfulFor"}}},
		{Code: `React.createElement("html", {[rel]: 1})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "onlyMeaningfulFor"}}},
		{Code: `React.createElement("a", {[rel]() { return 1; }})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noMethod"}}},
		{Code: `React.createElement("a", {rel: ("invalid")})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `React.createElement("a", {rel: ("")})`}}}}},
		{Code: `React.createElement("a", {rel: (["invalid"])})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `React.createElement("a", {rel: ([""])})`}}}}},
		{Code: `React.createElement("a", {rel: 1n})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `React.createElement("a", {rel: n})`}}}}},
		{Code: `React.createElement("a", {rel: "canonical"})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notValidFor", Message: `“"canonical"” is not a valid “rel” attribute value for <a>.`}}},
		{Code: `React.createElement("a", {rel: "canonical"})`, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "h"}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notValidFor", Message: `“"canonical"” is not a valid “rel” attribute value for <a>.`}}},
		{Code: `var x = <a rel={/invalid/}/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "onlyStrings", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveNonString", Output: `var x = <a />`}}}}},
		{Code: `React.createElement("a", {rel: /invalid/})`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `React.createElement("a", {rel: //})`}}}}},
		{Code: `var x = <a rel={1n}/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "onlyStrings", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveNonString", Output: `var x = <a />`}}}}},
		{Code: `var x = <a rel={("invalid")}/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `var x = <a rel={("")}/>`}}}}},
		{Code: `var x = <a rel={"invalid"}/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Line: 1, Column: 17, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `var x = <a rel={""}/>`}}}}},
		{Code: `var x = <a rel="invalid"/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Line: 1, Column: 16, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `var x = <a rel=""/>`}}}}},
		{Code: `var x = <a rel="noopener  noreferrer"/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spaceDelimited", Line: 1, Column: 16, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveWhitespaces", Output: `var x = <a rel="noopener noreferrer"/>`}}}}},
		// Locks in upstream checkLiteralValueNode() pair branch with a second token.
		{Code: `var x = <link rel="shortcut nope"/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `var x = <link rel="shortcut "/>`}}}, {MessageId: "notPaired"}}},
		{Code: "var x = <link rel=\"shortcut\tfoo\"/>", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: "var x = <link rel=\"shortcut\t\"/>"}}}, {MessageId: "spaceDelimited", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveWhitespaces", Output: `var x = <link rel="shortcut foo"/>`}}}}},
		// ---- Real-user: eslint-plugin-react#3172 ----
		{Code: `var x = <a rel="home"/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "neverValid", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `var x = <a rel=""/>`}}}}},
		// ---- Real-user: eslint-plugin-react#3172 ----
		{Code: `var x = <a rel="shortcut"/>`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notValidFor", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestRemoveInvalid", Output: `var x = <a rel=""/>`}}}, {MessageId: "notAlone"}}},
	})
}
