# no-throw-literal

## Rule Details

This rule restricts what can be thrown as an exception. When ESLint was originally written, only literals were forbidden, but the rule has since been expanded to disallow any expression which cannot possibly be an `Error` object.

Examples of **incorrect** code for this rule:

```javascript
throw "error";

throw 0;

throw undefined;

throw null;

const err = new Error();
throw "an " + err;

const err2 = new Error();
throw `${err2}`;
```

Examples of **correct** code for this rule:

```javascript
throw new Error();

throw new Error("error");

const e = new Error("error");
throw e;

try {
  throw new Error("error");
} catch (e) {
  throw e;
}
```

## Original Documentation

- [ESLint: no-throw-literal](https://eslint.org/docs/latest/rules/no-throw-literal)
