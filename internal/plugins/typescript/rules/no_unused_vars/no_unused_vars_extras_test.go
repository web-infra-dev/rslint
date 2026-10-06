package no_unused_vars

import (
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
