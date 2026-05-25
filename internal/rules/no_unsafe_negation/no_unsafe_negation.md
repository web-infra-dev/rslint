# no-unsafe-negation

## Rule Details

Disallows negating the left operand of relational operators. The code `!a in b` is parsed as `(!a) in b`, not `!(a in b)`, which is usually not the intended behavior. The same applies to `instanceof`.

With the `enforceForOrderingRelations` option enabled, this rule also checks `<`, `>`, `<=`, and `>=` operators.

Examples of **incorrect** code for this rule:

```javascript
if ((!key) in object) {
}

if ((!obj) instanceof Ctor) {
}
```

Examples of **correct** code for this rule:

```javascript
if (!(key in object)) {
}

if (!(obj instanceof Ctor)) {
}

if ((!key) in object) {
}
```

## Options

### `enforceForOrderingRelations`

When set to `true`, also disallows negating the left operand of `<`, `>`, `<=`, and `>=` operators. Default is `false`.

Examples of **incorrect** code with `{ "enforceForOrderingRelations": true }`:

```javascript
if (!a < b) {
}

if (!a >= b) {
}
```

## Original Documentation

- [ESLint no-unsafe-negation](https://eslint.org/docs/latest/rules/no-unsafe-negation)
