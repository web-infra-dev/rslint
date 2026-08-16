# id-length

## Rule Details

Very short identifier names like `e`, `x`, `_t` or very long ones like `hashGeneratorResultOutputContainerObject` can make code harder to read and potentially less maintainable. This rule enforces a minimum and/or maximum identifier length convention.

This rule counts [graphemes](https://unicode.org/reports/tr29/#Default_Grapheme_Cluster_Table) instead of using UTF-16 code unit length.

Examples of **incorrect** code for this rule with the default options:

```javascript
const x = 5;
obj.e = document.body;
const foo = function (e) {};
try {
  dangerousStuff();
} catch (e) {
  // ignore as many do
}
const myObj = { a: 1 };
(a) => {
  a * a;
};
class y {}
class Foo {
  x() {}
}
function bar(...x) {}
function baz([x]) {}
const [z] = arr;
const {
  prop: [i],
} = {};
function qux({ x }) {}
const { j } = {};
const { prop: a } = {};
({ prop: obj.x } = {});
```

Examples of **correct** code for this rule with the default options:

```javascript
const num = 5;
function _f() {
  return 42;
}
obj.el = document.body;
const foo = function (evt) {
  /* do stuff */
};
try {
  dangerousStuff();
} catch (error) {
  // ignore as many do
}
const myObj = { apple: 1 };
(num) => {
  num * num;
};
function bar(num = 0) {}
class MyClass {}
class Foo {
  method() {}
}
function baz(...args) {}
function qux([longName]) {}
const { prop } = {};
const [longName] = arr;
function foobar({ prop }) {}
const { a: property } = {};
({ prop: obj.longName } = {});
const data = { x: 1 }; // excused because of quotes
data["y"] = 3; // excused because of calculated property access
```

## Options

This rule has an object option:

- `"min"` (default: `2`) enforces a minimum identifier length
- `"max"` (default: unlimited) enforces a maximum identifier length
- `"properties": "always"` (default) enforces identifier length convention for property names
- `"properties": "never"` ignores identifier length convention for property names
- `"exceptions"` allows an array of specified identifier names
- `"exceptionPatterns"` array of strings representing regular expression patterns, allows identifiers that match any of the patterns

### min

Examples of **incorrect** code for this rule with `{ "min": 4 }`:

```json
{ "id-length": ["error", { "min": 4 }] }
```

```javascript
const val = 5;
function foo(e) {}
```

### max

Examples of **incorrect** code for this rule with `{ "max": 10 }`:

```json
{ "id-length": ["error", { "max": 10 }] }
```

```javascript
const reallyLongVarName = 5;
function reallyLongFuncName() {
  return 42;
}
```

### properties

Examples of **correct** code for this rule with `{ "properties": "never" }`:

```json
{ "id-length": ["error", { "properties": "never" }] }
```

```javascript
const myObj = { a: 1 };
({ a: obj.x.y.z } = {});
```

### exceptions

Examples of additional **correct** code for this rule with `{ "exceptions": ["x", "y", "z"] }`:

```json
{ "id-length": ["error", { "exceptions": ["x", "y", "z"] }] }
```

```javascript
const x = 5;
function y() {
  return 42;
}
```

### exceptionPatterns

Examples of additional **correct** code for this rule with `{ "exceptionPatterns": ["^E|S$"] }`:

```json
{ "id-length": ["error", { "exceptionPatterns": ["^E|S$"] }] }
```

```javascript
const E = 5;
function S() {
  return 42;
}
```

## Differences from ESLint

- `class Foo extends B {}` does not flag `B`, even when too short/long: this rule only checks a class's own name, never its superclass expression.
- In a destructuring assignment (not a `var`/`let`/`const` declaration) where a property's key and value are written out identically, e.g. `({ a: a } = {})`, the key is not checked — only the value would be, and here it's the same text either way. `var { a: a } = {}` (a real declaration) is unaffected and still checks the key as usual.

## Original Documentation

- [ESLint: id-length](https://eslint.org/docs/latest/rules/id-length)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/id-length.js)
