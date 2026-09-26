# no-global-object-property-assignment

Disallow assigning properties on the global object.

Global object mutation makes state harder to trace and can overwrite existing globals. Prefer module scope, explicit imports/exports, dependency injection, or a local singleton.

```js
// Incorrect
globalThis.foo = value;
window.foo += 1;
self.foo ||= value;

// Correct
export const foo = value;
const singleton = { foo: value };
globalThis.foo;
```

The rule checks statically named properties on the effective global `global`, `globalThis`, `self`, and `window` bindings. Shadowed bindings and dynamic computed properties are ignored. Reads and `delete` expressions are also allowed.

This port follows eslint-plugin-unicorn v75.0.0.

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/docs/rules/no-global-object-property-assignment.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-global-object-property-assignment.js)
