import { RuleTester } from '../rule-tester';

// Mirrors eslint-plugin-effector v0.19.0 src/rules/no-watch/no-watch.test.ts.
// `effector` resolves to the `declare module` stub in ../files/effector.d.ts.
const ruleTester = new RuleTester();

const restricted = (line: number, column: number) => ({
  messageId: 'restricted',
  message:
    'Using `.watch` method leads to imperative code. Replace it with an operator `sample` or use the `target` parameter of `sample` operator.',
  line,
  column,
});

ruleTester.run('no-watch', {} as never, {
  valid: [
    {
      name: 'correct sample usage',
      code: `
import { createEvent, sample } from 'effector';

const clock = createEvent();
const derived = sample({ clock, fn: () => true });
`,
    },
    {
      name: 'unrelated watch method',
      code: `
const event = { watch: (fn: () => void) => fn() };

event.watch(console.log);
`,
    },
    {
      name: 'createWatch',
      code: `
import { createEvent, sample, createWatch } from 'effector';

const clock = createEvent();
const derived = sample({ clock, fn: () => true });

createWatch({ unit: derived, fn: console.log });
`,
    },
  ],
  invalid: [
    {
      name: 'store units',
      code: `
import { createStore } from 'effector';

const watcher = (v: unknown) => v;
const $watchable = createStore(0);

$watchable.watch(watcher);
$watchable.updates.watch(watcher);
`,
      errors: [restricted(7, 1), restricted(8, 1)],
    },
    {
      name: 'sample',
      code: `
import { createEvent, sample } from 'effector';

const watcher = (v: unknown) => v;
const clock = createEvent();
const derived = sample({ clock, fn: () => true });

sample({ clock, fn: () => true }).watch(watcher);
derived.watch(watcher);
`,
      errors: [restricted(8, 1), restricted(9, 1)],
    },
    {
      name: 'guard',
      code: `
import { createEvent, guard } from 'effector';

const watcher = (v: unknown) => v;
const clock = createEvent();

guard({ clock, filter: () => Math.random() > 0.5 }).watch(watcher);
`,
      errors: [restricted(7, 1)],
    },
    {
      name: 'event',
      code: `
import { createEvent } from 'effector';

const watcher = (v: unknown) => v;
const event = createEvent();

event.watch(watcher);
`,
      errors: [restricted(7, 1)],
    },
    {
      name: 'effect',
      code: `
import { createEffect } from 'effector';

const watcher = (v: unknown) => v;
const effectFx = createEffect();

effectFx.watch(watcher);

effectFx.done.watch(watcher);
effectFx.doneData.watch(watcher);

effectFx.fail.watch(watcher);
effectFx.failData.watch(watcher);

effectFx.finally.watch(watcher);
`,
      errors: [
        restricted(7, 1),
        restricted(9, 1),
        restricted(10, 1),
        restricted(12, 1),
        restricted(13, 1),
        restricted(15, 1),
      ],
    },
    {
      name: 'imported units',
      code: `
import { $store, event } from '../files/units';

$store.watch(console.log);
event.watch(console.log);
`,
      errors: [restricted(4, 1), restricted(5, 1)],
    },
  ],
});
