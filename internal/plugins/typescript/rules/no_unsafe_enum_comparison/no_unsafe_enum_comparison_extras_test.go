// no_unsafe_enum_comparison_extras_test.go covers rslint-specific edge cases.
// Migrated upstream cases live in no_unsafe_enum_comparison_upstream_test.go.
package no_unsafe_enum_comparison

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeEnumComparisonExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnsafeEnumComparisonRule, nil, []rule_tester.InvalidTestCase{
		{
			Code: `
enum Num {
  A = 1,
}
declare const num: Num;
num === (1);
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mismatchedCondition",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceValueWithEnum",
					Output: `
enum Num {
  A = 1,
}
declare const num: Num;
num === (Num.A);
`,
				}},
			}},
		},
		{
			Code: `
enum Num {
  A = 1,
}
declare const num: Num;
((1)) === num;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mismatchedCondition",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceValueWithEnum",
					Output: `
enum Num {
  A = 1,
}
declare const num: Num;
((Num.A)) === num;
`,
				}},
			}},
		},
		{
			Code: `
enum ComputedKey {
  ['test-key' /* with comment */] = 1,
}
declare const computedKey: ComputedKey;
computedKey === 1;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mismatchedCondition",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceValueWithEnum",
					Output: `
enum ComputedKey {
  ['test-key' /* with comment */] = 1,
}
declare const computedKey: ComputedKey;
computedKey === ComputedKey['test-key'];
`,
				}},
			}},
		},
		{
			Code: `
enum ComputedKey {
  [` + "`" + `test-key` + "`" + ` /* with comment */] = 1,
}
declare const computedKey: ComputedKey;
computedKey === 1;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mismatchedCondition",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceValueWithEnum",
					Output: `
enum ComputedKey {
  [` + "`" + `test-key` + "`" + ` /* with comment */] = 1,
}
declare const computedKey: ComputedKey;
computedKey === ComputedKey[` + "`" + `test-key` + "`" + `];
`,
				}},
			}},
		},
		{
			Code: `
enum ComputedKey {
  [` + "`" + `test-
  key` + "`" + ` /* with comment */] = 1,
}
declare const computedKey: ComputedKey;
computedKey === 1;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mismatchedCondition",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceValueWithEnum",
					Output: `
enum ComputedKey {
  [` + "`" + `test-
  key` + "`" + ` /* with comment */] = 1,
}
declare const computedKey: ComputedKey;
computedKey === ComputedKey[` + "`" + `test-
  key` + "`" + `];
`,
				}},
			}},
		},
	})
}
