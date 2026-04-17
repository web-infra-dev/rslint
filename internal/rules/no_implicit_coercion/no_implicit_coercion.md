# no-implicit-coercion

Disallow shorthand type conversions in favor of explicit `Boolean()` / `Number()` / `String()` calls.

## Rule Details

In JavaScript, type conversions are often performed using shorthand syntax. While these idioms work, they obscure intent; the rule flags them so they can be replaced with explicit conversions.

Patterns flagged:

- `!!foo` instead of `Boolean(foo)`
- `~foo.indexOf(bar)` (or `lastIndexOf`) instead of `foo.indexOf(bar) !== -1`
- `+foo` or `-(-foo)` instead of `Number(foo)`
- `1 * foo` or `foo * 1` instead of `Number(foo)`
- `foo - 0` instead of `Number(foo)`
- `"" + foo` / `foo + ""` (including ` ` ``) instead of `String(foo)`
- `foo += ""` instead of `foo = String(foo)`
- Template shorthand `` `${foo}` `` instead of `String(foo)` — only when the `disallowTemplateShorthand` option is enabled.

### Options

```json
{
  "no-implicit-coercion": [
    "error",
    {
      "boolean": true,
      "number": true,
      "string": true,
      "disallowTemplateShorthand": false,
      "allow": []
    }
  ]
}
```

- `boolean` (default `true`) — disallow boolean shorthand conversions (`!!`, `~`).
- `number` (default `true`) — disallow numeric shorthand conversions (`+`, `- -`, `- 0`, `* 1`).
- `string` (default `true`) — disallow string shorthand conversions (`"" +`, `+= ""`).
- `disallowTemplateShorthand` (default `false`) — also disallow `` `${foo}` `` as a coercion.
- `allow` — operators to exempt from the above. Allowed values: `"~"`, `"!!"`, `"+"`, `"- -"`, `"-"`, `"*"`.

### Fix vs suggestion

Only `!!foo` → `Boolean(foo)` applies as an autofix (and only when `Boolean` is not shadowed in scope). The other rewrites are offered as suggestions because they can change runtime behavior — `Number(1n)` throws, `foo.indexOf(x) !== -1` differs from `~foo.indexOf(x)` on non-array targets, etc.

## Examples

Incorrect:

```javascript
const b = !!foo;
const b1 = ~foo.indexOf('.');
const n = +foo;
const n2 = foo - 0;
const n3 = 1 * foo;
const s = '' + foo;
foo += '';
```

Correct:

```javascript
const b = Boolean(foo);
const b1 = foo.indexOf('.') !== -1;
const n = Number(foo);
const s = String(foo);
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-implicit-coercion
