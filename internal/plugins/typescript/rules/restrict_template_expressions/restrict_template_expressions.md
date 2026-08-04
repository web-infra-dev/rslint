# restrict-template-expressions

## Rule Details

Enforce template literal expressions to be of `string` type. When a value is interpolated into a template literal (`${expr}`), it is implicitly converted to a string, which produces results such as `"[object Object]"` for plain objects. This rule restricts which types may be interpolated.

By default, primitives that stringify predictably are permitted (`number`, `bigint`, `boolean`, `null`, `undefined`, `any`, `RegExp`), along with `Error`, `URL`, and `URLSearchParams` and their subclasses.

Examples of **incorrect** code for this rule:

```typescript
declare const obj: object;
const msg = `result: ${obj}`;

declare const arr: string[];
const msg2 = `items: ${arr}`;

declare const sym: symbol;
const msg3 = `symbol: ${sym}`;
```

Examples of **correct** code for this rule:

```typescript
const name = 'world';
const greeting = `Hello, ${name}`;

declare const obj: object;
const msg = `result: ${JSON.stringify(obj)}`;

declare const arr: string[];
const msg2 = `items: ${arr.join(', ')}`;
```

## Options

| Option         | Type                                     | Default                                     | Description                                                    |
| -------------- | ---------------------------------------- | ------------------------------------------- | -------------------------------------------------------------- |
| `allow`        | `(string \| TypeOrValueSpecifier)[]`     | `[{ from: 'lib', name: ['Error', 'URL', 'URLSearchParams'] }]` | Additional types to permit, matched against the type or any of its base types. |
| `allowAny`     | `boolean`                                | `true`                                      | Permit `any` typed values.                                     |
| `allowArray`   | `boolean`                                | `false`                                     | Permit arrays and tuples whose element type is itself permitted. |
| `allowBoolean` | `boolean`                                | `true`                                      | Permit `boolean` typed values.                                 |
| `allowNever`   | `boolean`                                | `false`                                     | Permit `never` typed values.                                   |
| `allowNullish` | `boolean`                                | `true`                                      | Permit `null` and `undefined`.                                 |
| `allowNumber`  | `boolean`                                | `true`                                      | Permit `number` and `bigint` typed values.                     |
| `allowRegExp`  | `boolean`                                | `true`                                      | Permit `RegExp` typed values.                                  |

To require every interpolated value to be a `string`, turn each `allow*` option off:

```json
{
  "@typescript-eslint/restrict-template-expressions": [
    "error",
    {
      "allowAny": false,
      "allowBoolean": false,
      "allowNever": false,
      "allowNullish": false,
      "allowNumber": false,
      "allowRegExp": false
    }
  ]
}
```

## Original Documentation

- [typescript-eslint restrict-template-expressions](https://typescript-eslint.io/rules/restrict-template-expressions)
