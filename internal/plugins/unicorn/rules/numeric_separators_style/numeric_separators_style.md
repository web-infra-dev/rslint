# numeric-separators-style

## Rule details

Enforce consistent `_` separators in numeric and BigInt literals. The rule can
automatically regroup digits while preserving the value and notation of each
literal.

Examples of **incorrect** code for this rule

```javascript
const total = 1_23_4444;
const fraction = 1_234.567_89;
const mask = 0xAB_C_D_EF;
const count = 1294287712n;
```

Examples of **correct** code for this rule

```javascript
const total = 1_234_444;
const fraction = 1234.56789;
const mask = 0xAB_CD_EF;
const count = 1_294_287_712n;
```

## Options

The rule accepts `binary`, `octal`, `hexadecimal`, and `number` option objects.

| Option object | `minimumDigits` | `groupLength` | `fractionGroupLength` |
| --- | ---: | ---: | --- |
| `binary` | 0 | 4 | — |
| `octal` | 0 | 4 | — |
| `hexadecimal` | 0 | 2 | — |
| `number` | 5 | 3 | Ungrouped |

`minimumDigits` sets the digit count at which grouping starts. For decimals,
it applies independently to the integer, fractional, and exponent parts.
`groupLength` groups digits from the right. `number.fractionGroupLength`
groups fractional digits from the decimal point toward the right.

`onlyIfContainsSeparator` defaults to `false`. Set it to `true` to check only
literals that already contain `_`. Each option object can override the
top-level setting.

```javascript
export default [
  {
    plugins: ['unicorn'],
    rules: {
      'unicorn/numeric-separators-style': [
        'error',
        { number: { fractionGroupLength: 5 } },
      ],
    },
  },
];
```

## Differences from upstream

Rslint checks JavaScript and TypeScript files. Use the upstream rule to check
TOML and SOML files.

Legacy literals such as `099999.12` and `0777777` produce a syntax error in
scripts. Write them as `99999.12` and `0o777777` to check their grouping.

With `languageOptions.ecmaVersion` set to 2020 or earlier, `12345` still gets a
diagnostic. The rule withholds the autofix to `12_345` because that syntax
requires ES2021.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/numeric-separators-style.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/numeric-separators-style.js)
