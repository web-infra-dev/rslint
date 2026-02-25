package class_methods_use_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestClassMethodsUseThisRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ClassMethodsUseThisRule, []rule_tester.ValidTestCase{
		{
			Code: `
class Foo {
  method() {
    this.value = 1;
  }
}
      `,
		},
		{
			Code: `
class Foo {
  method() {
    return () => this.value;
  }
}
      `,
		},
		{
			Code: `
class Foo {
  property = () => {
    this.value = 1;
  };
}
      `,
		},
		{
			Code: `
class Foo {
  property = () => {};
}
      `,
			Options: map[string]interface{}{"enforceForClassFields": false},
		},
		{
			Code: `
class Foo implements Bar {
  method() {}
}
      `,
			Options: map[string]interface{}{"ignoreClassesThatImplementAnInterface": true},
		},
		{
			Code: `
class Foo {
  override method() {}
}
      `,
			Options: map[string]interface{}{"ignoreOverrideMethods": true},
		},
		{
			Code: `
class Foo {
  method() {}
}
      `,
			Options: map[string]interface{}{"exceptMethods": []interface{}{"method"}},
		},
		{
			Code: `
class Foo {
  method(value = this.value) {}
}
      `,
		},
		{
			Code: `
class Foo {
  property = (value = this.value) => {
    return value;
  };
}
      `,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
class Foo {
  method() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingThis"}},
		},
		{
			Code: `
class Foo {
  get value() {
    return 1;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingThis"}},
		},
		{
			Code: `
class Foo {
  set value(next: number) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingThis"}},
		},
		{
			Code: `
class Foo {
  property = () => {};
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingThis"}},
		},
		{
			Code: `
class Foo {
  property = function () {};
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingThis"}},
		},
		{
			Code: `
class Foo implements Bar {
  method() {}
}
      `,
			Options: map[string]interface{}{"ignoreClassesThatImplementAnInterface": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingThis"}},
		},
		{
			Code: `
class Foo implements Bar {
  private method() {}
}
      `,
			Options: map[string]interface{}{"ignoreClassesThatImplementAnInterface": "public-fields"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingThis"}},
		},
		{
			Code: `
class Foo {
  override method() {}
}
      `,
			Options: map[string]interface{}{"ignoreOverrideMethods": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingThis"}},
		},
		{
			Code: `
class Foo {
  property = value => value;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "missingThis",
				Line:      3,
				Column:    3,
				EndLine:   3,
				EndColumn: 14,
			}},
		},
	})
}
