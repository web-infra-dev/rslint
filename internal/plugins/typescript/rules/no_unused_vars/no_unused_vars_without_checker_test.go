package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnusedVarsWithoutChecker covers the rule on a file linted without a
// type checker. Usages come from the binder either way, and the checker only
// backs a few fallbacks, so the rule has to fall back to the binder's answer
// instead of dereferencing a checker that is not there. A Vue component is
// always linted without one, which makes it the vehicle here.
func TestNoUnusedVarsWithoutChecker(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     "<script lang=\"ts\">\nconst value = 1;\nconsole.log(value);\n</script>\n",
				FileName: "used-binding.vue",
			},
			// Local export targets.
			{
				Code:     "<script lang=\"ts\">\nconst local = 1;\nexport { local };\n</script>\n",
				FileName: "exported-local.vue",
			},
			// Overload signatures without a body.
			{
				Code:     "<script lang=\"ts\">\nfunction format(value: string): string;\nfunction format(value: number): string;\nfunction format(value: unknown): string {\n  return String(value);\n}\nconsole.log(format(1));\n</script>\n",
				FileName: "overloads.vue",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Merged symbol lookup for a declaration in a global file.
			{
				Code:     "<script lang=\"ts\">\ninterface Shape { a: number }\n</script>\n",
				FileName: "unused-interface.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2}},
			},
			// Namespace augmentation check.
			{
				Code:     "<script lang=\"ts\">\nnamespace Unused {}\n</script>\n",
				FileName: "unused-namespace.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2}},
			},
		},
	)
}
