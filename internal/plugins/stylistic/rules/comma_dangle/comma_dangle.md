# comma-dangle

Require or disallow trailing commas.

## Rule Details

This rule enforces consistent use of trailing commas in object and array literals, function parameters and arguments, import / export specifier lists, dynamic imports, import attributes, TS enums, TS tuple types, TS generics, and TS function types.

Trailing commas simplify adding and removing items, since only the lines you are modifying must be touched.

A trailing comma after a rest binding (`function f(a, ...rest)`, `[a, ...rest]`, `{a, ...rest}`) is always disallowed — even with `"always"` — because JavaScript does not permit it. In `.tsx` files, the single-parameter generic `<T,>` is the JSX-disambiguation form and is never flagged. Holes at the end of an array (`[a,,]`, `[,,]`) are not treated as the last item and are skipped.

## Options

This rule has a string option:

- `"never"` (default) disallows trailing commas.
- `"always"` requires trailing commas.
- `"always-multiline"` requires trailing commas when the closing bracket is on a different line from the last item, and disallows them otherwise.
- `"only-multiline"` allows (but does not require) trailing commas for multi-line lists, and disallows them for single-line lists.

This rule also has an object option, with one key per list kind. Each key takes the same string enum plus `"ignore"` (skip checking for that list kind):

- `"arrays"` — array literals and array destructuring patterns.
- `"objects"` — object literals and object destructuring patterns.
- `"imports"` — ES module named imports.
- `"exports"` — ES module named exports.
- `"functions"` — function declarations, expressions, methods, arrow functions, and call / `new` arguments.
- `"importAttributes"` — import / export `with { ... }` attribute clauses.
- `"dynamicImports"` — `import(source, options)` argument lists.
- `"enums"` — TS `enum Foo { Bar }` member lists.
- `"generics"` — TS `<T, U>` type parameter lists.
- `"tuples"` — TS `[a, b]` tuple types.

Any key left unset falls back to `"never"`. TS type-argument lists (e.g. `Bar<T>`) are always treated as `"never"` regardless of the `generics` setting.

### never

Examples of **incorrect** code for this rule with the default `"never"` option:

```javascript
var foo = { bar: 'baz', qux: 'quux', };
var arr = [1, 2,];
foo({ bar: 'baz', qux: 'quux', });
```

Examples of **correct** code for this rule with the default `"never"` option:

```javascript
var foo = { bar: 'baz', qux: 'quux' };
var arr = [1, 2];
foo({ bar: 'baz', qux: 'quux' });
```

### always

Examples of **incorrect** code for this rule with the `"always"` option:

```json
{ "@stylistic/comma-dangle": ["error", "always"] }
```

```javascript
var foo = { bar: 'baz', qux: 'quux' };
var arr = [1, 2];
foo({ bar: 'baz', qux: 'quux' });
```

Examples of **correct** code for this rule with the `"always"` option:

```json
{ "@stylistic/comma-dangle": ["error", "always"] }
```

```javascript
var foo = { bar: 'baz', qux: 'quux', };
var arr = [1, 2,];
foo({ bar: 'baz', qux: 'quux', });
```

### always-multiline

Examples of **incorrect** code for this rule with the `"always-multiline"` option:

```json
{ "@stylistic/comma-dangle": ["error", "always-multiline"] }
```

```javascript
var foo = {
    bar: 'baz',
    qux: 'quux'
};

var foo = { bar: 'baz', qux: 'quux', };

var arr = [1, 2,];

var arr = [
    1,
    2
];
```

Examples of **correct** code for this rule with the `"always-multiline"` option:

```json
{ "@stylistic/comma-dangle": ["error", "always-multiline"] }
```

```javascript
var foo = {
    bar: 'baz',
    qux: 'quux',
};

var foo = { bar: 'baz', qux: 'quux' };

var arr = [1, 2];

var arr = [
    1,
    2,
];
```

### only-multiline

Examples of **incorrect** code for this rule with the `"only-multiline"` option:

```json
{ "@stylistic/comma-dangle": ["error", "only-multiline"] }
```

```javascript
var foo = { bar: 'baz', qux: 'quux', };

var arr = [1, 2,];
```

Examples of **correct** code for this rule with the `"only-multiline"` option:

```json
{ "@stylistic/comma-dangle": ["error", "only-multiline"] }
```

```javascript
var foo = {
    bar: 'baz',
    qux: 'quux',
};

var foo = {
    bar: 'baz',
    qux: 'quux'
};

var foo = { bar: 'baz', qux: 'quux' };

var arr = [
    1,
    2,
];

var arr = [
    1,
    2
];
```

### functions

Examples of **incorrect** code for this rule with the `{ "functions": "always" }` option:

```json
{ "@stylistic/comma-dangle": ["error", { "functions": "always" }] }
```

```javascript
function foo(a, b) {}

foo(a, b);
new foo(a, b);
```

Examples of **correct** code for this rule with the `{ "functions": "always" }` option:

```json
{ "@stylistic/comma-dangle": ["error", { "functions": "always" }] }
```

```javascript
function foo(a, b,) {}

foo(a, b,);
new foo(a, b,);
```

## Original Documentation

- [@stylistic/comma-dangle](https://eslint.style/rules/comma-dangle)
