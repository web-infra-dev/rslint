package switch_exhaustiveness_check

import (
	"reflect"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSwitchExhaustivenessCheckEmptyOptionsUseDefaults(t *testing.T) {
	withoutOptions := parseOptions(nil)
	withEmptyObject := parseOptions([]any{map[string]any{}})
	if !reflect.DeepEqual(withoutOptions, withEmptyObject) {
		t.Fatalf("no options = %#v, [{}] = %#v", withoutOptions, withEmptyObject)
	}
}

func TestSwitchExhaustivenessCheckEmptyOptionsMatchDefaultDiagnostics(t *testing.T) {
	const code = `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
}
`
	const output = `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  case "b": { throw new Error('Not implemented yet: "b" case') }
}
`
	expected := []rule_tester.InvalidTestCaseError{
		{
			MessageId: "switchIsNotExhaustive",
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "addMissingCases", Output: output},
			},
		},
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SwitchExhaustivenessCheckRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{Code: code, Errors: expected},
			{Code: code, Options: []any{map[string]any{}}, Errors: expected},
		},
	)
}
