package no_focused_tests

import (
	jestNoFocusedTests "github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_focused_tests"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Rstest has no `fit`/`fdescribe` aliases, so no focus prefix is configured;
// only the `.only` modifier is reported.
var NoFocusedTestsRule rule.Rule = jestNoFocusedTests.NewRule(jestNoFocusedTests.Config{
	Name:        "rstest/no-focused-tests",
	ParseConfig: rstestUtils.RstestFnCallParseConfig(),
})
