# no-restricted-exports

## Rule Details

In a project, certain names may be disallowed from being used as exported names for various reasons.

This rule disallows specified names from being used as exported names.

## Options

By default, this rule doesn't disallow any names. Only the names you specify in the configuration will be disallowed.

This rule has an object option:

- `"restrictedNamedExports"` is an array of strings, where each string is a name to be restricted.
- `"restrictedNamedExportsPattern"` is a string representing a regular expression pattern. Named exports matching this pattern will be restricted. This option does not apply to `default` named exports.
- `"restrictDefaultExports"` is an object option with boolean properties to restrict certain default export declarations. The option works only if the `restrictedNamedExports` option does not contain the `"default"` value. The following properties are allowed:
  - `direct`: restricts `export default` declarations.
  - `named`: restricts `export { foo as default };` declarations.
  - `defaultFrom`: restricts `export { default } from 'foo';` declarations.
  - `namedFrom`: restricts `export { foo as default } from 'foo';` declarations.
  - `namespaceFrom`: restricts `export * as default from 'foo';` declarations.

### restrictedNamedExports

Examples of **incorrect** code for the `"restrictedNamedExports"` option:

```json
{ "no-restricted-exports": ["error", { "restrictedNamedExports": ["foo", "bar", "Baz", "a", "b", "c", "d", "e", "👍"] }] }
```

```javascript
export const foo = 1;

export function bar() {}

export class Baz {}

const a = {};
export { a };

function someFunction() {}
export { someFunction as b };

export { c } from "some_module";

export { "d" } from "some_module";

export { something as e } from "some_module";

export { "👍" } from "some_module";
```

Examples of **correct** code for the `"restrictedNamedExports"` option:

```json
{ "no-restricted-exports": ["error", { "restrictedNamedExports": ["foo", "bar", "Baz", "a", "b", "c", "d", "e", "👍"] }] }
```

```javascript
export const quux = 1;

export function myFunction() {}

export class MyClass {}

const a = {};
export { a as myObject };

function someFunction() {}
export { someFunction };

export { c as someName } from "some_module";

export { "d" as " d " } from "some_module";

export { something } from "some_module";

export { "👍" as thumbsUp } from "some_module";
```

#### Default exports

By design, the `"restrictedNamedExports"` option doesn't disallow `export default` declarations. If you configure `"default"` as a restricted name, that restriction will apply only to named export declarations.

Examples of additional **incorrect** code for the `"restrictedNamedExports": ["default"]` option:

```json
{ "no-restricted-exports": ["error", { "restrictedNamedExports": ["default"] }] }
```

```javascript
function foo() {}

export { foo as default };
```

```json
{ "no-restricted-exports": ["error", { "restrictedNamedExports": ["default"] }] }
```

```javascript
export { default } from "some_module";
```

Examples of additional **correct** code for the `"restrictedNamedExports": ["default"]` option:

```json
{ "no-restricted-exports": ["error", { "restrictedNamedExports": ["default", "foo"] }] }
```

```javascript
export default function foo() {}
```

### restrictedNamedExportsPattern

Example of **incorrect** code for the `"restrictedNamedExportsPattern"` option:

```json
{ "no-restricted-exports": ["error", { "restrictedNamedExportsPattern": "bar$" }] }
```

```javascript
export const foobar = 1;
```

Example of **correct** code for the `"restrictedNamedExportsPattern"` option:

```json
{ "no-restricted-exports": ["error", { "restrictedNamedExportsPattern": "bar$" }] }
```

```javascript
export const abc = 1;
```

Note that this option does not apply to `export default` or any `default` named exports. If you want to also restrict `default` exports, use the `restrictDefaultExports` option.

### restrictDefaultExports

This option allows you to restrict certain `default` declarations. The option works only if the `restrictedNamedExports` option does not contain the `"default"` value. This option accepts the following properties:

#### direct

Examples of **incorrect** code for the `"restrictDefaultExports": { "direct": true }` option:

```json
{ "no-restricted-exports": ["error", { "restrictDefaultExports": { "direct": true } }] }
```

```javascript
export default foo;
```

```javascript
export default 42;
```

```javascript
export default function foo() {}
```

#### named

Examples of **incorrect** code for the `"restrictDefaultExports": { "named": true }` option:

```json
{ "no-restricted-exports": ["error", { "restrictDefaultExports": { "named": true } }] }
```

```javascript
const foo = 123;

export { foo as default };
```

#### defaultFrom

Examples of **incorrect** code for the `"restrictDefaultExports": { "defaultFrom": true }` option:

```json
{ "no-restricted-exports": ["error", { "restrictDefaultExports": { "defaultFrom": true } }] }
```

```javascript
export { default } from "foo";
```

```javascript
export { default as default } from "foo";
```

#### namedFrom

Examples of **incorrect** code for the `"restrictDefaultExports": { "namedFrom": true }` option:

```json
{ "no-restricted-exports": ["error", { "restrictDefaultExports": { "namedFrom": true } }] }
```

```javascript
export { foo as default } from "foo";
```

#### namespaceFrom

Examples of **incorrect** code for the `"restrictDefaultExports": { "namespaceFrom": true }` option:

```json
{ "no-restricted-exports": ["error", { "restrictDefaultExports": { "namespaceFrom": true } }] }
```

```javascript
export * as default from "foo";
```

## Known Limitations

This rule doesn't inspect the content of source modules in re-export declarations. In particular, if you are re-exporting everything from another module's export, that export may include a restricted name. This rule cannot detect such cases.

```javascript
//----- some_module.js -----
export function foo() {}

//----- my_module.js -----
// { "restrictedNamedExports": ["foo"] }
export * from "some_module"; // allowed, although this declaration exports "foo" from my_module
```

This rule inspects named export declarations purely syntactically, matching ESLint's behavior: named exports of TypeScript-only declarations (`interface`, `type`, `enum`) are not checked, but `export type { name }` re-export specifiers are checked the same as value exports.

## Original Documentation

- [ESLint: no-restricted-exports](https://eslint.org/docs/latest/rules/no-restricted-exports)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-restricted-exports.js)
