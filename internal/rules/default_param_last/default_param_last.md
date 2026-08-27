# default-param-last

## Rule Details

Default parameters are most useful at the end of a parameter list, where callers can omit them. This rule reports default or optional parameters followed by required parameters.

Examples of **incorrect** code for this rule:

```javascript
function createUser(isAdmin = false, id) {}

function connect(host = "localhost", port) {}
```

Examples of **correct** code for this rule:

```javascript
function createUser(id, isAdmin = false) {}

function connect(port, host = "localhost") {}
```

The rule also supports TypeScript optional parameters and parameter properties.

Examples of **incorrect** TypeScript code for this rule:

```typescript
function format(value?: string, radix: number) {}

class Client {
  constructor(public endpoint = "localhost", private retries: number) {}
}
```

Examples of **correct** TypeScript code for this rule:

```typescript
function format(radix: number, value?: string) {}

class Client {
  constructor(private retries: number, public endpoint = "localhost") {}
}
```

## Original Documentation

- [ESLint: default-param-last](https://eslint.org/docs/latest/rules/default-param-last)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.1/lib/rules/default-param-last.js)
