package no_unsafe_member_access

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeMemberAccessRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noImplicitThis.json", t, &NoUnsafeMemberAccessRule, []rule_tester.ValidTestCase{
		{Code: `
function foo(x: { a: number }, y: any) {
  x[y++];
}
    `},
		{Code: `
function foo(x: { a: number }) {
  x.a;
}
    `},
		{Code: `
function foo(x?: { a: number }) {
  x?.a;
}
    `},
		{Code: `
function foo(x: { a: number }) {
  x['a'];
}
    `},
		{Code: `
function foo(x?: { a: number }) {
  x?.['a'];
}
    `},
		{Code: `
function foo(x: { a: number }, y: string) {
  x[y];
}
    `},
		{Code: `
function foo(x?: { a: number }, y: string) {
  x?.[y];
}
    `},
		{Code: `
function foo(x: string[]) {
  x[1];
}
    `},
		{Code: `
class B implements FG.A {}
    `},
		{Code: `
interface B extends FG.A {}
    `},
		{Code: `
class B implements F.S.T.A {}
    `},
		{Code: `
interface B extends F.S.T.A {}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
function foo(x: any) {
  x.a;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeMemberExpression",
					Line:      3,
					Column:    5,
					EndColumn: 6,
				},
			},
		},
		{
			Code: `
function foo(x: any) {
  x.a.b.c.d.e.f.g;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeMemberExpression",
					Line:      3,
					Column:    5,
					EndColumn: 6,
				},
			},
		},
		{
			Code: `
function foo(x: { a: any }) {
  x.a.b.c.d.e.f.g;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeMemberExpression",
					Line:      3,
					Column:    7,
					EndColumn: 8,
				},
			},
		},
		{
			Code: `
function foo(x: any) {
  x['a'];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeMemberExpression",
					Line:      3,
					Column:    5,
					EndColumn: 8,
				},
			},
		},
		{
			Code: `
function foo(x: any) {
  x['a']['b']['c'];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeMemberExpression",
					Line:      3,
					Column:    5,
					EndColumn: 8,
				},
			},
		},
		{
			Code: `
let value: NotKnown;

value.property;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeMemberExpression",
					Line:      4,
					Column:    7,
					EndColumn: 15,
				},
			},
		},
		{
			Code: `
function foo(x: { a: number }, y: any) {
  x[y];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndColumn: 6,
				},
			},
		},
		{
			Code: `
function foo(x?: { a: number }, y: any) {
  x?.[y];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    7,
					EndColumn: 8,
				},
			},
		},
		{
			Code: `
function foo(x: { a: number }, y: any) {
  x[(y += 1)];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndColumn: 13,
				},
			},
		},
		{
			Code: `
function foo(x: { a: number }, y: any) {
  x[1 as any];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndColumn: 13,
				},
			},
		},
		{
			Code: `
function foo(x: { a: number }, y: any) {
  x[y()];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndColumn: 8,
				},
			},
		},
		{
			Code: `
function foo(x: string[], y: any) {
  x[y];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndColumn: 6,
				},
			},
		},
		{
			Code: `
function foo(x: { a: number }, y: NotKnown) {
  x[y];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndColumn: 6,
				},
			},
		},
		{
			Code: `
const methods = {
  methodA() {
    return this.methodB()
  },
  methodB() {
    const getProperty = () => Math.random() > 0.5 ? 'methodB' : 'methodC'
    return this[getProperty()]()
  },
  methodC() {
    return true
  },
  methodD() {
    return (this?.methodA)?.()
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeThisMemberExpression",
					Line:      4,
					Column:    17,
					EndColumn: 24,
				},
				{
					MessageId: "unsafeThisMemberExpression",
					Line:      8,
					Column:    17,
					EndColumn: 30,
				},
				{
					MessageId: "unsafeThisMemberExpression",
					Line:      14,
					Column:    19,
					EndColumn: 26,
				},
			},
		},
		{
			Code: `
class C {
  getObs$: any;
  getPopularDepartments(): void {
    this.getObs$.pipe().subscribe(res => {
      log(res);
    });
  }
}
function log(arg: unknown) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeMemberExpression",
					Line:      5,
					Column:    25,
					EndColumn: 34,
				},
				{
					MessageId: "unsafeMemberExpression",
					Line:      5,
					Column:    18,
					EndColumn: 22,
				},
			},
		},
	})
}
