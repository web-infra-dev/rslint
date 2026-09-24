package prefer_mock_promise_shorthand

import (
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_mock_promise_shorthand"
)

var PreferMockPromiseShorthandRule = shared.NewRule(shared.Config{
	Name: "jest/prefer-mock-promise-shorthand",
})
