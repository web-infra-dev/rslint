// TestNoWatchUpstream migrates the full valid/invalid suite from upstream
// src/rules/no-watch/no-watch.test.ts (eslint-plugin-effector v0.19.0).
// Upstream runs against the real `effector` package; these cases use the
// `declare module 'effector'` stub in fixtures/effector.d.ts. rslint-specific
// cases live in no_watch_extras_test.go.

package no_watch_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/effector/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/effector/rules/no_watch"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const restrictedMessage = "Using `.watch` method leads to imperative code. Replace it with an operator `sample` or use the `target` parameter of `sample` operator."

func restricted(line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: "restricted", Message: restrictedMessage, Line: line, Column: column}
}

func TestNoWatchUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_watch.NoWatchRule,
		[]rule_tester.ValidTestCase{
			// correct sample usage
			{Code: `
import { createEvent, sample } from 'effector';

const clock = createEvent();
const derived = sample({ clock, fn: () => true });
`},
			// unrelated watch method
			{Code: `
const event = { watch: (fn: () => void) => fn() };

event.watch(console.log);
`},
			// createWatch
			{Code: `
import { createEvent, sample, createWatch } from 'effector';

const clock = createEvent();
const derived = sample({ clock, fn: () => true });

createWatch({ unit: derived, fn: console.log });
`},
		},
		[]rule_tester.InvalidTestCase{
			// store units
			{
				Code: `
import { createStore } from 'effector';

const watcher = (v: unknown) => v;
const $watchable = createStore(0);

$watchable.watch(watcher);
$watchable.updates.watch(watcher);
`,
				Errors: []rule_tester.InvalidTestCaseError{restricted(7, 1), restricted(8, 1)},
			},
			// sample
			{
				Code: `
import { createEvent, sample } from 'effector';

const watcher = (v: unknown) => v;
const clock = createEvent();
const derived = sample({ clock, fn: () => true });

sample({ clock, fn: () => true }).watch(watcher);
derived.watch(watcher);
`,
				Errors: []rule_tester.InvalidTestCaseError{restricted(8, 1), restricted(9, 1)},
			},
			// guard
			{
				Code: `
import { createEvent, guard } from 'effector';

const watcher = (v: unknown) => v;
const clock = createEvent();

guard({ clock, filter: () => Math.random() > 0.5 }).watch(watcher);
`,
				Errors: []rule_tester.InvalidTestCaseError{restricted(7, 1)},
			},
			// event
			{
				Code: `
import { createEvent } from 'effector';

const watcher = (v: unknown) => v;
const event = createEvent();

event.watch(watcher);
`,
				Errors: []rule_tester.InvalidTestCaseError{restricted(7, 1)},
			},
			// effect
			{
				Code: `
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
				Errors: []rule_tester.InvalidTestCaseError{
					restricted(7, 1),
					restricted(9, 1),
					restricted(10, 1),
					restricted(12, 1),
					restricted(13, 1),
					restricted(15, 1),
				},
			},
			// imported units
			{
				Code: `
import { $store, event } from './units';

$store.watch(console.log);
event.watch(console.log);
`,
				Errors: []rule_tester.InvalidTestCaseError{restricted(4, 1), restricted(5, 1)},
			},
		},
	)
}
