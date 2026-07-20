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
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		suite.Valid,
		suite.Invalid,
	)
}
