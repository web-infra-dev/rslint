package no_unused_vars

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils"
	"reflect"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsIgnorePatternMessages(t *testing.T) {
	options := map[string]interface{}{
		"args":                           "all",
		"argsIgnorePattern":              "^_",
		"caughtErrors":                   "all",
		"caughtErrorsIgnorePattern":      "^_",
		"destructuredArrayIgnorePattern": "^_",
		"varsIgnorePattern":              "^_",
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code:    `const value = 1;`,
				Options: options,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Message:   "'value' is assigned a value but never used. Allowed unused vars must match /^_/u.",
					Line:      1,
					Column:    7,
				}},
			},
			{
				Code:    `const value = 1;`,
				Options: map[string]interface{}{"varsIgnorePattern": "a/b"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Message:   `'value' is assigned a value but never used. Allowed unused vars must match /a\/b/u.`,
					Line:      1,
					Column:    7,
				}},
			},
			{
				Code:    `const value = 1;`,
				Options: map[string]interface{}{"varsIgnorePattern": `a\/b`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Message:   `'value' is assigned a value but never used. Allowed unused vars must match /a\/b/u.`,
					Line:      1,
					Column:    7,
				}},
			},
			{
				Code:    `const value = 1;`,
				Options: map[string]interface{}{"varsIgnorePattern": "a\nb"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Message:   `'value' is assigned a value but never used. Allowed unused vars must match /a\nb/u.`,
					Line:      1,
					Column:    7,
				}},
			},
			{
				Code:    `export function run(value: number) {}`,
				Options: options,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Message:   "'value' is defined but never used. Allowed unused args must match /^_/u.",
					Line:      1,
					Column:    21,
				}},
			},
			{
				Code:    `try {} catch (error) {}`,
				Options: options,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Message:   "'error' is defined but never used. Allowed unused caught errors must match /^_/u.",
					Line:      1,
					Column:    15,
				}},
			},
			{
				Code: `declare const source: { value?: number, nested?: { value?: number } };
export function objectParameter({ value }: { value?: number }) {}
export function defaultedObjectParameter({ value = 1 }: { value?: number }) {}
export function arrayParameter([value]: number[]) {}
export function defaultedArrayParameter([value = 1]: number[]) {}
export function defaultedParameter(value = 1) {}
export function defaultedPattern({ nested: { value } } = { nested: {} }) {}
export function nestedDefault({ nested: { value = 1 } }) {}
try { throw source; } catch ({ value }) {}
try { throw source; } catch ({ value = 1 }) {}`,
				Options: options,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used. Allowed unused args must match /^_/u.", Line: 2, Column: 35},
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used. Allowed unused args must match /^_/u.", Line: 3, Column: 44},
					{MessageId: "unusedVar", Message: "'value' is defined but never used. Allowed unused elements of array destructuring must match /^_/u.", Line: 4, Column: 33},
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used. Allowed unused args must match /^_/u.", Line: 5, Column: 42},
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used. Allowed unused args must match /^_/u.", Line: 6, Column: 36},
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used. Allowed unused args must match /^_/u.", Line: 7, Column: 46},
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used. Allowed unused args must match /^_/u.", Line: 8, Column: 43},
					{MessageId: "unusedVar", Message: "'value' is defined but never used. Allowed unused caught errors must match /^_/u.", Line: 9, Column: 32},
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used. Allowed unused caught errors must match /^_/u.", Line: 10, Column: 32},
				},
			},
			{
				Code:    `const [value] = [1];`,
				Options: options,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Message:   "'value' is assigned a value but never used. Allowed unused elements of array destructuring must match /^_/u.",
					Line:      1,
					Column:    8,
				}},
			},
			{
				Code:    `const [value = 1] = [];`,
				Options: options,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Message:   "'value' is assigned a value but never used. Allowed unused vars must match /^_/u.",
					Line:      1,
					Column:    8,
				}},
			},
			{
				Code:    `const value = 1; export type Value = typeof value;`,
				Options: options,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "usedOnlyAsType",
					Message:   "'value' is assigned a value but only used as a type. Allowed unused vars must match /^_/u.",
					Line:      1,
					Column:    7,
				}},
			},
			{
				Code:    `const _value = 1; console.log(_value);`,
				Options: map[string]interface{}{"varsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "usedIgnoredVar",
					Message:   "'_value' is marked as ignored but is used. Used vars must not match /^_/u.",
					Line:      1,
					Column:    7,
				}},
			},
			{
				Code:    `const [_value] = [1]; console.log(_value);`,
				Options: map[string]interface{}{"destructuredArrayIgnorePattern": "^_", "varsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "usedIgnoredVar",
					Message:   "'_value' is marked as ignored but is used. Used elements of array destructuring must not match /^_/u.",
					Line:      1,
					Column:    8,
				}},
			},
		},
	)
}

func TestNoUnusedVarsMessageData(t *testing.T) {
	tests := []struct {
		name    string
		message rule.RuleMessage
		want    map[string]string
	}{
		{
			name:    "unused",
			message: buildUnusedVarMessage("value", true, ". Allowed unused vars must match /^_/u"),
			want: map[string]string{
				"varName":    "value",
				"action":     "assigned a value",
				"additional": ". Allowed unused vars must match /^_/u",
			},
		},
		{
			name:    "used only as type",
			message: buildUsedOnlyAsTypeMessage("value", false, ""),
			want: map[string]string{
				"varName":    "value",
				"action":     "defined",
				"additional": "",
			},
		},
		{
			name:    "used ignored",
			message: buildUsedIgnoredVarMessage("_value", ". Used vars must not match /^_/u"),
			want: map[string]string{
				"varName":    "_value",
				"additional": ". Used vars must not match /^_/u",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.message.Data, test.want) {
				t.Fatalf("message data = %#v, want %#v", test.message.Data, test.want)
			}
		})
	}
}

func TestNoUnusedVarsIgnorePatternMessageEscaping(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "slash", source: "a/b", want: `. Allowed unused vars must match /a\/b/u`},
		{name: "escaped slash", source: `a\/b`, want: `. Allowed unused vars must match /a\/b/u`},
		{name: "even backslashes before slash", source: `a\\/b`, want: `. Allowed unused vars must match /a\\\/b/u`},
		{name: "line feed", source: "a\nb", want: `. Allowed unused vars must match /a\nb/u`},
		{name: "carriage return", source: "a\rb", want: `. Allowed unused vars must match /a\rb/u`},
		{name: "line separator", source: "a\u2028b", want: `. Allowed unused vars must match /a\u2028b/u`},
		{name: "paragraph separator", source: "a\u2029b", want: `. Allowed unused vars must match /a\u2029b/u`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ignorePatternMessage("vars", test.source, false); got != test.want {
				t.Fatalf("message = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNoUnusedVarsFunctionTypeSignatureParameterReferences(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `const functionTypeParameter = 'key';
export type FunctionType = (value: typeof functionTypeParameter) => void;

const constructorTypeParameter = 'key';
export type ConstructorType = new (value: typeof constructorTypeParameter) => object;

const callSignatureParameter = 'key';
export interface CallSignature { (value: typeof callSignatureParameter): void; }

const constructSignatureParameter = 'key';
export interface ConstructSignature { new (value: typeof constructSignatureParameter): object; }

const methodSignatureParameter = 'key';
export interface MethodSignature { method(value: typeof methodSignatureParameter): void; }

const declaredFunctionParameter = 'key';
declare function declaredFunction(value: typeof declaredFunctionParameter): void;
export { declaredFunction };`,
			},
			{
				Code: `const MOBILE_AUTH_FACE_NOT_SUPPORT_TIP_KEY = 'mobile_auth_face_not_support_tip';
export const getMobileAuthFaceNotSupportTip = (
  t: (key: typeof MOBILE_AUTH_FACE_NOT_SUPPORT_TIP_KEY) => string,
) => t('mobile_auth_face_not_support_tip');`,
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `const functionTypeReturn = 'key';
export type FunctionType = () => typeof functionTypeReturn;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "usedOnlyAsType",
					Line:      1,
					Column:    7,
				}},
			},
			{
				Code: `const methodSignatureReturn = 'key';
export interface MethodSignature { method(): typeof methodSignatureReturn; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "usedOnlyAsType",
					Line:      1,
					Column:    7,
				}},
			},
		},
	)
}

func TestNoUnusedVarsSignatureTypeParameters(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			{Code: `export type Result<T> = T extends (...args: infer _Args) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends (...args: infer Args) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends (value: infer Args) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends (value?: infer Args) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends (...args: [infer Args]) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends (value: Array<infer Args>) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends (...args: infer Args extends unknown[]) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends new (...args: infer Args) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends abstract new (...args: infer Args) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends { (...args: infer Args): infer R } ? R : never;`},
			{Code: `export type Result<T> = T extends { new (...args: infer Args): infer R } ? R : never;`},
			{Code: `export type Result<T> = T extends { method(...args: infer Args): infer R } ? R : never;`},
			// The whole parameter is visited, including nested signatures and their returns.
			{Code: `export type Result<T> = T extends (callback: () => infer Args) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends (callback: (value: infer Args) => void) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends (value: { property: infer Args }) => infer R ? R : never;`},
			{Code: `export type Result<T> = T extends (value: (unknown extends infer Args ? unknown : never)) => infer R ? R : never;`},
			// Repeated inferences share a binding, even when its first declaration is a return.
			{Code: `export type Result<T> = T extends { (): infer Args; (...args: infer Args): infer R } ? R : never;`},
			{Code: `export type Result<T> = T extends (one: infer Args, two: infer Args) => infer R ? R : never;`},
			{Code: `export type Result = (callback: <Unused>() => void) => void;`},
			{Code: `export declare function run<T>(value: T extends infer U ? unknown : never): void;`},
			{Code: `export abstract class A<T> { abstract run(value: T extends infer U ? unknown : never): void; }`},
			{Code: `export class A<T> { run(value: T extends infer U ? unknown : never): void; run(value: unknown): void { console.log(value); } }`},
			{Code: `export class A<T> { constructor(value: T extends infer U ? unknown : never); constructor(value: unknown) { console.log(value); } }`},
			{Code: `export class A<T> { set value(value: T extends infer U ? unknown : never) {} }`},
			{Code: `export type Result<T> = (value: { [key: string]: T extends infer U ? unknown : never }) => void;`},
			{Code: `export type Result<T> = ({ value }: { value: T extends infer U ? unknown : never }) => void;`},
			{Code: `export type Result<T> = (value: T extends infer U ? unknown : never) => value is never;`},
			// Upstream also visits names that are not TypeScript references.
			{Code: `export type Result<T> = T extends (arg: { U: unknown }) => infer U ? true : false;`},
			{Code: `export type Result<T> = T extends (...arg: [U: unknown]) => infer U ? true : false;`},
			{Code: `export type Result<U> = (arg: { U: unknown }) => void;`},
			{Code: `export type Result<U> = (...arg: [U: unknown]) => void;`},
			{Code: `export type Result<T> = T extends (arg: { [K in string]: infer U }) => void ? true : false;`},
			{Code: `export type Result<T> = T extends (arg: (unknown extends string ? never : infer U)) => void ? true : false;`},
			{Code: `export type Result<T> = T extends (arg: (unknown extends infer U ? U : never)) => void ? true : false;`},
			{Code: `export type Result<T> = T extends (arg: { [U: string]: infer U }) => void ? true : false;`},
			{Code: `export type Result<T> = T extends { (): infer U; <U>(arg: infer U): void; (arg: infer U): void } ? true : false;`},
			{Code: `export type Result<T> = (arg: <U extends T>() => void) => void;`},
			{Code: `export type Result<T> = (arg: () => (T extends infer U ? unknown : never)) => void;`},
			{Code: `export function run<T>(): (arg: T extends infer U ? unknown : never) => void { return () => {}; }`},
			{Code: `const value = 1; export abstract class A { abstract run(arg: typeof value): void; }`},
			{Code: `const value = 1; export class A { constructor(arg: typeof value); constructor(arg: number) { console.log(arg); } }`},
			{Code: `const value = 1; export class A { set value(arg: typeof value) {} }`},
			{Code: `const value = 1; export type Result = (arg: { [key: string]: typeof value }) => void;`},
			{
				Code: `export interface Used<T> { consume(value: T): T; }
export type Result<T> = T extends (...args: infer Args) => infer R ? R : never;
export type PropertyName<U> = (value: { U: unknown }) => void;`,
			},
			{
				Code:    `export type Result<T> = T extends (...args: infer _Args) => infer R ? R : never;`,
				Options: map[string]interface{}{"argsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `export type Result<T> = T extends (...args: infer Args) => infer R ? Args : never;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 66}},
			},
			{
				Code:   `export type Result<T> = T extends Array<infer U> ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 47}},
			},
			{
				Code:   `export type Result<T> = T extends [infer U] ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 42}},
			},
			{
				Code:   `export type Result<T> = T extends { value: infer U } ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 50}},
			},
			{
				Code:   `export type Result<T> = T extends { [key: string]: infer U } ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 58}},
			},
			{
				Code:   `const _Unused = 1; export {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
			},
			{
				Code:   `export type Result = <Unused>() => void;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 23}},
			},
			// A declaration in one scope must not mark a same-named binding in another.
			{
				Code:   `export type Result<Unused> = (callback: <Unused>() => void) => void;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 20}},
			},
			{
				Code:   `export type Result<T, Args> = T extends (...args: infer Args) => infer R ? R : never;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 23}},
			},
			{
				Code:   `export type Result<T> = [T extends (...args: infer U) => infer R ? R : never, T extends Array<infer U> ? true : false];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 101}},
			},
			{
				Code:   `export type Result<T> = T extends (arg: <U>(x: U) => infer U) => infer R ? R : never;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 60}},
			},
			{
				Code:   `export declare function run<T>(value: T): T extends infer U ? unknown : never;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 59}},
			},
			{
				Code:   `export function run<T>(value: T extends infer U ? unknown : never): void { console.log(value); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 47}},
			},
			{
				Code:   `export type Result<T> = T extends <U>(arg: infer U) => void ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 50}},
			},
			{
				Code:   `export type Result<T> = T extends (U: infer U) => void ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 45}},
			},
			{
				Code:   `export type Result<T> = T extends (...Args: infer Args) => void ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 51}},
			},
			{
				Code:   `export type Result<T> = T extends (U: unknown) => infer U ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 57}},
			},
			{
				Code:   `export type Result<U> = () => { U: unknown };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 20}},
			},
			{
				Code:   `export type Result<U> = () => [U: unknown];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 20}},
			},
			{
				Code:   `export type Result<T> = T extends { (): infer U; <U>(arg: infer U): void } ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 47}},
			},
			{
				Code:   `export type Result<T> = T extends (arg: { [U in string]: infer U }) => void ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 64}},
			},
			{
				Code:   `export type Result<T> = T extends (arg: { [U in string]: <U>() => infer U }) => void ? true : false;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 73}},
			},
			{
				Code:    `export type Result<U> = (arg: { U: unknown }) => void;`,
				Options: map[string]interface{}{"varsIgnorePattern": "^U$", "reportUsedIgnorePattern": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 1, Column: 20}},
			},
			{
				Code: `export interface Used<_T> { consume(value: _T): _T; }
export type Result<T> = T extends (...args: infer _Args) => infer R ? R : never;`,
				Options: map[string]interface{}{"varsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedIgnoredVar", Line: 1, Column: 23},
					{MessageId: "usedIgnoredVar", Line: 2, Column: 51},
				},
			},
			{
				Code: `export type Unused<T> = number;
export type Result<T> = T extends (...args: infer Args) => infer R ? R : never;
export type UnusedAgain<T> = number;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Line: 1, Column: 20},
					{MessageId: "unusedVar", Line: 3, Column: 25},
				},
			},
			{
				Code:    `export type Result<T> = T extends (arg: { [U in string]: infer U }) => void ? true : false;`,
				Options: map[string]interface{}{"varsIgnorePattern": "^U$", "reportUsedIgnorePattern": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 1, Column: 44}},
			},
			{
				Code:    `export type Result<T> = T extends (arg: { [U in string]: <U>() => infer U }) => void ? true : false;`,
				Options: map[string]interface{}{"varsIgnorePattern": "^U$", "reportUsedIgnorePattern": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedIgnoredVar", Line: 1, Column: 44},
					{MessageId: "usedIgnoredVar", Line: 1, Column: 59},
				},
			},
			{
				Code:    `export type Result<T> = T extends (...args: infer _Args) => infer R ? R : never;`,
				Options: map[string]interface{}{"varsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "usedIgnoredVar",
					Message:   "'_Args' is marked as ignored but is used. Used vars must not match /^_/u.",
					Line:      1,
					Column:    51,
				}},
			},
		},
	)
}

func TestNoUnusedVarsReportsWriteFromDeclarationScope(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{{
			Code: `// eslint-disable-next-line test
let pendingShareInfo: unknown;
export function update(value: unknown) {
  pendingShareInfo = value;
}`,
		}},
		[]rule_tester.InvalidTestCase{
			{
				Code: `let pendingShareInfo: unknown;
export function update(value: unknown) {
  pendingShareInfo = value;
}`,
				Options: map[string]interface{}{"varsIgnorePattern": "^_"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Message:   "'pendingShareInfo' is assigned a value but never used. Allowed unused vars must match /^_/u.",
					Line:      1,
					Column:    5,
				}},
			},
			{
				Code: `let pendingShareInfo: unknown;
export function update(value: unknown) {
  // eslint-disable-next-line test
  pendingShareInfo = value;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Line:      1,
					Column:    5,
				}},
			},
			{
				Code: `let pendingShareInfo: unknown;
pendingShareInfo = "same scope";
export function update(value: unknown) {
  pendingShareInfo = value;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar",
					Line:      2,
					Column:    1,
				}},
			},
		},
	)
}

// TestNoUnusedVarsUpstreamParityEdges covers boundaries of behavior aligned
// with typescript-eslint v8.72.1; each expectation was checked against it.
func TestNoUnusedVarsUpstreamParityEdges(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// A destructuring-assignment property with a rest sibling is ignored
			// for renamed targets too.
			{
				Code:    "let a, rest;\n({ key: a, ...rest } = obj);\nconsole.log(rest);\n",
				Options: map[string]interface{}{"ignoreRestSiblings": true},
			},
			// A function passed as an argument inside a self-assigned IIFE can run
			// later, so its read is a real use.
			{Code: "let cb = 0;\ncb = (function () {\n  register(function () {\n    cb();\n  });\n})();\n"},
			// A value read after a type-only one marks the later parameter used.
			{
				Code:    "export function g(a: string, b: number): typeof b {\n  return b;\n}\n",
				Options: map[string]interface{}{"args": "after-used"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Exporting only the type half of a merged declaration, even when it
			// comes first, leaves the value unused.
			{
				Code: "export interface Foo {}\nconst Foo = 1;\n",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar", Message: "'Foo' is assigned a value but never used.",
					Line: 2, Column: 7, EndLine: 2, EndColumn: 10,
				}},
			},
			// Without an initializer, the report stays on the first declaration.
			{
				Code: "export type Foo = string;\nlet Foo;\n",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar", Message: "'Foo' is defined but never used.",
					Line: 1, Column: 13, EndLine: 1, EndColumn: 16,
				}},
			},
			// A defaulted property is not a direct rest sibling.
			{
				Code:    "let a, rest;\n({ a = 1, ...rest } = obj);\nconsole.log(rest);\n",
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar", Message: "'a' is assigned a value but never used.",
					Line: 2, Column: 4, EndLine: 2, EndColumn: 5,
				}},
			},
			// Unlike ESLint core, an unused explicit global ignores `vars`.
			{
				Code:    "/* global foo */\n",
				Options: map[string]interface{}{"vars": "local"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar", Message: "'foo' is defined but never used.",
					Line: 1, Column: 11, EndLine: 1, EndColumn: 14,
				}},
			},
			{
				Code: "/* global a, b */\na;\n",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar", Message: "'b' is defined but never used.",
					Line: 1, Column: 14, EndLine: 1, EndColumn: 15,
				}},
			},
			{
				Code: "for (const k in obj) {\n}\n",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar", Message: "'k' is assigned a value but never used.",
					Line: 1, Column: 12, EndLine: 1, EndColumn: 13,
				}},
			},
			// Only the default import is used: remove `, { ... }`.
			{
				Code: "import Used, { Unused } from 'm';\nUsed;\n",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unusedVar", Message: "'Unused' is defined but never used.",
					Line: 1, Column: 16, EndLine: 1, EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    "import Used from 'm';\nUsed;\n",
					}},
				}},
			},
		},
	)
}

// TestNoUnusedVarsInlineGlobalBindings distinguishes authored global bindings
// from comment-only globals and same-name local declarations.
// Expectations were checked with typescript-eslint v8.72.1 and ESLint 10.14.0;
// JavaScript files use Espree and TypeScript files use the TypeScript parser.
// TypeScript listeners emit declaration diagnostics before EOF comment reports;
// public API diagnostics are sorted by source position.
func TestNoUnusedVarsInlineGlobalBindings(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// used var value = 1;
			{
				Code:            "/* global value */\nvar value = 1;\nconsume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// used let value = 1;
			{
				Code:            "/* global value */\nlet value = 1;\nconsume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// used const value = 1;
			{
				Code:            "/* global value */\nconst value = 1;\nconsume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// used function value() {}
			{
				Code:            "/* global value */\nfunction value() {}\nconsume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// used class value {}
			{
				Code:            "/* global value */\nclass value {}\nconsume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// used interface value {}
			{
				Code:            "/* global value */\ninterface value {}\nconsume(null as value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// used type value = number;
			{
				Code:            "/* global value */\ntype value = number;\nconsume(null as value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// used { var value = 1; }
			{
				Code:            "/* global value */\n{ var value = 1; }\nconsume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// exported comment
			{
				Code:            "/* global value:writable */\n/* exported value */\nvalue = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// exported reverse order
			{
				Code:            "/* exported value */\n/* global value */",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// exported authored
			{
				Code:            "/* global value */\n/* exported value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// hoisted read
			{
				Code:            "/* global value */\nconsume(value);\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// authored local option
			{
				Code:            "/* global value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Options:         map[string]interface{}{"vars": "local"},
			},
			// authored ignored
			{
				Code:            "/* global value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Options:         map[string]interface{}{"varsIgnorePattern": "^value$"},
			},
			// module file.ts exported global
			{
				Code:            "/* global value */\n/* exported value */\nvalue = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			// module file.js exported global
			{
				Code:            "/* global value */\n/* exported value */\nvalue = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				TSConfig:        "tsconfig.allow-js.json",
				FileName:        "inline-global.js",
			},
			// commonjs file.ts local read
			{
				Code:            "/* global value */\nvar value = 1; consume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			// commonjs file.ts exported global
			{
				Code:            "/* global value */\n/* exported value */\nvalue = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			// commonjs file.ts exported local
			{
				Code:            "/* global value */\n/* exported value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			// commonjs file.js exported global
			{
				Code:            "/* global value */\n/* exported value */\nvalue = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				TSConfig:        "tsconfig.allow-js.json",
				FileName:        "inline-global.js",
			},
		},
		[]rule_tester.InvalidTestCase{
			// unused var value = 1;
			{
				Code:            "/* global value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 5, EndLine: 2, EndColumn: 10},
				},
			},
			// unused let value = 1;
			{
				Code:            "/* global value */\nlet value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 5, EndLine: 2, EndColumn: 10},
				},
			},
			// unused const value = 1;
			{
				Code:            "/* global value */\nconst value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 7, EndLine: 2, EndColumn: 12},
				},
			},
			// unused function value() {}
			{
				Code:            "/* global value */\nfunction value() {}",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 2, Column: 10, EndLine: 2, EndColumn: 15},
				},
			},
			// unused class value {}
			{
				Code:            "/* global value */\nclass value {}",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 2, Column: 7, EndLine: 2, EndColumn: 12},
				},
			},
			// unused interface value {}
			{
				Code:            "/* global value */\ninterface value {}",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 2, Column: 11, EndLine: 2, EndColumn: 16},
				},
			},
			// unused type value = number;
			{
				Code:            "/* global value */\ntype value = number;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 2, Column: 6, EndLine: 2, EndColumn: 11},
				},
			},
			// unused { var value = 1; }
			{
				Code:            "/* global value */\n{ var value = 1; }",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 7, EndLine: 2, EndColumn: 12},
				},
			},
			// synthetic JSDoc is not a declaration
			{
				Code:            "/* global value */\n/** @typedef {number} value */",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				TSConfig:        "tsconfig.allow-js.json",
				FileName:        "inline-global.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// duplicate comments
			{
				Code:            "/* global value */\n/* global value:writable */",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// duplicate vars
			{
				Code:            "/* global value */\nvar value = 1;\nvar value;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 5, EndLine: 2, EndColumn: 10},
				},
			},
			// block shadow
			{
				Code:            "/* global value */\n{ let value = 1; consume(value); }",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// parameter shadow
			{
				Code:            "/* global value */\nfunction read(value: number) { return value; }\nread(1);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// exported shadow unused
			{
				Code:            "/* global value */\n/* exported value */\n{ let value = 1; }",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 3, Column: 7, EndLine: 3, EndColumn: 12},
				},
			},
			// global read and local shadow
			{
				Code:            "/* global value */\nconsume(value);\n{ let value = 1; }",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 3, Column: 7, EndLine: 3, EndColumn: 12},
				},
			},
			// comment local option
			{
				Code:            "/* global value */",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Options:         map[string]interface{}{"vars": "local"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// authored ignored used
			{
				Code:            "/* global value */\nvar value = 1; consume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Options:         map[string]interface{}{"varsIgnorePattern": "^value$", "reportUsedIgnorePattern": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedIgnoredVar", Message: "'value' is marked as ignored but is used. Used vars must not match /^value$/u.", Line: 2, Column: 5, EndLine: 2, EndColumn: 10},
				},
			},
			// module file.ts local read
			{
				Code:            "/* global value */\nvar value = 1; consume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// module file.ts local unused
			{
				Code:            "/* global value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 5, EndLine: 2, EndColumn: 10},
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// module file.ts exported local
			{
				Code:            "/* global value */\n/* exported value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 3, Column: 5, EndLine: 3, EndColumn: 10},
				},
			},
			// module file.js local read
			{
				Code:            "/* global value */\nvar value = 1; consume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				TSConfig:        "tsconfig.allow-js.json",
				FileName:        "inline-global.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// module file.js local unused
			{
				Code:            "/* global value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				TSConfig:        "tsconfig.allow-js.json",
				FileName:        "inline-global.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 5, EndLine: 2, EndColumn: 10},
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// module file.js exported local
			{
				Code:            "/* global value */\n/* exported value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				TSConfig:        "tsconfig.allow-js.json",
				FileName:        "inline-global.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 3, Column: 5, EndLine: 3, EndColumn: 10},
				},
			},
			// commonjs file.ts local unused
			{
				Code:            "/* global value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 5, EndLine: 2, EndColumn: 10},
				},
			},
			// commonjs file.js local read
			{
				Code:            "/* global value */\nvar value = 1; consume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				TSConfig:        "tsconfig.allow-js.json",
				FileName:        "inline-global.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// commonjs file.js local unused
			{
				Code:            "/* global value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				TSConfig:        "tsconfig.allow-js.json",
				FileName:        "inline-global.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 5, EndLine: 2, EndColumn: 10},
					{MessageId: "unusedVar", Message: "'value' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
				},
			},
			// commonjs file.js exported local
			{
				Code:            "/* global value */\n/* exported value */\nvar value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				TSConfig:        "tsconfig.allow-js.json",
				FileName:        "inline-global.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 3, Column: 5, EndLine: 3, EndColumn: 10},
				},
			},
		},
	)
}

// TestNoUnusedVarsNestedSelfAssignedFunctions distinguishes nested immediate
// calls from stored callbacks and externally consumed assignment results.
// Expectations were checked with typescript-eslint v8.72.1 and ESLint 10.14.0;
// JavaScript files use Espree and TypeScript files use the TypeScript parser.
// TypeScript listeners emit declaration diagnostics before EOF comment reports;
// public API diagnostics are sorted by source position.
func TestNoUnusedVarsNestedSelfAssignedFunctions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// typescript-eslint v8.72.1 reports these as unused because it checks
			// only the nearest function. Keep Rslint's existing exemption: the
			// outer callback escapes and its inner IIFE reads value when called.
			{Code: "let value = 0;\nvalue = register(() => (() => value)());"},
			{Code: "let value = 0;\nvalue = (() => register(() => (() => value)()))();"},
			{Code: "let value = 0;\nvalue = (() => { const read = () => (() => value)(); return read; })();"},

			// stored argument
			{
				Code:            "let value = 0;\nvalue = (() => register(() => value))();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			// stored variable
			{
				Code:            "let value = 0;\nvalue = (() => { const later = () => value; return later; })();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			// stored property
			{
				Code:            "let value = 0;\nvalue = (() => register({ read: () => value }))();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			// returned function
			{
				Code:            "let value = 0;\nvalue = (() => { return () => value; })();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			// function declaration
			{
				Code:            "let value = 0;\nvalue = (() => { function later() { return value; } return later; })();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			// consumed assignment
			{
				Code:            "let value = 0;\nconsume(value = (() => (() => value + 1)())());",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			// later read
			{
				Code:            "let value = 0;\nvalue = (() => (() => value + 1)())();\nconsume(value);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			// loop assignment
			{
				Code:            "let value = 0;\nwhile (again()) { value = (() => (() => value + 1)())(); }",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
			// cross function assignment
			{
				Code:            "let value = 0;\nfunction update() { value = (() => (() => value + 1)())(); }\nupdate();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// double arrow
			{
				Code:            "let value = 0;\nvalue = (() => (() => value + 1)())();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
			// triple arrow
			{
				Code:            "let value = 0;\nvalue = (() => (() => (() => value + 1)())())();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
			// block arrows
			{
				Code:            "let value = 0;\nvalue = (() => { return (() => { return value + 1; })(); })();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
			// mixed functions
			{
				Code:            "let value = 0;\nvalue = (function () { return (() => value + 1)(); })();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
			// arithmetic wrapper
			{
				Code:            "let value = 0;\nvalue = 1 + (() => (() => value + 1)())();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
			// comma callee
			{
				Code:            "let value = 0;\nvalue = (() => (0, () => value + 1)())();",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'value' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
		},
	)
}

// TestNoUnusedVarsReturnedCallbacks keeps concise arrow returns equivalent to
// block returns. typescript-eslint v8.72.1 shares the concise-return false
// positive; these callbacks can be retained and invoked after registration.
func TestNoUnusedVarsReturnedCallbacks(t *testing.T) {
	valid := []rule_tester.ValidTestCase{{Code: `
const callbacks: Array<() => string> = [];
function register(handler: () => string) {
  callbacks.push(handler);
  return () => "cancelled";
}
let cancel = () => "";
cancel = register((() => () => cancel())());
console.log(callbacks[0]());
`}}
	for _, expression := range []string{
		`(() => () => cancel())()`,
		`(() => { return () => cancel(); })()`,
		`(() => (() => cancel()))()`,
		`(() => (() => cancel()) as (() => string))()`,
		`(() => (() => cancel()) satisfies (() => string))()`,
		`(() => condition ? () => cancel() : () => "")()`,
		`(() => (touch(), () => cancel()))()`,
		`(() => ({ handler: () => cancel() }))()`,
		`(() => [() => cancel()])()`,
		`(() => (() => () => cancel())())()`,
		`(() => function () { return cancel(); })()`,
	} {
		valid = append(valid, rule_tester.ValidTestCase{
			Code: "let cancel = () => \"\";\ncancel = register(" + expression + ");",
		})
	}
	var invalid []rule_tester.InvalidTestCase
	for _, expression := range []string{
		`(() => (() => value + 1)())()`,
		`(() => { return (() => value + 1)(); })()`,
		`(() => (touch(), (() => value + 1)()))()`,
		`(() => condition ? (() => value + 1)() : 0)()`,
		`(() => (() => value, 0))()`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: "let value = 0;\nvalue = " + expression + ";",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Message: "'value' is assigned a value but never used.",
				Line: 2, Column: 1, EndLine: 2, EndColumn: 6,
			}},
		})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, valid, invalid)
}

// TestNoUnusedVarsReviewBoundaries covers cross-file declarations, reference
// namespaces, parameter writes, implicit libraries, and function value consumers.
// Expected diagnostics follow v8.72.1 except the documented callback fixes.
func TestNoUnusedVarsReviewBoundaries(t *testing.T) {
	root := fixtures.GetRootDir()
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "round3-legacy.json"):    `{"extends":"./tsconfig.json","compilerOptions":{"moduleDetection":"legacy"}}`,
		tspath.ResolvePath(root.Dir, "round3-no-lib.json"):    `{"compilerOptions":{"target":"esnext","strict":true,"noLib":true,"moduleDetection":"legacy","types":[]}}`,
		tspath.ResolvePath(root.Dir, "round3-initialized.ts"): `var RoundThreeInitialized = class { value = 1; };`,
		tspath.ResolvePath(root.Dir, "round3-no-dom.json"):    `{"extends":"./round3-legacy.json","compilerOptions":{"lib":["esnext"]}}`,
		tspath.ResolvePath(root.Dir, "round3-host.d.ts"):      `declare var RoundThreeExternal: { new(): any }; declare function roundThreeHost(): void; interface RoundThreeHostType {}`,
	})
	rule_tester.RunRuleTester(root, "round3-legacy.json", t, &NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// window
			{Code: "interface Window { customSize: number }\n", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			// document
			{Code: "interface Document { customSize: number }\n", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			// value-global
			{Code: "/* global launch */\nexport function run() { type launch = number; const code: launch = 1; launch(code); }", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// type-global
			{Code: "/* global Shape */\nexport function run() { const Shape = 1; let x: Shape; consume(x, Shape); }", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// parameter-value-read
			{Code: "export function g(a: string, b: number): typeof b { return b; }", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// library
			{Code: "/* global Promise, Array */", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// factory
			{Code: "function Factory() {}\nexport type Factory = number;\nexport const n: Factory = 1;", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// factory-reverse
			{Code: "export type Factory = number;\nfunction Factory() {}\nexport const n: Factory = 1;", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// v8.72.1 mistakes bind for executing the callback. The bound
			// function can be retained by subscribe, so its read is a use.
			{Code: "declare function subscribe(callback: () => void): () => void;\nlet off = () => {};\noff = subscribe(function () { off(); }.bind(null));", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// returned-bind
			{Code: "let off = () => {};\noff = subscribe((() => () => off())().bind(null));", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// bound-returned-escape
			{Code: "let off = () => {};\noff = subscribe((() => () => off()).bind(null)());", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// bind-chain
			{Code: "let off = () => {};\noff = subscribe((() => () => off())().bind(null).bind(null));", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// bind-computed
			{Code: "let off = () => {};\noff = subscribe((function () { off(); })[\"bind\"](null));", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// returned-wrapper-call
			{Code: "let off = () => {};\noff = subscribe((() => () => off()).call(null));", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
			// unknown-method
			{Code: "let off = () => {};\noff = subscribe((() => off()).custom());", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		},
		[]rule_tester.InvalidTestCase{
			// window-module
			{Code: "interface Window { customSize: number }\n", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'Window' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			// external-merge
			{Code: "interface RoundThreeExternal { value: number }\n", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'RoundThreeExternal' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 29},
				},
			},
			// local-type-only
			{Code: "/* global launch */\nexport function run() { type launch = number; const code: launch = 1; consume(code); }", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'launch' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			// local-value-only
			{Code: "/* global launch */\nexport function run() { const launch = () => 1; launch(); }", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'launch' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			// parameter-write
			{Code: "export function g(a: string, b: number): typeof b { b = 3; return 1; }", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedOnlyAsType", Message: "'b' is assigned a value but only used as a type.", Line: 1, Column: 53, EndLine: 1, EndColumn: 54},
				},
			},
			// parameter-only-type
			{Code: "export function g(a: string, b: number): typeof b { return 1; }", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'a' is defined but never used.", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "usedOnlyAsType", Message: "'b' is defined but only used as a type.", Line: 1, Column: 30, EndLine: 1, EndColumn: 31},
				},
			},
			// parameter-destructure
			{Code: "export function g(a: string, b: number): typeof b { [b] = [3]; return 1; }", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedOnlyAsType", Message: "'b' is assigned a value but only used as a type.", Line: 1, Column: 54, EndLine: 1, EndColumn: 55},
				},
			},
			// parameter-compound
			{Code: "export function g(a: string, b: number): typeof b { b += 3; return 1; }", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedOnlyAsType", Message: "'b' is assigned a value but only used as a type.", Line: 1, Column: 53, EndLine: 1, EndColumn: 54},
				},
			},
			// library-off
			{Code: "/* global Promise, Array */", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				TSConfig: "round3-no-lib.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'Promise' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
					{MessageId: "unusedVar", Message: "'Array' is defined but never used.", Line: 1, Column: 20, EndLine: 1, EndColumn: 25},
				},
			},
			// mixed-library-host
			{Code: "/* global Promise, roundThreeHost */", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'roundThreeHost' is defined but never used.", Line: 1, Column: 20, EndLine: 1, EndColumn: 34},
				},
			},
			// library-local
			{Code: "/* global Promise */\nconst Promise = 1;", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'Promise' is assigned a value but never used.", Line: 2, Column: 7, EndLine: 2, EndColumn: 14},
				},
			},
			// host-dts
			{Code: "/* global roundThreeHost */", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'roundThreeHost' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
				},
			},
			// library-intrinsics
			{Code: "/* global undefined, NaN, Infinity */", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'undefined' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 20},
					{MessageId: "unusedVar", Message: "'NaN' is defined but never used.", Line: 1, Column: 22, EndLine: 1, EndColumn: 25},
					{MessageId: "unusedVar", Message: "'Infinity' is defined but never used.", Line: 1, Column: 27, EndLine: 1, EndColumn: 35},
				},
			},
			// factory-unused
			{Code: "function Factory() {}\nexport type Factory = number;", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'Factory' is defined but never used.", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			// const-factory
			{Code: "const Factory = () => {};\nexport type Factory = number;\nexport const n: Factory = 1;", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedOnlyAsType", Message: "'Factory' is assigned a value but only used as a type.", Line: 1, Column: 7, EndLine: 1, EndColumn: 14},
				},
			},
			// bound-immediate
			{Code: "let x = 0;\nx = (() => x + 1).bind(null)();", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'x' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			// returned-immediate
			{Code: "let x = 0;\nx = (() => () => x + 1)()();", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'x' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			// v8.72.1 mistakes explicit return for storage. This returned
			// function is immediately called and only updates x.
			{Code: "let x = 0;\nx = (() => { return () => x + 1; })()();", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'x' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			// bound-returned-immediate
			{Code: "let x = 0;\nx = (() => () => x + 1).bind(null)()();", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'x' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			// three-returned-immediate
			{Code: "let x = 0;\nx = (() => () => () => x + 1)()()();", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'x' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			// returned-call
			{Code: "let x = 0;\nx = (() => () => x + 1)().call(null);", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'x' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			// returned-apply
			{Code: "let x = 0;\nx = (() => () => x + 1)().apply(null, []);", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'x' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			// parameter-write-only
			{Code: "export function g(a: string, b: number) { b = 3; }", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'b' is assigned a value but never used.", Line: 1, Column: 43, EndLine: 1, EndColumn: 44},
				},
			},
			// external-initializer
			{Code: "interface RoundThreeInitialized { value: number }\n", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'RoundThreeInitialized' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32},
				},
			},
			// returned-discarded
			{Code: "let x = 0;\nx = ((() => () => x + 1)(), 1);", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'x' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
				},
			},
			// library-no-dom
			{Code: "/* global Window, Promise */", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				TSConfig: "round3-no-dom.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'Window' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			// host-type-dts
			{Code: "/* global RoundThreeHostType */", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Message: "'RoundThreeHostType' is defined but never used.", Line: 1, Column: 11, EndLine: 1, EndColumn: 29},
				},
			},
		},
	)
}

// TestNoUnusedVarsCallbackExecutionBoundaries covers deferred execution, literal
// extraction, and metadata reads without invoking a callback. ESLint 10.14.0
// and typescript-eslint 8.72.1 also miss some deferred reads and explicit
// return consumers; these expectations preserve runtime behavior instead.
func TestNoUnusedVarsCallbackExecutionBoundaries(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `let x; x = (async () => { await 0; (() => consume(x))(); })();`},
		{Code: `let x; x = (function* () { (() => consume(x))(); })();`},
		{Code: `let x; x = (() => { return async () => { await 0; (() => consume(x))(); }; })()();`},
		{Code: `let x; x = (() => x).bind(null, consume(() => x)).name;`},
		{Code: `let x; x = (() => { return async () => { await 0; consume(x); }; })()();`},
		{Code: `let x; x = (() => async () => { await 0; consume(x); })()();`},
		{Code: `let x; x = (async () => { await 0; consume(x); })();`},
		{Code: `let x; x = (() => { return function* () { yield x; }; })()();`},
		{Code: `let x; x = (function* () { yield x; })();`},
		{Code: `let x; x = (() => async () => { await 0; consume(x); })().call(null);`},
		{Code: `let x; x = (() => function* () { yield x; })().bind(null)();`},
		{Code: `let x; x = [(async () => { await 0; consume(x); })][0]();`},
		{Code: `let x; x = ({ async run() { await 0; consume(x); } }).run();`},
		{Code: `let x; x = [() => x][index]();`},
		{Code: `let x; x = ({run: () => x})[key]();`},
		{Code: `let x; x = ({run: () => x, ...source}).run();`},
		{Code: `let x; x = ({run: () => x, [key]: replacement}).run();`},
		{Code: `let x; x = ({get run() { return () => x; }}).run();`},
		{Code: `let x; x = consume([() => x]);`},
		{Code: `let x; x = consume({run: () => x});`},
		{Code: `let x; x = (() => x).custom();`},
		{Code: `let x; x = (() => x).bind(null, consume(() => x)).length;`},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, expression := range []string{
		`[() => x + 1][0]()`,
		`[() => x + 1]["0"]()`,
		`({run: () => x + 1}).run()`,
		`({run() { return x + 1; }}).run()`,
		`({["run"]: () => x + 1})["run"]()`,
		`({run: () => x + 1}).run.call(null)`,
		`[{run: () => x + 1}][0].run()`,
		`({run: [() => x + 1]}).run[0]()`,
		`(() => [() => x + 1])()[0]()`,
		`(() => { return {run: () => x + 1}; })().run()`,
		`(() => x).length`,
		`(() => x)["name"]`,
		`(() => x).bind(null).length`,
		`(() => () => x)().name`,
		`[() => x][0].length`,
		`({run() { return x; }}).run.name`,
		`(async () => x).length`,
		`(function* () { yield x; }).name`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:   "let x = 0;\nx = " + expression + ";",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Message: "'x' is assigned a value but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2}},
		})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, valid, invalid)
}

func TestNoUnusedVarsAuthoredJavaScriptLibraryNames(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, sourceType := range []string{"script", "module"} {
		for _, name := range []string{"RequestInit", "Window", "Promise"} {
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: "var " + name + " = {};", FileName: "authored-library.js", TSConfig: "tsconfig.allow-js.json",
				LanguageOptions: rule.LanguageOptions{SourceType: sourceType},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Message: "'" + name + "' is assigned a value but never used."}},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, nil, invalid)
}
