# no-interpolation-in-snapshots

## Rule Details

Disallows interpolation inside inline snapshots. Updating a snapshot rewrites the literal in the source file, so interpolated values can be silently lost.

Only `toMatchInlineSnapshot` and `toThrowErrorMatchingInlineSnapshot` are checked. Other snapshot matchers keep their expected value outside the source file and are not affected by snapshot updates in the same way.

## Incorrect

```ts
expect(user).toMatchInlineSnapshot(`
  {
    "id": ${user.id},
  }
`);
```

## Correct

```ts
expect(user).toMatchInlineSnapshot(`
  {
    "id": 1,
  }
`);

expect(user).toMatchInlineSnapshot({ id: expect.any(Number) });
```
