# func-name-matching

## Rule Details

This rule requires function names to match the name of the variable or property to which they are assigned. The rule will ignore property assignments where the property name is a literal that is not a valid identifier in the ECMAScript version specified in your configuration (default ES5).

Examples of **incorrect** code for this rule:

```javascript
let foo = function bar() {};
foo = function bar() {};
const obj = { foo: function bar() {} };
obj.foo = function bar() {};
obj["foo"] = function bar() {};
({ ["foo"]: function bar() {} });

class C {
  foo = function bar() {};
}
```

Examples of **incorrect** code for this rule with `"never"`:

```json
{ "func-name-matching": ["error", "never"] }
```

```javascript
let foo = function foo() {};
foo = function foo() {};
const obj = { foo: function foo() {} };
obj.foo = function foo() {};
obj["foo"] = function foo() {};
({ ["foo"]: function foo() {} });

class C {
  foo = function foo() {};
}
```

Examples of **correct** code for this rule:

```javascript
const foo = function foo() {};
const foo1 = function () {};
const foo2 = () => {};
foo = function foo() {};

const obj = { foo: function foo() {} };
obj.foo = function foo() {};
obj["foo"] = function foo() {};
obj["foo//bar"] = function foo() {};
obj[foo] = function bar() {};

const obj1 = { [foo]: function bar() {} };
const obj2 = { "foo//bar": function foo() {} };
const obj3 = { foo: function () {} };

obj["x" + 2] = function bar() {};
const [bar] = [function bar() {}];
({ [foo]: function bar() {} });

class C {
  foo = function foo() {};
  baz = function () {};
}

// private names are ignored
class D {
  #foo = function foo() {};
  #bar = function foo() {};
  baz() {
    this.#foo = function foo() {};
    this.#foo = function bar() {};
  }
}

module.exports = function foo(name) {};
module["exports"] = function foo(name) {};
```

Examples of **correct** code for this rule with `"never"`:

```json
{ "func-name-matching": ["error", "never"] }
```

```javascript
let foo = function bar() {};
const foo1 = function () {};
const foo2 = () => {};
foo = function bar() {};

const obj = { foo: function bar() {} };
obj.foo = function bar() {};
obj["foo"] = function bar() {};
obj["foo//bar"] = function foo() {};
obj[foo] = function foo() {};

const obj1 = { foo: function bar() {} };
const obj2 = { [foo]: function foo() {} };
const obj3 = { "foo//bar": function foo() {} };
const obj4 = { foo: function () {} };

obj["x" + 2] = function bar() {};
const [bar] = [function bar() {}];
({ [foo]: function bar() {} });

class C {
  foo = function bar() {};
  baz = function () {};
}

// private names are ignored
class D {
  #foo = function foo() {};
  #bar = function foo() {};
  baz() {
    this.#foo = function foo() {};
    this.#foo = function bar() {};
  }
}

module.exports = function foo(name) {};
module["exports"] = function foo(name) {};
```

## Options

This rule takes an optional string of `"always"` or `"never"` (when omitted, it defaults to `"always"`), and an optional options object with two properties `considerPropertyDescriptor` and `includeCommonJSModuleExports`.

### considerPropertyDescriptor

A boolean value that defaults to `false`. If `considerPropertyDescriptor` is set to true, the check will take into account the use of `Object.create`, `Object.defineProperty`, `Object.defineProperties`, and `Reflect.defineProperty`.

Examples of **correct** code for the `{ "considerPropertyDescriptor": true }` option:

```json
{ "func-name-matching": ["error", { "considerPropertyDescriptor": true }] }
```

```javascript
const obj = {};
Object.create(obj, { foo: { value: function foo() {} } });
Object.defineProperty(obj, "bar", { value: function bar() {} });
Object.defineProperties(obj, { baz: { value: function baz() {} } });
Reflect.defineProperty(obj, "foo", { value: function foo() {} });
```

Examples of **incorrect** code for the `{ "considerPropertyDescriptor": true }` option:

```json
{ "func-name-matching": ["error", { "considerPropertyDescriptor": true }] }
```

```javascript
const obj = {};
Object.create(obj, { foo: { value: function bar() {} } });
Object.defineProperty(obj, "bar", { value: function baz() {} });
Object.defineProperties(obj, { baz: { value: function foo() {} } });
Reflect.defineProperty(obj, "foo", { value: function value() {} });
```

### includeCommonJSModuleExports

A boolean value that defaults to `false`. If `includeCommonJSModuleExports` is set to true, `module.exports` and `module["exports"]` will be checked by this rule.

Examples of **incorrect** code for the `{ "includeCommonJSModuleExports": true }` option:

```json
{ "func-name-matching": ["error", { "includeCommonJSModuleExports": true }] }
```

```javascript
module.exports = function foo(name) {};
module["exports"] = function foo(name) {};
```

## Differences from ESLint

- For a property whose name comes from a string literal (e.g. `{ "ᢅ": function foo() {} }`), rslint decides whether that name is a valid identifier using the same character rules at every configured ECMAScript version. ESLint's `ecmaVersion: 5` mode instead uses an older, frozen table of valid identifier characters, so a handful of characters that were only added to the identifier rules in later Unicode versions are treated as invalid identifiers by ESLint at `ecmaVersion: 5` — and the property is left unchecked — while rslint treats them as valid identifiers and checks the property at every ECMAScript version, including 5.

- With `considerPropertyDescriptor` enabled, an `Object.defineProperties()` or `Object.create()` descriptor map is checked against its entry's key when that key is an identifier (e.g. `{ bar: { value: function bar() {} } }`). ESLint also reports entries keyed by a string or numeric literal (e.g. `{ "bar": { value: function baz() {} } }`), but names the property `undefined` in the message; rslint leaves those entries unchecked.

## Original Documentation

- [ESLint: func-name-matching](https://eslint.org/docs/latest/rules/func-name-matching)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/func-name-matching.js)
