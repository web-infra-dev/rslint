package prefer_string_starts_ends_with_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_string_starts_ends_with"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Generated from the pinned reference, including each test() group, snapshot
// input, runtime assertion input and documentation example. Parser-only repeats
// are deduplicated; the group records where each source first appeared.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-string-starts-ends-with.js
//
//go:embed testdata/upstream.json
var upstreamJSON []byte

func TestPreferStringStartsEndsWithUpstream(t *testing.T) {
	var cases []struct {
		Group string `json:"group"`
		rule_tester.InvalidTestCase
	}
	if err := json.Unmarshal(upstreamJSON, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 358 {
		t.Fatalf("expected 358 upstream cases, got %d", len(cases))
	}
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, test := range cases {
		if len(test.Errors) == 0 {
			valid = append(valid, rule_tester.ValidTestCase{Code: test.Code, FileName: test.FileName})
		} else {
			invalid = append(invalid, test.InvalidTestCase)
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_string_starts_ends_with.PreferStringStartsEndsWithRule, valid, invalid)
}
