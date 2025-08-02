package no_import_type_side_effects

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoImportTypeSideEffectsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoImportTypeSideEffectsRule, []rule_tester.ValidTestCase{
		// Valid cases
		{Code: "import T from 'mod';"},
		{Code: "import * as T from 'mod';"},
		{Code: "import { T } from 'mod';"},
		{Code: "import type { T } from 'mod';"},
		{Code: "import type { T, U } from 'mod';"},
		{Code: "import { type T, U } from 'mod';"},
		{Code: "import { T, type U } from 'mod';"},
		{Code: "import type T from 'mod';"},
		{Code: "import type T, { U } from 'mod';"},
		{Code: "import T, { type U } from 'mod';"},
		{Code: "import type * as T from 'mod';"},
		{Code: "import 'mod';"},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases
		{
			Code:   "import { type A } from 'mod';",
			Output: []string{"import type { A } from 'mod';"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier"},
			},
		},
		{
			Code:   "import { type A as AA } from 'mod';",
			Output: []string{"import type { A as AA } from 'mod';"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier"},
			},
		},
		{
			Code:   "import { type A, type B } from 'mod';",
			Output: []string{"import type { A, B } from 'mod';"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier"},
			},
		},
		{
			Code:   "import { type A as AA, type B as BB } from 'mod';",
			Output: []string{"import type { A as AA, B as BB } from 'mod';"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier"},
			},
		},
	})
}
