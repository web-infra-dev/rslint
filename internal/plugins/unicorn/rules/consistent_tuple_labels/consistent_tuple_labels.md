# consistent-tuple-labels

## Rule Details

Require every element of a tuple type to have a label when any element is
labeled. Tuples whose elements are all labeled or all unlabeled are left
unchanged.

Examples of **incorrect** code for this rule:

```typescript
type Point = [x: number, number];
type Route = [name: string, ...string[]];
```

Examples of **correct** code for this rule:

```typescript
type Point = [x: number, y: number];
type Coordinates = [number, number];
type Route = [name: string, ...segments: string[]];
```

The rule does not provide an autofix because it cannot infer meaningful labels.

## Original Documentation

- [eslint-plugin-unicorn: consistent-tuple-labels](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/consistent-tuple-labels.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/consistent-tuple-labels.js)
