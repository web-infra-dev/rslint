# comma-spacing

Enforce consistent spacing before and after commas.

## Rule Details

This rule normalizes the whitespace around the `,` separator in variable declarations, array literals, array destructuring patterns, object literals, object destructuring patterns, function parameter lists, function call arguments, sequence expressions, and TypeScript-specific list-shaped syntax (type arguments, type parameters, tuple types, enum members, import / export specifiers, heritage clauses).

The rule does **not** apply between:

- Two consecutive commas (`[1, , 2]` — the comma terminating the null slot is exempt)
- A comma and the immediately following `)`, `]`, or `}` (delegated to dedicated bracket-spacing rules)
- A comma and the immediately following line comment when the option forbids a space after — removing the space would fuse the comma into `//`

Examples of **incorrect** code for this rule with the default `{ "before": false, "after": true }` option:

```javascript
var foo = 1 ,bar = 2;
var arr = [1 , 2];
var obj = { foo: 'bar' ,baz: 'qur' };
foo(a ,b);
type Foo<T ,P> = Bar<T ,P>;
```

Examples of **correct** code for this rule with the default `{ "before": false, "after": true }` option:

```javascript
var foo = 1, bar = 2;
var arr = [1, 2];
var obj = { foo: 'bar', baz: 'qur' };
foo(a, b);
type Foo<T, P> = Bar<T, P>;
```

## Options

This rule has an object option:

- `"before": false` (default) forbids a space before the comma. Set to `true` to require one.
- `"after": true` (default) requires a space after the comma. Set to `false` to forbid one.

Either key may be omitted; a missing key falls back to its default.

### before

Examples of **incorrect** code for this rule with the `{ "before": true }` option:

```json
{ "@stylistic/comma-spacing": ["error", { "before": true }] }
```

```javascript
var foo = 1, bar = 2;
var arr = [1, 2];
foo(a, b);
```

Examples of **correct** code for this rule with the `{ "before": true }` option:

```json
{ "@stylistic/comma-spacing": ["error", { "before": true }] }
```

```javascript
var foo = 1 , bar = 2;
var arr = [1 , 2];
foo(a , b);
```

### after

Examples of **incorrect** code for this rule with the `{ "after": false }` option:

```json
{ "@stylistic/comma-spacing": ["error", { "after": false }] }
```

```javascript
var foo = 1, bar = 2;
var arr = [1, 2];
foo(a, b);
```

Examples of **correct** code for this rule with the `{ "after": false }` option:

```json
{ "@stylistic/comma-spacing": ["error", { "after": false }] }
```

```javascript
var foo = 1,bar = 2;
var arr = [1,2];
foo(a,b);
```

## Original Documentation

- [@stylistic/comma-spacing](https://eslint.style/rules/comma-spacing)
