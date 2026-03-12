# ban-tslint-comment

## Rule Details

Disallow TSLint directive comments such as `// tslint:disable` and
`// tslint:disable-next-line`. These directives are not used by ESLint and are
typically left behind when migrating from TSLint.

Examples of **incorrect** code for this rule:

```javascript
/* tslint:disable */
/* tslint:enable */
// tslint:disable-next-line
someCode(); // tslint:disable-line
```

Examples of **correct** code for this rule:

```javascript
// some other comment
/* another comment that mentions tslint */
```

## Original Documentation

- [typescript-eslint ban-tslint-comment](https://typescript-eslint.io/rules/ban-tslint-comment)
