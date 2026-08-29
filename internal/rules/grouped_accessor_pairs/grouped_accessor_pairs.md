# grouped-accessor-pairs

## Rule Details

This rule requires getter and setter definitions for the same property to be adjacent in object literals and classes. It can also enforce whether the getter or setter appears first and, when configured, checks TypeScript interface and type-literal accessors.

Examples of **incorrect** code for this rule:

```javascript
const value = {
  get name() {},
  other: true,
  set name(next) {},
};
```

Examples of **correct** code for this rule:

```javascript
const value = {
  get name() {},
  set name(next) {},
  other: true,
};
```

Examples of **incorrect** code for this rule with `"getBeforeSet"`:

```json
{ "grouped-accessor-pairs": ["error", "getBeforeSet"] }
```

```javascript
const value = {
  set name(next) {},
  get name() {},
};
```

Examples of **incorrect** TypeScript code when type accessors are enforced:

```json
{
  "grouped-accessor-pairs": [
    "error",
    "anyOrder",
    { "enforceForTSTypes": true }
  ]
}
```

```typescript
interface Value {
  get name(): string;
  other: boolean;
  set name(next: string);
}
```

## Original Documentation

- [ESLint: grouped-accessor-pairs](https://eslint.org/docs/latest/rules/grouped-accessor-pairs)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.1/lib/rules/grouped-accessor-pairs.js)
