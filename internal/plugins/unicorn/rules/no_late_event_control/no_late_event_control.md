# no-late-event-control

## Rule details

Disallow calling `preventDefault()`, `stopPropagation()`, or
`stopImmediatePropagation()` on an event after synchronous dispatch has
finished. Call these methods before the handler suspends or defers work.

The rule recognizes common event parameter names such as `e`, `event`, `evt`,
and `mouseDownEvent`, as well as DOM and React event types. A known non-event
type takes precedence over the name.

Examples of **incorrect** code for this rule

```javascript
element.addEventListener("click", async (event) => {
  await save();
  event.preventDefault();
});

element.addEventListener("click", (event) => {
  setTimeout(() => {
    event.stopPropagation();
  });
});
```

Examples of **correct** code for this rule

```javascript
element.addEventListener("click", async (event) => {
  event.preventDefault();
  await save();
});

element.addEventListener("click", (event) => {
  event.stopPropagation();
  setTimeout(() => {
    update();
  });
});
```

## Limitations

- The rule reports calls inside nested functions, including callbacks that
  run synchronously. It also reports calls in generator handler bodies, which
  run when the iterator advances.
- The rule treats an earlier `await` or `yield` as a suspension even when
  the suspension and the call are in mutually exclusive branches. It also
  checks repeated parts of loops that can suspend.
- The rule checks calls on the original parameter. Calls on aliases or in
  a separate function that receives the event can go undetected.

## Differences from upstream

Rslint recognizes event types passed through a local generic alias even
without TypeScript project information. For example, if `type Identity<T> = T`,
a parameter declared as `input: Identity<MouseEvent>` is still checked for
late calls. Upstream needs project information to report that case.

Rslint reports calls in instance field initializers because those run when
an instance is created. Static fields and computed names are checked in the
surrounding handler because they run during class evaluation.

Generator parameter defaults run synchronously when the handler is called,
so Rslint allows direct event-control calls there. It reports calls in the
generator body and callbacks that capture the event.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-late-event-control.md)
- [Upstream source](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/no-late-event-control.js)
