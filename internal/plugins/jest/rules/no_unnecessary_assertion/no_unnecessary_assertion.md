# no-unnecessary-assertion

## Rule Details

When using TypeScript, runtime assertions based on types can often be omitted provided that the types are accurate.

This rule warns when you do an assertion about being `null`, `undefined`, or `NaN` on something that cannot be those types, as that indicates either the assertion can be removed or the types need to be adjusted. It checks `toBeNull`, `toBeUndefined`, `toBeDefined` and `toBeNaN`, with or without `.not`; assertions using `.resolves` or `.rejects` are not checked.

This rule requires type information. It reports a configuration warning when the `strictNullChecks` compiler option is disabled, because without it no type can rule out `null` or `undefined`.

Examples of **incorrect** code for this rule:

```ts
expect('hello world'.match('sunshine') ?? []).toBeNull();

expect(User.findOrThrow(1)).toBeDefined();

expect(map.getOrInsert('key', 'default')).not.toBeUndefined();

expect(user.name).not.toBeNaN();
```

Examples of **correct** code for this rule:

```ts
expect('hello world'.match('sunshine')).toBeNull();

expect(User.findOrNull(1)).toBeNull();

expect(map.get('key')).not.toBeUndefined();

expect(user.age).not.toBeNaN();
```

## Original Documentation

- [eslint-plugin-jest: no-unnecessary-assertion](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/no-unnecessary-assertion.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/no-unnecessary-assertion.ts)
