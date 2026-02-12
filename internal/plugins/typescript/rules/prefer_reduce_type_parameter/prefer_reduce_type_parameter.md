# prefer-reduce-type-parameter

## Rule Details

Enforce using type parameters for `Array#reduce` instead of type assertions on the initial value. When calling `Array#reduce`, it is common to use a type assertion (`as`) on the initial value to specify the result type. However, `Array#reduce` accepts a type parameter that achieves the same result in a cleaner way without requiring a type assertion.

Examples of **incorrect** code for this rule:

```typescript
[1, 2, 3].reduce(
  (acc, val) => ({ ...acc, [val]: true }),
  {} as Record<number, boolean>,
);
['a', 'b'].reduce((acc, val) => [...acc, val], [] as string[]);
```

Examples of **correct** code for this rule:

```typescript
[1, 2, 3].reduce<Record<number, boolean>>(
  (acc, val) => ({ ...acc, [val]: true }),
  {},
);
['a', 'b'].reduce<string[]>((acc, val) => [...acc, val], []);
```

## Original Documentation

- [typescript-eslint prefer-reduce-type-parameter](https://typescript-eslint.io/rules/prefer-reduce-type-parameter)
