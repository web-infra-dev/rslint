package no_use_before_define

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoUseBeforeDefineRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUseBeforeDefineRule, []rule_tester.ValidTestCase{
		{
			Code: `
				type foo = 1;
				const x: foo = 1;
			`,
		},
		{
			Code: `
				var a = 10;
				alert(a);
			`,
		},
		{
			Code: `
				function b(a: any) {
					alert(a);
				}
			`,
		},
		{
			Code: `
				a();
				function a() {
					alert(arguments);
				}
			`,
			Options: "nofunc",
		},
		{
			Code: `
				function foo() {
					foo();
				}
			`,
		},
		{
			Code: `
				var foo = function () {
					foo();
				};
			`,
		},
		{
			Code: `
				var a: any;
				for (a in a) {
				}
			`,
		},
		{
			Code: `
				var a: any;
				for (a of a) {
				}
			`,
		},
		{
			Code: `
				function foo() {
					new A();
				}
				class A {}
			`,
			Options: map[string]interface{}{"classes": false},
		},
		{
			Code: `
				function foo() {
					bar;
				}
				var bar: any;
			`,
			Options: map[string]interface{}{"variables": false},
		},
		{
			Code: `
				var x: Foo = 2;
				type Foo = string | number;
			`,
			Options: map[string]interface{}{"typedefs": false},
		},
		{
			Code: `
				interface Bar {
					type: typeof Foo;
				}

				const Foo = 2;
			`,
			Options: map[string]interface{}{"ignoreTypeReferences": true},
		},
		{
			Code: `
				function foo(): Foo {
					return Foo.FOO;
				}

				enum Foo {
					FOO,
				}
			`,
			Options: map[string]interface{}{"enums": false},
		},
		{
			Code: `
				export { a };
				const a = 1;
			`,
			Options: map[string]interface{}{"allowNamedExports": true},
		},
		{
			Code: `
				export { a as b };
				const a = 1;
			`,
			Options: map[string]interface{}{"allowNamedExports": true},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
				a++;
				var a = 19;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      2,
					Column:    5,
				},
			},
		},
		{
			Code: `
				a();
				var a = function () {};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      2,
					Column:    5,
				},
			},
		},
		{
			Code: `
				alert(a[1]);
				var a = [1, 3];
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      2,
					Column:    11,
				},
			},
		},
		{
			Code: `
				a();
				function a() {
					alert(b);
					var b = 10;
					a();
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      2,
					Column:    5,
				},
				{
					MessageId: "noUseBeforeDefine",
					Line:      4,
					Column:    12,
				},
			},
		},
		{
			Code: `
				a();
				var a = function () {};
			`,
			Options: "nofunc",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      2,
					Column:    5,
				},
			},
		},
		{
			Code: `
				(() => {
					alert(a);
					var a = 42;
				})();
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      3,
					Column:    12,
				},
			},
		},
		{
			Code: `
				var f = () => a;
				var a: any;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      2,
					Column:    19,
				},
			},
		},
		{
			Code: `
				new A();
				class A {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      2,
					Column:    9,
				},
			},
		},
		{
			Code: `
				function foo() {
					new A();
				}
				class A {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
				interface Bar {
					type: typeof Foo;
				}

				const Foo = 2;
			`,
			Options: map[string]interface{}{"ignoreTypeReferences": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      3,
					Column:    19,
				},
			},
		},
		{
			Code: `
				function foo(): Foo {
					return Foo.FOO;
				}

				enum Foo {
					FOO,
				}
			`,
			Options: map[string]interface{}{"enums": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      3,
					Column:    13,
				},
			},
		},
		{
			Code: `
				export { a };
				const a = 1;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      2,
					Column:    14,
				},
			},
		},
		{
			Code: `
				export { a };
				const a = 1;
			`,
			Options: map[string]interface{}{"allowNamedExports": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noUseBeforeDefine",
					Line:      2,
					Column:    14,
				},
			},
		},
	})
}
