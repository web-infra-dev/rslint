# no-duplicate-type-constituents

## Rule Details

Disallows duplicate constituents in union or intersection types. Having the same type more than once in a union (`|`) or intersection (`&`) is redundant and can be removed without changing the type. This rule also flags explicit `undefined` on optional parameters, since the `?` modifier already implies `undefined`.

Examples of **incorrect** code for this rule:

```typescript
type Foo = string | string;

type Bar = number & number;

type Baz = 'a' | 'b' | 'a';

function fn(x?: string | undefined) {}
```

Examples of **correct** code for this rule:

```typescript
type Foo = string | number;

type Bar = { a: string } & { b: number };

type Baz = 'a' | 'b' | 'c';

function fn(x?: string) {}
```

## Differences from upstream

In typescript-eslint v8.72.2, autofixing a nested union or intersection with a leading separator can remove every member and leave invalid syntax:

```typescript
type T = string | number | (| string | number);
```

Rslint keeps at least one member in each surviving group during a fix pass. It still reports every duplicate, but defers removal of the last member so later passes can safely remove the redundant group. The same protection applies to explicit `undefined` in optional parameters. If there is no safe removal, the diagnostic is reported without an autofix.

Rslint also excludes a containing group when looking for a previous occurrence of the same type. For example, `(string | never)` simplifies to `string`, but that does not make its own `string` member a duplicate. Removing that member would change the type.

## Original Documentation

- [typescript-eslint: no-duplicate-type-constituents](https://typescript-eslint.io/rules/no-duplicate-type-constituents)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.72.2/packages/eslint-plugin/src/rules/no-duplicate-type-constituents.ts)
