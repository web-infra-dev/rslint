# no-sequences

## Rule Details

This rule forbids the use of the comma operator, with the following exceptions:

- In the initialization or update portions of a `for` statement.
- By default, if the expression sequence is explicitly wrapped in parentheses. This exception can be removed with the `"allowInParentheses": false` option.

Examples of **incorrect** code for this rule:

```javascript
foo = doSomething(), val;

0, eval("doSomething();");

do {} while (doSomething(), !!test);

for (; doSomething(), !!test; );

if (doSomething(), !!test);

switch (val = foo(), val) {}

while (val = foo(), val < 42);

with (doSomething(), val) {}

const foo = (val) => (console.log('bar'), val);
```

Examples of **correct** code for this rule:

```javascript
foo = (doSomething(), val);

(0, eval)("doSomething();");

do {} while ((doSomething(), !!test));

for (i = 0, j = 10; i < j; i++, j--);

if ((doSomething(), !!test));

switch ((val = foo(), val)) {}

while ((val = foo(), val < 42));

with ((doSomething(), val)) {}

const foo = (val) => ((console.log('bar'), val));
```

## Options

This rule takes one optional object argument:

- `allowInParentheses` — when set to `false`, disallows expression sequences even when explicitly wrapped in parentheses. Default `true`.

Examples of **incorrect** code for this rule with `{ "allowInParentheses": false }`:

```json
{ "no-sequences": ["error", { "allowInParentheses": false }] }
```

```javascript
var foo = (1, 2);

(0, eval)("doSomething();");

foo(a, (b, c), d);
```

## Original Documentation

- [ESLint rule: no-sequences](https://eslint.org/docs/latest/rules/no-sequences)
