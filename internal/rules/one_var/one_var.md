# one-var

Enforce variables to be declared either together or separately in functions.

> **Upstream status:** This rule is **frozen** in ESLint — it is no longer accepting new features and is not part of the recommended config. Bug-fix-only.

## Rule Details

This rule enforces a single variable declaration style in a scope. Variables can be declared together in a single statement or each in a separate statement, depending on the configured mode. The rule supports `var`, `let`, `const`, `using`, and `await using` declarations.

Examples of **incorrect** code for this rule:

```javascript
function foo() {
    var bar;
    var baz;
}

function foo() {
    let bar;
    let baz;
}

function foo() {
    const bar = 1;
    const baz = 2;
}
```

Examples of **correct** code for this rule:

```javascript
function foo() {
    var bar, baz;
}

function foo() {
    let bar, baz;
}

function foo() {
    const bar = 1, baz = 2;
}
```

## Options

This rule accepts a string or an object as its only option.

### String option

The string option configures the same mode for every declaration kind:

- `"always"` (default) — requires one variable declaration per scope.
- `"never"` — requires multiple separate variable declarations per scope.
- `"consecutive"` — requires consecutive declarations of the same kind to be combined into a single statement.

```json
{ "one-var": ["error", "never"] }
```

```javascript
function foo() {
    var bar;
    var baz;
}
```

```json
{ "one-var": ["error", "consecutive"] }
```

```javascript
function foo() {
    var bar, baz;
    qux();
    var quux;
}
```

### Object option (per kind)

The object form lets you choose a different mode for each declaration kind:

- `"var"`, `"let"`, `"const"`, `"using"`, `"awaitUsing"` — each accepts `"always"`, `"never"`, or `"consecutive"`.
- `"separateRequires"` — when `true`, treats `require()` calls as a separate group from other initialized declarations of the same kind.

```json
{ "one-var": ["error", { "var": "always", "let": "never", "const": "never" }] }
```

```javascript
function foo() {
    var bar, baz;
    let qux;
    let norf;
    const a = 1;
    const b = 2;
}
```

```json
{ "one-var": ["error", { "separateRequires": true, "var": "always" }] }
```

```javascript
var foo = require('foo');
var bar = require('bar');
var baz = 'baz', qux = 'qux';
```

### Object option (initialized / uninitialized)

The alternative object form discriminates by initialization status across all kinds:

- `"initialized"` — applies to declarations that have an initializer.
- `"uninitialized"` — applies to declarations without an initializer.

Both accept `"always"`, `"never"`, or `"consecutive"`.

```json
{ "one-var": ["error", { "initialized": "never", "uninitialized": "always" }] }
```

```javascript
function foo() {
    var bar, baz;
    var a = 1;
    var b = 2;
}
```

## Differences from ESLint

- **`declare`-modified declarations are reported but not auto-fixed.** ESLint v10 emits fixes that produce TypeScript parse errors on this input shape: splitting `declare var a, b;` yields `declare var a; var b;` (the second statement silently loses ambient semantics), and combining `declare var a; declare var b;` yields `declare var a,  var b;` (two `var` keywords, syntactically invalid). rslint deliberately suppresses the fix in these cases to avoid producing broken code; the diagnostic itself is still reported with identical position and message.

## Original Documentation

- [ESLint rule](https://eslint.org/docs/latest/rules/one-var)
- [Source code](https://github.com/eslint/eslint/blob/main/lib/rules/one-var.js)
