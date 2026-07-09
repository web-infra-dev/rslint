# no-default-export

## Rule Details

Forbids default exports and default re-exports. Use named exports instead.

Examples of **incorrect** code for this rule:

```javascript
export default function foo() {}

const foo = "foo";
export { foo as default };

export { default } from "./foo";
```

Examples of **correct** code for this rule:

```javascript
export function foo() {}

const foo = "foo";
export { foo };
```

## Original Documentation

https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/no-default-export.md
