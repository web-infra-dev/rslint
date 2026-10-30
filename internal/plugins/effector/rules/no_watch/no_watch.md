# no-watch

## Rule Details

Disallows calling `.watch` on an effector unit: a store, event, effect or domain. `.watch` leads to imperative code; use the `sample` operator or the `target` parameter of `sample` (or another operator) instead.

The rule uses TypeScript type information. A call is reported when the receiver's type is one of effector's unit types (`Store`, `StoreWritable`, `Event`, `EventCallable`, `Effect`, `Domain`) declared by the `effector` package, so a `.watch` method on an unrelated object is not reported. When the receiver is a union, every member must be a unit; a possibly `undefined` unit is not reported.

Examples of **incorrect** code for this rule:

```ts
const myFx = createEffect();
const myEvent = createEvent();
const $awesome = createStore(0);

myFx.finally.watch(myEvent);

myEvent.watch((payload) => {
  if (Boolean(payload)) {
    myFx(payload);
  }
});

$awesome.updates.watch((data) => {
  myEvent(identity(data));
});
```

Examples of **correct** code for this rule:

```ts
const myFx = createEffect();
const myEvent = createEvent();
const $awesome = createStore(0);

sample({
  clock: myFx.finally,
  target: myEvent,
});

sample({
  clock: myEvent,
  filter: Boolean,
  target: myFx,
});

sample({
  clock: $awesome.updates,
  fn: identity,
  target: myEvent,
});
```

## Original Documentation

- [eslint-plugin-effector: no-watch](https://github.com/effector/eslint-plugin/blob/v0.19.0/src/rules/no-watch/no-watch.md)
- [Source code](https://github.com/effector/eslint-plugin/blob/v0.19.0/src/rules/no-watch/no-watch.ts)
