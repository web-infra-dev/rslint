package no_focused_tests_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_focused_tests"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoFocusedTests(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_focused_tests.NoFocusedTestsRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{},
	)
}
