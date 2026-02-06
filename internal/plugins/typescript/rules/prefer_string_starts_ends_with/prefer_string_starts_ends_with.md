# prefer-string-starts-ends-with

## Rule Details

Enforce using `String#startsWith` and `String#endsWith` over other equivalent methods of checking substrings.

There are multiple ways to verify if a string starts or ends with a specific string, such as `foo.indexOf('bar') === 0`, `foo.charAt(0) === 'b'`, or regex tests like `/^bar/.test(foo)`. Since ES2015 has added `String#startsWith` and `String#endsWith`, this rule reports on other ways of checking, suggesting the use of the built-in methods instead.

Examples of **incorrect** code for this rule:

```typescript
declare const foo: string;

foo[0] === 'b';
foo.charAt(0) === 'b';
foo.indexOf('bar') === 0;
foo.slice(0, 3) === 'bar';
foo.substring(0, 3) === 'bar';
foo.match(/^bar/) != null;
/^bar/.test(foo);

foo[foo.length - 1] === 'b';
foo.charAt(foo.length - 1) === 'b';
foo.lastIndexOf('bar') === foo.length - 3;
foo.slice(-3) === 'bar';
foo.substring(foo.length - 3) === 'bar';
foo.match(/bar$/) != null;
/bar$/.test(foo);
```

Examples of **correct** code for this rule:

```typescript
declare const foo: string;

foo.startsWith('bar');
foo.endsWith('bar');

foo.startsWith('a');
foo.endsWith('a');
```

## Options

### `allowSingleElementEquality`

When set to `"always"`, allows equality checks for a single character (e.g. `foo[0] === 'a'` and `foo.charAt(0) === 'a'`).

```json
{
  "@typescript-eslint/prefer-string-starts-ends-with": [
    "warn",
    { "allowSingleElementEquality": "always" }
  ]
}
```

## Original Documentation

- [typescript-eslint prefer-string-starts-ends-with](https://typescript-eslint.io/rules/prefer-string-starts-ends-with)
