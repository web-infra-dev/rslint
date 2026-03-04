# no-redundant-type-constituents

## Rule Details

Disallow members of unions and intersections that do nothing or override other members.

Some types can override other types in a union or intersection, rendering them redundant. For example, `any` in a union type overrides all other members, and `never` in an intersection type overrides all other members. These redundant constituents can be misleading and should be removed.

Examples of **incorrect** code for this rule:

```typescript
type Union = any | string;
type Intersection = string & any;
type PrimitiveOverride = string | 'hello';
type NeverUnion = string | never;
```

Examples of **correct** code for this rule:

```typescript
type Union = string | number;
type Intersection = string & { foo: string };
type ValidUnion = string | number | boolean;
type ReturnType = string | never; // allowed in return type position
```

## Original Documentation

- [typescript-eslint no-redundant-type-constituents](https://typescript-eslint.io/rules/no-redundant-type-constituents)
