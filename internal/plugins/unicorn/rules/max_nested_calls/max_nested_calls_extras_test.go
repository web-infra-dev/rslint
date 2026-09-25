package max_nested_calls_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/max_nested_calls"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMaxNestedCallsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&max_nested_calls.MaxNestedCallsRule,
		[]rule_tester.ValidTestCase{
			{Code: "foo(() => bar(baz(qux(zed()))));"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "foo(bar());", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}, Options: []any{map[string]any{"max": 1}}},
		},
	)
}
