# no-unused-expressions

## Rule Details

Disallow expression statements that do not affect program state.

Examples of **incorrect** code for this rule:

```javascript
0;

a;

foo.bar;

a ? b() : c;

tag`template`;
```

Examples of **correct** code for this rule:

```javascript
a = b;

new Foo();

foo();

delete foo.bar;

void foo();
```

## Options

This rule has an object option:

- `allowShortCircuit` (default: `false`): allow short-circuit expressions when
  the right-hand side has an accepted side effect.
- `allowTernary` (default: `false`): allow ternary expressions when both
  branches have accepted side effects.
- `allowTaggedTemplates` (default: `false`): allow tagged template expression
  statements.
- `enforceForJSX` (default: `false`): report unused JSX expression statements.
- `ignoreDirectives` (default: `false`): accepted for ESLint config
  compatibility; directive prologues are always allowed in rslint.

Examples of **correct** code for this rule with `{ "allowShortCircuit": true }`:

```json
{ "no-unused-expressions": ["error", { "allowShortCircuit": true }] }
```

```javascript
condition && doSomething();
```

Examples of **correct** code for this rule with `{ "allowTernary": true }`:

```json
{ "no-unused-expressions": ["error", { "allowTernary": true }] }
```

```javascript
condition ? doSomething() : doSomethingElse();
```

Examples of **correct** code for this rule with `{ "allowTaggedTemplates": true }`:

```json
{ "no-unused-expressions": ["error", { "allowTaggedTemplates": true }] }
```

```javascript
tag`template`;
```

Examples of **incorrect** code for this rule with `{ "enforceForJSX": true }`:

```json
{ "no-unused-expressions": ["error", { "enforceForJSX": true }] }
```

```jsx
<div />;
```

## Original Documentation

- [ESLint: no-unused-expressions](https://eslint.org/docs/latest/rules/no-unused-expressions)
