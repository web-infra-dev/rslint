package no_commented_out_tests

import jestNoCommentedOutTests "github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_commented_out_tests"

var NoCommentedOutTestsRule = jestNoCommentedOutTests.NewRule(jestNoCommentedOutTests.Options{
	Name: "rstest/no-commented-out-tests",
})
