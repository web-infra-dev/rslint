// TestNoImportTypeSideEffectsUpstream migrates the full valid/invalid suite
// from upstream
// https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/no-import-type-side-effects.test.ts
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in
// no_import_type_side_effects_extras_test.go.
package no_import_type_side_effects

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImportTypeSideEffectsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoImportTypeSideEffectsRule, []rule_tester.ValidTestCase{
		{Code: `import T from 'mod';`},
		{Code: `import * as T from 'mod';`},
		{Code: `import { T } from 'mod';`},
		{Code: `import type { T } from 'mod';`},
		{Code: `import type { T, U } from 'mod';`},
		{Code: `import { type T, U } from 'mod';`},
		{Code: `import { T, type U } from 'mod';`},
		{Code: `import type T from 'mod';`},
		{Code: `import T, { type U } from 'mod';`},
		{Code: `import type * as T from 'mod';`},
		{Code: `import 'mod';`},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `import { type A } from 'mod';`,
			Output: []string{`import type { A } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		{
			Code:   `import { type A as AA } from 'mod';`,
			Output: []string{`import type { A as AA } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		{
			Code:   `import { type A, type B } from 'mod';`,
			Output: []string{`import type { A, B } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		{
			Code:   `import { type A as AA, type B as BB } from 'mod';`,
			Output: []string{`import type { A as AA, B as BB } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
	})
}
