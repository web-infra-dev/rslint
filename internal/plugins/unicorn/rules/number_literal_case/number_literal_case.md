# number-literal-case

## Rule Details

Enforce lowercase radix prefixes (`0x`, `0o`, and `0b`) and a lowercase `e`
in exponential notation. Hexadecimal digits use uppercase by default.
The rule checks both numbers and BigInts and provides an automatic fix.

Examples of **incorrect** code for this rule:

```javascript
const mask = 0xff;
const flags = 0B1010n;
const permissions = 0O755;
const timeout = 1E3;
```

Examples of **correct** code for this rule:

```javascript
const mask = 0xFF;
const flags = 0b1010n;
const permissions = 0o755;
const timeout = 1e3;
```

Numeric separators and the lowercase BigInt suffix `n` are preserved.
Strings containing numeric-looking text are not checked.

## Options

### `hexadecimalValue`

Type: `"uppercase" | "lowercase"`

Default: `"uppercase"`

Choose the case of hexadecimal digits. The `0x` prefix is always lowercase.

```javascript
{
  'unicorn/number-literal-case': ['error', { hexadecimalValue: 'lowercase' }]
}
```

With `hexadecimalValue: "lowercase"`, use `0xff` and `0xffn` instead of
`0xFF` and `0xFFn`.

## Differences from upstream

Vue single-file components and template expressions are not supported by
rslint's parser. For example, this rule does not check `{{ 1.2E3 }}` inside
a Vue template.

The TypeScript parser rejects legacy numeric spellings such as `0777` and
`0888` before lint rules run, even in scripts. Upstream accepts these in
sloppy-mode JavaScript and leaves them unchanged.

## Original Documentation

- [eslint-plugin-unicorn: number-literal-case](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/number-literal-case.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/number-literal-case.js)
