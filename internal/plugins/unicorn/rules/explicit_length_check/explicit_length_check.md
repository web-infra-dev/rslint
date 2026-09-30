# explicit-length-check

## Rule Details

Require explicit comparisons when checking whether `.length` or `.size` is zero.
Zero checks use `=== 0`; non-zero checks use `> 0` by default.

Examples of **incorrect** code with the default options:

```javascript
if (items.length) {}
const empty = !items.length;
const hasItems = items.length !== 0;
while (set.size) {}
```

Examples of **correct** code:

```javascript
if (items.length > 0) {}
const empty = items.length === 0;
const hasItems = items.length > 0;
while (set.size > 0) {}
```

The rule recognizes boolean coercions such as `Boolean(items.length)` and
comparisons such as `items.length >= 1` or `0 === items.length`. It ignores
computed properties, optional access, `this.length` and `this.size`, and known
values that cannot represent collection lengths. Checks alongside the same
object's `width`, `height`, or `depth` are also ignored.

As in upstream v76.0.0, TypeScript assertions around the property itself can
hide its boolean context: `if (items.length!) {}` is not reported. Prefer
`if (items.length > 0) {}` to make the intent explicit.

## Options

The `non-zero` option accepts `"greater-than"` (default) or `"not-equal"`.

```javascript
{
  'unicorn/explicit-length-check': ['error', { 'non-zero': 'not-equal' }]
}
```

With `"not-equal"`, use `items.length !== 0` for non-zero checks. Zero checks
still use `items.length === 0`.

## Fixes and Suggestions

Boolean checks are automatically fixable. Outside a boolean context, an
expression such as `items.length && render()` receives a suggestion because
replacing a numeric result with a boolean could change its caller's behavior.

The rule reports `!items.length > 0` without a fix or suggestion because
operator precedence makes that replacement unsafe. Write the intended
comparison explicitly.

## Differences from upstream

Automatic fixes preserve grouping and add a semicolon when removing a negation
could join two statements. For example, `1 + !items.length` becomes
`1 + (items.length === 0)`, and `Boolean(items.length).valueOf()` becomes
`(items.length > 0).valueOf()`. Upstream v76.0.0 omits these parentheses, which
can change the result or produce invalid code.

Vue template expressions such as `<div v-if="items.length">` are not checked.
Use the upstream rule in ESLint with `vue-eslint-parser` to check Vue templates.

## Original Documentation

- [eslint-plugin-unicorn: explicit-length-check](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/docs/rules/explicit-length-check.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/rules/explicit-length-check.js)
