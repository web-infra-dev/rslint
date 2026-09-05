package no_deprecated

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDeprecatedHeritage(t *testing.T) {
	// ESLint reports each member in the heritage expression independently;
	// React.DOM.div therefore reports React.DOM, not the complete type name.
	settings := map[string]any{"react": map[string]any{"version": "19.0.0"}}
	var valid []rule_tester.ValidTestCase
	for _, code := range []string{
		`type I = React.PropTypes;`,
		`type I = typeof React.PropTypes;`,
		`type I = React.DOM.div;`,
		`interface I extends Box<React.PropTypes> {}`,
		`class C implements Box<React.DOM.div> {}`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code, Settings: settings})
	}
	var invalid []rule_tester.InvalidTestCase
	for _, context := range []struct {
		prefix string
		column int
	}{
		{"interface I extends ", 21},
		{"class C implements ", 20},
		{"class C extends ", 17},
	} {
		for _, member := range []struct {
			name    string
			width   int
			message string
		}{
			{"React.PropTypes", 15, msg("React.PropTypes", "15.5.0", "the npm module prop-types", "")},
			{"React.DOM.div", 9, msg("React.DOM", "15.6.0", "the npm module react-dom-factories", "")},
		} {
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: context.prefix + member.name + " {}", Settings: settings,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "deprecated", Message: member.message,
					Line: 1, Column: context.column, EndLine: 1, EndColumn: context.column + member.width,
				}},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDeprecatedRule, valid, invalid)
}
