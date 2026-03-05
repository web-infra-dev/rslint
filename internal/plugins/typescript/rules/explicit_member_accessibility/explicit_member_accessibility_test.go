package explicit_member_accessibility

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExplicitMemberAccessibilityRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitMemberAccessibilityRule, []rule_tester.ValidTestCase{
		{Code: "class Test {\n  public getX() {}\n}\n"},
		{
			Code:    "class Test {\n  constructor(public foo: number) {}\n}\n",
			Options: []interface{}{map[string]interface{}{"accessibility": "no-public"}},
		},
		{
			Code:    "class Test {\n  getX() {}\n}\n",
			Options: []interface{}{map[string]interface{}{"ignoredMethodNames": []interface{}{"getX"}}},
		},
		{Code: "class Test {\n  #foo = 1;\n}\n"},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "class Test {\n  x: number;\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  public x: number;\n}\n"},
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  private x: number;\n}\n"},
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  protected x: number;\n}\n"},
					},
				},
			},
		},
		{
			Code: "class Test {\n  getX() {\n    return this.x;\n  }\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  public getX() {\n    return this.x;\n  }\n}\n"},
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  private getX() {\n    return this.x;\n  }\n}\n"},
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  protected getX() {\n    return this.x;\n  }\n}\n"},
					},
				},
			},
		},
		{
			Code: "class Test {\n  constructor(public x: number) {}\n}\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 14,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  public constructor(public x: number) {}\n}\n"},
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  private constructor(public x: number) {}\n}\n"},
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  protected constructor(public x: number) {}\n}\n"},
					},
				},
			},
		},
		{
			Code:    "class Test {\n  public x = 2;\n}\n",
			Options: []interface{}{map[string]interface{}{"accessibility": "no-public"}},
			Output:  []string{"class Test {\n  x = 2;\n}\n"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unwantedPublicAccessibility",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 9,
				},
			},
		},
		{
			Code:    "class Test {\n  constructor(public readonly foo: string) {}\n}\n",
			Options: []interface{}{map[string]interface{}{"accessibility": "no-public"}},
			Output:  []string{"class Test {\n  constructor(readonly foo: string) {}\n}\n"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unwantedPublicAccessibility",
					Line:      2,
					Column:    15,
					EndLine:   2,
					EndColumn: 21,
				},
			},
		},
		{
			Code:    "class Test {\n  constructor(readonly foo: string) {}\n}\n",
			Options: []interface{}{map[string]interface{}{"accessibility": "off", "overrides": map[string]interface{}{"parameterProperties": "explicit"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Line:      2,
					Column:    15,
					EndLine:   2,
					EndColumn: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  constructor(public readonly foo: string) {}\n}\n"},
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  constructor(private readonly foo: string) {}\n}\n"},
						{MessageId: "addExplicitAccessibility", Output: "class Test {\n  constructor(protected readonly foo: string) {}\n}\n"},
					},
				},
			},
		},
	})
}
