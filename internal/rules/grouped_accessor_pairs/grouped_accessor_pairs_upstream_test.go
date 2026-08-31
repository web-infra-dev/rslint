package grouped_accessor_pairs

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"gotest.tools/v3/assert"
)

// TestGroupedAccessorPairsUpstream migrates the complete ESLint v10.9.1
// grouped-accessor-pairs suite. rslint-specific edge-shape and branch lock-in
// cases live in grouped_accessor_pairs_extras_test.go.

//go:embed grouped_accessor_pairs_upstream.json
var upstreamSuiteJSON []byte

func TestGroupedAccessorPairsUpstream(t *testing.T) {
	var suite rule_tester.TestSuite
	assert.NilError(t, json.Unmarshal(upstreamSuiteJSON, &suite))
	assert.Equal(t, len(suite.Valid), 88)
	assert.Equal(t, len(suite.Invalid), 62)

	var provenance struct {
		Upstream struct {
			Rule          string `json:"rule"`
			ESLintVersion string `json:"eslintVersion"`
			ESLintCommit  string `json:"eslintCommit"`
		} `json:"upstream"`
	}
	assert.NilError(t, json.Unmarshal(upstreamSuiteJSON, &provenance))
	assert.Equal(t, provenance.Upstream.Rule, "grouped-accessor-pairs")
	assert.Equal(t, provenance.Upstream.ESLintVersion, "10.9.1")
	assert.Equal(t, provenance.Upstream.ESLintCommit, "5c8c2417b9ff462f2dc4e54a062c59135b45b845")

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&GroupedAccessorPairsRule,
		suite.Valid,
		suite.Invalid,
	)
}
