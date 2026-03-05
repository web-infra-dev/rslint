# no-unsafe-member-access

## Rule Details

Disallow member access on a value with type `any`.

Accessing a member (property or element) on an `any`-typed value is unsafe because the result will also be typed as `any`, propagating the lack of type safety. This rule flags both dot-notation property access and bracket-notation element access on `any`-typed values, as well as computed member access where the index expression is typed as `any`.

Examples of **incorrect** code for this rule:

```typescript
declare const anyVal: any;
anyVal.foo;
anyVal['bar'];
anyVal[0];

declare const key: any;
declare const obj: { [k: string]: number };
obj[key];
```

Examples of **correct** code for this rule:

```typescript
declare const obj: { foo: string };
obj.foo;

declare const arr: string[];
arr[0];

declare const key: string;
declare const map: { [k: string]: number };
map[key];
```

## Original Documentation

- [typescript-eslint no-unsafe-member-access](https://typescript-eslint.io/rules/no-unsafe-member-access)
