package prefer_mock_return_shorthand

import (
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_mock_return_shorthand"
)

var PreferMockReturnShorthandRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-mock-return-shorthand",
})
