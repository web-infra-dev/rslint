package no_unsafe_member_access

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnsafeMemberAccessUpstream migrates the full valid/invalid suite from
// packages/eslint-plugin/tests/rules/no-unsafe-member-access.test.ts at
// typescript-eslint v8.67.0 1:1. rslint-specific lock-ins live in
// no_unsafe_member_access_extras_test.go.
func TestNoUnsafeMemberAccessUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.noImplicitThis.json",
		t,
		&NoUnsafeMemberAccessRule,
		[]rule_tester.ValidTestCase{
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
			{
				Code: `
function foo(x?: { a: number }) {
  x?.a;
}
      `,
				Options: map[string]any{"allowOptionalChaining": true},
			},
			{
				Code: `
function foo(x?: { a: number }, y: string) {
  x?.[y];
}
      `,
				Options: map[string]any{"allowOptionalChaining": true},
			},
			{
				Code: `
function foo(x: { a: number }, y: 'a') {
  x?.[y];
}
      `,
				Options: map[string]any{"allowOptionalChaining": true},
			},
			{
				Code: `
function foo(x: { a: number }, y: NotKnown) {
  x?.[y];
}
      `,
				Options: map[string]any{"allowOptionalChaining": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
function foo(x: any) {
  x.a;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeMemberExpression",
					Line:      3,
					Column:    5,
					EndLine:   3,
					EndColumn: 6,
				}},
			},
			{
				Code: `
function foo(x: any) {
  x.a.b.c.d.e.f.g;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeMemberExpression",
					Line:      3,
					Column:    5,
					EndLine:   3,
					EndColumn: 6,
				}},
			},
			{
				Code: `
function foo(x: { a: any }) {
  x.a.b.c.d.e.f.g;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeMemberExpression",
					Line:      3,
					Column:    7,
					EndLine:   3,
					EndColumn: 8,
				}},
			},
			{
				Code: `
function foo(x: any) {
  x['a'];
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeMemberExpression",
					Line:      3,
					Column:    5,
					EndLine:   3,
					EndColumn: 8,
				}},
			},
			{
				Code: `
function foo(x: any) {
  x['a']['b']['c'];
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeMemberExpression",
					Line:      3,
					Column:    5,
					EndLine:   3,
					EndColumn: 8,
				}},
			},
			{
				Code: `
let value: NotKnown;

value.property;
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "errorMemberExpression",
					Line:      4,
					Column:    7,
					EndLine:   4,
					EndColumn: 15,
				}},
			},
			{
				Code: `
function foo(x: { a: number }, y: any) {
  x[y];
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndLine:   3,
					EndColumn: 6,
				}},
			},
			{
				Code: `
function foo(x?: { a: number }, y: any) {
  x?.[y];
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    7,
					EndLine:   3,
					EndColumn: 8,
				}},
			},
			{
				Code: `
function foo(x: { a: number }, y: any) {
  x[(y += 1)];
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    6,
					EndLine:   3,
					EndColumn: 12,
				}},
			},
			{
				Code: `
function foo(x: { a: number }, y: any) {
  x[1 as any];
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndLine:   3,
					EndColumn: 13,
				}},
			},
			{
				Code: `
function foo(x: { a: number }, y: any) {
  x[y()];
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndLine:   3,
					EndColumn: 8,
				}},
			},
			{
				Code: `
function foo(x: string[], y: any) {
  x[y];
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndLine:   3,
					EndColumn: 6,
				}},
			},
			{
				Code: `
function foo(x: { a: number }, y: NotKnown) {
  x[y];
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "errorComputedMemberAccess",
					Line:      3,
					Column:    5,
					EndLine:   3,
					EndColumn: 6,
				}},
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
					{MessageId: "unsafeThisMemberExpression", Line: 4, Column: 17, EndLine: 4, EndColumn: 24},
					{MessageId: "unsafeThisMemberExpression", Line: 8, Column: 17, EndLine: 8, EndColumn: 30},
					{MessageId: "unsafeThisMemberExpression", Line: 14, Column: 19, EndLine: 14, EndColumn: 26},
				},
			},
			{
				Code: `
class C {
  getObs$: any;
  getPopularDepartments(): void {
    this.getObs$.pipe().subscribe(res => {
      console.log(res);
    });
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeMemberExpression", Line: 5, Column: 18, EndLine: 5, EndColumn: 22},
					{MessageId: "unsafeMemberExpression", Line: 5, Column: 25, EndLine: 5, EndColumn: 34},
				},
			},
			{
				Code: `
let value: any;

value?.middle.inner;
      `,
				Options: map[string]any{"allowOptionalChaining": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeMemberExpression", Line: 4, Column: 15, EndLine: 4, EndColumn: 20,
				}},
			},
			{
				Code: `
let value: any;

value?.outer.middle.inner;
      `,
				Options: map[string]any{"allowOptionalChaining": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeMemberExpression", Line: 4, Column: 14, EndLine: 4, EndColumn: 20,
				}},
			},
			{
				Code: `
let value: any;

value.outer?.middle.inner;
      `,
				Options: map[string]any{"allowOptionalChaining": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeMemberExpression", Line: 4, Column: 7, EndLine: 4, EndColumn: 12},
					{MessageId: "unsafeMemberExpression", Line: 4, Column: 21, EndLine: 4, EndColumn: 26},
				},
			},
			{
				Code: `
let value: any;

value.outer.middle?.inner;
      `,
				Options: map[string]any{"allowOptionalChaining": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unsafeMemberExpression", Line: 4, Column: 7, EndLine: 4, EndColumn: 12,
				}},
			},
			{
				Code: `
function foo(x: { a: number }, y: NotKnown) {
  x[y];
}
      `,
				Options: map[string]any{"allowOptionalChaining": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "errorComputedMemberAccess", Line: 3, Column: 5, EndLine: 3, EndColumn: 6,
				}},
			},
		},
	)
}
