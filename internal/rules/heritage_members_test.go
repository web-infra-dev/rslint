package rules_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules"
)

// The same corpus is checked against ESLint and its TypeScript parser in the
// JS suite. Keep the expectations independent of the compiler's node kinds:
// heritage member names and ordinary qualified type names have different
// ESTree semantics even when the compiler represents both as QualifiedName.
func TestHeritageMemberSemantics(t *testing.T) {
	data, err := os.ReadFile("testdata/heritage_members.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Rule    string                             `json:"rule"`
		Name    string                             `json:"name"`
		Code    string                             `json:"code"`
		Options []any                              `json:"options"`
		Globals map[string]any                     `json:"globals"`
		Errors  []rule_tester.InvalidTestCaseError `json:"errors"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("heritage member corpus must not be empty")
	}
	for _, test := range cases {
		t.Run(test.Rule+"/"+test.Name, func(t *testing.T) {
			ruleImpl, ok := rules.All().Lookup(test.Rule)
			if !ok {
				t.Fatalf("unknown rule %q", test.Rule)
			}
			var valid []rule_tester.ValidTestCase
			var invalid []rule_tester.InvalidTestCase
			if len(test.Errors) == 0 {
				valid = []rule_tester.ValidTestCase{{Code: test.Code, Options: test.Options, Globals: test.Globals}}
			} else {
				invalid = []rule_tester.InvalidTestCase{{Code: test.Code, Options: test.Options, Globals: test.Globals, Errors: test.Errors}}
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ruleImpl, valid, invalid)
		})
	}
}
