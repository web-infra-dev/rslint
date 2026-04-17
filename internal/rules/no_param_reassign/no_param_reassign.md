# no-param-reassign

## Rule Details

Disallow reassigning function parameters. Reassigning a parameter mutates the caller-visible `arguments` object in non-strict code and can hide bugs caused by unintentional overwrites. With the `props` option, the rule also forbids modifying properties on parameters.

Examples of **incorrect** code for this rule:

```javascript
function foo(bar) {
  bar = 13;
}

function foo(bar) {
  bar++;
}

function foo(bar) {
  for (bar in baz) {
  }
}
```

Examples of **correct** code for this rule:

```javascript
function foo(bar) {
  var baz = bar;
}
```

## Options

```json
{
  "no-param-reassign": ["error", { "props": false }]
}
```

```json
{
  "no-param-reassign": [
    "error",
    {
      "props": true,
      "ignorePropertyModificationsFor": ["acc", "e"],
      "ignorePropertyModificationsForRegex": ["^ctx"]
    }
  ]
}
```

- `props` (default `false`) — when `true`, assignments to properties of a parameter (e.g. `bar.x = 0`, `delete bar.x`, `++bar.x`) are also reported.
- `ignorePropertyModificationsFor` (requires `props: true`) — parameter names for which property modifications are allowed.
- `ignorePropertyModificationsForRegex` (requires `props: true`) — regular expressions matching parameter names for which property modifications are allowed.

Examples of **incorrect** code with `{ "props": true }`:

```javascript
/*eslint no-param-reassign: ["error", { "props": true }]*/
function foo(bar) {
  bar.prop = 'value';
}

function foo(bar) {
  delete bar.aaa;
}

function foo(bar) {
  bar.aaa++;
}
```

Examples of **correct** code with `{ "props": true, "ignorePropertyModificationsFor": ["bar"] }`:

```javascript
/*eslint no-param-reassign: ["error", { "props": true, "ignorePropertyModificationsFor": ["bar"] }]*/
function foo(bar) {
  bar.prop = 'value';
}
```

## Original Documentation

- [ESLint no-param-reassign](https://eslint.org/docs/latest/rules/no-param-reassign)
