# no-restricted-types

## Rule Details

Disallow specified types. The rule reports any usage of a type whose name matches a key in the configured `types` map. Names are compared after stripping all whitespace, so a source-side `Banned<A, B>` matches a configured key of `Banned<A,B>` (and a configured key of `  NS.Banned  ` matches a source-side `NS.Banned`).

The rule recognizes these type-syntax forms:

- Primitive type keywords (`bigint`, `boolean`, `never`, `null`, `number`, `object`, `string`, `symbol`, `undefined`, `unknown`, `void`).
- Type references — both bare (`Banned`, `NS.Banned`) and parameterized (`Banned<A>`, `NS.Banned<A>`).
- The empty tuple type `[]`.
- The empty type literal `{}`.
- Heritage references — `class X implements Banned` and `interface X extends Banned`.

Each entry in `types` pairs a type name with one of:

- `true` — ban with the default message.
- `false` or `null` — explicitly do not ban this name.
- A `string` — ban with the string appended to the default message.
- An object `{ message?: string, fixWith?: string, suggest?: string[] }` — ban with an extra message, an optional auto-fix replacement, and/or one or more editor suggestions.

Examples of **incorrect** code for this rule with `{ "types": { "Banned": "Use Ok instead." } }`:

```json
{ "@typescript-eslint/no-restricted-types": ["error", { "types": { "Banned": "Use Ok instead." } }] }
```

```typescript
let value: Banned;
```

Examples of **correct** code for this rule with `{ "types": { "Banned": "Use Ok instead." } }`:

```json
{ "@typescript-eslint/no-restricted-types": ["error", { "types": { "Banned": "Use Ok instead." } }] }
```

```typescript
let value: Ok;
```

Examples of **incorrect** code for this rule with `{ "types": { "Banned": { "fixWith": "Ok", "message": "Use Ok instead." } } }`:

```json
{ "@typescript-eslint/no-restricted-types": ["error", { "types": { "Banned": { "fixWith": "Ok", "message": "Use Ok instead." } } }] }
```

```typescript
let value: Banned;
```

The auto-fix rewrites the type reference to `Ok`.

Examples of **incorrect** code for this rule with `{ "types": { "[]": "Use unknown[] instead." } }`:

```json
{ "@typescript-eslint/no-restricted-types": ["error", { "types": { "[]": "Use unknown[] instead." } }] }
```

```typescript
let value: [];
```

## Original Documentation

- [typescript-eslint no-restricted-types](https://typescript-eslint.io/rules/no-restricted-types)
