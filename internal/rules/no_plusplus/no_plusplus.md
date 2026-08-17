# no-plusplus

## Rule Details

Because the unary `++` and `--` operators are subject to automatic semicolon insertion, differences in whitespace can change the semantics of source code.

This rule disallows the unary operators `++` and `--`.

Examples of **incorrect** code for this rule:

```javascript
let foo = 0;
foo++;

let bar = 42;
bar--;

for (let i = 0; i < l; i++) {
  doSomething(i);
}
```

Examples of **correct** code for this rule:

```javascript
let foo = 0;
foo += 1;

let bar = 42;
bar -= 1;

for (let i = 0; i < l; i += 1) {
  doSomething(i);
}
```

## Options

This rule has an object option.

- `"allowForLoopAfterthoughts": true` allows unary operators `++` and `--` in the afterthought (final expression) of a `for` loop.

### allowForLoopAfterthoughts

Examples of **correct** code for this rule with the `{ "allowForLoopAfterthoughts": true }` option:

```json
{ "no-plusplus": ["error", { "allowForLoopAfterthoughts": true }] }
```

```javascript
for (let i = 0; i < l; i++) {
  doSomething(i);
}

for (let i = l; i >= 0; i--) {
  doSomething(i);
}

for (let i = 0, j = l; i < l; i++, j--) {
  doSomething(i, j);
}
```

Examples of **incorrect** code for this rule with the `{ "allowForLoopAfterthoughts": true }` option:

```javascript
for (let i = 0; i < l; j = i++) {
  doSomething(i, j);
}

for (let i = l; i--; ) {
  doSomething(i);
}

for (let i = 0; i < l; ) i++;
```

## Original Documentation

- [ESLint: no-plusplus](https://eslint.org/docs/latest/rules/no-plusplus)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-plusplus.js)
