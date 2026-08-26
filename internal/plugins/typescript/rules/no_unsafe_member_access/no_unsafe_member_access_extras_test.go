package no_unsafe_member_access

// cspell:ignore splitted

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnsafeMemberAccessExtras locks in branches and edge shapes that the
// upstream test suite does not exercise. The 1:1 migrated cases live in
// no_unsafe_member_access_upstream_test.go.
func TestNoUnsafeMemberAccessExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Real-user: typescript-eslint#3292 ----
		// Nested namespace heritage accesses are excluded at every depth.
		{Code: `
declare namespace XXX {
  namespace WWW {
    interface Temp2 { foo: 'bar' }
  }
}
interface Test2 extends XXX.WWW.Temp2 {}
class Test3 implements XXX.WWW.Temp2 { foo = 'bar' as const }
    `},
		// ---- Real-user: typescript-eslint#2728 ----
		{
			Code:    `declare const event: { detail: any }; const uuid: unknown = event.detail?.uuid;`,
			Options: map[string]any{"allowOptionalChaining": true},
		},

		// Parentheses and non-null assertions preserve an ordinary safe receiver.
		{Code: `
declare const value: { path: string };
((value)).path;
value!.path;
    `},
		// ESTree Literal and UpdateExpression fast paths cover the full literal
		// family even when the updated operand itself is any.
		{Code: `
declare const value: { [key: string]: unknown };
declare let anyValue: any;
value[0];
value['key'];
value['template'];
value[true];
value[false];
value[null];
value[anyValue++];
value[++anyValue];
    `},
		// Spread/rest, empty bodies, and declaration-only shapes are containers,
		// not member-access receiver targets.
		{Code: `
declare const safe: { value: number };
const object = { ...safe, value: safe.value };
const { value, ...rest } = object;
class Empty {}
function empty() {}
    `},
		// No options and an empty options object both retain the upstream default.
		{Code: `declare const safe: { value: number }; safe?.value;`},
		{Code: `declare const safe: { value: number }; safe?.value;`, Options: map[string]any{}},
		// Optional computed access is entirely skipped when allowed, including an
		// unresolved computed key.
		{
			Code:    `declare const safe: { value: number }; declare const key: NotKnown; safe?.[key];`,
			Options: map[string]any{"allowOptionalChaining": true},
		},
	}

	invalid := []rule_tester.InvalidTestCase{
		// Locks in exact text and selection for all six upstream message IDs.
		{
			Code: `declare const value: any;
value.property;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeMemberExpression",
				Message:   "Unsafe member access .property on an `any` value.",
				Line:      2,
				Column:    7,
				EndLine:   2,
				EndColumn: 15,
			}},
		},
		{
			Code: `declare const value: NotKnown;
value.property;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "errorMemberExpression",
				Message:   "Unsafe member access .property on a type that cannot be resolved.",
				Line:      2,
				Column:    7,
				EndLine:   2,
				EndColumn: 15,
			}},
		},
		{
			Code: `declare const safe: { value: number };
declare const key: any;
safe[key];`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeComputedMemberAccess",
				Message:   "Computed name [key] resolves to an `any` value.",
				Line:      3,
				Column:    6,
				EndLine:   3,
				EndColumn: 9,
			}},
		},
		{
			Code: `declare const safe: { value: number };
declare const key: NotKnown;
safe[key];`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "errorComputedMemberAccess",
				Message:   "The type of computed name [key] cannot be resolved.",
				Line:      3,
				Column:    6,
				EndLine:   3,
				EndColumn: 9,
			}},
		},
		{
			Code: `
const value = {
  read() {
    return this.property;
  },
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeThisMemberExpression",
				Message:   "Unsafe member access .property on an `any` value. `this` is typed as `any`.\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
				Line:      4,
				Column:    17,
				EndLine:   4,
				EndColumn: 25,
			}},
		},
		{
			Code: `function read(this: NotKnown) { return this.property; }`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "errorThisMemberExpression",
				Message:   "Unsafe member access .property. The type of `this` cannot be resolved.\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
				Line:      1,
				Column:    45,
				EndLine:   1,
				EndColumn: 53,
			}},
		},

		// Parentheses are transparent for ordinary member chains.
		{
			Code: `declare const value: any;
((value)).outer;
(value.outer).inner;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMemberExpression", Line: 2, Column: 11, EndLine: 2, EndColumn: 16},
				{MessageId: "unsafeMemberExpression", Line: 3, Column: 8, EndLine: 3, EndColumn: 13},
			},
		},
		// TypeScript wrappers preserve the checker-provided any type.
		{
			Code: `declare const value: unknown;
(value as any).property;
(value satisfies unknown as any).other;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMemberExpression", Line: 2, Column: 16, EndLine: 2, EndColumn: 24},
				{MessageId: "unsafeMemberExpression", Line: 3, Column: 34, EndLine: 3, EndColumn: 39},
			},
		},
		// An unsafe inner access reports once and suppresses the rest of the same
		// ordinary member chain.
		{
			Code: `declare const value: any; value.a.b.c;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeMemberExpression", Line: 1, Column: 33, EndLine: 1, EndColumn: 34,
			}},
		},
		// Calls end member-chain suppression, so the unsafe callee and unsafe call
		// result both report in source order.
		{
			Code: `declare const value: any;
value.method().result;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMemberExpression", Line: 2, Column: 7, EndLine: 2, EndColumn: 13},
				{MessageId: "unsafeMemberExpression", Line: 2, Column: 16, EndLine: 2, EndColumn: 22},
			},
		},
		// Explicit false retains the default behavior for optional access.
		{
			Code:    `declare const value: any; value?.property;`,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeMemberExpression", Line: 1, Column: 34, EndLine: 1, EndColumn: 42,
			}},
		},
		// A parenthesized optional chain is an ESTree ChainExpression boundary:
		// the optional link is allowed, but the following ordinary access is not.
		{
			Code: `declare const value: any;
(value?.property).other;`,
			Options: map[string]any{"allowOptionalChaining": true},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeMemberExpression", Line: 2, Column: 19, EndLine: 2, EndColumn: 24,
			}},
		},
		// Receiver and computed-key diagnostics are independent and share the
		// ESTree property range.
		{
			Code: `declare const value: any;
declare const key: any;
value[key];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMemberExpression", Line: 3, Column: 7, EndLine: 3, EndColumn: 10},
				{MessageId: "unsafeComputedMemberAccess", Line: 3, Column: 7, EndLine: 3, EndColumn: 10},
			},
		},
		// Class extends is a runtime expression; only class implements and
		// interface heritage are excluded by the upstream selector.
		{
			Code: `declare const unsafeBase: any;
class UnsafeDerived extends unsafeBase.Base {}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeMemberExpression", Line: 2, Column: 40, EndLine: 2, EndColumn: 44,
			}},
		},
		// The audit document's rspack shape is aligned under the repository's
		// TypeScript 6 baseline: omitted strictness defaults to enabled, so the
		// logical-and initializer and all three accesses remain any-typed.
		{
			Code: `
interface ParsedResource {
  path: string;
  query: string;
  fragment: string;
}
declare function parseResource(value: string): ParsedResource;
declare const value: any;
const splittedResource = value && parseResource(value);
splittedResource.path;
splittedResource.query;
splittedResource.fragment;
      `,
			TSConfig: "tsconfig.no-explicit-strict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMemberExpression", Line: 10, Column: 18, EndLine: 10, EndColumn: 22},
				{MessageId: "unsafeMemberExpression", Line: 11, Column: 18, EndLine: 11, EndColumn: 23},
				{MessageId: "unsafeMemberExpression", Line: 12, Column: 18, EndLine: 12, EndColumn: 26},
			},
		},
		// typescript-eslint#7037: an in check does not make any safe.
		{
			Code: `declare const anyValue: any;
if ('value' in anyValue) {
  anyValue.value;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeMemberExpression", Line: 3, Column: 12, EndLine: 3, EndColumn: 17,
			}},
		},
		// With noImplicitThis enabled, upstream deliberately uses the generic
		// member message for an explicit this: any receiver.
		{
			Code:     `function read(this: any) { return this.property; }`,
			TSConfig: "tsconfig.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeMemberExpression", Line: 1, Column: 40, EndLine: 1, EndColumn: 48,
			}},
		},
		// TypeScript 6 also enables noImplicitThis when strict is omitted.
		{
			Code:     `function read(this: any) { return this.property; }`,
			TSConfig: "tsconfig.no-explicit-strict.json",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeMemberExpression", Line: 1, Column: 40, EndLine: 1, EndColumn: 48,
			}},
		},
		// Disable directives suppress only their covered access.
		{
			Code: `declare const value: any;
// eslint-disable-next-line test
value.first;
value.second;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unsafeMemberExpression", Line: 4, Column: 7, EndLine: 4, EndColumn: 13,
			}},
		},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.noImplicitThis.json",
		t,
		&NoUnsafeMemberAccessRule,
		valid,
		invalid,
	)
}
