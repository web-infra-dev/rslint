package no_unnecessary_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryConditionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule,
		[]rule_tester.ValidTestCase{
			// Valid cases - conditions that are necessary
			{Code: `
declare const x: string | null;
if (x) {
	console.log(x);
}`},
			{Code: `
declare const x: number | undefined;
while (x) {
	console.log(x);
}`},
			{Code: `
declare const x: boolean;
if (x) {
	console.log('true');
}`},
			// Valid with allowConstantLoopConditions
			{
				Code:    `while (true) { break; }`,
				Options: map[string]any{"allowConstantLoopConditions": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Always truthy conditions
			{
				Code: `
declare const x: string;
if (x) {
	console.log(x);
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "alwaysTruthy",
					Line:      3,
					Column:    5,
				}},
			},
			// Always falsy conditions
			{
				Code: `
declare const x: null;
if (x) {
	console.log(x);
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "alwaysFalsy",
					Line:      3,
					Column:    5,
				}},
			},
			// Never type
			{
				Code: `
declare const x: never;
if (x) {
	console.log(x);
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "never",
					Line:      3,
					Column:    5,
				}},
			},
			// Unnecessary nullish coalescing
			{
				Code: `
declare const x: string;
const y = x ?? 'default';`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "neverNullish",
					Line:      3,
					Column:    11,
				}},
			},
		},
	)
}

func TestNoUnnecessaryConditionRuleWithStrictNullChecks(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    `declare const x: any; if (x) { }`,
				Options: map[string]any{"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `const x = null; if (x) { }`,
				Options: map[string]any{"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "alwaysFalsy",
					Line:      1,
					Column:    21,
				}},
			},
		},
	)
}
