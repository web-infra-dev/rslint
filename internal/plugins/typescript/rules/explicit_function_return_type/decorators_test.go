package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestDecoratorHeadLoc pins down the report location when a class member has
// decorators. typescript-eslint's getFunctionHeadLoc excludes leading
// decorators from the reported range, so the column should land on the first
// token after the last decorator (modifier, `*`, method name, etc.).
func TestDecoratorHeadLoc(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// Decorator on a plain method — column should be on `use`, not on `@`.
		{
			Code: `
declare function Middleware(): any;
@Middleware()
class HookMiddleware {
  async use() {
    return;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 5, Column: 3, EndLine: 5, EndColumn: 12},
			},
		},
		// Decorator on the method itself — column should be on `async`, after @LifecycleHook().
		{
			Code: `
declare function LifecycleHook(): any;
class HookLifeCycle {
  @LifecycleHook()
  async didLoad() {
    return;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 5, Column: 3, EndLine: 5, EndColumn: 16},
			},
		},
		// Multiple decorators — column should be on `method`, after all decorators.
		{
			Code: `
declare function A(): any;
declare function B(): any;
class Foo {
  @A() @B()
  method() {
    return 1;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 6, Column: 3, EndLine: 6, EndColumn: 9},
			},
		},
		// Decorator on a class property with arrow initializer.
		// Column should be on `foo`, after the decorator.
		{
			Code: `
declare function Bound(): any;
class Foo {
  @Bound()
  foo = () => 1;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 5, Column: 3, EndLine: 5, EndColumn: 9},
			},
		},
	})
}
