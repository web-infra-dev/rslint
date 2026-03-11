# no-unnecessary-boolean-literal-compare

## Rule Details

Disallow unnecessary equality comparisons against boolean literals.

Comparing a boolean value to `true` or `false` is redundant. The value can be used directly or negated instead. This rule flags such comparisons and provides auto-fixes.

### Options

- `allowComparingNullableBooleansToTrue` (default: `true`): When set to `true`, allows `nullableVar === true` or `nullableVar !== true` comparisons for nullable boolean types (`boolean | null | undefined`). Set to `false` to flag these comparisons.
- `allowComparingNullableBooleansToFalse` (default: `true`): When set to `true`, allows `nullableVar === false` or `nullableVar !== false` comparisons for nullable boolean types. Set to `false` to flag these comparisons and suggest using the `??` operator instead.
- `allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing` (default: `false`): When set to `true`, allows the rule to run even when `strictNullChecks` is not enabled. By default, the rule reports an error if `strictNullChecks` is off.

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

Examples of **incorrect** code with `{ allowComparingNullableBooleansToTrue: false }`:

```typescript
declare const nullableFlag: boolean | undefined;
if (nullableFlag === true) {
}
```

Examples of **incorrect** code with `{ allowComparingNullableBooleansToFalse: false }`:

```typescript
declare const nullableFlag: boolean | null;
if (nullableFlag === false) {
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
