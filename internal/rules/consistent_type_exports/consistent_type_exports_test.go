package consistent_type_exports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestConsistentTypeExports(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeExportsRule, []rule_tester.ValidTestCase{
		// Basic valid cases that don't need external modules
		{
			Code: `
				const value = 1;
				export { value };
			`,
		},
		{
			Code: `
				type Type = string;
				export type { Type };
			`,
		},
		{
			Code: `
				export { unknown } from 'unknown-module';
			`,
		},
	}, []rule_tester.InvalidTestCase{
		// Basic invalid cases - test without fixes first to debug
	})
}