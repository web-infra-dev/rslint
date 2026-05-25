# no-useless-computed-key

## Rule Details

Disallow computed property keys when their use is unnecessary. For example, `{ ["a"]: 1 }` can be rewritten as `{ a: 1 }` with the same behavior. The rule applies to object literals, destructuring patterns, and (by default) class members. A few keys retain distinct semantics in computed form and are therefore exempt:

- `{ ["__proto__"]: v }` defines a regular property, whereas `{ __proto__: v }` sets the object's prototype.
- In a class, `["constructor"]()` is a regular method whereas `constructor()` is the constructor.
- In a class, `static ["prototype"]` / `static ["prototype"]()` produce only a runtime error, whereas the unbracketed form is a parse error that breaks the whole script.

Examples of **incorrect** code for this rule:

```javascript
({ ['0']: 0 });
({ ['x']: 0 });
({ [0]: 0 });
({ ['x']() {} });
({ get ['foo']() {} });
var { ['x']: a } = obj;
class Foo {
  ['x']() {}
}
class Foo {
  ['0'];
}
```

Examples of **correct** code for this rule:

```javascript
({ a: 0, b() {} });
({ [x]: 0 });
({ ['__proto__']: [] });
class Foo {
  [x]() {}
}
class Foo {
  ['constructor']() {}
}
class Foo {
  static ['prototype']() {}
}
```

## Options

This rule has one option, an object with a single key:

- `enforceForClassMembers` (default `true`): when `false`, the rule only flags object literals and destructuring patterns and leaves class members alone.

Examples of **correct** code for this rule with `{ "enforceForClassMembers": false }`:

```json
{ "no-useless-computed-key": ["error", { "enforceForClassMembers": false }] }
```

```javascript
class Foo {
  ['x']() {}
}
```

## Original Documentation

- [ESLint — no-useless-computed-key](https://eslint.org/docs/latest/rules/no-useless-computed-key)
