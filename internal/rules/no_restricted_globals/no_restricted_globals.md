# no-restricted-globals

## Rule Details

Disallows specified global variable names. This is useful when a project wants to allow globals in general but forbid specific ones — for example, banning the deprecated global `event` in favor of an explicit handler parameter, or banning ambiguous DOM globals like `name` or `length`.

Examples of **incorrect** code for this rule with `["error", "event", "fdescribe"]`:

```javascript
function onClick() {
  console.log(event);
}

fdescribe("foo", function () {});
```

Examples of **correct** code for this rule with `["error", "event"]`:

```javascript
import event from "event-module";

const event2 = 1;
```

The rule also accepts an object form so a custom message can be attached to each restricted name:

```json
{ "no-restricted-globals": ["error", { "name": "event", "message": "Use the local event parameter instead." }] }
```

```javascript
function onClick() {
  console.log(event);
}
```

## Options

The rule accepts either an array of names/objects, or a single object with the following properties:

- `globals` — array of restricted names; each entry is either a string or `{ name, message? }`.
- `checkGlobalObject` — when `true`, also flags access through a global object (`window.foo`, `self.foo`, `globalThis.foo`, and any names configured via `globalObjects`). Defaults to `false`.
- `globalObjects` — additional global object names to check when `checkGlobalObject` is enabled. `globalThis`, `self`, and `window` are always included.

Examples of **incorrect** code for this rule with `{ "globals": ["Promise"], "checkGlobalObject": true }`:

```json
{ "no-restricted-globals": ["error", { "globals": ["Promise"], "checkGlobalObject": true }] }
```

```javascript
globalThis.Promise;
self.Promise;
window.Promise;
```

Examples of **incorrect** code for this rule with a custom `globalObjects` entry:

```json
{
  "no-restricted-globals": [
    "error",
    { "globals": ["Promise"], "checkGlobalObject": true, "globalObjects": ["myGlobal"] }
  ]
}
```

```javascript
myGlobal.Promise;
```

## Differences from ESLint

- When `checkGlobalObject` is enabled, rslint always treats `globalThis`, `self`, `window`, and any configured `globalObjects` as accessible global objects — it doesn't require them to be declared through an ESLint environment or `languageOptions.globals`. As a result, code like `window.foo()` is flagged as soon as `foo` is restricted and `window` isn't shadowed locally, even in projects that never configure a browser environment.

## Original Documentation

- [ESLint: no-restricted-globals](https://eslint.org/docs/latest/rules/no-restricted-globals)
