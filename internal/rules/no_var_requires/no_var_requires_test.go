package no_var_requires

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoVarRequiresRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoVarRequiresRule, []rule_tester.ValidTestCase{
		{Code: "import foo = require('foo');"},
		{Code: "import foo from 'foo';"},
		{Code: "import * as foo from 'foo';"},
		{Code: "import { bar } from 'foo';"},
		{Code: "require('foo');"},
		{Code: "require?.('foo');"},
		{Code: "const foo = require('foo');", Options: &Options{Allow: []string{"/foo/"}}},
		{Code: "const foo = require('./foo');", Options: &Options{Allow: []string{"/foo/"}}},
		{Code: "const foo = require('../foo');", Options: &Options{Allow: []string{"/foo/"}}},
		{Code: "const foo = require('foo/bar');", Options: &Options{Allow: []string{"/bar/"}}},
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
			Options: &Options{Allow: []string{"/bar/"}},
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
			Options: &Options{Allow: []string{"/bar/"}},
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
