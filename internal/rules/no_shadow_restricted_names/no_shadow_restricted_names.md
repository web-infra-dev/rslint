# no-shadow-restricted-names

## Rule Details

ECMAScript defines several special names that should not be redefined by user code: `NaN`, `Infinity`, `undefined`, `eval`, `arguments`, and `globalThis`. Shadowing these restricted names obscures runtime globals and makes programs harder to reason about.

This rule disallows shadowing of these restricted names by variable declarations, function names, function parameters, catch clause parameters, imported bindings, and class names.

Examples of **incorrect** code for this rule:

```javascript
function NaN() {}

!function (Infinity) {};

var undefined = 5;

try {} catch (eval) {}

class globalThis {}

import undefined from "foo";
```

Examples of **correct** code for this rule:

```javascript
var Object;

function f(a, b) {}

// A declaration that doesn't assign a value to `undefined` is safe:
var undefined;
```

### Options

This rule has an object option:

- `"reportGlobalThis": true` (default) — report shadowing of `globalThis`.
- `"reportGlobalThis": false` — allow shadowing `globalThis`.

Examples of **correct** code for this rule with `{ "reportGlobalThis": false }`:

```json
{ "no-shadow-restricted-names": ["error", { "reportGlobalThis": false }] }
```

```javascript
let globalThis;

class globalThis {}

import { baz as globalThis } from "foo";
```

## Original Documentation

- [ESLint no-shadow-restricted-names](https://eslint.org/docs/latest/rules/no-shadow-restricted-names)
