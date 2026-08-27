# new-cap

## Rule Details

This rule requires constructor names to begin with an uppercase letter and
requires functions whose names begin with an uppercase letter to be called with
`new`.

Examples of **incorrect** code for this rule:

```javascript
const friend = new person();
const colleague = Person();
```

Examples of **correct** code for this rule:

```javascript
const friend = new Person();
const value = Boolean(input);
```

The object option supports `newIsCap`, `capIsNew`, `newIsCapExceptions`,
`newIsCapExceptionPattern`, `capIsNewExceptions`,
`capIsNewExceptionPattern`, and `properties`. The two checks and property
checking are enabled by default.

```json
{
  "new-cap": [
    "error",
    {
      "newIsCapExceptions": ["events"],
      "capIsNewExceptionPattern": "\\.Factory$",
      "properties": false
    }
  ]
}
```

```javascript
const emitter = new events();
const value = library.Factory();
const widget = new library.widget();
```

## Differences from ESLint

ESLint's exception lookup also accepts names inherited from
`Object.prototype`, such as `constructor` and `toString`, because its exception
map is a plain JavaScript object. This is an upstream implementation accident.
rslint uses an explicit string set, so lowercase constructor names are reported
unless they are configured in `newIsCapExceptions`.

## Original Documentation

- [ESLint: new-cap](https://eslint.org/docs/latest/rules/new-cap)
- [Source code](https://github.com/eslint/eslint/blob/v10.9.1/lib/rules/new-cap.js)
