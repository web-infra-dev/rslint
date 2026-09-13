package process_exit_as_throw_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/process_exit_as_throw"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/consistent_return"
	"github.com/web-infra-dev/rslint/internal/rules/no_unreachable"
)

// This rule never reports. Exercise each upstream consumer through the normal
// linter with the same call predicate an enabled node/process-exit-as-throw adds.
func withProcessExit(consumer rule.Rule) rule.Rule {
	consumer.CallThrows = process_exit_as_throw.ProcessExitAsThrowRule.CallThrows
	return consumer
}

func unreachableAt(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unreachableCode", Message: "Unreachable code.",
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// All three tests from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/process-exit-as-throw.js
func TestProcessExitAsThrowUpstream(t *testing.T) {
	unreachable := withProcessExit(no_unreachable.NoUnreachableRule)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &unreachable, nil,
		[]rule_tester.InvalidTestCase{{
			Code:   "foo();\nprocess.exit(1);\nbar();",
			Errors: []rule_tester.InvalidTestCaseError{unreachableAt(3, 1, 3, 7)},
		}})
	t.Run("disabled", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unreachable.NoUnreachableRule,
			[]rule_tester.ValidTestCase{{Code: "foo();\nprocess.exit(1);\nbar();"}}, nil)
	})
	t.Run("consistent-return", func(t *testing.T) {
		consumer := withProcessExit(consistent_return.ConsistentReturnRule)
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer,
			[]rule_tester.ValidTestCase{{Code: `function foo() {
    if (a) {
        return 1;
    } else {
        process.exit(1);
    }
}`}}, nil)
	})
}

// The only JavaScript example in the pinned documentation, with its related rule:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/process-exit-as-throw.md
func TestProcessExitAsThrowDocumentation(t *testing.T) {
	consumer := withProcessExit(consistent_return.ConsistentReturnRule)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consumer,
		[]rule_tester.ValidTestCase{{Code: `function foo(a) {
    if (a) {
        return new Bar();
    } else {
        process.exit(1);
    }
}`}}, nil)
}
