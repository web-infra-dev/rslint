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

- ESLint stops tracking the global `Object` variable for the rest of the file once it is reassigned anywhere without a new declaration (e.g. `Object = {};`), and no longer reports any later `Object.assign(...)` call. rslint only recognizes shadowing introduced by a new declaration (`var`/`let`/`const`/`import`/a parameter/etc.); a bare reassignment like `Object = {};` does not suppress later reports.
- rslint always recognizes `globalThis.Object.assign(...)`, so it reports and fixes it the same as `Object.assign(...)`. ESLint only does this when `globalThis` is a declared global (e.g. under a sufficiently recent `ecmaVersion`) — under older configurations `globalThis.Object.assign(...)` is left untouched.

## Original Documentation

[prefer-object-spread](https://eslint.org/docs/latest/rules/prefer-object-spread)
