# no-new-native-nonconstructor

## Rule Details

Disallows using the `new` operator with native global functions that are not constructors.

Examples of **incorrect** code for this rule:

```javascript
const foo = new Symbol('foo');
const bar = new BigInt(9007199254740991);
```

Examples of **correct** code for this rule:

```javascript
const foo = Symbol('foo');
const bar = BigInt(9007199254740991);

function baz(Symbol) {
  const qux = new Symbol('baz');
}

const SymbolCtor = Symbol;
new SymbolCtor();
```

## Original Documentation

- [ESLint no-new-native-nonconstructor](https://eslint.org/docs/latest/rules/no-new-native-nonconstructor)
