# id-denylist

## Rule Details

Generic names can lead to hard-to-decipher code. This rule lets you list identifier names that should not be used, and reports every place the code introduces one.

The rule catches denied identifiers that are:

- variable declarations
- function declarations
- object properties assigned to during object creation
- class fields
- class methods

It leaves alone denied identifiers that are:

- function calls and their arguments, so you can keep calling functions you do not control
- property reads, so you can keep reading properties of objects you do not control
- references to global variables the code does not declare, whose names you do not control

Examples of **incorrect** code for this rule:

```json
{ "id-denylist": ["error", "data", "callback"] }
```

```javascript
const data = { ...values };

function callback() {
  // ...
}

element.callback = function () {
  // ...
};

const itemSet = {
  data: [...values],
};

class Foo {
  data = [];
}

class Bar {
  #data = [];
}

class Baz {
  callback() {}
}

class Qux {
  #callback() {}
}
```

Examples of **correct** code for this rule:

```json
{ "id-denylist": ["error", "data", "callback"] }
```

```javascript
const encodingOptions = { ...values };

function processFileResult() {
  // ...
}

element.successHandler = function () {
  // ...
};

const itemSet = {
  entities: [...values],
};

callback();

foo.callback();

foo.data;

class Foo {
  items = [];
}

class Bar {
  #items = [];
}

class Baz {
  method() {}
}

class Qux {
  #method() {}
}
```

## Options

The rule takes one or more strings: the names of the denied identifiers.

```json
{ "id-denylist": ["error", "data", "err", "e", "cb", "callback"] }
```

## Differences from ESLint

- When a denied identifier carries a TypeScript type annotation, the reported range covers the name only. ESLint's TypeScript parser stretches it over the annotation as well, so `let data: string` is reported as `data` here and as `data: string` there.

## Original Documentation

- [ESLint: id-denylist](https://eslint.org/docs/latest/rules/id-denylist)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/id-denylist.js)
