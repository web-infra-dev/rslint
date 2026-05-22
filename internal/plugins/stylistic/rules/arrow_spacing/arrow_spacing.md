# arrow-spacing

Enforce consistent spacing before and after the `=>` in arrow functions.

## Rule Details

This rule normalizes the whitespace around the `=>` token of arrow functions. It also applies to TypeScript function-type and constructor-type annotations (`type F = () => void`, `type C = new () => Foo`).

Examples of **incorrect** code for this rule with the default `{ "before": true, "after": true }` option:

```javascript
a=>a;
()=>{};
(a)=> {};
a =>a;
```

Examples of **correct** code for this rule with the default `{ "before": true, "after": true }` option:

```javascript
a => a;
() => {};
(a) => {};
```

## Options

This rule has an object option:

- `"before": true` (default) requires a space before `=>`. Set to `false` to forbid the space.
- `"after": true` (default) requires a space after `=>`. Set to `false` to forbid the space.

Either key may be omitted; a missing key falls back to its default `true`.

### before

Examples of **incorrect** code for this rule with the `{ "before": false }` option:

```json
{ "@stylistic/arrow-spacing": ["error", { "before": false }] }
```

```javascript
a =>a;
() =>{};
(a) =>{};
```

Examples of **correct** code for this rule with the `{ "before": false }` option:

```json
{ "@stylistic/arrow-spacing": ["error", { "before": false }] }
```

```javascript
a=> a;
()=> {};
(a)=> {};
```

### after

Examples of **incorrect** code for this rule with the `{ "after": false }` option:

```json
{ "@stylistic/arrow-spacing": ["error", { "after": false }] }
```

```javascript
a => a;
() => {};
(a) => {};
```

Examples of **correct** code for this rule with the `{ "after": false }` option:

```json
{ "@stylistic/arrow-spacing": ["error", { "after": false }] }
```

```javascript
a =>a;
() =>{};
(a) =>{};
```

## Original Documentation

- [@stylistic/arrow-spacing](https://eslint.style/rules/arrow-spacing)
