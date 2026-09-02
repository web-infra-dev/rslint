# no-restricted-rstest-methods

## Rule Details

Disallow members of the Rstest utilities object that a team has decided not to use, each with an optional message naming what to reach for instead. Nothing is disallowed until the option object says so.

The utilities object is reached as `rs` or `rstest`, and its members come in two kinds that this rule matches differently, because Rstest reaches them differently. `fn`, `spyOn`, `mocked`, the timer helpers and the stub helpers are ordinary functions, so the rule follows the binding: a rename such as `import { rs as mocker }`, a namespace as in `core.rs.fn()`, `import.meta.rstest.rs.fn()`, a `require` and the [globals](https://rstest.rs/config/test/globals) are all reported, a receiver the file declares itself is a different object and is left alone, and every call shape counts, so `rs['fn']()`, `rs.fn?.()` and `rs?.fn()` are reported alongside `rs.fn()`.

The module mock APIs — `mock`, `doMock`, `unmock`, `hoisted`, `importActual`, `requireMock` and the rest — are rewritten by Rstest's build rather than called, so the rule reads the receiver as written instead of following the binding. A renamed binding is not reported, because Rstest does not rewrite it either and the call throws where it stands; a receiver the file declares itself is reported, because the build rewrites the call regardless of that declaration. Only the shapes the build actually rewrites are reported: `rs.mock('./payments')` is, while `rs['mock']('./payments')`, `rs.mock?.('./payments')` and `rs?.mock('./payments')` are not, since none of them runs. `importActual` and `requireActual` are the two the build also reads through a bracketed string and an optional call, so `rs['importActual']('./payments')` is reported.

The rule reports where a disallowed member is used. Where the call stands and what it is passed are not its concern.

## Incorrect

```ts
"rstest/no-restricted-rstest-methods": ["error", { "useFakeTimers": "Use the clock fixture instead." }]
```

```ts
test('charges the card after the retry window', async () => {
  rs.useFakeTimers();
  await retryCharge(card);
});
```

## Correct

```ts
test('charges the card after the retry window', async () => {
  const clock = installClock();
  await retryCharge(card);
});
```

## Options

```json
{
  "rstest/no-restricted-rstest-methods": [
    "error",
    {
      "useFakeTimers": "Use the clock fixture instead.",
      "stubGlobal": null
    }
  ]
}
```

The option object maps a member name to the message reported in its place, or to `null` for the default message. A name that is not a member of the utilities object never matches anything.

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| _(member name)_ | `string \| null` | `{}` | Message to report where this member is used, or `null` for the default message. |
