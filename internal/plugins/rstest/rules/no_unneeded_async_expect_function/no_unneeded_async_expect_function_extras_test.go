package no_unneeded_async_expect_function

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnneededAsyncExpectFunctionExtras(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	addValid := func(code ...string) {
		for _, item := range code {
			valid = append(valid, rule_tester.ValidTestCase{Code: item})
		}
	}

	var invalid []rule_tester.InvalidTestCase
	// add reports the diagnostic on wrapper, which must appear once in code.
	// An empty output records a diagnostic that carries no suggestion.
	add := func(code, wrapper, output string) {
		t.Helper()
		start := strings.Index(code, wrapper)
		if start < 0 || start != strings.LastIndex(code, wrapper) {
			t.Fatalf("wrapper %q must appear exactly once in %q", wrapper, code)
		}
		end := start + len(wrapper)
		line := strings.Count(code[:start], "\n") + 1
		column := start - strings.LastIndex(code[:start], "\n")
		item := rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "noAsyncWrapperForExpectedPromise",
			Line:      line,
			Column:    column,
			EndLine:   strings.Count(code[:end], "\n") + 1,
			EndColumn: end - strings.LastIndex(code[:end], "\n"),
		}}}
		if output != "" {
			item.Errors[0].Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "suggestRemoveAsyncWrapper",
				Output:    output,
			}}
		}
		invalid = append(invalid, item)
	}

	// Without `resolves` or `rejects` the assertion is about the function
	// itself, and unwrapping it would assert on a promise instead.
	addValid(
		`expect(async () => { await run(); }).toBeInstanceOf(Function);`,
		`expect(async () => { await run(); }).toThrow();`,
		`expect(async () => { await run(); }).not.toBeUndefined();`,
		`expect(async () => { await run(); }).toBe(handler);`,
	)
	// Chains that never reach a matcher are not assertions.
	addValid(
		`expect(async () => { await run(); });`,
		`expect(async () => { await run(); }).rejects;`,
		`expect(async () => { await run(); }).rejects.toThrow;`,
	)
	// expect.poll() retries a callback and throws when combined with a promise
	// modifier, and expect.element() asserts on a browser locator.
	addValid(
		`expect.poll(async () => { await run(); }).rejects.toThrow();`,
		`expect.element(async () => { await run(); }).rejects.toThrow();`,
	)
	// An async generator returns an async iterator, not a promise, so the call
	// it wraps is not what the assertion would receive.
	addValid(
		`expect(async function* () { await run(); }).rejects.toThrow();`,
	)
	// The wrapper has to be async, hold exactly one awaited call, and await a
	// call rather than any other expression.
	addValid(
		`expect(() => { run(); }).rejects.toThrow();`,
		`expect(async () => { run(); }).rejects.toThrow();`,
		`expect(async () => { await run(); await stop(); }).rejects.toThrow();`,
		`expect(async () => { const value = await run(); }).rejects.toThrow();`,
		`expect(async () => { return await run(); }).rejects.toThrow();`,
		`expect(async () => { await pending; }).rejects.toThrow();`,
		`expect(async () => { await new Runner(); }).rejects.toThrow();`,
		`expect(async () => [await run()]).rejects.toThrow();`,
		`expect(handler, async () => { await run(); }).rejects.toThrow();`,
		`expect().rejects.toThrow();`,
	)
	// A local binding named expect is not the Rstest assertion factory.
	addValid(`
function expect(value) {
  return { rejects: { toThrow() {} } };
}
expect(async () => { await run(); }).rejects.toThrow();
`)

	// Every assertion factory that asserts on the value it receives, every way
	// of reaching expect, and both promise modifiers.
	add(
		`expect.soft(async () => { await run(); }).rejects.toThrow();`,
		`async () => { await run(); }`,
		`expect.soft(run()).rejects.toThrow();`,
	)
	add(
		`expect(async () => { await run(); }).resolves.toBe(1);`,
		`async () => { await run(); }`,
		`expect(run()).resolves.toBe(1);`,
	)
	add(
		`expect(async () => { await run(); }).rejects.not.toThrow();`,
		`async () => { await run(); }`,
		`expect(run()).rejects.not.toThrow();`,
	)
	add(`
import { expect as pleaseExpect } from '@rstest/core';
pleaseExpect(async () => { await run(); }).rejects.toThrow();
`,
		`async () => { await run(); }`,
		`
import { expect as pleaseExpect } from '@rstest/core';
pleaseExpect(run()).rejects.toThrow();
`)
	add(`
import * as rstest from '@rstest/core';
rstest.expect(async () => { await run(); }).rejects.toThrow();
`,
		`async () => { await run(); }`,
		`
import * as rstest from '@rstest/core';
rstest.expect(run()).rejects.toThrow();
`)
	add(`
const { expect } = require('@rstest/core');
expect(async () => { await run(); }).rejects.toThrow();
`,
		`async () => { await run(); }`,
		`
const { expect } = require('@rstest/core');
expect(run()).rejects.toThrow();
`)
	add(`
import { expect } from 'rstack/test';
expect(async () => { await run(); }).rejects.toThrow();
`,
		`async () => { await run(); }`,
		`
import { expect } from 'rstack/test';
expect(run()).rejects.toThrow();
`)
	add(`
import { test } from '@rstest/core';
test('reports', async ({ expect }) => {
  await expect(async () => { await run(); }).rejects.toThrow();
});
`,
		`async () => { await run(); }`,
		`
import { test } from '@rstest/core';
test('reports', async ({ expect }) => {
  await expect(run()).rejects.toThrow();
});
`)
	add(`
import { expect } from '@rstest/playwright';
expect(async () => { await run(); }).rejects.toThrow();
`,
		`async () => { await run(); }`,
		`
import { expect } from '@rstest/playwright';
expect(run()).rejects.toThrow();
`)
	add(
		`import.meta.rstest.expect(async () => { await run(); }).rejects.toThrow();`,
		`async () => { await run(); }`,
		`import.meta.rstest.expect(run()).rejects.toThrow();`,
	)
	// A non-null assertion on the optional import.meta.rstest namespace is not
	// recognized as an Rstest binding by the shared expect parser, so the
	// assertion behind it is left alone.
	addValid(`import.meta.rstest!.expect(async () => { await run(); }).rejects.toThrow();`)
	// A modifier written as a string key reads the same property.
	add(
		`expect(async () => { await run(); })["rejects"].toThrow();`,
		`async () => { await run(); }`,
		`expect(run())["rejects"].toThrow();`,
	)
	// An optional call on expect itself still hands the wrapper to rejects.
	add(
		`expect?.(async () => { await run(); }).rejects.toThrow();`,
		`async () => { await run(); }`,
		`expect?.(run()).rejects.toThrow();`,
	)
	// A later matcher in the chain asserts on the same promise either way.
	add(
		`expect(async () => { await run(); }).rejects.toThrow().toBe(1);`,
		`async () => { await run(); }`,
		`expect(run()).rejects.toThrow().toBe(1);`,
	)

	// The awaited call keeps its own spelling, and the wrapper's parentheses
	// and the remaining expect arguments are preserved.
	add(
		`expect(async () => { await run?.(); }).rejects.toThrow();`,
		`async () => { await run?.(); }`,
		`expect(run?.()).rejects.toThrow();`,
	)
	add(
		`expect(async () => { await (0, run)(); }).rejects.toThrow();`,
		`async () => { await (0, run)(); }`,
		`expect((0, run)()).rejects.toThrow();`,
	)
	add(
		`expect((async () => { await run(); })).rejects.toThrow();`,
		`(async () => { await run(); })`,
		`expect(run()).rejects.toThrow();`,
	)
	add(
		`expect(async () => { await run(); }, 'runs').rejects.toThrow();`,
		`async () => { await run(); }`,
		`expect(run(), 'runs').rejects.toThrow();`,
	)
	add(
		`expect(async () => { await run/* keeps its own comments */(); }).rejects.toThrow();`,
		`async () => { await run/* keeps its own comments */(); }`,
		`expect(run/* keeps its own comments */()).rejects.toThrow();`,
	)

	// Bindings the wrapper owns do not survive the unwrap, so the diagnostic
	// stands without a suggestion.
	add(
		`expect(async (value) => { await run(value); }).rejects.toThrow();`,
		`async (value) => { await run(value); }`,
		"",
	)
	add(
		`expect(async <T,>() => { await run<T>(); }).rejects.toThrow();`,
		`async <T,>() => { await run<T>(); }`,
		"",
	)
	add(
		`expect(async function retry() { await retry(); }).rejects.toThrow();`,
		`async function retry() { await retry(); }`,
		"",
	)
	add(
		`expect(async function () { await this.run(); }).rejects.toThrow();`,
		`async function () { await this.run(); }`,
		"",
	)
	add(
		`expect(async function () { await run(arguments); }).rejects.toThrow();`,
		`async function () { await run(arguments); }`,
		"",
	)
	add(
		`expect(async function () { await run(new.target); }).rejects.toThrow();`,
		`async function () { await run(new.target); }`,
		"",
	)
	// An arrow takes `this`, `arguments` and `new.target` from the enclosing
	// scope already.
	add(
		`expect(async () => { await this.run(); }).rejects.toThrow();`,
		`async () => { await this.run(); }`,
		`expect(this.run()).rejects.toThrow();`,
	)
	add(
		`function outer() { return expect(async () => { await run(new.target); }).rejects.toThrow(); }`,
		`async () => { await run(new.target); }`,
		`function outer() { return expect(run(new.target)).rejects.toThrow(); }`,
	)

	// An await inside the awaited call belongs to the wrapper, and would move
	// into the assertion's scope with the call.
	add(
		"test('rejects', () =>\n  expect(async () => { await run(await load()); }).rejects.toThrow(),\n);",
		`async () => { await run(await load()); }`,
		"",
	)
	add(
		`test('rejects', async () => { await expect(async () => { await run(await load()); }).rejects.toThrow(); });`,
		`async () => { await run(await load()); }`,
		"",
	)
	add(
		`expect(async () => await run(await load())).rejects.toThrow();`,
		`async () => await run(await load())`,
		"",
	)
	add(
		`expect(async () => { await run({ async [await key()]() {} }); }).rejects.toThrow();`,
		`async () => { await run({ async [await key()]() {} }); }`,
		"",
	)
	// An await inside a nested function belongs to that function and moves
	// with it.
	add(
		`expect(async () => { await run(async () => await load()); }).rejects.toThrow();`,
		`async () => { await run(async () => await load()); }`,
		`expect(run(async () => await load())).rejects.toThrow();`,
	)
	add(
		`expect(async () => { await run({ async load() { await fetch(); } }); }).rejects.toThrow();`,
		`async () => { await run({ async load() { await fetch(); } }); }`,
		`expect(run({ async load() { await fetch(); } })).rejects.toThrow();`,
	)
	// The type argument describes the wrapper, not the awaited value.
	add(
		`expect<() => Promise<void>>(async () => { await run(); }).rejects.toThrow();`,
		`async () => { await run(); }`,
		"",
	)
	// A comment anywhere else inside the wrapper would be deleted.
	add(`
expect(async () => {
  // waits for the queue to drain
  await run();
}).rejects.toThrow();
`,
		`async () => {
  // waits for the queue to drain
  await run();
}`,
		"",
	)
	add(`
expect(async () => {
  await run();
  // nothing else happens
}).rejects.toThrow();
`,
		`async () => {
  await run();
  // nothing else happens
}`,
		"",
	)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnneededAsyncExpectFunctionRule, valid, invalid)
}
