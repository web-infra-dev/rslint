# no-underscore-dangle

## Rule Details

This rule disallows dangling underscores in identifiers — an underscore at the beginning or the end of a name, such as `_foo` or `foo_`. A name that is exactly `_` is always allowed.

By default the rule checks variable declarations, function declaration names, and member access. Function parameters, destructured names, method names, and class field names are each governed by an option.

Examples of **incorrect** code for this rule:

```javascript
let foo_;
const __proto__ = {};
foo._bar();
```

Examples of **correct** code for this rule:

```javascript
const _ = require('underscore');
const obj = _.contains(items, item);
obj.__proto__ = {};
const file = __filename;
function foo(_bar) {}
const bar = { onClick(_bar) {} };
const baz = (_bar) => {};
```

## Options

This rule has an object option:

- `"allow"` — a list of identifiers that may have dangling underscores
- `"allowAfterThis": false` (default) — disallows dangling underscores in members of the `this` object
- `"allowAfterSuper": false` (default) — disallows dangling underscores in members of the `super` object
- `"allowAfterThisConstructor": false` (default) — disallows dangling underscores in members of the `this.constructor` object
- `"enforceInMethodNames": false` (default) — allows dangling underscores in method names
- `"enforceInClassFields": false` (default) — allows dangling underscores in class field names
- `"allowInArrayDestructuring": true` (default) — allows dangling underscores in names bound by array destructuring
- `"allowInObjectDestructuring": true` (default) — allows dangling underscores in names bound by object destructuring
- `"allowFunctionParams": true` (default) — allows dangling underscores in function parameter names

### allow

Examples of additional **correct** code for this rule with the `{ "allow": ["foo_", "_bar"] }` option:

```json
{ "no-underscore-dangle": ["error", { "allow": ["foo_", "_bar"] }] }
```

```javascript
let foo_;
foo._bar();
```

### allowAfterThis

Examples of **correct** code for this rule with the `{ "allowAfterThis": true }` option:

```json
{ "no-underscore-dangle": ["error", { "allowAfterThis": true }] }
```

```javascript
const a = this.foo_;
this._bar();
```

### allowAfterSuper

Examples of **correct** code for this rule with the `{ "allowAfterSuper": true }` option:

```json
{ "no-underscore-dangle": ["error", { "allowAfterSuper": true }] }
```

```javascript
class Foo extends Bar {
  doSomething() {
    const a = super.foo_;
    super._bar();
  }
}
```

### allowAfterThisConstructor

Examples of **correct** code for this rule with the `{ "allowAfterThisConstructor": true }` option:

```json
{ "no-underscore-dangle": ["error", { "allowAfterThisConstructor": true }] }
```

```javascript
const a = this.constructor.foo_;
this.constructor._bar();
```

### enforceInMethodNames

Examples of **incorrect** code for this rule with the `{ "enforceInMethodNames": true }` option:

```json
{ "no-underscore-dangle": ["error", { "enforceInMethodNames": true }] }
```

```javascript
class Foo {
  _bar() {}
}

class Bar {
  bar_() {}
}

const o1 = {
  _bar() {},
};

const o2 = {
  bar_() {},
};
```

### enforceInClassFields

Examples of **incorrect** code for this rule with the `{ "enforceInClassFields": true }` option:

```json
{ "no-underscore-dangle": ["error", { "enforceInClassFields": true }] }
```

```javascript
class Foo {
  _bar;
}

class Bar {
  _bar = () => {};
}

class Baz {
  bar_;
}

class Qux {
  #_bar;
}

class FooBar {
  #bar_;
}
```

### allowInArrayDestructuring

Examples of **incorrect** code for this rule with the `{ "allowInArrayDestructuring": false }` option:

```json
{ "no-underscore-dangle": ["error", { "allowInArrayDestructuring": false }] }
```

```javascript
const [_foo, _bar] = list;
const [foo_, ..._qux] = list;
const [foo, [bar, _baz]] = list;
```

### allowInObjectDestructuring

Examples of **incorrect** code for this rule with the `{ "allowInObjectDestructuring": false }` option:

```json
{ "no-underscore-dangle": ["error", { "allowInObjectDestructuring": false }] }
```

```javascript
const { foo, bar: _bar } = collection;
const { qux, xyz, _baz } = collection;
```

Examples of **correct** code for this rule with the `{ "allowInObjectDestructuring": false }` option:

```json
{ "no-underscore-dangle": ["error", { "allowInObjectDestructuring": false }] }
```

```javascript
const {
  foo,
  bar,
  _baz: { a, b },
} = collection;
const { qux, xyz, _baz: baz } = collection;
```

### allowFunctionParams

Examples of **incorrect** code for this rule with the `{ "allowFunctionParams": false }` option:

```json
{ "no-underscore-dangle": ["error", { "allowFunctionParams": false }] }
```

```javascript
function foo1(_bar) {}
function foo2(_bar = 0) {}
function foo3(..._bar) {}

const foo4 = function onClick(_bar) {};
const foo5 = function onClick(_bar = 0) {};
const foo6 = function onClick(..._bar) {};

const foo7 = (_bar) => {};
const foo8 = (_bar = 0) => {};
const foo9 = (..._bar) => {};
```

## TypeScript

A member that only declares a signature has no implementation to name, so it is never reported: overload signatures, `declare function`, and `abstract` class members. The implementation that follows is checked as usual.

```json
{
  "no-underscore-dangle": [
    "error",
    { "enforceInMethodNames": true, "enforceInClassFields": true }
  ]
}
```

```typescript
declare function _read(): void;

abstract class Store {
  abstract _load(): void;
  abstract _cache: Map<string, string>;
}
```

A constructor parameter that also declares a property — one marked `private`, `protected`, `public`, `readonly`, or `override` — is exempt from `allowFunctionParams`.

```json
{ "no-underscore-dangle": ["error", { "allowFunctionParams": false }] }
```

```typescript
class Service {
  constructor(private readonly _http: Http) {}
}
```

## When Not To Use It

If you want to allow dangling underscores in identifiers, then you can safely turn this rule off.

## Original Documentation

- [ESLint: no-underscore-dangle](https://eslint.org/docs/latest/rules/no-underscore-dangle)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-underscore-dangle.js)
