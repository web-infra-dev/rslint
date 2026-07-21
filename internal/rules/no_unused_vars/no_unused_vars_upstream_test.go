package no_unused_vars

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"gotest.tools/v3/assert"
)

// TestNoUnusedVarsUpstream migrates the full valid/invalid suite from
// eslint/eslint tests/lib/rules/no-unused-vars.js at fb09aa8 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific lock-in
// cases live in no_unused_vars_extras_test.go.
//
//go:embed no_unused_vars_upstream.json
var upstreamSuiteJSON []byte

func TestNoUnusedVarsUpstream(t *testing.T) {
	var suite rule_tester.TestSuite
	assert.NilError(t, json.Unmarshal(upstreamSuiteJSON, &suite))
	// These counts come from executing RuleTester.run in the pinned upstream
	// source. Keep them explicit so a dropped or duplicated generated case
	// cannot hide behind a still-valid JSON fixture.
	assert.Equal(t, len(suite.Valid), 167)
	assert.Equal(t, len(suite.Invalid), 281)
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
	// The remaining skips require either ESLint's cross-rule
	// markVariableAsUsed API or parsing invalid strict-mode legacy octal syntax.
	assert.Equal(t, skipped, 8)

	var provenance struct {
		Upstream struct {
			Rule          string `json:"rule"`
			ESLintVersion string `json:"eslintVersion"`
			ESLintCommit  string `json:"eslintCommit"`
		} `json:"upstream"`
	}
	assert.NilError(t, json.Unmarshal(upstreamSuiteJSON, &provenance))
	assert.Equal(t, provenance.Upstream.Rule, "no-unused-vars")
	assert.Equal(t, provenance.Upstream.ESLintVersion, "10.7.0")
	assert.Equal(t, provenance.Upstream.ESLintCommit, "fb09aa8ff09730d3ccf68859e065f99666b52466")

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		suite.Valid,
		suite.Invalid,
	)
}
