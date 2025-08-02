package explicit_member_accessibility

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestExplicitMemberAccessibilityRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		{
			Code: `
class Test {
  protected name: string;
  private x: number;
  public getX() {
    return this.x;
  }
}`,
			Options: map[string]interface{}{"accessibility": "explicit"},
		},
		{
			Code: `
class Test {
  name: string;
  foo?: string;
  getX() {
    return this.x;
  }
}`,
			Options: map[string]interface{}{"accessibility": "no-public"},
		},
		{
			Code: `
class Test {
  public constructor(private foo: string) {}
}`,
			Options: map[string]interface{}{
				"accessibility": "explicit",
				"overrides": map[string]interface{}{
					"parameterProperties": "explicit",
				},
			},
		},
		{
			Code: `
class Test {
  public getX() {
    return this.x;
  }
}`,
			Options: map[string]interface{}{
				"ignoredMethodNames": []interface{}{"getX"},
			},
		},
		{
			Code: `
class Test {
  #foo = 1;
  #bar() {}
}`,
			Options: map[string]interface{}{"accessibility": "explicit"},
		},
		{
			Code: `
class Test {
  private accessor foo = 1;
}`,
			Options: map[string]interface{}{"accessibility": "explicit"},
		},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		{
			Code: `
class Test {
  x: number;
  public getX() {
    return this.x;
  }
}`,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Test {
  private x: number;
  getX() {
    return this.x;
  }
}`,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Test {
  protected name: string;
  public foo?: string;
  getX() {
    return this.x;
  }
}`,
			Options: map[string]interface{}{"accessibility": "no-public"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unwantedPublicAccessibility",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `
export class WithParameterProperty {
  public constructor(readonly value: string) {}
}`,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Line:      3,
					Column:    22,
				},
			},
		},
		{
			Code: `
abstract class SomeClass {
  abstract method(): string;
}`,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
class SomeClass {
  accessor foo = 1;
}`,
			Options: map[string]interface{}{"accessibility": "explicit"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAccessibility",
					Line:      3,
					Column:    3,
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitMemberAccessibilityRule, validTestCases, invalidTestCases)
}