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
			{Code: "foo(bar(() => baz(qux(zed()))));"},
			{Code: "foo();", Options: []any{map[string]any{}}},
			{Code: "foo(import('x'));", Options: []any{map[string]any{"max": 1}}},
			{Code: "foo(bar(baz(import('x'))));"},
			{Code: "outer(inner({[key](){ foo(); }}));", Options: []any{map[string]any{"max": 2}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "foo(bar());", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 1.", Line: 1, Column: 5, EndLine: 1, EndColumn: 10}}, Options: []any{map[string]any{"max": 1}}},
			{Code: "foo(bar(baz(qux())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 3."}}, Options: []any{map[string]any{}}},
			{Code: "outer(inner({[foo()](){}}));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 2."}}, Options: []any{map[string]any{"max": 2}}},
			{Code: "outer(inner({get [foo()](){ return value; }}));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 2."}}, Options: []any{map[string]any{"max": 2}}},
			{Code: "outer(inner({[foo()]: value}));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 2."}}, Options: []any{map[string]any{"max": 2}}},
		},
	)
}
