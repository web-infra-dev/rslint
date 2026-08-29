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
			// Sequence expressions fold to their final value, while preserving
			// their outer parentheses in the replacement.
			Code: `
enum Num {
  A = 2,
}
declare const num: Num;
num === (1, 2);
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mismatchedCondition",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceValueWithEnum",
					Output: `
enum Num {
  A = 2,
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
  A = 2,
}
declare const num: Num;
num === 'ab'.length;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mismatchedCondition",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceValueWithEnum",
					Output: `
enum Num {
  A = 2,
}
declare const num: Num;
num === Num.A;
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
num === 'ab'.indexOf('b');
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
num === Num.A;
`,
				}},
			}},
		},
		{
			// NaN is not equal to itself, so a textual NaN match must not offer
			// an enum replacement suggestion.
			Code: `
enum Num {
  A = 0 / 0,
}
declare const num: Num;
num === 0 / 0;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mismatchedCondition",
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
