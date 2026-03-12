package prefer_regexp_exec

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferRegExpExecRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferRegExpExecRule, []rule_tester.ValidTestCase{
		{Code: `const value = "foo"; /foo/g.exec(value);`},
		{Code: `const value = "foo"; value.match(/foo/g);`},
		{Code: `const value = "foo"; value.match(pattern);`},
		{Code: `let search = /foo/; const value = "foo"; value.match(search);`},
		{Code: `const value = "foo"; declare const flags: string; value.match(new RegExp("foo", flags));`},
		{Code: `const value = "foo"; value.match("[a-z");`},
		{Code: `const value = "foo"; value.search(/foo/);`},
		{Code: `const value: { match(v: string): any } = { match: () => null as any }; value.match("foo");`},
		{Code: `const value = "foo"; const pattern = /foo/g as RegExp; value.match(pattern);`},
		{Code: `const value = "foo"; value.match(new RegExp(/foo/g));`},
		{Code: `const value = "foo"; value.match(new RegExp(/foo/g, undefined));`},
		{Code: `function test(value: string, undefined: string) { value.match(new RegExp("foo", undefined)); }`},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `const value = "foo"; value.match(/foo/);`,
			Output: []string{`const value = "foo"; /foo/.exec(value);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch", Line: 1, Column: 28}},
		},
		{
			Code:   `const value = "foo"; value.match("foo");`,
			Output: []string{`const value = "foo"; /foo/.exec(value);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch", Line: 1, Column: 28}},
		},
		{
			Code: `
const value = "foo";
const reg: RegExp = /foo/;
value.match(reg);`,
			Output: []string{`
const value = "foo";
const reg: RegExp = /foo/;
reg.exec(value);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch", Line: 4, Column: 7}},
		},
		{
			Code:   `const value = "foo"; value["match"](/foo/);`,
			Output: []string{`const value = "foo"; /foo/.exec(value);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch"}},
		},
		{
			Code:   `const value = "foo"; value.match(new RegExp("foo", undefined));`,
			Output: []string{`const value = "foo"; new RegExp("foo", undefined).exec(value);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch"}},
		},
		{
			Code:   `const value = "foo"; value.match("\\d+");`,
			Output: []string{`const value = "foo"; /\d+/.exec(value);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch"}},
		},
		{
			Code:   `const value = "foo"; value.match("a\nb");`,
			Output: []string{`const value = "foo"; /a\nb/.exec(value);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch"}},
		},
		{
			Code: `
function test(value: string, pattern: string) {
  value.match(pattern);
}`,
			Output: []string{`
function test(value: string, pattern: string) {
  RegExp(pattern).exec(value);
}`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch"}},
		},
		{
			Code: `
function test(value: string, a: string, b: string, cond: boolean) {
  value.match(cond ? a : b);
}`,
			Output: []string{`
function test(value: string, a: string, b: string, cond: boolean) {
  RegExp((cond ? a : b)).exec(value);
}`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch"}},
		},
		{
			Code: `
const value = "foo";
const reg = /foo/ as RegExp;
value.match(reg);`,
			Output: []string{`
const value = "foo";
const reg = /foo/ as RegExp;
reg.exec(value);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch"}},
		},
		{
			Code: `
const value = "foo";
const reg: RegExp | null = /foo/;
value.match(reg!);`,
			Output: []string{`
const value = "foo";
const reg: RegExp | null = /foo/;
(reg!).exec(value);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch"}},
		},
		{
			Code:   `const value = "foo"; value.match(/foo/).toString();`,
			Output: []string{`const value = "foo"; (/foo/.exec(value)).toString();`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "regExpExecOverStringMatch"}},
		},
	})
}
