# prefer-object-spread

## Rule Details

When `Object.assign` is called using an object literal as the first argument, this rule requires using the object spread syntax instead. This rule also warns on cases where an `Object.assign` call is made using a single argument that is an object literal, in this case, the `Object.assign` call is not needed.

Object spread is a declarative alternative which may perform better than the more dynamic, imperative `Object.assign`.

Examples of **incorrect** code for this rule:

```javascript
Object.assign({}, foo);

Object.assign({}, { foo: 'bar' });

Object.assign({ foo: 'bar' }, baz);

Object.assign({}, baz, { foo: 'bar' });

Object.assign({}, { ...baz });

// Object.assign with a single argument that is an object literal
Object.assign({});

Object.assign({ foo: bar });
```

Examples of **correct** code for this rule:

```javascript
({ ...foo });

({ ...baz, foo: 'bar' });

// Any Object.assign call without an object literal as the first argument
Object.assign(foo, { bar: baz });

Object.assign(foo, bar);

Object.assign(foo, { bar, baz });

Object.assign(foo, { ...baz });
```

## Options

This rule has no options.

## Differences from ESLint

- ESLint stops tracking the global `Object` variable for the whole file once it is reassigned anywhere without a new declaration (e.g. `Object = {};`), including calls textually before the reassignment. rslint tracks this flow-sensitively: calls before the first bare write to `Object` (and aliases captured before it) are still reported; only references after the write are untracked.
- rslint follows aliases flow-sensitively and reports calls that ESLint's ReferenceTracker misses: aliases established by a plain assignment (`let o; o = Object; o.assign({}, x)`), nested destructuring (`const { Object: { assign } } = globalThis; assign({}, x)`), and alias chains of any length. Conversely, an alias that has been reassigned to something else by the time of the call (`let o = Object; o = foo; o.assign({}, x)`) is not reported.
- When a source argument is an object literal with a prototype-setting `__proto__:` property, the autofix keeps that literal whole behind a spread (`Object.assign({}, { __proto__: p })` → `({ ...{ __proto__: p } })`) instead of merging its properties into the result literal, where `__proto__:` would change the result's prototype. ESLint's fixer merges it and changes behavior.
- The autofix parenthesizes the resulting object literal when the call is the callee of another call (`Object.assign({}, foo)()` → `({ ...foo})()`); ESLint's fixer produces unparsable output there.
- rslint always recognizes `globalThis.Object.assign(...)`, so it reports and fixes it the same as `Object.assign(...)`. ESLint only does this when `globalThis` is a declared global (e.g. under a sufficiently recent `ecmaVersion`) — under older configurations `globalThis.Object.assign(...)` is left untouched.

## Original Documentation

[prefer-object-spread](https://eslint.org/docs/latest/rules/prefer-object-spread)
