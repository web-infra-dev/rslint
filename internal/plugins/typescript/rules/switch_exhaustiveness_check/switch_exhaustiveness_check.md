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

Regular expression for a trailing comment that stands in for a missing `default` clause. Configured expressions use ECMAScript Unicode (`u`) semantics. The default is `/^no default$/iu`.

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

- Upstream reads TypeScript's internal `escapedName` and emits it as a bare expression for unique-symbol cases. That breaks identifiers beginning with `__` and qualified imports such as `ns.value`. rslint keeps the source-level name in diagnostics and asks the type checker for an accessible runtime expression in suggestions.
- Upstream formats ordinary enum cases from the enum-literal type text and uses the enum type's bare symbol name for bracket notation. Either path can lose the value alias or namespace qualification that is bound at the switch. rslint resolves every enum member through the type checker, preserving references such as `A.B.z`, `A.B['x-y']`, or `Alias.z`. It then chooses dot or bracket notation from the configured TypeScript target; for example, the ESNext-valid member `Ϳ` is emitted as `E.Ϳ`, while ES5 uses `E['Ϳ']`.
- If a missing enum or unique-symbol branch has no checker-proven runtime path—for example, it is visible only through a type-only import or declared only as a property in a type alias—upstream offers a suggestion that cannot be used as a runtime `case` expression. rslint keeps the diagnostic but suppresses that unsafe suggestion.
- Upstream's bracket-member suggestion escapes quotes and line endings, but not backslashes, other control characters, or lone UTF-16 surrogates. rslint fully escapes the single-quoted property key so the generated case remains parseable and refers to the original enum member.
- Configured regular expressions use ECMAScript syntax and Unicode 17 data pinned by rslint's standalone regexp engine, while upstream follows the syntax and Unicode version of the host JavaScript runtime.
- For process safety, rslint rejects a configured pattern above 131,072 UTF-8 bytes, 1,024 nested groups, 4,096 total group openings, 128 Unicode property escapes, or 100,000 weighted matcher operations that consume no engine step. It also rejects patterns whose weighted lookaround executions and actual capture count could require more than 1,000,000 unmetered capture-slot operations. Each match is limited to 100,000 engine steps and a 10,000,000 capture-slot work envelope; the step limit is lowered for zero-step-, capture-, and lookaround-heavy patterns. Unmetered zero-step and lookaround work is also accumulated per file, and crossing a compile-time limit trips that file's matcher. After a pattern exhausts a budget, remaining candidate comments in the same file are treated as non-matches. Upstream delegates its resource limits to the host JavaScript runtime, so otherwise valid but resource-intensive patterns can behave differently.

## Original Documentation

- [typescript-eslint: switch-exhaustiveness-check](https://typescript-eslint.io/rules/switch-exhaustiveness-check)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/switch-exhaustiveness-check.ts)
