# no-unnecessary-splice

## Rule details

This rule reports `splice()` calls that can be removed or expressed using
`shift()`, `unshift()`, `pop()`, `push()`, or a `length = 0` assignment.
It preserves calls that use the removed elements when emptying an array.
For other cases, it can report a call whose result is used and leave the
rewrite to you. The rule withholds autofixes when it detects that an edit
would remove a comment or a side effect.

Examples of **incorrect** code for this rule

```javascript
array.splice(index, 0);
array.splice(0, 1);
array.splice(0, 0, item);
array.splice(array.length - 1, 1);
array.splice(array.length, 0, item);
array.splice(0);
```

Examples of **correct** code for this rule

```javascript
array.shift();
array.unshift(item);
array.pop();
array.push(item);
array.length = 0;
```

The rule recognizes non-array receivers documented with JSDoc annotations
such as `@param`, `@type`, and `@returns`.

## Options

This rule has no options.

## Differences from upstream

Rslint skips calls such as `array.splice(...[0, 1], 0)` because a spread in
the first argument can provide the delete count. Unicorn v79.0.0 can report
and remove this call even though it removes the first element.

Rslint keeps the diagnostic when it cannot safely remove a call such as
`array.splice(index, 0)`. An index with an unknown value may call
`valueOf()` or throw during conversion. A getter read inside an expression,
such as `array.splice(options.index + 1, 0)`, is also preserved.

Rslint withholds the `push()` autofix for `array.splice(array.length, 0, ...items)`
when iteration might run user code. An iterator that appends to `array`
can change where the inserted items end up.

When removing a no-op call directly before a string statement in a script
or function body, Rslint leaves a semicolon to keep strings such as
`"use strict"` from becoming directives.

Rslint leaves calls such as `array.items.splice((array?.items).length - 1, 1)`
alone because the optional chain can affect how the length is evaluated.

Rslint fixes `do(array.splice)(0); while (ready);` to
`do array.length = 0; while (ready);` with the space required after `do`.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-unnecessary-splice.md)
- [Upstream implementation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-unnecessary-splice.js)
