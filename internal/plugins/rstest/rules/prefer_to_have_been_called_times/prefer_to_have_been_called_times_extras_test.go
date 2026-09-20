package prefer_to_have_been_called_times

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToHaveBeenCalledTimesExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{}
	addValid := func(code ...string) {
		for _, item := range code {
			valid = append(valid, rule_tester.ValidTestCase{Code: item})
		}
	}
	var invalid []rule_tester.InvalidTestCase
	// add reports the diagnostic on the bare `toHaveLength` accessor. Cases whose
	// matcher is written as a string or template key are listed explicitly.
	add := func(code, output string) {
		t.Helper()
		start := strings.Index(code, "toHaveLength")
		line := strings.Count(code[:start], "\n") + 1
		column := start - strings.LastIndex(code[:start], "\n")
		item := rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "preferMatcher", Line: line, Column: column, EndLine: line, EndColumn: column + len("toHaveLength")},
		}}
		if output != "" {
			item.Output = []string{output}
		}
		invalid = append(invalid, item)
	}

	// Optional links would leave a dangling `?.` or drop a guard when removed.
	addValid(
		`expect(fn?.mock.calls).toHaveLength(1);`,
		`expect(fn.mock?.calls).toHaveLength(1);`,
		`expect(fn?.["mock"].calls).toHaveLength(1);`,
		`expect(fn["mock"]?.["calls"]).toHaveLength(1);`,
	)
	// A computed identifier key is a variable read, not the member it spells.
	addValid(
		`expect(fn[mock].calls).toHaveLength(1);`,
		`expect(fn.mock[calls]).toHaveLength(1);`,
		`expect(fn["mo" + "ck"].calls).toHaveLength(1);`,
		`expect(fn.mock[0]).toHaveLength(1);`,
	)
	// Type assertions and non-null wrappers are not mock accessors.
	addValid(
		`expect((fn.mock.calls as never)).toHaveLength(1);`,
		`expect(fn.mock.calls!).toHaveLength(1);`,
	)
	// Neighbouring mock state and unrelated subjects.
	addValid(
		`expect(fn.mock.results).toHaveLength(1);`,
		`expect(fn.mock.calls.length).toHaveLength(1);`,
		`expect(fn.mock.calls[0]).toHaveLength(1);`,
		`expect(calls).toHaveLength(1);`,
		`expect().toHaveLength(1);`,
		`expect(fn.mock.calls).to.have.length(1);`,
		`expect(fn.mock.calls).toHaveProperty('length', 1);`,
		`expect.assertions(1);`,
		// A dynamic matcher name is not resolved as a matcher at all.
		`expect(fn.mock.calls)[toHaveLength](1);`,
		`expect(fn.mock.calls)["toHave" + "Length"](1);`,
	)
	// expect.element asserts on a browser locator, which carries no mock context.
	addValid(`expect.element(fn.mock.calls).toHaveLength(1);`)
	// Roots that are not Rstest's expect.
	addValid(
		`import { expect } from 'vitest'; expect(fn.mock.calls).toHaveLength(1);`,
		`import { expect } from '@jest/globals'; expect(fn.mock.calls).toHaveLength(1);`,
		`import type { expect as check } from '@rstest/core'; check(fn.mock.calls).toHaveLength(1);`,
		`const expect = custom; expect(fn.mock.calls).toHaveLength(1);`,
		`import { expect } from '@rstest/core'; function run(expect: any) { expect(fn.mock.calls).toHaveLength(1); }`,
		`let core = require('@rstest/core'); core = other; core.expect(fn.mock.calls).toHaveLength(1);`,
	)

	// Accessor spellings and receivers the removal keeps balanced.
	for _, pair := range [][2]string{
		{`expect(fn.mock.calls).toHaveLength(1);`, `expect(fn).toHaveBeenCalledTimes(1);`},
		{`expect(fn["mock"]["calls"]).toHaveLength(1);`, `expect(fn).toHaveBeenCalledTimes(1);`},
		{"expect(fn[`mock`].calls).toHaveLength(1);", `expect(fn).toHaveBeenCalledTimes(1);`},
		{`expect((fn).mock.calls).toHaveLength(1);`, `expect((fn)).toHaveBeenCalledTimes(1);`},
		{`expect(((fn.mock.calls))).toHaveLength(1);`, `expect(((fn))).toHaveBeenCalledTimes(1);`},
		{`expect((fn.mock).calls).toHaveLength(1);`, `expect((fn)).toHaveBeenCalledTimes(1);`},
		{`expect(((fn).mock).calls).toHaveLength(1);`, `expect(((fn))).toHaveBeenCalledTimes(1);`},
		{`expect(this.fn.mock.calls).toHaveLength(1);`, `expect(this.fn).toHaveBeenCalledTimes(1);`},
		{`expect(getMock().mock.calls).toHaveLength(1);`, `expect(getMock()).toHaveBeenCalledTimes(1);`},
		{`expect(mocks[index].mock.calls).toHaveLength(1);`, `expect(mocks[index]).toHaveBeenCalledTimes(1);`},
		{`expect((fn as never).mock.calls).toHaveLength(1);`, `expect((fn as never)).toHaveBeenCalledTimes(1);`},
		{`expect(fn.mock.calls, 'called once').toHaveLength(1);`, `expect(fn, 'called once').toHaveBeenCalledTimes(1);`},
		{`expect(fn.mock.calls).toHaveLength(count);`, `expect(fn).toHaveBeenCalledTimes(count);`},
		{`expect(fn.mock.calls).toHaveLength(1, 2);`, `expect(fn).toHaveBeenCalledTimes(1, 2);`},
		{`expect(fn.mock.calls).toHaveLength();`, `expect(fn).toHaveBeenCalledTimes();`},
		{`expect(fn.mock.calls).not.toHaveLength(1);`, `expect(fn).not.toHaveBeenCalledTimes(1);`},
		{`expect(fn.mock.calls).rejects.not.toHaveLength(1);`, `expect(fn).rejects.not.toHaveBeenCalledTimes(1);`},
		{`expect(fn.mock.calls)?.toHaveLength(1);`, `expect(fn)?.toHaveBeenCalledTimes(1);`},
		{`expect(fn.mock.calls).toHaveLength?.(1);`, `expect(fn).toHaveBeenCalledTimes?.(1);`},
		{`(expect(fn.mock.calls).toHaveLength)(1);`, `(expect(fn).toHaveBeenCalledTimes)(1);`},
		{`(expect(fn.mock.calls).toHaveLength(1));`, `(expect(fn).toHaveBeenCalledTimes(1));`},
		{`await expect(fn.mock.calls).toHaveLength(1);`, `await expect(fn).toHaveBeenCalledTimes(1);`},
		{`expect(fn.mock.calls) /* keep */ .toHaveLength(1);`, `expect(fn) /* keep */ .toHaveBeenCalledTimes(1);`},
		{`expect(fn.mock.calls).toHaveLength(/* keep */ 1);`, `expect(fn).toHaveBeenCalledTimes(/* keep */ 1);`},
		{"expect(\n  fn.mock.calls,\n).toHaveLength(1);", "expect(\n  fn,\n).toHaveBeenCalledTimes(1);"},
	} {
		add(pair[0], pair[1])
	}
	for _, factory := range []string{"expect", "expect.soft"} {
		add(factory+`(fn.mock.calls).toHaveLength(1);`, factory+`(fn).toHaveBeenCalledTimes(1);`)
	}
	// Every Rstest expect provenance the parser resolves.
	for _, pair := range [][2]string{
		{`import { expect as check } from '@rstest/core';`, "check"},
		{`import { expect } from 'rstack/test';`, "expect"},
		{`import * as core from '@rstest/core';`, "core.expect"},
		{`const { expect: check } = require('@rstest/core');`, "check"},
		{`const core = require('rstack/test');`, "core.expect"},
		{`import { expect } from '@rstest/playwright';`, "expect"},
		{`const { expect: check } = import.meta.rstest;`, "check"},
		{"", "import.meta.rstest.expect"},
	} {
		add(pair[0]+"\n"+pair[1]+`(fn.mock.calls).toHaveLength(1);`,
			pair[0]+"\n"+pair[1]+`(fn).toHaveBeenCalledTimes(1);`)
	}
	// Test-context expect, including parameterized callbacks.
	add(`test('counts', ({ expect }) => { expect(fn.mock.calls).toHaveLength(1); });`,
		`test('counts', ({ expect }) => { expect(fn).toHaveBeenCalledTimes(1); });`)
	for _, code := range []string{
		`test.concurrent('counts', ({ expect }) => expect(fn.mock.calls).toHaveLength(1));`,
		`test.for([1])('counts', (row, ctx) => ctx.expect(fn.mock.calls).toHaveLength(1));`,
	} {
		add(code, "")
	}

	// Comments between the removed accessors survive the fix.
	add(`expect(fn /* keep */ .mock.calls).toHaveLength(1);`, `expect(fn /* keep */ ).toHaveBeenCalledTimes(1);`)
	add(`expect(fn.mock /* keep */ .calls).toHaveLength(1);`, `expect(fn /* keep */ ).toHaveBeenCalledTimes(1);`)
	// expect.poll re-invokes its argument, so rewriting the subject would call the mock.
	add(`await expect.poll(fn.mock.calls).toHaveLength(1);`, "")
	add(`await expect.poll(fn.mock.calls, { timeout: 10 }).not.toHaveLength(1);`, "")
	// The rewritten subject stays on the Chai assertion the chain returns.
	for _, code := range []string{
		`expect(fn.mock.calls).toHaveLength(1).toContainEqual([1]);`,
		`expect(fn.mock.calls).toHaveLength(1).and.called;`,
		`expect(fn.mock.calls).toHaveLength(1).message;`,
		`const assertion = expect(fn.mock.calls).toHaveLength(1);`,
		`const assertion = (expect(fn.mock.calls).toHaveLength(1));`,
		`const assertion = expect(fn.mock.calls).toHaveLength(1) as never;`,
		`assertion = expect(fn.mock.calls).toHaveLength(1);`,
		`function assertion() { return expect(fn.mock.calls).toHaveLength(1); }`,
		`const assertion = () => expect(fn.mock.calls).toHaveLength(1);`,
		`consume(expect(fn.mock.calls).toHaveLength(1));`,
		`const assertions = [expect(fn.mock.calls).toHaveLength(1)];`,
		`consume(condition ? expect(fn.mock.calls).toHaveLength(1) : other);`,
		`consume((sideEffect(), expect(fn.mock.calls).toHaveLength(1)));`,
	} {
		add(code, "")
	}
	// Both matchers of a chain are reported; neither can be rewritten.
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: `expect(fn.mock.calls).toHaveLength(1).toHaveLength(2);`,
		Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "preferMatcher", Line: 1, Column: 23, EndLine: 1, EndColumn: 35},
			{MessageId: "preferMatcher", Line: 1, Column: 39, EndLine: 1, EndColumn: 51},
		},
	})
	// Quoted matcher keys keep their delimiters; the report spans the whole literal.
	for _, pair := range [][2]string{
		{`expect(fn.mock.calls)['toHaveLength'](1);`, `expect(fn)['toHaveBeenCalledTimes'](1);`},
		{"expect(fn.mock.calls)[`toHaveLength`](1);", "expect(fn)[`toHaveBeenCalledTimes`](1);"},
		{`expect(fn.mock.calls)?.["toHaveLength"](1);`, `expect(fn)?.["toHaveBeenCalledTimes"](1);`},
	} {
		start := strings.Index(pair[0], "toHaveLength") - 1
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: pair[0], Output: []string{pair[1]},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferMatcher", Line: 1, Column: start + 1, EndLine: 1, EndColumn: start + 1 + len("toHaveLength") + 2},
			},
		})
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferToHaveBeenCalledTimesRule, valid, invalid)
}
