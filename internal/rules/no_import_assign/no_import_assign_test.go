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

			// Object/Reflect mutation functions on non-namespace imports are allowed
			{Code: `import mod from 'mod'; Object.assign(mod, {prop: 1})`},

			// Namespace import as second argument is fine
			{Code: `import * as mod from 'mod'; Object.assign({}, mod)`},

			// Block-scoped shadow - reassigning local variable is fine
			{Code: `import mod from 'mod'; { let mod = 0; mod = 1; }`},
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

			// Namespace import - element access assignment
			{
				Code: `import * as mod from 'mod'; mod["named"] = 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 29},
				},
			},

			// Namespace import - delete member
			{
				Code: `import * as mod from 'mod'; delete mod.named`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 36},
				},
			},

			// Namespace import - Object.assign(ns, ...)
			{
				Code: `import * as mod from 'mod'; Object.assign(mod, {prop: 1})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 43},
				},
			},

			// Namespace import - Object.defineProperty(ns, ...)
			{
				Code: `import * as mod from 'mod'; Object.defineProperty(mod, 'prop', {value: 1})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 51},
				},
			},

			// Namespace import - Object.defineProperties(ns, ...)
			{
				Code: `import * as mod from 'mod'; Object.defineProperties(mod, {prop: {value: 1}})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 53},
				},
			},

			// Namespace import - Object.setPrototypeOf(ns, ...)
			{
				Code: `import * as mod from 'mod'; Object.setPrototypeOf(mod, {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 51},
				},
			},

			// Namespace import - Reflect.defineProperty(ns, ...)
			{
				Code: `import * as mod from 'mod'; Reflect.defineProperty(mod, 'prop', {value: 1})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 52},
				},
			},

			// Namespace import - Reflect.deleteProperty(ns, ...)
			{
				Code: `import * as mod from 'mod'; Reflect.deleteProperty(mod, 'prop')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 52},
				},
			},

			// Namespace import - Reflect.set(ns, ...)
			{
				Code: `import * as mod from 'mod'; Reflect.set(mod, 'prop', 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 41},
				},
			},

			// Namespace import - Reflect.setPrototypeOf(ns, ...)
			{
				Code: `import * as mod from 'mod'; Reflect.setPrototypeOf(mod, {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonlyMember", Line: 1, Column: 52},
				},
			},

			// Default import - compound assignment
			{
				Code: `import mod from 'mod'; mod += 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Line: 1, Column: 24},
				},
			},

			// Default import - for-in
			{
				Code: `import mod from 'mod'; for (mod in foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "readonly", Line: 1, Column: 29},
				},
			},
		},
	)
}
