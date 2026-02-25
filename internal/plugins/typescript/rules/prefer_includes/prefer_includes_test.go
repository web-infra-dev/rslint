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
		{Code: `function f(a?: string): void {
  /bar/.test(a);
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
			Code: `function f(a?: string): void {
  a?.indexOf(b) === -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes", Line: 2, Column: 3}},
		},
		{
			Code: `function f(a: Uint8Array): void {
  a.indexOf(b) !== -1;
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferIncludes", Line: 2, Column: 3}},
			Output: []string{`function f(a: Uint8Array): void {
  a.includes(b);
}`},
		},
		{
			Code: `function f(a: string): void {
  /bar/.test(a);
}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferStringIncludes", Line: 2, Column: 3}},
			Output: []string{`function f(a: string): void {
  a.includes('bar');
}`},
		},
	})
}
