# no-unneeded-ternary

## Rule Details

Disallows ternary expressions when a simpler alternative exists. Two patterns
are flagged:

- **Boolean-literal selection** — `cond ? true : false` (or any combination of
  boolean literals on both arms) collapses to `cond`, `!cond`, `!!cond`, or
  the boolean literal itself. Always reported.
- **Default assignment** — `a ? a : b` is equivalent to `a || b`. Reported
  only when the `defaultAssignment` option is set to `false`.

Examples of **incorrect** code for this rule:

```javascript
var a = x === 2 ? true : false;
var b = x ? true : false;
```

Examples of **correct** code for this rule:

```javascript
var a = x === 2 ? "Yes" : "No";
var b = x !== false;
var c = x ? "Yes" : "No";
var d = x ? y : x;
```

Examples of **incorrect** code for this rule with `{ "defaultAssignment": false }`:

```json
{ "no-unneeded-ternary": ["error", { "defaultAssignment": false }] }
```

```javascript
var a = x ? x : 1;
f(x ? x : 1);
```

## Options

| Option              | Type    | Default | Description                                                                              |
| ------------------- | ------- | ------- | ---------------------------------------------------------------------------------------- |
| `defaultAssignment` | boolean | `true`  | When `false`, also flag the `a ? a : b` default-assignment pattern (auto-fixed to `a \|\| b`). |

## Differences from ESLint

- TypeScript-specific expressions (`as`, `satisfies`, type assertions, etc.)
  are unknown to ESLint's precedence table; ESLint treats them as the lowest
  precedence and wraps them in parentheses defensively when emitting `||`
  fixes. rslint mirrors that behavior so an alternate like `bar as any`
  becomes `(bar as any)` after the fix, matching ESLint's output.

## Original Documentation

[no-unneeded-ternary - ESLint](https://eslint.org/docs/latest/rules/no-unneeded-ternary)
