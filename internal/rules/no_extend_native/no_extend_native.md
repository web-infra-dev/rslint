# no-extend-native

Disallows directly modifying the prototype of native built-in objects (`Object`,
`Array`, `Function`, `String`, `Number`, `Boolean`, `Symbol`, `Map`, `Set`,
`Promise`, `Error`, `RegExp`, `Date`, `BigInt`, `WeakRef`,
`FinalizationRegistry`, etc.).

## Rule Details

Extending native prototypes is generally regarded as a bad practice because it
breaks assumptions other code makes about builtins, can collide with future
language additions, and is invisible to the rest of the program until it is
triggered at runtime.

The rule reports two extension patterns:

1. Direct assignment, including compound and logical assignments:
   `Builtin.prototype.foo = ...`, `Builtin.prototype.foo ??= ...`.
2. `Object.defineProperty(Builtin.prototype, ...)` and
   `Object.defineProperties(Builtin.prototype, ...)`.

References to the builtin that are shadowed by a local declaration in scope
(e.g. `function foo() { var Object = function () {}; Object.prototype.p = 0 }`)
are not reported.

Examples of **incorrect** code for this rule:

```javascript
Object.prototype.a = "a";
Object.defineProperty(Array.prototype, "times", { value: 999 });
```

Examples of **correct** code for this rule:

```javascript
// Modifications to user-defined objects are allowed.
x.prototype.p = 0;

// Property access on the constructor (not on its prototype) is allowed.
Object.toString.bind = 0;
```

## Options

```json
{
  "no-extend-native": ["error", { "exceptions": ["Object"] }]
}
```

| Option       | Type                | Default | Description                                                                |
| ------------ | ------------------- | ------- | -------------------------------------------------------------------------- |
| `exceptions` | `string[]` (unique) | `[]`    | Names of built-in objects whose prototype is allowed to be extended.       |

With `{ "exceptions": ["Object"] }`, the following becomes valid:

```javascript
Object.prototype.g = 0;
```

## Original Documentation

- [ESLint rule documentation](https://eslint.org/docs/latest/rules/no-extend-native)
- [Source code](https://github.com/eslint/eslint/blob/main/lib/rules/no-extend-native.js)
