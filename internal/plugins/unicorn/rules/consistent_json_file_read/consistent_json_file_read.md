# consistent-json-file-read

## Rule Details

Use a consistent encoding when passing the result of `readFile` or `readFileSync`
to `JSON.parse()`. The default is to read the file as a UTF-8 string, which
matches the argument type of `JSON.parse()` in TypeScript.

The rule checks direct reads and reads assigned to local variables used only
for parsing. It handles both member calls and named imports from Node.js `fs`
and `fs/promises`.

Examples of **incorrect** code for the default option:

```javascript
const config = JSON.parse(await fs.readFile('./config.json'));
```

```javascript
const pending = fs.readFile('./config.json');
const config = JSON.parse(await pending);
```

Examples of **correct** code for the default option:

```javascript
const config = JSON.parse(await fs.readFile('./config.json', 'utf8'));
```

```javascript
const pending = fs.readFile('./config.json', { encoding: 'utf8', signal });
const config = JSON.parse(await pending);
```

Reads with additional options, such as `signal` or `flag`, are excluded so
automatic fixes can safely change the encoding without dropping those options.

## Options

The rule accepts `"string"` (default) or `"buffer"`.

```json
{ "unicorn/consistent-json-file-read": ["error", "buffer"] }
```

Example of **incorrect** code with `"buffer"`:

```javascript
const config = JSON.parse(await fs.readFile('./config.json', 'utf8'));
```

Example of **correct** code with `"buffer"`:

```javascript
const config = JSON.parse(await fs.readFile('./config.json'));
```

## Differences from upstream

- Rslint skips custom `JSON` bindings and exported read results. For example,
  `export const contents = fs.readFileSync(file); JSON.parse(contents);`
  keeps the exported `Buffer` intact. Unicorn reports this read.
- Rslint skips options objects with detectable mutations, including through an
  alias. For example, if `const alias = options; alias.flag = 'r';` adds an
  option to `{ encoding: 'utf8' }` before the read, Unicorn can remove the whole
  argument in `"buffer"` mode. Rslint preserves the `flag` option. This also
  excludes writes after the read, such as `Object.defineProperty(options,
  'other', { set() {} })`.
- Rslint preserves option objects with computed keys that have side effects,
  such as `{ [(log(), 'encoding')]: 'utf8' }`. Unicorn can remove the argument
  in `"buffer"` mode. Rslint also preserves options with an explicit prototype,
  such as `{ __proto__: { flag: 'r' }, encoding: null }`, so the read keeps its
  inherited `flag` option.
- Rslint skips reads whose options may be changed by another function. For
  example, after `configure(options)`, the rule preserves the options argument
  so the read still honors any `signal` added by `configure`.
- Rslint reports an encoding mismatch without a fix when removing the argument
  would also remove a call such as `Object.freeze(options)`. A freeze in a
  separate options initializer still allows the fix.

## Original Documentation

- [eslint-plugin-unicorn: consistent-json-file-read](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/consistent-json-file-read.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/consistent-json-file-read.js)
