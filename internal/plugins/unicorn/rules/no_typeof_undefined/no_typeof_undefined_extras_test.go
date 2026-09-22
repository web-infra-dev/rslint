package no_typeof_undefined_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_typeof_undefined"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoTypeofUndefinedMultilineBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_typeof_undefined.NoTypeofUndefinedRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			invalidFixed(
				"function f(value) { throw typeof // comment\nvalue === \"undefined\"; }",
				"function f(value) { throw ( // comment\nvalue === undefined); }",
			),
			invalidFixed(
				"function f(value) { const check = typeof // comment\nvalue === \"undefined\"; return check; }",
				"function f(value) { const check = // comment\nvalue === undefined; return check; }",
			),
		},
	)
}
