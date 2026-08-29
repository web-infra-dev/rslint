# no-exports-in-scripts

## Rule Details

Scripts (files whose first line starts with `#!`) are meant to be executed
directly, not imported as modules. Exports in a script mix module and script
boundaries and can mislead readers about how the file is intended to run.

Examples of **incorrect** code for this rule:

```javascript
#!/usr/bin/env node
export const foo = 1;
```

```javascript
#!/usr/bin/env node
export default foo;
```

```javascript
#!/usr/bin/env node
const foo = 1;
export {foo};
```

```javascript
#!/usr/bin/env node
export {};
```

Examples of **correct** code for this rule:

```javascript
#!/usr/bin/env node
const foo = 1;
console.log(foo);
```

```javascript
// #!/usr/bin/env node
export const foo = 1;
```

```javascript
console.log('#!/usr/bin/env node');
export const foo = 1;
```

A file without a shebang is a module, and modules are free to use exports:

```javascript
export const foo = 1;
```

## Original Documentation

- [eslint-plugin-unicorn: no-exports-in-scripts](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v72.0.0/docs/rules/no-exports-in-scripts.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v72.0.0/rules/no-exports-in-scripts.js)
