# no-unassigned-vars

## Rule Details

This rule reports `let` and `var` variables that are read but never assigned a
value. These variables are always `undefined`, so reading them is usually a
programming mistake.

Examples of **incorrect** code for this rule:

```javascript
let status;
if (status === "ready") {
  console.log("Ready!");
}

let user;
greet(user);

function test() {
  let error;
  return error || "Unknown error";
}
```

Examples of **correct** code for this rule:

```javascript
let message = "hello";
console.log(message);

let user;
user = getUser();
console.log(user.name);

let temp;
```

Examples of **correct** TypeScript code for this rule:

```typescript
declare let value: number | undefined;
console.log(value);

declare module "my-module" {
  let value: string;
  export = value;
}
```

## Options

This rule has no options.

## Original Documentation

- [ESLint rule: no-unassigned-vars](https://eslint.org/docs/latest/rules/no-unassigned-vars)
