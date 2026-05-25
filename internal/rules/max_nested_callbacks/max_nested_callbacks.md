# max-nested-callbacks

## Rule Details

This rule enforces a maximum depth that callbacks can be nested to improve
code readability. A common anti-pattern is "callback hell" — deeply nested
callbacks that grow rightward and become hard to follow.

A function expression or arrow function counts toward the nesting depth only
when it is passed directly to a call (as a call argument or as the callee of
an immediately-invoked call). Function-likes assigned to variables, object or
class properties, array elements, JSX attributes, default parameter values,
or used as `new` arguments / tagged-template arguments do not increase the
counter.

Examples of **incorrect** code for this rule with the default `{ "max": 10 }`
option:

```javascript
foo1(function () {
  foo2(function () {
    foo3(function () {
      foo4(function () {
        foo5(function () {
          foo6(function () {
            foo7(function () {
              foo8(function () {
                foo9(function () {
                  foo10(function () {
                    foo11(function () {});
                  });
                });
              });
            });
          });
        });
      });
    });
  });
});
```

Examples of **correct** code for this rule with the default `{ "max": 10 }`
option:

```javascript
foo1(handleFoo1);

function handleFoo1() {
  foo2(handleFoo2);
}

function handleFoo2() {
  foo3(handleFoo3);
}
```

## Options

This rule accepts a number, or an object with the following properties:

- `max` (default `10`): the maximum nesting depth allowed.
- `maximum`: deprecated alias for `max`. When both keys are present and
  `maximum` is truthy, `maximum` wins (matching ESLint's
  `option.maximum || option.max` coercion).

### `max`

Examples of **incorrect** code for this rule with `{ "max": 3 }`:

```json
{ "max-nested-callbacks": ["error", { "max": 3 }] }
```

```javascript
foo1(function () {
  foo2(function () {
    foo3(function () {
      foo4(function () {});
    });
  });
});
```

Examples of **correct** code for this rule with `{ "max": 3 }`:

```json
{ "max-nested-callbacks": ["error", { "max": 3 }] }
```

```javascript
foo1(function () {
  foo2(function () {
    foo3(function () {});
  });
});
```

Arrow functions are counted the same as function expressions:

```javascript
foo1(() => {
  foo2(() => {
    foo3(() => {
      foo4(() => {});
    });
  });
});
```

## Original Documentation

- [https://eslint.org/docs/latest/rules/max-nested-callbacks](https://eslint.org/docs/latest/rules/max-nested-callbacks)
