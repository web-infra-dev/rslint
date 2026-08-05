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

To require every interpolated value to be a `string`, empty the `allow` list and turn each `allow*` option off:

```json
{
  "@typescript-eslint/restrict-template-expressions": [
    "error",
    {
      "allow": [],
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

## Differences from ESLint

- An `allow` entry with `from: 'package'` matches the package name exactly. ESLint matches any package whose name contains the given string, so `{ from: 'package', package: 'demo' }` also permits types declared in `demo-pkg` there. Write the package name in full to permit the same types in both.
- The type name in the message may order union constituents differently. Given `declare const input: object | undefined;` and `const value = input ?? 'fallback';`, interpolating `value` reports `Invalid type ""fallback" | object" of template literal expression.` here and `Invalid type "object | "fallback"" of template literal expression.` in ESLint.

## Original Documentation

- [typescript-eslint restrict-template-expressions](https://typescript-eslint.io/rules/restrict-template-expressions)
