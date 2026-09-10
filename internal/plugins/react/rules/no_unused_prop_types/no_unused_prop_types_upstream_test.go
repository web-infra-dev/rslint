// TestNoUnusedPropTypesUpstream runs the complete valid/invalid suite from
// eslint-plugin-react v7.37.5 tests/lib/rules/no-unused-prop-types.js. The
// fixture preserves every supported upstream case. The sibling
// no_unused_prop_types_extras_test.go file contains rslint-specific cases.
package no_unused_prop_types

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"gotest.tools/v3/assert"
)

//go:embed no_unused_prop_types_upstream.json
var upstreamSuiteJSON []byte

func TestNoUnusedPropTypesUpstream(t *testing.T) {
	var suite struct {
		Upstream struct {
			Rule          string `json:"Rule"`
			PluginVersion string `json:"PluginVersion"`
			Source        string `json:"Source"`
		} `json:"Upstream"`
		Valid   []rule_tester.ValidTestCase   `json:"Valid"`
		Invalid []rule_tester.InvalidTestCase `json:"Invalid"`
	}
	assert.NilError(t, json.Unmarshal(upstreamSuiteJSON, &suite))
	assert.Equal(t, suite.Upstream.Rule, "react/no-unused-prop-types")
	assert.Equal(t, suite.Upstream.PluginVersion, "7.37.5")
	assert.Equal(t, len(suite.Valid), 243)
	assert.Equal(t, len(suite.Invalid), 118)
	skipped := 0
	for _, testCase := range suite.Valid {
		if testCase.Skip {
			skipped++
		}
	}
	for _, testCase := range suite.Invalid {
		if testCase.Skip {
			skipped++
		}
	}
	assert.Equal(t, skipped, 4)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedPropTypesRule,
		suite.Valid,
		suite.Invalid,
	)
}
