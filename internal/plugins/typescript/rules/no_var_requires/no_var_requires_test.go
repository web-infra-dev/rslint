package no_var_requires

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoVarRequiresRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoVarRequiresRule, []rule_tester.ValidTestCase{
		{Code: "import foo = require('foo');"},
		{Code: "import foo from 'foo';"},
		{Code: "import * as foo from 'foo';"},
		{Code: "import { bar } from 'foo';"},
		{Code: "require('foo');"},
		{Code: "require?.('foo');"},
		{Code: "const foo = require('foo');", Options: map[string]interface{}{"allow": []interface{}{"foo"}}},
		{Code: "const foo = require('./foo');", Options: map[string]interface{}{"allow": []interface{}{"foo"}}},
		{Code: "const foo = require('../foo');", Options: map[string]interface{}{"allow": []interface{}{"foo"}}},
		{Code: "const foo = require('foo/bar');", Options: map[string]interface{}{"allow": []interface{}{"bar"}}},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "var foo = require('foo');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "const foo = require('foo');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: "let foo = require('foo');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    11,
				},
			},
		},
		{
			Code: "var foo = require('foo'), bar = require('bar');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    11,
				},
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    33,
				},
			},
		},
		{
			Code: "const { foo } = require('foo');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    17,
				},
			},
		},
		{
			Code: "const { foo, bar } = require('foo');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    22,
				},
			},
		},
		{
			Code: "const foo = require?.('foo');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code:    "const foo = require('foo'), bar = require('bar');",
			Options: map[string]interface{}{"allow": []interface{}{"bar"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code:    "const foo = require('./foo');",
			Options: map[string]interface{}{"allow": []interface{}{"bar"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: "const foo = require('foo') as Foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: "const foo: Foo = require('foo');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noVarReqs",
					Line:      1,
					Column:    18,
				},
			},
		},
	})
}

// TestNoVarRequiresAllowSchema locks in that each `allow` entry is validated
// as a regex. The rule compiles every entry to match require paths, so an
// unparsable one would otherwise be dropped and that path would keep
// reporting. Upstream's schema declares a plain array of strings.
func TestNoVarRequiresAllowSchema(t *testing.T) {
	invalid := []any{map[string]any{"allow": []any{"("}}}
	if err := NoVarRequiresRule.Schema.Validate(invalid); err == nil {
		t.Error("expected an invalid allow pattern to fail schema validation")
	}
	valid := []any{map[string]any{"allow": []any{"/package\\.json$"}}}
	if err := NoVarRequiresRule.Schema.Validate(valid); err != nil {
		t.Errorf("expected a valid allow pattern to pass schema validation, got: %v", err)
	}
}
