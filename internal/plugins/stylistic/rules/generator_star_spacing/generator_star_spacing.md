# generator-star-spacing

Enforce consistent spacing around `*` operators in generator functions.

## Rule Details

Generators are a type of function in ECMAScript 6 that can return multiple values over time. These functions are indicated by placing an `*` after the `function` keyword.

To keep a sense of consistency when using generators this rule standardizes the spacing around the `*` token.

Examples of **incorrect** code for this rule with the default `"before"` option:

```javascript
function*foo() {}
function* foo() {}
var foo = function*(){};
var foo = { * bar(){} };
class Foo { * bar(){} }
class Foo { static* bar(){} }
```

Examples of **correct** code for this rule with the default `"before"` option:

```javascript
function *foo() {}
var foo = function *(){};
var foo = { *bar(){} };
class Foo { *bar(){} }
class Foo { static *bar(){} }
```

## Options

This rule takes one option, which can be a string or an object.

String shorthands:

- `"before"` (default) requires a space before the `*` and forbids a space after — equivalent to `{ "before": true, "after": false }`.
- `"after"` requires a space after the `*` and forbids a space before — equivalent to `{ "before": false, "after": true }`.
- `"both"` requires a space both before and after the `*` — equivalent to `{ "before": true, "after": true }`.
- `"neither"` forbids a space both before and after the `*` — equivalent to `{ "before": false, "after": false }`.

Object option:

- `"before"` (boolean, default `true`) — controls the space before the `*` token.
- `"after"` (boolean, default `false`) — controls the space after the `*` token.

Per-kind overrides (each accepts a shorthand string or a `{ "before", "after" }` object):

- `"named"` — overrides the top-level setting for named functions.
- `"anonymous"` — overrides the top-level setting for anonymous function expressions.
- `"method"` — overrides the top-level setting for class methods and object shorthand methods.
- `"shorthand"` — overrides the `"method"` setting specifically for object shorthand methods; falls back to `"method"` when omitted.

### after

Examples of **incorrect** code for this rule with the `"after"` option:

```json
{ "@stylistic/generator-star-spacing": ["error", "after"] }
```

```javascript
function *foo() {}
var foo = function *(){};
class Foo { static *bar(){} }
```

Examples of **correct** code for this rule with the `"after"` option:

```json
{ "@stylistic/generator-star-spacing": ["error", "after"] }
```

```javascript
function* foo() {}
var foo = function* (){};
class Foo { static* bar(){} }
```

### both

Examples of **incorrect** code for this rule with the `"both"` option:

```json
{ "@stylistic/generator-star-spacing": ["error", "both"] }
```

```javascript
function*foo() {}
var foo = function*(){};
class Foo { static*bar(){} }
```

Examples of **correct** code for this rule with the `"both"` option:

```json
{ "@stylistic/generator-star-spacing": ["error", "both"] }
```

```javascript
function * foo() {}
var foo = function * (){};
class Foo { static * bar(){} }
```

### neither

Examples of **incorrect** code for this rule with the `"neither"` option:

```json
{ "@stylistic/generator-star-spacing": ["error", "neither"] }
```

```javascript
function * foo() {}
var foo = function * (){};
class Foo { static * bar(){} }
```

Examples of **correct** code for this rule with the `"neither"` option:

```json
{ "@stylistic/generator-star-spacing": ["error", "neither"] }
```

```javascript
function*foo() {}
var foo = function*(){};
class Foo { static*bar(){} }
```

### Overrides per function type

Examples of **incorrect** code for this rule with `{ "before": false, "after": true, "anonymous": "neither", "method": "both" }`:

```json
{
  "@stylistic/generator-star-spacing": [
    "error",
    { "before": false, "after": true, "anonymous": "neither", "method": "both" }
  ]
}
```

```javascript
function * generator() {}
var anonymous = function* () {};
class Foo { static* method() {} }
```

Examples of **correct** code for this rule with `{ "before": false, "after": true, "anonymous": "neither", "method": "both" }`:

```json
{
  "@stylistic/generator-star-spacing": [
    "error",
    { "before": false, "after": true, "anonymous": "neither", "method": "both" }
  ]
}
```

```javascript
function* generator() {}
var anonymous = function*() {};
class Foo { static * method() {} }
```

## Original Documentation

- [@stylistic/generator-star-spacing](https://eslint.style/rules/generator-star-spacing)
