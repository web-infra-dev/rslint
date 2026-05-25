package no_access_key

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// expectedError mirrors upstream's expected error shape — every invalid case
// emits the same single message. Centralized so a future text tweak lives in
// one place. Shared with no_access_key_extras_test.go.
var expectedError = rule_tester.InvalidTestCaseError{
	MessageId: "noAccessKey",
	Message:   errorMessage,
}

// TestNoAccessKeyUpstream covers the full valid/invalid suite migrated 1:1
// from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/no-access-key-test.js`. Order and grouping mirror the
// upstream file so a future audit can grep across both side-by-side.
//
// rslint-specific lock-ins (TS wrappers, spread literal, listener boundary
// repeats, position assertions, the upstream getPropValue-branch
// classifications that upstream never tests directly) live in
// no_access_key_extras_test.go.
func TestNoAccessKeyUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoAccessKeyRule, []rule_tester.ValidTestCase{
		{Code: `<div />;`, Tsx: true},
		{Code: `<div {...props} />`, Tsx: true},
		{Code: `<div accessKey={undefined} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `<div accesskey="h" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div accessKey="h" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div accessKey="h" {...props} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div acCesSKeY="y" />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div accessKey={"y"} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<div accessKey={`${y}`} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<div accessKey={`${undefined}y${undefined}`} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<div accessKey={`This is ${bad}`} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   `<div accessKey={accessKey} />`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<div accessKey={`${undefined}`} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
		{
			Code:   "<div accessKey={`${undefined}${undefined}`} />",
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{expectedError},
		},
	})
}
