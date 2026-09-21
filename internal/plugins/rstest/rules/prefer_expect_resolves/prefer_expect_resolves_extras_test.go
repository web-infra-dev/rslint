package prefer_expect_resolves_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	impl "github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_expect_resolves"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"strings"
	"testing"
)

func TestRstestResolvesExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{}
	for _, code := range []string{
		"expect(await p).resolves.toBe(true);",
		"expect(await p).rejects.toThrow('reason');",
		"expect.soft(await p).toBe(true);",
		"expect.poll(await p).toBe(true);",
		"expect.element(await p).toBeVisible();",
		"expect(await p).to.be.true;",
		"expect(await p).to.equal(1).and.equal(1);",
		"expect(await p).toHaveTitle('title');",
		"expect(await 1).toBe(1);",
		"declare function fail(): never; expect(await fail()).toBe(1);",
		"declare const p: Promise<number> | number; expect(await p).toBe(1);",
		"import { expect } from 'vitest'; expect(await p).toBe(1);",
		"function run(expect: any) { expect(await p).toBe(1); }",
		"let { expect } = require('@rstest/core'); expect = replacement; expect(await p).toBe(1);",
		"expect().nothing();", "expect.hasAssertions();",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code})
	}
	invalid := []rule_tester.InvalidTestCase{}
	for _, pair := range [][2]string{
		{"expect(await p).toBe(1);", "await expect(p).resolves.toBe(1);"},
		{"expect((await p)).not.toBe(2);", "await expect((p)).resolves.not.toBe(2);"},
		{"(expect(await p)).toBe(1);", "await (expect(p).resolves).toBe(1);"},
		{"await (expect(await p).toBe(1));", "await (expect(p).resolves.toBe(1));"},
		{"expect(await /* keep */ p).toBe(1);", "await expect(/* keep */ p).resolves.toBe(1);"},
		{"expect(await (p as Promise<number>)).toBe(1);", "await expect((p as Promise<number>)).resolves.toBe(1);"},
		{"expect(await p!).toBe(1);", "await expect(p!).resolves.toBe(1);"},
		{"expect(await p, 'result')['toBe'](1);", "await expect(p, 'result').resolves['toBe'](1);"},
		{"return expect(await p).toBe(1);", "return await expect(p).resolves.toBe(1);"},
	} {
		prefix := "declare const p: Promise<number>; async function run() { "
		code, output := prefix+pair[0]+" }", prefix+pair[1]+" }"
		start := strings.Index(code, "expect(")
		column := start + strings.Index(code[start:], "await") + 1
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Output: []string{output}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectResolves", Line: 1, Column: column}}})
	}
	for _, pair := range [][2]string{
		{"import { expect as check } from '@rstest/core'; ", "check"},
		{"import * as core from '@rstest/core'; ", "core.expect"},
		{"const { expect: check } = require('rstack/test'); ", "check"},
		{"const core = require('@rstest/core'); ", "core.expect"},
		{"const { expect: check } = import.meta.rstest; ", "check"},
		{"", "import.meta.rstest.expect"},
		{"import { expect } from '@rstest/playwright'; ", "expect"},
	} {
		code := pair[0] + pair[1] + "(await pending).toBe(1);"
		output := pair[0] + "await " + pair[1] + "(pending).resolves.toBe(1);"
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectResolves", Line: 1, Column: strings.Index(code, "await") + 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestExpectResolves", Output: output}}}}})
	}
	for _, code := range []string{
		"expect(await p).toBe(expected);",
		"expect(await p).to.equal(1);",
		"expect(await p).toBe(compute());",
		"expect(await p, message()).toBe(1);",
		"expect<number>(await p).toBe(1);",
		"expect?.(await p).toBe(1);",
		"expect(await p)?.toBe(1);",
		"const result = expect(await p).toBe(1);",
		"expect(await p).toThrow();",
		"expect(await p).toCustom(1);",
		"expect.extend({toBe(){return {pass:true,message:()=>''}}}); expect(await p).toBe(1);",
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectResolves", Line: 1, Column: strings.Index(code, "await") + 1}}})
	}
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code:   `const π = 1; expect(await pending).toBe(π);`,
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectResolves", Line: 1, Column: 21}},
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &impl.PreferExpectResolvesRule, valid, invalid)
}
