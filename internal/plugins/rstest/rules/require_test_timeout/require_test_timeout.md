# require-test-timeout

## Rule Details

Requires every test to run under a timeout that is written down somewhere, rather than falling back to the project-wide default. Stating the budget next to the test makes a slow test a visible decision instead of a silent one, and keeps a hang from consuming the default five seconds before anyone notices.

A test satisfies the rule in any of three ways, matching how Rstest itself resolves a timeout:

- **Its own timeout.** Either overload counts — the trailing number in `test(name, fn, 5000)` and the `timeout` property in `test(name, { timeout: 5000 }, fn)`. `0` disables the timeout in Rstest and so counts as a decision; a negative value does not.
- **A timeout on an enclosing suite.** Rstest applies suite options as defaults to the tests inside them, so `describe('slow', { timeout: 30_000 }, ...)` covers every test in that suite, however deeply nested.
- **A configured runtime timeout.** `rs.setConfig({ testTimeout: 30_000 })` covers the tests registered after it. A later `resetConfig()` on the same binding puts the default back, and the tests after that are reported again. The utility object is recognized as `rs`, `rstest`, `import.meta.rstest`, a namespace member, and a `require` binding, under whatever local name it was imported as.

Registrations with nothing to time out are left alone: `test.todo(...)`, `test.skip(...)`, and a registration with no callback at all — the last of those is `rstest/prefer-todo`'s subject. Parameterized registrations are included, and so are tests whose callback is a named function rather than an inline one.

The rule reports only what it can read. A timeout that arrives through an unreadable expression — an imported constant, a function call, a spread into the options object, a `let` binding — is taken at face value as a timeout the author wrote down, and the test is not reported. The same goes for a test inside a function this rule cannot tie back to a `describe`, such as one handed to `describe` by name: it could belong to a timed suite, so it is left alone.

## Incorrect

```ts
test('imports the catalog', async () => {
  await importCatalog(fixture);
});

describe('catalog', () => {
  it('exports the catalog', async () => {
    await exportCatalog();
  });
});
```

## Correct

```ts
test('imports the catalog', async () => {
  await importCatalog(fixture);
}, 30_000);

describe('catalog', { timeout: 30_000 }, () => {
  it('exports the catalog', async () => {
    await exportCatalog();
  });
});
```

```ts
rs.setConfig({ testTimeout: 30_000 });

test('imports the catalog', async () => {
  await importCatalog(fixture);
});
```

## Differences from the Vitest rule

Rstest resolves a test's timeout differently from Vitest, and the two Rstest-only steps are modeled here. A `timeout` declared on an enclosing `describe` exempts the tests inside it, because the Rstest runner applies suite options as defaults. A `resetConfig()` positioned between a `setConfig({ testTimeout })` and a test cancels the exemption that `setConfig` granted. The upstream rule models neither, because Vitest's rule looks no further than the registration and the `setConfig` call.

The upstream rule reports a timeout it cannot resolve to a number — `{ timeout: 'soon' }`, `{ ...options }`, a `let` binding, a value returned by a function. This rule stays silent on all of them. TypeScript already rejects a `timeout` that is not a number, so a lint report there adds nothing but the risk of flagging a test whose timeout is perfectly real and merely written somewhere this rule cannot follow.

A call with more arguments than either Rstest overload accepts matches neither one, and is left alone rather than scanned argument by argument for a timeout as upstream does.

Tests whose callback is passed by name are reported. Upstream looks for an inline function argument and so skips them, though such a test needs a timeout no less than an inline one does.

`xit` and the other `x`-prefixed aliases have no equivalent branch here, because Rstest does not export them.

This rule offers no autofix. What a test's timeout should be is a decision about the code under test, not something a linter can supply.
