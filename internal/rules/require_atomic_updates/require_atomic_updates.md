# require-atomic-updates

## Rule Details

Disallow assignments that can lead to race conditions due to usage of `await` or `yield`.

This rule reports assignments to variables or properties in cases where the assignments may be based on outdated values. When a variable is read, then an `await` or `yield` pauses execution, the variable might be modified by another concurrent operation before the assignment completes.

Examples of **incorrect** code for this rule:

```javascript
let result;

async function foo() {
  result += await something;
}

async function bar() {
  result = result + (await something);
}

function* baz() {
  result += yield;
}
```

Examples of **correct** code for this rule:

```javascript
let result;

async function foo() {
  result = (await something) + result;
}

async function bar() {
  const tmp = await something;
  result += tmp;
}

async function baz() {
  let localVar = 0;
  localVar += await something;
}
```

## Options

### `allowProperties`

When set to `true`, the rule does not report assignments to properties (only variables).

```javascript
/* eslint require-atomic-updates: ["error", { "allowProperties": true }] */

async function foo(obj) {
  if (!obj.done) {
    obj.something = await getSomething(); // OK with allowProperties
  }
}
```

## Original Documentation

[ESLint - require-atomic-updates](https://eslint.org/docs/latest/rules/require-atomic-updates)
