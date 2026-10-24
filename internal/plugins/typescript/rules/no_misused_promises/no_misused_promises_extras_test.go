package no_misused_promises

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMisusedPromisesArgumentTypes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedPromisesRule,
		[]rule_tester.ValidTestCase{
			{Code: `declare function invoke(): void; invoke();`},
			{Code: `declare function invoke(x: string): void; invoke("value");`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
export {};
declare global { interface String { (): Promise<void>; } }
declare function invoke(callback: () => void): void;
invoke("value");
`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "voidReturnArgument"}},
			},
		})
}
