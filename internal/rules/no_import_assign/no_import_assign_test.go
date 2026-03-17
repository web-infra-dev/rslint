package no_import_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImportAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoImportAssignRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Default import - member write is allowed
			{Code: `import mod from 'mod'; mod.prop = 0`},

			// Named import - member write is allowed
			{Code: `import {named} from 'mod'; named.prop = 0`},

			// Namespace import - nested member write is allowed
			{Code: `import * as mod from 'mod'; mod.named.prop = 0`},

			// Read-only usage (no writes)
			{Code: `import mod from 'mod'; console.log(mod)`},
			{Code: `import {named} from 'mod'; console.log(named)`},
			{Code: `import * as mod from 'mod'; console.log(mod)`},

			// Calling imports is not a write
			{Code: `import mod from 'mod'; mod()`},
			{Code: `import {named} from 'mod'; named()`},

			// Reading namespace member is not a write
			{Code: `import * as mod from 'mod'; console.log(mod.named)`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Default import - direct assignment
			{
				Code: `import mod from 'mod'; mod = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Line: 1, Column: 24},
				},
			},

			// Default import - increment
			{
				Code: `import mod from 'mod'; mod++`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Line: 1, Column: 24},
				},
			},

			// Named import - direct assignment
			{
				Code: `import {named} from 'mod'; named = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Line: 1, Column: 28},
				},
			},

			// Named import - increment
			{
				Code: `import {named} from 'mod'; named++`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Line: 1, Column: 28},
				},
			},

			// Namespace import - direct assignment
			{
				Code: `import * as mod from 'mod'; mod = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Line: 1, Column: 29},
				},
			},

			// Namespace import - member assignment
			{
				Code: `import * as mod from 'mod'; mod.named = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 29},
				},
			},

			// Namespace import - member increment
			{
				Code: `import * as mod from 'mod'; mod.named++`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 29},
				},
			},
		},
	)
}
