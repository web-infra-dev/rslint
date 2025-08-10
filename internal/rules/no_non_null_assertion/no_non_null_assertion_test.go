package no_non_null_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoNonNullAssertionRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// 基本有效情况 - 没有非空断言
		{Code: `const foo = "hello"; console.log(foo);`},
		{Code: `function foo(bar: string) { console.log(bar); }`},
		{Code: `const foo: string | null = "hello"; if (foo) { console.log(foo); }`},
		{Code: `const foo: string | undefined = "hello"; if (foo !== undefined) { console.log(foo); }`},
		{Code: `const foo: string | null = "hello"; const bar = foo || "default";`},
		{Code: `const foo: string | null = "hello"; const bar = foo ?? "default";`},
		{Code: `const foo: string | null = "hello"; const bar = foo?.length || 0;`},
		{Code: `const foo: string | null = "hello"; if (foo !== null) { console.log(foo); }`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// 基本非空断言 - 应该报告错误
		{
			Code: `const foo: string | null = "hello"; const bar = foo!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 属性访问中的非空断言
		{
			Code: `const foo: string | null = "hello"; const bar = foo!.length;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 函数调用中的非空断言
		{
			Code: `const foo: string | null = "hello"; const bar = foo!.toUpperCase();`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 数组访问中的非空断言
		{
			Code: `const foo: string[] | null = ["hello"]; const bar = foo![0];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 链式非空断言 - 应该报告2个错误
		{
			Code: `const foo: string | null = "hello"; const bar = foo!!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 条件表达式中的非空断言
		{
			Code: `const foo: string | null = "hello"; const bar = foo! ? "yes" : "no";`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 逻辑表达式中的非空断言
		{
			Code: `const foo: string | null = "hello"; const bar = foo! && "yes";`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 返回语句中的非空断言
		{
			Code: `function test(): string { const foo: string | null = "hello"; return foo!; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 变量声明中的非空断言
		{
			Code: `let foo: string | null = "hello"; foo = foo!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 参数中的非空断言
		{
			Code: `function test(foo: string | null) { const bar = foo!; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 对象属性中的非空断言
		{
			Code: `const obj = { foo: "hello" as string | null }; const bar = obj.foo!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 模板字符串中的非空断言
		{
			Code: "const foo: string | null = \"hello\"; const bar = `Value: ${foo!}`;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 类型断言中的非空断言
		{
			Code: `const foo: string | null = "hello"; const bar = (foo! as string).length;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 泛型中的非空断言
		{
			Code: `function test<T extends string | null>(foo: T): T { return foo!; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 联合类型中的非空断言
		{
			Code: `const foo: (string | null)[] = ["hello"]; const bar = foo[0]!;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 嵌套表达式中的非空断言
		{
			Code: `const foo: string | null = "hello"; const bar = (foo! + "world").length;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
			},
		},
		// 三元表达式中的非空断言
		{
			Code: `const foo: string | null = "hello"; const bar = foo! ? foo!.length : 0;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noNonNull",
				},
				{
					MessageId: "noNonNull",
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNonNullAssertionRule, validTestCases, invalidTestCases)
}

func TestNoNonNullAssertionOptionsParsing(t *testing.T) {
	// 测试规则基本信息
	rule := NoNonNullAssertionRule
	if rule.Name != "no-non-null-assertion" {
		t.Errorf("Expected rule name to be 'no-non-null-assertion', got %s", rule.Name)
	}
}

func TestNoNonNullAssertionMessage(t *testing.T) {
	msg := buildNoNonNullAssertionMessage()
	if msg.Id != "noNonNull" {
		t.Errorf("Expected message ID to be 'noNonNull', got %s", msg.Id)
	}
	if msg.Description != "Non-null assertion operator (!) is not allowed." {
		t.Errorf("Expected description to be 'Non-null assertion operator (!) is not allowed.', got %s", msg.Description)
	}
}

// 测试边界情况
func TestNoNonNullAssertionEdgeCases(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// 嵌套赋值表达式
		{Code: `let obj: { prop?: string } = {}; obj.prop! = "value";`},

		// 解构赋值中的非空断言
		{Code: `let arr: (string | null)[] = ["hello"]; [arr[0]!] = ["world"];`},

		// 复杂赋值表达式
		{Code: `let foo: string | null = "hello"; (foo! as any) = "world";`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// 这些测试用例已经在主测试函数中包含了
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNonNullAssertionRule, validTestCases, invalidTestCases)
}
