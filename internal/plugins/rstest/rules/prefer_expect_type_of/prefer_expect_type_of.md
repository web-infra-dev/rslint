# prefer-expect-type-of

## Rule Details

Prefers `toBeTypeOf` over comparing `typeof` with a string. The dedicated matcher makes the assertion's intent clearer and gives a better failure message.

The rule supports regular and soft Rstest assertions.

## Incorrect

```ts
expect(typeof value).toBe('string');
```

## Correct

```ts
expect(value).toBeTypeOf('string');
```

## Autofix

Rewrites the assertion to `toBeTypeOf`, leaving the `expect` factory, a message argument, and the modifier chain untouched.
