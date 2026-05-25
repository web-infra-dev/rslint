# no-empty-object-type

## Rule Details

Disallows accidental uses of the `{}` ("empty object") type. In TypeScript `{}` is the type of any non-nullish value, which is rarely what authors intend — typically they want `object` (any non-primitive value) or `unknown` (any value at all). The same problem applies to empty `interface` declarations: an empty interface with no members is equivalent to `{}`, and an empty interface that extends a single supertype is equivalent to a type alias of that supertype.

This rule reports both shapes and offers `object` / `unknown` suggestions.

Examples of **incorrect** code for this rule:

```typescript
let anyObject: {};
let anyValue: {};

interface AnyObject {}
interface AnyValue {}

type AnyObjectAlias = {};
type AnyValueAlias = {};
```

Examples of **correct** code for this rule:

```typescript
let anyObject: object;
let anyValue: unknown;

type AnyObjectAlias = object;
type AnyValueAlias = unknown;

let objectWith: {
  property: boolean;
};

interface InterfaceWith {
  property: boolean;
}

type TypeWith = {
  property: boolean;
};
```

### Options

- `allowInterfaces` (default `"never"`): how empty interfaces are treated.
  - `"never"`: empty interfaces are disallowed.
  - `"always"`: empty interfaces are always allowed.
  - `"with-single-extends"`: empty interfaces are allowed only when they extend exactly one supertype.
- `allowObjectTypes` (default `"never"`): how empty `{}` type literals are treated.
  - `"never"`: empty `{}` is disallowed.
  - `"always"`: empty `{}` is always allowed.
- `allowWithName`: a regular expression source string. When set, interfaces and `type` aliases whose name matches the pattern are exempted.

### `allowInterfaces: "with-single-extends"`

Examples of **correct** code:

```typescript
interface Base {
  value: boolean;
}

interface Derived extends Base {}
```

### `allowObjectTypes: "always"`

Examples of **correct** code:

```typescript
type AnyObject = {};

let value: {};
```

### `allowWithName: "Props$"`

Examples of **correct** code:

```typescript
interface ComponentProps {}

type DialogProps = {};
```

## Original Documentation

- [typescript-eslint no-empty-object-type](https://typescript-eslint.io/rules/no-empty-object-type)
