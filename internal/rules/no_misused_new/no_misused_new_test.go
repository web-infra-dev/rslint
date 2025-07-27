package no_misused_new

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestNoMisusedNewRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// Add valid test cases here
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// Add invalid test cases here
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedNewRule, validTestCases, invalidTestCases)
}