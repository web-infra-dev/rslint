# switch-exhaustiveness-check

## Rule Details

Require switch statements over union types to be exhaustive. When switching over a union type (such as a discriminated union or enum), it is easy to forget to handle all possible cases. This rule ensures that every possible value of the union is handled with a `case` clause, either explicitly or via a `default` clause, preventing runtime errors from unhandled values.

Examples of **incorrect** code for this rule:

```typescript
type Direction = 'north' | 'south' | 'east' | 'west';

function move(dir: Direction) {
  switch (dir) {
    case 'north':
      break;
    case 'south':
      break;
    // 'east' and 'west' are not handled
  }
}
```

Examples of **correct** code for this rule:

```typescript
type Direction = 'north' | 'south' | 'east' | 'west';

function move(dir: Direction) {
  switch (dir) {
    case 'north':
      break;
    case 'south':
      break;
    case 'east':
      break;
    case 'west':
      break;
  }
}
```

## Options

### `allowDefaultCaseForExhaustiveSwitch`

If `false`, a `default` clause is reported as unnecessary once every member of the union already has its own `case`.

```json
{ "switch-exhaustiveness-check": ["error", { "allowDefaultCaseForExhaustiveSwitch": false }] }
```

```typescript
type Direction = 'north' | 'south';

function move(dir: Direction) {
  switch (dir) {
    case 'north':
      break;
    case 'south':
      break;
    default:
      break;
  }
}
```

### `requireDefaultForNonUnion`

If `true`, also requires a `default` clause for switches over non-union types (such as `number` or `string`), so they are held to the same standard as unions.

```json
{ "switch-exhaustiveness-check": ["error", { "requireDefaultForNonUnion": true }] }
```

```typescript
declare const value: number;

switch (value) {
  case 0:
    break;
  case 1:
    break;
}
```

### `considerDefaultExhaustiveForUnions`

If `true`, a `default` clause on a switch over a union type is itself treated as covering every unhandled member, instead of requiring each member to have an explicit `case`.

```json
{ "switch-exhaustiveness-check": ["error", { "considerDefaultExhaustiveForUnions": true }] }
```

```typescript
type Direction = 'north' | 'south';

function move(dir: Direction) {
  switch (dir) {
    case 'north':
      break;
    default:
      break;
  }
}
```

### `defaultCaseCommentPattern`

Regular expression for a trailing comment that stands in for a missing `default` clause. Defaults to `/^no default$/i`.

```json
{ "switch-exhaustiveness-check": ["error", { "defaultCaseCommentPattern": "^skip default" }] }
```

```typescript
declare const value: 'a' | 'b';

switch (value) {
  case 'a':
    break;
  // skip default
}
```

## Differences from ESLint

- When a switch is missing more than one case, the order of the types listed in the `missingBranches` message and the order the fixer inserts the corresponding `case` clauses follow rslint's internal type ordering rather than the union's declaration order. For example, `type Day = 'Monday' | 'Tuesday' | 'Wednesday'` reports missing cases as `"Monday" | "Tuesday" | "Wednesday"`, alphabetized, even when the type alias declares them in a different order.

## Original Documentation

- [typescript-eslint: switch-exhaustiveness-check](https://typescript-eslint.io/rules/switch-exhaustiveness-check)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/switch-exhaustiveness-check.ts)
