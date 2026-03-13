package prefer_includes

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferIncludesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferIncludesRule, []rule_tester.ValidTestCase{
		{Code: `function f(a: string): void {
  a.indexOf(b);
}`},
		{Code: `function f(a: string): void {
  a.indexOf(b) + 0;
}`},
		{Code: `function f(a: string | { value: string }): void {
  a.indexOf(b) !== -1;
}`},
		{Code: `type UserDefined = {
  indexOf(x: any): number;
};
function f(a: UserDefined): void {
  a.indexOf(b) !== -1;
}`},
		{Code: `type UserDefined = {
  indexOf(x: any, fromIndex?: number): number;
  includes(x: any): boolean;
};
function f(a: UserDefined): void {
  a.indexOf(b) !== -1;
}`},
		{Code: `type UserDefined = {
  indexOf(x: any, fromIndex?: number): number;
  includes(x: any, fromIndex: number): boolean;
};
function f(a: UserDefined): void {
  a.indexOf(b) !== -1;
}`},
		{Code: `type UserDefined = {
  indexOf(x: any, fromIndex?: number): number;
  includes: boolean;
};
function f(a: UserDefined): void {
  a.indexOf(b) !== -1;
}`},
		{Code: `function f(a?: string): void {
	  /bar/.test(a);
	}`},
		{Code: `function f(a: string): void {
		  let pattern = /foo/;
		  pattern = /bar/;
		  pattern.test(a);
		}`},
		{Code: `function f(a: string): void {
		  /bar/y.test(a);
		}`},
		{Code: `function f(a: string): void {
		  /\01/.test(a);
		}`},
		{Code: `function f(a: string, undefined: string): void {
		  new RegExp('bar', undefined).test(a);
		}`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `function f(a: string): void {
  a.indexOf(b) !== -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes", Line: 2, Column: 3}},
			Output: []string{`function f(a: string): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f(a: string): void {
  (a.indexOf(b)) === -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes", Line: 2, Column: 3}},
			Output: []string{`function f(a: string): void {
  (!a.includes(b));
}`},
		},
		{
			Code: `function f(a: string): void {
  a.indexOf(b) === -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes", Line: 2, Column: 3}},
			Output: []string{`function f(a: string): void {
  !a.includes(b);
}`},
		},
		{
			Code: `function f(a: string): void {
  a.indexOf(b) == -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: string): void {
  !a.includes(b);
}`},
		},
		{
			Code: `function f(a?: string): void {
  a?.indexOf(b) === -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes", Line: 2, Column: 3}},
		},
		{
			Code: `function f(a?: string): void {
  a?.indexOf(b) !== -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
		},
		{
			Code: `function f(a?: { b: string[] }, c: string): void {
  a?.b.indexOf(c) === -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
		},
		{
			Code: `function f(a: Uint8Array): void {
  a.indexOf(b) !== -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: Uint8Array): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f(a: string): void {
  a.indexOf(b) >= 0x0;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: string): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f(a: string): void {
  a.indexOf(b) > -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: string): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f(a: string): void {
  a.indexOf(b) != -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: string): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f(a: string): void {
  a.indexOf(b) < 0;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: string): void {
  !a.includes(b);
}`},
		},
		{
			Code: `function f(a: string): void {
  a.indexOf(b) <= -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: string): void {
  !a.includes(b);
}`},
		},
		{
			Code: `function f(a: string): void {
  /bar/.test(a);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferStringIncludes"}},
			Output: []string{`function f(a: string): void {
  a.includes('bar');
}`},
		},
		{
			Code: `function f(a: string[]): void {
  /bar/.test(a);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferStringIncludes"}},
		},
		{
			Code: `function f(a: string): void {
  new RegExp('bar', undefined).test(a);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferStringIncludes"}},
			Output: []string{`function f(a: string): void {
  a.includes('bar');
}`},
		},
		{
			Code: `function f(a: string): void {
  const pattern = /bar/;
  pattern.test(a);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferStringIncludes"}},
			Output: []string{`function f(a: string): void {
  const pattern = /bar/;
  a.includes('bar');
}`},
		},
		{
			Code: `function f(a: string): void {
  /bar/u.test(a);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferStringIncludes"}},
			Output: []string{`function f(a: string): void {
  a.includes('bar');
}`},
		},
		{
			Code: `const pattern = /bar/;
function f(a: string): void {
  pattern?.test(a);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferStringIncludes"}},
			Output: []string{`const pattern = /bar/;
function f(a: string): void {
  a?.includes('bar');
}`},
		},
		{
			Code: `const pattern = /bar/;
function f(a: string, b: string): void {
  pattern.test(a + b);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferStringIncludes"}},
			Output: []string{`const pattern = /bar/;
function f(a: string, b: string): void {
  (a + b).includes('bar');
}`},
		},
		{
			Code: `function f(a: Int32Array): void {
  a.indexOf(b) !== -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: Int32Array): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f(a: Float64Array): void {
  a.indexOf(b) !== -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: Float64Array): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f(a: ReadonlyArray<any>): void {
  a.indexOf(b) !== -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: ReadonlyArray<any>): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f(a: Readonly<any[]>): void {
  a.indexOf(b) !== -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f(a: Readonly<any[]>): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f<T>(a: T[] | ReadonlyArray<T>): void {
  a.indexOf(b) !== -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes"}},
			Output: []string{`function f<T>(a: T[] | ReadonlyArray<T>): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f(a: string): void {
  /\0'\\\n\r\v\t\f/.test(a);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferStringIncludes"}},
			Output: []string{`function f(a: string): void {
  a.includes('\0\'\\\n\r\v\t\f');
}`},
		},
	})
}
