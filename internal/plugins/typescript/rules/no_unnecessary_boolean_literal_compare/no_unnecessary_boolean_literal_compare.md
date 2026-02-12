# no-unnecessary-boolean-literal-compare

## Rule Details

Disallow unnecessary equality comparisons against boolean literals.

Comparing a boolean value to `true` or `false` is redundant. The value can be used directly or negated instead. This rule flags such comparisons and provides auto-fixes.

Examples of **incorrect** code for this rule:

```typescript
declare const someCondition: boolean;
if (someCondition === true) {
}
if (someCondition !== true) {
}
if (someCondition === false) {
}
```

Examples of **correct** code for this rule:

```typescript
declare const someCondition: boolean;
if (someCondition) {
}
if (!someCondition) {
}

declare const nullableFlag: boolean | null;
if (nullableFlag === true) {
} // allowed by default for nullable booleans
```

## Original Documentation

- [typescript-eslint no-unnecessary-boolean-literal-compare](https://typescript-eslint.io/rules/no-unnecessary-boolean-literal-compare)
