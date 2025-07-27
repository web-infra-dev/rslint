package no_useless_constructor

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestNoUselessConstructorRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// TODO: Add valid test cases
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// TODO: Add invalid test cases
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUselessConstructorRule, validTestCases, invalidTestCases)
}